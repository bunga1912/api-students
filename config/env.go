package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv memuat .env. Bila berkas tidak ditemukan, program tetap lanjut —
// nilai bisa saja disuntikkan lewat environment variable sistem.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: .env tidak ditemukan, membaca dari environment sistem")
	}
}

// GetEnv membaca env var, atau memakai nilai bawaan bila kosong.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}