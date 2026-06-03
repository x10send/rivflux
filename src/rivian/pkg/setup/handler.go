package setup

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/x10send/rivflux/pkg/auth"
)

// Authenticator is the subset of auth.Authenticator used by the setup handler.
// Defined as an interface so tests can inject a fake without hitting the network.
type Authenticator interface {
	InitialLogin(username, password, outputFile string) (mfaRequired bool, err error)
	CompleteMFA(username, otpCode, outputFile string) error
}

const (
	maxEmailLen = 254
	maxOTPLen   = 8
)

var pageTmpl = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Rivflux Setup</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:system-ui,sans-serif;background:#0d1117;color:#e6edf3;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#161b22;border:1px solid #30363d;border-radius:12px;padding:2rem;width:100%;max-width:420px}
h1{font-size:1.25rem;margin-bottom:.25rem}
.sub{color:#8b949e;font-size:.875rem;margin-bottom:1.5rem}
label{display:block;font-size:.875rem;color:#8b949e;margin-bottom:.25rem;margin-top:1rem}
input{width:100%;padding:.625rem .75rem;background:#0d1117;border:1px solid #30363d;border-radius:6px;color:#e6edf3;font-size:.875rem;outline:none}
input:focus{border-color:#388bfd}
button{margin-top:1.5rem;width:100%;padding:.75rem;background:#238636;border:none;border-radius:6px;color:#fff;font-size:.875rem;font-weight:600;cursor:pointer}
button:hover{background:#2ea043}
.status{margin-top:1rem;padding:.75rem;border-radius:6px;font-size:.875rem}
.ok{background:#0d2818;border:1px solid #238636;color:#3fb950}
.err{background:#2d1012;border:1px solid #f85149;color:#f85149}
.info{background:#0d1b2a;border:1px solid #388bfd;color:#79c0ff}
.badge{display:inline-block;padding:.2rem .5rem;border-radius:4px;font-size:.75rem;font-weight:600;margin-left:.4rem}
.running{background:#0d2818;color:#3fb950}
.waiting{background:#1a1a0d;color:#e3b341}
</style>
</head>
<body>
<div class="card">
  <h1>Rivflux</h1>
  <p class="sub">Rivian &rarr; InfluxDB data collector</p>

  {{if .Error}}
  <div class="status err">{{.Error}}</div>
  {{end}}

  {{if eq .State "configured"}}
  <div class="status ok">Authenticated &amp; collecting data<span class="badge running">running</span></div>
  <p style="margin-top:1.25rem;font-size:.875rem;color:#8b949e">Re-authenticate if your Rivian session has expired.</p>
  <form method="POST" action="/setup/login">
    <label>Rivian email</label>
    <input name="username" type="email" required maxlength="254" autocomplete="username">
    <label>Password</label>
    <input name="password" type="password" required autocomplete="current-password">
    <button type="submit">Re-authenticate</button>
  </form>

  {{else if eq .State "mfa"}}
  <div class="status info">Check your email for a one-time code and enter it below.</div>
  <form method="POST" action="/setup/mfa">
    <input name="username" type="hidden" value="{{.Username}}">
    <label>One-time code</label>
    <input name="otp" type="text" required maxlength="8" autocomplete="one-time-code" autofocus inputmode="numeric">
    <button type="submit">Verify</button>
  </form>

  {{else}}
  <div class="status waiting">Not yet authenticated<span class="badge waiting">waiting</span></div>
  <form method="POST" action="/setup/login">
    <label>Rivian email</label>
    <input name="username" type="email" required maxlength="254" autocomplete="username">
    <label>Password</label>
    <input name="password" type="password" required autocomplete="current-password">
    <button type="submit">Authenticate</button>
  </form>
  {{end}}
</div>
</body>
</html>`))

type Handler struct {
	authFile         string
	onAuth           func()
	newAuthenticator func() Authenticator
	mu               sync.Mutex
	pendingUsername  string
}

func NewHandler(authFile string, onAuth func()) *Handler {
	return newHandlerWithAuth(authFile, onAuth, func() Authenticator {
		return auth.NewAuthenticator(false)
	})
}

func newHandlerWithAuth(authFile string, onAuth func(), newAuth func() Authenticator) *Handler {
	return &Handler{
		authFile:         authFile,
		onAuth:           onAuth,
		newAuthenticator: newAuth,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleIndex)
	mux.HandleFunc("/setup/login", h.handleLogin)
	mux.HandleFunc("/setup/mfa", h.handleMFA)
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	pending := h.pendingUsername
	h.mu.Unlock()

	state := "unconfigured"
	switch {
	case pending != "":
		state = "mfa"
	case fileExists(h.authFile):
		state = "configured"
	}

	pageTmpl.Execute(w, map[string]string{"State": state, "Username": pending}) //nolint:errcheck
}

// handleLogin calls InitialLogin. If MFA is required it stores the pending
// username; otherwise it signals the logger to start.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		h.renderError(w, "unconfigured", "", "Email and password are required.")
		return
	}
	if len(username) > maxEmailLen {
		h.renderError(w, "unconfigured", "", "Email address is too long.")
		return
	}

	mfaRequired, err := h.newAuthenticator().InitialLogin(username, password, h.authFile)
	if err != nil {
		log.Printf("setup: InitialLogin error for %s: %v", username, err)
		h.renderError(w, "unconfigured", "", "Authentication failed.")
		return
	}

	if mfaRequired {
		h.mu.Lock()
		h.pendingUsername = username
		h.mu.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.mu.Lock()
	h.pendingUsername = ""
	h.mu.Unlock()
	if h.onAuth != nil {
		go h.onAuth()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleMFA completes the login with the OTP code from the user's email.
func (h *Handler) handleMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username := r.FormValue("username")
	otp := r.FormValue("otp")
	if otp == "" {
		h.renderError(w, "mfa", username, "One-time code is required.")
		return
	}
	if len(otp) > maxOTPLen {
		h.renderError(w, "mfa", username, "One-time code is too long.")
		return
	}

	if err := h.newAuthenticator().CompleteMFA(username, otp, h.authFile); err != nil {
		log.Printf("setup: CompleteMFA error for %s: %v", username, err)
		h.renderError(w, "mfa", username, "Verification failed.")
		return
	}

	h.mu.Lock()
	h.pendingUsername = ""
	h.mu.Unlock()
	if h.onAuth != nil {
		go h.onAuth()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) renderError(w http.ResponseWriter, state, username, errMsg string) {
	pageTmpl.Execute(w, map[string]string{ //nolint:errcheck
		"State":    state,
		"Username": username,
		"Error":    errMsg,
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
