package router

import (
	"database/sql"

	"bank-data/backend/internal/config"
	"bank-data/backend/internal/handler"
	"bank-data/backend/internal/middleware"
	mysqlrepo "bank-data/backend/internal/repository/mysql"
	"bank-data/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, db *sql.DB, cfg *config.Config) {
	// rate limit global - lapisan pertama untuk semua request
	app.Use(middleware.GlobalRateLimit())
	// repository
	siswaRepo := mysqlrepo.NewSiswaRepo(db)
	importLogRepo := mysqlrepo.NewImportLogRepo(db)

	// service
	siswaService := service.NewSiswaService(siswaRepo)
	importService := service.NewImportService(siswaRepo, importLogRepo)

	// handler
	siswaHandler := handler.NewSiswaHandler(siswaService)
	importHandler := handler.NewImportHandler(importService)

	api := app.Group("/api/v1")

	// === internal routes (dipakai background workers) ===
	internal := api.Group("/internal", middleware.ServiceAuth(db))
	internal.Post("/import/pendaftaran", importHandler.ImportPendaftaran)
	internal.Post("/import/emis", importHandler.ImportEmis)
	internal.Get("/import/:jobId/status", importHandler.GetJobStatus)

	// tambahkan di bagian repository/service/handler setup
	activityLogRepo := mysqlrepo.NewActivityLogRepo(db)
	activityLogService := service.NewActivityLogService(activityLogRepo, siswaRepo)
	activityLogHandler := handler.NewActivityLogHandler(activityLogService)

	// === consumer routes (dipakai app-presensi, app-perpustakaan) ===
	consumer := api.Group("/consumer", middleware.APIKeyAuth(db))
	consumer.Get("/siswa/:id/ringkas", middleware.RequireScope("siswa:read:basic"), siswaHandler.GetRingkas)
	consumer.Get("/siswa/search", middleware.RequireScope("siswa:read:basic"), siswaHandler.Search)
	consumer.Post("/activity-log",
		middleware.RequireAnyScope("presensi:write", "perpustakaan:write"),
		activityLogHandler.Record,
	)

	apiClientRepo := mysqlrepo.NewAPIClientRepo(db)
	apiClientService := service.NewAPIClientService(apiClientRepo)
	apiClientHandler := handler.NewAPIClientHandler(apiClientService)

	scraperConfigRepo := mysqlrepo.NewScraperConfigRepo(db)
	scraperConfigService := service.NewScraperConfigService(scraperConfigRepo)
	scraperConfigHandler := handler.NewScraperConfigHandler(scraperConfigService)

	// === admin routes (dipakai dashboard MDM, butuh session login) ===
	admin := api.Group("/admin", middleware.AdminAuth())
	admin.Get("/siswa", siswaHandler.List)
	admin.Get("/siswa/:id", siswaHandler.GetDetail)
	admin.Put("/siswa/:id", siswaHandler.Update)
	admin.Delete("/siswa/:id", siswaHandler.Delete)
	admin.Get("/siswa/:id/activity-log", activityLogHandler.GetBySiswa)
	admin.Get("/activity-log", activityLogHandler.List)
	admin.Post("/api-clients", apiClientHandler.Create)
	admin.Get("/api-clients", apiClientHandler.List)
	admin.Get("/api-clients/:id", apiClientHandler.GetDetail)
	admin.Put("/api-clients/:id", apiClientHandler.Update)
	admin.Delete("/api-clients/:id", apiClientHandler.Delete)
	admin.Post("/api-clients/:id/regenerate-key", apiClientHandler.RegenerateKey)

	// admin - kelola config
	admin.Post("/scraper-config/:name", scraperConfigHandler.SetConfig)
	admin.Get("/scraper-config/:name", scraperConfigHandler.GetConfig)
	admin.Get("/scraper-config", scraperConfigHandler.List)

	// internal - scraper ambil config sendiri
	internal.Get("/scraper-config/:name", scraperConfigHandler.GetConfigForScraper)
	adminRepo := mysqlrepo.NewAdminRepo(db)

	authService := service.NewAuthService(adminRepo)
	authHandler := handler.NewAuthHandler(authService)

	exportService := service.NewExportService(siswaRepo)
	exportHandler := handler.NewExportHandler(exportService)

	// export routes - pakai AdminAuth supaya tidak sembarang orang bisa download data sensitif
	export := api.Group("/export", middleware.AdminAuth())
	export.Get("/json", exportHandler.ExportJSON)
	export.Get("/xlsx", exportHandler.ExportExcel)

	api.Post("/auth/login", authHandler.Login) // di luar group admin, tidak butuh token

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
