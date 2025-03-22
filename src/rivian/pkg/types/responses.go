package types

// AuthResponse represents the response from the login endpoints
type AuthResponse struct {
	Data struct {
		Login struct {
			TypeName         string `json:"__typename"`
			AccessToken      string `json:"accessToken"`
			RefreshToken     string `json:"refreshToken"`
			UserSessionToken string `json:"userSessionToken"`
			OTPToken        string `json:"otpToken"`
		} `json:"login"`
		LoginWithOTP struct {
			TypeName         string `json:"__typename"`
			AccessToken      string `json:"accessToken"`
			RefreshToken     string `json:"refreshToken"`
			UserSessionToken string `json:"userSessionToken"`
		} `json:"loginWithOTP"`
	} `json:"data"`
}

// VehicleState represents the current state of a vehicle
type VehicleState struct {
	Data struct {
		VehicleState struct {
			CabinClimateInteriorTemperature struct {
				Value float64 `json:"value"`
			} `json:"cabinClimateInteriorTemperature"`
			PowerState struct {
				Value string `json:"value"`
			} `json:"powerState"`
			DriveMode struct {
				Value string `json:"value"`
			} `json:"driveMode"`
			GearStatus struct {
				Value string `json:"value"`
			} `json:"gearStatus"`
			VehicleMileage struct {
				Value float64 `json:"value"`
			} `json:"vehicleMileage"`
			BatteryLevel struct {
				Value float64 `json:"value"`
			} `json:"batteryLevel"`
			Range struct {
				Value float64 `json:"value"`
			} `json:"range"`
			ChargerStatus struct {
				Value string `json:"value"`
			} `json:"chargerStatus"`
			ChargeState struct {
				Value string `json:"value"`
			} `json:"chargeState"`
			BatteryLimit struct {
				Value float64 `json:"value"`
			} `json:"batteryLimit"`
			ChargeEndTime struct {
				Value string `json:"value"`
			} `json:"chargeEndTime"`
			ChargeElapsedTime struct {
				Value float64 `json:"value"`
			} `json:"chargeElapsedTime"`
			ChargePower struct {
				Value float64 `json:"value"`
			} `json:"chargePower"`
			ChargeSession struct {
				Value float64 `json:"value"`
			} `json:"chargeSession"`
			BatteryCapacity struct {
				Value float64 `json:"value"`
			} `json:"batteryCapacity"`
			BatteryEnergyRemaining struct {
				Value float64 `json:"value"`
			} `json:"batteryEnergyRemaining"`
		} `json:"vehicleState"`
	} `json:"data"`
} 