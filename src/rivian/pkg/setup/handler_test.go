package setup

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeAuth implements Authenticator for tests without hitting the network.
type fakeAuth struct {
	loginErr    error
	mfaErr      error
	mfaRequired bool // when true, InitialLogin returns mfaRequired=true
}

func (f *fakeAuth) InitialLogin(username, password, outputFile string) (bool, error) {
	if f.loginErr != nil {
		return false, f.loginErr
	}
	if f.mfaRequired {
		return true, nil
	}
	return false, os.WriteFile(outputFile, []byte("dGVzdA=="), 0600)
}

func (f *fakeAuth) CompleteMFA(username, otpCode, outputFile string) error {
	if f.mfaErr != nil {
		return f.mfaErr
	}
	return os.WriteFile(outputFile, []byte("dGVzdA=="), 0600)
}

func newTestHandler(t *testing.T, a Authenticator) (*Handler, string) {
	t.Helper()
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.json")
	h := newHandlerWithAuth(authFile, nil, func() Authenticator { return a })
	return h, authFile
}

func serve(t *testing.T, h *Handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	return httptest.NewServer(mux)
}

func TestIndex_UnconfiguredState(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{})
	srv := serve(t, h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}

func TestIndex_ConfiguredState(t *testing.T) {
	h, authFile := newTestHandler(t, &fakeAuth{})
	os.WriteFile(authFile, []byte("dGVzdA=="), 0600) //nolint:errcheck
	srv := serve(t, h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}

func TestLogin_EmptyFields(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{})
	srv := serve(t, h)
	defer srv.Close()

	resp, err := http.PostForm(srv.URL+"/setup/login", url.Values{
		"username": {""},
		"password": {""},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should render error page, not redirect.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 (error page), got %d", resp.StatusCode)
	}
}

func TestLogin_AuthFailure(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{loginErr: errTest("bad credentials")})
	srv := serve(t, h)
	defer srv.Close()

	resp, err := http.PostForm(srv.URL+"/setup/login", url.Values{
		"username": {"user@example.com"},
		"password": {"wrong"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 (error page), got %d", resp.StatusCode)
	}
}

func TestLogin_SuccessNoMFA(t *testing.T) {
	onAuthCalled := make(chan struct{}, 1)
	h, authFile := newTestHandler(t, &fakeAuth{})
	h.onAuth = func() { onAuthCalled <- struct{}{} }
	srv := serve(t, h)
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.PostForm(srv.URL+"/setup/login", url.Values{
		"username": {"user@example.com"},
		"password": {"secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("want redirect, got %d", resp.StatusCode)
	}
	if !fileExists(authFile) {
		t.Fatal("auth file should exist after successful login")
	}
}

func TestLogin_MFARequired(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{mfaRequired: true})
	srv := serve(t, h)
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.PostForm(srv.URL+"/setup/login", url.Values{
		"username": {"user@example.com"},
		"password": {"secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("want redirect to MFA form, got %d", resp.StatusCode)
	}

	h.mu.Lock()
	pending := h.pendingUsername
	h.mu.Unlock()
	if pending == "" {
		t.Fatal("pendingUsername should be set after MFA-required login")
	}
}

func TestMFA_EmptyOTP(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{})
	h.mu.Lock()
	h.pendingUsername = "user@example.com"
	h.mu.Unlock()
	srv := serve(t, h)
	defer srv.Close()

	resp, err := http.PostForm(srv.URL+"/setup/mfa", url.Values{
		"username": {"user@example.com"},
		"otp":      {""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 (error page), got %d", resp.StatusCode)
	}
}

func TestMFA_Success(t *testing.T) {
	onAuthCalled := make(chan struct{}, 1)
	h, _ := newTestHandler(t, &fakeAuth{})
	h.onAuth = func() { onAuthCalled <- struct{}{} }
	h.mu.Lock()
	h.pendingUsername = "user@example.com"
	h.mu.Unlock()
	srv := serve(t, h)
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.PostForm(srv.URL+"/setup/mfa", url.Values{
		"username": {"user@example.com"},
		"otp":      {"123456"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("want redirect, got %d", resp.StatusCode)
	}

	h.mu.Lock()
	pending := h.pendingUsername
	h.mu.Unlock()
	if pending != "" {
		t.Fatal("pendingUsername should be cleared after successful MFA")
	}
}

func TestGetMethod_Redirects(t *testing.T) {
	h, _ := newTestHandler(t, &fakeAuth{})
	srv := serve(t, h)
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	for _, path := range []string{"/setup/login", "/setup/mfa"} {
		resp, err := client.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("%s: want redirect on GET, got %d", path, resp.StatusCode)
		}
	}
}

// errTest is a simple error type for tests.
type errTest string

func (e errTest) Error() string { return string(e) }

// Ensure errTest satisfies the error interface — kept here to catch regressions.
var _ error = errTest("")

// Ensure the HTML template contains expected state markers.
func TestPageTemplate(t *testing.T) {
	cases := []struct {
		state string
		want  string
	}{
		{"unconfigured", "Authenticate"},
		{"configured", "Re-authenticate"},
		{"mfa", "one-time-code"},
	}
	for _, c := range cases {
		var buf strings.Builder
		if err := pageTmpl.Execute(&buf, map[string]string{"State": c.state}); err != nil {
			t.Fatalf("state %s: template error: %v", c.state, err)
		}
		if !strings.Contains(buf.String(), c.want) {
			t.Errorf("state %s: want %q in output", c.state, c.want)
		}
	}
}
