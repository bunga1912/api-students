package middleware

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type attempt struct {
	count   int
	resetAt time.Time
}

// attemptLimiter membatasi jumlah percobaan per key dalam satu jendela waktu.
// State disimpan di memori proses (cukup untuk satu instance).
type attemptLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	now     func() time.Time
	entries map[string]*attempt
}

func newAttemptLimiter(max int, window time.Duration, now func() time.Time) *attemptLimiter {
	return &attemptLimiter{
		max:     max,
		window:  window,
		now:     now,
		entries: make(map[string]*attempt),
	}
}

// acquire memesan satu percobaan. Jika batas sudah tercapai, dikembalikan
// false beserta sisa waktu tunggu.
func (l *attemptLimiter) acquire(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()

	// Bersihkan entri kedaluwarsa agar map tidak tumbuh tanpa batas.
	if len(l.entries) > 1024 {
		for k, e := range l.entries {
			if !now.Before(e.resetAt) {
				delete(l.entries, k)
			}
		}
	}

	e, ok := l.entries[key]
	if ok && !now.Before(e.resetAt) {
		delete(l.entries, key)
		ok = false
	}

	if !ok {
		l.entries[key] = &attempt{count: 1, resetAt: now.Add(l.window)}
		return true, 0
	}

	if e.count >= l.max {
		return false, e.resetAt.Sub(now)
	}

	e.count++

	return true, 0
}

// refund mengembalikan satu percobaan (untuk request yang bukan login gagal).
func (l *attemptLimiter) refund(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		return
	}

	e.count--
	if e.count <= 0 {
		delete(l.entries, key)
	}
}

// reset menghapus hitungan (dipakai saat login sukses).
func (l *attemptLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.entries, key)
}

// LoginRateLimit membatasi login gagal: setelah max kali gagal (401) dalam
// satu window, percobaan berikutnya dari IP yang sama dijawab 429.
func LoginRateLimit(max int, window time.Duration) fiber.Handler {
	return loginRateLimit(newAttemptLimiter(max, window, time.Now))
}

func loginRateLimit(l *attemptLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()

		allowed, retryAfter := l.acquire(key)
		if !allowed {
			secs := int(math.Ceil(retryAfter.Seconds()))
			if secs < 1 {
				secs = 1
			}

			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(secs))

			return abort(
				c,
				fiber.StatusTooManyRequests,
				fmt.Sprintf("Terlalu banyak percobaan login. Coba lagi dalam %d detik", secs),
			)
		}

		if err := c.Next(); err != nil {
			l.refund(key)
			return err
		}

		switch c.Response().StatusCode() {
		case fiber.StatusUnauthorized:
			// login gagal: percobaan tetap dihitung
		case fiber.StatusOK:
			l.reset(key)
		default:
			// mis. 422 validasi: bukan kegagalan kredensial
			l.refund(key)
		}

		return nil
	}
}