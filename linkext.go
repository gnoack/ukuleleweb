package ukuleleweb

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	pageNameRE = regexp.MustCompile(`\b([A-ZÄÖÜ][a-zäöüß]+){2,}\b`)
	goLinkRE   = regexp.MustCompile(`\bgo/[A-Za-z0-9/_+öäüÖÄÜß-]+\b`)
)

// wikiLinkExt is a goldmark extension for recognizing WikiLinks and go/links.
type wikiLinkExt struct{}

var WikiLinkExt = &wikiLinkExt{}

func (e *wikiLinkExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			// One less than the linkify one - we don't want to mess up http links.
			util.Prioritized(&wikiLinkParser{}, 998),
		),
	)
}

// A parser for WikiLinks (resolving to /WikiLinks) and go/links
// (resolving to http://go/links).
type wikiLinkParser struct{}

func (w *wikiLinkParser) Trigger() []byte {
	return []byte{' '}
}

func (s *wikiLinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) (res ast.Node) {
	if pc.IsInLinkLabel() {
		return nil
	}
	line, segment := block.PeekLine()

	// Implementation note:
	// The trigger ' ' above triggers for any space character, as well as for newlines.
	// Parse() below must be able to recognize both lines starting with "WikiLink..."
	// as well as lines starting with " WikiLink..." (for any leading space character).
	// If the line does start with a space, then *on a successful parse*,
	// that space must be inserted into the parent node before returning.
	if len(line) > 0 && util.IsSpace(line[0]) {
		spaceSeg := segment.WithStop(segment.Start + 1)

		// Move line and segment one character further
		// and continue the parsing as if we had not started with a space.
		block.Advance(1)
		line = line[1:]
		segment = segment.WithStart(segment.Start + 1)

		// Insert the leading space into the parent AST, if parse was a success.
		defer func() {
			if res == nil {
				return
			}
			ast.MergeOrAppendTextSegment(parent, spaceSeg)
		}()
	}

	m := pageNameRE.FindSubmatchIndex(line)
	if m == nil {
		m = goLinkRE.FindSubmatchIndex(line)
	}
	if m == nil || m[0] != 0 {
		return nil
	}

	linkText := line[0:m[1]]

	block.Advance(m[1])

	link := ast.NewLink()
	link.AppendChild(link, ast.NewTextSegment(text.NewSegment(segment.Start, segment.Start+m[1])))
	if bytes.HasPrefix(linkText, []byte("go/")) {
		link.Destination = append([]byte("http://"), linkText...)
	} else {
		link.Destination = append([]byte{'/'}, linkText...)
	}
	return link
}
