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
	return mux
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
	mux.Handle("/{pageName}", handler)
	mux.Handle("/edit/{pageName}", handler)
}
