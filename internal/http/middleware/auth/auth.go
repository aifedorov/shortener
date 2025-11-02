package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
)

// ContextKey represents a type for context keys used in authentication.
type ContextKey string

// UserIDKey is the context key used to store the user ID in request context.
const UserIDKey ContextKey = "user_id"

// GetUserID extracts the user ID from the context.
// Returns an error if user ID is not present in the context or is empty.
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}

var ErrUserIDNotFound = errors.New("user_id not found")

const (
	// tokenExp defines the JWT token expiration time.
	tokenExp = time.Hour * 3
	// cookieExp defines the cookie expiration time (24 hours).
	cookieExp = time.Hour * 24
	// tokenName is the name of the JWT cookie.
	tokenName = "JWT"
)

// Middleware provides JWT-based authentication middleware for HTTP handlers.
type Middleware struct {
	jwtManager *jwt.Manager
}

// NewMiddleware creates a new authentication middleware instance.
// The secretKey is used for JWT token signing and validation.
func NewMiddleware(secretKey string) *Middleware {
	return &Middleware{
		jwtManager: jwt.NewManager(secretKey, tokenExp),
	}
}

// GetJWTManager returns the JWT manager instance for use in other components (e.g., gRPC interceptors).
func (m *Middleware) GetJWTManager() *jwt.Manager {
	return m.jwtManager
}

// JWTAuth provides JWT-based authentication middleware.
// It extracts user ID from JWT cookies and creates new cookies for unauthenticated users.
func (m *Middleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(tokenName)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("auth: cookie not present", zap.String("name", tokenName))

			logger.Log.Debug("auth: creating new user_id")
			userID := uuid.NewString()
			setNewCookies(userID, m.jwtManager, w)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID, err := m.jwtManager.Verify(cookie.Value)
		if err != nil {
			logger.Log.Error("auth: failed to get cookie", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setNewCookies(userID string, jwtManager *jwt.Manager, w http.ResponseWriter) {
	logger.Log.Debug("auth: setting new cookies", zap.String("user_id", userID))
	token, err := jwtManager.Generate(userID)
	if err != nil {
		logger.Log.Error("auth: failed to build JWT token", zap.String("error", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("auth: setting cookie with JWT token")
	cookie := http.Cookie{
		Name:     tokenName,
		Value:    token,
		Expires:  time.Now().Add(cookieExp),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}
