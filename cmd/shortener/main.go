package main

import (
	"log"
	"net/http"

	"github.com/aifedorov/shortener/internal/app"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.ShortUrlHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
