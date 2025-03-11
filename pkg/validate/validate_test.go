package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name       string
		urlChecker URLChecker
		url        string
		hasErr     bool
	}{
		{
			name:       "url is empty",
			urlChecker: NewService(),
			url:        "",
			hasErr:     false,
		},
		{
			name:       "url is not valid",
			urlChecker: NewService(),
			url:        "abc",
			hasErr:     true,
		},
		{
			name:       "url is valid: https://google.com",
			urlChecker: NewService(),
			url:        "https://google.com",
			hasErr:     false,
		},
		{
			name:       "url is valid: http://google.com",
			urlChecker: NewService(),
			url:        "http://google.com",
			hasErr:     false,
		},
		{
			name:       "complex url is valid: https://google.tr.com",
			urlChecker: NewService(),
			url:        "https://google.tr.com",
			hasErr:     false,
		},
		{
			name:       "complex url is valid: https://google.tr.com/",
			urlChecker: NewService(),
			url:        "https://google.tr.com/",
			hasErr:     false,
		},
		{
			name:       "complex url is valid: https://google.tr.com/2a/2b/2c",
			urlChecker: NewService(),
			url:        "https://google.tr.com/2a/2b/2c",
			hasErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.urlChecker.CheckURL(tt.url)
			assert.Equal(t, tt.hasErr, err != nil)
		})
	}
}
