package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Login(
	ctx context.Context,
	req LoginRequest,
) (*LoginResponse, error) {

	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email dan password wajib diisi")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("email atau password salah")
		}

		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		return nil, errors.New("email atau password salah")
	}

	token, err := helper.GenerateJWT(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		return nil, err
	}

	user.Password = ""

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
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
