package repository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

const (
	// FilePermissionsWrite defines the file permissions for write operations.
	FilePermissionsWrite = 0644
	// FilePermissionsRead defines the file permissions for read operations.
	FilePermissionsRead = 0444
	// FileOpenFlagsWrite defines the flags for opening files in write mode.
	FileOpenFlagsWrite = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	// FileOpenFlagsRead defines the flags for opening files in read mode.
	FileOpenFlagsRead = os.O_RDONLY
)

// URLMapping represents a single URL mapping stored in the file repository.
// It contains the user ID, short URL, and original URL for persistence.
type URLMapping struct {
	// ID is the user ID who created the URL mapping.
	ID string `json:"id"`
	// ShortURL is the generated short URL path.
	ShortURL string `json:"short_url"`
	// OriginalURL is the original URL that was shortened.
	OriginalURL string `json:"original_url"`
}

// FileRepository provides a file-based implementation of the Repository interface.
// It stores URL mappings in a JSON file with append-only writes for persistence.
type FileRepository struct {
	// fname is the path to the storage file.
	fname string
	// file is the open file handle for writing.
	file *os.File
	// pathToURL stores all URL mappings in memory for fast access.
	pathToURL []URLMapping
	// rand is used for generating random short URL identifiers.
	rand random.Randomizer
}

// NewFileRepository creates a new file-based repository instance.
// The repository will use the specified file path for persistence.
func NewFileRepository(filePath string) *FileRepository {
	return &FileRepository{
		fname:     filePath,
		pathToURL: make([]URLMapping, 0),
		rand:      random.NewService(),
	}
}

// Run initializes the file repository by opening the storage file.
func (fs *FileRepository) Run() error {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsWrite, FilePermissionsWrite)
	fs.file = file
	if err != nil {
		logger.Log.Error("fileStorage: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return err
	}
	return nil
}

// Ping checks the health of the file repository connection.
func (fs *FileRepository) Ping() error {
	return nil
}

// Close closes the file repository connection and performs cleanup.
func (fs *FileRepository) Close() error {
	err := fs.file.Close()
	if err != nil {
		logger.Log.Error("fileStorage: failed to close file", zap.String("file", fs.fname), zap.Error(err))
	}
	return nil
}

// Get retrieves the original URL for a given short URL from the file storage.
func (fs *FileRepository) Get(shortURL string) (string, error) {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsRead, FilePermissionsRead)
	if err != nil {
		logger.Log.Error("fileStorage: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("fileStorage: failed to close file", zap.String("file", fs.fname), zap.Error(err))
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLMapping
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			logger.Log.Error("fileStorage: failed to unmarshal record", zap.String("file", fs.fname), zap.Error(err))
			return "", err
		}
		if record.ShortURL == shortURL {
			return record.OriginalURL, nil
		}
	}

	logger.Log.Debug("fileStorage: url is not found", zap.String("short_url", shortURL))
	return "", ErrShortURLNotFound
}

// GetAll retrieves all URLs belonging to a specific user from the file storage.
func (fs *FileRepository) GetAll(_, _ string) ([]URLOutput, error) {
	panic("implement me")
}

// Store saves a new URL to the file storage and returns the generated short URL.
func (fs *FileRepository) Store(_, baseURL, targetURL string) (string, error) {
	alias, err := fs.rand.GenRandomString()
	if err != nil {
		logger.Log.Error("fileStorage: generate random string failed", zap.Error(err))
		return "", err
	}

	shortURL := baseURL + "/" + alias
	err = fs.addNewURL(alias, targetURL)
	if err != nil {
		logger.Log.Error("fileStorage: failed to add new url", zap.Error(err))
		return "", err
	}

	logger.Log.Debug("fileStorage: saved url to file", zap.String("file", fs.fname), zap.String("res_url", shortURL))
	return shortURL, nil
}

// StoreBatch saves multiple URLs to the file storage in a single operation.
func (fs *FileRepository) StoreBatch(_, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	logger.Log.Debug("fileStorage: storing batch of urls", zap.Int("count", len(urls)))
	res, err := fs.addNewURLs(baseURL, urls)
	if err != nil {
		logger.Log.Error("fileStorage: failed to add url", zap.Error(err))
		return nil, err
	}

	return res, nil
}

// DeleteBatch marks multiple URLs as deleted for a specific user.
func (fs *FileRepository) DeleteBatch(_ string, _ []string) error {
	panic("implement me")
}

// addNewURL adds a new URL mapping to the file storage.
func (fs *FileRepository) addNewURL(shortURL string, originalURL string) error {
	logger.Log.Debug("fileStorage: storing new url", zap.String("short_url", shortURL), zap.String("original_url", originalURL))
	record := URLMapping{
		ID:          uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	data, err := json.Marshal(record)
	if err != nil {
		logger.Log.Error("fileStorage: failed to marshal record", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	writer := bufio.NewWriter(fs.file)

	if _, err := writer.Write(data); err != nil {
		logger.Log.Error("fileStorage: failed to write data to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		logger.Log.Error("fileStorage: failed to write newline to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.Flush(); err != nil {
		logger.Log.Error("fileStorage: failed to flush buffer to disk", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	return nil
}

// addNewURLs adds multiple URL mappings to the file storage in batch.
func (fs *FileRepository) addNewURLs(baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
	res := make([]BatchURLOutput, len(urls))
	writer := bufio.NewWriter(fs.file)
	for i, url := range urls {
		alias, err := fs.rand.GenRandomString()
		if err != nil {
			logger.Log.Debug("fileStorage: generation of random string failed", zap.Error(err))
			return nil, err
		}

		record := URLMapping{
			ID:          uuid.New().String(),
			ShortURL:    alias,
			OriginalURL: url.OriginalURL,
		}

		logger.Log.Debug("fileStorage: storing new url", zap.String("short_url", url.OriginalURL), zap.String("original_url", url.OriginalURL))
		data, err := json.Marshal(record)
		if err != nil {
			logger.Log.Error("fileStorage: failed to marshal record", zap.String("file", fs.fname), zap.Error(err))
			return nil, err
		}

		if _, err := writer.Write(data); err != nil {
			logger.Log.Error("fileStorage: failed to write data to buffer", zap.String("file", fs.fname), zap.Error(err))
			return nil, err
		}

		if err := writer.WriteByte('\n'); err != nil {
			logger.Log.Error("fileStorage: failed to write newline to buffer", zap.String("file", fs.fname), zap.Error(err))
			return nil, err
		}

		resURL := baseURL + "/" + alias
		ou := BatchURLOutput{
			CID:      url.CID,
			ShortURL: resURL,
		}
		res[i] = ou
	}

	if err := writer.Flush(); err != nil {
		logger.Log.Error("fileStorage: failed to flush buffer to disk", zap.String("file", fs.fname), zap.Error(err))
		return nil, err
	}
	return res, nil
}
