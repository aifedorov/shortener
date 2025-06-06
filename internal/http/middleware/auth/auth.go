package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

const (
	tokenExp  = time.Hour * 3
	tokenName = "JWT"
)

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

			logger.Log.Debug("auth: creating new user_id")
			userID := uuid.NewString()
			setNewCookies(userID, m.secretKey, w)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
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

func setNewCookies(userID, secretKey string, w http.ResponseWriter) {
	logger.Log.Debug("auth: setting new cookies", zap.String("user_id", userID))
	token, err := buildJWTString(userID, secretKey)
	if err != nil {
		logger.Log.Error("auth: failed to build JWT token", zap.String("error", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("auth: setting cookie with JWT token")
	cookie := http.Cookie{
		Name:     tokenName,
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}

func buildJWTString(userID, secretKey string) (string, error) {
	logger.Log.Debug("auth: building JWT token with user_id", zap.String("user_id", userID))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Error("auth: failed to sign JWT token", zap.String("error", err.Error()))
		return "", err
	}
	return tokenString, nil
}
