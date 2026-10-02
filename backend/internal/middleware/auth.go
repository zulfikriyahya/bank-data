package middleware

import (
	"database/sql"
	"os"
	"strings"

	"bank-data/backend/pkg/jwtutil"

	"github.com/gofiber/fiber/v2"
)

func ServiceAuth(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "header Authorization wajib diisi"})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		validTokens := []string{
			os.Getenv("SERVICE_TOKEN_PENDAFTARAN"),
			os.Getenv("SERVICE_TOKEN_EMIS"),
		}

		for _, valid := range validTokens {
			if valid != "" && token == valid {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "token tidak valid"})
	}
}

// AdminAuth - validasi JWT sungguhan, dipakai endpoint /admin/*
func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "header Authorization wajib diisi"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtutil.ValidateToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}

		c.Locals("admin_id", claims.UserID)
		c.Locals("admin_username", claims.Username)
		c.Locals("admin_role", claims.Role)

		return c.Next()
	}
}
