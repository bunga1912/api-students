package model

import "errors"

var (
	// Umum
	ErrNotFound      = errors.New("data tidak ditemukan")
	ErrAlreadyExists = errors.New("data sudah ada")
	ErrForbidden     = errors.New("akses ditolak")

	// Auth
	ErrInvalidCredentials = errors.New("email atau password salah")

	// ErrAccountInactive: token masih berlaku tetapi akun mahasiswanya
	// sudah di-soft delete.
	ErrAccountInactive = errors.New("akun tidak aktif")

	// KRS
	ErrSudahDiambil = errors.New("mata kuliah sudah diambil pada tahun akademik tersebut")
	ErrKuotaPenuh   = errors.New("kuota mata kuliah sudah penuh")
	ErrSKSMelebihi  = errors.New("pengambilan mata kuliah melebihi batas SKS")
)

// ValidationError menampung pesan error per field.
// Dipetakan ke HTTP 422 dengan format {"errors": {"field": ["pesan"]}}.
type ValidationError struct {
	Fields map[string][]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Fields: make(map[string][]string),
	}
}

func (e *ValidationError) Add(field, message string) {
	e.Fields[field] = append(e.Fields[field], message)
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

func (e *ValidationError) Error() string {
	return "validasi gagal"
}

// FieldError membuat ValidationError untuk satu field.
func FieldError(field, message string) error {
	v := NewValidationError()
	v.Add(field, message)

	return v
}