package main

import (
	"fmt"
	"log"

	"github.com/rivflux/rivian/pkg/auth"
)

func runAuth(username, password, otpCode, outputFile string, debug bool) error {
	authenticator := auth.NewAuthenticator(debug)

	if otpCode != "" {
		// Complete MFA login
		if err := authenticator.CompleteMFA(username, password, otpCode, outputFile); err != nil {
			return fmt.Errorf("MFA completion failed: %v", err)
		}
		log.Println("MFA authentication successful")
		return nil
	}

	// Initial login
	if err := authenticator.InitialLogin(username, password, outputFile); err != nil {
		return fmt.Errorf("Initial login failed: %v", err)
	}
	log.Println("Initial login successful")
	return nil
} 