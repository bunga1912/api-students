package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// LocalsAuthUser adalah kunci penyimpanan identitas pengguna di dalam
// context request. Dibuat sebagai konstanta agar tidak ada salah ketik
// antara tempat menyimpan (middleware) dan tempat membaca (service).
const LocalsAuthUser = "authUser"

// requestTimeout adalah batas waktu seluruh pekerjaan database untuk satu request.
const requestTimeout = 5 * time.Second

// CurrentUser membaca identitas pengguna yang sudah diverifikasi
// oleh middleware RequireAuth.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}

func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}

// RequestContext membuat context dengan batas waktu untuk satu request.
// Pemanggil WAJIB memanggil cancel (defer cancel()).
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), requestTimeout)
}