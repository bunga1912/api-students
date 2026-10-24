package model

import (
	"context"
	"time"
)

type Enrollment struct {
	ID            int64     `json:"id"`
	StudentID     int64     `json:"student_id"`
	CourseID      int64     `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type EnrollmentDetail struct {
	ID            int64   `json:"id"`
	StudentID     int64   `json:"student_id"`
	NIM           string  `json:"nim"`
	Nama          string  `json:"nama"`
	CourseID      int64   `json:"course_id"`
	KodeMK        string  `json:"kode_mk"`
	NamaMK        string  `json:"nama_mk"`
	SKS           int     `json:"sks"`
	TahunAkademik string  `json:"tahun_akademik"`
	Nilai         *string `json:"nilai,omitempty"`
}

type EnrollmentRepository interface {
	FindByID(ctx context.Context, id int64) (*Enrollment, error)

	FindByStudentCourseYear(
		ctx context.Context,
		studentID int64,
		courseID int64,
		tahunAkademik string,
	) (*Enrollment, error)

	ListByStudent(
		ctx context.Context,
		studentID int64,
		tahunAkademik string,
	) ([]EnrollmentDetail, error)

	// ListDetailByStudent seperti ListByStudent, tetapi tahunAkademik kosong
	// berarti semua tahun akademik.
	ListDetailByStudent(
		ctx context.Context,
		studentID int64,
		tahunAkademik string,
	) ([]EnrollmentDetail, error)

	CountByCourse(
		ctx context.Context,
		courseID int64,
		tahunAkademik string,
	) (int, error)

	GetTotalSKS(
		ctx context.Context,
		studentID int64,
		tahunAkademik string,
	) (int, error)

	Create(
		ctx context.Context,
		enrollment *Enrollment,
	) (*Enrollment, error)

	Delete(ctx context.Context, id int64) error
}