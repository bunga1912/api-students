package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"api-students/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)

			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("access token kedaluwarsa")
			}
			return helper.Unauthorized("access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

// RequirePermission menolak request bila role pemanggil tidak memiliki
// permission yang diminta. Dipasang SETELAH RequireAuth, sehingga
// identitas pengguna sudah tersedia di Locals.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}

		if !perms.Can(authUser.Role, permission) {
			return helper.Forbidden("anda tidak memiliki hak untuk mengakses sumber daya ini")
		}

		return c.Next()
	}
}

// RequireRole menolak request bila role pemanggil bukan salah satu
// dari yang diizinkan. Dipakai untuk endpoint yang haknya ditentukan
// langsung oleh role, bukan oleh permission granular.
func RequireRole(allowed ...string) fiber.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		authUser, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}

		if _, ok := allowedSet[authUser.Role]; !ok {
			return helper.Forbidden("anda tidak memiliki hak untuk mengakses sumber daya ini")
		}

		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}

	return token, nil
}

// LoginRateLimiter membatasi jumlah percobaan login dari satu alamat IP.
// Tanpa pembatasan ini, penyerang dapat mencoba ribuan password per menit
// (serangan brute force) tanpa hambatan apa pun.
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "60")
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam satu menit")
		},
	})
}