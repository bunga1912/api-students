package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type NilaiRepository interface {
	FindByNIM(ctx context.Context, nim string) ([]model.Nilai, error)
}

type postgresNilaiRepository struct {
	db *pgxpool.Pool
}

func NewNilaiRepository(db *pgxpool.Pool) NilaiRepository {
	return &postgresNilaiRepository{db: db}
}

func (r *postgresNilaiRepository) FindByNIM(ctx context.Context, nim string) ([]model.Nilai, error) {
	query := `
		SELECT n.id, n.nama_matkul, n.nilai, s.nim, s.name
		FROM nilai n
		JOIN students s ON n.id_student = s.id
		WHERE s.nim = $1
		ORDER BY n.nama_matkul
	`
	rows, err := r.db.Query(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []model.Nilai
	for rows.Next() {
		var n model.Nilai
		if err := rows.Scan(&n.ID, &n.NamaMatkul, &n.Nilai, &n.NIM, &n.NamaStudent); err != nil {
			return nil, err
		}
		hasil = append(hasil, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hasil, nil
}