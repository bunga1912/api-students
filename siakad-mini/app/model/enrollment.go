package model

import "time"

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
