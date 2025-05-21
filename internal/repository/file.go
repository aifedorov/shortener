package repository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/aifedorov/shortener/pkg/random"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
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
	file      *os.File
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
	file, err := os.OpenFile(fs.fname, FileOpenFlagsWrite, FilePermissionsWrite)
	fs.file = file
	if err != nil {
		logger.Log.Error("fileStorage: failed to open file", zap.String("file", fs.fname), zap.Error(err))
		return err
	}
	return nil
}

func (fs *FileRepository) Ping() error {
	return nil
}

func (fs *FileRepository) Close() error {
	err := fs.file.Close()
	if err != nil {
		logger.Log.Error("fileStorage: failed to close file", zap.String("file", fs.fname), zap.Error(err))
	}
	return nil
}

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

func (fs *FileRepository) GetAll(userID, baseURL string) ([]URLOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (fs *FileRepository) Store(userID, baseURL, targetURL string) (string, error) {
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

func (fs *FileRepository) StoreBatch(userID, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
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

func (fs *FileRepository) DeleteBatch(userID string, aliases []string) error {
	//TODO implement me
	panic("implement me")
}

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
