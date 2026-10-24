package service

import (
	"context"
	"errors"
	"net/mail"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/app/model"
)

// TokenIssuer dibuat di sisi service supaya service tidak bergantung
// pada implementasi JWT (helper) maupun config.
type TokenIssuer interface {
	Generate(userID int64, email, role string) (string, error)
}

type AuthService struct {
	userRepo model.UserRepository
	tokens   TokenIssuer
}

func NewAuthService(
	userRepo model.UserRepository,
	tokens TokenIssuer,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		tokens:   tokens,
	}
}

func validateLogin(email, password string) error {
	verr := model.NewValidationError()

	if email == "" {
		verr.Add("email", "wajib diisi")
	} else if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		verr.Add("email", "format email tidak valid")
	}

	switch {
	case password == "":
		verr.Add("password", "wajib diisi")
	case len(password) < 8:
		verr.Add("password", "minimal 8 karakter")
	}

	if verr.HasErrors() {
		return verr
	}

	return nil
}

// Login mengembalikan token dan data user (password sudah dikosongkan).
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, *model.User, error) {

	if err := validateLogin(email, password); err != nil {
		return "", nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return "", nil, model.ErrInvalidCredentials
		}

		return "", nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return "", nil, model.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		return "", nil, err
	}

	user.Password = ""

	return token, user, nil
}

func (s *AuthService) GetMe(
	ctx context.Context,
	userID int64,
) (*model.User, error) {

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return user, nil
}