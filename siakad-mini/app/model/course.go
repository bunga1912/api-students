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

// CourseWithQuota adalah mata kuliah beserta kuota yang sudah terpakai.
// Terisi dihitung dari tabel enrollments, bukan disimpan di tabel courses.
type CourseWithQuota struct {
	Course
	Terisi    int `json:"terisi"`
	SisaKuota int `json:"sisa_kuota"`
}

// CourseFilter adalah parameter pada GET /courses.
// Semester 0 berarti semua semester, string kosong berarti tidak difilter.
type CourseFilter struct {
	Semester      int
	Search        string // mencari di kode_mk atau nama_mk
	OnlyAvailable bool   // hanya yang kuotanya belum penuh

	// TahunAkademik membatasi penghitungan terisi pada satu tahun akademik.
	// Kosong berarti semua tahun akademik.
	TahunAkademik string
}

type CourseRepository interface {
	FindByID(ctx context.Context, id int64) (*Course, error)

	// FindByIDForUpdate sama seperti FindByID tetapi mengunci barisnya
	// (SELECT ... FOR UPDATE). Hanya bermakna di dalam transaction.
	FindByIDForUpdate(ctx context.Context, id int64) (*Course, error)

	FindByCode(ctx context.Context, kodeMK string) (*Course, error)

	// ListWithQuota mengembalikan mata kuliah beserta terisi dan sisa_kuota.
	ListWithQuota(ctx context.Context, filter CourseFilter) ([]CourseWithQuota, error)

	List(ctx context.Context) ([]Course, error)
	ListBySemester(ctx context.Context, semester int) ([]Course, error)
	Create(ctx context.Context, course *Course) (*Course, error)
	Update(ctx context.Context, course *Course) (*Course, error)
	Delete(ctx context.Context, id int64) error
}