package ukuleleweb

import (
	"bytes"
	"sort"
	"strings"

	"github.com/gopikchr/goldmark-pikchr"
	"github.com/peterbourgon/diskv/v3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var gmark = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Typographer,
		WikiLinkExt,
		&pikchr.Extender{},
	),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func RenderHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := gmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

// OutgoingLinks returns the outgoing wiki links in a given Markdown input.
// The outgoing links are a map of page names to true.
func OutgoingLinks(md string) map[string]bool {
	found := make(map[string]bool)
	reader := text.NewReader([]byte(md))
	doc := gmark.Parser().Parse(reader)
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
