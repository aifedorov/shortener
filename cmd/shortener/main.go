package main

import (
	"github.com/aifedorov/shortener/internal/app"
)

func main() {
	server := app.NewServer()
	server.ListenAndServe()
}
