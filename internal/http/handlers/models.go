package handlers

import (
	"fmt"
)

// RequestBody represents the request body for URL shortening operations.
// Used in JSON API endpoints for single URL shortening.
type RequestBody struct {
	// URL is the original URL to be shortened.
	URL string `json:"url"`
}

// String returns a string representation of the RequestBody.
func (r RequestBody) String() string {
	return fmt.Sprintf("{url: %s}", r.URL)
}

// Response represents the response body for URL shortening operations.
// Returned by JSON API endpoints after successful URL shortening.
type Response struct {
	// ShortURL is the generated short URL.
	ShortURL string `json:"result"`
}

// String returns a string representation of the Response.
func (r Response) String() string {
	return fmt.Sprintf("{shortURL: %s}", r.ShortURL)
}

// BatchRequest represents a single URL in a batch shortening request.
// Used in batch API endpoints for processing multiple URLs at once.
type BatchRequest struct {
	// CID is the correlation ID to match request and response items.
	CID string `json:"correlation_id"`
	// OriginalURL is the original URL to be shortened.
	OriginalURL string `json:"original_url"`
}

// String returns a string representation of the BatchRequest.
func (r BatchRequest) String() string {
	return fmt.Sprintf("{correlation_id: %s, original_url: %s}", r.CID, r.OriginalURL)
}

// BatchResponse represents a single URL in a batch shortening response.
// Returned by batch API endpoints after successful URL shortening.
type BatchResponse struct {
	// CID is the correlation ID matching the original request.
	CID string `json:"correlation_id"`
	// ShortURL is the generated short URL.
	ShortURL string `json:"short_url"`
}

// String returns a string representation of the BatchResponse.
func (r BatchResponse) String() string {
	return fmt.Sprintf("{correlation_id: %s, short_url: %s}", r.CID, r.ShortURL)
}
