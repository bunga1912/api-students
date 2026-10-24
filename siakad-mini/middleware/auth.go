package middleware

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
)

// abort mengirim error berformat seragam.
// Tidak memakai package handler supaya tidak terjadi import melingkar.
func abort(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// Auth memvalidasi JWT dari header "Authorization: Bearer <token>".
func Auth(tokens *helper.JWT) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw, err := helper.ExtractBearerToken(c.Get("Authorization"))
		if err != nil {
			return abort(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
		}

		claims, err := tokens.Parse(raw)
		if err != nil {
			return abort(c, fiber.StatusUnauthorized, "Token tidak valid atau sudah kedaluwarsa")
		}

		c.Locals(CtxUserID, claims.UserID)
		c.Locals(CtxRole, claims.Role)
		return c.Next()
	}
}

// RequireRole membatasi akses endpoint hanya untuk role tertentu.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(CtxRole).(string)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return abort(c, fiber.StatusForbidden, "Akses ditolak")
	}
}