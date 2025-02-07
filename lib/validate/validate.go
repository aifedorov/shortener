package validate

import (
	"errors"
	"regexp"
)

const (
	urlPatternString = `^https?:\/\/[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var (
	ErrURLMissing = errors.New("URL is missing")
	ErrURLInvalid = errors.New("URL is not valid")
	urlPattern    = regexp.MustCompile(urlPatternString)
)

func ValidateURL(url string) error {
	if url == "" {
		return ErrURLMissing
	}

	if !urlPattern.MatchString(url) {
		return ErrURLInvalid
	}

	return nil
}
