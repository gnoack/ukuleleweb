package ukuleleweb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/peterbourgon/diskv/v3"
)

// noRedirectClient is an HTTP client that does not follow redirects.
var noRedirectClient = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := NewServer(&Config{
		Store: diskv.New(diskv.Options{
			BasePath:     t.TempDir(),
			CacheSizeMax: 1024 * 1024,
		}),
	})
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)
	return ts
}

func TestInvalidPageName(t *testing.T) {
	ts := testServer(t)

	for _, tt := range []struct {
		name, method, path string
	}{
		{"ViewInvalidPage", "GET", "/notapage"},
		{"EditInvalidPage", "GET", "/edit/notapage"},
		{"SaveInvalidPage", "POST", "/notapage"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("http.NewRequest(%q, %q): %v", tt.method, tt.path, err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("http.DefaultClient.Do(%s %s): %v", tt.method, tt.path, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusNotFound)
			}
		})
	}
}

func TestRootRedirectsToMainPage(t *testing.T) {
	ts := testServer(t)

	resp, err := noRedirectClient.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusMovedPermanently)
	}
	if got := resp.Header.Get("Location"); got != "/MainPage" {
		t.Errorf("Location = %q, want %q", got, "/MainPage")
	}
}

func TestPreview(t *testing.T) {
	ts := testServer(t)

	resp, err := http.Post(ts.URL+"/preview", "text/plain", strings.NewReader("Hello *World*!"))
	if err != nil {
		t.Fatalf("POST /preview: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	if got, want := string(body), "<p>Hello <em>World</em>!</p>\n"; got != want {
		t.Errorf("Body = %q, want %q", got, want)
	}
}

func TestSavePage(t *testing.T) {
	ts := testServer(t)

	resp, err := noRedirectClient.Post(ts.URL+"/TestPage", "application/x-www-form-urlencoded", strings.NewReader("content=hello"))
	if err != nil {
		t.Fatalf("POST /TestPage: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusFound)
	}
	if got := resp.Header.Get("Location"); got != "/TestPage" {
		t.Errorf("Location = %q, want %q", got, "/TestPage")
	}

	// Verify the saved content is served back.
	resp, err = http.Get(ts.URL + "/TestPage")
	if err != nil {
		t.Fatalf("GET /TestPage: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	if !strings.Contains(string(body), "hello") {
		t.Errorf("GET /TestPage body does not contain saved content %q", "hello")
	}
}

func TestIsPageName(t *testing.T) {
	for _, pn := range []string{
		"PageName",
		"AlsoPageName",
		"WönderfülPägeNäme",
		// XXX: It's unclear to me why this is not recognized. Unicode shenanigans?
		// "ÄtschiBätschi",
	} {
		if !isPageName(pn) {
			t.Errorf("isPageName(%q) = false, want true", pn)
		}
	}
}

func TestIsNotPageName(t *testing.T) {
	for _, pn := range []string{
		"foo PageName bar",
		"/AlsoPageName/",
		"Oneword",
		"123",
	} {
		if isPageName(pn) {
			t.Errorf("isPageName(%q) = true, want false", pn)
		}
	}
}
