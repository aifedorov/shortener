package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

const tokenName = "JWT"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type Middleware struct {
	secretKey string
}

func NewMiddleware(secretKey string) *Middleware {
	return &Middleware{
		secretKey: secretKey,
	}
}

func (m *Middleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(tokenName)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("auth: cookie not present", zap.String("name", tokenName))
			next.ServeHTTP(w, r)
			return
		}

		userID, err := parseUserID(cookie.Value, m.secretKey)
		if err != nil {
			logger.Log.Error("auth: failed to get cookie", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseUserID(tokenString, secretKey string) (string, error) {
	logger.Log.Debug("auth: parsing token")
	if tokenString == "" {
		logger.Log.Error("auth: empty token")
		return "", errors.New("auth: token is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		logger.Log.Error("auth: error parsing token", zap.Error(err))
		return "", errors.New("auth: invalid token")
	}

	logger.Log.Debug("auth: checking token")
	if !token.Valid {
		logger.Log.Error("auth: invalid token")
		return "", errors.New("auth: invalid token")
	}
	return claims.UserID, nil
}
