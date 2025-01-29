package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

var pathToURL map[string]string = make(map[string]string)

func genShortURL(url, host string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:8])
	return fmt.Sprintf("http://%s/%s", host, encoded)
}

func shortUrlHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	host := req.Host

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	// TODO: Validate body - not empty and valid URL.

	if len(body) == 0 {
		http.Error(res, "Empty request body.", http.StatusBadRequest)
		return
	}

	shortURL := genShortURL(string(body), host)
	if _, ok := pathToURL[shortURL]; ok {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(shortURL))
		return
	}

	pathToURL[shortURL] = string(body)

	res.WriteHeader(http.StatusCreated)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortUrlHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
