package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"case/internal/config"
	"case/internal/routes"
	"case/internal/services"

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

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DBSource())
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

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

	// Initialize handler config
	handlerConfig := config.Config{}

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(fiberLogger.New())

	// Set up routes
	routes.SetRoute(app, db, store, sl, handlerConfig, smsService, voiceService)

	// Start server
	log.Fatal(app.Listen(cfg.Addr()))
}
