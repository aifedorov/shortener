package url

import "errors"

var (
	// ErrInvalidUserID occurs when UserID is empty or invalid.
	ErrInvalidUserID = errors.New("user ID cannot be empty")

	// ErrInvalidAlias occurs when alias is empty or invalid.
	ErrInvalidAlias = errors.New("alias cannot be empty")

	// ErrInvalidOriginalURL occurs when original URL is empty or invalid.
	ErrInvalidOriginalURL = errors.New("original URL cannot be empty")

	// ErrAlreadyDeleted occurs when attempting to delete an already deleted URL.
	ErrAlreadyDeleted = errors.New("URL is already deleted")

	// ErrShortURLNotFound occurs when short URL is not found in storage.
	ErrShortURLNotFound = errors.New("short URL not found")

	// ErrInvalidURL occurs when URL does not meet validation requirements.
	ErrInvalidURL = errors.New("invalid URL format")

	// ErrEmptyBatch occurs when attempting to save an empty batch.
	ErrEmptyBatch = errors.New("batch cannot be empty")
)
