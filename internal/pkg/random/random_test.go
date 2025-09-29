package random

import (
	"crypto/md5"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with default size", func(t *testing.T) {
		service := NewService()

		assert.NotNil(t, service)
		assert.Equal(t, ShortURLDefaultSize, service.ShortURLSize)
	})
}

func TestGenRandomString(t *testing.T) {
	tests := []struct {
		name       string
		randomizer Randomizer
		wantErr    bool
		errorMsg   string
	}{
		{
			name: "size is zero value",
			randomizer: &Service{
				ShortURLSize: 0,
			},
			wantErr:  true,
			errorMsg: "random: invalid size 0, must be > 0 and <= 16",
		},
		{
			name: "size is negative value",
			randomizer: &Service{
				ShortURLSize: -1,
			},
			wantErr:  true,
			errorMsg: "random: invalid size -1, must be > 0 and <= 16",
		},
		{
			name: "size is valid minimum",
			randomizer: &Service{
				ShortURLSize: 1,
			},
			wantErr: false,
		},
		{
			name: "size is valid default",
			randomizer: &Service{
				ShortURLSize: 2,
			},
			wantErr: false,
		},
		{
			name: "size is valid maximum",
			randomizer: &Service{
				ShortURLSize: md5.Size,
			},
			wantErr: false,
		},
		{
			name: "size is too big value",
			randomizer: &Service{
				ShortURLSize: md5.Size + 1,
			},
			wantErr:  true,
			errorMsg: "random: invalid size 17, must be > 0 and <= 16",
		},
		{
			name: "size is way too big value",
			randomizer: &Service{
				ShortURLSize: 100,
			},
			wantErr:  true,
			errorMsg: "random: invalid size 100, must be > 0 and <= 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.randomizer.GenRandomString()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.Greater(t, len(result), 0)
			}
		})
	}
}

func TestGenRandomString_Uniqueness(t *testing.T) {
	service := NewService()

	generated := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		result, err := service.GenRandomString()
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Greater(t, len(result), 0)

		assert.False(t, generated[result], "Generated duplicate string: %s", result)
		generated[result] = true
	}

	assert.Equal(t, iterations, len(generated))
}

func TestGenRandomString_ValidCharacters(t *testing.T) {
	service := NewService()

	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for i := 0; i < 50; i++ {
		result, err := service.GenRandomString()
		assert.NoError(t, err)

		for _, char := range result {
			assert.Contains(t, validChars, string(char), "Invalid character '%c' in result: %s", char, result)
		}
	}
}

func TestGenRandomString_DifferentSizes(t *testing.T) {
	sizes := []int{1, 4, 8, 12, 16}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			service := &Service{ShortURLSize: size}

			result, err := service.GenRandomString()
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Greater(t, len(result), 0)
		})
	}
}

func TestGenRandomString_EdgeCases(t *testing.T) {
	t.Run("boundary value 1", func(t *testing.T) {
		service := &Service{ShortURLSize: 1}
		result, err := service.GenRandomString()
		assert.NoError(t, err)
		assert.Greater(t, len(result), 0)
	})

	t.Run("boundary value md5.Size", func(t *testing.T) {
		service := &Service{ShortURLSize: md5.Size}
		result, err := service.GenRandomString()
		assert.NoError(t, err)
		assert.Greater(t, len(result), 0)
	})

	t.Run("boundary value md5.Size + 1", func(t *testing.T) {
		service := &Service{ShortURLSize: md5.Size + 1}
		_, err := service.GenRandomString()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid size")
	})
}
