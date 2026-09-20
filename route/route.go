package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

func Register(
	app *fiber.App, pool *pgxpool.Pool,
	studentService *service.StudentService, nilaiService *service.NilaiService,
	authService *service.AuthService, jwtManager *helper.JWTManager,
	permissions *helper.PermissionSet,
) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))

	// --- autentikasi (publik, kecuali /me) ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// --- wajib login, hak akses diperiksa per endpoint ---
	students := api.Group("/students",
		middleware.RequireJSON, middleware.RequireAuth(jwtManager))

	// Hak dapat diputuskan TANPA melihat data -> middleware.
	students.Get("/",
		middleware.RequirePermission(permissions, "student:list"),
		studentService.List)
	students.Post("/",
		middleware.RequirePermission(permissions, "student:create"),
		studentService.Create)
	students.Delete("/:id",
		middleware.RequirePermission(permissions, "student:delete"),
		studentService.Delete)

	// Hak bergantung pada KEPEMILIKAN data -> diperiksa di service.
	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)

	students.Get("/:nim/nilai", nilaiService.GetByNIM)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}