package repository

import (
	"errors"
)

var (
	ErrShortURLNotFound = errors.New("repository: short url not found")
	ErrURLExists        = errors.New("repository: url exists")
	ErrGenShortURL      = errors.New("repository: generation short url is failed")
)

type Repository interface {
	Get(shortURL string) (string, error)
	Store(baseURL, targetURL string) (string, error)
}
