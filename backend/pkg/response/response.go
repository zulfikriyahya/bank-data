package response

import "github.com/gofiber/fiber/v2"

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{"data": data})
}
