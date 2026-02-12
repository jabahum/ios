package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"

	_ "github.com/lib/pq"
        _ "github.com/joho/godotenv/autoload"

	_ "case/cmd/web/docs" // Swagger documentation
	"case/internal/handlers"
	"case/internal/models"
	"case/internal/routes"
	"case/internal/services"
)

// Initialize session storage
var store *session.Store

func init() {
	// Initialize session store with in-memory storage
	store = session.New(session.Config{
		Expiration:     24 * 60 * 60, // 24 hours in seconds
		KeyLookup:      "cookie:fiber_sess",
		CookieSecure:   false, // Set to true in production with HTTPS
		CookieHTTPOnly: true,
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSameSite: "Lax",
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
	// database.ConnectDb()
	config := getConfig()

	mlogger := initLogger(config.LogFile)

	// Initialize Fiber app
	app := fiber.New()

	// Debug middleware to log session cookie and session data for every request
	app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("fiber_sess")
		log.Printf("DEBUG: Incoming request session cookie: %s", cookie)
		log.Printf("DEBUG: Request path: %s", c.Path())
		sess, err := store.Get(c)
		if err != nil {
			log.Printf("DEBUG: Error getting session: %v", err)
		} else {
			log.Printf("DEBUG: Session keys: %v", sess.Keys())
			log.Printf("DEBUG: Session isAuthenticated: %v", sess.Get("isAuthenticated"))
			log.Printf("DEBUG: Session user: %v", sess.Get("user"))
			log.Printf("DEBUG: Session user_id: %v", sess.Get("user_id"))
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

		return c.JSON(fiber.Map{
			"status":    "success",
			"sessionID": sess.ID(),
			"keys":      sess.Keys(),
			"storage":   "memory",
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

	// Session test endpoint for debugging
	app.Get("/session/test", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to get session",
				"error":   err.Error(),
			})
		}

		// Set a test value in session
		sess.Set("test_value", "session_working")
		if err := sess.Save(); err != nil {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to save session",
				"error":   err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":     "success",
			"message":    "Session test value set",
			"session_id": sess.ID(),
			"test_value": sess.Get("test_value"),
		})
	})

	// Serve static files (support running from project root or cmd/web)
	uiRoot := "ui"
	if _, err := os.Stat(uiRoot); os.IsNotExist(err) {
		uiRoot = filepath.Join("..", "ui")
	}
	app.Static("/", uiRoot)

	// Serve audio files
	app.Static("/audios", "./audios")

	// Add Logger middleware
	app.Use(logger.New())

	db := getDB(config, mlogger)
	defer db.Close()

	// Enable SQL logging for generated models
	models.SetLogger(log.Printf)

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

	// Initialize Voice service
	voiceConfig := services.VoiceConfig{
		VoiceURL: config.VoiceURL,
	}

	// Log Voice configuration
	mlogger.Info("Initializing Voice service",
		"voice_url", voiceConfig.VoiceURL)

	voiceService := services.NewVoiceService(voiceConfig)

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

	// Debug: List all users
	app.Get("/debug-users", func(c *fiber.Ctx) error {
		rows, err := db.QueryContext(context.Background(), "SELECT user_id, user_name, user_pass FROM users LIMIT 10")
		if err != nil {
			return c.SendString("Error querying users: " + err.Error())
		}
		defer rows.Close()

		var result string
		for rows.Next() {
			var id int
			var name, pass string
			err := rows.Scan(&id, &name, &pass)
			if err != nil {
				continue
			}
			result += fmt.Sprintf("ID: %d, Name: %s, Pass: %s\n", id, name, pass)
		}

		return c.SendString(result)
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
	routes.SetRoute(app, db, store, mlogger, config, smsService, voiceService)

	mlogger.Info("starting server...")
	// Start the app

	if err := app.Listen(config.Address); err != nil {
		mlogger.Error("Failed to start server", "error", err.Error())
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

// connect to database
func getDB(config handlers.Config, sl *slog.Logger) *sql.DB {
	// Use proper PostgreSQL connection string format
	// For local development, use localhost. For Docker, use "db"
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // Default to localhost for local development
	}

	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		dbHost, config.Ux, config.Px, config.Dx)

	sl.Info("Connecting to database", "host", dbHost, "user", config.Ux, "database", config.Dx)

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
