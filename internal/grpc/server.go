package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/aifedorov/shortener/api/grpc/gen/shortener/v1"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	grpc       *grpc.Server
	repo       repository.Repository
	urlChecker validate.URLChecker
}

func NewShortenerServer(repo repository.Repository, urlChecker validate.URLChecker) *ShortenerServer {
	return &ShortenerServer{
		repo:       repo,
		urlChecker: urlChecker,
	}
}

func (s *ShortenerServer) Run() error {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.grpc = grpc.NewServer()
	pb.RegisterShortenerServiceServer(s.grpc, s)

	// Включаем gRPC reflection для тестирования через grpcurl/grpcui
	reflection.Register(s.grpc)

	return s.grpc.Serve(listen)
}

func (s *ShortenerServer) CreateShortURL(ctx context.Context, request *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortenerServer) BatchCreateShortURL(ctx context.Context, request *pb.BatchCreateShortURLRequest) (*pb.BatchCreateShortURLResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortenerServer) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.repo.Ping()
	if err != nil {
		logger.Log.Error("ping: failed to ping repository", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to ping")
	}
	return &pb.PingResponse{}, nil
}

func (s *ShortenerServer) GetShortURL(ctx context.Context, request *pb.GetShortURLRequest) (*pb.GetShortURLResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortenerServer) GetStats(ctx context.Context, request *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortenerServer) ListShortURLs(ctx context.Context, request *pb.ListShortURLsRequest) (*pb.ListShortURLsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortenerServer) Shutdown() {
	s.grpc.GracefulStop()
}
