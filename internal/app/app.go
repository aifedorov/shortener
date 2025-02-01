package app

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aifedorov/shortener/cmd/config"
)

var (
	ErrShortURLMissing = errors.New("short URL is missing")
)

type Server struct {
	router    *chi.Mux
	pathToURL map[string]string
	config    *config.Config
}

func NewServer() *Server {
	return &Server{
		router:    chi.NewRouter(),
		pathToURL: make(map[string]string),
		config:    new(config.Config),
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
	s.router.Post("/", s.methodPostHandler)
	s.router.Get("/{shortURL}", s.methodGetHandler)
	s.router.Get("/", func(res http.ResponseWriter, r *http.Request) {
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
	})
}

func (s *Server) methodPostHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	host := s.config.ShortURLAddr
	shortURL := genShortURL(string(body))
	resURL := host + "/" + shortURL

	if _, ok := s.pathToURL[shortURL]; ok {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte(resURL))

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	s.pathToURL[shortURL] = string(body)
	res.WriteHeader(http.StatusCreated)

	_, writeErr := res.Write([]byte(resURL))
	if writeErr != nil {
		http.Error(res, writeErr.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) methodGetHandler(res http.ResponseWriter, req *http.Request) {
	path := chi.URLParam(req, "shortURL")

	targetURL, exists := s.pathToURL[path]
	if !exists {
		http.NotFound(res, req)
		return
	}

	http.Redirect(res, req, targetURL, http.StatusTemporaryRedirect)
}

func genShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}
