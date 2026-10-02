package handler

import (
	"strconv"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ActivityLogHandler struct {
	service *service.ActivityLogService
}

func NewActivityLogHandler(s *service.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{service: s}
}

// Record - dipanggil consumer app (presensi/perpustakaan) untuk catat aktivitas
func (h *ActivityLogHandler) Record(c *fiber.Ctx) error {
	clientID, ok := c.Locals("client_id").(int64)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "client_id tidak ditemukan di context"})
	}

	var input domain.ActivityLogInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	if input.SiswaID == 0 || input.ActivityType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "siswa_id dan activity_type wajib diisi"})
	}

	id, err := h.service.Record(clientID, input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":      id,
		"message": "aktivitas berhasil dicatat",
	})
}

// GetBySiswa - admin lihat riwayat aktivitas seorang siswa
func (h *ActivityLogHandler) GetBySiswa(c *fiber.Ctx) error {
	siswaID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	logs, err := h.service.GetBySiswa(siswaID, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": logs})
}

// List - admin lihat semua aktivitas, bisa difilter per jenis
func (h *ActivityLogHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	activityType := c.Query("type", "")

	logs, total, err := h.service.List(page, perPage, activityType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":     logs,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}
