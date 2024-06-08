package ukuleleweb

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/peterbourgon/diskv/v3"
)

var (
	cssURL     = flag.String("brand.css_url", "/static/style.css", "The URL for the CSS file")
	faviconURL = flag.String("brand.favicon_url", "/static/favicon.svg", "The URL for the favicon")
)

//go:embed templates/*
var templateFiles embed.FS

var (
	baseTmpl = template.Must(template.New("layout").ParseFS(templateFiles, "templates/base/*.html"))
	PageTmpl = template.Must(template.Must(baseTmpl.Clone()).ParseFS(templateFiles, "templates/contents/page.html"))
	EditTmpl = template.Must(template.Must(baseTmpl.Clone()).ParseFS(templateFiles, "templates/contents/edit.html"))
)

type pageValues struct {
	Title         string
	PageName      string
	HTMLContent   template.HTML
	SourceContent string
	Error         string
	ReverseLinks  []string
	FaviconURL    string
	CSSURL        string
}

type PageHandler struct {
	MainPage string
	D        *diskv.Diskv

	// A cached version of the reverse links.
	revLinksMu sync.RWMutex
	revLinks   map[string][]string
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy
	w.Header().Set("Referrer-Policy", "no-referrer")

	pageName := r.PathValue("pageName")
	if !isPageName(pageName) {
		http.Error(w, "Invalid page name", http.StatusNotFound)
		return
	}

	tmpl := PageTmpl
	pv := &pageValues{
		Title:      insertSpaces(pageName),
		PageName:   pageName,
		FaviconURL: *faviconURL,
		CSSURL:     *cssURL,
	}

	if r.FormValue("edit") == "1" {
		tmpl = EditTmpl
		content := contentValue(r)
		if content == "" {
			content = h.D.ReadString(pageName)
		}
		pv.SourceContent = content

		// Search engines do not need to index the edit page.
		w.Header().Set("X-Robots-Tag", "noindex")
	} else if r.Method == "POST" {
		content := contentValue(r)
		err := h.D.WriteString(pageName, content)
		if err == nil { // Success saving! This is the default case.
			// TODO: Potentially do it in a background job?
			h.recalculateRevLinks()
			http.Redirect(w, r, "/"+pageName, http.StatusFound)
			return
		}
		// On error, render edit form with the error message.
		w.WriteHeader(http.StatusInternalServerError)
		tmpl = EditTmpl
		log.Printf("ERROR: diskv.WriteString(%q, ...): %v\n", pageName, err)
		pv.Error = "Internal error writing page"
		pv.SourceContent = content
	} else {
		tmpl = PageTmpl
		content := h.D.ReadString(pageName)
		rendered, err := RenderHTML(content)
		if err != nil {
			http.Error(w, "Failed to render markdown", http.StatusInternalServerError)
			return
		}
		pv.HTMLContent = template.HTML(rendered)
		pv.ReverseLinks = h.reverseLinks(pageName)

	}
	err := tmpl.Execute(w, pv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PageHandler) reverseLinks(pagename string) []string {
	h.revLinksMu.RLock()
	defer h.revLinksMu.RUnlock()

	if h.revLinks == nil {
		// Recalculate at read only if not done before.
		// This should only happen on startup.
		h.revLinksMu.RUnlock()
		h.recalculateRevLinks()
		h.revLinksMu.RLock()
	}
	return h.revLinks[pagename]
}

func (h *PageHandler) recalculateRevLinks() {
	rl := AllReverseLinks(h.D)

	h.revLinksMu.Lock()
	defer h.revLinksMu.Unlock()
	h.revLinks = rl
}

var fullPageNameRE = regexp.MustCompile(`^` + pageNameRE.String() + `$`)

// isPageName returns true iff pn is a camel case page name.
func isPageName(pn string) bool {
	return fullPageNameRE.MatchString(pn)
}

// insertSpaces inserts spaces before every non-leading capital letter
func insertSpaces(pn string) string {
	var bld strings.Builder
	bld.Grow(len(pn) + 3)
	for pos, rn := range pn {
		if unicode.IsUpper(rn) && pos > 0 {
			bld.WriteByte(' ')
		}
		// Guaranteed to not return errors.
		bld.WriteRune(rn)
	}
	return bld.String()
}

func contentValue(r *http.Request) string {
	return strings.ReplaceAll(r.FormValue("content"), "\r\n", "\n")
}
