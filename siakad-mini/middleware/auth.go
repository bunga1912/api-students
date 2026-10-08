package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"

	RoleAdmin     = "admin"
	RoleMahasiswa = "mahasiswa"
)

// Auth memvalidasi JWT dari header "Authorization: Bearer <token>".
func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		parts := strings.SplitN(c.Get("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "token tidak ditemukan"})
		}

		// ASUMSI: helper.ParseJWT mengembalikan claims berisi UserID (int64) & Role (string).
		claims, err := helper.ParseJWT(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "token tidak valid"})
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
		return c.Status(fiber.StatusForbidden).
			JSON(fiber.Map{"error": "akses ditolak"})
	}
}