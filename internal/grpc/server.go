package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	pb "github.com/aifedorov/shortener/api/grpc/gen/shortener/v1"
	"github.com/aifedorov/shortener/internal/config"
	urlDomain "github.com/aifedorov/shortener/internal/domain/url"
	userDomain "github.com/aifedorov/shortener/internal/domain/user"
	"github.com/aifedorov/shortener/internal/grpc/middleware/auth"
	"github.com/aifedorov/shortener/internal/grpc/middleware/ipcheck"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	errMsgInvalidURL      = "invalid url"
	errMsgUnauthenticated = "unauthenticated"
	errMsgNotFound        = "not found"
	errMsgURLDeleted      = "url has been deleted"
	errMsgFailedToCreate  = "failed to create url"
	errMsgFailedToPing    = "failed to ping"
	errMsgFailedToGetURLs = "failed to get urls"
	errMsgFailedToStats   = "failed to get stats"
	errMsgInternalError   = "internal server error"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	cfg         *config.Config
	grpc        *grpc.Server
	repo        repository.Repository
	urlService  urlDomain.Service
	userService userDomain.Service
	jwtChecker  jwt.JWT
}

func NewShortenerServer(
	cfg *config.Config,
	repo repository.Repository,
	urlService urlDomain.Service,
	userService userDomain.Service,
	jwtChecker jwt.JWT,
) *ShortenerServer {
	return &ShortenerServer{
		cfg:         cfg,
		repo:        repo,
		urlService:  urlService,
		userService: userService,
		jwtChecker:  jwtChecker,
	}
}

func (s *ShortenerServer) Run() error {
	listen, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	authInterceptor := auth.NewInterceptor(s.jwtChecker, s.userService)
	ipcheckInterceptor := ipcheck.NewInterceptor(s.cfg.TrustedIPNet)

	s.grpc = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.UnaryAuthInterceptor,
			ipcheckInterceptor.UnaryIPCheckInterceptor,
		),
	)

	pb.RegisterShortenerServiceServer(s.grpc, s)

	reflection.Register(s.grpc)

	return s.grpc.Serve(listen)
}

func (s *ShortenerServer) CreateShortURL(ctx context.Context, request *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	url := request.GetUrl()
	if err := s.urlService.ValidateURL(url); err != nil {
		logger.Log.Error("grpc: create short url: invalid url", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidURL)
	}

	userID, err := s.userService.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, errMsgUnauthenticated)
	}

	resURL, err := s.repo.Store(userID, s.cfg.BaseURL, url)
	var cErr *repository.ConflictError
	if errors.As(err, &cErr) {
		return nil, status.Error(codes.AlreadyExists, cErr.ShortURL)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, errMsgFailedToCreate)
	}
	return &pb.CreateShortURLResponse{ShortUrl: resURL}, nil
}

func (s *ShortenerServer) BatchCreateShortURL(ctx context.Context, request *pb.BatchCreateShortURLRequest) (*pb.BatchCreateShortURLResponse, error) {
	urls := request.GetUrls()
	inputURLs := make([]repository.BatchURLInput, len(urls))

	for i, url := range urls {
		if err := s.urlService.ValidateURL(url.OriginalUrl); err != nil {
			logger.Log.Error("grpc: batch create short url: invalid url", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, errMsgInvalidURL)
		}

		inputURLs[i] = repository.BatchURLInput{
			CID:         url.Cid,
			OriginalURL: url.OriginalUrl,
		}
	}

	userID, err := s.userService.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, errMsgUnauthenticated)
	}

	storedURLs, err := s.repo.StoreBatch(userID, s.cfg.BaseURL, inputURLs)
	var cErr *repository.ConflictError
	if errors.As(err, &cErr) {
		return nil, status.Error(codes.AlreadyExists, cErr.ShortURL)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, errMsgFailedToCreate)
	}

	resURLs := make([]*pb.BatchCreateShortURLResponse_URLOutput, len(storedURLs))
	for i, url := range storedURLs {
		resURLs[i] = &pb.BatchCreateShortURLResponse_URLOutput{
			Cid:      url.CID,
			ShortUrl: url.ShortURL,
		}
	}

	return &pb.BatchCreateShortURLResponse{
		Urls: resURLs,
	}, nil
}

func (s *ShortenerServer) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.repo.Ping()
	if err != nil {
		logger.Log.Error("grpc: failed to ping repository", zap.Error(err))
		return nil, status.Error(codes.Internal, errMsgFailedToPing)
	}
	return &pb.PingResponse{}, nil
}

func (s *ShortenerServer) GetShortURL(_ context.Context, request *pb.GetShortURLRequest) (*pb.GetShortURLResponse, error) {
	shortURL := request.GetShortUrl()
	url, err := s.repo.Get(shortURL)
	if errors.Is(err, repository.ErrShortURLNotFound) {
		logger.Log.Info("redirect: short url not found", zap.String("alias", shortURL))
		return nil, status.Error(codes.NotFound, errMsgNotFound)
	}
	if errors.Is(err, repository.ErrURLDeleted) {
		logger.Log.Info("redirect: url deleted", zap.String("alias", shortURL))
		return nil, status.Error(codes.NotFound, errMsgURLDeleted)
	}
	if err != nil {
		logger.Log.Error("grpc: get short url: failed to get url", zap.Error(err))
		return nil, status.Error(codes.NotFound, errMsgNotFound)
	}
	return &pb.GetShortURLResponse{OriginalUrl: url}, nil
}

func (s *ShortenerServer) GetStats(_ context.Context, _ *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.repo.GetStats()
	if err != nil {
		logger.Log.Error("grpc: failed to get stats", zap.Error(err))
		return nil, status.Error(codes.Internal, errMsgFailedToStats)
	}

	return &pb.GetStatsResponse{
		Urls:  uint64(stats.TotalURLs),
		Users: uint64(stats.TotalUsers),
	}, nil
}

func (s *ShortenerServer) ListShortURLs(ctx context.Context, _ *pb.ListShortURLsRequest) (*pb.ListShortURLsResponse, error) {
	userID, err := s.userService.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, errMsgUnauthenticated)
	}

	urls, err := s.repo.GetAll(userID, s.cfg.BaseURL)
	if errors.Is(err, repository.ErrUserHasNoData) {
		logger.Log.Info("user don't have any urls", zap.String("user_id", userID))
		return nil, status.Error(codes.NotFound, errMsgNotFound)
	}
	if err != nil {
		logger.Log.Error("grpc: list short urls: failed to get urls", zap.Error(err))
		return nil, status.Error(codes.Internal, errMsgFailedToGetURLs)
	}

	resURLs := make([]*pb.ListShortURLsResponse_URLItem, len(urls))
	for i, url := range urls {
		resURLs[i] = &pb.ListShortURLsResponse_URLItem{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		}
	}

	return &pb.ListShortURLsResponse{
		Urls: resURLs,
	}, nil
}

func (s *ShortenerServer) DeleteURLs(ctx context.Context, request *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	userID, err := s.userService.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, errMsgUnauthenticated)
	}

	shortURLs := request.GetShortUrls()
	if len(shortURLs) == 0 {
		logger.Log.Warn("grpc: delete urls: empty list", zap.String("user_id", userID))
		return &pb.DeleteURLsResponse{}, nil
	}

	err = s.repo.DeleteBatch(userID, shortURLs)
	if err != nil {
		logger.Log.Error("grpc: delete urls: failed to delete urls", zap.Error(err), zap.String("user_id", userID))
		return nil, status.Error(codes.Internal, errMsgInternalError)
	}

	logger.Log.Info("grpc: delete urls: successfully deleted", zap.Int("count", len(shortURLs)), zap.String("user_id", userID))
	return &pb.DeleteURLsResponse{}, nil
}

func (s *ShortenerServer) Shutdown() {
	s.grpc.GracefulStop()
}
