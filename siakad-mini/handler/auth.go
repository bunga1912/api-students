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

type loginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
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

	token, user, err := h.svc.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Login berhasil", loginResponse{
		Token: token,
		User:  user,
	})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, err := h.svc.GetMe(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Profil berhasil diambil", user)
}