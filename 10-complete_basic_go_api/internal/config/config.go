package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const insecureDevSecret = "dev-only-secret-change-me-32chars!!"

type Config struct {
	Env                string
	HTTPAddr           string
	Store              string
	DatabaseURL        string
	JWTSecret          string
	JWTIssuer          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CORSOrigins        []string
	LogLevel           string
	RateLimitRPS       float64
	RateLimitBurst     int
	AuthRateLimitRPS   float64
	AuthRateLimitBurst int
	AutoMigrate        bool
	TrustProxy         bool
	ShutdownTimeout    time.Duration
	CleanupInterval    time.Duration
	RequestTimeout     time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:         getenv("API_ENV", "development"),
		HTTPAddr:    getenv("API_HTTP_ADDR", ":8080"),
		Store:       getenv("API_STORE", "postgres"),
		DatabaseURL: os.Getenv("API_DATABASE_URL"),
		JWTSecret:   os.Getenv("API_JWT_SECRET"),
		JWTIssuer:   getenv("API_JWT_ISSUER", "go-api"),
		LogLevel:    getenv("API_LOG_LEVEL", "info"),
		CORSOrigins: splitCSV(os.Getenv("API_CORS_ORIGINS")),
	}

	var err error
	if cfg.AccessTokenTTL, err = durationEnv("API_ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTokenTTL, err = durationEnv("API_REFRESH_TOKEN_TTL", 168*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = durationEnv("API_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.CleanupInterval, err = durationEnv("API_CLEANUP_INTERVAL", time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.RequestTimeout, err = durationEnv("API_REQUEST_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RateLimitRPS, err = floatEnv("API_RATE_LIMIT_RPS", 10); err != nil {
		return Config{}, err
	}
	if cfg.RateLimitBurst, err = intEnv("API_RATE_LIMIT_BURST", 20); err != nil {
		return Config{}, err
	}
	if cfg.AuthRateLimitRPS, err = floatEnv("API_AUTH_RATE_LIMIT_RPS", 0.5); err != nil {
		return Config{}, err
	}
	if cfg.AuthRateLimitBurst, err = intEnv("API_AUTH_RATE_LIMIT_BURST", 5); err != nil {
		return Config{}, err
	}
	if cfg.AutoMigrate, err = boolEnv("API_AUTO_MIGRATE", false); err != nil {
		return Config{}, err
	}
	if cfg.TrustProxy, err = boolEnv("API_TRUST_PROXY", false); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg Config) Validate() error {
	switch cfg.Env {
	case "development", "test", "production":
	default:
		return fmt.Errorf("API_ENV must be development, test, or production")
	}
	switch cfg.Store {
	case "memory", "postgres":
	default:
		return fmt.Errorf("API_STORE must be memory or postgres")
	}
	if cfg.HTTPAddr == "" {
		return errors.New("API_HTTP_ADDR is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return errors.New("API_JWT_SECRET must be at least 32 bytes")
	}
	if cfg.AccessTokenTTL <= 0 || cfg.AccessTokenTTL > 24*time.Hour {
		return errors.New("API_ACCESS_TOKEN_TTL must be between 1ns and 24h")
	}
	if cfg.RefreshTokenTTL <= cfg.AccessTokenTTL || cfg.RefreshTokenTTL > 30*24*time.Hour {
		return errors.New("API_REFRESH_TOKEN_TTL must be longer than the access token and at most 30 days")
	}
	if cfg.ShutdownTimeout < time.Second || cfg.ShutdownTimeout > time.Minute {
		return errors.New("API_SHUTDOWN_TIMEOUT must be between 1s and 1m")
	}
	if cfg.CleanupInterval < time.Minute || cfg.CleanupInterval > 24*time.Hour {
		return errors.New("API_CLEANUP_INTERVAL must be between 1m and 24h")
	}
	if cfg.RequestTimeout < time.Second || cfg.RequestTimeout > time.Minute {
		return errors.New("API_REQUEST_TIMEOUT must be between 1s and 1m")
	}
	if cfg.RateLimitRPS <= 0 || cfg.RateLimitRPS > 100000 || cfg.RateLimitBurst < 1 || cfg.RateLimitBurst > 10000 {
		return errors.New("invalid API_RATE_LIMIT_RPS or API_RATE_LIMIT_BURST")
	}
	if cfg.AuthRateLimitRPS <= 0 || cfg.AuthRateLimitRPS > 100000 || cfg.AuthRateLimitBurst < 1 || cfg.AuthRateLimitBurst > 10000 {
		return errors.New("invalid API_AUTH_RATE_LIMIT_RPS or API_AUTH_RATE_LIMIT_BURST")
	}
	for _, origin := range cfg.CORSOrigins {
		if err := validateOrigin(origin); err != nil {
			return err
		}
	}
	if cfg.Store == "postgres" && strings.TrimSpace(cfg.DatabaseURL) == "" {
		return errors.New("API_DATABASE_URL is required when API_STORE=postgres")
	}

	if cfg.Env == "production" {
		if cfg.Store == "memory" {
			return errors.New("API_STORE=memory is not allowed in production")
		}
		if cfg.JWTSecret == insecureDevSecret {
			return errors.New("replace API_JWT_SECRET before running in production")
		}
		if !secureDatabaseURL(cfg.DatabaseURL) {
			return errors.New("production database url must set sslmode to require, verify-ca, or verify-full")
		}
		for _, origin := range cfg.CORSOrigins {
			if origin == "*" {
				return errors.New("API_CORS_ORIGINS cannot contain * in production")
			}
		}
	}
	return nil
}

func secureDatabaseURL(raw string) bool {
	return strings.Contains(raw, "sslmode=require") ||
		strings.Contains(raw, "sslmode=verify-ca") ||
		strings.Contains(raw, "sslmode=verify-full")
}

func validateOrigin(origin string) error {
	if origin == "*" {
		return nil
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid CORS origin %q", origin)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("invalid CORS origin %q", origin)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return fmt.Errorf("invalid CORS origin %q", origin)
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return b, nil
}

func intEnv(key string, fallback int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return n, nil
}

func floatEnv(key string, fallback float64) (float64, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return n, nil
}

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
