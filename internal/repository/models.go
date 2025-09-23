package repository

// BatchURLInput represents a single URL input for batch operations.
type BatchURLInput struct {
	CID         string
	OriginalURL string
}

// BatchURLOutput represents a single URL output from batch operations.
type BatchURLOutput struct {
	CID      string
	ShortURL string
}

// URLOutput represents a URL entry in the user's URL list.
type URLOutput struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
