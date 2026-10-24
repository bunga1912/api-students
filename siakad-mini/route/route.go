package route

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/handler"
	"siakad-mini/middleware"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	Course     *handler.CourseHandler
	Student    *handler.StudentHandler
	Enrollment *handler.EnrollmentHandler
}

// auth adalah middleware autentikasi (middleware.Auth) yang dibuat di main.go.
func Setup(app *fiber.App, h Handlers, auth fiber.Handler) {
	api := app.Group("/api/v1")

	admin := middleware.RequireRole(model.RoleAdmin)
	mahasiswa := middleware.RequireRole(model.RoleMahasiswa)
	mahasiswaOrAdmin := middleware.RequireRole(model.RoleAdmin, model.RoleMahasiswa)

	// 1. Login (publik)
	api.Post("/auth/login", h.Auth.Login)

	// Semua route di bawah ini wajib token.
	authed := api.Group("", auth)

	// 2. Profil user yang sedang login
	authed.Get("/auth/me", h.Auth.Me)

	// 3-7. Students
	students := authed.Group("/students")
	students.Get("/", admin, h.Student.List)
	students.Post("/", admin, h.Student.Create)
	students.Get("/:id", mahasiswaOrAdmin, h.Student.GetByID) // mahasiswa: data sendiri
	students.Put("/:id", admin, h.Student.Update)
	students.Delete("/:id", admin, h.Student.Delete)

	// 8. Courses
	authed.Get("/courses", h.Course.List)

	// 9-10. Enrollments (KRS): khusus mahasiswa
	enrollments := authed.Group("/enrollments")
	enrollments.Post("/", mahasiswa, h.Enrollment.Create)
	enrollments.Delete("/:id", mahasiswa, h.Enrollment.Delete)
}