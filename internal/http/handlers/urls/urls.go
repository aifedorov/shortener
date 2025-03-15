package urls

import (
	"errors"
	"net/http"
	"time"

	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
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
	UserID string
}

func NewURLsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if userID == "" {
			logger.Log.Info("user is not authorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(userID))
		if err != nil {
			logger.Log.Error("failed to write response", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func getUserID(tokenString string) string {
	if tokenString == "" {
		return ""
	}

	logger.Log.Debug("parsing token")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		logger.Log.Error("error parsing token", zap.Error(err))
		return ""
	}

	logger.Log.Debug("checking token")
	if !token.Valid {
		logger.Log.Error("invalid token")
		return ""
	}

	return claims.UserID
}

func setNewCookies(w http.ResponseWriter) {
	logger.Log.Debug("building JWT token")
	userID := uuid.NewString()
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

func buildJWTString(userID string) (string, error) {
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
