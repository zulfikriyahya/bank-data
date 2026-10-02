package handler

import (
	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ScraperConfigHandler struct {
	service *service.ScraperConfigService
}

func NewScraperConfigHandler(s *service.ScraperConfigService) *ScraperConfigHandler {
	return &ScraperConfigHandler{service: s}
}

// SetConfig - admin simpan/update config (dienkripsi otomatis di service)
func (h *ScraperConfigHandler) SetConfig(c *fiber.Ctx) error {
	scraperName := c.Params("name")

	var input domain.SetScraperConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	adminUsername, _ := c.Locals("admin_username").(string)

	if err := h.service.Set(scraperName, input, adminUsername); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "config berhasil disimpan"})
}

// GetConfig - admin lihat config (password disensor)
func (h *ScraperConfigHandler) GetConfig(c *fiber.Ctx) error {
	scraperName := c.Params("name")

	config, err := h.service.GetMasked(scraperName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(config)
}

func (h *ScraperConfigHandler) List(c *fiber.Ctx) error {
	configs, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": configs})
}

// GetConfigForScraper - DIPANGGIL SCRAPER SENDIRI (internal, service token), full tanpa masking
func (h *ScraperConfigHandler) GetConfigForScraper(c *fiber.Ctx) error {
	scraperName := c.Params("name")

	config, err := h.service.GetFull(scraperName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"config": config})
}
