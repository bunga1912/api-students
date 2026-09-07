package main

import (
	"log"
	"log/slog"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
)

// main hanya berisi urutan perakitan. Tidak ada logika bisnis,
// tidak ada query, tidak ada satu pun handler di sini.
func main() {
	// 1. Konfigurasi dan logger
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Database
	pool, err := database.NewPostgresPool()
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		log.Fatal(err)
	}
	defer pool.Close()

	// 3. Perakitan dari dalam ke luar: repository -> service
	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo)

	nilaiRepo := repository.NewNilaiRepository(pool)
	nilaiService := service.NewNilaiService(nilaiRepo)

	// 4. Aplikasi
	app := config.NewApp(logger, pool, studentService, nilaiService)

	port := config.GetEnv("APP_PORT", "8000")
	logger.Info("server berjalan", slog.String("port", port))
	log.Fatal(app.Listen(":" + port))
}