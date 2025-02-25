package main

import (
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http"
	"github.com/aifedorov/shortener/internal/storage"
)

func main() {
	cfg := config.NewConfig()
	cfg.ParseFlags()

	store := storage.NewFileStorage(cfg.FileStoragePath)
	srv := server.NewServer(cfg, store)

	srv.Run()
}
