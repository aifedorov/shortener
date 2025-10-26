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
// This endpoint is available to all users without authentication.
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

		ip, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			logger.Log.Error("stats: failed to parse trusted subnet", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		logger.Log.Debug("stats: checking IP", zap.String("ip", rIP.String()))
		logger.Log.Debug("stats: masked IP", zap.String("mask", ipnet.Mask.String()))

		if !ip.Equal(rIP.Mask(ipnet.Mask)) {
			logger.Log.Info("stats: request IP is not in trusted subnet", zap.String("ip", rIP.String()))
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
	ip := net.ParseIP(ipStr)
	if ipStr == "" {
		ips := r.Header.Get("X-Forwarded-For")
		if len(ips) == 0 {
			return nil, nil
		}
		ipStrs := strings.Split(ips, `,`)
		ip = net.ParseIP(ipStrs[0])
	}
	if ip == nil {
		return nil, errors.New("failed to resolve IP")
	}
	return ip, nil
}
