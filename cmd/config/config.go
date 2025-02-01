package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr      string `env:"SERVER_ADDRESS"`
	ShortBaseURL string `env:"BASE_URL"`
}

func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8888", "address and port to run server")
	flag.StringVar(&cfg.ShortBaseURL, "b", "http://localhost:8000", "address and port for short url")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envShortBaseURL := os.Getenv("BASE_URL"); envShortBaseURL != "" {
		cfg.ShortBaseURL = envShortBaseURL
	}
}
