package handlers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
)

// NewStatsHandler creates a new HTTP handler for retrieving service statistics.
// This endpoint requires the client IP to be within the trusted subnet.
// It returns the total number of shortened URLs and users in the service.
// Only the GET method is allowed.
func NewStatsHandler(cfg *config.Config, repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rIP, err := resolveIP(r)
		if err != nil {
			logger.Log.Error("stats: failed to resolve IP", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			logger.Log.Error("stats: failed to parse trusted subnet", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if !ipnet.Contains(rIP) {
			logger.Log.Info("stats: request IP is not in trusted subnet",
				zap.String("ip", rIP.String()),
				zap.String("subnet", cfg.TrustedSubnet))
			http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		rw.Header().Set("Content-Type", "application/json")

		stats, err := repo.GetStats()
		if err != nil {
			logger.Log.Error("stats: failed to get stats", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response := StatsResponse{
			URLs:  stats.TotalURLs,
			Users: stats.TotalUsers,
		}

		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(response); err != nil {
			logger.Log.Error("stats: failed to encode response", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
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
	if ips == "" {
		return nil, errors.New("no IP headers found")
	}

	ipStrs := strings.Split(ips, ",")
	if len(ipStrs) == 0 {
		return nil, errors.New("empty X-Forwarded-For header")
	}

	ip := net.ParseIP(strings.TrimSpace(ipStrs[0]))
	if ip == nil {
		return nil, errors.New("failed to parse X-Forwarded-For")
	}
	return ip, nil
}
