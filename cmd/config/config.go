package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr      string
	ShortBaseURL string
}

func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.ShortBaseURL, "b", "http://localhost:8080", "address and port for short url")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envShortBaseURL := os.Getenv("BASE_URL"); envShortBaseURL != "" {
		cfg.ShortBaseURL = envShortBaseURL
	}
}
