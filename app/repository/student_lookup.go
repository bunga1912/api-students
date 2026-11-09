package repository

import (
	"context"
	"fmt"
)

// StudentLookup menjawab satu pertanyaan saja: apakah seorang mahasiswa ada di
// PostgreSQL. Dipakai layer lain, misalnya layanan portofolio di MongoDB, yang
// tidak boleh mengandalkan FOREIGN KEY karena datanya berada di database lain.
type StudentLookup interface {
	Exists(ctx context.Context, id int) (bool, error)
}

type studentLookupPostgres struct{ db DBTX }

func NewStudentLookup(db DBTX) StudentLookup {
	return &studentLookupPostgres{db: db}
}

func (r *studentLookupPostgres) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM students WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("memeriksa mahasiswa: %w", err)
	}
	return exists, nil
}