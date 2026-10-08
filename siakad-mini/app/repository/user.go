package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"siakad-mini/app/model"
)

var (
	ErrNotFound      = errors.New("data tidak ditemukan")
	ErrAlreadyExists = errors.New("data sudah ada")
)

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) (*model.User, error)
}

type userRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) UserRepository {
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
			return nil, ErrNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByEmail(
	ctx context.Context,
	email string,
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
		WHERE LOWER(email) = LOWER($1)
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
			return nil, ErrNotFound
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
		return nil, err
	}

	return result, nil
}
