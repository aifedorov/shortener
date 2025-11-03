package handlers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aifedorov/shortener/internal/config"
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
		realIP         string
		trustedSubnet  string
		setupMock      func(*mocks.MockRepository)
		expectedStatus int
		expectedBody   *StatsResponse
		checkError     bool
	}{
		{
			name:          "GET - successful stats retrieval with trusted IP",
			method:        http.MethodGet,
			realIP:        "192.168.1.10",
			trustedSubnet: "192.168.1.0/24",
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
			name:          "GET - zero stats with trusted IP",
			method:        http.MethodGet,
			realIP:        "10.0.0.5",
			trustedSubnet: "10.0.0.0/24",
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
			name:          "GET - large numbers with trusted IP",
			method:        http.MethodGet,
			realIP:        "172.16.0.100",
			trustedSubnet: "172.16.0.0/16",
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
			name:          "GET - repository error with trusted IP",
			method:        http.MethodGet,
			realIP:        "192.168.1.1",
			trustedSubnet: "192.168.1.0/24",
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
			name:           "GET - IP not in trusted subnet",
			method:         http.MethodGet,
			realIP:         "10.0.0.1",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "GET - empty trusted subnet (all requests forbidden)",
			method:         http.MethodGet,
			realIP:         "192.168.1.10",
			trustedSubnet:  "",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "GET - missing X-Real-IP header",
			method:         http.MethodGet,
			realIP:         "",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "GET - invalid IP address format",
			method:         http.MethodGet,
			realIP:         "invalid-ip",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "GET - IPv6 address not in IPv4 subnet",
			method:         http.MethodGet,
			realIP:         "2001:db8::1",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "GET - IP outside subnet range",
			method:         http.MethodGet,
			realIP:         "192.168.2.1",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "POST - method not allowed (even with trusted IP)",
			method:         http.MethodPost,
			realIP:         "192.168.1.10",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "PUT - method not allowed",
			method:         http.MethodPut,
			realIP:         "192.168.1.10",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "DELETE - method not allowed",
			method:         http.MethodDelete,
			realIP:         "192.168.1.10",
			trustedSubnet:  "192.168.1.0/24",
			setupMock:      func(mockRepo *mocks.MockRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			checkError:     true,
		},
		{
			name:           "PATCH - method not allowed",
			method:         http.MethodPatch,
			realIP:         "192.168.1.10",
			trustedSubnet:  "192.168.1.0/24",
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

			cfg := &config.Config{
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: tt.trustedSubnet,
			}

			// Parse the trusted subnet into IPNet if provided
			if tt.trustedSubnet != "" {
				_, ipnet, err := net.ParseCIDR(tt.trustedSubnet)
				if err == nil {
					cfg.TrustedIPNet = ipnet
				}
			}

			handler := NewStatsHandler(cfg, mockRepo)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/internal/stats", nil)

			// Set X-Real-IP header if provided
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

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
