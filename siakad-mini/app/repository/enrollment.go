package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

type EnrollmentRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Enrollment, error)

	FindByStudentCourseYear(
		ctx context.Context,
		studentID int64,
		courseID int64,
		tahunAkademik string,
	) (*model.Enrollment, error)

	ListByStudent(
		ctx context.Context,
		studentID int64,
		tahunAkademik string,
	) ([]model.EnrollmentDetail, error)

	CountByCourse(
		ctx context.Context,
		courseID int64,
		tahunAkademik string,
	) (int, error)

	GetTotalSKS(
		ctx context.Context,
		studentID int64,
		tahunAkademik string,
	) (int, error)

	Create(
		ctx context.Context,
		enrollment *model.Enrollment,
	) (*model.Enrollment, error)

	Delete(ctx context.Context, id int64) error
}

type enrollmentRepository struct {
	db DBTX
}

func NewEnrollmentRepository(db DBTX) EnrollmentRepository {
	return &enrollmentRepository{
		db: db,
	}
}

func (r *enrollmentRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.Enrollment, error) {

	enrollment := &model.Enrollment{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			student_id,
			course_id,
			tahun_akademik,
			created_at
		FROM enrollments
		WHERE id = $1
		`,
		id,
	).Scan(
		&enrollment.ID,
		&enrollment.StudentID,
		&enrollment.CourseID,
		&enrollment.TahunAkademik,
		&enrollment.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return enrollment, nil
}

func (r *enrollmentRepository) FindByStudentCourseYear(
	ctx context.Context,
	studentID int64,
	courseID int64,
	tahunAkademik string,
) (*model.Enrollment, error) {

	enrollment := &model.Enrollment{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			student_id,
			course_id,
			tahun_akademik,
			created_at
		FROM enrollments
		WHERE student_id = $1
		  AND course_id = $2
		  AND tahun_akademik = $3
		`,
		studentID,
		courseID,
		tahunAkademik,
	).Scan(
		&enrollment.ID,
		&enrollment.StudentID,
		&enrollment.CourseID,
		&enrollment.TahunAkademik,
		&enrollment.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return enrollment, nil
}

func (r *enrollmentRepository) ListByStudent(
	ctx context.Context,
	studentID int64,
	tahunAkademik string,
) ([]model.EnrollmentDetail, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			e.id,
			e.student_id,
			s.nim,
			s.nama,
			e.course_id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			e.tahun_akademik
		FROM enrollments e
		JOIN students s
			ON s.id = e.student_id
		JOIN courses c
			ON c.id = e.course_id
		WHERE e.student_id = $1
		  AND e.tahun_akademik = $2
		ORDER BY c.kode_mk
		`,
		studentID,
		tahunAkademik,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	enrollments := make([]model.EnrollmentDetail, 0)

	for rows.Next() {

		var enrollment model.EnrollmentDetail

		err := rows.Scan(
			&enrollment.ID,
			&enrollment.StudentID,
			&enrollment.NIM,
			&enrollment.Nama,
			&enrollment.CourseID,
			&enrollment.KodeMK,
			&enrollment.NamaMK,
			&enrollment.SKS,
			&enrollment.TahunAkademik,
		)

		if err != nil {
			return nil, err
		}

		enrollments = append(enrollments, enrollment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return enrollments, nil
}

func (r *enrollmentRepository) CountByCourse(
	ctx context.Context,
	courseID int64,
	tahunAkademik string,
) (int, error) {

	var count int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1
		  AND tahun_akademik = $2
		`,
		courseID,
		tahunAkademik,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *enrollmentRepository) GetTotalSKS(
	ctx context.Context,
	studentID int64,
	tahunAkademik string,
) (int, error) {

	var totalSKS int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c
			ON c.id = e.course_id
		WHERE e.student_id = $1
		  AND e.tahun_akademik = $2
		`,
		studentID,
		tahunAkademik,
	).Scan(&totalSKS)

	if err != nil {
		return 0, err
	}

	return totalSKS, nil
}

func (r *enrollmentRepository) Create(
	ctx context.Context,
	enrollment *model.Enrollment,
) (*model.Enrollment, error) {

	result := &model.Enrollment{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO enrollments (
			student_id,
			course_id,
			tahun_akademik
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			student_id,
			course_id,
			tahun_akademik,
			created_at
		`,
		enrollment.StudentID,
		enrollment.CourseID,
		enrollment.TahunAkademik,
	).Scan(
		&result.ID,
		&result.StudentID,
		&result.CourseID,
		&result.TahunAkademik,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *enrollmentRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	commandTag, err := r.db.Exec(
		ctx,
		`
		DELETE FROM enrollments
		WHERE id = $1
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
