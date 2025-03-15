package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/handlers"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/http/middleware/compress"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/validate"
)

var (
	ErrShortURLMissing = errors.New("short URL is missing")
)

var supportedContentTypes = []string{
	"application/json",
	"text/plain",
	"text/html",
	"application/x-gzip",
}

type Server struct {
	router     *chi.Mux
	config     *config.Config
	repo       repository.Repository
	urlChecker validate.URLChecker
	ctx        context.Context
}

func NewServer(cfg *config.Config, repo repository.Repository) *Server {
	return &Server{
		router:     chi.NewRouter(),
		repo:       repo,
		config:     cfg,
		urlChecker: validate.NewService(),
		ctx:        context.Background(),
	}
}

func (s *Server) Run() {
	if err := logger.Initialize(s.config.LogLevel); err != nil {
		log.Fatal(err)
	}

	err := s.repo.Run()
	if err != nil {
		logger.Log.Fatal("server: failed to run repository", zap.Error(err))
	}
	defer func() {
		err := s.repo.Close()
		if err != nil {
			logger.Log.Fatal("server: failed to close repository", zap.Error(err))
		}
	}()

	s.router.Use(chimiddleware.AllowContentType(supportedContentTypes...))
	s.router.Use(compress.GzipMiddleware)
	s.router.Use(logger.RequestLogger)
	s.router.Use(logger.ResponseLogger)
	s.router.Use(auth.JWTAuth)

	s.mountHandlers()

	logger.Log.Info("server: running on", zap.String("address", s.config.RunAddr))
	lsErr := http.ListenAndServe(s.config.RunAddr, s.router)
	if lsErr != nil {
		logger.Log.Fatal("server: failed to run", zap.Error(lsErr))
	}
}

func (s *Server) mountHandlers() {
	s.router.Post("/", handlers.NewSavePlainTextHandler(s.config, s.repo, s.urlChecker))
	s.router.Post("/api/shorten", handlers.NewSaveJSONHandler(s.config, s.repo, s.urlChecker))
	s.router.Post("/api/shorten/batch", handlers.NewSaveJSONBatchHandler(s.config, s.repo, s.urlChecker))
	s.router.Get("/{shortURL}", handlers.NewRedirectHandler(s.repo))
	s.router.Get("/", func(res http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("server: got request with bad data", zap.String("method", r.Method))
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
	})
	s.router.Get("/ping", handlers.NewPingHandler(s.repo))
	s.router.Get("/api/user/urls", handlers.NewURLsHandler(s.config, s.repo))
}
