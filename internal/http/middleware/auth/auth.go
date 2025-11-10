package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/aifedorov/shortener/internal/domain/user"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/jwt"
)

const (
	// cookieExp defines the cookie expiration time (24 hours).
	cookieExp = time.Hour * 24
	// tokenName is the name of the JWT cookie.
	tokenName = "JWT"
)

// Middleware provides JWT-based authentication middleware for HTTP handlers.
type Middleware struct {
	jwtService  jwt.JWT
	userService user.Service
}

// NewMiddleware creates a new authentication middleware instance.
// The secretKey is used for JWT token signing and validation.
func NewMiddleware(authService jwt.JWT, userService user.Service) *Middleware {
	return &Middleware{
		jwtService:  authService,
		userService: userService,
	}
}

// GetJWTService returns the JWT manager instance for use in other components (e.g., gRPC interceptors).
func (m *Middleware) GetJWTService() jwt.JWT {
	return m.jwtService
}

// JWTAuth provides JWT-based authentication middleware.
// It extracts user ID from JWT cookies and creates new cookies for unauthenticated users.
func (m *Middleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(tokenName)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("auth: cookie not present", zap.String("name", tokenName))

			logger.Log.Debug("auth: creating new user_id")
			userID := m.userService.GenerateUserID()
			setNewCookies(userID, m.jwtService, w)

			ctx := m.userService.SetUserIDToContext(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID, err := m.jwtService.ParseWithUserID(cookie.Value)
		if err != nil {
			logger.Log.Error("auth: failed to get cookie", zap.String("name", tokenName), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx := m.userService.SetUserIDToContext(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setNewCookies(userID string, jwtService jwt.JWT, w http.ResponseWriter) {
	logger.Log.Debug("auth: setting new cookies", zap.String("user_id", userID))
	token, err := jwtService.Generate(userID)
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
