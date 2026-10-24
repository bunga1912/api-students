package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

// likeEscaper meng-escape karakter khusus LIKE supaya input pencarian
// diperlakukan sebagai teks biasa.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type studentRepository struct {
	db DBTX
}

func NewStudentRepository(db DBTX) model.StudentRepository {
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
			return nil, model.ErrNotFound
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
			return nil, model.ErrNotFound
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
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return student, nil
}

func (r *studentRepository) ListPaged(
	ctx context.Context,
	f model.StudentFilter,
) ([]model.Student, int, error) {

	where := []string{"deleted_at IS NULL"}
	args := make([]interface{}, 0, 5)

	if f.Prodi != "" {
		args = append(args, f.Prodi)
		where = append(where, fmt.Sprintf("LOWER(prodi) = LOWER($%d)", len(args)))
	}

	if f.Angkatan > 0 {
		args = append(args, f.Angkatan)
		where = append(where, fmt.Sprintf("angkatan = $%d", len(args)))
	}

	if f.Search != "" {
		args = append(args, "%"+likeEscaper.Replace(f.Search)+"%")
		where = append(where, fmt.Sprintf(
			"(nim ILIKE $%[1]d OR nama ILIKE $%[1]d)",
			len(args),
		))
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int

	err := r.db.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM students "+whereSQL,
		args...,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	// orderBy hanya berasal dari konstanta di bawah (bukan input user),
	// jadi aman disisipkan ke SQL.
	orderBy := "nim"

	switch f.Sort {
	case model.StudentSortNama:
		orderBy = "nama, id"
	case model.StudentSortIPKDesc:
		orderBy = "ipk_terakhir DESC, id"
	}

	args = append(args, f.PerPage, (f.Page-1)*f.PerPage)

	query := fmt.Sprintf(
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
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
		`,
		whereSQL,
		orderBy,
		len(args)-1,
		len(args),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	students := make([]model.Student, 0, f.PerPage)

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
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
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
		if constraint, ok := uniqueViolationConstraint(err); ok {
			// Termasuk NIM milik mahasiswa yang sudah di-soft delete,
			// karena constraint UNIQUE tetap berlaku untuk baris tersebut.
			if strings.Contains(constraint, "nim") {
				return nil, model.FieldError("nim", "NIM sudah terdaftar")
			}

			return nil, model.ErrAlreadyExists
		}

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
			return nil, model.ErrNotFound
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
		return model.ErrNotFound
	}

	return nil
}