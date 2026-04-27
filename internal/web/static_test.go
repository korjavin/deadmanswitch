package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// projectRoot walks up from this test file to find the repo root
// (the directory that contains web/static). We do this so the tests
// don't depend on the working directory the `go test` invocation chose.
func projectRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate test file")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "web", "static")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate project root containing web/static (started from %s)", thisFile)
	return ""
}

// staticFileServer constructs an http.Handler equivalent to the one
// the live server mounts at /static/, but rooted at the test-discovered
// web/static directory so it works regardless of $PWD.
func staticFileServer(t *testing.T) http.Handler {
	t.Helper()
	root := projectRoot(t)
	return http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(root, "web", "static"))))
}

// TestStaticServesHeartbeatCSS asserts the redesign foundation CSS
// files are present on disk and served with the correct content type.
func TestStaticServesHeartbeatCSS(t *testing.T) {
	srv := staticFileServer(t)
	cases := []struct {
		path     string
		wantType string
	}{
		{"/static/css/tokens.css", "text/css"},
		{"/static/css/heartbeat.css", "text/css"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("got status %d, want 200", rr.Code)
			}
			ct := rr.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, tc.wantType) {
				t.Errorf("got content-type %q, want prefix %q", ct, tc.wantType)
			}
			if rr.Body.Len() == 0 {
				t.Error("empty body")
			}
		})
	}
}

// TestStaticServesJetBrainsMonoFonts asserts the self-hosted woff2
// fonts are present and served as font/woff2 (Go's mime package
// recognizes the .woff2 extension).
func TestStaticServesJetBrainsMonoFonts(t *testing.T) {
	srv := staticFileServer(t)
	for _, p := range []string{
		"/static/fonts/JetBrainsMono-400.woff2",
		"/static/fonts/JetBrainsMono-400italic.woff2",
	} {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("got status %d, want 200", rr.Code)
			}
			ct := rr.Header().Get("Content-Type")
			// Go's net/http file server determines content-type via mime.
			// font/woff2 is the registered IANA type; some Go versions
			// report application/octet-stream if mime DB is bare. Accept
			// either, but assert the body looks like a woff2 (magic 'wOF2').
			if rr.Body.Len() < 4 {
				t.Fatalf("font body too short: %d bytes", rr.Body.Len())
			}
			head := rr.Body.Bytes()[:4]
			if string(head) != "wOF2" {
				t.Errorf("font header = %q, want wOF2 (content-type was %q)", head, ct)
			}
		})
	}
}

// TestStaticServesDesignAssets asserts the dotgrid background and
// blueprint plate SVGs are present.
func TestStaticServesDesignAssets(t *testing.T) {
	srv := staticFileServer(t)
	for _, p := range []string{
		"/static/assets/dotgrid-bg.svg",
		"/static/assets/blueprint-plate-01.svg",
	} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("%s: got %d, want 200", p, rr.Code)
		}
		if !strings.Contains(strings.ToLower(rr.Body.String()), "<svg") {
			t.Errorf("%s: body does not look like SVG", p)
		}
	}
}

// TestLayoutLoadsHeartbeatStylesheets is a smoke test that reads
// layout.html from disk and asserts the redesign CSS links are
// present in the <head> while the old stylesheet links are gone.
// This is the cheapest possible "did Task 1 land" check; later
// tasks will mount full template-render assertions.
func TestLayoutLoadsHeartbeatStylesheets(t *testing.T) {
	root := projectRoot(t)
	body, err := os.ReadFile(filepath.Join(root, "web", "templates", "layout.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)

	mustContain := []string{
		`/static/css/tokens.css`,
		`/static/css/heartbeat.css`,
	}
	mustNotContain := []string{
		`/static/css/normalize.css`,
		`/static/css/main.css`,
	}
	for _, s := range mustContain {
		if !strings.Contains(html, s) {
			t.Errorf("layout.html missing reference to %s", s)
		}
	}
	for _, s := range mustNotContain {
		if strings.Contains(html, s) {
			t.Errorf("layout.html still references old stylesheet %s", s)
		}
	}
}
