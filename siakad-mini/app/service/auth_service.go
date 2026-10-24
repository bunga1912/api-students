package service

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/app/model"
)

// TokenIssuer dibuat di sisi service supaya service tidak bergantung
// pada implementasi JWT (helper) maupun config.
type TokenIssuer interface {
	Generate(userID int64, email, role string) (string, error)
	TTL() time.Duration
}

// TokenTypeBearer adalah nilai token_type pada response login.
const TokenTypeBearer = "Bearer"

type AuthService struct {
	userRepo    model.UserRepository
	studentRepo model.StudentRepository
	tokens      TokenIssuer
}

func NewAuthService(
	userRepo model.UserRepository,
	studentRepo model.StudentRepository,
	tokens TokenIssuer,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		tokens:      tokens,
	}
}

// LoginResult adalah hasil login yang berhasil.
type LoginResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64 // detik
	User        *model.User
}

// MeResult adalah profil user yang sedang login.
// Student hanya terisi jika role-nya mahasiswa.
type MeResult struct {
	User    *model.User
	Student *model.Student
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
) (*LoginResult, error) {

	if err := validateLogin(email, password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrInvalidCredentials
		}

		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return nil, model.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		return nil, err
	}

	user.Password = ""

	return &LoginResult{
		AccessToken: token,
		TokenType:   TokenTypeBearer,
		ExpiresIn:   int64(s.tokens.TTL() / time.Second),
		User:        user,
	}, nil
}

// GetMe mengembalikan profil user. Untuk role mahasiswa, data mahasiswanya
// (nim, nama, prodi, angkatan) ikut disertakan.
//
// Mahasiswa yang sudah di-soft delete tidak boleh lagi memakai token lama,
// sehingga dikembalikan ErrAccountInactive (401).
func (s *AuthService) GetMe(
	ctx context.Context,
	userID int64,
) (*MeResult, error) {

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	result := &MeResult{User: user}

	if user.Role != model.RoleMahasiswa {
		return result, nil
	}

	student, err := s.studentRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrAccountInactive
		}

		return nil, err
	}

	result.Student = student

	return result, nil
}