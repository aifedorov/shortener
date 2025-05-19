package handlers

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/validate"
)

func decodeRequest(r *http.Request) (RequestBody, error) {
	logger.Log.Debug("decoding request body")
	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if errors.Is(err, io.EOF) {
		return RequestBody{}, errors.New("request body is empty")
	}
	if err != nil {
		logger.Log.Error("failed to decode request", zap.Error(err))
		return RequestBody{}, errors.New("failed to decode request body")
	}
	return body, nil
}

func decodeBatchRequest(r *http.Request) ([]BatchRequest, error) {
	logger.Log.Debug("decoding request body")
	var urls []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		logger.Log.Error("failed to decode request", zap.Error(err))
		return nil, errors.New("failed to decode request body")
	}
	return urls, nil
}

func decodeAliasesRequest(r *http.Request) ([]string, error) {
	logger.Log.Debug("decoding request body")
	var aliases []string
	if err := json.NewDecoder(r.Body).Decode(&aliases); err != nil {
		logger.Log.Error("failed to decode request", zap.Error(err))
		return nil, errors.New("failed to decode request body")
	}
	return aliases, nil
}

func encodeResponse(rw http.ResponseWriter, resURL string) error {
	logger.Log.Debug("encoding response")
	encoder := json.NewEncoder(rw)
	resp := Response{
		ShortURL: resURL,
	}

	if err := encoder.Encode(resp); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		return errors.New("failed to encode response")
	}
	return nil
}

func encodeBatchResponse(rw http.ResponseWriter, urls []repository.BatchURLOutput) error {
	logger.Log.Debug("encoding response")
	encoder := json.NewEncoder(rw)
	resp := make([]BatchResponse, len(urls))
	for url := range urls {
		r := BatchResponse{
			CID:      urls[url].CID,
			ShortURL: urls[url].ShortURL,
		}
		resp[url] = r
	}

	if err := encoder.Encode(resp); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		return errors.New("failed to encode response")
	}
	return nil
}

func encodeURLsResponse(rw http.ResponseWriter, urls []repository.URLOutput) error {
	logger.Log.Debug("encoding response")
	encoder := json.NewEncoder(rw)
	resp := make([]repository.URLOutput, len(urls))
	for i, url := range urls {
		r := repository.URLOutput{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
		resp[i] = r
	}

	if err := encoder.Encode(resp); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		return errors.New("failed to encode response")
	}
	return nil
}

func validateURLs(reqURLs []BatchRequest, urlChecker validate.URLChecker) ([]repository.BatchURLInput, error) {
	logger.Log.Debug("validating url")
	var urls = make([]repository.BatchURLInput, len(reqURLs))
	for i, reqBodyURL := range reqURLs {
		if err := urlChecker.CheckURL(reqBodyURL.OriginalURL); err != nil {
			logger.Log.Error("invalid url", zap.String("url", reqBodyURL.OriginalURL), zap.Error(err))
			return nil, errors.New("invalid url")
		}
		urls[i] = repository.BatchURLInput{
			CID:         reqBodyURL.CID,
			OriginalURL: reqBodyURL.OriginalURL,
		}
	}
	return urls, nil
}

func getUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		logger.Log.Error("user_id not found")
		return "", errors.New("user_id not found")
	}
	return userID, nil
}
