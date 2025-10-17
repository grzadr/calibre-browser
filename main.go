package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/grzadr/calibre-browser/internal/arguments"
	"github.com/grzadr/calibre-browser/internal/booksdb"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

func createSearchHandler() http.HandlerFunc {
	// 1. Parse template at startup (happens once)
	search := template.Must(template.ParseFS(templateFiles,
		"templates/search-results.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		// 2. Get search query
		query := r.FormValue("search")

		entries := booksdb.GetBooksEntries()

		// 3. Perform search
		args := strings.Fields(query)
		results, _ := booksdb.SelectEntriesByTitleCommand(entries, args)

		log.Println("search completed")

		// 4. Execute template with results
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// THIS IS WHERE WE USE tmpl! ↓↓↓
		err := search.Execute(w, results)

		log.Println("search template executed")
		//     ^^^^^^^^^^^^^
		// Converts Go data → HTML using template
		if err != nil {
			log.Printf("template error: %v", err)
		}
	}
}

func createIndexHandler() http.HandlerFunc {
	// Parse template once at startup
	tmpl := template.Must(template.ParseFS(templateFiles,
		"templates/index.html"))

	// Pre-render template with initial data
	var buf bytes.Buffer

	entries := booksdb.GetBooksEntries()

	initialData := struct {
		Title     string
		BookCount int
		Generated time.Time // Added for the footer
	}{
		Title:     "Book Search",
		BookCount: entries.NumBooks(),
		Generated: time.Now(), // Server startup time
	}

	if err := tmpl.Execute(&buf, initialData); err != nil {
		log.Fatal("Failed to pre-render index template:", err)
	}

	// Pre-calculate values for efficiency
	content := buf.Bytes()
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(content))
	contentLength := strconv.Itoa(len(content))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)

			return
		}

		// Set headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
		w.Header().Set("Content-Length", contentLength)

		// Check if client has cached version
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)

			return
		}

		w.Write(content)
	}
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Method-based routing (Go 1.22+)
	mux.HandleFunc("GET /", createIndexHandler())
	mux.HandleFunc("POST /search", createSearchHandler())

	// FIX: Use fs.Sub to serve from the static subdirectory
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("Failed to create static sub-filesystem:", err)
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServerFS(staticFS)))

	return mux
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := arguments.ParseArgsServer(os.Args)
	if err != nil {
		log.Fatalln(fmt.Errorf("error parsing args: %w", err))
	}

	if err := booksdb.PopulateBooksRepository(conf.DbPath, ctx); err != nil {
		log.Fatalf("error initializng database %q: %s\n", conf.DbPath, err)
	}

	mux := setupRoutes()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// Each connection gets its own goroutine automatically
	}

	// This single ListenAndServe handles thousands of concurrent connections
	server.ListenAndServe()
}
