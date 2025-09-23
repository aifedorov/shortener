package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewPingHandler(t *testing.T) {
	tests := []struct {
		name           string
		pingError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful ping",
			pingError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "ping failed",
			pingError:      errors.New("connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "database error",
			pingError:      errors.New("database connection lost"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().Ping().Return(tt.pingError)

			handler := NewPingHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
