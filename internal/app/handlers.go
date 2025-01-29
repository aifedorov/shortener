package app

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// TODO: It's not thread safe collection.
var pathToURL = make(map[string]string)

func ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		methodPostHandler(res, req)
	case http.MethodGet:
		methodGetHandler(res, req)
	default:
		http.Error(res, "Only GET/POST requests are allowed.", http.StatusBadRequest)
	}
}

func methodPostHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	if !isRequestBodyValid(string(body)) {
		http.Error(res, "The request must include only one URL formatted as follows: https://example.com.", http.StatusBadRequest)
		return
	}

	host := req.Host
	shortURL := genShortURL(string(body))
	resURL := fmt.Sprintf("http://%s/%s", host, shortURL)

	if _, ok := pathToURL[shortURL]; ok {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte(resURL))

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	pathToURL[shortURL] = string(body)
	res.WriteHeader(http.StatusCreated)

	_, writeErr := res.Write([]byte(resURL))
	if writeErr != nil {
		http.Error(res, writeErr.Error(), http.StatusBadRequest)
		return
	}
}

func methodGetHandler(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	if path == "" {
		http.Error(res, "Short URL is missing.", http.StatusBadRequest)
		return
	}

	targetURL, exists := pathToURL[path]
	if !exists {
		http.Error(res, "URL doesn't found.", http.StatusBadRequest)
		return
	}

	http.Redirect(res, req, targetURL, http.StatusTemporaryRedirect)
}

func genShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
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
