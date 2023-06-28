package ukuleleweb

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	for _, tt := range []struct{ Input, Want string }{
		{
			Input: "Just a paragraph.",
			Want:  "<p>Just a paragraph.</p>\n",
		},
		{
			Input: "Hello *World*!",
			Want:  "<p>Hello <em>World</em>!</p>\n",
		},
		{
			Input: "Hello WikiLink!",
			Want:  `<p>Hello <a href="/WikiLink">WikiLink</a>!</p>` + "\n",
		},
		{
			Input: "WikiLink at start",
			Want:  `<p><a href="/WikiLink">WikiLink</a> at start</p>` + "\n",
		},
		{
			Input: "WikiLink and UkuleleLink",
			Want:  `<p><a href="/WikiLink">WikiLink</a> and <a href="/UkuleleLink">UkuleleLink</a></p>` + "\n",
		},
		{
			Input: "at the end a WikiLink",
			Want:  `<p>at the end a <a href="/WikiLink">WikiLink</a></p>` + "\n",
		},
		{
			Input: "at the end a   WikiLink",
			Want:  `<p>at the end a   <a href="/WikiLink">WikiLink</a></p>` + "\n",
		},
		{
			Input: `<a href="http://wiki/">To the wiki!</a>`,
			Want:  `<p><a href="http://wiki/">To the wiki!</a></p>` + "\n",
		},
		{
			Input: "Hello go/go-link!",
			Want:  `<p>Hello <a href="http://go/go-link">go/go-link</a>!</p>` + "\n",
		},
		{
			Input: "Hello go/long/go-link!",
			Want:  `<p>Hello <a href="http://go/long/go-link">go/long/go-link</a>!</p>` + "\n",
		},
		{
			// Should not recognize the inner mention of 'ExamplePage'.
			Input: `<a href="http://wiki/ExamplePage">To the wiki!</a>`,
			Want:  `<p><a href="http://wiki/ExamplePage">To the wiki!</a></p>` + "\n",
		},
		{
			Input: "[not a WikiLink](http://stuff/)",
			Want:  `<p><a href="http://stuff/">not a WikiLink</a></p>` + "\n",
		},
		{
			Input: "<!-- not a WikiLink -->",
			Want:  "<!-- not a WikiLink -->",
		},
	} {
		got, err := RenderHTML(tt.Input)
		if err != nil {
			t.Errorf("RenderHTML(%q): %v, want success", tt.Input, err)
		}
		if got != tt.Want {
			t.Errorf("RenderHTML(%q) = %q, want %q", tt.Input, got, tt.Want)
		}
	}
}

func TestOutgoingLinks(t *testing.T) {
	for _, tt := range []struct {
		Input string
		Want  []string
	}{
		{
			Input: "A WikiLink and AnotherOne.",
			Want:  []string{"AnotherOne", "WikiLink"},
		},
		{
			Input: "<!-- not a WikiLink -->",
			Want:  []string{},
		},
	} {
		gotMap := OutgoingLinks(tt.Input)
		got := sortedStringSlice(gotMap)

		if strings.Join(got, ",") != strings.Join(tt.Want, ",") {
			t.Errorf("OutgoingLinks(%q) = %v, want %v", tt.Input, got, tt.Want)
		}
	}
}
