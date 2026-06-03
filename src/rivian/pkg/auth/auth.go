package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/x10send/rivflux/pkg/httpclient"
	"github.com/x10send/rivflux/pkg/types"
)

type Authenticator struct {
	client *httpclient.Client
	debug  bool
}

func NewAuthenticator(debug bool) *Authenticator {
	return &Authenticator{
		client: httpclient.NewClient(types.RivianGatewayPath, debug),
		debug:  debug,
	}
}

func (a *Authenticator) GetCSRFToken() (*types.CSRFResponse, error) {
	return a.client.GetCSRFToken()
}

func (a *Authenticator) InitialLogin(username, password, outputFile string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	csrfResp, err := a.GetCSRFToken()
	if err != nil {
		return err
	}

	// Use proper JSON marshaling for variables to prevent injection.
	body, err := json.Marshal(map[string]any{
		"operationName": "Login",
		"variables":     map[string]string{"email": username, "password": password},
		"query":         "mutation Login($email: String!, $password: String!) { login(email: $email, password: $password) { __typename ... on MobileLoginResponse { accessToken refreshToken userSessionToken } ... on MobileMFALoginResponse { otpToken } } }",
	})
	if err != nil {
		return fmt.Errorf("error building login request: %v", err)
	}

	headers := map[string]string{
		"Csrf-Token": csrfResp.Data.CreateCsrfToken.CSRFToken,
		"A-Sess":     csrfResp.Data.CreateCsrfToken.AppSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "", string(body), headers, &authResp); err != nil {
		return fmt.Errorf("initial login failed: %v", err)
	}

	if authResp.Data.Login.TypeName == "MobileMFALoginResponse" {
		mfaData := types.MFAData{
			Username:        username,
			CSRFToken:       csrfResp.Data.CreateCsrfToken.CSRFToken,
			AppSessionToken: csrfResp.Data.CreateCsrfToken.AppSessionToken,
			OTPToken:        authResp.Data.Login.OTPToken,
			Timestamp:       time.Now().Unix(),
		}
		jsonData, err := json.Marshal(mfaData)
		if err != nil {
			return fmt.Errorf("error marshaling MFA data: %v", err)
		}
		if err := os.WriteFile(outputFile+".mfa", jsonData, 0600); err != nil {
			return fmt.Errorf("error writing MFA file: %v", err)
		}
		return nil
	}

	return a.CompleteAuth(
		authResp.Data.Login.AccessToken,
		authResp.Data.Login.RefreshToken,
		authResp.Data.Login.UserSessionToken,
		csrfResp.Data.CreateCsrfToken.CSRFToken,
		csrfResp.Data.CreateCsrfToken.AppSessionToken,
		outputFile,
	)
}

func (a *Authenticator) CompleteAuth(accessToken, refreshToken, userSessionToken, csrfToken, appSessionToken, outputFile string) error {
	vehiclesQuery, _ := json.Marshal(map[string]any{
		"operationName": "getUserInfo",
		"query":         "query getUserInfo { currentUser { vehicles { id } } }",
		"variables":     nil,
	})

	headers := map[string]string{
		"Csrf-Token": csrfToken,
		"A-Sess":     appSessionToken,
		"U-Sess":     userSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var vehiclesResp types.VehiclesResponse
	if err := a.client.DoRequest("POST", "", string(vehiclesQuery), headers, &vehiclesResp); err != nil {
		return fmt.Errorf("failed to get vehicles: %v", err)
	}

	if len(vehiclesResp.Data.CurrentUser.Vehicles) == 0 {
		return fmt.Errorf("no vehicles found")
	}

	authData := types.AuthData{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserSessionToken: userSessionToken,
		CSRFToken:        csrfToken,
		AppSessionToken:  appSessionToken,
		VehicleID:        vehiclesResp.Data.CurrentUser.Vehicles[0].ID,
	}
	jsonData, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("error marshaling auth data: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	if err := os.WriteFile(outputFile, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("error writing auth file: %v", err)
	}
	return nil
}

// CompleteMFA completes the MFA login process using the interim .mfa temp file written
// by InitialLogin. The password parameter is unused (OTP flow doesn't re-send credentials).
func (a *Authenticator) CompleteMFA(username, _ /* password */, otpCode, outputFile string) error {
	tempFile := outputFile + ".mfa"
	mfaDataBytes, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("error reading MFA file: %v", err)
	}

	var mfaData types.MFAData
	if err := json.Unmarshal(mfaDataBytes, &mfaData); err != nil {
		return fmt.Errorf("error decoding MFA data: %v", err)
	}

	if time.Now().Unix()-mfaData.Timestamp > 3600 {
		return fmt.Errorf("MFA session expired — please log in again")
	}

	body, err := json.Marshal(map[string]any{
		"operationName": "LoginWithOTP",
		"variables": map[string]string{
			"email":    username,
			"otpCode":  otpCode,
			"otpToken": mfaData.OTPToken,
		},
		"query": "mutation LoginWithOTP($email: String!, $otpCode: String!, $otpToken: String!) { loginWithOTP(email: $email, otpCode: $otpCode, otpToken: $otpToken) { __typename accessToken refreshToken userSessionToken } }",
	})
	if err != nil {
		return fmt.Errorf("error building OTP request: %v", err)
	}

	headers := map[string]string{
		"Csrf-Token": mfaData.CSRFToken,
		"A-Sess":     mfaData.AppSessionToken,
		"Dc-Cid":     fmt.Sprintf("m-ios-%s", uuid.New().String()),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "", string(body), headers, &authResp); err != nil {
		return fmt.Errorf("mfa login failed: %v", err)
	}

	os.Remove(tempFile)

	return a.CompleteAuth(
		authResp.Data.LoginWithOTP.AccessToken,
		authResp.Data.LoginWithOTP.RefreshToken,
		authResp.Data.LoginWithOTP.UserSessionToken,
		mfaData.CSRFToken,
		mfaData.AppSessionToken,
		outputFile,
	)
}
