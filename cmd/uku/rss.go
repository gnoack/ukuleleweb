package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/gnoack/ukuleleweb"
	"github.com/gorilla/feeds"
)

func runRss(args []string) {
	fs := flag.NewFlagSet("uku rss", flag.ExitOnError)
	fs.Usage = func() {
		o := fs.Output()
		fmt.Fprintf(o, "Usage: uku rss [FLAGS] [FILENAME...]\n\n")
		fmt.Fprintf(o, "Flags:\n")
		fs.PrintDefaults()
	}

	title := fs.String("wiki.title", "", "Wiki title")
	baseURL := fs.String("wiki.base_url", "", "Wiki base URL")
	description := fs.String("wiki.description", "", "Wiki description")
	maxItems := fs.Int("feed.max_items", 20, "Maximum number of items")
	suppress := fs.String("feed.suppress", "", "If set, suppress pages whose markdown contains this string")

	fs.Parse(args)

	filenames := fs.Args()
	if *title == "" {
		log.Fatalf("missing --wiki.title")
	}
	if *baseURL == "" {
		log.Fatalf("missing --wiki.base_url")
	}

	// Canonicalize base URL (must end with /)
	*baseURL, _ = strings.CutSuffix(*baseURL, "/")
	*baseURL = *baseURL + "/"

	feed := feeds.Feed{
		Title:       *title,
		Link:        &feeds.Link{Href: *baseURL},
		Description: *description,
		Created:     time.Now(),
		// XXX: Language
	}
	for _, fn := range filenames {
		st, err := times.Stat(fn)
		if err != nil {
			log.Fatalf("Stat(%q): %v\n", fn, err)
			continue
		}
		md, err := os.ReadFile(fn)
		if err != nil {
			log.Fatalf("ReadFile(%q): %v\n", fn, err)
			continue
		}

		if len(*suppress) > 0 && strings.Contains(string(md), *suppress) {
			continue // Skip draft pages
		}

		html, err := ukuleleweb.RenderHTML(string(md))
		if err != nil {
			log.Fatalf("RenderHTML(ReadFile(%q)): %v", fn, err)
			continue
		}
		if len(html) == 0 {
			continue // Skip empty pages
		}

		if st.ModTime().After(feed.Updated) {
			feed.Updated = st.ModTime()
		}
		feed.Add(&feeds.Item{
			Title:       ukuleleweb.ToTitle(filepath.Base(fn)),
			Link:        &feeds.Link{Href: *baseURL + filepath.Base(fn)},
			Created:     earlier(st.BirthTime(), st.ModTime()),
			Updated:     st.ModTime(),
			Description: html,
		})
	}
	slices.SortFunc(feed.Items, func(a, b *feeds.Item) int {
		if a.Created.After(b.Created) {
			return -1
		} else {
			return 1
		}
	})
	if len(feed.Items) > *maxItems {
		feed.Items = feed.Items[:*maxItems]
	}

	err := feed.WriteRss(os.Stdout)
	if err != nil {
		log.Fatalf("template rendering: %v", err)
	}
}

func earlier(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
