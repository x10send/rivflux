package types

// AuthData represents the authentication data stored in the auth file
type AuthData struct {
	Token           string `json:"token"`
	RefreshToken    string `json:"refresh_token"`
	UserSessionToken string `json:"user_session_token"`
	CSRFToken       string `json:"csrf_token"`
	AppSessionToken string `json:"app_session_token"`
	VehicleID       string `json:"vehicle_id"`
}

// MFAData represents the MFA data stored temporarily
type MFAData struct {
	Username        string `json:"username"`
	CSRFToken      string `json:"csrf_token"`
	AppSessionToken string `json:"app_session_token"`
	OTPToken       string `json:"otp_token"`
	Timestamp      int64  `json:"timestamp"`
}

// CSRFResponse represents the response from the CSRF token endpoint
type CSRFResponse struct {
	Data struct {
		CreateCsrfToken struct {
			TypeName        string `json:"__typename"`
			CSRFToken      string `json:"csrfToken"`
			AppSessionToken string `json:"appSessionToken"`
		} `json:"createCsrfToken"`
	} `json:"data"`
}

// VehiclesResponse represents the response from the vehicles endpoint
type VehiclesResponse struct {
	Data struct {
		CurrentUser struct {
			Vehicles []struct {
				ID string `json:"id"`
			} `json:"vehicles"`
		} `json:"currentUser"`
	} `json:"data"`
} 