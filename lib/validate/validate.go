package validate

import (
	"errors"
	"regexp"
)

const (
	urlPatternString = `^https?:\/\/[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\/?$`
)

var (
	ErrURLInvalid = errors.New("URL is not valid")
	urlPattern    = regexp.MustCompile(urlPatternString)
)

func CheckURL(url string) error {
	if url != "" && !urlPattern.MatchString(url) {
		return ErrURLInvalid
	}

	return nil
}
