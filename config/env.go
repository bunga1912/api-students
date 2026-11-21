package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv memuat konfigurasi dari file .env.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: .env tidak ditemukan, membaca environment sistem")
	}
}

// GetEnv membaca environment variable.
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvInt membaca environment variable bertipe integer.
func GetEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("peringatan: %s bukan integer, menggunakan nilai default", key)
		return fallback
	}

	return result
}