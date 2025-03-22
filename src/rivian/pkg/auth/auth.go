package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rivflux/rivian/pkg/httpclient"
	"github.com/rivflux/rivian/pkg/types"
)

// Authenticator handles Rivian authentication
type Authenticator struct {
	client *httpclient.Client
	debug  bool
}

// NewAuthenticator creates a new Authenticator instance
func NewAuthenticator(debug bool) *Authenticator {
	return &Authenticator{
		client: httpclient.NewClient(types.RivianGatewayPath, debug),
		debug:  debug,
	}
}

// GetCSRFToken gets a CSRF token from the Rivian API
func (a *Authenticator) GetCSRFToken() (*types.CSRFResponse, error) {
	return a.client.GetCSRFToken()
}

// InitialLogin performs the initial login attempt
func (a *Authenticator) InitialLogin(username, password, outputFile string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Step 1: Get CSRF token
	csrfResp, err := a.GetCSRFToken()
	if err != nil {
		return err
	}

	// Step 2: Login
	loginQuery := `{
		"operationName": "Login",
		"variables": {
			"email": "%s",
			"password": "%s"
		},
		"query": "mutation Login($email: String!, $password: String!) { login(email: $email, password: $password) { __typename ... on MobileLoginResponse { accessToken refreshToken userSessionToken } ... on MobileMFALoginResponse { otpToken } } }"
	}`

	headers := map[string]string{
		"Csrf-Token": csrfResp.Data.CreateCsrfToken.CSRFToken,
		"A-Sess":     csrfResp.Data.CreateCsrfToken.AppSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "/graphql", fmt.Sprintf(loginQuery, username, password), headers, &authResp); err != nil {
		return err
	}

	// Check if MFA is required
	if authResp.Data.Login.TypeName == "MobileMFALoginResponse" {
		// Store MFA data for later use
		mfaData := types.MFAData{
			Username:        username,
			CSRFToken:      csrfResp.Data.CreateCsrfToken.CSRFToken,
			AppSessionToken: csrfResp.Data.CreateCsrfToken.AppSessionToken,
			OTPToken:       authResp.Data.Login.OTPToken,
			Timestamp:      time.Now().Unix(),
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(mfaData)
		if err != nil {
			return fmt.Errorf("error marshaling MFA data: %v", err)
		}

		// Write to temporary file
		tempFile := outputFile + ".mfa"
		if err := ioutil.WriteFile(tempFile, jsonData, 0600); err != nil {
			return fmt.Errorf("error writing MFA file: %v", err)
		}

		return nil
	}

	// If no MFA required, proceed with getting vehicles and saving auth data
	return a.CompleteAuth(
		authResp.Data.Login.AccessToken,
		authResp.Data.Login.RefreshToken,
		authResp.Data.Login.UserSessionToken,
		csrfResp.Data.CreateCsrfToken.CSRFToken,
		csrfResp.Data.CreateCsrfToken.AppSessionToken,
		outputFile,
	)
}

// CompleteAuth completes the authentication process
func (a *Authenticator) CompleteAuth(accessToken, refreshToken, userSessionToken, csrfToken, appSessionToken, outputFile string) error {
	// Get vehicle ID
	vehiclesQuery := `{
		"operationName": "getUserInfo",
		"query": "query getUserInfo { currentUser { vehicles { id } } }",
		"variables": null
	}`

	headers := map[string]string{
		"Csrf-Token": csrfToken,
		"A-Sess":     appSessionToken,
		"U-Sess":     userSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var vehiclesResp types.VehiclesResponse
	if err := a.client.DoRequest("POST", "/graphql", vehiclesQuery, headers, &vehiclesResp); err != nil {
		return err
	}

	if len(vehiclesResp.Data.CurrentUser.Vehicles) == 0 {
		return fmt.Errorf("no vehicles found")
	}

	// Save auth data
	authData := types.AuthData{
		Token:           accessToken,
		RefreshToken:    refreshToken,
		UserSessionToken: userSessionToken,
		CSRFToken:       csrfToken,
		AppSessionToken: appSessionToken,
		VehicleID:       vehiclesResp.Data.CurrentUser.Vehicles[0].ID,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("error marshaling auth data: %v", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	if err := ioutil.WriteFile(outputFile, []byte(encodedData), 0600); err != nil {
		return fmt.Errorf("error writing auth file: %v", err)
	}

	return nil
}

// CompleteMFA completes the MFA login process
func (a *Authenticator) CompleteMFA(username, password, otpCode, outputFile string) error {
	// Read MFA data from temporary file
	tempFile := outputFile + ".mfa"
	mfaDataBytes, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("error reading MFA file: %v", err)
	}

	var mfaData types.MFAData
	if err := json.Unmarshal(mfaDataBytes, &mfaData); err != nil {
		return fmt.Errorf("error decoding MFA data: %v", err)
	}

	// Check if MFA data is expired (1 hour)
	if time.Now().Unix()-mfaData.Timestamp > 3600 {
		return fmt.Errorf("MFA data has expired. Please run the initial login again")
	}

	// Step 1: Login with OTP
	otpQuery := `{
		"operationName": "LoginWithOTP",
		"variables": {
			"email": "%s",
			"otpCode": "%s",
			"otpToken": "%s"
		},
		"query": "mutation LoginWithOTP($email: String!, $otpCode: String!, $otpToken: String!) { loginWithOTP(email: $email, otpCode: $otpCode, otpToken: $otpToken) { __typename accessToken refreshToken userSessionToken } }"
	}`

	headers := map[string]string{
		"Csrf-Token": mfaData.CSRFToken,
		"A-Sess":     mfaData.AppSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "/graphql", fmt.Sprintf(otpQuery, username, otpCode, mfaData.OTPToken), headers, &authResp); err != nil {
		return err
	}

	// Clean up temporary MFA file
	os.Remove(tempFile)

	// Complete the authentication process
	return a.CompleteAuth(
		authResp.Data.LoginWithOTP.AccessToken,
		authResp.Data.LoginWithOTP.RefreshToken,
		authResp.Data.LoginWithOTP.UserSessionToken,
		mfaData.CSRFToken,
		mfaData.AppSessionToken,
		outputFile,
	)
} 