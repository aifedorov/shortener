package handlers

import (
	"fmt"
	"net/http/httptest"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
)

// ExampleNewSavePlainTextHandler demonstrates how to create a plain text URL shortening handler.
func ExampleNewSavePlainTextHandler() {
	// Create a test configuration
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// Create a mock repository
	repo := &mockRepository{}

	// Create URL checker
	urlChecker := validate.NewService()

	// Create the handler
	_ = NewSavePlainTextHandler(cfg, repo, urlChecker)

	// The handler is now ready to process plain text URL shortening requests
	// Note: In a real application, this handler requires user authentication
	fmt.Printf("Handler created for base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Handler accepts: POST / with text/plain content\n")

	// Output:
	// Handler created for base URL: http://localhost:8080
	// Handler accepts: POST / with text/plain content
}

// ExampleNewSaveJSONHandler demonstrates how to create a JSON URL shortening handler.
func ExampleNewSaveJSONHandler() {
	// Create a test configuration
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// Create a mock repository
	repo := &mockRepository{}

	// Create URL checker
	urlChecker := validate.NewService()

	// Create the handler
	_ = NewSaveJSONHandler(cfg, repo, urlChecker)

	// The handler is now ready to process JSON URL shortening requests
	// Note: In a real application, this handler requires user authentication
	fmt.Printf("Handler created for base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Handler accepts: POST /api/shorten with application/json content\n")

	// Output:
	// Handler created for base URL: http://localhost:8080
	// Handler accepts: POST /api/shorten with application/json content
}

// ExampleNewSaveJSONBatchHandler demonstrates how to create a batch URL shortening handler.
func ExampleNewSaveJSONBatchHandler() {
	// Create a test configuration
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// Create a mock repository
	repo := &mockRepository{}

	// Create URL checker
	urlChecker := validate.NewService()

	// Create the handler
	_ = NewSaveJSONBatchHandler(cfg, repo, urlChecker)

	// The handler is now ready to process batch URL shortening requests
	// Note: In a real application, this handler requires user authentication
	fmt.Printf("Handler created for base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Handler accepts: POST /api/shorten/batch with application/json content\n")

	// Output:
	// Handler created for base URL: http://localhost:8080
	// Handler accepts: POST /api/shorten/batch with application/json content
}

// ExampleNewRedirectHandler demonstrates how to create a redirect handler.
func ExampleNewRedirectHandler() {
	// Create a mock repository
	repo := &mockRepository{}

	// Create the handler
	_ = NewRedirectHandler(repo)

	// The handler is now ready to redirect short URLs to their original URLs
	// Note: This handler is available to all users (no authentication required)
	fmt.Printf("Handler created for URL redirection\n")
	fmt.Printf("Handler accepts: GET /{shortURL}\n")

	// Output:
	// Handler created for URL redirection
	// Handler accepts: GET /{shortURL}
}

// ExampleNewURLsHandler demonstrates how to create a user URLs handler.
func ExampleNewURLsHandler() {
	// Create a test configuration
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// Create a mock repository
	repo := &mockRepository{}

	// Create the handler
	_ = NewURLsHandler(cfg, repo)

	// The handler is now ready to retrieve user URLs
	// Note: In a real application, this handler requires user authentication
	fmt.Printf("Handler created for base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Handler accepts: GET /api/user/urls\n")

	// Output:
	// Handler created for base URL: http://localhost:8080
	// Handler accepts: GET /api/user/urls
}

// ExampleNewDeleteHandler demonstrates how to create a delete handler.
func ExampleNewDeleteHandler() {
	// Create a mock repository
	repo := &mockRepository{}

	// Create the handler
	_ = NewDeleteHandler(repo)

	// The handler is now ready to delete user URLs
	// Note: In a real application, this handler requires user authentication
	fmt.Printf("Handler created for URL deletion\n")
	fmt.Printf("Handler accepts: DELETE /api/user/urls with application/json content\n")

	// Output:
	// Handler created for URL deletion
	// Handler accepts: DELETE /api/user/urls with application/json content
}

// ExampleNewPingHandler demonstrates how to use the ping handler.
func ExampleNewPingHandler() {
	// Create a mock repository
	repo := &mockRepository{}

	// Create the handler
	handler := NewPingHandler(repo)

	// Create a test request
	req := httptest.NewRequest("GET", "/ping", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler(rr, req)

	// Check the response
	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Response: %s\n", rr.Body.String())
	// Output:
	// Status: 200
	// Response:
}

// mockRepository is a mock implementation of the repository.Repository interface for testing.
type mockRepository struct {
	urls     map[string]string
	userURLs map[string][]repository.URLOutput
}

func (m *mockRepository) Run() error {
	if m.urls == nil {
		m.urls = make(map[string]string)
	}
	if m.userURLs == nil {
		m.userURLs = make(map[string][]repository.URLOutput)
	}
	return nil
}

func (m *mockRepository) Ping() error {
	return nil
}

func (m *mockRepository) Close() error {
	return nil
}

func (m *mockRepository) Get(shortURL string) (string, error) {
	if url, exists := m.urls[shortURL]; exists {
		return url, nil
	}
	return "", repository.ErrShortURLNotFound
}

func (m *mockRepository) GetAll(userID, baseURL string) ([]repository.URLOutput, error) {
	if urls, exists := m.userURLs[userID]; exists {
		return urls, nil
	}
	return nil, repository.ErrUserHasNoData
}

func (m *mockRepository) Store(userID, baseURL, targetURL string) (string, error) {
	shortURL := "abc123"
	m.urls[shortURL] = targetURL
	return baseURL + "/" + shortURL, nil
}

func (m *mockRepository) StoreBatch(userID, baseURL string, urls []repository.BatchURLInput) ([]repository.BatchURLOutput, error) {
	var results []repository.BatchURLOutput
	for i, url := range urls {
		shortURL := fmt.Sprintf("abc%d", i+1)
		m.urls[shortURL] = url.OriginalURL
		results = append(results, repository.BatchURLOutput{
			CID:      url.CID,
			ShortURL: baseURL + "/" + shortURL,
		})
	}
	return results, nil
}

func (m *mockRepository) DeleteBatch(userID string, aliases []string) error {
	// Mock implementation - just return success
	return nil
}
