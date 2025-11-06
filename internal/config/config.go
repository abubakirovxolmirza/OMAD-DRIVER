package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
	Telegram TelegramConfig
	CORS     CORSConfig
	Pricing  PricingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
	Env  string
	Domain string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	Directory   string
	MaxFileSize int64
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken     string
	AdminGroupID string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
}

// PricingConfig holds default pricing configuration
type PricingConfig struct {
	Discount1Person      float64
	Discount2Person      float64
	Discount3Person      float64
	DiscountFullCar      float64
	ServiceFeePercentage float64
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:   getEnv("SERVER_PORT", "8080"),
			Host:   getEnv("SERVER_HOST", "0.0.0.0"),
			Env:    getEnv("ENV", "development"),
			Domain: getEnv("SERVER_DOMAIN", "api.omad-driver.uz"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "taxi_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your_secret_key"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 720),
		},
		Upload: UploadConfig{
			Directory:   getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSize: getEnvAsInt64("MAX_UPLOAD_SIZE", 10485760), // 10MB
		},
		Telegram: TelegramConfig{
			BotToken:     getEnv("TELEGRAM_BOT_TOKEN", ""),
			AdminGroupID: getEnv("TELEGRAM_ADMIN_GROUP_ID", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "https://api.omad-driver.uz,https://*.omad-driver.uz,https://docs.omad-driver.uz,http://localhost:3000,http://localhost:5173"),
		},
		Pricing: PricingConfig{
			Discount1Person:      getEnvAsFloat("DISCOUNT_1_PERSON", 0),
			Discount2Person:      getEnvAsFloat("DISCOUNT_2_PERSON", 10),
			Discount3Person:      getEnvAsFloat("DISCOUNT_3_PERSON", 15),
			DiscountFullCar:      getEnvAsFloat("DISCOUNT_FULL_CAR", 20),
			ServiceFeePercentage: getEnvAsFloat("SERVICE_FEE_PERCENTAGE", 15),
		},
	}

	return cfg, nil
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
