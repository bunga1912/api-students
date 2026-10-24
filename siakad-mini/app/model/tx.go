package model

import "context"

// TxRepositories adalah kumpulan repository yang terikat ke satu transaction.
type TxRepositories struct {
	Users       UserRepository
	Students    StudentRepository
	Courses     CourseRepository
	Enrollments EnrollmentRepository
}

// TxManager menjalankan fn di dalam satu transaction.
// Jika fn mengembalikan error (atau panic), transaction di-rollback.
// Jika fn berhasil, transaction di-commit.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(repos TxRepositories) error) error
}