package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

// TxBeginner dipenuhi oleh *pgxpool.Pool.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type txManager struct {
	db TxBeginner
}

func NewTxManager(db TxBeginner) model.TxManager {
	return &txManager{
		db: db,
	}
}

func (m *txManager) WithinTx(
	ctx context.Context,
	fn func(repos model.TxRepositories) error,
) error {

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	// Aman dipanggil setelah Commit: hasilnya diabaikan (ErrTxClosed).
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	repos := model.TxRepositories{
		Users:       NewUserRepository(tx),
		Students:    NewStudentRepository(tx),
		Courses:     NewCourseRepository(tx),
		Enrollments: NewEnrollmentRepository(tx),
	}

	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit(ctx)
}