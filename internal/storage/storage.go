package storage

import (
	"errors"
)

var (
	ErrShortURLNotFound = errors.New("storage: short url not found")
	ErrURLExists        = errors.New("storage: url exists")
	ErrGenShortURL      = errors.New("storage: generation short url is failed")
)

type Storage interface {
	GetURL(shortURL string) (string, error)
	SaveURL(baseURL, targetURL string) (string, error)
}
