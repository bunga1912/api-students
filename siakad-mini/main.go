package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/handler"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"
)

func main() {
	// Load environment variable dari .env
	config.LoadEnv()

	ctx := context.Background()

	// Konfigurasi database
	dbConfig := database.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnv("DB_PORT", "5432"),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", ""),
		Name:     config.GetEnv("DB_NAME", "siakad_mini"),
		SSLMode:  config.GetEnv("DB_SSLMODE", "disable"),
	}

	// Koneksi database
	pool, err := database.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	log.Println("database berhasil terhubung")

	// =========================
	// JWT (gagal cepat jika secret belum diisi)
	// =========================

	tokens, err := helper.NewJWT(
		config.GetEnv("JWT_SECRET", ""),
		24*time.Hour,
	)
	if err != nil {
		log.Fatalf("konfigurasi JWT: %v", err)
	}

	// =========================
	// REPOSITORY
	// =========================

	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	txManager := repository.NewTxManager(pool)

	// =========================
	// SERVICE
	// =========================

	authService := service.NewAuthService(userRepo, tokens)
	studentService := service.NewStudentService(studentRepo, enrollmentRepo, txManager)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, txManager)

	// =========================
	// FIBER APP + ROUTES
	// =========================

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	route.Setup(
		app,
		route.Handlers{
			Auth:       handler.NewAuthHandler(authService),
			Course:     handler.NewCourseHandler(courseService),
			Student:    handler.NewStudentHandler(studentService),
			Enrollment: handler.NewEnrollmentHandler(enrollmentService, studentService),
		},
		middleware.Auth(tokens),
	)

	// =========================
	// SERVER (graceful shutdown)
	// =========================

	port := config.GetEnv("APP_PORT", "8001")

	go func() {
		log.Printf("server berjalan di http://localhost:%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("menghentikan server...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}