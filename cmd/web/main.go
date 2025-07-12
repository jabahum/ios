package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"

	_ "github.com/lib/pq"

	"case/internal/handlers"
	"case/internal/models"
	"case/internal/routes"
	"case/internal/services"
	"context"
)

var store = session.New() // Session store

func trace() string {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return "?"
	}
	fn := runtime.FuncForPC(pc)
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}

func main() {
	config := getConfig()

	mlogger := initLogger(config.LogFile)

	// Initialize Fiber app
	app := fiber.New()

	// Serve static files
	app.Static("/static", "../../ui/static")

	// Add Logger middleware
	app.Use(logger.New())

	db := getDB(config, mlogger)
	defer db.Close()

	// Initialize SMS service
	smsConfig := services.SMSConfig{
		BaseURL:  config.SMSBaseURL,
		Username: config.SMSUsername,
		Password: config.SMSPassword,
	}

	// Log SMS configuration
	mlogger.Info("Initializing SMS service",
		"base_url", smsConfig.BaseURL,
		"username", smsConfig.Username)

	smsService := services.NewSMSService(smsConfig)

	// TEMPORARY: Reset philip user password
	app.Get("/reset-philip", func(c *fiber.Ctx) error {
		// Reset philip user password to "123456"
		_, err := db.ExecContext(context.Background(),
			"UPDATE users SET user_pass = $1 WHERE user_name = $2",
			models.Encrypt("123456"), "philip")

		if err != nil {
			return c.SendString("Error resetting password: " + err.Error())
		}

		// Check if user exists
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM users WHERE user_name = $1", "philip").Scan(&count)

		if err != nil {
			return c.SendString("Error checking user: " + err.Error())
		}

		if count == 0 {
			return c.SendString("User 'philip' not found!")
		}

		return c.SendString("Password for user 'philip' has been reset to '123456'. You can now log in.")
	})

	// Set up routes
	routes.SetRoute(app, db, store, mlogger, config, smsService)

	// VHF CIF routes
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, mlogger, store, config, smsService)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, mlogger, store, config, smsService)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_investigator")
	})
	app.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFSuccess(c, db, mlogger, store, config)
	})

	mlogger.Info("starting server...")
	// Start the app

	app.Listen(config.Address)
}

// connect to database
func getDB(config handlers.Config, sl *slog.Logger) *sql.DB {
	// Use proper PostgreSQL connection string format
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable",
		config.Ux, config.Px, config.Dx)

	sl.Info("Connecting to database", "host", "localhost", "user", config.Ux, "database", config.Dx)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		sl.Error("Failed to open database connection", "error", err.Error())
		panic(fmt.Sprintf("Cannot open database connection: %v", err))
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		sl.Error("Failed to ping database", "error", err.Error())
		panic(fmt.Sprintf("Cannot reach database: %v", err))
	}

	sl.Info("Successfully connected to database")
	return db
}

// get config details
func getConfig() (config handlers.Config) {
	// Try multiple possible config file locations
	configPaths := []string{
		"config.json",
		"cmd/web/config.json",
		"../../cmd/web/config.json",
	}

	var configFile *os.File
	var err error

	for _, path := range configPaths {
		configFile, err = os.Open(path)
		if err == nil {
			fmt.Printf("Found config file: %s\n", path)
			break
		}
	}

	if configFile == nil {
		panic(fmt.Sprintf("Could not find config.json in any of these locations: %v", configPaths))
	}
	defer configFile.Close()

	// Decode the JSON data into a Config struct
	decoder := json.NewDecoder(configFile)

	if err := decoder.Decode(&config); err != nil {
		panic(fmt.Sprintf("Error decoding config JSON: %v", err))
	}

	return config
}

func initLogger(logFile string) *slog.Logger {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a handler for writing logs to the file
	fileHandler := slog.NewTextHandler(file, nil)
	logger := slog.New(fileHandler)

	// Set this logger as the default
	slog.SetDefault(logger)

	return logger
}
