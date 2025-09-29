package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http"
	"github.com/aifedorov/shortener/internal/repository"
)

func main() {
	cfg := config.NewConfig()
	cfg.ParseFlags()

	repo := repository.NewRepository(context.Background(), cfg)
	srv := server.NewServer(cfg, repo)
	srv.Run()
}
