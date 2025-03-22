package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAuth(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name       string
		username   string
		password   string
		otpCode    string
		outputFile string
		debug      bool
		wantErr    bool
	}{
		{
			name:       "missing credentials",
			username:   "",
			password:   "",
			otpCode:    "",
			outputFile: filepath.Join(tmpDir, "auth1.json"),
			debug:      true,
			wantErr:    true,
		},
		{
			name:       "invalid credentials",
			username:   "test@example.com",
			password:   "wrongpassword",
			otpCode:    "",
			outputFile: filepath.Join(tmpDir, "auth2.json"),
			debug:      true,
			wantErr:    true,
		},
		{
			name:       "invalid otp",
			username:   "test@example.com",
			password:   "testpass",
			otpCode:    "123456",
			outputFile: filepath.Join(tmpDir, "auth3.json"),
			debug:      true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runAuth(tt.username, tt.password, tt.otpCode, tt.outputFile, tt.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("runAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} 