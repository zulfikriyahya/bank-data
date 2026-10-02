package main

import (
	"log"
	"os"

	"bank-data/backend/internal/config"
	"bank-data/backend/internal/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	defer db.Close()

	app := fiber.New(fiber.Config{
		AppName: "bank-data-core-app",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	router.Setup(app, db, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("core-app berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
