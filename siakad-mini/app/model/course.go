package model

import "context"

type Course struct {
	ID       int64  `json:"id"`
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
	Kuota    int    `json:"kuota"`
}

type CourseRepository interface {
	FindByID(ctx context.Context, id int64) (*Course, error)

	// FindByIDForUpdate sama seperti FindByID tetapi mengunci barisnya
	// (SELECT ... FOR UPDATE). Hanya bermakna di dalam transaction.
	FindByIDForUpdate(ctx context.Context, id int64) (*Course, error)

	FindByCode(ctx context.Context, kodeMK string) (*Course, error)
	List(ctx context.Context) ([]Course, error)
	ListBySemester(ctx context.Context, semester int) ([]Course, error)
	Create(ctx context.Context, course *Course) (*Course, error)
	Update(ctx context.Context, course *Course) (*Course, error)
	Delete(ctx context.Context, id int64) error
}