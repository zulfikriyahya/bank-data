package handler

import (
	"fmt"
	"time"

	"bank-data/backend/internal/repository"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ExportHandler struct {
	service *service.ExportService
}

func NewExportHandler(s *service.ExportService) *ExportHandler {
	return &ExportHandler{service: s}
}

func (h *ExportHandler) buildFilter(c *fiber.Ctx) repository.ExportFilter {
	return repository.ExportFilter{
		Search:   c.Query("search", ""),
		Source:   c.Query("source", ""),
		DateFrom: c.Query("date_from", ""),
		DateTo:   c.Query("date_to", ""),
	}
}

func (h *ExportHandler) ExportJSON(c *fiber.Ctx) error {
	filter := h.buildFilter(c)

	jsonBytes, err := h.service.ExportToJSON(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal export data", "detail": err.Error()})
	}

	filename := fmt.Sprintf("siswa_export_%s.json", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Type", "application/json")

	return c.Send(jsonBytes)
}

func (h *ExportHandler) ExportExcel(c *fiber.Ctx) error {
	filter := h.buildFilter(c)

	file, err := h.service.ExportToExcel(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal export data", "detail": err.Error()})
	}
	defer file.Close()

	filename := fmt.Sprintf("siswa_export_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	buf, err := file.WriteToBuffer()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal membuat file excel"})
	}

	return c.Send(buf.Bytes())
}
