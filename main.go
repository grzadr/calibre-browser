package main

import (
	"embed"
	"fmt"
	"html"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Method-based routing (Go 1.22+)
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /search", handleSearch)

	// Static files with embedded FS
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServerFS(staticFiles)))

	return mux
}

func main() {
	mux := setupRoutes()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// Each connection gets its own goroutine automatically
	}

	// This single ListenAndServe handles thousands of concurrent connections
	server.ListenAndServe()
}
