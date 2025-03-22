package main

import (
	"flag"
	"log"
)

func main() {
	// Parse command line flags
	username := flag.String("username", "", "Rivian account username")
	password := flag.String("password", "", "Rivian account password")
	otpCode := flag.String("otp", "", "OTP code for MFA")
	outputFile := flag.String("output", "auth.json", "Output file path")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	// Validate required flags
	if *username == "" || *password == "" {
		log.Fatal("username and password are required")
	}

	// Run authentication
	if err := runAuth(*username, *password, *otpCode, *outputFile, *debug); err != nil {
		log.Fatal(err)
	}
} 