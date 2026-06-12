package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	JWTExpiration time.Duration
	Port          string
	AllowedOrigin string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://taskmanager:secretpassword@db:5432/taskmanager?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-this-to-a-very-long-random-secret-key-in-production"
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	origin := os.Getenv("API_ALLOWED_ORIGINS")
	if origin == "" {
		origin = "http://localhost:3000"
	}

	return &Config{
		DatabaseURL:   dbURL,
		JWTSecret:     jwtSecret,
		JWTExpiration: 24 * time.Hour,
		Port:          port,
		AllowedOrigin: origin,
	}
}