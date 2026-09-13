package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// LocalsAuthUser adalah kunci penyimpanan identitas pengguna di dalam
// context request. Dibuat sebagai konstanta agar tidak ada salah ketik
// antara tempat menyimpan (middleware) dan tempat membaca (service).
const LocalsAuthUser = "authUser"

// CurrentUser membaca identitas pengguna yang sudah diverifikasi
// oleh middleware RequireAuth.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}