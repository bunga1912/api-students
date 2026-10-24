package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestAttemptLimiter(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	l := newAttemptLimiter(5, time.Minute, func() time.Time { return now })

	for i := 1; i <= 5; i++ {
		if ok, _ := l.acquire("ip-a"); !ok {
			t.Fatalf("percobaan ke-%d seharusnya diizinkan", i)
		}
	}

	ok, wait := l.acquire("ip-a")
	if ok {
		t.Fatal("percobaan ke-6 seharusnya ditolak")
	}
	if wait <= 0 || wait > time.Minute {
		t.Fatalf("sisa waktu tunggu tidak masuk akal: %v", wait)
	}

	if ok, _ := l.acquire("ip-b"); !ok {
		t.Fatal("IP lain tidak boleh ikut terblokir")
	}

	now = now.Add(time.Minute)

	if ok, _ := l.acquire("ip-a"); !ok {
		t.Fatal("setelah jendela waktu lewat, percobaan harus diizinkan lagi")
	}
}

func TestAttemptLimiterRefundAndReset(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	l := newAttemptLimiter(2, time.Minute, func() time.Time { return now })

	l.acquire("ip")
	l.refund("ip")
	l.acquire("ip")
	l.acquire("ip")

	if ok, _ := l.acquire("ip"); ok {
		t.Fatal("seharusnya terblokir setelah 2 percobaan tercatat")
	}

	l.reset("ip")

	if ok, _ := l.acquire("ip"); !ok {
		t.Fatal("setelah reset, percobaan harus diizinkan")
	}
}

func TestLoginRateLimit(t *testing.T) {
	app := fiber.New()
	app.Post("/login", LoginRateLimit(5, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusUnauthorized)
	})

	for i := 1; i <= 5; i++ {
		resp, err := app.Test(httptest.NewRequest("POST", "/login", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("percobaan ke-%d: status = %d, want 401", i, resp.StatusCode)
		}
	}

	resp, err := app.Test(httptest.NewRequest("POST", "/login", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("percobaan ke-6: status = %d, want 429", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Fatal("header Retry-After tidak ada")
	}
}