package handler

import (
	"github.com/gofiber/fiber/v2"

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

// resolveStudentID: mahasiswa selalu memakai data dirinya sendiri,
// admin harus menentukan student_id secara eksplisit.
// Mengembalikan (id, ok, err). Jika ok == false, response sudah dikirim (err dari c.JSON).
func (h *EnrollmentHandler) resolveStudentID(c *fiber.Ctx, explicit int64) (int64, bool, error) {
	if isAdmin(c) {
		if explicit <= 0 {
			return 0, false, badRequest(c, "student_id wajib diisi")
		}
		return explicit, true, nil
	}

	student, err := h.studentSvc.GetByUserID(c.UserContext(), currentUserID(c))
	if err != nil {
		return 0, false, respondError(c, err)
	}
	return student.ID, true, nil
}

type createEnrollmentRequest struct {
	StudentID     int64  `json:"student_id"` // hanya dipakai admin
	CourseID      int64  `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

// POST /enrollments
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var req createEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "format request tidak valid")
	}

	studentID, ok, err := h.resolveStudentID(c, req.StudentID)
	if !ok {
		return err
	}

	enrollment, err := h.enrollmentSvc.Create(
		c.UserContext(),
		studentID,
		req.CourseID,
		req.TahunAkademik,
	)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(enrollment)
}

// GET /enrollments?tahun_akademik=2025/2026[&student_id=1 (admin)]
func (h *EnrollmentHandler) List(c *fiber.Ctx) error {
	studentID, ok, err := h.resolveStudentID(c, int64(c.QueryInt("student_id", 0)))
	if !ok {
		return err
	}

	list, err := h.enrollmentSvc.ListByStudent(
		c.UserContext(),
		studentID,
		c.Query("tahun_akademik"),
	)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(list)
}

// GET /enrollments/total-sks?tahun_akademik=2025/2026[&student_id=1 (admin)]
func (h *EnrollmentHandler) TotalSKS(c *fiber.Ctx) error {
	studentID, ok, err := h.resolveStudentID(c, int64(c.QueryInt("student_id", 0)))
	if !ok {
		return err
	}

	total, err := h.enrollmentSvc.GetTotalSKS(
		c.UserContext(),
		studentID,
		c.Query("tahun_akademik"),
	)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(fiber.Map{"total_sks": total})
}

// DELETE /enrollments/:id
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	if err := h.enrollmentSvc.Delete(c.UserContext(), id); err != nil {
		return respondError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}