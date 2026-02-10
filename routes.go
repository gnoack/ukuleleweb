package ukuleleweb

import (
	"embed"
	"net/http"

	"github.com/peterbourgon/diskv/v3"
)

//go:embed static/*
var staticFiles embed.FS

type Config struct {
	MainPage string
	Store    *diskv.Diskv
}

func NewServer(cfg *Config) http.Handler {
	if cfg.MainPage == "" {
		cfg.MainPage = "MainPage"
	}

	mux := http.NewServeMux()
	addRoutes(mux, cfg)
	return noReferrer(mux)
}

// noReferrer wraps a handler to set a no-referrer Referrer-Policy header.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy
func noReferrer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "no-referrer")
		h.ServeHTTP(w, r)
	})
}

// addRoutes adds the Ukuleleweb routes to the given ServeMux.
func addRoutes(mux *http.ServeMux, cfg *Config) {
	mux.Handle("GET /static/", http.FileServer(http.FS(staticFiles)))
	mux.HandleFunc("POST /preview", previewHandler)

	handler := &PageHandler{
		MainPage: cfg.MainPage,
		D:        cfg.Store,
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+cfg.MainPage, http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /edit/{pageName}", handler.serveEdit)
	mux.HandleFunc("POST /{pageName}", handler.serveSave)
	mux.HandleFunc("GET /{pageName}", handler.serveView)
}
