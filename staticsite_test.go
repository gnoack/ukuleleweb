package ukuleleweb

import (
	"bytes"
	"flag"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
)

func renderStatic(t *testing.T, gmark goldmark.Markdown, md string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := gmark.Convert([]byte(md), &buf); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	return buf.String()
}

var update = flag.Bool("update", false, "update golden files")

func TestWriteStaticAssets(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStaticAssets(dir); err != nil {
		t.Fatalf("WriteStaticAssets: %v", err)
	}

	for _, fn := range []string{"static/style.css", "static/favicon.svg"} {
		if _, err := os.Stat(filepath.Join(dir, fn)); err != nil {
			t.Errorf("WriteStaticAssets: missing file %q", fn)
		}
	}

	// wiki.js is only used by the dynamic wiki editor, not static pages.
	if _, err := os.Stat(filepath.Join(dir, "static/wiki.js")); !os.IsNotExist(err) {
		t.Error("WriteStaticAssets: unexpectedly copied wiki.js")
	}
}

func TestStaticDestFuncDefaultMatchesRenderHTML(t *testing.T) {
	// With baseURL="/" and urlStyle="dir", RenderStaticHTML should produce
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
			got := renderStatic(t, NewGoldmark(StaticDestFunc("/", "dir")), md)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("NewGoldmark(StaticDestFunc(\"/\", \"dir\")) != RenderHTML (-got +want):\n%v", diff)
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
			got := renderStatic(t, NewGoldmark(StaticDestFunc(tt.baseURL, tt.urlStyle)), md)
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

func TestStaticPageTmpl(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values StaticPageValues
		want   []string
	}{
		{
			name: "BasicPage",
			values: StaticPageValues{
				Title:       "Hello World",
				HTMLContent: template.HTML("<p>Body text.</p>"),
				CSSURL:      "/static/style.css",
				FaviconURL:  "/static/favicon.svg",
			},
			want: []string{
				`<meta charset="utf-8">`,
				"<title>Hello World</title>",
				`<link rel="stylesheet" href="/static/style.css">`,
				`<link rel="icon" href="/static/favicon.svg">`,
				"<h1>Hello World</h1>",
				"<p>Body text.</p>",
			},
		},
		{
			name: "PageWithSiteTitle",
			values: StaticPageValues{
				Title:     "Hello World",
				SiteTitle: "My Wiki",
			},
			want: []string{
				"<title>Hello World — My Wiki</title>",
			},
		},
		{
			name: "PageWithoutSiteTitle",
			values: StaticPageValues{
				Title: "Hello World",
			},
			want: []string{
				"<title>Hello World</title>",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			if err := StaticPageTmpl.Execute(&buf, tt.values); err != nil {
				t.Fatalf("StaticPageTmpl.Execute: %v", err)
			}
			got := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("StaticPageTmpl.Execute output missing %q", want)
				}
			}
		})
	}
}
