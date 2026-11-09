package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

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
	reviewService *service.ReviewService,
	portfolioService *service.PortfolioService,
	mongoClient *mongo.Client,
) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool, mongoClient))

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

	// Hak akses berdasarkan kepemilikan data.
	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)

	students.Get("/:nim/nilai", nilaiService.GetByNIM)

	// --- Portfolio mahasiswa (C.2) ---

	// GET: membaca portofolio mahasiswa.
	students.Get(
		"/:id/portfolio",
		portfolioService.Get,
	)

	// PUT: membuat atau mengganti seluruh portofolio.
	students.Put(
		"/:id/portfolio",
		middleware.RequirePermission(
			permissions,
			"portfolio:manage",
		),
		portfolioService.Put,
	)

	// POST: menambahkan satu proyek dengan $push.
	students.Post(
		"/:id/portfolio/projects",
		middleware.RequirePermission(
			permissions,
			"portfolio:manage",
		),
		portfolioService.AddProject,
	)

	// DELETE: menghapus proyek dengan $pull.
	students.Delete(
		"/:id/portfolio/projects/:projectId",
		middleware.RequirePermission(
			permissions,
			"portfolio:manage",
		),
		portfolioService.RemoveProject,
	)

	// --- kelas dan KRS ---
	courses := api.Group(
		"/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	courses.Get("/", courseService.List)
	courses.Get("/mine", courseService.Mine)
	courses.Get("/:id", courseService.Get)

	courses.Post("/:id/enroll", courseService.Enroll)
	courses.Delete("/:id/enroll", courseService.Drop)

	courses.Patch(
		"/:id",
		middleware.RequirePermission(permissions, "course:manage"),
		courseService.UpdateQuota,
	)

	// --- ulasan kelas ---
	courses.Get("/:id/reviews", reviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", reviewService.Summary)
	courses.Post("/:id/reviews", reviewService.Create)

	enrollments := api.Group(
		"/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	enrollments.Post("/swap", courseService.Swap)

	// --- pengelolaan ulasan ---
	reviews := api.Group(
		"/reviews",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	reviews.Get("/:id", reviewService.Get)
	reviews.Patch("/:id", reviewService.Update)
	reviews.Delete("/:id", reviewService.Delete)
	reviews.Post("/:id/replies", reviewService.Reply)
	reviews.Post("/:id/helpful", reviewService.MarkHelpful)

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

	// --- perpindahan pembimbing ---
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

	// --- statistik teknologi portofolio ---
	portfolios := api.Group(
		"/portfolios",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	portfolios.Get(
		"/stats/technologies",
		portfolioService.TechnologyStats,
	)
}

// healthCheck memeriksa PostgreSQL dan MongoDB.
func healthCheck(
	pool *pgxpool.Pool,
	mongoClient *mongo.Client,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable(
				"PostgreSQL tidak dapat dihubungi",
			)
		}

		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable(
				"MongoDB tidak dapat dihubungi",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"server, PostgreSQL, dan MongoDB berjalan",
			nil,
		)
	}
}
