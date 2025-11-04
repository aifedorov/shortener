package ipcheck

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"
)

type Middleware struct {
	trustedIPNet *net.IPNet
}

func NewMiddleware(trustedIPNet *net.IPNet) *Middleware {
	return &Middleware{
		trustedIPNet: trustedIPNet,
	}
}

func (m *Middleware) IPCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rIP, err := resolveIP(r)
		if err != nil {
			logger.Log.Error("stats: failed to resolve IP", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if m.trustedIPNet == nil || !m.trustedIPNet.Contains(rIP) {
			logger.Log.Info("stats: request IP is not in trusted subnet",
				zap.String("ip", rIP.String()),
				zap.String("subnet", m.trustedIPNet.String()))
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func resolveIP(r *http.Request) (net.IP, error) {
	ipStr := r.Header.Get("X-Real-IP")
	if ipStr != "" {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, errors.New("failed to parse X-Real-IP")
		}
		return ip, nil
	}

	ips := r.Header.Get("X-Forwarded-For")
	if ips != "" {
		ipStrs := strings.Split(ips, ",")
		if len(ipStrs) == 0 {
			return nil, errors.New("empty X-Forwarded-For header")
		}

		ip := net.ParseIP(strings.TrimSpace(ipStrs[0]))
		if ip == nil {
			return nil, errors.New("failed to parse X-Forwarded-For")
		}
	}

	ipstrs := strings.Split(r.RemoteAddr, ":")
	if len(ipstrs) == 0 {
		return nil, errors.New("failed to parse IP address from RemoteAddr")
	}

	ip := net.ParseIP(ipstrs[0])
	if ip == nil {
		return nil, errors.New("failed to parse IP address from RemoteAddr")
	}

	return ip, nil
}
