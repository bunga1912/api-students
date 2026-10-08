package config

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/database"
)

type App struct {
	DB     *pgxpool.Pool
	Server *http.Server
}

func NewApp() (*App, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	maxConns, err := strconv.Atoi(
		GetEnv("DB_MAX_CONNS", "10"),
	)

	if err != nil {
		maxConns = 10
	}

	db, err := database.NewPool(
		ctx,
		database.Config{
			Host:     GetEnv("DB_HOST", "localhost"),
			Port:     GetEnv("DB_PORT", "5432"),
			User:     GetEnv("DB_USER", "postgres"),
			Password: GetEnv("DB_PASSWORD", ""),
			Name:     GetEnv("DB_NAME", "siakad_mini"),
			SSLMode:  GetEnv("DB_SSLMODE", "disable"),
			MaxConns: int32(maxConns),
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"gagal koneksi database: %w",
			err,
		)
	}

	port := GetEnv("APP_PORT", "8001")

	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		DB:     db,
		Server: server,
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
