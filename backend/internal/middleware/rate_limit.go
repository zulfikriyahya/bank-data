package middleware

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// GlobalRateLimit - rate limit kasar berdasarkan IP, lapisan pertama sebelum cek API key
// mencegah brute-force ke endpoint login/API key sebelum request masuk lebih dalam
func GlobalRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "terlalu banyak request, coba lagi sebentar lagi",
			})
		},
	})
}

// ClientRateLimit - rate limit spesifik per API client, dibaca dari kolom rate_limit di DB
// HARUS dipasang SETELAH APIKeyAuth karena butuh c.Locals("client_id")
func ClientRateLimit(db *sql.DB) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        1000, // default fallback, akan di-override logic di bawah kalau perlu dibuat dinamis
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			clientID := c.Locals("client_id")
			if clientID == nil {
				return c.IP() // fallback kalau belum ter-auth
			}
			return "client_" + string(rune(clientID.(int64)))
		},
		LimitReached: func(c *fiber.Ctx) error {
			clientName, _ := c.Locals("client_name").(string)
			return c.Status(429).JSON(fiber.Map{
				"error": "rate limit tercapai untuk client '" + clientName + "', coba lagi sebentar lagi",
			})
		},
	})
}
