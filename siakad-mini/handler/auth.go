package handler

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// loginResponse: access_token, token_type, expires_in, dan data user.
type loginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        userResponse `json:"user"`
}

type studentSummary struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

// meResponse: data user, ditambah "student" jika role mahasiswa.
type meResponse struct {
	userResponse
	Student *studentSummary `json:"student,omitempty"`
}

func newUserResponse(u *model.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Role: u.Role}
}

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidRequest(c, "Format request tidak valid")
	}

	result, err := h.svc.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Login berhasil", loginResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
		User:        newUserResponse(result.User),
	})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	result, err := h.svc.GetMe(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}

	resp := meResponse{userResponse: newUserResponse(result.User)}

	if result.Student != nil {
		resp.Student = &studentSummary{
			NIM:      result.Student.NIM,
			Nama:     result.Student.Nama,
			Prodi:    result.Student.Prodi,
			Angkatan: result.Student.Angkatan,
		}
	}

	return respondOK(c, "Profil berhasil diambil", resp)
}