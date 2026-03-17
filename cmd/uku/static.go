package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnoack/ukuleleweb"
)

// wikiLinkRE matches href attributes pointing to wiki pages (e.g. href="/PageName").
var wikiLinkRE = regexp.MustCompile(`href="/([A-Z][a-zA-Z0-9]*)"`)

func runStatic(args []string) {
	fs := flag.NewFlagSet("uku static", flag.ExitOnError)
	fs.Usage = func() {
		o := fs.Output()
		fmt.Fprintf(o, "Usage: uku static [FLAGS] [FILENAME...]\n\n")
		fmt.Fprintf(o, "Render wiki pages to a static website.\n\n")
		fmt.Fprintf(o, "Flags:\n")
		fs.PrintDefaults()
	}

	outDir := fs.String("out_dir", "", "Output directory for static files")
	baseURL := fs.String("base_url", "/", "Base URL where the site will be deployed")
	siteTitle := fs.String("site_title", "", "Site title appended to each page's <title>")
	urlStyle := fs.String("url_style", "dir", `URL style: "dir" (PageName/index.html) or "flat" (PageName.html)`)

	fs.Parse(args)

	if *outDir == "" {
		log.Fatalf("missing --out_dir")
	}
	if *urlStyle != "dir" && *urlStyle != "flat" {
		log.Fatalf(`--url_style must be "dir" or "flat"`)
	}

	// Canonicalize base URL (must end with /)
	*baseURL, _ = strings.CutSuffix(*baseURL, "/")
	*baseURL = *baseURL + "/"

	cssURL := *baseURL + "static/style.css"
	faviconURL := *baseURL + "static/favicon.svg"

	if err := ukuleleweb.WriteStaticAssets(*outDir); err != nil {
		log.Fatalf("WriteStaticAssets: %v", err)
	}

	for _, fn := range fs.Args() {
		pageName := filepath.Base(fn)
		md, err := os.ReadFile(fn)
		if err != nil {
			log.Fatalf("ReadFile(%q): %v", fn, err)
		}

		html, err := ukuleleweb.RenderHTML(string(md))
		if err != nil {
			log.Fatalf("RenderHTML(%q): %v", pageName, err)
		}
		html = rewriteLinks(html, *baseURL, *urlStyle)

		outPath := pageOutputPath(*outDir, pageName, *urlStyle)
		if err := os.MkdirAll(filepath.Dir(outPath), 0777); err != nil {
			log.Fatalf("MkdirAll(%q): %v", filepath.Dir(outPath), err)
		}

		if err := writeStaticPage(outPath, ukuleleweb.StaticPageValues{
			Title:       ukuleleweb.ToTitle(pageName),
			SiteTitle:   *siteTitle,
			HTMLContent: template.HTML(html),
			CSSURL:      cssURL,
			FaviconURL:  faviconURL,
		}); err != nil {
			log.Fatalf("writeStaticPage(%q): %v", pageName, err)
		}
	}
}

func writeStaticPage(path string, values ukuleleweb.StaticPageValues) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return ukuleleweb.StaticPageTmpl.Execute(f, values)
}

func pageOutputPath(outDir, pageName, urlStyle string) string {
	if urlStyle == "flat" {
		return filepath.Join(outDir, pageName+".html")
	}
	return filepath.Join(outDir, pageName, "index.html")
}

func rewriteLinks(html, baseURL, urlStyle string) string {
	return wikiLinkRE.ReplaceAllStringFunc(html, func(match string) string {
		pageName := wikiLinkRE.FindStringSubmatch(match)[1]
		if urlStyle == "flat" {
			return fmt.Sprintf(`href="%s%s.html"`, baseURL, pageName)
		}
		return fmt.Sprintf(`href="%s%s"`, baseURL, pageName)
	})
}
