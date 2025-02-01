package config

import "flag"

type Config struct {
	RunAddr      string
	ShortURLAddr string
}

func (conf *Config) ParseFlags() {
	flag.StringVar(&conf.RunAddr, "a", ":8888", "address and port to run server")
	flag.StringVar(&conf.ShortURLAddr, "b", "http://localhost:8000", "address and port for short url")
	flag.Parse()
}
