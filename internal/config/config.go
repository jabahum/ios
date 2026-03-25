package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AuthConfig struct {
	Enabled             bool   `mapstructure:"ENABLED"`
	KeycloakInternalURL string `mapstructure:"KEYCLOAK_INTERNAL_URL"`
	KeycloakPublicURL   string `mapstructure:"KEYCLOAK_PUBLIC_URL"`
	KeycloakBaseURL     string `mapstructure:"KEYCLOAK_BASE_URL"`
	KeycloakRealm       string `mapstructure:"KEYCLOAK_REALM"`
	ClientID            string `mapstructure:"CLIENT_ID"`
	ClientSecret        string `mapstructure:"CLIENT_SECRET"`
	AdminClientID       string `mapstructure:"ADMIN_CLIENT_ID"`
	AdminClientSecret   string `mapstructure:"ADMIN_CLIENT_SECRET"`
	RedirectURI         string `mapstructure:"REDIRECT_URI"`
	SuccessRedirect     string `mapstructure:"SUCCESS_REDIRECT"`
	LogoutRedirect      string `mapstructure:"LOGOUT_REDIRECT"`
	CookieDomain        string `mapstructure:"COOKIE_DOMAIN"`
	CookieSecure        bool   `mapstructure:"COOKIE_SECURE"`
}

type APIAuthConfig struct {
	BaseURL  string `mapstructure:"BASE_URL"`
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}

type Config struct {
	// server
	AppEnv       string `mapstructure:"APP_ENV"`
	Port         string `mapstructure:"PORT"`
	Address      string `mapstructure:"ADDRESS"`
	ReadTimeout  int64  `mapstructure:"READ_TIMEOUT"`
	WriteTimeout int64  `mapstructure:"WRITE_TIMEOUT"`
	Static       string `mapstructure:"STATIC"`

	// db
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`

	// legacy/custom db fields from old config
	Ux string `mapstructure:"UX"`
	Px string `mapstructure:"PX"`
	Dx string `mapstructure:"DX"`

	// africas talking
	ATAPIKey   string `mapstructure:"AT_API_KEY"`
	ATUsername string `mapstructure:"AT_USERNAME"`
	ATPhone    string `mapstructure:"AT_PHONE"`

	// auth provider toggle
	AuthProvider bool `mapstructure:"AUTH_PROVIDER"`

	// sms
	SMSBaseURL  string `mapstructure:"SMS_BASE_URL"`
	SMSUsername string `mapstructure:"SMS_USERNAME"`
	SMSPassword string `mapstructure:"SMS_PASSWORD"`

	// voice
	VoiceURL string `mapstructure:"VOICE_URL"`

	// keycloak urls
	KeycloakInternalURL string `mapstructure:"KEYCLOAK_INTERNAL_URL"`
	KeycloakPublicURL   string `mapstructure:"KEYCLOAK_PUBLIC_URL"`
	KeycloakBaseURL     string `mapstructure:"KEYCLOAK_BASE_URL"`

	// keycloak client
	KeycloakRealm        string `mapstructure:"KEYCLOAK_REALM"`
	KeycloakClientID     string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	KeycloakClientSecret string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
	KeycloakAudience     string `mapstructure:"KEYCLOAK_AUDIENCE"`
	KeycloakRedirectURI  string `mapstructure:"KEYCLOAK_REDIRECT_URI"`

	// auth cookie/session
	AuthCookieDomain    string `mapstructure:"AUTH_COOKIE_DOMAIN"`
	AuthCookiePath      string `mapstructure:"AUTH_COOKIE_PATH"`
	AuthCookieSecure    bool   `mapstructure:"AUTH_COOKIE_SECURE"`
	AuthSuccessRedirect string `mapstructure:"AUTH_SUCCESS_REDIRECT"`
	AuthLogoutRedirect  string `mapstructure:"AUTH_LOGOUT_REDIRECT"`

	AuthAccessTokenName  string `mapstructure:"AUTH_ACCESS_TOKEN_NAME"`
	AuthRefreshTokenName string `mapstructure:"AUTH_REFRESH_TOKEN_NAME"`
	AuthSessionMaxAge    int    `mapstructure:"AUTH_SESSION_MAX_AGE"`

	// admin
	KeycloakAdminClientID     string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_ID"`
	KeycloakAdminClientSecret string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_SECRET"`

	// keycloak bootstrap/admin for docker-compose
	KeycloakVersion       string `mapstructure:"KEYCLOAK_VERSION"`
	KeycloakDB            string `mapstructure:"KEYCLOAK_DB"`
	KeycloakDBName        string `mapstructure:"KEYCLOAK_DB_NAME"`
	KeycloakDBUser        string `mapstructure:"KEYCLOAK_DB_USER"`
	KeycloakDBPassword    string `mapstructure:"KEYCLOAK_DB_PASSWORD"`
	KeycloakAdmin         string `mapstructure:"KEYCLOAK_ADMIN"`
	KeycloakAdminPassword string `mapstructure:"KEYCLOAK_ADMIN_PASSWORD"`
	KeycloakHostname      string `mapstructure:"KEYCLOAK_HOSTNAME"`

	// log
	LogData string `mapstructure:"LOG_DATA"`
	LogFile string `mapstructure:"LOG_FILE"`

	// facility
	Facility string `mapstructure:"FACILITY"`

	// external APIs
	DHIS2API  APIAuthConfig `mapstructure:"DHIS2_API"`
	AlertsAPI APIAuthConfig `mapstructure:"ALERTS_API"`

	// grouped auth config
	Auth AuthConfig `mapstructure:"AUTH"`
}

func LoadConfig(configPaths ...string) (Config, error) {
	var cfg Config

	v := viper.New()

	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// defaults
	v.SetDefault("APP_ENV", "local")
	v.SetDefault("PORT", "3000")

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")

	v.SetDefault("AUTH_PROVIDER", true)
	v.SetDefault("AUTH_COOKIE_PATH", "/")
	v.SetDefault("AUTH_COOKIE_SECURE", false)
	v.SetDefault("AUTH_ACCESS_TOKEN_NAME", "access_token")
	v.SetDefault("AUTH_REFRESH_TOKEN_NAME", "refresh_token")
	v.SetDefault("AUTH_SESSION_MAX_AGE", 86400)
	v.SetDefault("AUTH_SUCCESS_REDIRECT", "/home")
	v.SetDefault("AUTH_LOGOUT_REDIRECT", "/login")

	// grouped auth defaults
	v.SetDefault("AUTH.ENABLED", true)
	v.SetDefault("AUTH.COOKIE_SECURE", false)
	v.SetDefault("AUTH.SUCCESS_REDIRECT", "/home")
	v.SetDefault("AUTH.LOGOUT_REDIRECT", "/login")

	// keycloak bootstrap defaults
	v.SetDefault("KEYCLOAK_VERSION", "26.0.5")
	v.SetDefault("KEYCLOAK_DB", "postgres")
	v.SetDefault("KEYCLOAK_DB_NAME", "keycloak")
	v.SetDefault("KEYCLOAK_DB_USER", "keycloak")
	v.SetDefault("KEYCLOAK_DB_PASSWORD", "keycloak")
	v.SetDefault("KEYCLOAK_ADMIN", "admin")
	v.SetDefault("KEYCLOAK_ADMIN_PASSWORD", "admin")
	v.SetDefault("KEYCLOAK_HOSTNAME", "http://localhost:8081")

	// optional env file support
	for _, path := range configPaths {
		if path == "" {
			continue
		}

		v.SetConfigFile(path)
		if err := v.MergeInConfig(); err != nil {
			return cfg, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// sync nested Auth config with flat fields when nested values are not provided
	if cfg.Auth.KeycloakInternalURL == "" {
		cfg.Auth.KeycloakInternalURL = cfg.KeycloakInternalURL
	}
	if cfg.Auth.KeycloakPublicURL == "" {
		cfg.Auth.KeycloakPublicURL = cfg.KeycloakPublicURL
	}
	if cfg.Auth.KeycloakBaseURL == "" {
		cfg.Auth.KeycloakBaseURL = cfg.KeycloakBaseURL
	}
	if cfg.Auth.KeycloakRealm == "" {
		cfg.Auth.KeycloakRealm = cfg.KeycloakRealm
	}
	if cfg.Auth.ClientID == "" {
		cfg.Auth.ClientID = cfg.KeycloakClientID
	}
	if cfg.Auth.ClientSecret == "" {
		cfg.Auth.ClientSecret = cfg.KeycloakClientSecret
	}
	if cfg.Auth.AdminClientID == "" {
		cfg.Auth.AdminClientID = cfg.KeycloakAdminClientID
	}
	if cfg.Auth.AdminClientSecret == "" {
		cfg.Auth.AdminClientSecret = cfg.KeycloakAdminClientSecret
	}
	if cfg.Auth.RedirectURI == "" {
		cfg.Auth.RedirectURI = cfg.KeycloakRedirectURI
	}
	if cfg.Auth.SuccessRedirect == "" {
		cfg.Auth.SuccessRedirect = cfg.AuthSuccessRedirect
	}
	if cfg.Auth.LogoutRedirect == "" {
		cfg.Auth.LogoutRedirect = cfg.AuthLogoutRedirect
	}
	if cfg.Auth.CookieDomain == "" {
		cfg.Auth.CookieDomain = cfg.AuthCookieDomain
	}
	if !cfg.Auth.CookieSecure {
		cfg.Auth.CookieSecure = cfg.AuthCookieSecure
	}

	return cfg, nil
}

func (c Config) DBSource() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}

func (c Config) Addr() string {
	if c.Address != "" {
		return c.Address
	}
	return "0.0.0.0:" + c.Port
}

func (c Config) Validate() error {
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.KeycloakRealm == "" && c.Auth.KeycloakRealm == "" {
		return fmt.Errorf("KEYCLOAK_REALM is required")
	}
	if c.KeycloakClientID == "" && c.Auth.ClientID == "" {
		return fmt.Errorf("KEYCLOAK_CLIENT_ID is required")
	}
	if c.KeycloakRedirectURI == "" && c.Auth.RedirectURI == "" {
		return fmt.Errorf("KEYCLOAK_REDIRECT_URI is required")
	}
	return nil
}
