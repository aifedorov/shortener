package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	pb "github.com/aifedorov/shortener/api/grpc/gen/shortener/v1"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/grpc/mw"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	cfg        *config.Config
	grpc       *grpc.Server
	repo       repository.Repository
	urlChecker validate.URLChecker
	ipnet      *net.IPNet
}

func NewShortenerServer(cfg *config.Config, repo repository.Repository, urlChecker validate.URLChecker) *ShortenerServer {
	return &ShortenerServer{
		cfg:        cfg,
		repo:       repo,
		urlChecker: urlChecker,
	}
}

func (s *ShortenerServer) Run() error {
	_, ipnet, err := net.ParseCIDR(s.cfg.TrustedSubnet)
	if err != nil {
		logger.Log.Error("grpc: failed to parse trusted subnet", zap.Error(err))
		return fmt.Errorf("grpc: failed to parse trusted subnet: %w", err)
	}
	s.ipnet = ipnet

	listen, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.grpc = grpc.NewServer(grpc.UnaryInterceptor(mw.AuthJWTInterceptor))
	pb.RegisterShortenerServiceServer(s.grpc, s)

	reflection.Register(s.grpc)

	return s.grpc.Serve(listen)
}

func (s *ShortenerServer) CreateShortURL(ctx context.Context, request *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	url := request.GetUrl()
	if err := s.urlChecker.CheckURL(url); err != nil {
		logger.Log.Error("grpc: create short url: invalid url", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	// TODO: add userID and auth.
	resURL, err := s.repo.Store("1", s.cfg.BaseURL, url)
	var cErr *repository.ConflictError
	if errors.As(err, &cErr) {
		return nil, status.Error(codes.AlreadyExists, cErr.ShortURL)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create url")
	}
	return &pb.CreateShortURLResponse{ShortUrl: resURL}, nil
}

func (s *ShortenerServer) BatchCreateShortURL(ctx context.Context, request *pb.BatchCreateShortURLRequest) (*pb.BatchCreateShortURLResponse, error) {
	urls := request.GetUrls()
	inputURLs := make([]repository.BatchURLInput, len(urls))

	for i, url := range urls {
		if err := s.urlChecker.CheckURL(url.OriginalUrl); err != nil {
			logger.Log.Error("grpc: batch create short url: invalid url", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "invalid url")
		}

		inputURLs[i] = repository.BatchURLInput{
			CID:         url.Cid,
			OriginalURL: url.OriginalUrl,
		}
	}

	// TODO: Add userID and auth.
	storedURLs, err := s.repo.StoreBatch("1", s.cfg.BaseURL, inputURLs)
	var cErr *repository.ConflictError
	if errors.As(err, &cErr) {
		return nil, status.Error(codes.AlreadyExists, cErr.ShortURL)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create url")
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

func (s *ShortenerServer) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.repo.Ping()
	if err != nil {
		logger.Log.Error("grpc: failed to ping repository", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to ping")
	}
	return &pb.PingResponse{}, nil
}

func (s *ShortenerServer) GetShortURL(ctx context.Context, request *pb.GetShortURLRequest) (*pb.GetShortURLResponse, error) {
	shortURL := request.GetShortUrl()
	url, err := s.repo.Get(shortURL)
	if errors.Is(err, repository.ErrShortURLNotFound) {
		logger.Log.Info("redirect: short url not found", zap.String("alias", shortURL))
		return nil, status.Error(codes.NotFound, "not found")
	}
	if errors.Is(err, repository.ErrURLDeleted) {
		logger.Log.Info("redirect: url deleted", zap.String("alias", shortURL))
		return nil, status.Error(codes.NotFound, "url has been deleted")
	}
	if err != nil {
		logger.Log.Error("grpc: get short url: failed to get url", zap.Error(err))
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &pb.GetShortURLResponse{OriginalUrl: url}, nil
}

func (s *ShortenerServer) GetStats(ctx context.Context, request *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	if s.cfg.TrustedSubnet == "" {
		logger.Log.Warn("grpc: GetStats called but trusted subnet is not configured")
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Log.Error("grpc: failed to get peer from context")
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	tcpAddr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		logger.Log.Error("grpc: peer address is not TCP")
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	clientIP := tcpAddr.IP

	if s.ipnet == nil || !s.ipnet.Contains(clientIP) {
		logger.Log.Info("grpc: request IP is not in trusted subnet",
			zap.String("ip", clientIP.String()),
			zap.String("subnet", s.cfg.TrustedSubnet))
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	stats, err := s.repo.GetStats()
	if err != nil {
		logger.Log.Error("grpc: failed to get stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return &pb.GetStatsResponse{
		Urls:  uint64(stats.TotalURLs),
		Users: uint64(stats.TotalUsers),
	}, nil
}

func (s *ShortenerServer) ListShortURLs(ctx context.Context, request *pb.ListShortURLsRequest) (*pb.ListShortURLsResponse, error) {
	// TODO: Add userID and auth.
	urls, err := s.repo.GetAll("1", s.cfg.BaseURL)
	if errors.Is(err, repository.ErrUserHasNoData) {
		logger.Log.Info("user don't have any urls", zap.String("user_id", "1"))
		return nil, status.Error(codes.NotFound, "not found")
	}
	if err != nil {
		logger.Log.Error("grpc: list short urls: failed to get urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get urls")
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

func (s *ShortenerServer) Shutdown() {
	s.grpc.GracefulStop()
}
