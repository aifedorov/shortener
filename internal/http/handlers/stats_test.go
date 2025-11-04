package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setupMock      func(*mocks.MockRepository)
		expectedStatus int
		expectedBody   *StatsResponse
		checkError     bool
	}{
		{
			name:   "GET - successful stats retrieval",
			method: http.MethodGet,
			setupMock: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					GetStats().
					Return(repository.StatsOutput{
						TotalURLs:  100,
						TotalUsers: 50,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &StatsResponse{
				URLs:  100,
				Users: 50,
			},
			checkError: false,
		},
		{
			name:   "GET - zero stats",
			method: http.MethodGet,
			setupMock: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					GetStats().
					Return(repository.StatsOutput{
						TotalURLs:  0,
						TotalUsers: 0,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &StatsResponse{
				URLs:  0,
				Users: 0,
			},
			checkError: false,
		},
		{
			name:   "GET - large numbers",
			method: http.MethodGet,
			setupMock: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					GetStats().
					Return(repository.StatsOutput{
						TotalURLs:  1000000,
						TotalUsers: 50000,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &StatsResponse{
				URLs:  1000000,
				Users: 50000,
			},
			checkError: false,
		},
		{
			name:   "GET - repository error",
			method: http.MethodGet,
			setupMock: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					GetStats().
					Return(repository.StatsOutput{}, errors.New("database connection failed")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "POST - method not allowed",
			method:         http.MethodPost,
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "PUT - method not allowed",
			method:         http.MethodPut,
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "DELETE - method not allowed",
			method:         http.MethodDelete,
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "PATCH - method not allowed",
			method:         http.MethodPatch,
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tt.setupMock(mockRepo)

			handler := NewStatsHandler(mockRepo)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/internal/stats", nil)
			rr := httptest.NewRecorder()

			// Execute
			handler(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "Status code mismatch")

			if !tt.checkError {
				// Check Content-Type header for successful requests
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type header mismatch")

				// Decode and verify response
				var response StatsResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err, "Failed to decode response")

				assert.Equal(t, tt.expectedBody.URLs, response.URLs, "URLs count mismatch")
				assert.Equal(t, tt.expectedBody.Users, response.Users, "Users count mismatch")
			}
		})
	}
}
