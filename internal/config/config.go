package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	DBDriver                string
	FirebaseCredentials     string
	FirebaseCredentialsJSON string
	Port                string
	SellerPhone         string
	AppURL              string
	AllowedOrigins      string
	AdminAPIKey         string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading from environment")
	}

	return &Config{
		DatabaseURL:         mustGet("DATABASE_URL"),
		DBDriver:                getOrDefault("DB_DRIVER", "postgres"),
		FirebaseCredentials:     getOrDefault("FIREBASE_CREDENTIALS", "firebase-credentials.json"),
		FirebaseCredentialsJSON: os.Getenv("FIREBASE_CREDENTIALS_JSON"),
		Port:                getOrDefault("PORT", "8080"),
		SellerPhone:         mustGet("SELLER_WHATSAPP_PHONE"),
		AppURL:              getOrDefault("APP_URL", "http://localhost:8080"),
		AllowedOrigins:      getOrDefault("ALLOWED_ORIGINS", "*"),
		AdminAPIKey:         mustGet("ADMIN_API_KEY"),
		CloudinaryCloudName: mustGet("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    mustGet("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: mustGet("CLOUDINARY_API_SECRET"),
	}
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Required env variable %s not set", key)
	}
	return val
}

func getOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}