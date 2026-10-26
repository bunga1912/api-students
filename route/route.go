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
	app *fiber.App,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	nilaiService *service.NilaiService,
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
	permissions *helper.PermissionSet,
	courseService *service.CourseService,
	supervisorService *service.SupervisorService,
) {
	api := app.Group("/api/v1")

	// --- health check ---
	api.Get("/health", healthCheck(pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// --- mahasiswa ---
	students := api.Group(
		"/students",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	students.Get(
		"/",
		middleware.RequirePermission(permissions, "student:list"),
		studentService.List,
	)

	students.Get(
		"/cursor",
		middleware.RequirePermission(permissions, "student:list"),
		studentService.ListCursor,
	)

	students.Post(
		"/",
		middleware.RequirePermission(permissions, "student:create"),
		studentService.Create,
	)

	students.Delete(
		"/:id",
		middleware.RequirePermission(permissions, "student:delete"),
		studentService.Delete,
	)

	// Hak akses berdasarkan kepemilikan data diperiksa di service.
	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Get("/:nim/nilai", nilaiService.GetByNIM)

	// --- kelas dan KRS ---
	courses := api.Group(
		"/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	courses.Get("/", courseService.List)

	// Daftarkan /mine sebelum /:id.
	courses.Get("/mine", courseService.Mine)
	courses.Get("/:id", courseService.Get)

	courses.Post("/:id/enroll", courseService.Enroll)
	courses.Delete("/:id/enroll", courseService.Drop)

	courses.Patch(
		"/:id",
		middleware.RequirePermission(permissions, "course:manage"),
		courseService.UpdateQuota,
	)

	enrollments := api.Group(
		"/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	enrollments.Post("/swap", courseService.Swap)

	// --- pembimbing tugas akhir ---
	supervisors := api.Group(
		"/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	supervisors.Get("/", supervisorService.List)
	supervisors.Get("/:id", supervisorService.Get)

	supervisors.Post(
		"/:id/assign",
		middleware.RequirePermission(permissions, "supervision:manage"),
		supervisorService.Assign,
	)

	supervisors.Patch(
		"/:id",
		middleware.RequirePermission(permissions, "supervision:manage"),
		supervisorService.UpdateQuota,
	)

	// --- pengelolaan pembimbing tugas akhir ---
	supervisions := api.Group(
		"/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	supervisions.Post(
		"/switch",
		middleware.RequirePermission(permissions, "supervision:manage"),
		supervisorService.Switch,
	)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable(
				"database tidak dapat dihubungi",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"server dan database berjalan",
			nil,
		)
	}
}