package config

import (
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	SecureCookies bool
}

func Load() Config {
	_ = godotenv.Load()

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getEnv("DB_USER", "uptime"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432")),
		Path:     getEnv("DB_NAME", "uptime"),
		RawQuery: "sslmode=disable",
	}

	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   dbURL.String(),
		SecureCookies: getEnv("COOKIE_SECURE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
