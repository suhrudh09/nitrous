package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	JWTSecret              string
	Port                   string
	JolpicaBaseURL         string
	OpenF1BaseURL          string
	SportsDBBaseURL        string
	SportsDBAPIKey         string
	JolpicaSyncInterval    string
	OpenF1ActiveInterval   string
	OpenF1IdleInterval     string
	SportsDBSyncInterval   string
	ExternalRequestTimeout string
	F1YouTubeLiveURL       string
	F1TwitchLiveURL        string
}

var AppConfig Config

func LoadConfig() {
	// Load .env file (optional in production)
	godotenv.Load()

	AppConfig = Config{
		DBHost:                 getEnv("DB_HOST", "localhost"),
		DBPort:                 getEnv("DB_PORT", "5432"),
		DBUser:                 getEnv("DB_USER", "postgres"),
		DBPassword:             getEnv("DB_PASSWORD", "postgres"),
		DBName:                 getEnv("DB_NAME", "nitrous"),
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key-change-this"),
		Port:                   getEnv("PORT", "8080"),
		JolpicaBaseURL:         getEnv("JOLPICA_BASE_URL", "https://api.jolpi.ca/ergast/f1"),
		OpenF1BaseURL:          getEnv("OPENF1_BASE_URL", "https://api.openf1.org/v1"),
		SportsDBBaseURL:        getEnv("SPORTSDB_BASE_URL", "https://www.thesportsdb.com/api/v1/json"),
		SportsDBAPIKey:         getEnv("SPORTSDB_API_KEY", "123"),
		JolpicaSyncInterval:    getEnv("JOLPICA_SYNC_INTERVAL", "24h"),
		OpenF1ActiveInterval:   getEnv("OPENF1_ACTIVE_INTERVAL", "5s"),
		OpenF1IdleInterval:     getEnv("OPENF1_IDLE_INTERVAL", "1h"),
		SportsDBSyncInterval:   getEnv("SPORTSDB_SYNC_INTERVAL", "168h"),
		ExternalRequestTimeout: getEnv("EXTERNAL_REQUEST_TIMEOUT", "10s"),
		F1YouTubeLiveURL:       getEnv("F1_YOUTUBE_LIVE_URL", "https://www.youtube.com/results?search_query=formula+1+live"),
		F1TwitchLiveURL:        getEnv("F1_TWITCH_LIVE_URL", "https://www.twitch.tv/directory/category/sports"),
	}

	log.Println("✓ Configuration loaded")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
