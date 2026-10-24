package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
)

type CourseHandler struct {
	svc *service.CourseService
}

func NewCourseHandler(svc *service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// GET /api/v1/courses
//
// Query (semua opsional):
//
//	semester=3
//	search=basis            (kode_mk atau nama_mk, tidak case-sensitive)
//	available=true          (hanya yang kuotanya belum penuh)
//	tahun_akademik=2026/2027-Ganjil  (terisi dihitung hanya pada tahun ini)
func (h *CourseHandler) List(c *fiber.Ctx) error {
	verr := model.NewValidationError()

	filter := model.CourseFilter{
		Search:        c.Query("search"),
		TahunAkademik: c.Query("tahun_akademik"),
	}

	if raw := strings.TrimSpace(c.Query("semester")); raw != "" {
		n, err := strconv.Atoi(raw)

		switch {
		case err != nil:
			verr.Add("semester", "harus berupa angka")
		case n < 1:
			verr.Add("semester", "minimal 1")
		default:
			filter.Semester = n
		}
	}

	switch strings.ToLower(strings.TrimSpace(c.Query("available"))) {
	case "", "false":
		// semua mata kuliah
	case "true":
		filter.OnlyAvailable = true
	default:
		verr.Add("available", "harus true atau false")
	}

	if verr.HasErrors() {
		return respondError(c, verr)
	}

	courses, err := h.svc.ListWithQuota(c.UserContext(), filter)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Data mata kuliah berhasil diambil", courses)
}