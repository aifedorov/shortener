package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewDeleteHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		userID         string
		deleteError    error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful delete",
			requestBody:    `["abc123", "def456"]`,
			userID:         "user123",
			deleteError:    nil,
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:           "invalid JSON",
			requestBody:    `["abc123", "def456"`,
			userID:         "user123",
			deleteError:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty request body",
			requestBody:    ``,
			userID:         "user123",
			deleteError:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty aliases array",
			requestBody:    `[]`,
			userID:         "user123",
			deleteError:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    `["abc123"]`,
			userID:         "",
			deleteError:    nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "repository error",
			requestBody:    `["abc123"]`,
			userID:         "user123",
			deleteError:    errors.New("database error"),
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:           "single alias",
			requestBody:    `["abc123"]`,
			userID:         "user123",
			deleteError:    nil,
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			if tt.userID != "" {
				var aliases []string
				if err := json.Unmarshal([]byte(tt.requestBody), &aliases); err == nil && len(aliases) > 0 {
					mockRepo.EXPECT().DeleteBatch(tt.userID, aliases).Return(tt.deleteError).AnyTimes()
				}
			}

			handler := NewDeleteHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
			if tt.expectedStatus == http.StatusAccepted {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
			}
		})
	}
}
