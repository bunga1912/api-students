package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"api-students/config"
)

// NewMongoClient menyiapkan client MongoDB dan memastikan servernya benar-benar
// dapat dihubungi.
//
// Berbeda dari pgxpool, mongo.Connect pada driver v2 TIDAK menghubungi server
// sama sekali: ia hanya menyiapkan client. Alamat yang salah atau server yang
// mati baru ketahuan saat query pertama. Karena itu Ping wajib dilakukan di
// sini, supaya aplikasi menolak menyala daripada gagal pada request pertama.
func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	maxPool := uint64(20)
	if n, err := strconv.Atoi(config.GetEnv("MONGO_MAX_POOL", "20")); err == nil && n > 0 {
		maxPool = uint64(n)
	}

	opts := options.Client().
		ApplyURI(config.GetEnv("MONGO_URI", "mongodb://localhost:27017")).
		SetServerSelectionTimeout(5 * time.Second).
		SetMaxPoolSize(maxPool)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("menyiapkan client MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}