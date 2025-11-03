package server

import (
	"context"
	"errors"
	"net"
	"testing"

	pb "github.com/aifedorov/shortener/api/grpc/gen/shortener/v1"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestShortenerServer_CreateShortURL(t *testing.T) {
	tests := []struct {
		name      string
		request   *pb.CreateShortURLRequest
		setupMock func(*mocks.MockRepository, *mocks.MockURLChecker)
		wantResp  *pb.CreateShortURLResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success",
			request: &pb.CreateShortURLRequest{
				Url: "https://example.com",
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().Store("1", "http://localhost:8080", "https://example.com").
					Return("http://localhost:8080/abc123", nil)
			},
			wantResp: &pb.CreateShortURLResponse{
				ShortUrl: "http://localhost:8080/abc123",
			},
			wantErr: false,
		},
		{
			name: "invalid url",
			request: &pb.CreateShortURLRequest{
				Url: "not-a-valid-url",
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("not-a-valid-url").Return(errors.New("invalid url"))
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "empty url",
			request: &pb.CreateShortURLRequest{
				Url: "",
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("").Return(errors.New("empty url"))
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "url already exists - conflict",
			request: &pb.CreateShortURLRequest{
				Url: "https://example.com",
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().Store("1", "http://localhost:8080", "https://example.com").
					Return("", &repository.ConflictError{ShortURL: "http://localhost:8080/existing"})
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "repository internal error",
			request: &pb.CreateShortURLRequest{
				Url: "https://example.com",
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().Store("1", "http://localhost:8080", "https://example.com").
					Return("", errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo, mockChecker)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.CreateShortURL(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResp.ShortUrl, resp.ShortUrl)
			}
		})
	}
}

func TestShortenerServer_BatchCreateShortURL(t *testing.T) {
	tests := []struct {
		name      string
		request   *pb.BatchCreateShortURLRequest
		setupMock func(*mocks.MockRepository, *mocks.MockURLChecker)
		wantResp  *pb.BatchCreateShortURLResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success - single url",
			request: &pb.BatchCreateShortURLRequest{
				Urls: []*pb.BatchCreateShortURLRequest_URLInput{
					{Cid: "1", OriginalUrl: "https://example.com"},
				},
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().StoreBatch("1", "http://localhost:8080", gomock.Any()).
					DoAndReturn(func(userID, baseURL string, urls []repository.BatchURLInput) ([]repository.BatchURLOutput, error) {
						return []repository.BatchURLOutput{
							{CID: "1", ShortURL: "http://localhost:8080/abc123"},
						}, nil
					})
			},
			wantResp: &pb.BatchCreateShortURLResponse{
				Urls: []*pb.BatchCreateShortURLResponse_URLOutput{
					{Cid: "1", ShortUrl: "http://localhost:8080/abc123"},
				},
			},
			wantErr: false,
		},
		{
			name: "success - multiple urls",
			request: &pb.BatchCreateShortURLRequest{
				Urls: []*pb.BatchCreateShortURLRequest_URLInput{
					{Cid: "1", OriginalUrl: "https://example.com"},
					{Cid: "2", OriginalUrl: "https://google.com"},
				},
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				checker.EXPECT().CheckURL("https://google.com").Return(nil)
				repo.EXPECT().StoreBatch("1", "http://localhost:8080", gomock.Any()).
					DoAndReturn(func(userID, baseURL string, urls []repository.BatchURLInput) ([]repository.BatchURLOutput, error) {
						return []repository.BatchURLOutput{
							{CID: "1", ShortURL: "http://localhost:8080/abc123"},
							{CID: "2", ShortURL: "http://localhost:8080/def456"},
						}, nil
					})
			},
			wantResp: &pb.BatchCreateShortURLResponse{
				Urls: []*pb.BatchCreateShortURLResponse_URLOutput{
					{Cid: "1", ShortUrl: "http://localhost:8080/abc123"},
					{Cid: "2", ShortUrl: "http://localhost:8080/def456"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid url in batch",
			request: &pb.BatchCreateShortURLRequest{
				Urls: []*pb.BatchCreateShortURLRequest_URLInput{
					{Cid: "1", OriginalUrl: "https://example.com"},
					{Cid: "2", OriginalUrl: "invalid-url"},
				},
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				checker.EXPECT().CheckURL("invalid-url").Return(errors.New("invalid url"))
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "repository conflict error",
			request: &pb.BatchCreateShortURLRequest{
				Urls: []*pb.BatchCreateShortURLRequest_URLInput{
					{Cid: "1", OriginalUrl: "https://example.com"},
				},
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().StoreBatch("1", "http://localhost:8080", gomock.Any()).
					Return(nil, &repository.ConflictError{ShortURL: "http://localhost:8080/existing"})
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "repository internal error",
			request: &pb.BatchCreateShortURLRequest{
				Urls: []*pb.BatchCreateShortURLRequest_URLInput{
					{Cid: "1", OriginalUrl: "https://example.com"},
				},
			},
			setupMock: func(repo *mocks.MockRepository, checker *mocks.MockURLChecker) {
				checker.EXPECT().CheckURL("https://example.com").Return(nil)
				repo.EXPECT().StoreBatch("1", "http://localhost:8080", gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo, mockChecker)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.BatchCreateShortURL(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, len(tt.wantResp.Urls), len(resp.Urls))
				for i, url := range tt.wantResp.Urls {
					assert.Equal(t, url.Cid, resp.Urls[i].Cid)
					assert.Equal(t, url.ShortUrl, resp.Urls[i].ShortUrl)
				}
			}
		})
	}
}

func TestShortenerServer_GetShortURL(t *testing.T) {
	tests := []struct {
		name      string
		request   *pb.GetShortURLRequest
		setupMock func(*mocks.MockRepository)
		wantResp  *pb.GetShortURLResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success",
			request: &pb.GetShortURLRequest{
				ShortUrl: "abc123",
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Get("abc123").Return("https://example.com", nil)
			},
			wantResp: &pb.GetShortURLResponse{
				OriginalUrl: "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "url not found",
			request: &pb.GetShortURLRequest{
				ShortUrl: "nonexistent",
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Get("nonexistent").Return("", repository.ErrShortURLNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "url deleted",
			request: &pb.GetShortURLRequest{
				ShortUrl: "deleted",
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Get("deleted").Return("", repository.ErrURLDeleted)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "repository error",
			request: &pb.GetShortURLRequest{
				ShortUrl: "error",
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Get("error").Return("", errors.New("database error"))
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.GetShortURL(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResp.OriginalUrl, resp.OriginalUrl)
			}
		})
	}
}

func TestShortenerServer_Ping(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockRepository)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Ping().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().Ping().Return(errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.Ping(ctx, &pb.PingRequest{})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

func TestShortenerServer_ListShortURLs(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockRepository)
		wantResp  *pb.ListShortURLsResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success - single url",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAll("1", "http://localhost:8080").Return([]repository.URLOutput{
					{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
				}, nil)
			},
			wantResp: &pb.ListShortURLsResponse{
				Urls: []*pb.ListShortURLsResponse_URLItem{
					{ShortUrl: "http://localhost:8080/abc123", OriginalUrl: "https://example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "success - multiple urls",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAll("1", "http://localhost:8080").Return([]repository.URLOutput{
					{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
					{ShortURL: "http://localhost:8080/def456", OriginalURL: "https://google.com"},
				}, nil)
			},
			wantResp: &pb.ListShortURLsResponse{
				Urls: []*pb.ListShortURLsResponse_URLItem{
					{ShortUrl: "http://localhost:8080/abc123", OriginalUrl: "https://example.com"},
					{ShortUrl: "http://localhost:8080/def456", OriginalUrl: "https://google.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "user has no urls",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAll("1", "http://localhost:8080").Return(nil, repository.ErrUserHasNoData)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "repository error",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAll("1", "http://localhost:8080").Return(nil, errors.New("database error"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.ListShortURLs(ctx, &pb.ListShortURLsRequest{})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, len(tt.wantResp.Urls), len(resp.Urls))
				for i, url := range tt.wantResp.Urls {
					assert.Equal(t, url.ShortUrl, resp.Urls[i].ShortUrl)
					assert.Equal(t, url.OriginalUrl, resp.Urls[i].OriginalUrl)
				}
			}
		})
	}
}

func TestShortenerServer_GetStats(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		ctx       context.Context
		setupMock func(*mocks.MockRepository)
		wantResp  *pb.GetStatsResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success - trusted ip",
			cfg: &config.Config{
				RunAddr:       ":8080",
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: "192.168.1.0/24",
			},
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.10"), Port: 12345},
			}),
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetStats().Return(repository.StatsOutput{
					TotalURLs:  100,
					TotalUsers: 10,
				}, nil)
			},
			wantResp: &pb.GetStatsResponse{
				Urls:  100,
				Users: 10,
			},
			wantErr: false,
		},
		{
			name: "no trusted subnet configured",
			cfg: &config.Config{
				RunAddr:       ":8080",
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: "",
			},
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.10"), Port: 12345},
			}),
			setupMock: func(repo *mocks.MockRepository) {
				// No expectations - should fail before repository call
			},
			wantCode: codes.PermissionDenied,
			wantErr:  true,
		},
		{
			name: "ip not in trusted subnet",
			cfg: &config.Config{
				RunAddr:       ":8080",
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: "192.168.1.0/24",
			},
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 12345},
			}),
			setupMock: func(repo *mocks.MockRepository) {
				// No expectations - should fail before repository call
			},
			wantCode: codes.PermissionDenied,
			wantErr:  true,
		},
		{
			name: "no peer in context",
			cfg: &config.Config{
				RunAddr:       ":8080",
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: "192.168.1.0/24",
			},
			ctx: context.Background(),
			setupMock: func(repo *mocks.MockRepository) {
				// No expectations - should fail before repository call
			},
			wantCode: codes.PermissionDenied,
			wantErr:  true,
		},
		{
			name: "repository error",
			cfg: &config.Config{
				RunAddr:       ":8080",
				BaseURL:       "http://localhost:8080",
				TrustedSubnet: "192.168.1.0/24",
			},
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.10"), Port: 12345},
			}),
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetStats().Return(repository.StatsOutput{}, errors.New("database error"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo)

			server := NewShortenerServer(tt.cfg, mockRepo, mockChecker, mockJWT)
			if tt.cfg.TrustedSubnet != "" {
				_, ipnet, err := net.ParseCIDR(tt.cfg.TrustedSubnet)
				if err == nil {
					tt.cfg.TrustedIPNet = ipnet
				}
			}
			resp, err := server.GetStats(tt.ctx, &pb.GetStatsRequest{})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantResp.Urls, resp.Urls)
				assert.Equal(t, tt.wantResp.Users, resp.Users)
			}
		})
	}
}

func TestShortenerServer_DeleteURLs(t *testing.T) {
	tests := []struct {
		name      string
		request   *pb.DeleteURLsRequest
		setupMock func(*mocks.MockRepository)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "success - single url",
			request: &pb.DeleteURLsRequest{
				ShortUrls: []string{"abc123"},
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteBatch("1", []string{"abc123"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - multiple urls",
			request: &pb.DeleteURLsRequest{
				ShortUrls: []string{"abc123", "def456", "ghi789"},
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteBatch("1", []string{"abc123", "def456", "ghi789"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty list",
			request: &pb.DeleteURLsRequest{
				ShortUrls: []string{},
			},
			setupMock: func(repo *mocks.MockRepository) {
				// No expectations - should return early
			},
			wantErr: false,
		},
		{
			name: "repository error",
			request: &pb.DeleteURLsRequest{
				ShortUrls: []string{"abc123"},
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteBatch("1", []string{"abc123"}).Return(errors.New("database error"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockChecker := mocks.NewMockURLChecker(ctrl)
			mockJWT := mocks.NewMockJWT(ctrl)
			tt.setupMock(mockRepo)

			server := NewShortenerServer(newMockConfig(), mockRepo, mockChecker, mockJWT)
			ctx := context.WithValue(context.Background(), "userID", "1")
			resp, err := server.DeleteURLs(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

func TestNewShortenerServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := newMockConfig()
	mockRepo := mocks.NewMockRepository(ctrl)
	mockChecker := mocks.NewMockURLChecker(ctrl)
	mockJWT := mocks.NewMockJWT(ctrl)

	server := NewShortenerServer(cfg, mockRepo, mockChecker, mockJWT)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, mockRepo, server.repo)
	assert.Equal(t, mockChecker, server.urlChecker)
}

func newMockConfig() *config.Config {
	return &config.Config{
		RunAddr:       ":8080",
		GRPCAddr:      ":9090",
		BaseURL:       "http://localhost:8080",
		LogLevel:      "info",
		TrustedSubnet: "",
	}
}
