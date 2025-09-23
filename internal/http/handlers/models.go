package handlers

import (
	"fmt"
)

// RequestBody represents the request body for URL shortening operations.
type RequestBody struct {
	URL string `json:"url"`
}

func (r RequestBody) String() string {
	return fmt.Sprintf("{url: %s}", r.URL)
}

// Response represents the response body for URL shortening operations.
type Response struct {
	ShortURL string `json:"result"`
}

func (r Response) String() string {
	return fmt.Sprintf("{shortURL: %s}", r.ShortURL)
}

// BatchRequest represents a single URL in a batch shortening request.
type BatchRequest struct {
	CID         string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

func (r BatchRequest) String() string {
	return fmt.Sprintf("{correlation_id: %s, original_url: %s}", r.CID, r.OriginalURL)
}

// BatchResponse represents a single URL in a batch shortening response.
type BatchResponse struct {
	CID      string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

func (r BatchResponse) String() string {
	return fmt.Sprintf("{correlation_id: %s, short_url: %s}", r.CID, r.ShortURL)
}
