package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		urlChecker URLChecker
		url        string
		wantErr    error
	}{
		{
			name:       "url is empty",
			urlChecker: NewService(),
			url:        "",
			wantErr:    nil,
		},
		{
			name:       "url is not valid",
			urlChecker: NewService(),
			url:        "abc",
			wantErr:    ErrURLInvalid,
		},
		{
			name:       "url is valid: https://google.com",
			urlChecker: NewService(),
			url:        "https://google.com",
			wantErr:    nil,
		},
		{
			name:       "url is valid: http://google.com",
			urlChecker: NewService(),
			url:        "http://google.com",
			wantErr:    nil,
		},
		{
			name:       "complex url is valid: https://google.tr.com",
			urlChecker: NewService(),
			url:        "https://google.tr.com",
			wantErr:    nil,
		},
		{
			name:       "complex url is valid: https://google.tr.com/",
			urlChecker: NewService(),
			url:        "https://google.tr.com/",
			wantErr:    nil,
		},
		{
			name:       "complex url is valid: https://google.tr.com/2a/2b/2c",
			urlChecker: NewService(),
			url:        "https://google.tr.com/2a/2b/2c",
			wantErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.urlChecker.CheckURL(tt.url)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
