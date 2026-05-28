package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	AppURL          string
	AppKey          string
	FrontendURL     string
	DBHost          string
	DBPort          string
	DBDatabase      string
	DBUsername      string
	DBPassword      string
	FilesystemDisk  string
	StoragePath     string
	AWSAccessKeyID  string
	AWSSecretKey    string
	AWSRegion       string
	AWSBucket       string
	AWSEndpoint     string
	ResendAPIKey    string
	MailFromAddress string
	MailFromName    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppURL:          getEnv("APP_URL", "http://localhost:8080"),
		AppKey:          getEnv("APP_KEY", "default-secret-key-change-in-production"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:4323"),
		DBHost:          getEnv("DB_HOST", "127.0.0.1"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBDatabase:      getEnv("DB_DATABASE", "trustwired"),
		DBUsername:      getEnv("DB_USERNAME", "root"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		FilesystemDisk:  getEnv("FILESYSTEM_DISK", "local"),
		StoragePath:     getEnv("STORAGE_PATH", "storage"),
		AWSAccessKeyID:  getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:    getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:       getEnv("AWS_DEFAULT_REGION", "sgp1"),
		AWSBucket:       getEnv("AWS_BUCKET", ""),
		AWSEndpoint:     getEnv("AWS_ENDPOINT", ""),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		MailFromAddress: getEnv("MAIL_FROM_ADDRESS", "noreply@internal-t.com"),
		MailFromName:    getEnv("MAIL_FROM_NAME", "internal-t"),
	}
}

// AllowedOrigins returns a slice of allowed CORS origins.
// Set CORS_ORIGINS=http://localhost:4321,https://app.yourdomain.com to allow multiple.
// Falls back to FRONTEND_URL if CORS_ORIGINS is not set.
func (c *Config) AllowedOrigins() []string {
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		var origins []string
		for _, o := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		return origins
	}
	return []string{c.FrontendURL}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
