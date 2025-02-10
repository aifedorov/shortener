package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/handlers/redirect"
	"github.com/aifedorov/shortener/internal/http/handlers/save"
	"github.com/aifedorov/shortener/internal/storage"
)

var (
	ErrShortURLMissing = errors.New("short URL is missing")
)

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

func (s *Server) ListenAndServe() {
	s.config.ParseFlags()
	s.mountHandlers()

	fmt.Println("Running server on", s.config.RunAddr)
	err := http.ListenAndServe(s.config.RunAddr, s.router)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) mountHandlers() {
	s.router.Post("/", save.NewURLSaveHandler(s.config, s.store))
	s.router.Get("/{shortURL}", redirect.NewRedirectHandler(s.store))
	s.router.Get("/", func(res http.ResponseWriter, r *http.Request) {
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
	})
}
