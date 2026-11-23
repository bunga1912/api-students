package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	_ "api-students/docs"
	"api-students/helper"
)

const minSecretLength = 32

// @title                       API KRS Praktikum Backend Lanjut
// @version                     1.0
// @description                 API pendaftaran kelas (KRS), ulasan kelas, dan notifikasi.
// @description                 Seluruh endpoint selain /auth memerlukan access token.
// @BasePath                    /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Isi dengan: Bearer <access_token>

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	// Satu program, tiga cara menjalankan:
	//   go run . -role=api
	//   go run . -role=worker
	//   go run . -role=all
	role := flag.String("role", config.GetEnv("APP_ROLE", "all"), "api | worker | all")
	flag.Parse()

	runAPI := *role == "api" || *role == "all"
	runWorker := *role == "worker" || *role == "all"
	if !runAPI && !runWorker {
		logger.Error("role tidak dikenal", slog.String("role", *role))
		os.Exit(1)
	}

	// --- PostgreSQL ---
	pool, err := database.NewPostgresPool()
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var app *fiber.App
	if runAPI {
		app = startAPI(logger, pool)
		if app == nil {
			os.Exit(1)
		}
	}

	var stopWorker func()
	if runWorker {
		stopWorker = startWorker(ctx, logger, pool)
		if stopWorker == nil {
			os.Exit(1)
		}
	}

	logger.Info("aplikasi berjalan", slog.String("role", *role))
	<-ctx.Done()

	if app != nil {
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			logger.Error("shutdown HTTP tidak rapi", slog.String("error", err.Error()))
		}
	}
	if stopWorker != nil {
		stopWorker()
	}
	logger.Info("aplikasi berhenti dengan rapi")
}

func startAPI(logger *slog.Logger, pool *pgxpool.Pool) *fiber.App {
	// --- JWT ---
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error(
			"JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength),
		)
		return nil
	}

	// --- MongoDB ---
	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		return nil
	}
	mongoDB := mongoClient.Database(
		config.GetEnv("MONGO_DATABASE", "api_students"),
	)

	// --- JWT Manager ---
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "api-students"),
		time.Duration(mustAtoi(config.GetEnv("JWT_ACCESS_TTL_MINUTES", "15")))*time.Minute,
	)

	// --- Permissions ---
	roleRepo := repository.NewRoleRepository(pool)
	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		return nil
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	// --- Students ---
	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo, permissions)

	// --- Nilai ---
	nilaiRepo := repository.NewNilaiRepository(pool)
	nilaiService := service.NewNilaiService(nilaiRepo)

	// --- Authentication ---
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
		permissions,
		time.Duration(mustAtoi(config.GetEnv("JWT_REFRESH_TTL_DAYS", "7")))*24*time.Hour,
	)

	// --- Courses / KRS ---
	courseRepository := repository.NewCourseRepository(pool)
	outboxRepository := repository.NewOutboxRepository(pool)
	notificationRepository := repository.NewNotificationRepository(pool)
	txManager := repository.NewTxManager(pool)

	courseService := service.NewCourseService(
		courseRepository,
		outboxRepository,
		txManager,
	)

	// --- Pembimbing Tugas Akhir ---
	supervisorRepository := repository.NewSupervisorRepository(pool)
	supervisorService := service.NewSupervisorService(supervisorRepository, txManager)

	// --- Review ---
	reviewRepository := repository.NewReviewRepository(mongoDB)
	if err := reviewRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index Review MongoDB", slog.String("error", err.Error()))
		return nil
	}
	reviewService := service.NewReviewService(reviewRepository, courseRepository, permissions)

	// --- Portfolio ---
	studentLookup := repository.NewStudentLookup(pool)
	portfolioRepository := repository.NewPortfolioRepository(mongoDB)

	indexCtx, cancelIndex := context.WithTimeout(context.Background(), 10*time.Second)
	err = portfolioRepository.EnsureIndexes(indexCtx)
	cancelIndex()
	if err != nil {
		logger.Error("gagal membuat index Portfolio MongoDB", slog.String("error", err.Error()))
		return nil
	}
	portfolioService := service.NewPortfolioService(portfolioRepository, studentLookup)

	// --- Notifikasi ---
	inspector := asynq.NewInspector(config.RedisOpt())
	notificationService := service.NewNotificationService(
		notificationRepository,
		outboxRepository,
		inspector,
	)

	// --- Swagger ---
	enableSwagger := config.GetEnv("SWAGGER_ENABLED", "false") == "true"

	// --- Application ---
	app := config.NewApp(
		logger,
		pool,
		studentService,
		nilaiService,
		authService,
		jwtManager,
		permissions,
		courseService,
		supervisorService,
		reviewService,
		portfolioService,
		mongoClient,
		notificationService,
		enableSwagger,
	)

	app.Hooks().OnShutdown(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = inspector.Close()
		return mongoClient.Disconnect(ctx)
	})

	port := config.GetEnv("APP_PORT", "8000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))
	return app
}

func mustAtoi(s string) int {
	n := 0

	for _, r := range s {
		if r < '0' || r > '9' {
			return 15
		}

		n = n*10 + int(r-'0')
	}

	return n
}
