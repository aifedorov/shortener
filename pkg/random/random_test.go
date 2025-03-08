package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenRandomString(t *testing.T) {
	tests := []struct {
		name       string
		randomizer Randomizer
		str        string
		wantErr    bool
	}{
		{
			name: "size is zero value",
			randomizer: &Service{
				ShortURLSize: 0,
			},
			str:     "abc",
			wantErr: true,
		},
		{
			name: "size is negative value",
			randomizer: &Service{
				ShortURLSize: -1,
			},
			str:     "abc",
			wantErr: true,
		},
		{
			name: "size is valid",
			randomizer: &Service{
				ShortURLSize: 2,
			},
			str:     "abc",
			wantErr: false,
		},
		{
			name: "size is too big value",
			randomizer: &Service{
				ShortURLSize: 33,
			},
			str:     "abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.randomizer.GenRandomString(tt.str)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
