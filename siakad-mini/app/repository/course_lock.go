package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

func (r *courseRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*model.Course, error) {

	course := &model.Course{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			kode_mk,
			nama_mk,
			sks,
			semester,
			kuota
		FROM courses
		WHERE id = $1
		FOR UPDATE
		`,
		id,
	).Scan(
		&course.ID,
		&course.KodeMK,
		&course.NamaMK,
		&course.SKS,
		&course.Semester,
		&course.Kuota,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return course, nil
}