package ukuleleweb

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
