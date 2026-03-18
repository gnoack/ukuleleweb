package ukuleleweb

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"

	_ "embed"
)

var update = flag.Bool("update", false, "update golden files")

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
			Input: "Hello unknown/go-link!",
			Want:  `<p>Hello unknown/go-link!</p>` + "\n",
		},
		{
			Input: "Hello go/link.with.dots!",
			Want:  `<p>Hello <a href="http://go/link.with.dots">go/link.with.dots</a>!</p>` + "\n",
		},
		{
			Input: "A link: go/link.with.dots.",
			Want:  `<p>A link: <a href="http://go/link.with.dots">go/link.with.dots</a>.</p>` + "\n",
		},
		{
			Input: "A link: go/link.with.dots, right?",
			Want:  `<p>A link: <a href="http://go/link.with.dots">go/link.with.dots</a>, right?</p>` + "\n",
		},
		{
			Input: "A link: (go/link.in.brackets).",
			Want:  `<p>A link: (<a href="http://go/link.in.brackets">go/link.in.brackets</a>).</p>` + "\n",
		},
		{
			Input: "A link: go/a/b/c.",
			Want:  `<p>A link: <a href="http://go/a/b/c">go/a/b/c</a>.</p>` + "\n",
		},
		{
			Input: "A link: go/a.b/c.",
			Want:  `<p>A link: <a href="http://go/a.b/c">go/a.b/c</a>.</p>` + "\n",
		},
		{
			Input: "A link with comma: go/a,b/c.",
			Want:  `<p>A link with comma: <a href="http://go/a,b/c">go/a,b/c</a>.</p>` + "\n",
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
			Input: "<!-- not a WikiLink -->\n",
			Want:  "<!-- not a WikiLink -->\n",
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
		pageName := dirent.Name()
		t.Run(pageName, func(t *testing.T) {
			var (
				wikiPath = filepath.Join("testdata", "wiki", pageName)
				wantPath = filepath.Join("testdata", "want", pageName)
				md       = mustReadFile(t, wikiPath)
			)
			got, err := RenderHTML(md)
			if err != nil {
				t.Fatalf("RenderHTML: unexpected error: %v", err)
			}
			if *update {
				if err := os.WriteFile(wantPath, []byte(got), 0666); err != nil {
					t.Fatalf("os.WriteFile(%q): %v", wantPath, err)
				}
				return
			}
			want := mustReadFile(t, wantPath)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("RenderHTML: unexpected output (-got +want):\n%v", diff)
			}
		})
	}
}

func staticDestFunc(baseURL, urlStyle string) func(string) string {
	return func(pageName string) string {
		if urlStyle == "flat" {
			return baseURL + pageName + ".html"
		}
		return baseURL + pageName
	}
}

func renderWithGoldmark(t *testing.T, gmark goldmark.Markdown, md string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := gmark.Convert([]byte(md), &buf); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	return buf.String()
}

func TestStaticDestFuncDefaultMatchesRenderHTML(t *testing.T) {
	// With baseURL="/" and urlStyle="dir", staticDestFunc should produce
	// identical output to RenderHTML since link destinations are unchanged.
	entries, err := os.ReadDir(filepath.Join("testdata", "wiki"))
	if err != nil {
		t.Fatalf("os.ReadDir: %v", err)
	}
	for _, dirent := range entries {
		pageName := dirent.Name()
		md := mustReadFile(t, filepath.Join("testdata", "wiki", pageName))
		t.Run(pageName, func(t *testing.T) {
			want, err := RenderHTML(md)
			if err != nil {
				t.Fatalf("RenderHTML: %v", err)
			}
			got := renderWithGoldmark(t, NewGoldmark(staticDestFunc("/", "dir")), md)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("NewGoldmark(staticDestFunc(\"/\", \"dir\")) != RenderHTML (-got +want):\n%v", diff)
			}
		})
	}
}

func TestRenderStaticHTMLGolden(t *testing.T) {
	for _, tt := range []struct {
		pageName string
		baseURL  string
		urlStyle string
	}{
		{"UkuleleWeb", "/", "flat"},
		{"UkuleleWeb", "https://example.com/wiki/", "dir"},
	} {
		goldenName := tt.pageName + "." + tt.urlStyle
		t.Run(goldenName, func(t *testing.T) {
			md := mustReadFile(t, filepath.Join("testdata", "wiki", tt.pageName))
			got := renderWithGoldmark(t, NewGoldmark(staticDestFunc(tt.baseURL, tt.urlStyle)), md)
			wantPath := filepath.Join("testdata", "static_want", goldenName)
			if *update {
				if err := os.MkdirAll(filepath.Dir(wantPath), 0777); err != nil {
					t.Fatalf("os.MkdirAll: %v", err)
				}
				if err := os.WriteFile(wantPath, []byte(got), 0666); err != nil {
					t.Fatalf("os.WriteFile(%q): %v", wantPath, err)
				}
				return
			}
			want := mustReadFile(t, wantPath)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("RenderStaticHTML(%q, %q, %q): unexpected output (-got +want):\n%v",
					tt.pageName, tt.baseURL, tt.urlStyle, diff)
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
