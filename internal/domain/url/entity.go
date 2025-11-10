package url

import (
	"time"
)

// URL represents a domain entity - shortened URL.
type URL struct {
	id          string
	userID      string
	alias       string
	originalURL string
	createdAt   time.Time
	isDeleted   bool
}

// NewURL creates a new URL entity with validation.
func NewURL(userID, alias, originalURL string) (*URL, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if alias == "" {
		return nil, ErrInvalidAlias
	}
	if originalURL == "" {
		return nil, ErrInvalidOriginalURL
	}

	return &URL{
		userID:      userID,
		alias:       alias,
		originalURL: originalURL,
		createdAt:   time.Now(),
		isDeleted:   false,
	}, nil
}

// ID returns the URL identifier.
func (u *URL) ID() string {
	return u.id
}

// UserID returns the owner's identifier.
func (u *URL) UserID() string {
	return u.userID
}

// Alias returns the short URL alias.
func (u *URL) Alias() string {
	return u.alias
}

// OriginalURL returns the original URL.
func (u *URL) OriginalURL() string {
	return u.originalURL
}

// CreatedAt returns the creation time.
func (u *URL) CreatedAt() time.Time {
	return u.createdAt
}

// IsDeleted checks if the URL has been deleted.
func (u *URL) IsDeleted() bool {
	return u.isDeleted
}

// MarkAsDeleted marks the URL as deleted (soft delete).
func (u *URL) MarkAsDeleted() error {
	if u.isDeleted {
		return ErrAlreadyDeleted
	}
	u.isDeleted = true
	return nil
}

// IsOwnedBy checks if the URL belongs to the specified user.
func (u *URL) IsOwnedBy(userID string) bool {
	return u.userID == userID
}

// CanBeAccessed checks if the URL is available for use.
func (u *URL) CanBeAccessed() bool {
	return !u.isDeleted
}

// SetID sets the ID.
func (u *URL) SetID(id string) {
	u.id = id
}
