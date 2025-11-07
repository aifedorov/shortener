package user

import "github.com/google/uuid"

// Service contains business logic for working with user ID.
// In the current implementation, users are identified through JWT tokens.
type Service struct{}

// NewService creates a new domain service for working with user IDs.
func NewService() *Service {
	return &Service{}
}

// GenerateUserID generates a new unique user ID.
func (s *Service) GenerateUserID() string {
	return uuid.New().String()
}
