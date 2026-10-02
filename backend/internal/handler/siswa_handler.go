package handler

import (
	"strconv"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type SiswaHandler struct {
	service *service.SiswaService
}

func NewSiswaHandler(s *service.SiswaService) *SiswaHandler {
	return &SiswaHandler{service: s}
}

func (h *SiswaHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")

	siswas, total, err := h.service.List(page, perPage, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":     siswas,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *SiswaHandler) GetDetail(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	siswa, err := h.service.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(siswa)
}

// GetRingkas - versi minim untuk konsumsi app eksternal (presensi, perpustakaan)
func (h *SiswaHandler) GetRingkas(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	siswa, err := h.service.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"id":             siswa.ID,
		"uuid":           siswa.UUID,
		"nama":           siswa.Nama,
		"nisn":           siswa.NISN,
		"nik":            siswa.NIK,
		"jenis_kelamin":  siswa.JenisKelamin,
		"rfid":           siswa.RFID,
		"fingerprint":    siswa.Fingerprint,
	})
}

func (h *SiswaHandler) Search(c *fiber.Ctx) error {
	nik := c.Query("nik", "")
	rfid := c.Query("rfid", "")
	nisn := c.Query("nisn", "")

	var field, value string
	switch {
	case rfid != "":
		field, value = "rfid", rfid
	case nik != "":
		field, value = "nik", nik
	case nisn != "":
		field, value = "nisn", nisn
	default:
		return c.Status(400).JSON(fiber.Map{"error": "sertakan parameter nik, rfid, atau nisn"})
	}

	siswa, err := h.service.FindByField(field, value)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	// sama seperti GetRingkas - batasi field yang diekspos ke consumer app
	return c.JSON(fiber.Map{
		"id":            siswa.ID,
		"uuid":          siswa.UUID,
		"nama":          siswa.Nama,
		"nisn":          siswa.NISN,
		"nik":           siswa.NIK,
		"jenis_kelamin": siswa.JenisKelamin,
		"rfid":          siswa.RFID,
		"fingerprint":   siswa.Fingerprint,
	})
}

func (h *SiswaHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	var siswa domain.Siswa
	if err := c.BodyParser(&siswa); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	if err := h.service.Update(id, &siswa); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "data berhasil diperbarui"})
}

func (h *SiswaHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id tidak valid"})
	}

	if err := h.service.Delete(id); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "data berhasil dihapus"})
}
