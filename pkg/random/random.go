package random

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const ShortURLDefaultSize = 8

type Randomizer interface {
	GenRandomString() (string, error)
}

type Service struct {
	ShortURLSize int
}

func NewService() *Service {
	return &Service{
		ShortURLSize: ShortURLDefaultSize,
	}
}

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
