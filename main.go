package main

import (
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

	// Rahasia diperiksa SEBELUM server menyala. Lebih baik gagal seketika
	// daripada berjalan dengan token yang mudah dipalsukan.
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

	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo)

	nilaiRepo := repository.NewNilaiRepository(pool)
	nilaiService := service.NewNilaiService(nilaiRepo)

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	authService := service.NewAuthService(
		userRepo, tokenRepo, jwtManager,
		time.Duration(mustAtoi(config.GetEnv("JWT_REFRESH_TTL_DAYS", "7")))*24*time.Hour,
	)

	app := config.NewApp(logger, pool, studentService, nilaiService, authService, jwtManager)

	port := config.GetEnv("APP_PORT", "8000")
	logger.Info("server berjalan", slog.String("port", port))
	log.Fatal(app.Listen(":" + port))
}

func mustAtoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 15 // fallback aman bila env var rusak
		}
		n = n*10 + int(r-'0')
	}
	return n
}