package testutil

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/x10send/rivflux/pkg/types"
)

// WriteAuthFile creates a temp auth.json containing base64-encoded AuthData.
// The file is placed in t.TempDir() and removed automatically when the test ends.
func WriteAuthFile(t *testing.T, data types.AuthData) string {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp(t.TempDir(), "auth_*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(base64.StdEncoding.EncodeToString(raw)); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
