package repository

import (
	"context"

	"siakad-mini/app/model"
)

// ListDetailByStudent mengembalikan KRS seorang mahasiswa.
// tahunAkademik kosong berarti semua tahun akademik.
func (r *enrollmentRepository) ListDetailByStudent(
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
		  AND ($2::text = '' OR e.tahun_akademik = $2)
		ORDER BY e.tahun_akademik, c.kode_mk
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