package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

// pageMeta adalah objek "meta" pada response berpagination.
type pageMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

func respondPaginated(
	c *fiber.Ctx,
	message string,
	data interface{},
	meta pageMeta,
) error {
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// queryInt membaca query parameter bertipe angka.
// Jika bukan angka, error dicatat di verr dan default dikembalikan.
func queryInt(c *fiber.Ctx, key string, def int, verr *model.ValidationError) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return def
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		verr.Add(key, "harus berupa angka")
		return def
	}

	return n
}