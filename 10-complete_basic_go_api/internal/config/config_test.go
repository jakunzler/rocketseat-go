package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDevelopmentMemory(t *testing.T) {
	t.Setenv("API_ENV", "development")
	t.Setenv("API_STORE", "memory")
	t.Setenv("API_HTTP_ADDR", ":8080")
	t.Setenv("API_DATABASE_URL", "")
	t.Setenv("API_JWT_SECRET", insecureDevSecret)
	t.Setenv("API_JWT_ISSUER", "go-api")
	t.Setenv("API_CORS_ORIGINS", "http://localhost:3000")
	t.Setenv("API_ACCESS_TOKEN_TTL", "15m")
	t.Setenv("API_REFRESH_TOKEN_TTL", "168h")
	t.Setenv("API_AUTO_MIGRATE", "false")
	t.Setenv("API_TRUST_PROXY", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Store != "memory" || cfg.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if len(cfg.CORSOrigins) != 1 {
		t.Fatalf("origins = %#v", cfg.CORSOrigins)
	}
}

func TestLoadProductionGuards(t *testing.T) {
	base := map[string]string{
		"API_ENV":               "production",
		"API_STORE":             "postgres",
		"API_HTTP_ADDR":         ":8080",
		"API_DATABASE_URL":      "postgres://api:api@db:5432/api?sslmode=verify-full",
		"API_JWT_SECRET":        strings.Repeat("p", 48),
		"API_CORS_ORIGINS":      "https://app.example.com",
		"API_AUTO_MIGRATE":      "false",
		"API_TRUST_PROXY":       "true",
		"API_ACCESS_TOKEN_TTL":  "15m",
		"API_REFRESH_TOKEN_TTL": "168h",
	}
	apply := func(overrides map[string]string) {
		for key, value := range base {
			t.Setenv(key, value)
		}
		for key, value := range overrides {
			t.Setenv(key, value)
		}
	}

	apply(nil)
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}

	cases := []map[string]string{
		{"API_STORE": "memory"},
		{"API_JWT_SECRET": insecureDevSecret},
		{"API_DATABASE_URL": "postgres://api:api@db:5432/api?sslmode=disable"},
		{"API_CORS_ORIGINS": "*"},
		{"API_JWT_SECRET": "too-short"},
	}
	for _, overrides := range cases {
		t.Run(overrides["API_STORE"]+overrides["API_DATABASE_URL"]+overrides["API_CORS_ORIGINS"], func(t *testing.T) {
			apply(overrides)
			if _, err := Load(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
