package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

func (r *studentRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*model.Student, error) {

	student := &model.Student{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at,
			created_at
		FROM students
		WHERE id = $1
		  AND deleted_at IS NULL
		FOR UPDATE
		`,
		id,
	).Scan(
		&student.ID,
		&student.UserID,
		&student.NIM,
		&student.Nama,
		&student.Prodi,
		&student.Angkatan,
		&student.IPKTerakhir,
		&student.DeletedAt,
		&student.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return student, nil
}