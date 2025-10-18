package main

import (
	"context"
	"fmt"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http"
	"github.com/aifedorov/shortener/internal/repository"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("failed to load config", zap.Error(err))
	}
	repo := repository.NewRepository(context.Background(), cfg)
	srv := server.NewServer(cfg, repo)

	if err := srv.Run(); err != nil {
		logger.Log.Fatal("server: failed to run", zap.Error(err))
	}
}
