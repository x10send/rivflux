package auth

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/x10send/rivflux/pkg/types"
)

func TestInitialLogin_EmptyCredentials(t *testing.T) {
	err := NewAuthenticator(false).InitialLogin("", "", "auth.json")
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestInitialLogin_NetworkFailure(t *testing.T) {
	// Points at a port nothing is listening on.
	err := NewAuthenticator(false).InitialLogin("user@example.com", "pass", "auth.json")
	if err == nil {
		t.Fatal("expected error when Rivian API is unreachable")
	}
}

func TestCompleteMFA_MissingMFAFile(t *testing.T) {
	err := NewAuthenticator(false).CompleteMFA("user@example.com", "", "123456", "/nonexistent/auth.json")
	if err == nil {
		t.Fatal("expected error when .mfa file is missing")
	}
}

func TestCompleteMFA_CorruptedMFAFile(t *testing.T) {
	dir := t.TempDir()
	authFile := dir + "/auth.json"
	mfaFile := authFile + ".mfa"

	if err := os.WriteFile(mfaFile, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	err := NewAuthenticator(false).CompleteMFA("user@example.com", "", "123456", authFile)
	if err == nil {
		t.Fatal("expected error for corrupted .mfa file")
	}
}

func TestCompleteMFA_ExpiredSession(t *testing.T) {
	dir := t.TempDir()
	authFile := dir + "/auth.json"
	mfaFile := authFile + ".mfa"

	expired := types.MFAData{
		Username:        "user@example.com",
		CSRFToken:       "csrf",
		AppSessionToken: "app",
		OTPToken:        "otp",
		Timestamp:       time.Now().Unix() - 7200, // 2 hours ago
	}
	data, _ := json.Marshal(expired)
	if err := os.WriteFile(mfaFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	err := NewAuthenticator(false).CompleteMFA("user@example.com", "", "123456", authFile)
	if err == nil {
		t.Fatal("expected error for expired MFA session")
	}
}

func TestCompleteMFA_NetworkFailure(t *testing.T) {
	// Write a valid, non-expired .mfa file so we get past local validation
	// and hit the (non-existent) network — confirming the OTP request fires.
	dir := t.TempDir()
	authFile := dir + "/auth.json"
	mfaFile := authFile + ".mfa"

	valid := types.MFAData{
		Username:        "user@example.com",
		CSRFToken:       "csrf",
		AppSessionToken: "app",
		OTPToken:        "otp-token",
		Timestamp:       time.Now().Unix(),
	}
	data, _ := json.Marshal(valid)
	if err := os.WriteFile(mfaFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	err := NewAuthenticator(false).CompleteMFA("user@example.com", "", "123456", authFile)
	if err == nil {
		t.Fatal("expected network error when Rivian is unreachable")
	}
}
