package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenRandomString(t *testing.T) {
	tests := []struct {
		name       string
		randomizer Randomizer
		wantErr    bool
	}{
		{
			name: "size is zero value",
			randomizer: &Service{
				ShortURLSize: 0,
			},
			wantErr: true,
		},
		{
			name: "size is negative value",
			randomizer: &Service{
				ShortURLSize: -1,
			},
			wantErr: true,
		},
		{
			name: "size is valid",
			randomizer: &Service{
				ShortURLSize: 2,
			},
			wantErr: false,
		},
		{
			name: "size is too big value",
			randomizer: &Service{
				ShortURLSize: 33,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.randomizer.GenRandomString()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
