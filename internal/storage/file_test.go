package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_addNewURLMapping(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	storage := NewFileStorage(tempFile)

	shortURL := "qwerty123"
	originalURL := "https://google.com"

	err := storage.addNewURLMapping(shortURL, originalURL)
	require.NoError(t, err)

	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	var record URLMapping
	err = json.Unmarshal(data[:len(data)-1], &record)
	require.NoError(t, err)

	assert.Equal(t, shortURL, record.ShortURL)
	assert.Equal(t, originalURL, record.OriginalURL)
	assert.NotEmpty(t, record.ID)
}

func TestFileStorage_SaveURL(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	storage := NewFileStorage(tempFile)

	baseURL := "http://localhost:8080"
	targetURL := "https://example.com"

	shortURL, err := storage.SaveURL(baseURL, targetURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortURL)

	assert.Contains(t, shortURL, baseURL)

	_, err = os.Stat(tempFile)
	assert.NoError(t, err)
}

func TestFileStorage_GetURL(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	storage := NewFileStorage(tempFile)

	baseURL := "http://localhost:8080"
	targetURL := "https://example.com"

	shortURL, err := storage.SaveURL(baseURL, targetURL)
	require.NoError(t, err)

	shortID := shortURL[len(baseURL)+1:]

	originalURL, err := storage.GetURL(shortID)
	require.NoError(t, err)
	assert.Equal(t, targetURL, originalURL)
}

func TestFileStorage_GetURL_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	storage := NewFileStorage(tempFile)

	_, err := storage.GetURL("non_existent_short_url")
	assert.Error(t, err)
}

func TestFileStorage_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	testRecord := URLMapping{
		ID:          "test-id",
		ShortURL:    "test-short",
		OriginalURL: "https://test-example.com",
	}

	file, err := os.OpenFile(tempFile, FileOpenFlagsWrite, FilePermissionsWrite)
	require.NoError(t, err)

	data, err := json.Marshal(testRecord)
	require.NoError(t, err)

	_, err = file.Write(append(data, byte('\n')))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	storage := NewFileStorage(tempFile)

	originalURL, err := storage.GetURL("test-short")
	require.NoError(t, err)
	assert.Equal(t, "https://test-example.com", originalURL)

	baseURL := "http://localhost:8080"
	targetURL := "https://new-example.com"

	shortURL, err := storage.SaveURL(baseURL, targetURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortURL)
}

func TestFileStorage_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	file, err := os.OpenFile(tempFile, FileOpenFlagsWrite, FilePermissionsWrite)
	require.NoError(t, err)

	_, err = file.WriteString("{ invalid json }\n")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	storage := NewFileStorage(tempFile)

	_, err = storage.GetURL("any-short-url")
	assert.Error(t, err)

	baseURL := "http://localhost:8080"
	targetURL := "https://example.com"

	shortURL, err := storage.SaveURL(baseURL, targetURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortURL)
}

func TestFileStorage_AddNewURLMapping(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_urls.json")

	storage := NewFileStorage(tempFile)

	shortURL := "test-short"
	originalURL := "https://example.com"

	err := storage.addNewURLMapping(shortURL, originalURL)
	require.NoError(t, err)

	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	file, err := os.Open(tempFile)
	require.NoError(t, err)
	defer file.Close()

	var record URLMapping
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&record)
	require.NoError(t, err)

	assert.Equal(t, shortURL, record.ShortURL)
	assert.Equal(t, originalURL, record.OriginalURL)
	assert.NotEmpty(t, record.ID)
}
