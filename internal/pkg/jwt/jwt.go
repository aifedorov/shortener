package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

var (
	ErrEmptyToken           = errors.New("token is empty")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("unexpected signing method")
)

//go:generate mockgen -destination=../../mocks/jwt_mock.go -package=mocks github.com/aifedorov/shortener/internal/pkg/jwt JWT

type JWT interface {
	Generate(userID string) (string, error)
	ParseWithUserID(tokenString string) (string, error)
}

type service struct {
	secretKey string
	tokenExp  time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func NewService(secretKey string, tokenExp time.Duration) JWT {
	return &service{
		secretKey: secretKey,
		tokenExp:  tokenExp,
	}
}

func (m *service) Generate(userID string) (string, error) {
	logger.Log.Debug("jwt: generating token", zap.String("user_id", userID))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		logger.Log.Error("jwt: failed to sign token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

func (m *service) ParseWithUserID(tokenString string) (string, error) {
	logger.Log.Debug("jwt: verifying token")

	if tokenString == "" {
		logger.Log.Error("jwt: empty token")
		return "", ErrEmptyToken
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, t.Header["alg"])
			}
			return []byte(m.secretKey), nil
		})

	if err != nil {
		logger.Log.Error("jwt: error parsing token", zap.Error(err))
		return "", ErrInvalidToken
	}

	if !token.Valid {
		logger.Log.Error("jwt: invalid token")
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}
