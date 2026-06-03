package main

import (
	"fmt"
	"log"

	"github.com/x10send/rivflux/pkg/auth"
)

func runAuth(username, password, otpCode, outputFile string, debug bool) error {
	authenticator := auth.NewAuthenticator(debug)

	if otpCode != "" {
		if err := authenticator.CompleteMFA(username, otpCode, outputFile); err != nil {
			return fmt.Errorf("MFA completion failed: %v", err)
		}
		log.Println("MFA authentication successful")
		return nil
	}

	mfaRequired, err := authenticator.InitialLogin(username, password, outputFile)
	if err != nil {
		return fmt.Errorf("initial login failed: %v", err)
	}
	if mfaRequired {
		log.Println("MFA required — check your email and run auth again with -otp flag")
		return nil
	}
	log.Println("Authentication successful")
	return nil
}
