package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"case/internal/config"
	flogger "case/internal/log"
	"case/internal/routes"
	"case/internal/services"

	dbpkg "case/internal/database"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"
)

func main() {
	// Load application config
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	appLogger := flogger.InitLogger(cfg.LogFile)

	// Initialize database connection
	primaryDB, err := dbpkg.InitDB(context.Background(), appLogger, dbpkg.DBConfig{
		Driver:          "postgres",
		DSN:             cfg.DBSource(),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		WaitTimeout:     30 * time.Second,
	})
	if err != nil {
		appLogger.Error("failed to initialize primary DB", "error", err)
	}
	defer primaryDB.Close()

	// Initialize logger
	sl := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Initialize SMS service
	smsService := services.NewSMSService(services.SMSConfig{
		BaseURL:  cfg.SMSBaseURL,
		Username: cfg.SMSUsername,
		Password: cfg.SMSPassword,
	})

	// Initialize Voice service
	voiceService := services.NewVoiceService(services.VoiceConfig{
		VoiceURL: cfg.VoiceURL,
	})

	// Initialize session store
	store := session.New()

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(fiberLogger.New())

	// Set up routes
	routes.SetRoute(app, primaryDB, store, sl, cfg, smsService, voiceService)

	// Start server
	log.Fatal(app.Listen(cfg.Addr()))
}
