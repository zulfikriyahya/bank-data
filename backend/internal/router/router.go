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

	// === consumer routes (dipakai app-presensi, app-perpustakaan) ===
	consumer := api.Group("/consumer", middleware.APIKeyAuth(db))
	consumer.Get("/siswa/:id/ringkas", siswaHandler.GetRingkas)
	consumer.Get("/siswa/search", siswaHandler.Search)
	consumer.Get("/siswa/:id/ringkas", middleware.RequireScope("siswa:read:basic"), siswaHandler.GetRingkas)
	consumer.Get("/siswa/search", middleware.RequireScope("siswa:read:basic"), siswaHandler.Search)

	// === admin routes (dipakai dashboard MDM, butuh session login) ===
	admin := api.Group("/admin", middleware.AdminAuth())
	admin.Get("/siswa", siswaHandler.List)
	admin.Get("/siswa/:id", siswaHandler.GetDetail)
	admin.Put("/siswa/:id", siswaHandler.Update)
	admin.Delete("/siswa/:id", siswaHandler.Delete)

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
