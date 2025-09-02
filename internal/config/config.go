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
	Cache    CacheConfig
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
	Username        string
	Password        string
	AuthDB          string
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	TTL          time.Duration
	MaxSize      int
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
			Username:        getEnv("MONGO_USERNAME", ""),
			Password:        getEnv("MONGO_PASSWORD", ""),
			AuthDB:          getEnv("MONGO_AUTH_DB", "admin"),
			ConnectTimeout:  getDurationEnv("DB_CONNECT_TIMEOUT", 10*time.Second),
			QueryTimeout:    getDurationEnv("DB_QUERY_TIMEOUT", 5*time.Second),
		},
		Cache: CacheConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 3),
			MaxRetries:   getIntEnv("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			TTL:          getDurationEnv("REDIS_TTL", 1*time.Hour),
			MaxSize:      getIntEnv("REDIS_MAX_SIZE", 10000),
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

// Address returns the server address.
func (s ServerConfig) Address() string {
	return s.Host + ":" + s.Port
}

// IsProduction returns true if the environment is production.
func (a AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// IsDevelopment returns true if the environment is development.
func (a AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// HasAuthentication returns true if username and password are provided.
func (d DatabaseConfig) HasAuthentication() bool {
	return d.Username != "" && d.Password != ""
}

// Helper functions for environment variable parsing.

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

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return fallback
}
