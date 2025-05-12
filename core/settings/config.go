package settings

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBPath     string

	JWTSecret     string
	JWTExpiration int

	ServerPort int
	ServerHost string

	LogLevel  string
	LogFormat string

	AppName  string
	AppEnv   string
	AppDebug bool
}

var cfg *Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	cfg = &Config{
		DBType:     getEnv("DB_TYPE", "sqlite"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "bank"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBPath:     getEnv("DB_PATH", "bank.db"),

		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION", 24),

		ServerPort: getEnvAsInt("SERVER_PORT", 8080),
		ServerHost: getEnv("SERVER_HOST", "localhost"),

		LogLevel:  getEnv("LOG_LEVEL", "debug"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		AppName:  getEnv("APP_NAME", "FinanceGolang"),
		AppEnv:   getEnv("APP_ENV", "development"),
		AppDebug: getEnvAsBool("APP_DEBUG", true),
	}

	return nil
}

func Get() *Config {
	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
