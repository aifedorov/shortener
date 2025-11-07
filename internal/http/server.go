package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/aifedorov/shortener/internal/config"
	urlDomain "github.com/aifedorov/shortener/internal/domain/url"
	userDomain "github.com/aifedorov/shortener/internal/domain/user"
	"github.com/aifedorov/shortener/internal/http/handlers"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/http/middleware/compress"
	"github.com/aifedorov/shortener/internal/http/middleware/ipcheck"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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
	// config holds the application configuration settings.
	config *config.Config
	// repo is the repository interface for data persistence.
	repo repository.Repository
	// authService is the JWT authentication service.
	authService jwt.JWT
	// urlService is the domain service for URL operations.
	urlService urlDomain.Service
	// userService is the domain service for user operations.
	userService *userDomain.Service
	// ctx is the background context for the server.
	ctx context.Context
	// srv is the HTTP server instance.
	srv *http.Server
}

// NewServer creates a new HTTP server instance with the provided configuration and repository.
// The server is initialized with Chi router, domain services, and background context.
func NewServer(
	ctx context.Context,
	cfg *config.Config,
	repo repository.Repository,
	authService jwt.JWT,
	urlService urlDomain.Service,
	userService *userDomain.Service,
) *Server {
	return &Server{
		config:      cfg,
		repo:        repo,
		authService: authService,
		urlService:  urlService,
		userService: userService,
		ctx:         ctx,
		srv: &http.Server{
			Addr:    cfg.RunAddr,
			Handler: newRouter(cfg, repo, urlService, authService),
		},
	}
}

// Run starts the HTTP server and begins listening for requests.
// It initializes the logger, repository, middleware, and mounts all route handlers.
func (s *Server) Run() error {
	err := s.repo.Run()
	if err != nil {
		logger.Log.Fatal("server: failed to run repository", zap.Error(err))
	}

	if s.config.EnableHTTPS {
		logger.Log.Info("HTTPS server: running on", zap.String("address", s.config.RunAddr))
		return s.srv.ListenAndServeTLS("cert.pem", "key.pem")
	} else {
		logger.Log.Info("HTTP server: running on", zap.String("address", s.config.RunAddr))
		return s.srv.ListenAndServe()
	}
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown() error {
	return s.srv.Shutdown(s.ctx)
}

// NewRouter create a new roture then registers all HTTP route handlers and middleware.
func newRouter(cfg *config.Config, repo repository.Repository, urlService urlDomain.Service, authService jwt.JWT) *chi.Mux {
	router := chi.NewRouter()

	router.Use(chimiddleware.AllowContentType(supportedContentTypes...))
	router.Use(compress.GzipMiddleware)
	router.Use(logger.RequestLogger)
	router.Use(logger.ResponseLogger)

	authm := auth.NewMiddleware(authService)
	router.Use(authm.JWTAuth)

	router.Post("/", handlers.NewSavePlainTextHandler(cfg, repo, urlService))
	router.Post("/api/shorten", handlers.NewSaveJSONHandler(cfg, repo, urlService))
	router.Post("/api/shorten/batch", handlers.NewSaveJSONBatchHandler(cfg, repo, urlService))
	router.Get("/{shortURL}", handlers.NewRedirectHandler(repo))
	router.Get("/", func(res http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("server: got request with bad data", zap.String("method", r.Method))
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
	})
	router.Get("/ping", handlers.NewPingHandler(repo))
	router.Get("/api/user/urls", handlers.NewURLsHandler(cfg, repo))
	router.Delete("/api/user/urls", handlers.NewDeleteHandler(repo))

	ipcheckm := ipcheck.NewMiddleware(cfg.TrustedIPNet)
	router.With(ipcheckm.IPCheck).Get("/api/internal/stats", handlers.NewStatsHandler(repo))

	return router
}
