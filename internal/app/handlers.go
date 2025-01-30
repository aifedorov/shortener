package app

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	ErrShortURLMissing  = errors.New("short URL is missing")
	ErrMethodNotAllowed = errors.New("method is not allowed")
)

type Server struct {
	mux       *http.ServeMux
	pathToURL map[string]string
}

func NewServer() *Server {
	return &Server{
		mux:       http.NewServeMux(),
		pathToURL: make(map[string]string),
	}
}

func (s *Server) ListenAndServe() {
	s.mux.HandleFunc("/", s.shortURLHandler)
	err := http.ListenAndServe(":8080", s.mux)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) shortURLHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		s.methodPostHandler(res, req)
	case http.MethodGet:
		s.methodGetHandler(res, req)
	default:
		http.Error(res, ErrMethodNotAllowed.Error(), http.StatusBadRequest)
	}
}

func (s *Server) methodPostHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	host := req.Host
	shortURL := genShortURL(string(body))
	resURL := fmt.Sprintf("http://%s/%s", host, shortURL)

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
	path := strings.TrimPrefix(req.URL.Path, "/")
	if path == "" {
		http.Error(res, ErrShortURLMissing.Error(), http.StatusBadRequest)
		return
	}

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
