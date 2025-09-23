package repository

// BatchURLInput represents a single URL input for batch operations.
// Used internally by the repository layer for batch URL storage.
type BatchURLInput struct {
	// CID is the correlation ID to match input and output items.
	CID string
	// OriginalURL is the original URL to be shortened.
	OriginalURL string
}

// BatchURLOutput represents a single URL output from batch operations.
// Returned by the repository layer after successful batch URL storage.
type BatchURLOutput struct {
	// CID is the correlation ID matching the original input.
	CID string
	// ShortURL is the generated short URL.
	ShortURL string
}

// URLOutput represents a URL entry in the user's URL list.
// Used for returning user's URLs in API responses.
type URLOutput struct {
	// ShortURL is the generated short URL.
	ShortURL string `json:"short_url"`
	// OriginalURL is the original URL that was shortened.
	OriginalURL string `json:"original_url"`
}
