package handler

import (
	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ImportHandler struct {
	service *service.ImportService
}

func NewImportHandler(s *service.ImportService) *ImportHandler {
	return &ImportHandler{service: s}
}

func (h *ImportHandler) ImportPendaftaran(c *fiber.Ctx) error {
	var payloads []domain.SiswaImportPayload

	if err := c.BodyParser(&payloads); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "format JSON tidak valid",
			"detail": err.Error(),
		})
	}

	if len(payloads) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "payload kosong"})
	}

	mode := c.Query("mode", "upsert") // default upsert
	if mode != "insert" && mode != "upsert" {
		return c.Status(400).JSON(fiber.Map{"error": "mode harus insert atau upsert"})
	}

	jobID := uuid.New().String()

	// langsung proses sync kalau payload kecil, atau taruh ke queue kalau besar
	// untuk versi awal, proses sync dulu supaya simpel
	result, err := h.service.ProcessImport(jobID, "pendaftaran", mode, payloads)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal memproses import", "detail": err.Error()})
	}

	return c.Status(202).JSON(fiber.Map{
		"job_id": jobID,
		"status": "completed", // karena sync, langsung completed
		"result": result,
	})
}

func (h *ImportHandler) ImportEmis(c *fiber.Ctx) error {
	// sama seperti ImportPendaftaran tapi source = "emis_scraping"
	var payloads []domain.SiswaImportPayload

	if err := c.BodyParser(&payloads); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format JSON tidak valid"})
	}

	mode := c.Query("mode", "upsert")
	jobID := uuid.New().String()

	result, err := h.service.ProcessImport(jobID, "emis_scraping", mode, payloads)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal memproses import"})
	}

	return c.Status(202).JSON(fiber.Map{
		"job_id": jobID,
		"status": "completed",
		"result": result,
	})
}

func (h *ImportHandler) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	status, err := h.service.GetJobStatus(jobID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
	}
	return c.JSON(status)
}
