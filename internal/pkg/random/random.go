package random

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// ShortURLDefaultSize defines the default length for generated short URL strings.
const ShortURLDefaultSize = 8

// Randomizer defines the interface for generating random strings.
// It provides a method to generate cryptographically secure random strings
// suitable for use as short URL identifiers.
type Randomizer interface {
	// GenRandomString generates a random string of the configured length.
	// Returns an error if the generation fails.
	GenRandomString() (string, error)
}

// Service provides random string generation functionality for short URLs.
// It implements the Randomizer interface using cryptographic random number generation.
type Service struct {
	// ShortURLSize specifies the length of generated random strings.
	ShortURLSize int
}

// NewService creates a new random string service with the default short URL size.
// The service is ready to use immediately after creation.
func NewService() *Service {
	return &Service{
		ShortURLSize: ShortURLDefaultSize,
	}
}

// GenRandomString generates a cryptographically secure random string of the configured length.
// The string is base64 URL-encoded and suitable for use in short URLs.
func (s *Service) GenRandomString() (string, error) {
	if s.ShortURLSize <= 0 || s.ShortURLSize > md5.Size {
		return "", fmt.Errorf("random: invalid size %d, must be > 0 and <= %d", s.ShortURLSize, md5.Size)
	}
	b := make([]byte, s.ShortURLSize)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	hash := md5.Sum(b)
	return base64.RawURLEncoding.EncodeToString(hash[:s.ShortURLSize]), nil
}
