package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		configFile     string
		configContent  string
		expectedError  bool
		expectedConfig *Config
	}{
		{
			name: "successful load with environment variables",
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":9090",
				"BASE_URL":          "http://localhost:9090",
				"LOG_LEVEL":         "debug",
				"FILE_STORAGE_PATH": "/tmp/storage",
				"DATABASE_DSN":      "postgres://test",
				"ENABLE_HTTPS":      "true",
				"SECRET_KEY":        "test-secret-key",
			},
			expectedError: false,
			expectedConfig: &Config{
				RunAddr:         ":9090",
				BaseURL:         "http://localhost:9090",
				LogLevel:        "debug",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
				EnableHTTPS:     true,
				SecretKey:       "test-secret-key",
			},
		},
		{
			name: "missing secret key",
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":9090",
				"BASE_URL":          "http://localhost:9090",
				"LOG_LEVEL":         "debug",
				"FILE_STORAGE_PATH": "/tmp/storage",
				"DATABASE_DSN":      "postgres://test",
			},
			expectedError: true,
		},
		{
			name: "invalid ENABLE_HTTPS value",
			envVars: map[string]string{
				"SECRET_KEY":   "test-secret-key",
				"ENABLE_HTTPS": "invalid-bool",
			},
			expectedError: true,
		},
		{
			name: "config file with environment override",
			envVars: map[string]string{
				"SECRET_KEY":     "test-secret-key",
				"SERVER_ADDRESS": ":7070",
			},
			configFile: "test_config.json",
			configContent: `{
				"server_address": ":6060",
				"base_url": "http://localhost:6060",
				"log_level": "warn",
				"file_storage_path": "/config/storage",
				"database_dsn": "postgres://config",
				"enable_https": true
			}`,
			expectedError: false,
			expectedConfig: &Config{
				RunAddr:         ":7070",
				BaseURL:         "http://localhost:6060",
				LogLevel:        "warn",
				FileStoragePath: "/config/storage",
				DSN:             "postgres://config",
				EnableHTTPS:     true,
				SecretKey:       "test-secret-key",
				ConfigPath:      "test_config.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnv()

			if tt.configFile != "" && tt.configContent != "" {
				err := os.WriteFile(tt.configFile, []byte(tt.configContent), 0644)
				assert.NoError(t, err)
				defer os.Remove(tt.configFile)

				os.Setenv("CONFIG", tt.configFile)
			}

			cfg, err := LoadConfig()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.expectedConfig.RunAddr, cfg.RunAddr)
				assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
				assert.Equal(t, tt.expectedConfig.LogLevel, cfg.LogLevel)
				assert.Equal(t, tt.expectedConfig.FileStoragePath, cfg.FileStoragePath)
				assert.Equal(t, tt.expectedConfig.DSN, cfg.DSN)
				assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
				assert.Equal(t, tt.expectedConfig.SecretKey, cfg.SecretKey)
			}
		})
	}
}

func TestParseEnvs(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  bool
		expectedConfig *Config
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":8081",
				"BASE_URL":          "http://example.com",
				"LOG_LEVEL":         "error",
				"FILE_STORAGE_PATH": "/tmp/test",
				"DATABASE_DSN":      "postgres://user:pass@localhost/db",
				"ENABLE_HTTPS":      "true",
				"CONFIG":            "/path/to/config",
				"SECRET_KEY":        "my-secret-key",
			},
			expectedError: false,
			expectedConfig: &Config{
				RunAddr:         ":8081",
				BaseURL:         "http://example.com",
				LogLevel:        "error",
				FileStoragePath: "/tmp/test",
				DSN:             "postgres://user:pass@localhost/db",
				EnableHTTPS:     true,
				ConfigPath:      "/path/to/config",
				SecretKey:       "my-secret-key",
			},
		},
		{
			name: "minimal environment variables",
			envVars: map[string]string{
				"SECRET_KEY": "minimal-key",
			},
			expectedError: false,
			expectedConfig: &Config{
				SecretKey: "minimal-key",
			},
		},
		{
			name:          "missing secret key",
			envVars:       map[string]string{},
			expectedError: true,
		},
		{
			name: "invalid ENABLE_HTTPS",
			envVars: map[string]string{
				"SECRET_KEY":   "test-key",
				"ENABLE_HTTPS": "not-a-boolean",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnv()

			cfg, err := parseEnvs()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.expectedConfig != nil {
					assert.Equal(t, tt.expectedConfig.RunAddr, cfg.RunAddr)
					assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
					assert.Equal(t, tt.expectedConfig.LogLevel, cfg.LogLevel)
					assert.Equal(t, tt.expectedConfig.FileStoragePath, cfg.FileStoragePath)
					assert.Equal(t, tt.expectedConfig.DSN, cfg.DSN)
					assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
					assert.Equal(t, tt.expectedConfig.ConfigPath, cfg.ConfigPath)
					assert.Equal(t, tt.expectedConfig.SecretKey, cfg.SecretKey)
				}
			}
		})
	}
}

func TestParseConfigFromFile(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedError  bool
		expectedConfig *Config
	}{
		{
			name: "valid config file",
			configContent: `{
				"server_address": ":3000",
				"base_url": "https://short.ly",
				"log_level": "debug",
				"file_storage_path": "/data/urls",
				"database_dsn": "postgres://user@localhost/shortener",
				"enable_https": true
			}`,
			expectedError: false,
			expectedConfig: &Config{
				RunAddr:         ":3000",
				BaseURL:         "https://short.ly",
				LogLevel:        "debug",
				FileStoragePath: "/data/urls",
				DSN:             "postgres://user@localhost/shortener",
				EnableHTTPS:     true,
			},
		},
		{
			name:          "invalid JSON",
			configContent: `{"server_address": ":3000"`,
			expectedError: true,
		},
		{
			name:           "empty config file",
			configContent:  `{}`,
			expectedError:  false,
			expectedConfig: &Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := "test_config_" + tt.name + ".json"
			err := os.WriteFile(tempFile, []byte(tt.configContent), 0644)
			assert.NoError(t, err)
			defer os.Remove(tempFile)

			cfg, err := parseConfigFromFile(tempFile)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.expectedConfig != nil {
					assert.Equal(t, tt.expectedConfig.RunAddr, cfg.RunAddr)
					assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
					assert.Equal(t, tt.expectedConfig.LogLevel, cfg.LogLevel)
					assert.Equal(t, tt.expectedConfig.FileStoragePath, cfg.FileStoragePath)
					assert.Equal(t, tt.expectedConfig.DSN, cfg.DSN)
					assert.Equal(t, tt.expectedConfig.EnableHTTPS, cfg.EnableHTTPS)
				}
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name           string
		dst            *Config
		src            *Config
		expectedError  bool
		expectedResult *Config
	}{
		{
			name: "merge non-empty values",
			dst: &Config{
				RunAddr:  ":8080",
				BaseURL:  "http://localhost:8080",
				LogLevel: "info",
			},
			src: &Config{
				RunAddr:         ":9090",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
			},
			expectedError: false,
			expectedResult: &Config{
				RunAddr:         ":9090",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
			},
		},
		{
			name: "nil source",
			dst: &Config{
				RunAddr: ":8080",
			},
			src:           nil,
			expectedError: true,
		},
		{
			name:          "nil destination",
			dst:           nil,
			src:           &Config{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeConfigs(tt.dst, tt.src)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.RunAddr, tt.dst.RunAddr)
					assert.Equal(t, tt.expectedResult.BaseURL, tt.dst.BaseURL)
					assert.Equal(t, tt.expectedResult.LogLevel, tt.dst.LogLevel)
					assert.Equal(t, tt.expectedResult.FileStoragePath, tt.dst.FileStoragePath)
					assert.Equal(t, tt.expectedResult.DSN, tt.dst.DSN)
					assert.Equal(t, tt.expectedResult.EnableHTTPS, tt.dst.EnableHTTPS)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError bool
	}{
		{
			name: "valid config",
			config: &Config{
				RunAddr:         ":8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
				SecretKey:       "secret",
			},
			expectedError: false,
		},
		{
			name: "missing run address",
			config: &Config{
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
				SecretKey:       "secret",
			},
			expectedError: true,
		},
		{
			name: "missing base URL",
			config: &Config{
				RunAddr:         ":8080",
				LogLevel:        "info",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
				SecretKey:       "secret",
			},
			expectedError: true,
		},
		{
			name: "missing secret key",
			config: &Config{
				RunAddr:         ":8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "/tmp/storage",
				DSN:             "postgres://test",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadConfigFromFile(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		fileContent    string
		createFile     bool
		expectedError  bool
		expectedConfig *Config
	}{
		{
			name:     "valid config file",
			filePath: "valid_config.json",
			fileContent: `{
				"server_address": ":4000",
				"base_url": "https://example.com",
				"log_level": "warn"
			}`,
			createFile:    true,
			expectedError: false,
			expectedConfig: &Config{
				RunAddr:  ":4000",
				BaseURL:  "https://example.com",
				LogLevel: "warn",
			},
		},
		{
			name:          "empty file path",
			filePath:      "",
			expectedError: true,
		},
		{
			name:          "non-existent file",
			filePath:      "non_existent.json",
			createFile:    false,
			expectedError: true,
		},
		{
			name:          "invalid JSON",
			filePath:      "invalid.json",
			fileContent:   `{"invalid": json}`,
			createFile:    true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFile && tt.filePath != "" {
				err := os.WriteFile(tt.filePath, []byte(tt.fileContent), 0644)
				assert.NoError(t, err)
				defer os.Remove(tt.filePath)
			}

			cfg, err := readConfigFromFile(tt.filePath)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.expectedConfig != nil {
					assert.Equal(t, tt.expectedConfig.RunAddr, cfg.RunAddr)
					assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL)
					assert.Equal(t, tt.expectedConfig.LogLevel, cfg.LogLevel)
				}
			}
		})
	}
}

func clearEnv() {
	envVars := []string{
		"SERVER_ADDRESS",
		"BASE_URL",
		"LOG_LEVEL",
		"FILE_STORAGE_PATH",
		"DATABASE_DSN",
		"ENABLE_HTTPS",
		"CONFIG",
		"SECRET_KEY",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
