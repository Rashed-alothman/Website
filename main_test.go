package main

// Tests must run from the project root (where templates/ and static/ live)
// so that newServer() can find the template files.
// Run with: go test ./...

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestServer creates a server for use in tests.
// It fails the test immediately if template parsing fails.
func newTestServer(t *testing.T) *server {
	t.Helper()
	srv, err := newServer()
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	return srv
}

// ── Home handler tests ────────────────────────────────────────────────────────

func TestHome_ReturnsOK(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.home(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html content-type, got %q", ct)
	}
}

func TestHome_RendersHeroName(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.home(w, req)

	if !strings.Contains(w.Body.String(), srv.data.Hero.Name) {
		t.Errorf("body does not contain hero name %q", srv.data.Hero.Name)
	}
}

func TestHome_ReturnsNotFoundForUnknownPaths(t *testing.T) {
	srv := newTestServer(t)

	paths := []string{"/about", "/unknown", "/contact/extra"}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.home(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("path %q: expected 404, got %d", path, w.Code)
		}
	}
}

// ── Contact handler tests ─────────────────────────────────────────────────────

func TestContact_GetShowsForm(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()

	srv.contact(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<form") {
		t.Error("expected <form> element in body")
	}
}

func TestContact_PostRedirectsToPRG(t *testing.T) {
	srv := newTestServer(t)
	form := url.Values{
		"name":    {"Rashed"},
		"email":   {"r@example.com"},
		"message": {"Hello from the test suite"},
	}
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.contact(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303 SeeOther, got %d", w.Code)
	}
	if got := w.Header().Get("Location"); got != "/contact?sent=1" {
		t.Errorf("expected redirect to /contact?sent=1, got %q", got)
	}
}

func TestContact_SentQueryShowsThankYou(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/contact?sent=1", nil)
	w := httptest.NewRecorder()

	srv.contact(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "received") {
		t.Error("expected thank-you message in body")
	}
}

// ── jsJSON helper tests ───────────────────────────────────────────────────────

func TestJsJSON_StringSlice(t *testing.T) {
	got := string(jsJSON([]string{"Go", "Python"}))
	if got != `["Go","Python"]` {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestJsJSON_EmptySlice(t *testing.T) {
	got := string(jsJSON([]string{}))
	if got != `[]` {
		t.Errorf("expected [], got %s", got)
	}
}
