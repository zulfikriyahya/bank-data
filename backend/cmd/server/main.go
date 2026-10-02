package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bank-data/backend/internal/config"
	"bank-data/backend/internal/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	app := fiber.New(fiber.Config{
		AppName:      "bank-data-core-app",
		BodyLimit:    10 * 1024 * 1024,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key",
		AllowCredentials: true,
	}))

	router.Setup(app, db, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// jalankan server di goroutine terpisah supaya main() bisa dengar sinyal shutdown
	go func() {
		log.Printf("core-app berjalan di port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("server berhenti: %v", err)
		}
	}()

	// tunggu sinyal interrupt/terminate (Ctrl+C, atau PM2 restart/stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("menerima sinyal shutdown, menghentikan server dengan aman...")

	// beri waktu 10 detik untuk request yang sedang berjalan selesai
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("error saat shutdown: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("error menutup koneksi database: %v", err)
	}

	log.Println("server berhasil dimatikan dengan aman")
}

func getAllowedOrigins() string {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origins == "" {
		return "http://localhost:3000"
	}
	return origins
}
