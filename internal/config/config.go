package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// File repository constants
const (
	// FilePermissionsRead defines the file permissions for read operations.
	FilePermissionsRead = 0444
	// FileOpenFlagsRead defines the flags for opening files in read mode.
	FileOpenFlagsRead = os.O_RDONLY
)

// Config holds the application configuration settings.
// Configuration can be set via command line flags or environment variables.
type Config struct {
	// RunAddr is the address and port where the server will listen to.
	RunAddr string `json:"server_address"`
	// BaseURL is the base URL used for generating short URLs.
	BaseURL string `json:"base_url"`
	// LogLevel specifies the logging level (debug, info, warn, error).
	LogLevel string `json:"log_level"`
	// FileStoragePath is the path to the file-based storage (optional).
	FileStoragePath string `json:"file_storage_path"`
	// DSN is the PostgreSQL database connection string (optional).
	DSN string `json:"database_dsn"`
	// SecretKey is used for JWT token signing and validation.
	SecretKey string `json:"-"`
	// EnableHTTPS specifies whether to enable HTTPS.
	EnableHTTPS bool `json:"enable_https"`
	// ConfigPath is the path to the config file.
	ConfigPath string `json:"-"`
}

// LoadConfig parses command line flags, environment variables to populate the configuration, then reads the JSON config file.
// Priority Order: CLI flags > Environment > Config file > Defaults
func LoadConfig() (*Config, error) {
	cfgEnvs, err := parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse envs: %w", err)
	}

	cfgFlags := parseFlags()

	var cfgFile *Config
	if cfgFlags.ConfigPath != "" {
		cfgFile, err = parseConfigFromFile(cfgFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	cfg := &Config{}
	if cfgFile != nil {
		err = mergeConfigs(cfg, cfgFile)
		if err != nil {
			log.Fatal("failed to merge configs: ", err)
		}
	}

	err = mergeConfigs(cfg, cfgEnvs)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	err = mergeConfigs(cfg, cfgFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	err = validateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

// parseEnvs parses configuration from environment variables.
// Returns a Config with values set from the following environment variables:
//   - SERVER_ADDRESS: server listen address
//   - BASE_URL: base URL for short URLs
//   - LOG_LEVEL: logging level
//   - FILE_STORAGE_PATH: path to file storage
//   - DATABASE_DSN: PostgreSQL connection string
//   - ENABLE_HTTPS: enable HTTPS (true/false)
//   - CONFIG: path to JSON config file
//   - SECRET_KEY: JWT signing key (required)
//
//nolint:cyclop
func parseEnvs() (*Config, error) {
	cfg := &Config{}
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

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		val, err := strconv.ParseBool(envEnableHTTPS)
		if err != nil {
			return nil, fmt.Errorf("invalid ENABLE_HTTPS value '%s': %w", envEnableHTTPS, err)
		}
		cfg.EnableHTTPS = val
	}

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		cfg.ConfigPath = envConfigPath
	}

	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		cfg.SecretKey = secretKey
	}

	if cfg.SecretKey == "" {
		return nil, errors.New("secret key is empty")
	}

	return cfg, nil
}

// parseFlags parses configuration from command line flags.
// Returns a Config with values set from the following flags:
//
//	-a: server address (default ":8080")
//	-b: base URL (default "http://localhost:8080")
//	-l: log level (default "info")
//	-f: file storage path
//	-d: database connection string
//	-s: enable HTTPS (default false)
//	-c: path to JSON config file
func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port for short url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file repository path")
	flag.StringVar(&cfg.DSN, "d", "", "postgres connection string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&cfg.ConfigPath, "c", "", "path to json config file")
	flag.Parse()
	return cfg
}

// parseConfigFromFile reads and parses configuration from a JSON file.
// Returns an error if the file cannot be read or contains invalid JSON.
func parseConfigFromFile(path string) (*Config, error) {
	cfg, err := readConfigFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return cfg, nil
}

// mergeConfigs merges configuration values from src into dst.
// Only non-empty values from src are copied to dst, preserving
// existing values in dst when src has empty/zero values.
// This implements the configuration priority system.
//
//nolint:cyclop
func mergeConfigs(dst *Config, src *Config) error {
	if src == nil {
		return errors.New("src is nil")
	}
	if dst == nil {
		return errors.New("dst is nil")
	}

	if src.RunAddr != "" {
		dst.RunAddr = src.RunAddr
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	if src.FileStoragePath != "" {
		dst.FileStoragePath = src.FileStoragePath
	}
	if src.DSN != "" {
		dst.DSN = src.DSN
	}
	if !src.EnableHTTPS {
		dst.EnableHTTPS = src.EnableHTTPS
	}
	if src.ConfigPath != "" {
		dst.ConfigPath = src.ConfigPath
	}
	if src.SecretKey != "" {
		dst.SecretKey = src.SecretKey
	}

	return nil
}

// validateConfig validates that all required configuration fields are set.
// Returns an error if any required field is empty or invalid.
// Required fields: RunAddr, BaseURL, LogLevel, FileStoragePath, DSN, SecretKey.
func validateConfig(cfg *Config) error {
	if cfg.RunAddr == "" {
		return errors.New("run address is empty")
	}
	if cfg.BaseURL == "" {
		return errors.New("base url is empty")
	}
	if cfg.LogLevel == "" {
		return errors.New("log level is empty")
	}
	if cfg.FileStoragePath == "" {
		return errors.New("file storage path is empty")
	}
	if cfg.DSN == "" {
		return errors.New("database connection string is empty")
	}
	if cfg.SecretKey == "" {
		return errors.New("secret key is empty")
	}

	return nil
}

// readConfigFromFile reads a JSON configuration file and unmarshals it into a Config struct.
// Returns an error if the file cannot be opened or contains invalid JSON.
func readConfigFromFile(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("failed to read config file: path is empty")
	}

	file, err := os.OpenFile(path, FileOpenFlagsRead, FilePermissionsRead)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg := Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return &cfg, nil
}
