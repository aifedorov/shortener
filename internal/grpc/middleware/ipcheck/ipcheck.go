package ipcheck

import (
	"context"
	"net"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	errMsgAccessDenied = "access denied"
)

type Interceptor struct {
	trustedIPNet *net.IPNet
}

func NewInterceptor(trustedIPNet *net.IPNet) *Interceptor {
	return &Interceptor{
		trustedIPNet: trustedIPNet,
	}
}

var protectedMethods = map[string]bool{
	"/shortener.v1.ShortenerService/GetStats": true,
}

func (i *Interceptor) UnaryIPCheckInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !protectedMethods[info.FullMethod] {
		logger.Log.Debug("ipcheck: IP check not required",
			zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}

	if i.trustedIPNet == nil {
		logger.Log.Warn("ipcheck: getStats called but trusted subnet is not configured")
		return nil, status.Error(codes.PermissionDenied, errMsgAccessDenied)
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Log.Error("ipcheck: failed to get peer from context")
		return nil, status.Error(codes.PermissionDenied, errMsgAccessDenied)
	}

	tcpAddr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		logger.Log.Error("ipcheck: peer address is not TCP")
		return nil, status.Error(codes.PermissionDenied, errMsgAccessDenied)
	}

	clientIP := tcpAddr.IP

	if !i.trustedIPNet.Contains(clientIP) {
		logger.Log.Info("ipcheck: request IP is not in trusted subnet",
			zap.String("ip", clientIP.String()),
			zap.String("subnet", i.trustedIPNet.String()),
		)
		return nil, status.Error(codes.PermissionDenied, errMsgAccessDenied)
	}

	return handler(ctx, req)
}
