package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	DataBase DatabaseConfig
	App      AppConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	URI             string
	Name            string
	QueryCollection string
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

// AppConfig holds general app configuration.
type AppConfig struct {
	Name                string
	Version             string
	Environment         string
	LogLevel            string
	DefaultSuggestions  int
	MaxSuggestionsLimit int
}

// Load loads configuration from environment variables and .env file
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		DataBase: DatabaseConfig{
			URI:             getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Name:            getEnv("DATABASE_NAME", "autocomplete_db"),
			QueryCollection: getEnv("QUERY_COLLECTION", "queries"),
			ConnectTimeout:  getDurationEnv("DB_CONNECT_TIMEOUT", 10*time.Second),
			QueryTimeout:    getDurationEnv("DB_QUERY_TIMEOUT", 5*time.Second),
		},
		App: AppConfig{
			Name:                getEnv("APP_NAME", "Autocomplete System"),
			Version:             getEnv("APP_VERSION", "1.0.0"),
			Environment:         getEnv("APP_ENV", "development"),
			LogLevel:            getEnv("LOG_LEVEL", "info"),
			DefaultSuggestions:  getIntEnv("DEFAULT_SUGGESTIONS", 5),
			MaxSuggestionsLimit: getIntEnv("MAX_SUGGESTIONS_LIMIT", 20),
		},
	}
}

// Address returns the server address
func (s ServerConfig) Address() string {
	return s.Host + ":" + s.Port
}

// IsProduction returns true if the environment is production
func (a AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (a AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// Helper functions for environment variable parsing

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return fallback
}
