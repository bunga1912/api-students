package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

func NewApp(
	logger *slog.Logger, pool *pgxpool.Pool,
	studentService *service.StudentService, nilaiService *service.NilaiService,
	authService *service.AuthService, jwtManager *helper.JWTManager,
	permissions *helper.PermissionSet,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Tugas Mandiri - API Students"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
	})

	middleware.Register(app, logger)
	route.Register(app, pool, studentService, nilaiService, authService, jwtManager, permissions)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler adalah SATU-SATUNYA tempat error berubah menjadi
// response HTTP di seluruh aplikasi.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)

		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Kegagalan yang sudah kita rencanakan.

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}

		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status: fiberErr.Code, Code: "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		// Hanya kegagalan sisi server yang dicatat sebagai Error.
		// Kegagalan 4xx adalah kesalahan pemakai API, bukan kerusakan
		// sistem; mencatatnya sebagai Error membuat log penuh bising
		// sehingga kerusakan yang sesungguhnya justru tenggelam.
		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", appErr.Error()))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}