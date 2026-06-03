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

const (
	mfaFileSuffix          = ".mfa"
	mobileMFALoginResponse = "MobileMFALoginResponse"
	mfaSessionTTL          = time.Hour
)

type Authenticator struct {
	client *httpclient.Client
}

func NewAuthenticator(debug bool) *Authenticator {
	return &Authenticator{
		client: httpclient.NewClient(types.RivianGatewayPath, debug),
	}
}

func (a *Authenticator) GetCSRFToken() (*types.CSRFResponse, error) {
	return a.client.GetCSRFToken()
}

// InitialLogin starts the Rivian login flow. Returns mfaRequired=true if an OTP
// was sent to the user's email; in that case call CompleteMFA next.
func (a *Authenticator) InitialLogin(username, password, outputFile string) (mfaRequired bool, err error) {
	if username == "" || password == "" {
		return false, fmt.Errorf("username and password are required")
	}

	csrfResp, err := a.GetCSRFToken()
	if err != nil {
		return false, err
	}

	body, err := json.Marshal(map[string]any{
		"operationName": "Login",
		"variables":     map[string]string{"email": username, "password": password},
		"query":         "mutation Login($email: String!, $password: String!) { login(email: $email, password: $password) { __typename ... on MobileLoginResponse { accessToken refreshToken userSessionToken } ... on MobileMFALoginResponse { otpToken } } }",
	})
	if err != nil {
		return false, fmt.Errorf("error building login request: %v", err)
	}

	headers := map[string]string{
		"Csrf-Token": csrfResp.Data.CreateCsrfToken.CSRFToken,
		"A-Sess":     csrfResp.Data.CreateCsrfToken.AppSessionToken,
		"Dc-Cid":     newDeviceID(),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "", string(body), headers, &authResp); err != nil {
		return false, fmt.Errorf("initial login failed: %v", err)
	}

	if authResp.Data.Login.TypeName == mobileMFALoginResponse {
		mfaData := types.MFAData{
			Username:        username,
			CSRFToken:       csrfResp.Data.CreateCsrfToken.CSRFToken,
			AppSessionToken: csrfResp.Data.CreateCsrfToken.AppSessionToken,
			OTPToken:        authResp.Data.Login.OTPToken,
			Timestamp:       time.Now().Unix(),
		}
		jsonData, err := json.Marshal(mfaData)
		if err != nil {
			return false, fmt.Errorf("error marshaling MFA data: %v", err)
		}
		if err := os.WriteFile(outputFile+mfaFileSuffix, jsonData, 0600); err != nil {
			return false, fmt.Errorf("error writing MFA session file: %v", err)
		}
		return true, nil
	}

	return false, a.CompleteAuth(
		authResp.Data.Login.AccessToken,
		authResp.Data.Login.RefreshToken,
		authResp.Data.Login.UserSessionToken,
		csrfResp.Data.CreateCsrfToken.CSRFToken,
		csrfResp.Data.CreateCsrfToken.AppSessionToken,
		outputFile,
	)
}

func (a *Authenticator) CompleteAuth(accessToken, refreshToken, userSessionToken, csrfToken, appSessionToken, outputFile string) error {
	vehiclesQuery, err := json.Marshal(map[string]any{
		"operationName": "getUserInfo",
		"query":         "query getUserInfo { currentUser { vehicles { id } } }",
		"variables":     nil,
	})
	if err != nil {
		return fmt.Errorf("error building vehicles request: %v", err)
	}

	headers := map[string]string{
		"Csrf-Token": csrfToken,
		"A-Sess":     appSessionToken,
		"U-Sess":     userSessionToken,
		"Dc-Cid":     newDeviceID(),
	}

	var vehiclesResp types.VehiclesResponse
	if err := a.client.DoRequest("POST", "", string(vehiclesQuery), headers, &vehiclesResp); err != nil {
		return fmt.Errorf("failed to get vehicles: %v", err)
	}

	if len(vehiclesResp.Data.CurrentUser.Vehicles) == 0 {
		return fmt.Errorf("no vehicles found for this account")
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
		return fmt.Errorf("error writing auth file %s: %v", outputFile, err)
	}
	return nil
}

// CompleteMFA finishes an MFA login using the OTP code sent to the user's email.
// It reads the interim session file written by InitialLogin and removes it on completion.
func (a *Authenticator) CompleteMFA(username, otpCode, outputFile string) error {
	tempFile := outputFile + mfaFileSuffix
	mfaDataBytes, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("error reading MFA session file: %v", err)
	}
	defer os.Remove(tempFile) // always clean up, whether we succeed or fail

	var mfaData types.MFAData
	if err := json.Unmarshal(mfaDataBytes, &mfaData); err != nil {
		return fmt.Errorf("error decoding MFA session: %v", err)
	}

	if time.Since(time.Unix(mfaData.Timestamp, 0)) > mfaSessionTTL {
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
		"Dc-Cid":     newDeviceID(),
	}

	var authResp types.AuthResponse
	if err := a.client.DoRequest("POST", "", string(body), headers, &authResp); err != nil {
		return fmt.Errorf("MFA verification failed: %v", err)
	}

	return a.CompleteAuth(
		authResp.Data.LoginWithOTP.AccessToken,
		authResp.Data.LoginWithOTP.RefreshToken,
		authResp.Data.LoginWithOTP.UserSessionToken,
		mfaData.CSRFToken,
		mfaData.AppSessionToken,
		outputFile,
	)
}

func newDeviceID() string {
	return "m-ios-" + uuid.New().String()
}
