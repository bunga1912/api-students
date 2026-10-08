package route

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/handler"
	"siakad-mini/middleware"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	Course     *handler.CourseHandler
	Student    *handler.StudentHandler
	Enrollment *handler.EnrollmentHandler
}

func Setup(app *fiber.App, h Handlers) {
	api := app.Group("/api")

	// ---------- Public ----------
	api.Post("/auth/login", h.Auth.Login)

	// ---------- Authenticated ----------
	authed := api.Group("", middleware.Auth())

	authed.Get("/auth/me", h.Auth.Me)

	admin := middleware.RequireRole(middleware.RoleAdmin)
	mahasiswaOrAdmin := middleware.RequireRole(middleware.RoleAdmin, middleware.RoleMahasiswa)

	// Courses: baca untuk semua user login, tulis khusus admin
	courses := authed.Group("/courses")
	courses.Get("/", h.Course.List) // ?semester=
	courses.Get("/code/:code", h.Course.GetByCode)
	courses.Get("/:id", h.Course.GetByID)
	courses.Post("/", admin, h.Course.Create)
	courses.Put("/:id", admin, h.Course.Update)
	courses.Delete("/:id", admin, h.Course.Delete)

	// Students (rute statis "/me" & "/nim/:nim" didaftarkan sebelum "/:id")
	students := authed.Group("/students")
	students.Get("/me", mahasiswaOrAdmin, h.Student.Me)
	students.Get("/nim/:nim", admin, h.Student.GetByNIM)
	students.Get("/", admin, h.Student.List)
	students.Get("/:id", admin, h.Student.GetByID)
	students.Post("/", admin, h.Student.Create)
	students.Put("/:id", admin, h.Student.Update)
	students.Delete("/:id", admin, h.Student.Delete)

	// Enrollments (KRS)
	enrollments := authed.Group("/enrollments")
	enrollments.Post("/", mahasiswaOrAdmin, h.Enrollment.Create)
	enrollments.Get("/", mahasiswaOrAdmin, h.Enrollment.List)
	enrollments.Get("/total-sks", mahasiswaOrAdmin, h.Enrollment.TotalSKS)
	enrollments.Delete("/:id", admin, h.Enrollment.Delete)
}