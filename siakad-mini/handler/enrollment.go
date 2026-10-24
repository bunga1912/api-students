package handler

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
)

type EnrollmentHandler struct {
	enrollmentSvc *service.EnrollmentService
	studentSvc    *service.StudentService
}

func NewEnrollmentHandler(
	enrollmentSvc *service.EnrollmentService,
	studentSvc *service.StudentService,
) *EnrollmentHandler {
	return &EnrollmentHandler{
		enrollmentSvc: enrollmentSvc,
		studentSvc:    studentSvc,
	}
}

type createEnrollmentRequest struct {
	CourseID      int64  `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

// POST /api/v1/enrollments  (khusus mahasiswa; role dicek di route)
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var req createEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidRequest(c, "Format request tidak valid")
	}

	student, err := h.studentSvc.GetByUserID(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}

	enrollment, err := h.enrollmentSvc.Create(
		c.UserContext(),
		student.ID,
		req.CourseID,
		req.TahunAkademik,
	)
	if err != nil {
		return respondError(c, err)
	}

	return respondCreated(c, "Mata kuliah berhasil diambil", enrollment)
}

// DELETE /api/v1/enrollments/:id  (khusus mahasiswa pemilik)
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return respondError(c, model.ErrNotFound)
	}

	student, err := h.studentSvc.GetByUserID(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}

	if err := h.enrollmentSvc.Delete(c.UserContext(), student.ID, id); err != nil {
		return respondError(c, err)
	}

	return respondNoContent(c)
}