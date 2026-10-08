package handler

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "format request tidak valid")
	}

	res, err := h.svc.Login(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, err := h.svc.GetMe(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(user)
}