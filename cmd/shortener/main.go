package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	repo := repository.NewRepository(ctx, cfg)
	srv := server.NewServer(ctx, cfg, repo)

	go func() {
		<-ctx.Done()

		logger.Log.Info("start gracefully shutting down")

		err = srv.Shutdown()
		if err != nil {
			log.Printf("failed to shutdown server (ignoring): %v", err)
		}

		err := repo.Close()
		if err != nil {
			log.Printf("failed to close repository (ignoring): %v", err)
		}

		logger.Log.Info("finish gracefully shutting down")

		if err := logger.Log.Sync(); err != nil {
			log.Printf("failed to sync logger (ignoring): %v", err)
		}
	}()

	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Fatal("failed to run server", zap.Error(err))
	}
}
