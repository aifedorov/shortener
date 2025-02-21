package server

import (
	"errors"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/handlers/redirect"
	"github.com/aifedorov/shortener/internal/http/handlers/save"
	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/internal/middleware"
	"github.com/aifedorov/shortener/internal/storage"
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
	router *chi.Mux
	store  storage.Storage
	config *config.Config
}

func NewServer() *Server {
	return &Server{
		router: chi.NewRouter(),
		store:  storage.NewMemoryStorage(),
		config: config.NewConfig(),
	}
}

func (s *Server) Run() {
	s.config.ParseFlags()

	if err := logger.Initialize(s.config.LogLevel); err != nil {
		log.Fatal(err)
	}

	s.router.Use(chimiddleware.AllowContentType(supportedContentTypes...))

	s.router.Use(middleware.GzipMiddleware)
	s.router.Use(logger.RequestLogger)
	s.router.Use(logger.ResponseLogger)

	s.mountHandlers()

	logger.Log.Info("Running server on", zap.String("address", s.config.RunAddr))
	err := http.ListenAndServe(s.config.RunAddr, s.router)
	if err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}

func (s *Server) mountHandlers() {
	s.router.Post("/", save.NewSavePlainTextHandler(s.config, s.store))
	s.router.Post("/api/shorten", save.NewSaveJSONHandler(s.config, s.store))
	s.router.Get("/{shortURL}", redirect.NewRedirectHandler(s.store))
	s.router.Get("/", func(res http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("got request with bad method", zap.String("method", r.Method))
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
	})
}
