package auth

import (
	"context"
	"errors"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	jwyKey    = "jwt"
	userIDKey = "userID"
)

// Public methods that don't require authentication
var publicMethods = map[string]bool{
	"/shortener.v1.ShortenerService/Ping":        true,
	"/shortener.v1.ShortenerService/GetShortURL": true,
	"/shortener.v1.ShortenerService/GetStats":    true,
}

var ErrUserIDNotFound = errors.New("user_id not found")

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}

type Interceptor struct {
	authService jwt.JWT
}

func NewInterceptor(authService jwt.JWT) Interceptor {
	return Interceptor{
		authService: authService,
	}
}

func (i *Interceptor) UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Check if the method is public
	if publicMethods[info.FullMethod] {
		logger.Log.Debug("auth: public method, skipping authentication",
			zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "no metadata")
	}

	token := md.Get(jwyKey)
	if len(token) == 0 {
		logger.Log.Debug("auth: generating new token")
		userID := uuid.NewString()
		jwtToken, err := i.authService.Generate(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "server internal error")
		}

		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = metadata.AppendToOutgoingContext(ctx, jwyKey, jwtToken)
		return handler(ctx, req)
	}

	userID, err := i.authService.ParseWithUserID(token[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify token: %v", err)
	}
	ctx = context.WithValue(ctx, userIDKey, userID)

	return handler(ctx, req)
}
