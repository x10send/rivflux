package rivian

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/x10send/rivflux/pkg/httpclient"
	"github.com/x10send/rivflux/pkg/types"
)

// newTestHTTPClient returns an httpclient.Client pointed at baseURL, used to
// inject a test server into rivian.Client without touching the real Rivian API.
func newTestHTTPClient(baseURL string) *httpclient.Client {
	return httpclient.NewClient(baseURL, false)
}

func TestNewClient(t *testing.T) {
	c := NewClient("test_auth.json", true)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.authFile != "test_auth.json" {
		t.Errorf("authFile = %q, want test_auth.json", c.authFile)
	}
}

// writeAuthFile creates a temp auth file with the given AuthData and returns its path.
func writeAuthFile(t *testing.T, data types.AuthData) string {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp(t.TempDir(), "auth_*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(base64.StdEncoding.EncodeToString(raw)) //nolint:errcheck
	f.Close()
	return f.Name()
}

func TestGetAuthData_Valid(t *testing.T) {
	want := types.AuthData{
		Token:            "tok",
		RefreshToken:     "ref",
		UserSessionToken: "sess",
		CSRFToken:        "csrf",
		AppSessionToken:  "app",
		VehicleID:        "v123",
	}
	path := writeAuthFile(t, want)

	got, err := NewClient(path, false).GetAuthData()
	if err != nil {
		t.Fatalf("GetAuthData error: %v", err)
	}
	if got.Token != want.Token || got.VehicleID != want.VehicleID {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetAuthData_MissingFile(t *testing.T) {
	_, err := NewClient("/nonexistent/auth.json", false).GetAuthData()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestGetAuthData_BadBase64(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "auth_*.json")
	f.WriteString("!!!not-base64!!!") //nolint:errcheck
	f.Close()

	_, err := NewClient(f.Name(), false).GetAuthData()
	if err == nil {
		t.Fatal("expected error for bad base64")
	}
}

func TestGetAuthData_BadJSON(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "auth_*.json")
	f.WriteString(base64.StdEncoding.EncodeToString([]byte("not json"))) //nolint:errcheck
	f.Close()

	_, err := NewClient(f.Name(), false).GetAuthData()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestGetAuthData_MissingVehicleID(t *testing.T) {
	path := writeAuthFile(t, types.AuthData{Token: "tok"}) // VehicleID empty
	_, err := NewClient(path, false).GetAuthData()
	if err == nil {
		t.Fatal("expected error when VehicleID is missing")
	}
}

func TestGetCSRFToken(t *testing.T) {
	resp := map[string]any{
		"data": map[string]any{
			"createCsrfToken": map[string]string{
				"__typename":      "CreateCsrfTokenResponse",
				"csrfToken":       "csrf-xyz",
				"appSessionToken": "app-xyz",
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	// Override the client's baseURL by reaching into the unexported field via a
	// custom constructor — we expose it through NewClient then swap the inner URL.
	c := &Client{
		client: newTestHTTPClient(srv.URL),
		debug:  false,
	}
	got, err := c.GetCSRFToken()
	if err != nil {
		t.Fatalf("GetCSRFToken error: %v", err)
	}
	if got.Data.CreateCsrfToken.CSRFToken != "csrf-xyz" {
		t.Errorf("CSRFToken = %q, want csrf-xyz", got.Data.CreateCsrfToken.CSRFToken)
	}
}

func TestGetVehicleState(t *testing.T) {
	authPath := writeAuthFile(t, types.AuthData{
		Token: "bearer-tok", VehicleID: "v999",
	})

	vehicleResp := map[string]any{
		"data": map[string]any{
			"vehicleState": map[string]any{
				"batteryLevel": map[string]any{"value": 80.0},
				"range":        map[string]any{"value": 200.0},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(vehicleResp) //nolint:errcheck
	}))
	defer srv.Close()

	c := &Client{
		client:   newTestHTTPClient(srv.URL),
		authFile: authPath,
		debug:    false,
	}
	got, err := c.GetVehicleState()
	if err != nil {
		t.Fatalf("GetVehicleState error: %v", err)
	}
	if got.Data.VehicleState.BatteryLevel.Value != 80.0 {
		t.Errorf("BatteryLevel = %v, want 80.0", got.Data.VehicleState.BatteryLevel.Value)
	}
}

func TestGetVehicleState_MissingAuthFile(t *testing.T) {
	_, err := NewClient("/nonexistent/auth.json", false).GetVehicleState()
	if err == nil {
		t.Fatal("expected error when auth file is missing")
	}
}
