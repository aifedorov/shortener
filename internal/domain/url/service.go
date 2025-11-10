package url

import (
	"net/url"
	"strings"

	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/aifedorov/shortener/internal/pkg/validate"
)

type Service interface {
	GenerateAlias() (string, error)
	ValidateURL(url string) error
	NormalizeURL(rawURL string) (string, error)
	CanUserDeleteURLs(userID string, urls []*URL) error
	ValidateBatch(items []BatchItem) error
}

// Service contains business logic for working with URLs.
type service struct {
	generator random.Randomizer
	validator validate.URLChecker
}

// NewService creates a new domain service.
func NewService(generator random.Randomizer, validator validate.URLChecker) Service {
	return &service{
		generator: generator,
		validator: validator,
	}
}

// GenerateAlias generates a unique alias for a URL.
func (s *service) GenerateAlias() (string, error) {
	return s.generator.GenRandomString()
}

// ValidateURL checks URL correctness according to business rules.
func (s *service) ValidateURL(url string) error {
	if url == "" {
		return ErrInvalidOriginalURL
	}

	return s.validator.CheckURL(url)
}

// NormalizeURL normalizes a URL (converts to a standard form).
func (s *service) NormalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	return parsedURL.String(), nil
}

// CanUserDeleteURLs checks business rules for URL deletion.
func (s *service) CanUserDeleteURLs(userID string, urls []*URL) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	if len(urls) == 0 {
		return ErrEmptyBatch
	}

	for _, u := range urls {
		if !u.IsOwnedBy(userID) {
			return ErrShortURLNotFound
		}
	}

	return nil
}

// ValidateBatch checks the correctness of a batch operation.
func (s *service) ValidateBatch(items []BatchItem) error {
	if len(items) == 0 {
		return ErrEmptyBatch
	}

	for _, item := range items {
		if item.CorrelationID == "" {
			return ErrInvalidAlias
		}
		if item.URL == nil {
			return ErrInvalidOriginalURL
		}
	}

	return nil
}
