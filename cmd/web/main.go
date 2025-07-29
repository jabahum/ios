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
	"github.com/gofiber/storage/redis"

	_ "github.com/lib/pq"

	"case/internal/handlers"
	"case/internal/models"
	"case/internal/routes"
	"case/internal/services"
	"context"
)

// Initialize Redis storage for sessions
var store *session.Store

func init() {
	// Initialize session store with in-memory storage as fallback
	store = session.New(session.Config{
		Expiration: 24 * 60 * 60, // 24 hours in seconds
		KeyLookup:  "cookie:fiber_sess",
	})
}

// checkRedisConnection checks if Redis is available and returns connection status
func checkRedisConnection(config handlers.Config) (bool, string) {
	if !config.RedisEnabled {
		return false, "Redis is disabled in configuration"
	}

	// Set default Redis configuration if not provided
	if config.RedisHost == "" {
		config.RedisHost = "localhost"
	}
	if config.RedisPort == 0 {
		config.RedisPort = 6379
	}

	// Create Redis storage for testing
	redisStorage := redis.New(redis.Config{
		Host:      config.RedisHost,
		Port:      config.RedisPort,
		Username:  "",
		Password:  config.RedisPassword,
		Database:  config.RedisDatabase,
		Reset:     false,
		TLSConfig: nil,
	})

	// Test Redis connection
	if err := redisStorage.Conn().Ping(context.Background()).Err(); err != nil {
		return false, fmt.Sprintf("Redis connection failed: %v", err)
	}

	return true, "Redis connection successful"
}

// initializeRedisSession initializes Redis session storage
func initializeRedisSession(config handlers.Config, logger *slog.Logger) *session.Store {
	if !config.RedisEnabled {
		logger.Info("Redis is disabled, using in-memory session storage")
		return session.New(session.Config{
			Expiration: 24 * 60 * 60, // 24 hours in seconds
			KeyLookup:  "cookie:fiber_sess",
		})
	}

	// Set default Redis configuration if not provided
	if config.RedisHost == "" {
		config.RedisHost = "localhost"
	}
	if config.RedisPort == 0 {
		config.RedisPort = 6379
	}
	if config.RedisDatabase == 0 {
		config.RedisDatabase = 0
	}

	logger.Info("Initializing Redis session storage",
		"host", config.RedisHost,
		"port", config.RedisPort,
		"database", config.RedisDatabase)

	// Create Redis storage
	redisStorage := redis.New(redis.Config{
		Host:      config.RedisHost,
		Port:      config.RedisPort,
		Username:  "", // Redis username (if needed)
		Password:  config.RedisPassword,
		Database:  config.RedisDatabase,
		Reset:     false, // Do not flush DB on startup
		TLSConfig: nil,   // TLS config (if using TLS)
	})

	// Test Redis connection
	if err := redisStorage.Conn().Ping(context.Background()).Err(); err != nil {
		logger.Warn("Failed to connect to Redis, falling back to in-memory storage",
			"error", err.Error(),
			"host", config.RedisHost,
			"port", config.RedisPort)

		// Fallback to in-memory storage
		return session.New(session.Config{
			Expiration: 24 * 60 * 60, // 24 hours in seconds
			KeyLookup:  "cookie:fiber_sess",
		})
	}

	logger.Info("Successfully connected to Redis for session storage")

	// Initialize session store with Redis storage
	return session.New(session.Config{
		Storage:    redisStorage,
		Expiration: 24 * 60 * 60, // 24 hours in seconds
		KeyLookup:  "cookie:fiber_sess",
	})
}

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

	// Initialize Redis session storage
	store = initializeRedisSession(config, mlogger)

	// Initialize Fiber app
	app := fiber.New()

	// Debug middleware to log session cookie and session data for every request
	app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("fiber_sess")
		log.Printf("DEBUG: Incoming request session cookie: %s", cookie)
		sess, err := store.Get(c)
		if err != nil {
			log.Printf("DEBUG: Error getting session: %v", err)
		} else {
			log.Printf("DEBUG: Session keys: %v", sess.Keys())
			log.Printf("DEBUG: Session isAuthenticated: %v", sess.Get("isAuthenticated"))
			log.Printf("DEBUG: Session userID: %v", sess.Get("userID"))
			log.Printf("DEBUG: Session username: %v", sess.Get("username"))
		}
		return c.Next()
	})

	// Session health check endpoint
	app.Get("/session/health", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to get session",
				"error":   err.Error(),
			})
		}

		// Check Redis connection status
		redisConnected, redisMessage := checkRedisConnection(config)

		return c.JSON(fiber.Map{
			"status":         "success",
			"sessionID":      sess.ID(),
			"keys":           sess.Keys(),
			"redisEnabled":   config.RedisEnabled,
			"redisConnected": redisConnected,
			"redisMessage":   redisMessage,
			"storage": func() string {
				if config.RedisEnabled && redisConnected {
					return "redis"
				}
				return "memory"
			}(),
		})
	})

	// Session clear endpoint for debugging
	app.Get("/session/clear", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to get session",
				"error":   err.Error(),
			})
		}

		if err := sess.Destroy(); err != nil {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to destroy session",
				"error":   err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Session cleared",
		})
	})

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

	// Add test routes for session debugging
	app.Get("/test-session", func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Set("foo", "bar")
		sess.Save()
		return c.SendString("Session set: foo=bar")
	})

	app.Get("/check-session", func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		val := sess.Get("foo")
		return c.SendString(fmt.Sprintf("foo: %v", val))
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
