package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

type userRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) model.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {

	user := &model.User{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			email,
			password,
			role,
			created_at
		FROM users
		WHERE id = $1
		`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return user, nil
}

// FindByEmail dipakai untuk login. User yang mahasiswanya sudah di-soft delete
// diperlakukan seperti tidak ada, sehingga tidak bisa login.
func (r *userRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {

	user := &model.User{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			u.id,
			u.email,
			u.password,
			u.role,
			u.created_at
		FROM users u
		WHERE LOWER(u.email) = LOWER($1)
		  AND NOT EXISTS (
			SELECT 1
			FROM students s
			WHERE s.user_id = u.id
			  AND s.deleted_at IS NOT NULL
		  )
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *userRepository) Create(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {

	result := &model.User{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			email,
			password,
			role,
			created_at
		`,
		user.Email,
		user.Password,
		user.Role,
	).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&result.Role,
		&result.CreatedAt,
	)

	if err != nil {
		if constraint, ok := uniqueViolationConstraint(err); ok {
			if strings.Contains(constraint, "email") {
				return nil, model.FieldError("email", "email sudah terdaftar")
			}

			return nil, model.ErrAlreadyExists
		}

		return nil, err
	}

	return result, nil
}