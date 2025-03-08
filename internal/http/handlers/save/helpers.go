package save

import (
	"encoding/json"
	"errors"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

func encodeResponse(rw http.ResponseWriter, resURL string) error {
	encoder := json.NewEncoder(rw)
	resp := Response{
		ShortURL: resURL,
	}

	logger.Log.Debug("encoding response", zap.Any("response", resp))
	if err := encoder.Encode(resp); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		return errors.New("failed to encode response")
	}
	return nil
}

func encodeBatchResponse(rw http.ResponseWriter, urls []repository.URLOutput) error {
	encoder := json.NewEncoder(rw)
	resp := make([]BatchResponse, len(urls))
	for url := range urls {
		r := BatchResponse{
			CID:      urls[url].CID,
			ShortURL: urls[url].ShortURL,
		}
		resp[url] = r
	}

	logger.Log.Debug("encoding response", zap.Any("response", resp))
	if err := encoder.Encode(resp); err != nil {
		logger.Log.Error("failed to encode batch response", zap.Error(err))
		return errors.New("failed to encode batch response")
	}
	return nil
}
