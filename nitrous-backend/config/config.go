package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string
}

var AppConfig Config

func LoadConfig() {
	// Load .env file — optional in production (env vars take precedence)
	godotenv.Load()

	AppConfig = Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "nitrous"),
		JWTSecret:  getEnv("JWT_SECRET", "nitrous-dev-secret-change-in-production"),
		Port:       getEnv("PORT", "8080"),
	}

	// Fail fast — an empty JWT secret lets anyone forge tokens
	if AppConfig.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable must be set")
	}

	log.Println("✓ Configuration loaded")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
