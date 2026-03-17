package ukuleleweb

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

// StaticPageValues holds template data for a rendered static wiki page.
type StaticPageValues struct {
	Title       string
	SiteTitle   string
	HTMLContent template.HTML
	CSSURL      string
	FaviconURL  string
}

// StaticPageTmpl is the template for rendering a static wiki page.
var StaticPageTmpl = template.Must(
	template.ParseFS(templateFiles, "templates/static/page.html"),
)

// StaticDestFunc returns a destFunc for NewGoldmark that rewrites wiki link
// destinations for static site deployment. baseURL is prepended to each page
// name; if urlStyle is "flat", a ".html" suffix is appended.
func StaticDestFunc(baseURL, urlStyle string) func(string) string {
	return func(pageName string) string {
		if urlStyle == "flat" {
			return baseURL + pageName + ".html"
		}
		return baseURL + pageName
	}
}

// WriteStaticAssets copies the embedded static resources to dir/static/.
func WriteStaticAssets(dir string) error {
	return fs.WalkDir(staticFiles, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(dest, 0777)
		}
		// wiki.js is only needed for the dynamic wiki (editor form submission).
		if d.Name() == "wiki.js" {
			return nil
		}
		data, err := staticFiles.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0666)
	})
}
