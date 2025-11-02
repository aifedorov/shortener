package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aifedorov/shortener/internal/config"
	grpcserver "github.com/aifedorov/shortener/internal/grpc"
	httpserver "github.com/aifedorov/shortener/internal/http"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
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
		log.Fatalf("failed to load config: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	repo := repository.NewRepository(ctx, cfg)
	hSrv := httpserver.NewServer(ctx, cfg, repo)
	gSrv := grpcserver.NewShortenerServer(cfg, repo, validate.NewService())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Info("starting gRPC server", zap.String("address", cfg.GRPCAddr))
		if err := gSrv.Run(); err != nil {
			logger.Log.Error("gRPC server stopped", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Info("starting HTTP server", zap.String("address", cfg.RunAddr))
		if err := hSrv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("HTTP server stopped", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Log.Info("received shutdown signal, starting graceful shutdown")

	if err := hSrv.Shutdown(); err != nil {
		logger.Log.Error("failed to shutdown HTTP server", zap.Error(err))
	}

	gSrv.Shutdown()

	wg.Wait()

	if err := repo.Close(); err != nil {
		logger.Log.Error("failed to close repository", zap.Error(err))
	}

	logger.Log.Info("graceful shutdown completed")

	if err := logger.Log.Sync(); err != nil {
		log.Printf("failed to sync logger: %v", err)
	}
}
