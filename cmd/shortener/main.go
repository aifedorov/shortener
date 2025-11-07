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
	"time"

	"github.com/aifedorov/shortener/internal/config"
	urlDomain "github.com/aifedorov/shortener/internal/domain/url"
	userDomain "github.com/aifedorov/shortener/internal/domain/user"
	grpcserver "github.com/aifedorov/shortener/internal/grpc"
	httpserver "github.com/aifedorov/shortener/internal/http"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// tokenExp defines the JWT token expiration time.
const tokenExp = time.Hour * 3

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

	authService := jwt.NewService(cfg.SecretKey, tokenExp)
	validator := validate.NewService()
	randomizer := random.NewService()

	urlService := urlDomain.NewService(randomizer, validator)
	userService := userDomain.NewService()

	repo := repository.NewRepository(ctx, cfg)

	hSrv := httpserver.NewServer(ctx, cfg, chi.NewRouter(), repo, authService, urlService, userService)
	gSrv := grpcserver.NewShortenerServer(cfg, repo, urlService, userService, authService)

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
