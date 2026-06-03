package rivian

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/x10send/rivflux/pkg/types"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_auth.json", true)
	if client == nil {
		t.Error("NewClient returned nil")
	}
	if client.authFile != "test_auth.json" {
		t.Errorf("Expected authFile to be 'test_auth.json', got '%s'", client.authFile)
	}
	if !client.debug {
		t.Error("Expected debug to be true")
	}
}

func TestGetAuthData(t *testing.T) {
	// Create a temporary auth file for testing
	authData := types.AuthData{
		Token:           "test-token",
		RefreshToken:    "test-refresh",
		UserSessionToken: "test-session",
		CSRFToken:       "test-csrf",
		AppSessionToken: "test-app-session",
		VehicleID:       "test-vehicle",
	}

	// Write test data to a temporary file
	tmpFile, err := os.CreateTemp("", "auth_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Convert to JSON and base64 encode
	jsonData, err := json.Marshal(authData)
	if err != nil {
		t.Fatalf("Failed to marshal auth data: %v", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	if err := os.WriteFile(tmpFile.Name(), []byte(encodedData), 0600); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	client := NewClient(tmpFile.Name(), true)
	
	// Test with valid data
	got, err := client.GetAuthData()
	if err != nil {
		t.Fatalf("GetAuthData failed: %v", err)
	}

	if got.Token != authData.Token {
		t.Errorf("Token = %v, want %v", got.Token, authData.Token)
	}
	if got.VehicleID != authData.VehicleID {
		t.Errorf("VehicleID = %v, want %v", got.VehicleID, authData.VehicleID)
	}
}

func TestGetCSRFToken(t *testing.T) {
	client := NewClient("test_auth.json", true)
	
	// Test without auth data
	_, err := client.GetCSRFToken()
	if err == nil {
		t.Error("Expected error when getting CSRF token without auth data")
	}

	// TODO: Implement test with mock HTTP server
}

func TestGetVehicleState(t *testing.T) {
	client := NewClient("test_auth.json", true)
	
	// Test without auth data
	_, err := client.GetVehicleState()
	if err == nil {
		t.Error("Expected error when getting vehicle state without auth data")
	}

	// TODO: Implement test with mock HTTP server
} 