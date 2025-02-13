package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr      string
	ShortBaseURL string
	LogLevel     string
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.ShortBaseURL, "b", "http://localhost:8080", "address and port for short url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envShortBaseURL := os.Getenv("BASE_URL"); envShortBaseURL != "" {
		cfg.ShortBaseURL = envShortBaseURL
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
}
