package urls

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// TODO: Use a secret store for key.
const (
	tokenExp  = time.Hour * 3
	tokenName = "token"
	secretKey = "1q2w3e4r"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURLsHandler(cfg *config.Config, repo repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cookie, err := r.Cookie(tokenName)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("cookie not present", zap.String("name", tokenName))
			setNewCookies(w)
			return
		}
		if err != nil {
			logger.Log.Error("failed to get cookie", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userID := getUserID(cookie.Value)
		if userID == -1 {
			logger.Log.Info("user is not authorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		logger.Log.Debug("sending HTTP 200 response")
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "user_id: %d", userID)
		if err != nil {
			logger.Log.Error("failed to write response", zap.String("name", tokenName), zap.Error(err))
			return
		}
	}
}

func getUserID(tokenString string) int {
	if tokenString == "" {
		return -1
	}

	logger.Log.Debug("parsing token")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		logger.Log.Error("error parsing token", zap.Error(err))
		return -1
	}

	logger.Log.Debug("checking token")
	if !token.Valid {
		logger.Log.Error("invalid token")
		return -1
	}
	return claims.UserID
}

func setNewCookies(w http.ResponseWriter) {
	logger.Log.Debug("building JWT token")
	// TODO: Store userID in DB.
	userID := rand.Intn(1000)
	token, err := buildJWTString(userID)
	if err != nil {
		logger.Log.Error("failed to build JWT token", zap.String("error", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("setting cookie with JWT token")
	cookie := http.Cookie{
		Name:     tokenName,
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/api/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusNoContent)
}

func buildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func encodeResponse(rw http.ResponseWriter, urls []repository.URLOutput) error {
	rw.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(rw)
	resp := make([]URLResponse, len(urls))
	for i, url := range urls {
		r := URLResponse{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
		resp[i] = r
	}

	if err := encoder.Encode(resp); err != nil {
		return err
	}
	return nil
}
