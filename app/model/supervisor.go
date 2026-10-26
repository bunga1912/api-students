package model

import "time"

type Supervisor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Quota           int       `json:"quota"`
	SupervisedCount int       `json:"supervised_count"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
}

// AssignSupervisorRequest dipakai POST /supervisors/:id/assign.
type AssignSupervisorRequest struct {
	StudentID int `json:"student_id" validate:"required,min=1"`
}

// SwitchSupervisionRequest dipakai POST /supervisions/switch.
// Pembimbing lama tidak dikirim client: ia dibaca dari database di dalam
// transaction, karena seorang mahasiswa hanya punya satu pembimbing.
type SwitchSupervisionRequest struct {
	StudentID      int `json:"student_id" validate:"required,min=1"`
	ToSupervisorID int `json:"to_supervisor_id" validate:"required,min=1"`
}

// UpdateSupervisorQuotaRequest membawa version yang terakhir DILIHAT client
// (optimistic locking).
type UpdateSupervisorQuotaRequest struct {
	Quota   int `json:"quota" validate:"required,min=1,max=500"`
	Version int `json:"version" validate:"required,min=1"`
}