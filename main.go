package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	// --- JWT ---
	jwtSecret := config.GetEnv("JWT_SECRET", "")

	if len(jwtSecret) < minSecretLength {
		logger.Error(
			"JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength),
		)
		os.Exit(1)
	}

	// --- PostgreSQL ---
	pool, err := database.NewPostgresPool()
	if err != nil {
		logger.Error(
			"gagal terhubung ke database",
			slog.String("error", err.Error()),
		)
		log.Fatal(err)
	}
	defer pool.Close()

	// --- MongoDB ---
	mongoClient, err := database.NewMongoClient(
		context.Background(),
	)
	if err != nil {
		logger.Error(
			"gagal terhubung ke MongoDB",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_ = mongoClient.Disconnect(ctx)
	}()

	mongoDB := mongoClient.Database(
		config.GetEnv("MONGO_DATABASE", "api_students"),
	)

	// --- JWT Manager ---
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "api-students"),
		time.Duration(
			mustAtoi(
				config.GetEnv("JWT_ACCESS_TTL_MINUTES", "15"),
			),
		)*time.Minute,
	)

	// --- Permissions ---
	roleRepo := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepo.LoadPermissions(
		context.Background(),
	)
	if err != nil {
		logger.Error(
			"gagal memuat permission",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)

	logger.Info(
		"permission dimuat",
		slog.Any("roles", permissions.KnownRoles()),
	)

	// --- Students ---
	studentRepo := repository.NewStudentRepository(pool)

	studentService := service.NewStudentService(
		studentRepo,
		permissions,
	)

	// --- Nilai ---
	nilaiRepo := repository.NewNilaiRepository(pool)

	nilaiService := service.NewNilaiService(
		nilaiRepo,
	)

	// --- Authentication ---
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)

	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
		permissions,
		time.Duration(
			mustAtoi(
				config.GetEnv("JWT_REFRESH_TTL_DAYS", "7"),
			),
		)*24*time.Hour,
	)

	// --- Courses / KRS ---
	courseRepository := repository.NewCourseRepository(pool)
	txManager := repository.NewTxManager(pool)

	courseService := service.NewCourseService(
		courseRepository,
		txManager,
	)

	// --- Pembimbing Tugas Akhir ---
	supervisorRepository := repository.NewSupervisorRepository(pool)

	supervisorService := service.NewSupervisorService(
		supervisorRepository,
		txManager,
	)

	// --- Review ---
	reviewRepository := repository.NewReviewRepository(
		mongoDB,
	)

	if err := reviewRepository.EnsureIndexes(
		context.Background(),
	); err != nil {
		logger.Error(
			"gagal membuat index Review MongoDB",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	reviewService := service.NewReviewService(
		reviewRepository,
		courseRepository,
		permissions,
	)

	// --- Portfolio ---

	// Repository PostgreSQL untuk memeriksa mahasiswa.
	studentLookup := repository.NewStudentLookup(pool)

	// Repository MongoDB untuk portofolio.
	portfolioRepository := repository.NewPortfolioRepository(
		mongoDB,
	)

	// Membuat index portofolio.
	indexCtx, cancelIndex := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	err = portfolioRepository.EnsureIndexes(indexCtx)
	cancelIndex()

	if err != nil {
		logger.Error(
			"gagal membuat index Portfolio MongoDB",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// Service portofolio.
	portfolioService := service.NewPortfolioService(
		portfolioRepository,
		studentLookup,
	)

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
	)

	// --- Run Server ---
	port := config.GetEnv("APP_PORT", "8000")

	logger.Info(
		"server berjalan",
		slog.String("port", port),
	)

	log.Fatal(app.Listen(":" + port))
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
