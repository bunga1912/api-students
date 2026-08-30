package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"api-students/app/repository"
	"api-students/database"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// requireJSON menolak request berisi body yang Content-Type-nya bukan JSON.
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}

func main() {
	// Muat kredensial dari .env. Bila berkas tidak ditemukan, program
	// tetap lanjut — nilai bisa saja disuntikkan lewat environment
	// variable sistem (misalnya saat deploy).
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: .env tidak ditemukan, membaca dari environment sistem")
	}

	pool, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal("gagal terhubung ke database: ", err)
	}
	defer pool.Close()

	repo = repository.NewStudentRepository(pool)

	app := fiber.New(fiber.Config{
		AppName: "Tugas Mandiri Pertemuan 3 - API Students",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return fail(c, status, pesan)
		},
	})

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New())

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
		}
		return ok(c, "server dan database berjalan", fiber.Map{"status": "up"})
	})

	s := api.Group("/students", requireJSON)
	s.Get("/", listStudents)
	s.Get("/:id", getStudent)
	s.Post("/", createStudent)
	s.Put("/:id", replaceStudent)
	s.Patch("/:id", patchStudent)
	s.Delete("/:id", deleteStudent)

	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	fmt.Println("Server berjalan di http://localhost:8000")
	log.Fatal(app.Listen(":8000"))
}