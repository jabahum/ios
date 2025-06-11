package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"case/internal/handlers"
	"case/internal/routes"
	"case/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize logger
	sl := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Initialize SMS service
	smsService := services.NewSMSService(services.SMSConfig{
		BaseURL:  os.Getenv("SMS_BASE_URL"),
		Username: os.Getenv("SMS_USERNAME"),
		Password: os.Getenv("SMS_PASSWORD"),
	})

	// Initialize session store
	store := session.New()

	// Initialize config
	config := handlers.Config{
		// Add any required config values here
	}

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(logger.New())

	// Set up routes
	routes.SetRoute(app, db, store, sl, config, smsService)

	// Start server
	log.Fatal(app.Listen(":3000"))
}
