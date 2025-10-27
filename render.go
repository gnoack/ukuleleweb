package ukuleleweb

import (
	"bytes"
	"flag"
	"log"
	"sort"
	"strings"
	"sync"

	shortlink "github.com/gnoack/goldmark-shortlink"
	pikchr "github.com/gopikchr/goldmark-pikchr"
	attributes "github.com/mdigger/goldmark-attributes"
	"github.com/peterbourgon/diskv/v3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var (
	once  sync.Once
	gmark goldmark.Markdown

	shortlinkPrefixes = flag.String("md.shortlinks", "go=http://go/", "Accepted shortlink prefixes, comma-separated list of prefix=URL pairs")
)

func mustParseShortlinkFlag(sl string) map[string]string {
	res := make(map[string]string)
	for _, pair := range strings.Split(sl, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid shortlink pair %q, expected a key=value pair", pair)
		}
		// Sanitize URLs a bit, it should be possible to pass bare domain names
		// and URLs that are missing the trailing slash.
		url := parts[1]
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		res[strings.TrimSpace(parts[0])] = strings.TrimSpace(url)
	}
	return res
}

func wikiGmark() goldmark.Markdown {
	once.Do(func() {
		gmark = goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Typographer,
				WikiLinkExt,
				&shortlink.Extender{Prefixes: mustParseShortlinkFlag(*shortlinkPrefixes)},
				&pikchr.Extender{},
				attributes.Extension,
			),
			goldmark.WithParserOptions(
				parser.WithAttribute(),
			),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		)
	})
	return gmark
}

func RenderHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := wikiGmark().Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

// OutgoingLinks returns the outgoing wiki links in a given Markdown input.
// The outgoing links are a map of page names to true.
func OutgoingLinks(md string) map[string]bool {
	found := make(map[string]bool)
	reader := text.NewReader([]byte(md))
	doc := wikiGmark().Parser().Parse(reader)
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		l, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		URL := string(l.Destination)
		if strings.HasPrefix(URL, "/") {
			found[URL[1:]] = true
		}
		return ast.WalkContinue, nil
	})
	return found
}

// AllReverseLinks calculates the reverse link map for the whole wiki.
// The returned map maps page names to a list of pages linking to it.
// Sets of pages are represented as sorted lists.
func AllReverseLinks(d *diskv.Diskv) map[string][]string {
	revLinks := make(map[string]map[string]bool)
	for p := range d.Keys(nil) {
		pOut := OutgoingLinks(d.ReadString(p))
		for q := range pOut {
			qIn, ok := revLinks[q]
			if !ok {
				qIn = make(map[string]bool)
				revLinks[q] = qIn
			}
			qIn[p] = true
		}
	}

	revLinksSorted := make(map[string][]string)
	for p, s := range revLinks {
		revLinksSorted[p] = sortedStringSlice(s)
	}
	return revLinksSorted
}

func sortedStringSlice(a map[string]bool) []string {
	var res []string
	for k := range a {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}
