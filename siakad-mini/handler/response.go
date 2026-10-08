package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/middleware"
	"siakad-mini/app/repository"
)

func respondError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "data tidak ditemukan"})
	}
	return c.Status(fiber.StatusInternalServerError).
		JSON(fiber.Map{"error": err.Error()})
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func parseID(c *fiber.Ctx, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func currentUserID(c *fiber.Ctx) int64 {
	id, _ := c.Locals(middleware.CtxUserID).(int64)
	return id
}

func isAdmin(c *fiber.Ctx) bool {
	role, _ := c.Locals(middleware.CtxRole).(string)
	return role == middleware.RoleAdmin
}