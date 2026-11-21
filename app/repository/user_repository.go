package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, u model.User) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)

	// FindByCourse mengembalikan seluruh peserta sebuah kelas, dipakai
	// worker untuk mengirim notifikasi ke semua peserta (D.2).
	FindByCourse(ctx context.Context, courseID int) ([]model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	query := `
		INSERT INTO users (username, email, password, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password, role, is_active, created_at
	`
	var hasil model.User
	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role, u.IsActive).
		Scan(&hasil.ID, &hasil.Username, &hasil.Email, &hasil.Password,
			&hasil.Role, &hasil.IsActive, &hasil.CreatedAt)

	if isUniqueViolation(err) {
		return model.User{}, ErrDuplicate
	}
	if err != nil {
		return model.User{}, fmt.Errorf("membuat user: %w", err)
	}
	return hasil, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at
		FROM users WHERE id = $1
	`
	var u model.User
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindByUsername dipakai saat login. Pencocokan tidak membedakan
// huruf besar dan kecil.
func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at
		FROM users WHERE LOWER(username) = LOWER($1)
	`
	var u model.User
	err := r.pool.QueryRow(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindByCourse mengambil peserta sebuah kelas, diurutkan menurut id agar
// urutan pengiriman email selalu sama pada setiap attempt.
func (r *userPostgresRepository) FindByCourse(ctx context.Context, courseID int) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.username, u.email, u.password, u.role, u.is_active, u.created_at
		FROM users u
		JOIN enrollments e ON e.user_id = u.id
		WHERE e.course_id = $1
		ORDER BY u.id`, courseID)
	if err != nil {
		return nil, fmt.Errorf("mengambil peserta kelas: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password,
			&u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca row peserta: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}