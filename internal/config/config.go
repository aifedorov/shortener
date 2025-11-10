package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"
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
	// GRPCAddr is the address and port where the gRPC server will listen to.
	GRPCAddr string `json:"grpc_address"`
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
	// TrustedSubnet is the subnet allowed to access the API.
	TrustedSubnet string `json:"trusted_subnet"`
	// TrustedIPNet is the parsed CIDR network from TrustedSubnet.
	TrustedIPNet *net.IPNet `json:"-"`
}

// LoadConfig parses command line flags, environment variables to populate the configuration, then reads the JSON config file.
// Priority Order: CLI flags > Environment > Config file > Defaults
//
//nolint:cyclop
func LoadConfig() (*Config, error) {
	cfgEnvs, err := parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse envs: %w", err)
	}

	cfgFlags := parseFlags()

	var cfgFile *Config
	configPath := cfgFlags.ConfigPath
	if configPath == "" && cfgEnvs.ConfigPath != "" {
		configPath = cfgEnvs.ConfigPath
	}
	if configPath != "" {
		cfgFile, err = parseConfigFromFile(configPath)
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
//   - GRPC_ADDRESS: gRPC server listen address
//   - BASE_URL: base URL for short URLs
//   - LOG_LEVEL: logging level
//   - FILE_STORAGE_PATH: path to file storage
//   - DATABASE_DSN: PostgreSQL connection string
//   - ENABLE_HTTPS: enable HTTPS (true/false)
//   - CONFIG: path to JSON config file
//   - SECRET_KEY: JWT signing key (required)
//   - TRUSTED_SUBNET: subnet allowed to access the API
//
//nolint:cyclop
func parseEnvs() (*Config, error) {
	cfg := &Config{}

	if envRunAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.RunAddr = envRunAddr
	}

	if envGRPCAddr, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		cfg.GRPCAddr = envGRPCAddr
	}

	if envShortBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = envShortBaseURL
	}

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = envLogLevel
	}

	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DSN = envDSN
	}

	if envEnableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		val, err := strconv.ParseBool(envEnableHTTPS)
		if err != nil {
			return nil, fmt.Errorf("invalid ENABLE_HTTPS value '%s': %w", envEnableHTTPS, err)
		}
		cfg.EnableHTTPS = val
	}

	if envConfigPath, ok := os.LookupEnv("CONFIG"); ok {
		cfg.ConfigPath = envConfigPath
	}

	if envTrustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		cfg.TrustedSubnet = envTrustedSubnet
	}

	if secretKey, ok := os.LookupEnv("SECRET_KEY"); ok {
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
//		-a: server address (default ":8080")
//		-g: gRPC server address (default ":9090")
//		-b: base URL (default "http://localhost:8080")
//		-l: log level (default "info")
//		-f: file storage path
//		-d: database connection string
//		-s: enable HTTPS (default false)
//		-c: path to JSON config file
//	    -t: trusted subnet
func parseFlags() *Config {
	cfg := &Config{}
	if !flag.Parsed() {
		flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
		flag.StringVar(&cfg.GRPCAddr, "g", ":9090", "address and port to run gRPC server")
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port for short url")
		flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
		flag.StringVar(&cfg.FileStoragePath, "f", "", "file repository path")
		flag.StringVar(&cfg.DSN, "d", "", "postgres connection string")
		flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")
		flag.StringVar(&cfg.ConfigPath, "c", "", "path to json config file")
		flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet")
		flag.Parse()
	}
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
	if src.GRPCAddr != "" {
		dst.GRPCAddr = src.GRPCAddr
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
	if src.EnableHTTPS {
		dst.EnableHTTPS = src.EnableHTTPS
	}
	if src.ConfigPath != "" {
		dst.ConfigPath = src.ConfigPath
	}
	if src.TrustedSubnet != "" {
		dst.TrustedSubnet = src.TrustedSubnet
	}
	if src.SecretKey != "" {
		dst.SecretKey = src.SecretKey
	}

	return nil
}

// validateConfig validates that all required configuration fields are set.
// Returns an error if any required field is empty or invalid.
// Required fields: RunAddr, GRPCAddr, BaseURL, LogLevel, SecretKey.
// Optional: FileStoragePath, DSN (at least one storage type will be selected)
func validateConfig(cfg *Config) error {
	if cfg.RunAddr == "" {
		return errors.New("run address is empty")
	}
	if cfg.GRPCAddr == "" {
		return errors.New("grpc address is empty")
	}
	if cfg.BaseURL == "" {
		return errors.New("base url is empty")
	}
	if cfg.LogLevel == "" {
		return errors.New("log level is empty")
	}
	if cfg.SecretKey == "" {
		return errors.New("secret key is empty")
	}
	if len(cfg.TrustedSubnet) != 0 {
		_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("trusted subnet is invalid: %w", err)
		}
		cfg.TrustedIPNet = ipnet
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
		err = file.Close()
		if err != nil {
			logger.Log.Error("failed to close config file", zap.Error(err))
		}
	}()

	cfg := Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return &cfg, nil
}
