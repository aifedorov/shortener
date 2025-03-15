package save

import (
	"encoding/json"
	"github.com/aifedorov/shortener/pkg/validate"
	"net/http"

	"github.com/aifedorov/shortener/internal/repository"
)

func decodeRequest(req *http.Request) ([]BatchRequest, error) {
	var reqURLs []BatchRequest
	if err := json.NewDecoder(req.Body).Decode(&reqURLs); err != nil {
		return nil, err
	}
	return reqURLs, nil
}

func encodeResponse(rw http.ResponseWriter, resURL string) error {
	encoder := json.NewEncoder(rw)
	resp := Response{
		ShortURL: resURL,
	}

	if err := encoder.Encode(resp); err != nil {
		return err
	}
	return nil
}

func encodeBatchResponse(rw http.ResponseWriter, urls []repository.BatchURLOutput) error {
	encoder := json.NewEncoder(rw)
	resp := make([]BatchResponse, len(urls))
	for url := range urls {
		r := BatchResponse{
			CID:      urls[url].CID,
			ShortURL: urls[url].ShortURL,
		}
		resp[url] = r
	}

	if err := encoder.Encode(resp); err != nil {
		return err
	}
	return nil
}

func validateURLs(reqURLs []BatchRequest, urlChecker validate.URLChecker) ([]repository.BatchURLInput, error) {
	var urls = make([]repository.BatchURLInput, len(reqURLs))
	for i, reqBodyURL := range reqURLs {
		if err := urlChecker.CheckURL(reqBodyURL.OriginalURL); err != nil {
			return nil, err
		}
		urls[i] = repository.BatchURLInput{
			CID:         reqBodyURL.CID,
			OriginalURL: reqBodyURL.OriginalURL,
		}
	}
	return urls, nil
}
