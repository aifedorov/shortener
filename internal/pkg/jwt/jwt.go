// Package jwt provides JWT token generation and validation functionality.
//
// This package implements JSON Web Token (JWT) authentication using the HS256 signing method.
// It is used by both HTTP and gRPC authentication middleware to manage user sessions.
//
// Example usage:
//
//	jwtService := jwt.NewService("your-secret-key", 24*time.Hour)
//	token, err := jwtService.Generate("user123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	userID, err := jwtService.ParseWithUserID(token)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
	// ErrEmptyToken is returned when an empty token string is provided for parsing.
	ErrEmptyToken = errors.New("token is empty")

	// ErrInvalidToken is returned when token parsing or validation fails.
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidSigningMethod is returned when the token uses an unexpected signing method.
	ErrInvalidSigningMethod = errors.New("unexpected signing method")
)

//go:generate mockgen -destination=../../mocks/jwt_mock.go -package=mocks github.com/aifedorov/shortener/internal/pkg/jwt JWT

// JWT defines the interface for JWT token operations.
// It provides methods to generate new tokens and parse existing ones to extract user IDs.
type JWT interface {
	// Generate creates a new JWT token for the given user ID.
	// The token is signed using HS256 and includes an expiration time.
	//
	// Parameters:
	//   - userID: The unique identifier for the user
	//
	// Returns:
	//   - string: The signed JWT token
	//   - error: An error if token generation fails
	Generate(userID string) (string, error)

	// ParseWithUserID validates a JWT token and extracts the user ID from it.
	// It verifies the token signature, expiration, and signing method.
	//
	// Parameters:
	//   - tokenString: The JWT token to parse
	//
	// Returns:
	//   - string: The user ID extracted from the token
	//   - error: ErrEmptyToken if token is empty, ErrInvalidToken if validation fails
	ParseWithUserID(tokenString string) (string, error)
}

// service is the concrete implementation of the JWT interface.
type service struct {
	secretKey string        // The secret key used for signing tokens
	tokenExp  time.Duration // The token expiration duration
}

// Claims represents the JWT claims structure.
// It extends the standard RegisteredClaims with a custom UserID field.
type Claims struct {
	jwt.RegisteredClaims
	UserID string // The user identifier stored in the token
}

// NewService creates a new JWT service instance.
//
// Parameters:
//   - secretKey: The secret key used for signing tokens (must be kept secure)
//   - tokenExp: The duration after which tokens will expire
//
// Returns:
//   - JWT: A new JWT service instance
func NewService(secretKey string, tokenExp time.Duration) JWT {
	return &service{
		secretKey: secretKey,
		tokenExp:  tokenExp,
	}
}

// Generate creates a new JWT token for the specified user ID.
// The token is signed using HS256 HMAC algorithm with the configured secret key.
//
// The generated token includes:
//   - UserID: Custom claim containing the user identifier
//   - ExpiresAt: Token expiration timestamp
//
// Returns an error if token signing fails.
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

// ParseWithUserID parses and validates a JWT token, returning the user ID.
//
// This method performs the following validations:
//   - Checks if the token string is not empty
//   - Verifies the token signature using the configured secret key
//   - Validates the signing method is HS256
//   - Checks token expiration
//   - Extracts and returns the user ID from claims
//
// Returns:
//   - string: The user ID extracted from the token
//   - error: ErrEmptyToken if token is empty, ErrInvalidToken for validation failures,
//     ErrInvalidSigningMethod if signing method is unexpected
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
