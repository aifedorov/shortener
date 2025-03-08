package repository

import (
	"bufio"
	"encoding/json"
	"github.com/aifedorov/shortener/pkg/logger"
	"os"

	"github.com/aifedorov/shortener/pkg/random"
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

type FileRepository struct {
	fname     string
	pathToURL []URLMapping
	rand      random.Randomizer
}

func NewFileRepository(filePath string) *FileRepository {
	return &FileRepository{
		fname:     filePath,
		pathToURL: make([]URLMapping, 0),
		rand:      random.NewService(),
	}
}

func (fs *FileRepository) Run() error {
	return nil
}

func (fs *FileRepository) Ping() error {
	return nil
}

func (fs *FileRepository) Close() error {
	return nil
}

func (fs *FileRepository) Get(shortURL string) (string, error) {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsRead, FilePermissionsRead)
	if err != nil {
		logger.Log.Error("repository: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("repository: failed to close file", zap.String("file", fs.fname), zap.Error(err))
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLMapping
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			logger.Log.Error("repository: failed to unmarshal record", zap.String("file", fs.fname), zap.Error(err))
			return "", err
		}
		if record.ShortURL == shortURL {
			return record.OriginalURL, nil
		}
	}

	logger.Log.Debug("repository: url is not found", zap.String("short_url", shortURL))
	return "", ErrShortURLNotFound
}

func (fs *FileRepository) Store(baseURL, targetURL string) (string, error) {
	alias, genErr := fs.rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Error("repository: generate random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	shortURL := baseURL + "/" + alias
	err := fs.addNewURLMapping(alias, targetURL)
	if err != nil {
		logger.Log.Error("repository: failed to add new url", zap.Error(err))
		return "", err
	}

	logger.Log.Debug("repository: saved url to file", zap.String("file", fs.fname), zap.String("res_url", shortURL))
	return shortURL, nil
}

func (fs *FileRepository) addNewURLMapping(shortURL string, originalURL string) error {
	file, err := os.OpenFile(fs.fname, FileOpenFlagsWrite, FilePermissionsWrite)
	if err != nil {
		logger.Log.Error("repository: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("repository: failed to close file", zap.String("file", fs.fname), zap.Error(err))
		}
	}()

	record := URLMapping{
		ID:          uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	data, err := json.Marshal(record)
	if err != nil {
		logger.Log.Error("repository: failed to marshal record", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	writer := bufio.NewWriter(file)

	if _, err := writer.Write(data); err != nil {
		logger.Log.Error("repository: failed to write data to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		logger.Log.Error("repository: failed to write newline to buffer", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	if err := writer.Flush(); err != nil {
		logger.Log.Error("repository: failed to flush buffer to disk", zap.String("file", fs.fname), zap.Error(err))
		return err
	}

	return nil
}
