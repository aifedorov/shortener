package random

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const ShortURLDefaultSize = 8

type Randomizer interface {
	GenRandomString(str string) (string, error)
}

type Service struct {
	ShortURLSize int
}

func NewService() *Service {
	return &Service{
		ShortURLSize: ShortURLDefaultSize,
	}
}

func (s *Service) GenRandomString(str string) (string, error) {
	if s.ShortURLSize <= 0 || s.ShortURLSize > sha256.Size {
		return "", fmt.Errorf("random: invalid size %d, must be > 0 and <= %d", s.ShortURLSize, sha256.Size)
	}
	hash := sha256.Sum256([]byte(str))
	return base64.RawURLEncoding.EncodeToString(hash[:s.ShortURLSize]), nil
}
