package httpclient

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://example.com", false)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
}

func TestDoRequest_Success(t *testing.T) {
	want := map[string]string{"key": "value"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(want) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got map[string]string
	if err := c.DoRequest("POST", "", `{}`, nil, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got %v, want key=value", got)
	}
}

func TestDoRequest_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got any
	err := c.DoRequest("POST", "", `{}`, nil, &got)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestDoRequest_GzipResponse(t *testing.T) {
	want := map[string]string{"compressed": "true"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		json.NewEncoder(gz).Encode(want) //nolint:errcheck
		gz.Close()                       //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got map[string]string
	if err := c.DoRequest("POST", "", `{}`, nil, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["compressed"] != "true" {
		t.Errorf("got %v, want compressed=true", got)
	}
}

func TestDoRequest_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json")) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got map[string]string
	if err := c.DoRequest("POST", "", `{}`, nil, &got); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestDoRequest_CustomHeaders(t *testing.T) {
	var received http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header
		json.NewEncoder(w).Encode(map[string]string{}) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got any
	headers := map[string]string{"X-Custom": "test-value"}
	if err := c.DoRequest("POST", "", `{}`, headers, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := received.Get("X-Custom"); got != "test-value" {
		t.Errorf("X-Custom header = %q, want %q", got, "test-value")
	}
}

func TestDoRequest_DefaultHeadersSet(t *testing.T) {
	var received http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header
		json.NewEncoder(w).Encode(map[string]string{}) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got any
	if err := c.DoRequest("POST", "", `{}`, nil, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ua := received.Get("User-Agent"); ua == "" {
		t.Error("User-Agent header not set")
	}
	if ct := received.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestDoRequest_BodySent(t *testing.T) {
	var body bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body.ReadFrom(r.Body) //nolint:errcheck
		json.NewEncoder(w).Encode(map[string]string{}) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	var got any
	payload := `{"test":"payload"}`
	if err := c.DoRequest("POST", "", payload, nil, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body.String() != payload {
		t.Errorf("body = %q, want %q", body.String(), payload)
	}
}

func TestGetCSRFToken(t *testing.T) {
	resp := map[string]any{
		"data": map[string]any{
			"createCsrfToken": map[string]string{
				"__typename":      "CreateCsrfTokenResponse",
				"csrfToken":       "csrf-abc",
				"appSessionToken": "app-session-xyz",
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	got, err := c.GetCSRFToken()
	if err != nil {
		t.Fatalf("GetCSRFToken error: %v", err)
	}
	if got.Data.CreateCsrfToken.CSRFToken != "csrf-abc" {
		t.Errorf("CSRFToken = %q, want csrf-abc", got.Data.CreateCsrfToken.CSRFToken)
	}
	if got.Data.CreateCsrfToken.AppSessionToken != "app-session-xyz" {
		t.Errorf("AppSessionToken = %q, want app-session-xyz", got.Data.CreateCsrfToken.AppSessionToken)
	}
}
