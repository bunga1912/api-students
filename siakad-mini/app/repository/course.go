package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

type courseRepository struct {
	db DBTX
}

func NewCourseRepository(db DBTX) model.CourseRepository {
	return &courseRepository{
		db: db,
	}
}

func (r *courseRepository) FindByID(
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

func (r *courseRepository) FindByCode(
	ctx context.Context,
	kodeMK string,
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
		WHERE kode_mk = $1
		`,
		kodeMK,
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

func (r *courseRepository) List(
	ctx context.Context,
) ([]model.Course, error) {

	rows, err := r.db.Query(
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
		ORDER BY semester, kode_mk
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	courses := make([]model.Course, 0)

	for rows.Next() {

		var course model.Course

		err := rows.Scan(
			&course.ID,
			&course.KodeMK,
			&course.NamaMK,
			&course.SKS,
			&course.Semester,
			&course.Kuota,
		)

		if err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *courseRepository) ListBySemester(
	ctx context.Context,
	semester int,
) ([]model.Course, error) {

	rows, err := r.db.Query(
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
		WHERE semester = $1
		ORDER BY kode_mk
		`,
		semester,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	courses := make([]model.Course, 0)

	for rows.Next() {

		var course model.Course

		err := rows.Scan(
			&course.ID,
			&course.KodeMK,
			&course.NamaMK,
			&course.SKS,
			&course.Semester,
			&course.Kuota,
		)

		if err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *courseRepository) Create(
	ctx context.Context,
	course *model.Course,
) (*model.Course, error) {

	result := &model.Course{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO courses (
			kode_mk,
			nama_mk,
			sks,
			semester,
			kuota
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			kode_mk,
			nama_mk,
			sks,
			semester,
			kuota
		`,
		course.KodeMK,
		course.NamaMK,
		course.SKS,
		course.Semester,
		course.Kuota,
	).Scan(
		&result.ID,
		&result.KodeMK,
		&result.NamaMK,
		&result.SKS,
		&result.Semester,
		&result.Kuota,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrAlreadyExists
		}

		return nil, err
	}

	return result, nil
}

func (r *courseRepository) Update(
	ctx context.Context,
	course *model.Course,
) (*model.Course, error) {

	result := &model.Course{}

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE courses
		SET
			kode_mk = $1,
			nama_mk = $2,
			sks = $3,
			semester = $4,
			kuota = $5
		WHERE id = $6
		RETURNING
			id,
			kode_mk,
			nama_mk,
			sks,
			semester,
			kuota
		`,
		course.KodeMK,
		course.NamaMK,
		course.SKS,
		course.Semester,
		course.Kuota,
		course.ID,
	).Scan(
		&result.ID,
		&result.KodeMK,
		&result.NamaMK,
		&result.SKS,
		&result.Semester,
		&result.Kuota,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		if isUniqueViolation(err) {
			return nil, model.ErrAlreadyExists
		}

		return nil, err
	}

	return result, nil
}

func (r *courseRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	commandTag, err := r.db.Exec(
		ctx,
		`
		DELETE FROM courses
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}