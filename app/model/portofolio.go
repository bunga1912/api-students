package model

import "time"

// Portfolio adalah bentuk portofolio yang dikenal seluruh aplikasi.
//
// Seperti Review, tidak ada tipe milik driver MongoDB di sini: ID disimpan
// sebagai string. student_id merujuk students.id di PostgreSQL; tidak ada
// FOREIGN KEY lintas database, sehingga service yang memeriksanya.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Project disimpan DI DALAM dokumen portofolio (embedding). Setiap proyek
// punya id sendiri agar dapat dihapus satu per satu dengan $pull.
type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

// Certificate memakai tanggal berformat YYYY-MM-DD. Di database ia disimpan
// sebagai tipe date BSON.
type Certificate struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

type ProjectInput struct {
	Title        string   `json:"title" validate:"required,min=2,max=100"`
	Technologies []string `json:"technologies" validate:"min=1,max=10,dive,min=1,max=30"`
	URL          string   `json:"url" validate:"omitempty,http_url,max=300"`
	Year         int      `json:"year" validate:"required,min=2000,max=2100"`
}

type CertificateInput struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Issuer string `json:"issuer" validate:"required,min=2,max=100"`
	Date   string `json:"date" validate:"required,datetime=2006-01-02"`
}

// PortfolioRequest dipakai PUT: mengganti SELURUH isi portofolio.
type PortfolioRequest struct {
	Headline     string             `json:"headline" validate:"required,min=5,max=100"`
	Bio          string             `json:"bio" validate:"max=1000"`
	Projects     []ProjectInput     `json:"projects" validate:"max=20,dive"`
	Certificates []CertificateInput `json:"certificates" validate:"max=20,dive"`
}

type TechnologyCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}