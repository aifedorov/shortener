package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

const (
	tokenExp  = time.Hour * 3
	tokenName = "JWT"
	// TODO:  Use encrypted storage for the key.
	secretKey = "1q2w3e4r"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(tokenName)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("cookie not present", zap.String("name", tokenName))

			logger.Log.Debug("creating new user_id")
			userID := uuid.NewString()
			setNewCookies(userID, w)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID, err := getUserID(cookie.Value)
		if err != nil {
			logger.Log.Error("failed to get cookie", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserID(tokenString string) (string, error) {
	logger.Log.Debug("parsing token")
	if tokenString == "" {
		logger.Log.Error("empty token")
		return "", errors.New("token is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		logger.Log.Error("error parsing token", zap.Error(err))
		return "", errors.New("invalid token")
	}

	logger.Log.Debug("checking token")
	if !token.Valid {
		logger.Log.Error("invalid token")
		return "", errors.New("invalid token")
	}
	return claims.UserID, nil
}

func setNewCookies(userID string, w http.ResponseWriter) {
	logger.Log.Debug("setting new cookies", zap.String("user_id", userID))
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
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}

func buildJWTString(userID string) (string, error) {
	logger.Log.Debug("building JWT token with user_id", zap.String("user_id", userID))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Error("failed to sign JWT token", zap.String("error", err.Error()))
		return "", err
	}
	return tokenString, nil
}
