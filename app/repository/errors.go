package repository

import (
	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation mengenali kode error PostgreSQL 23505
// (unique_violation), yang muncul saat constraint UNIQUE dilanggar —
// dalam kasus ini, saat NIM yang sama dimasukkan dua kali.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if err == nil {
		return false
	}
	if e, ok := err.(*pgconn.PgError); ok {
		pgErr = e
	}
	return pgErr != nil && pgErr.Code == "23505"
}