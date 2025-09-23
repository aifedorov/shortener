package validate

import (
	"errors"
	"regexp"
)

const (
	urlPatternString = `(https?:\/\/)([\da-z\.-]+)\.([a-z]{2,6})(\/[\da-z\.-]+)*`
)

// URLChecker defines the interface for URL validation operations.
type URLChecker interface {
	CheckURL(url string) error
}

// Service provides URL validation functionality using regular expressions.
type Service struct {
	URLPattern *regexp.Regexp
}

// NewService creates a new URL validation service with the default URL pattern.
func NewService() *Service {
	return &Service{
		URLPattern: regexp.MustCompile(urlPatternString),
	}
}

// CheckURL validates a URL against the configured pattern.
// Returns an error if the URL is invalid or empty.
func (s *Service) CheckURL(url string) error {
	if url != "" && !s.URLPattern.MatchString(url) {
		return errors.New("url is not valid")
	}
	return nil
}
