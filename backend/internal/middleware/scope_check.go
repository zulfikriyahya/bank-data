package middleware

import (
	"database/sql"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func APIKeyAuth(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(401).JSON(fiber.Map{"error": "header X-API-Key wajib diisi"})
		}

		rows, err := db.Query(`SELECT id, client_name, api_key_hash, scopes, is_active FROM api_clients WHERE is_active = TRUE`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "gagal validasi API key"})
		}
		defer rows.Close()

		var matchedClientID int64
		var matchedClientName string
		var matchedScopesRaw string
		found := false

		for rows.Next() {
			var id int64
			var clientName, hash, scopesRaw string
			var active bool
			if err := rows.Scan(&id, &clientName, &hash, &scopesRaw, &active); err != nil {
				continue
			}

			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(apiKey)) == nil {
				matchedClientID = id
				matchedClientName = clientName
				matchedScopesRaw = scopesRaw
				found = true
				break
			}
		}

		if !found {
			return c.Status(403).JSON(fiber.Map{"error": "API key tidak valid"})
		}

		var scopes []string
		if err := json.Unmarshal([]byte(matchedScopesRaw), &scopes); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "konfigurasi scope client rusak"})
		}

		c.Locals("client_id", matchedClientID)
		c.Locals("client_name", matchedClientName)
		c.Locals("client_scopes", scopes)

		return c.Next()
	}
}

// RequireScope - dipasang per-endpoint, cek apakah client (dari APIKeyAuth) punya scope yang diminta
// HARUS dipasang SETELAH APIKeyAuth di chain middleware, karena butuh c.Locals("client_scopes")
func RequireScope(requiredScope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopesRaw := c.Locals("client_scopes")
		if scopesRaw == nil {
			// harusnya tidak pernah terjadi kalau dipasang benar setelah APIKeyAuth
			return c.Status(500).JSON(fiber.Map{"error": "scope belum divalidasi, cek urutan middleware"})
		}

		scopes, ok := scopesRaw.([]string)
		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "format scope tidak valid"})
		}

		for _, s := range scopes {
			if s == requiredScope {
				return c.Next()
			}
		}

		clientName, _ := c.Locals("client_name").(string)
		return c.Status(403).JSON(fiber.Map{
			"error": "client '" + clientName + "' tidak punya izin untuk endpoint ini",
			"required_scope": requiredScope,
		})
	}
}
