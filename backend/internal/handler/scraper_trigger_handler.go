package handler

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"bank-data/backend/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type ScraperTriggerHandler struct {
	scraperDir string
	mu         sync.Mutex
	running    bool
	startedAt  time.Time
}

func NewScraperTriggerHandler(scraperDir string) *ScraperTriggerHandler {
	return &ScraperTriggerHandler{scraperDir: scraperDir}
}

func mapSameSite(s string) string {
	switch s {
	case "no_restriction":
		return "None"
	case "strict":
		return "Strict"
	default:
		return "Lax"
	}
}

func buildStorageState(input domain.ScraperAuthInput) domain.PlaywrightStorageState {
	cookies := make([]domain.PlaywrightCookie, 0, len(input.Cookies))
	for _, c := range input.Cookies {
		expires := c.ExpirationDate
		if expires == 0 {
			expires = -1
		}
		cookies = append(cookies, domain.PlaywrightCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  expires,
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
			SameSite: mapSameSite(c.SameSite),
		})
	}

	return domain.PlaywrightStorageState{
		Cookies: cookies,
		Origins: []domain.PlaywrightOrigin{
			{Origin: "https://emis.kemenag.go.id", LocalStorage: input.LocalStorage},
		},
	}
}

// TriggerJSON - terima cookies + localStorage langsung dari extension, susun auth.json, jalankan scraper
func (h *ScraperTriggerHandler) TriggerJSON(c *fiber.Ctx) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return c.Status(409).JSON(fiber.Map{"error": "scraping sedang berjalan, tunggu sampai selesai"})
	}
	h.mu.Unlock()

	var input domain.ScraperAuthInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format JSON tidak valid"})
	}
	if len(input.Cookies) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "cookies kosong, pastikan sudah login ke EMIS"})
	}
	if len(input.LocalStorage) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "localStorage kosong, pastikan sudah login ke EMIS"})
	}

	storageState := buildStorageState(input)
	authBytes, err := json.MarshalIndent(storageState, "", "  ")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal menyusun auth.json"})
	}

	authPath := filepath.Join(h.scraperDir, "auth.json")
	if err := os.WriteFile(authPath, authBytes, 0600); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal menyimpan auth.json"})
	}

	mode := c.Query("mode", "resume")
	logPath := filepath.Join(h.scraperDir, "scraping.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal membuat file log"})
	}

	cmd := exec.Command(filepath.Join(h.scraperDir, "run_scraper.sh"), mode)
	cmd.Dir = h.scraperDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return c.Status(500).JSON(fiber.Map{"error": "gagal menjalankan scraper: " + err.Error()})
	}

	h.mu.Lock()
	h.running = true
	h.startedAt = time.Now()
	h.mu.Unlock()

	go func() {
		cmd.Wait()
		logFile.Close()
		h.mu.Lock()
		h.running = false
		h.mu.Unlock()
	}()

	return c.Status(202).JSON(fiber.Map{"message": "scraping dimulai di background", "mode": mode})
}

func (h *ScraperTriggerHandler) Status(c *fiber.Ctx) error {
	h.mu.Lock()
	running := h.running
	startedAt := h.startedAt
	h.mu.Unlock()

	lines := tailFile(filepath.Join(h.scraperDir, "scraping.log"), 80)

	return c.JSON(fiber.Map{"running": running, "started_at": startedAt, "log_tail": lines})
}

func tailFile(path string, n int) []string {
	file, err := os.Open(path)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return lines
}
