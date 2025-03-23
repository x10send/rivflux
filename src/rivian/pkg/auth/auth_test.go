package auth

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/rivflux/rivian/pkg/types"
)

func TestAuthenticator_InitialLogin(t *testing.T) {
	// Test with empty credentials
	err := NewAuthenticator(true).InitialLogin("", "", "test_auth.json")
	if err == nil {
		t.Error("Expected error with empty credentials")
	}

	// Test with invalid file path
	err = NewAuthenticator(true).InitialLogin("test@example.com", "password", "/invalid/path/auth.json")
	if err == nil {
		t.Error("Expected error with invalid file path")
	}
}

func TestAuthenticator_CompleteMFA(t *testing.T) {
	// Create a temporary MFA file
	tmpFile, err := os.CreateTemp("", "auth_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create test MFA data
	mfaData := types.MFAData{
		Username:        "test@example.com",
		CSRFToken:      "test-csrf",
		AppSessionToken: "test-app-session",
		OTPToken:       "test-otp",
		Timestamp:      time.Now().Unix(),
	}

	// Write MFA data to file
	jsonData, err := json.Marshal(mfaData)
	if err != nil {
		t.Fatalf("Failed to marshal MFA data: %v", err)
	}

	mfaFile := tmpFile.Name() + ".mfa"
	if err := os.WriteFile(mfaFile, jsonData, 0600); err != nil {
		t.Fatalf("Failed to write MFA file: %v", err)
	}
	defer os.Remove(mfaFile)

	// Test with empty OTP code
	err = NewAuthenticator(true).CompleteMFA("test@example.com", "", mfaFile, "test_auth.json")
	if err == nil {
		t.Error("Expected error with empty OTP code")
	}

	// Test with invalid MFA file
	err = NewAuthenticator(true).CompleteMFA("test@example.com", "123456", "/invalid/path/auth.mfa", "test_auth.json")
	if err == nil {
		t.Error("Expected error with invalid MFA file")
	}
} 