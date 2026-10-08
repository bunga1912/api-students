package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

type StudentRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Student, error)
	FindByUserID(ctx context.Context, userID int64) (*model.Student, error)
	FindByNIM(ctx context.Context, nim string) (*model.Student, error)
	List(ctx context.Context) ([]model.Student, error)
	Create(ctx context.Context, student *model.Student) (*model.Student, error)
	Update(ctx context.Context, student *model.Student) (*model.Student, error)
	Delete(ctx context.Context, id int64) error
}

type studentRepository struct {
	db DBTX
}

func NewStudentRepository(db DBTX) StudentRepository {
	return &studentRepository{
		db: db,
	}
}

func (r *studentRepository) FindByID(
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
			return nil, ErrNotFound
		}

		return nil, err
	}

	return student, nil
}

func (r *studentRepository) FindByUserID(
	ctx context.Context,
	userID int64,
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
		WHERE user_id = $1
		  AND deleted_at IS NULL
		`,
		userID,
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
			return nil, ErrNotFound
		}

		return nil, err
	}

	return student, nil
}

func (r *studentRepository) FindByNIM(
	ctx context.Context,
	nim string,
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
		WHERE nim = $1
		  AND deleted_at IS NULL
		`,
		nim,
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
			return nil, ErrNotFound
		}

		return nil, err
	}

	return student, nil
}

func (r *studentRepository) List(
	ctx context.Context,
) ([]model.Student, error) {

	rows, err := r.db.Query(
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
		WHERE deleted_at IS NULL
		ORDER BY nim
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {

		var student model.Student

		err := rows.Scan(
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
			return nil, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (r *studentRepository) Create(
	ctx context.Context,
	student *model.Student,
) (*model.Student, error) {

	result := &model.Student{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO students (
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at,
			created_at
		`,
		student.UserID,
		student.NIM,
		student.Nama,
		student.Prodi,
		student.Angkatan,
		student.IPKTerakhir,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.NIM,
		&result.Nama,
		&result.Prodi,
		&result.Angkatan,
		&result.IPKTerakhir,
		&result.DeletedAt,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *studentRepository) Update(
	ctx context.Context,
	student *model.Student,
) (*model.Student, error) {

	result := &model.Student{}

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE students
		SET
			nama = $1,
			prodi = $2,
			angkatan = $3,
			ipk_terakhir = $4
		WHERE id = $5
		  AND deleted_at IS NULL
		RETURNING
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at,
			created_at
		`,
		student.Nama,
		student.Prodi,
		student.Angkatan,
		student.IPKTerakhir,
		student.ID,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.NIM,
		&result.Nama,
		&result.Prodi,
		&result.Angkatan,
		&result.IPKTerakhir,
		&result.DeletedAt,
		&result.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return result, nil
}

func (r *studentRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	commandTag, err := r.db.Exec(
		ctx,
		`
		UPDATE students
		SET deleted_at = NOW()
		WHERE id = $1
		  AND deleted_at IS NULL
		`,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
