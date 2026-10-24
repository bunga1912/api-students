package model

import (
	"context"
	"time"
)

type Student struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Nilai sort yang diizinkan pada GET /students.
const (
	StudentSortNama    = "nama"
	StudentSortIPKDesc = "-ipk_terakhir"
)

// StudentFilter adalah parameter pencarian daftar mahasiswa.
// Angkatan 0 berarti semua angkatan, string kosong berarti tidak difilter.
type StudentFilter struct {
	Prodi    string
	Angkatan int
	Search   string // mencari di nim atau nama
	Sort     string // "", StudentSortNama, atau StudentSortIPKDesc
	Page     int
	PerPage  int
}

// StudentPage adalah satu halaman hasil pencarian.
type StudentPage struct {
	Items    []Student
	Total    int
	Page     int
	PerPage  int
	LastPage int
}

type StudentRepository interface {
	FindByID(ctx context.Context, id int64) (*Student, error)

	// FindByIDForUpdate sama seperti FindByID tetapi mengunci barisnya
	// (SELECT ... FOR UPDATE). Hanya bermakna di dalam transaction.
	FindByIDForUpdate(ctx context.Context, id int64) (*Student, error)

	FindByUserID(ctx context.Context, userID int64) (*Student, error)
	FindByNIM(ctx context.Context, nim string) (*Student, error)

	// ListPaged mengembalikan satu halaman mahasiswa (yang belum di-soft delete)
	// beserta total seluruh data yang cocok dengan filter.
	ListPaged(ctx context.Context, filter StudentFilter) ([]Student, int, error)

	Create(ctx context.Context, student *Student) (*Student, error)
	Update(ctx context.Context, student *Student) (*Student, error)
	Delete(ctx context.Context, id int64) error
}