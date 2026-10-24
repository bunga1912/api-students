package repository

import (
	"context"
	"fmt"
	"strings"

	"siakad-mini/app/model"
)

// ListWithQuota mengembalikan mata kuliah beserta terisi dan sisa_kuota.
// terisi dihitung dari tabel enrollments (opsional dibatasi satu tahun akademik).
func (r *courseRepository) ListWithQuota(
	ctx context.Context,
	f model.CourseFilter,
) ([]model.CourseWithQuota, error) {

	args := make([]interface{}, 0, 4)

	// Filter tahun akademik berlaku di subquery penghitung terisi.
	countWhere := ""
	if f.TahunAkademik != "" {
		args = append(args, f.TahunAkademik)
		countWhere = fmt.Sprintf("WHERE tahun_akademik = $%d", len(args))
	}

	where := make([]string, 0, 3)

	if f.Semester > 0 {
		args = append(args, f.Semester)
		where = append(where, fmt.Sprintf("c.semester = $%d", len(args)))
	}

	if f.Search != "" {
		args = append(args, "%"+likeEscaper.Replace(f.Search)+"%")
		where = append(where, fmt.Sprintf(
			"(c.kode_mk ILIKE $%[1]d OR c.nama_mk ILIKE $%[1]d)",
			len(args),
		))
	}

	if f.OnlyAvailable {
		where = append(where, "c.kuota > COALESCE(e.terisi, 0)")
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// countWhere dan whereSQL dibentuk dari konstanta di atas (bukan input
	// user); nilai dari user selalu lewat placeholder $n.
	query := fmt.Sprintf(
		`
		SELECT
			c.id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			c.semester,
			c.kuota,
			COALESCE(e.terisi, 0) AS terisi,
			GREATEST(c.kuota - COALESCE(e.terisi, 0), 0) AS sisa_kuota
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*)::int AS terisi
			FROM enrollments
			%s
			GROUP BY course_id
		) e ON e.course_id = c.id
		%s
		ORDER BY c.semester, c.kode_mk
		`,
		countWhere,
		whereSQL,
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	courses := make([]model.CourseWithQuota, 0)

	for rows.Next() {

		var c model.CourseWithQuota

		err := rows.Scan(
			&c.ID,
			&c.KodeMK,
			&c.NamaMK,
			&c.SKS,
			&c.Semester,
			&c.Kuota,
			&c.Terisi,
			&c.SisaKuota,
		)

		if err != nil {
			return nil, err
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}