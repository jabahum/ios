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
	"case/internal/routes"
	"case/internal/services"
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

	// Set up routes
	routes.SetRoute(app, db, store, mlogger, config, smsService)

	// VHF CIF routes
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, mlogger, store, config, smsService)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, mlogger, store, config)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_investigator")
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
	connStr := "host=localhost user=" + config.Ux + " password='" + config.Px + "' dbname=" + config.Dx + " sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Cannot reach db: ", err.Error())
		sl.Error("Failed to connect to database: " + err.Error())
	}

	if err = db.Ping(); err != nil {
		fmt.Println("Cannot reach db: ", err.Error())
		sl.Error("Cannot reach db: " + err.Error())
	}

	return db
}

// get config details
func getConfig() (config handlers.Config) {
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err.Error())
		return config
	}

	// Decode the JSON data into a Config struct
	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return config
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
