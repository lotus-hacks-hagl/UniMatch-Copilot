package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	AIServiceURL  string
	PublicBaseURL string
	Env           string
	JWTSecret     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	return &Config{
		Port:          getEnv("PORT", "8894"),
		DatabaseURL:   mustEnv("DATABASE_URL"),
		AIServiceURL:  getEnv("AI_SERVICE_URL", "http://localhost:8895"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8894"),
		Env:           getEnv("ENV", "development"),
		JWTSecret:     getEnv("JWT_SECRET", "super-secret-unimatch-key-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
