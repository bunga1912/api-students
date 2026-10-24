package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

type DBTX interface {
	Exec(
		ctx context.Context,
		sql string,
		arguments ...interface{},
	) (pgconn.CommandTag, error)

	Query(
		ctx context.Context,
		sql string,
		args ...interface{},
	) (pgx.Rows, error)

	QueryRow(
		ctx context.Context,
		sql string,
		args ...interface{},
	) pgx.Row
}

// uniqueViolationConstraint mengembalikan nama constraint jika err adalah
// pelanggaran UNIQUE di PostgreSQL.
func uniqueViolationConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return pgErr.ConstraintName, true
	}

	return "", false
}

// isUniqueViolation true jika error berasal dari pelanggaran UNIQUE.
func isUniqueViolation(err error) bool {
	_, ok := uniqueViolationConstraint(err)

	return ok
}