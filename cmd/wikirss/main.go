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

var (
	title       = flag.String("wiki.title", "", "Wiki title")
	baseURL     = flag.String("wiki.baseURL", "", "Wiki base URL")
	description = flag.String("wiki.description", "", "Wiki description")
	maxItems    = flag.Int("feed.maxItems", 20, "Maximum number of items")
	suppress    = flag.String("suppress", "", "If set, suppress pages whose markdown contains this string")
)

func main() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(o, "  %s [FLAGS] [FILENAME...]\n\n", os.Args[0])
		fmt.Fprintf(o, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	filenames := flag.Args()
	if *title == "" {
		log.Fatalf("missing --wiki.title")
	}
	if *baseURL == "" {
		log.Fatalf("missing --wiki.baseURL")
	}

	type Page struct {
		Filename string
		Title    string
		Link     string
		PubDate  time.Time
		HTML     string
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
			Created:     st.BirthTime(),
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
