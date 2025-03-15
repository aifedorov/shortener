package auth

import (
	"errors"
	"fmt"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TODO: Use a secret store for key.
const (
	TokenExp  = time.Hour * 3
	TokenName = "token"
	SecretKey = "1q2w3e4r"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("cookie not present", zap.String("name", TokenName))
			setNewCookies(w)
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			logger.Log.Error("failed to get cookie", zap.String("name", TokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserID(tokenString string) (string, error) {
	if tokenString == "" {
		return "", errors.New("token is empty")
	}

	logger.Log.Debug("parsing token")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
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

func setNewCookies(w http.ResponseWriter) {
	logger.Log.Debug("building JWT token")
	token, err := buildJWTString(uuid.NewString())
	if err != nil {
		logger.Log.Error("failed to build JWT token", zap.String("error", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("setting cookie with JWT token")
	cookie := http.Cookie{
		Name:     TokenName,
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/api/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

func buildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
