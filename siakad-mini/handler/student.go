package handler

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
)

type StudentHandler struct {
	svc *service.StudentService
}

func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	students, err := h.svc.List(c.UserContext())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(students)
}

// GET /students/me -> profil mahasiswa dari user yang login
func (h *StudentHandler) Me(c *fiber.Ctx) error {
	student, err := h.svc.GetByUserID(c.UserContext(), currentUserID(c))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(student)
}

func (h *StudentHandler) GetByID(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	student, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(student)
}

func (h *StudentHandler) GetByNIM(c *fiber.Ctx) error {
	student, err := h.svc.GetByNIM(c.UserContext(), c.Params("nim"))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(student)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var student model.Student
	if err := c.BodyParser(&student); err != nil {
		return badRequest(c, "format request tidak valid")
	}
	created, err := h.svc.Create(c.UserContext(), &student)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	var student model.Student
	if err := c.BodyParser(&student); err != nil {
		return badRequest(c, "format request tidak valid")
	}
	student.ID = id

	updated, err := h.svc.Update(c.UserContext(), &student)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(updated)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return respondError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}