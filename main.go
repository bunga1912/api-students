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

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPostgresPool()
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		log.Fatal(err)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "api-students"),
		time.Duration(mustAtoi(config.GetEnv("JWT_ACCESS_TTL_MINUTES", "15")))*time.Minute,
	)

	// Pemetaan role ke permission dibaca SEKALI saat aplikasi menyala.
	roleRepo := repository.NewRoleRepository(pool)
	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo, permissions)

	nilaiRepo := repository.NewNilaiRepository(pool)
	nilaiService := service.NewNilaiService(nilaiRepo)

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	authService := service.NewAuthService(
		userRepo, tokenRepo, jwtManager, permissions,
		time.Duration(mustAtoi(config.GetEnv("JWT_REFRESH_TTL_DAYS", "7")))*24*time.Hour,
	)

	app := config.NewApp(logger, pool, studentService, nilaiService, authService, jwtManager, permissions)

	port := config.GetEnv("APP_PORT", "8000")
	logger.Info("server berjalan", slog.String("port", port))
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