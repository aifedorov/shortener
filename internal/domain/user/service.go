package user

import (
	"context"

	"github.com/google/uuid"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// idKey is the context key used to store the user ID in request context.
const idKey = ContextKey("user_id")

// Service defines the interface for user ID operations.
type Service interface {
	GetUserIDFromContext(ctx context.Context) (string, error)
	SetUserIDToContext(ctx context.Context, userID string) context.Context
	GenerateUserID() string
}

// Service contains business logic for working with user ID.
// In the current implementation, users are identified through JWT tokens.
type service struct{}

// NewService creates a new domain service for working with user IDs.
func NewService() Service {
	return &service{}
}

// GenerateUserID generates a new unique user ID.
func (s *service) GenerateUserID() string {
	return uuid.New().String()
}

// GetUserIDFromContext returns the user ID from the request context.
func (s *service) GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(idKey).(string)
	if !ok || userID == "" {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}

// SetUserIDToContext sets the user ID in the request context.
func (s *service) SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, idKey, userID)
}
