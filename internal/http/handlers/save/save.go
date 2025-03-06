package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aifedorov/shortener/pkg/logger"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/validate"
)

type Request struct {
	URL string `json:"url"`
}

func (r Request) String() string {
	return fmt.Sprintf("{url: %s}", r.URL)
}

type Response struct {
	ShortURL string `json:"result"`
}

func (r Response) String() string {
	return fmt.Sprintf("{shortURL: %s}", r.ShortURL)
}

func NewSavePlainTextHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")

		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			logger.Log.Error("failed to read request body", zap.Error(readErr))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		url := string(body)
		if err := urlChecker.CheckURL(url); err != nil {
			logger.Log.Error("invalid url", zap.String("url", url))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resURL, saveErr := repo.Store(config.ShortBaseURL, url)
		if errors.Is(saveErr, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusOK)
			_, writeErr := rw.Write([]byte(resURL))
			if writeErr != nil {
				logger.Log.Error("Failed to write response", zap.Error(writeErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		rw.WriteHeader(http.StatusCreated)
		_, writeErr := rw.Write([]byte(resURL))
		if writeErr != nil {
			logger.Log.Error("Failed to write response", zap.Error(writeErr))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func NewSaveJSONHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		logger.Log.Debug("decoding request body")
		var reqBody Request

		if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
			logger.Log.Error("failed to decode request", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger.Log.Debug("request body decoded", zap.String("request", reqBody.String()))

		if err := urlChecker.CheckURL(reqBody.URL); err != nil {
			logger.Log.Error("invalid url parameter in request", zap.String("url", reqBody.URL))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resURL, saveErr := repo.Store(config.ShortBaseURL, reqBody.URL)
		if errors.Is(saveErr, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusOK)
			if err := encodeResponse(rw, resURL); err != nil {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			logger.Log.Debug("sending HTTP 200 response")
			return
		}

		rw.WriteHeader(http.StatusCreated)
		if err := encodeResponse(rw, resURL); err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("sending HTTP 201 response")
	}
}

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
