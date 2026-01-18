package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI        string
	MongoDBName     string
	GoogleClientID  string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Port 			string
}

func Load() (*Config, error) {
	 _ = godotenv.Load()
	cfg := &Config{
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:     getEnv("MONGO_DB_NAME", "notesapp"),
		GoogleClientID:  getEnv("GOOGLE_CLIENT_ID", ""),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		AccessTokenTTL:  getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		Port:			 getEnv("PORT","3000"),
	}

	// Hard fail on required secrets
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if cfg.GoogleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	return cfg, nil
}
