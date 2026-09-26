package model

type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
	OwnerID  *int    `json:"owner_id,omitempty"`
}

// POST — semua field wajib
type CreateStudentRequest struct {
	NIM   string  `json:"nim" validate:"required"`
	Name  string  `json:"name" validate:"required"`
	Grade float64 `json:"grade" validate:"min=0,max=100"`
}

// PUT — ganti seluruh isi, field bertipe biasa dan semuanya wajib
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

// PATCH — ubah sebagian, field bertipe pointer supaya bisa dibedakan
// antara "tidak dikirim" (nil) dan "dikirim bernilai kosong".
// omitnil dipakai, bukan omitempty, karena maksudnya adalah
// "lewati hanya bila nil" — bukan "lewati bila kosong".
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,min=1"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=1"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Amplop baku untuk semua respons
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}