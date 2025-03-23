package rivian

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/rivflux/rivian/pkg/httpclient"
	"github.com/rivflux/rivian/pkg/types"
)

// Client represents a Rivian API client
type Client struct {
	client   *httpclient.Client
	authFile string
	debug    bool
}

// NewClient creates a new Rivian API client
func NewClient(authFile string, debug bool) *Client {
	return &Client{
		client:   httpclient.NewClient(types.RivianAPIPath, debug),
		authFile: authFile,
		debug:    debug,
	}
}

// GetAuthData reads and decodes the auth data from the auth file
func (c *Client) GetAuthData() (*types.AuthData, error) {
	// Read auth file
	encodedData, err := ioutil.ReadFile(c.authFile)
	if err != nil {
		return nil, fmt.Errorf("error reading auth file: %v", err)
	}

	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(string(encodedData))
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 data: %v", err)
	}

	var auth types.AuthData
	if err := json.Unmarshal(decodedData, &auth); err != nil {
		return nil, fmt.Errorf("error parsing auth data: %v", err)
	}

	if auth.VehicleID == "" {
		return nil, fmt.Errorf("no vehicle ID found in auth file")
	}

	return &auth, nil
}

// GetCSRFToken gets a CSRF token from the Rivian API
func (c *Client) GetCSRFToken() (*types.CSRFResponse, error) {
	csrfQuery := `{
		"operationName": "CreateCSRFToken",
		"variables": null,
		"query": "mutation CreateCSRFToken { createCsrfToken { __typename csrfToken appSessionToken } }"
	}`

	var csrfResp types.CSRFResponse
	if err := c.client.DoRequest("POST", "", csrfQuery, nil, &csrfResp); err != nil {
		return nil, err
	}

	return &csrfResp, nil
}

// GetVehicleState gets the current state of the vehicle
func (c *Client) GetVehicleState() (*types.VehicleState, error) {
	auth, err := c.GetAuthData()
	if err != nil {
		return nil, fmt.Errorf("error getting auth data: %v", err)
	}

	query := `{
		"operationName": "GetVehicleState",
		"variables": {
			"vehicleId": "%s"
		},
		"query": "query GetVehicleState($vehicleId: ID!) { vehicleState(vehicleId: $vehicleId) { cabinClimateInteriorTemperature { value } powerState { value } driveMode { value } gearStatus { value } vehicleMileage { value } batteryLevel { value } range { value } chargerStatus { value } chargeState { value } batteryLimit { value } chargeEndTime { value } chargeElapsedTime { value } chargePower { value } chargeSession { value } batteryCapacity { value } batteryEnergyRemaining { value } } }"
	}`

	headers := map[string]string{
		"Authorization": "Bearer " + auth.Token,
	}

	var vehicleState types.VehicleState
	if err := c.client.DoRequest("POST", "", fmt.Sprintf(query, auth.VehicleID), headers, &vehicleState); err != nil {
		return nil, err
	}

	return &vehicleState, nil
} 