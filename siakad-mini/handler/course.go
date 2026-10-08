package handler

import (
	"strconv"

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

// GET /courses?semester=3  (semester opsional)
func (h *CourseHandler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if s := c.Query("semester"); s != "" {
		semester, err := strconv.Atoi(s)
		if err != nil {
			return badRequest(c, "semester tidak valid")
		}
		courses, err := h.svc.ListBySemester(ctx, semester)
		if err != nil {
			return respondError(c, err)
		}
		return c.JSON(courses)
	}

	courses, err := h.svc.List(ctx)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(courses)
}

func (h *CourseHandler) GetByID(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	course, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(course)
}

func (h *CourseHandler) GetByCode(c *fiber.Ctx) error {
	course, err := h.svc.GetByCode(c.UserContext(), c.Params("code"))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(course)
}

func (h *CourseHandler) Create(c *fiber.Ctx) error {
	var course model.Course
	if err := c.BodyParser(&course); err != nil {
		return badRequest(c, "format request tidak valid")
	}
	created, err := h.svc.Create(c.UserContext(), &course)
	if err != nil {
		return respondError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *CourseHandler) Update(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	var course model.Course
	if err := c.BodyParser(&course); err != nil {
		return badRequest(c, "format request tidak valid")
	}
	course.ID = id

	updated, err := h.svc.Update(c.UserContext(), &course)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(updated)
}

func (h *CourseHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return badRequest(c, "parameter id tidak valid")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return respondError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}