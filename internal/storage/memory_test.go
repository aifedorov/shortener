package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_GetURL(t *testing.T) {
	tests := []struct {
		name     string
		storage  *MemoryStorage
		shortURL string
		want     string
		wantErr  error
	}{
		{
			name: "get URL with existing value",
			storage: func() *MemoryStorage {
				ms := NewMemoryStorage()
				ms.PathToURL.Store("1", "https://google.com")
				return ms
			}(),
			shortURL: "1",
			want:     "https://google.com",
			wantErr:  nil,
		},
		{
			name: "get URL with not existing value",
			storage: func() *MemoryStorage {
				ms := NewMemoryStorage()
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
			got, err := tt.storage.GetURL(tt.shortURL)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryStorage_SaveURL(t *testing.T) {
	tests := []struct {
		name       string
		storage    *MemoryStorage
		baseURL    string
		targetURL  string
		wantPrefix string
		want       string
		wantErr    error
	}{
		{
			name:      "save new URL with empty targetURL",
			storage:   NewMemoryStorage(),
			baseURL:   "https://google.com",
			targetURL: "",
			want:      "",
			wantErr:   nil,
		},
		{
			name:       "save new URL with valid targetURL",
			storage:    NewMemoryStorage(),
			baseURL:    "https://localhost:80",
			targetURL:  "https://google.com",
			wantPrefix: "https://localhost:80/",
			wantErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.SaveURL(tt.baseURL, tt.targetURL)
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
