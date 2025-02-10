package validate

import (
	"errors"
	"regexp"
)

const (
	urlPatternString = `(https?:\/\/)([\da-z\.-]+)\.([a-z]{2,6})(\/[\da-z\.-]+)*`
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
