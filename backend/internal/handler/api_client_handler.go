package handler

import (
	"strconv"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type APIClientHandler struct {
	service *service.APIClientService
}

func NewAPIClientHandler(s *service.APIClientService) *APIClientHandler {
	return &APIClientHandler{service: s}
}

func (h *APIClientHandler) Create(c *fiber.Ctx) error {
	var input domain.CreateAPIClientInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	apiKey, client, err := h.service.Create(input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"client":  client,
		"api_key": apiKey,
		"warning": "simpan api_key ini sekarang, tidak akan ditampilkan lagi setelah ini",
	})
}

func (h *APIClientHandler) List(c *fiber.Ctx) error {
	clients, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": clients})
}

func (h *APIClientHandler) GetDetail(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	client, err := h.service.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(client)
}

func (h *APIClientHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	var input domain.UpdateAPIClientInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	if err := h.service.Update(id, input); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "client berhasil diperbarui"})
}

func (h *APIClientHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	if err := h.service.Delete(id); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "client berhasil dihapus"})
}

func (h *APIClientHandler) RegenerateKey(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	apiKey, err := h.service.RegenerateKey(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"api_key": apiKey,
		"warning": "key lama sudah tidak berlaku, simpan key baru ini sekarang",
	})
}
