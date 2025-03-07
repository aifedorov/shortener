package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_GetURL(t *testing.T) {
	tests := []struct {
		name     string
		storage  *MemoryRepository
		shortURL string
		want     string
		wantErr  error
	}{
		{
			name: "get URL with existing value",
			storage: func() *MemoryRepository {
				ms := NewMemoryRepository()
				ms.PathToURL.Store("1", "https://google.com")
				return ms
			}(),
			shortURL: "1",
			want:     "https://google.com",
			wantErr:  nil,
		},
		{
			name: "get URL with not existing value",
			storage: func() *MemoryRepository {
				ms := NewMemoryRepository()
				ms.PathToURL.Store("1", "https://google.com")
				return ms
			}(),
			shortURL: "2",
			want:     "",
			wantErr:  ErrShortURLNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.Get(tt.shortURL)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryStorage_SaveURL(t *testing.T) {
	tests := []struct {
		name       string
		storage    *MemoryRepository
		baseURL    string
		targetURL  string
		wantPrefix string
		want       string
		wantErr    error
	}{
		{
			name:      "save new URL with empty targetURL",
			storage:   NewMemoryRepository(),
			baseURL:   "https://google.com",
			targetURL: "",
			want:      "",
			wantErr:   nil,
		},
		{
			name:       "save new URL with valid targetURL",
			storage:    NewMemoryRepository(),
			baseURL:    "https://localhost:80",
			targetURL:  "https://google.com",
			wantPrefix: "https://localhost:80/",
			wantErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.Store(tt.baseURL, tt.targetURL)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantPrefix != "" {
				assert.Contains(t, got, tt.wantPrefix)
			}

			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
