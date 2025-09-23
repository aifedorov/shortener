package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDecodeRequest(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expected    RequestBody
		expectError bool
	}{
		{
			name:        "valid JSON",
			requestBody: `{"url": "https://example.com"}`,
			expected:    RequestBody{URL: "https://example.com"},
			expectError: false,
		},
		{
			name:        "empty body",
			requestBody: ``,
			expected:    RequestBody{},
			expectError: true,
		},
		{
			name:        "invalid JSON",
			requestBody: `{"url": "https://example.com"`,
			expected:    RequestBody{},
			expectError: true,
		},
		{
			name:        "empty URL",
			requestBody: `{"url": ""}`,
			expected:    RequestBody{URL: ""},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			result, err := decodeRequest(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDecodeBatchRequest(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expected    []BatchRequest
		expectError bool
	}{
		{
			name:        "valid batch JSON",
			requestBody: `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			expected:    []BatchRequest{{CID: "1", OriginalURL: "https://example.com"}},
			expectError: false,
		},
		{
			name:        "empty array",
			requestBody: `[]`,
			expected:    []BatchRequest{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			requestBody: `[{"correlation_id": "1", "original_url": "https://example.com"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty body",
			requestBody: ``,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			result, err := decodeBatchRequest(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDecodeAliasesRequest(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expected    []string
		expectError bool
	}{
		{
			name:        "valid aliases JSON",
			requestBody: `["abc123", "def456"]`,
			expected:    []string{"abc123", "def456"},
			expectError: false,
		},
		{
			name:        "empty array",
			requestBody: `[]`,
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			requestBody: `["abc123", "def456"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty body",
			requestBody: ``,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			result, err := decodeAliasesRequest(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		expectError bool
	}{
		{
			name:        "valid user ID",
			userID:      "user123",
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			expectError: true,
		},
		{
			name:        "no user ID in context",
			userID:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.name != "no user ID in context" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			result, err := getUserID(req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, result)
			}
		})
	}
}

func TestValidateURLs(t *testing.T) {
	tests := []struct {
		name        string
		reqURLs     []BatchRequest
		checkError  error
		expected    []repository.BatchURLInput
		expectError bool
	}{
		{
			name: "valid URLs",
			reqURLs: []BatchRequest{
				{CID: "1", OriginalURL: "https://example.com"},
				{CID: "2", OriginalURL: "https://google.com"},
			},
			checkError: nil,
			expected: []repository.BatchURLInput{
				{CID: "1", OriginalURL: "https://example.com"},
				{CID: "2", OriginalURL: "https://google.com"},
			},
			expectError: false,
		},
		{
			name: "invalid URL",
			reqURLs: []BatchRequest{
				{CID: "1", OriginalURL: "invalid-url"},
			},
			checkError:  errors.New("invalid URL"),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty URLs",
			reqURLs:     []BatchRequest{},
			checkError:  nil,
			expected:    []repository.BatchURLInput{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockURLChecker := mocks.NewMockURLChecker(ctrl)
			for _, reqURL := range tt.reqURLs {
				mockURLChecker.EXPECT().CheckURL(reqURL.OriginalURL).Return(tt.checkError)
			}

			result, err := validateURLs(tt.reqURLs, mockURLChecker)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
