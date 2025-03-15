package validate

import (
	"errors"
	"regexp"
)

const (
	urlPatternString = `(https?:\/\/)([\da-z\.-]+)\.([a-z]{2,6})(\/[\da-z\.-]+)*`
)

type URLChecker interface {
	CheckURL(url string) error
}

type Service struct {
	URLPattern *regexp.Regexp
}

func NewService() *Service {
	return &Service{
		URLPattern: regexp.MustCompile(urlPatternString),
	}
}

func (s *Service) CheckURL(url string) error {
	if url != "" && !s.URLPattern.MatchString(url) {
		return errors.New("url is not valid")
	}
	return nil
}
