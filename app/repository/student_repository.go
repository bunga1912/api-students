package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("nim sudah terdaftar")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q StudentListFilter) ([]Student, int, error)
	FindByID(ctx context.Context, id int) (Student, error)
	Create(ctx context.Context, s Student) (Student, error)
	Update(ctx context.Context, s Student) (Student, error)
	Delete(ctx context.Context, id int) error
}

type Student struct {
	ID       int
	NIM      string
	Name     string
	Grade    float64
	IsActive bool
	OwnerID  *int
}

type StudentListFilter struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

type postgresStudentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &postgresStudentRepository{db: db}
}

func (r *postgresStudentRepository) FindAll(ctx context.Context, f StudentListFilter) ([]Student, int, error) {
	query := `
		SELECT id, nim, name, grade, is_active, owner_id
		FROM students
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		  AND ($2::boolean IS NULL OR is_active = $2)
		ORDER BY ` + f.Sort + ` ` + f.Order + `
		LIMIT $3 OFFSET $4
	`
	offset := (f.Page - 1) * f.Limit

	rows, err := r.db.Query(ctx, query, f.Search, f.IsActive, f.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var hasil []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID); err != nil {
			return nil, 0, err
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM students
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		  AND ($2::boolean IS NULL OR is_active = $2)
	`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, f.Search, f.IsActive).Scan(&total); err != nil {
		return nil, 0, err
	}

	return hasil, total, nil
}

func (r *postgresStudentRepository) FindByID(ctx context.Context, id int) (Student, error) {
	query := `SELECT id, nim, name, grade, is_active, owner_id FROM students WHERE id = $1`

	var s Student
	err := r.db.QueryRow(ctx, query, id).
		Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Student{}, ErrNotFound
	}
	if err != nil {
		return Student{}, err
	}
	return s, nil
}

// Create menyimpan owner_id sesuai yang dikirim service. Owner_id TIDAK
// pernah berasal dari body request klien — ia diambil dari identitas
// pemanggil (lihat student_service.go), sehingga klien tidak mungkin
// mengaku sebagai pemilik user lain.
func (r *postgresStudentRepository) Create(ctx context.Context, s Student) (Student, error) {
	query := `
		INSERT INTO students (nim, name, grade, is_active, owner_id)
		VALUES ($1, $2, $3, true, $4)
		RETURNING id, nim, name, grade, is_active, owner_id
	`
	var hasil Student
	err := r.db.QueryRow(ctx, query, s.NIM, s.Name, s.Grade, s.OwnerID).
		Scan(&hasil.ID, &hasil.NIM, &hasil.Name, &hasil.Grade, &hasil.IsActive, &hasil.OwnerID)

	if isUniqueViolation(err) {
		return Student{}, ErrDuplicate
	}
	if err != nil {
		return Student{}, err
	}
	return hasil, nil
}

func (r *postgresStudentRepository) Update(ctx context.Context, s Student) (Student, error) {
	query := `
		UPDATE students
		SET nim = $1, name = $2, grade = $3, is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, owner_id
	`
	var hasil Student
	err := r.db.QueryRow(ctx, query, s.NIM, s.Name, s.Grade, s.IsActive, s.ID).
		Scan(&hasil.ID, &hasil.NIM, &hasil.Name, &hasil.Grade, &hasil.IsActive, &hasil.OwnerID)

	if errors.Is(err, pgx.ErrNoRows) {
		return Student{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Student{}, ErrDuplicate
	}
	if err != nil {
		return Student{}, err
	}
	return hasil, nil
}

func (r *postgresStudentRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM students WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}