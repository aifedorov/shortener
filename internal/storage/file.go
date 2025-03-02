package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/lib/random"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	FilePermissionsWrite = 0644
	FilePermissionsRead  = 0444
	FileOpenFlagsWrite   = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	FileOpenFlagsRead    = os.O_RDONLY
)

type URLMapping struct {
	ID          string `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	fname     string
	pathToURL []URLMapping
	rand      random.Randomizer
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		fname:     filePath,
		pathToURL: make([]URLMapping, 0),
		rand:      random.NewService(),
	}
}

func (fs *FileStorage) addNewURLMapping(shortURL string, originalURL string) error {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsWrite, FilePermissionsWrite)
	if err != nil {
		logger.Log.Error("storage: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("storage: failed to close file", zap.String("file", fs.fname), zap.Error(err))
		}
	}()

	record := URLMapping{
		ID:          uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	data, err := json.Marshal(record)
	if err != nil {
		logger.Log.Error("storage: failed to marshal record", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	writer := bufio.NewWriter(file)

	if _, err := writer.Write(data); err != nil {
		logger.Log.Error("storage: failed to write data to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		logger.Log.Error("storage: failed to write newline to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.Flush(); err != nil {
		logger.Log.Error("storage: failed to flush buffer to disk", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	return nil
}

func (fs *FileStorage) GetURL(shortURL string) (string, error) {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsRead, FilePermissionsRead)
	if err != nil {
		logger.Log.Error("storage: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("storage: failed to close file", zap.String("file", fs.fname), zap.Error(err))
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLMapping
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			logger.Log.Error("storage: failed to unmarshal record", zap.String("file", fs.fname), zap.Error(err))
			return "", err
		}
		if record.ShortURL == shortURL {
			return record.OriginalURL, nil
		}
	}

	logger.Log.Debug("storage: url is not found", zap.String("short_url", shortURL))
	return "", ErrShortURLNotFound
}

func (fs *FileStorage) SaveURL(baseURL, targetURL string) (string, error) {
	alias, genErr := fs.rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Error("storage: generate random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	shortURL := baseURL + "/" + alias
	err := fs.addNewURLMapping(alias, targetURL)
	if err != nil {
		logger.Log.Error("storage: failed to add new url", zap.Error(err))
		return "", err
	}

	logger.Log.Debug("storage: saved url to file", zap.String("file", fs.fname), zap.String("res_url", shortURL))
	return shortURL, nil
}
