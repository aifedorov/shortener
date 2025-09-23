package config

import (
	"flag"
	"log"
	"os"
)

// Config holds the application configuration settings.
// Configuration can be set via command line flags or environment variables.
type Config struct {
	// RunAddr is the address and port where the server will listen.
	RunAddr string
	// BaseURL is the base URL used for generating short URLs.
	BaseURL string
	// LogLevel specifies the logging level (debug, info, warn, error).
	LogLevel string
	// FileStoragePath is the path to the file-based storage (optional).
	FileStoragePath string
	// DSN is the PostgreSQL database connection string (optional).
	DSN string
	// SecretKey is used for JWT token signing and validation.
	SecretKey string
}

// NewConfig creates a new Config instance with default values.
// The configuration should be populated using ParseFlags() before use.
func NewConfig() *Config {
	return &Config{}
}

// ParseFlags parses command line flags and environment variables to populate the configuration.
// Command line flags take precedence over environment variables.
func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port for short url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file repository path")
	flag.StringVar(&cfg.DSN, "d", "", "postgres connection string")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envShortBaseURL := os.Getenv("BASE_URL"); envShortBaseURL != "" {
		cfg.BaseURL = envShortBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}

	secretKey := os.Getenv("SECRET_KEY")
	cfg.SecretKey = secretKey
	if secretKey == "" {
		log.Fatal("secret key is not set")
	}
}
