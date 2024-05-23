package ukuleleweb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	_ "embed"
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
			Input: "just (WikiLinkInBrackets)",
			Want:  `<p>just (<a href="/WikiLinkInBrackets">WikiLinkInBrackets</a>)</p>` + "\n",
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
			Input: "Hello go/link#with-anchor!",
			Want:  `<p>Hello <a href="http://go/link#with-anchor">go/link#with-anchor</a>!</p>` + "\n",
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
			Input: "A go/go-link and a subsequent [regular link](http://link/)!",
			Want:  `<p>A <a href="http://go/go-link">go/go-link</a> and a subsequent <a href="http://link/">regular link</a>!</p>` + "\n",
		},
		{
			Input: "A go/go-link and a subsequent WikiLink!",
			Want:  `<p>A <a href="http://go/go-link">go/go-link</a> and a subsequent <a href="/WikiLink">WikiLink</a>!</p>` + "\n",
		},
		{
			Input: "A WikiLink and a subsequent go/go-link!",
			Want:  `<p>A <a href="/WikiLink">WikiLink</a> and a subsequent <a href="http://go/go-link">go/go-link</a>!</p>` + "\n",
		},
		{
			Input: "A [regular link](http://link) and a subsequent WikiLink!",
			Want:  `<p>A <a href="http://link">regular link</a> and a subsequent <a href="/WikiLink">WikiLink</a>!</p>` + "\n",
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
			t.Errorf("RenderHTML(%q)\n\t   = %q,\n\twant %q", tt.Input, got, tt.Want)
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

func mustReadFile(t testing.TB, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("io.ReadAll(%q): %v", path, err)
	}
	return string(b)
}

func TestFullPageRendering(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("testdata", "wiki"))
	if err != nil {
		t.Fatalf("os.ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("Missing test data")
	}

	for _, dirent := range entries {
		var (
			pageName = dirent.Name()
			wikiPath = filepath.Join("testdata", "wiki", pageName)
			wantPath = filepath.Join("testdata", "want", pageName)
			md       = mustReadFile(t, wikiPath)
			want     = mustReadFile(t, wantPath)
		)

		t.Run(pageName, func(t *testing.T) {
			got, err := RenderHTML(md)
			if err != nil {
				t.Fatalf("RenderHTML: unexpected error: %v", err)
			}
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("RenderHTML: unexpected output (-got +want):\n%v", diff)
			}
		})
	}
}

//go:embed testdata/wiki/UkuleleWeb
var ukuleleWebPage string

func BenchmarkRender(b *testing.B) {
	for range b.N {
		_, _ = RenderHTML(ukuleleWebPage)
	}
}
