package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/handlers"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/http/middleware/compress"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
)

// Server error definitions
var (
	// ErrShortURLMissing is returned when a request is missing the required short URL parameter.
	ErrShortURLMissing = errors.New("short URL is missing")
)

// supportedContentTypes defines the content types that the server accepts.
var supportedContentTypes = []string{
	"application/json",
	"text/plain",
	"text/html",
	"application/x-gzip",
}

// Server represents the HTTP server for the URL shortener application.
// It manages HTTP routes, middleware, and coordinates between handlers and the repository.
type Server struct {
	// router is the Chi router instance for handling HTTP routes.
	router *chi.Mux
	// config holds the application configuration settings.
	config *config.Config
	// repo is the repository interface for data persistence.
	repo repository.Repository
	// urlChecker is used for validating URLs before processing.
	urlChecker validate.URLChecker
	// ctx is the background context for the server.
	ctx context.Context
}

// NewServer creates a new HTTP server instance with the provided configuration and repository.
// The server is initialized with Chi router, URL validation service, and background context.
func NewServer(cfg *config.Config, repo repository.Repository) *Server {
	return &Server{
		router:     chi.NewRouter(),
		repo:       repo,
		config:     cfg,
		urlChecker: validate.NewService(),
		ctx:        context.Background(),
	}
}

// Run starts the HTTP server and begins listening for requests.
// It initializes the logger, repository, middleware, and mounts all route handlers.
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

	m := auth.NewMiddleware(s.config.SecretKey)
	s.router.Use(m.JWTAuth)

	s.mountHandlers()

	logger.Log.Info("server: running on", zap.String("address", s.config.RunAddr))
	lsErr := http.ListenAndServe(s.config.RunAddr, s.router)
	if lsErr != nil {
		logger.Log.Fatal("server: failed to run", zap.Error(lsErr))
	}

	s.router.Mount("/debug", chimiddleware.Profiler())
}

// mountHandlers registers all HTTP route handlers with the router.
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
	s.router.Delete("/api/user/urls", handlers.NewDeleteHandler(s.repo))
}
