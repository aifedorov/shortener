package auth

import (
	"context"

	userDomain "github.com/aifedorov/shortener/internal/domain/user"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// jwyKey is the context key for storing JWT token
	jwyKey ContextKey = "jwt"
)

var publicMethods = map[string]bool{
	"/shortener.v1.ShortenerService/Ping":        true,
	"/shortener.v1.ShortenerService/GetShortURL": true,
	"/shortener.v1.ShortenerService/GetStats":    true,
}

type Interceptor struct {
	authService jwt.JWT
	userService userDomain.Service
}

func NewInterceptor(authService jwt.JWT, userService userDomain.Service) Interceptor {
	return Interceptor{
		authService: authService,
		userService: userService,
	}
}

func (i *Interceptor) UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if publicMethods[info.FullMethod] {
		logger.Log.Debug("auth: public method, skipping authentication",
			zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "no metadata")
	}

	token := md.Get(string(jwyKey))
	if len(token) == 0 {
		logger.Log.Debug("auth: generating new token")
		userID := uuid.NewString()
		jwtToken, err := i.authService.Generate(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "server internal error")
		}

		ctx = i.userService.SetUserIDToContext(ctx, userID)
		ctx = metadata.AppendToOutgoingContext(ctx, string(jwyKey), jwtToken)
		return handler(ctx, req)
	}

	userID, err := i.authService.ParseWithUserID(token[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify token: %v", err)
	}
	ctx = i.userService.SetUserIDToContext(ctx, userID)

	return handler(ctx, req)
}
