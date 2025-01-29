package app

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var pathToURL = make(map[string]string)

func ShortUrlHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	if !isRequestBodyValid(string(body)) {
		http.Error(res, "The request must include only one URL formatted as follows: https://example.com.", http.StatusBadRequest)
		return
	}

	host := req.Host
	shortURL := genShortURL(string(body), host)
	if _, ok := pathToURL[shortURL]; ok {
		res.WriteHeader(http.StatusOK)
		return
	}

	pathToURL[shortURL] = string(body)

	res.WriteHeader(http.StatusCreated)
}

func genShortURL(url, host string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:8])
	return fmt.Sprintf("http://%s/%s", host, encoded)
}

func isRequestBodyValid(body string) bool {
	if len(body) == 0 {
		return false
	}

	match, err := regexp.MatchString(`^(https|http)://[a-z0-9]+\.+[a-z0-9]+\.*[a-z]+/*:*[0-9]*$`, body)
	if err != nil {
		return false
	}

	return match
}
