package ukuleleweb_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/gnoack/ukuleleweb"
	"github.com/peterbourgon/diskv/v3"
)

func TestWebDriver(t *testing.T) {
	h := ukuleleweb.NewServer(&ukuleleweb.Config{
		Store: diskv.New(diskv.Options{
			BasePath:     t.TempDir(),
			CacheSizeMax: 1024 * 1024, // 1MB
		}),
	})

	ts := httptest.NewServer(h)
	defer ts.Close()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Headless,
	}
	ctx, cancel := chromedp.NewExecAllocator(t.Context(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(t.Logf))
	defer cancel()

	var title string
	mustRun(ctx, t,
		chromedp.Navigate(ts.URL+"/UkuleleWeb"),
		chromedp.Title(&title),
	)

	if want := "Ukulele Web"; title != want {
		t.Fatalf("/UkuleleWeb title = %q, want %q", title, want)
	}

	// Click "Edit page"
	mustRun(ctx, t,
		chromedp.Click(`//a[text()="Edit page"]`, chromedp.NodeVisible),
	)

	// On the edit page, find the form and submit button
	mustRun(ctx, t,
		chromedp.WaitVisible(`//textarea[@name="content"]`),
		chromedp.WaitVisible(`//button[@type="submit"]`),
	)

	// Add markdown and submit
	const markdown = "## My Test Header"
	mustRun(ctx, t,
		chromedp.SendKeys(`//textarea[@name="content"]`, markdown),
		chromedp.Click(`//button[@type="submit"]`),
	)

	// After reload, we should be back on the UkuleleWeb page
	// and see the rendered markdown.
	var mainHTML string
	mustRun(ctx, t,
		chromedp.WaitVisible(`//a[text()="Edit page"]`),
		chromedp.InnerHTML(`//main`, &mainHTML, chromedp.NodeVisible),
	)

	if !strings.Contains(mainHTML, `<h2>My Test Header</h2>`) {
		t.Errorf("Did not find rendered markdown in <main> tag. Got:\n%s", mainHTML)
	}
}

func mustRun(ctx context.Context, t *testing.T, actions ...chromedp.Action) {
	t.Helper()
	if err := chromedp.Run(ctx, actions...); err != nil {
		t.Fatalf("chromedp.Run: %v", err)
	}
}
