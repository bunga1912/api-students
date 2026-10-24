package handler

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/middleware"
)

// Format response sukses:
//
//	{"success": true, "message": "...", "data": ..., "meta": {...}}
type response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// Format response error:
//
//	{"success": false, "message": "...", "errors": {"field": ["pesan"]}}
type errorResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func respondOK(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondCreated(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondNoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// respondError memetakan error domain ke status HTTP.
// Error yang tidak dikenal dicatat di log dan TIDAK dikirim ke client.
func respondError(c *fiber.Ctx, err error) error {
	var verr *model.ValidationError
	if errors.As(err, &verr) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse{
			Success: false,
			Message: "Validasi gagal",
			Errors:  verr.Fields,
		})
	}

	status := fiber.StatusInternalServerError

	switch {
	case errors.Is(err, model.ErrNotFound):
		status = fiber.StatusNotFound

	case errors.Is(err, model.ErrInvalidCredentials),
		errors.Is(err, model.ErrAccountInactive):
		status = fiber.StatusUnauthorized

	case errors.Is(err, model.ErrForbidden):
		status = fiber.StatusForbidden

	case errors.Is(err, model.ErrSudahDiambil):
		status = fiber.StatusConflict

	case errors.Is(err, model.ErrKuotaPenuh),
		errors.Is(err, model.ErrSKSMelebihi),
		errors.Is(err, model.ErrAlreadyExists):
		status = fiber.StatusUnprocessableEntity

	default:
		log.Printf("internal error: %v", err)
	}

	msg := "Terjadi kesalahan pada server"
	if status != fiber.StatusInternalServerError {
		msg = capitalize(err.Error())
	}

	return c.Status(status).JSON(errorResponse{
		Success: false,
		Message: msg,
	})
}

// validationFailed mengirim 422 untuk satu field.
func validationFailed(c *fiber.Ctx, field, message string) error {
	return respondError(c, model.FieldError(field, message))
}

// invalidRequest dipakai untuk body yang tidak bisa di-parse (JSON rusak).
// Spec tidak memakai 400, jadi dikirim sebagai 422.
func invalidRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse{
		Success: false,
		Message: message,
	})
}

// ErrorHandler dipasang di fiber.Config supaya error bawaan Fiber
// (route tidak ada, panic yang di-recover, dsb.) juga berformat JSON seragam
// dan tidak membocorkan detail di status 5xx.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		msg := fe.Message
		if fe.Code >= fiber.StatusInternalServerError {
			log.Printf("internal error: %v", err)
			msg = "Terjadi kesalahan pada server"
		}

		return c.Status(fe.Code).JSON(errorResponse{
			Success: false,
			Message: msg,
		})
	}

	return respondError(c, err)
}

func parseID(c *fiber.Ctx, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func currentUserID(c *fiber.Ctx) int64 {
	id, _ := c.Locals(middleware.CtxUserID).(int64)
	return id
}

func isAdmin(c *fiber.Ctx) bool {
	role, _ := c.Locals(middleware.CtxRole).(string)
	return role == model.RoleAdmin
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}