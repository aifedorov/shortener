package main

import (
	"github.com/aifedorov/shortener/internal/http"
)

func main() {
	srv := server.NewServer()
	srv.ListenAndServe()
}
