package ukuleleweb

import (
	"embed"
	"net/http"

	"github.com/peterbourgon/diskv/v3"
)

//go:embed static/*
var staticFiles embed.FS

func NewServer(mainPage string, d *diskv.Diskv) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, mainPage, d)
	return mux
}

// addRoutes adds the Ukuleleweb routes to the given ServeMux.
func addRoutes(mux *http.ServeMux, mainPage string, d *diskv.Diskv) {
	mux.Handle("GET /static/", http.FileServer(http.FS(staticFiles)))

	handler := &PageHandler{
		MainPage: mainPage,
		D:        d,
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+mainPage, http.StatusMovedPermanently)
	})
	mux.Handle("/{pageName}", handler)
	mux.Handle("/edit/{pageName}", handler)
}
