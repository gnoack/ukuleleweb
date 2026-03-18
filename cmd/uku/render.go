package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/gnoack/ukuleleweb"
)

// templateValues holds template data for uku render --out.template.
type templateValues struct {
	Title       string
	SiteTitle   string
	HTMLContent template.HTML
}

func runRender(args []string) {
	fs := flag.NewFlagSet("uku render", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: uku render [FLAGS] FILENAME\n\n")
		fmt.Fprintf(fs.Output(), "Render a markdown file to HTML.\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	tmplFile := fs.String("out.template", "", "Go html/template file to wrap rendered content (variables: .Title, .SiteTitle, .HTMLContent)")
	baseURL := fs.String("wiki.base_url", "/", "Base URL for wiki link rewriting")
	urlStyle := fs.String("out.url_style", "dir", `URL style for wiki links: "dir" (PageName/) or "flat" (PageName.html)`)
	siteTitle := fs.String("wiki.title", "", "Site title, exposed as .SiteTitle in the template")

	fs.Parse(args)

	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(1)
	}

	if *urlStyle != "dir" && *urlStyle != "flat" {
		log.Fatalf(`--out.url_style must be "dir" or "flat"`)
	}
	if *siteTitle != "" && *tmplFile == "" {
		log.Fatalf("--wiki.title requires --out.template")
	}

	var tmpl *template.Template
	if *tmplFile != "" {
		var err error
		tmpl, err = template.ParseFiles(*tmplFile)
		if err != nil {
			log.Fatalf("template.ParseFiles(%q): %v", *tmplFile, err)
		}
	}

	gmark := ukuleleweb.NewGoldmark(staticDestFunc(*baseURL, *urlStyle))

	fn := fs.Arg(0)
	md, err := os.ReadFile(fn)
	if err != nil {
		log.Fatalf("ReadFile(%q): %v", fn, err)
	}

	var buf bytes.Buffer
	if err := gmark.Convert(md, &buf); err != nil {
		log.Fatalf("gmark.Convert(%q): %v", fn, err)
	}

	if tmpl != nil {
		values := templateValues{
			Title:       ukuleleweb.ToTitle(filepath.Base(fn)),
			SiteTitle:   *siteTitle,
			HTMLContent: template.HTML(buf.String()),
		}
		if err := tmpl.Execute(os.Stdout, values); err != nil {
			log.Fatalf("tmpl.Execute(%q): %v", fn, err)
		}
	} else {
		fmt.Print(buf.String())
	}
}

// staticDestFunc returns a destFunc for NewGoldmark that rewrites wiki link
// destinations for static site deployment. baseURL is prepended to each page
// name; if urlStyle is "flat", a ".html" suffix is appended.
func staticDestFunc(baseURL, urlStyle string) func(string) string {
	return func(pageName string) string {
		if urlStyle == "flat" {
			return baseURL + pageName + ".html"
		}
		return baseURL + pageName
	}
}
