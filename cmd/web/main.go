package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"

	_ "github.com/lib/pq"

	_ "case/cmd/web/docs"
	"case/internal/config"
	dbpkg "case/internal/database"
	flogger "case/internal/log"
	"case/internal/models"
	"case/internal/routes"
	"case/internal/services"
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	appLogger := flogger.InitLogger(cfg.LogFile)

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

	// Initialize session store using config
	store := session.New(session.Config{
		Expiration:     time.Duration(cfg.AuthSessionMaxAge) * time.Second,
		KeyLookup:      "cookie:fiber_sess",
		CookieSecure:   cfg.AuthCookieSecure,
		CookieHTTPOnly: true,
		CookiePath:     cfg.AuthCookiePath,
		CookieDomain:   cfg.AuthCookieDomain,
		CookieSameSite: "Lax",
	})

	app := fiber.New(fiber.Config{
		AppName:               "Integrated Outbreak System",
		ReadTimeout:           time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:          time.Duration(cfg.WriteTimeout) * time.Second,
		Prefork:               false,
		ServerHeader:          "IOS",
		DisableStartupMessage: cfg.AppEnv == "prod",
	})

	app.Use(recover.New())

	app.Use(fiberlogger.New())

	// Optional debug middleware in non-production only
	if cfg.AppEnv != "prod" {
		app.Use(func(c *fiber.Ctx) error {
			sess, err := store.Get(c)
			if err != nil {
				log.Printf("debug: failed to get session: %v", err)
			} else {
				log.Printf(
					"debug: path=%s session_id=%s authenticated=%v",
					c.Path(),
					sess.ID(),
					sess.Get("isAuthenticated"),
				)
			}
			return c.Next()
		})
	}

	// Enable SQL logging for generated models
	models.SetLogger(log.Printf)

	// Serve static files
	uiRoot := resolveUIRoot(cfg.Static)
	app.Static("/", uiRoot)

	// Serve audio files if present
	if _, err := os.Stat("./audios"); err == nil {
		app.Static("/audios", "./audios")
	}

	// Initialize SMS service
	smsConfig := services.SMSConfig{
		BaseURL:  cfg.SMSBaseURL,
		Username: cfg.SMSUsername,
		Password: cfg.SMSPassword,
	}

	appLogger.Info("initializing SMS service",
		"base_url", smsConfig.BaseURL,
		"username", smsConfig.Username,
	)

	smsService := services.NewSMSService(smsConfig)

	// Initialize Voice service
	voiceConfig := services.VoiceConfig{
		VoiceURL: cfg.VoiceURL,
	}

	appLogger.Info("initializing Voice service",
		"voice_url", voiceConfig.VoiceURL,
	)

	voiceService := services.NewVoiceService(voiceConfig)

	// Set up routes
	routes.SetRoute(app, primaryDB, store, appLogger, cfg, smsService, voiceService)

	addr := cfg.Addr()
	appLogger.Info("starting server", "address", addr, "env", cfg.AppEnv)

	if err := app.Listen(addr); err != nil {
		appLogger.Error("failed to start server", "error", err)
	}
}

func resolveUIRoot(configured string) string {
	candidates := []string{}

	if configured != "" {
		candidates = append(candidates, configured)
	}

	candidates = append(candidates,
		"ui",
		filepath.Join("..", "ui"),
	)

	for _, path := range candidates {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "ui"
}
