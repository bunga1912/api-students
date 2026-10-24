package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/service"
)

type CourseHandler struct {
	svc *service.CourseService
}

func NewCourseHandler(svc *service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// GET /api/v1/courses?semester=3  (semester opsional)
func (h *CourseHandler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if s := c.Query("semester"); s != "" {
		semester, err := strconv.Atoi(s)
		if err != nil {
			return validationFailed(c, "semester", "harus berupa angka")
		}

		courses, err := h.svc.ListBySemester(ctx, semester)
		if err != nil {
			return respondError(c, err)
		}

		return respondOK(c, "Data mata kuliah berhasil diambil", courses)
	}

	courses, err := h.svc.List(ctx)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Data mata kuliah berhasil diambil", courses)
}