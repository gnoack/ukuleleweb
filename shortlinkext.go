package ukuleleweb

import (
	"regexp"
	"slices"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var shortLinkRE = regexp.MustCompile(`^([a-z]+)/([^ \t\n]+)\b`)

// ShortLinkExt is a goldmark extension for recognizing shortlinks
// like go/links.
//
// A shortlink starts with a shortlink name (e.g. "go"),
// followed by a slash and a URL path.
//
// Valid shortlink examples:
//
// * go/foo
// * files/my/document.txt
// * wiki/FooBar
// * wiki/FooBar#Header1
type ShortLinkExt struct {
	// A map of short link prefixes to their full URL prefixes.
	Prefixes map[string]string
}

func (s *ShortLinkExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			// One less than the linkify one - we don't want to mess up http links.
			util.Prioritized(&shortLinkParser{prefixes: s.Prefixes}, 998),
		),
	)
}

// A parser for shortlinks like go/links
type shortLinkParser struct {
	prefixes map[string]string
}

func (s *shortLinkParser) Trigger() []byte {
	return []byte{' ', '('}
}

func (s *shortLinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) (res ast.Node) {
	if pc.IsInLinkLabel() {
		return nil
	}
	line, segment := block.PeekLine()
	// Implementation note:
	// The trigger above triggers for the given characters, as well as for newlines.
	// Parse() below must be able to recognize both lines starting with "go/link..."
	// as well as lines starting with " go/link..." (for any leading trigger character).
	// If the line does start with a trigger, then *on a successful parse*,
	// that trigger must be inserted into the parent node before returning.
	if len(line) > 0 && slices.Contains(s.Trigger(), line[0]) {
		prefixSeg := segment.WithStop(segment.Start + 1)

		// Move line and segment one character further
		// and continue the parsing as if we had not started with a space.
		// e.g. line = "go/foo ..." instead of " go/foo ..."
		block.Advance(1)
		line = line[1:]
		segment = segment.WithStart(segment.Start + 1)

		// Insert the leading space into the parent AST, if parse was a success.
		defer func() {
			if res == nil {
				return
			}
			ast.MergeOrAppendTextSegment(parent, prefixSeg)
		}()
	}

	// Match must be at the beginning of the line either way.
	m := shortLinkRE.FindSubmatchIndex(line)
	if m == nil || m[0] != 0 {
		return nil
	}
	shortlinkKey := line[m[2]:m[3]]
	shortlinkPath := line[m[4]:m[5]]
	shortlinkDomain, ok := s.prefixes[string(shortlinkKey)]
	if !ok {
		return nil
	}

	block.Advance(m[1])

	link := ast.NewLink()
	link.AppendChild(link, ast.NewTextSegment(text.NewSegment(segment.Start, segment.Start+m[1])))
	link.Destination = append([]byte(shortlinkDomain), shortlinkPath...)
	return link
}
