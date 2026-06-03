package rivian

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/x10send/rivflux/pkg/httpclient"
	"github.com/x10send/rivflux/pkg/types"
)

type Client struct {
	client   *httpclient.Client
	authFile string
}

func NewClient(authFile string, debug bool) *Client {
	return &Client{
		client:   httpclient.NewClient(types.RivianAPIPath, debug),
		authFile: authFile,
	}
}

func (c *Client) GetAuthData() (*types.AuthData, error) {
	encodedData, err := os.ReadFile(c.authFile)
	if err != nil {
		return nil, fmt.Errorf("error reading auth file %s: %v", c.authFile, err)
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(encodedData))
	if err != nil {
		return nil, fmt.Errorf("error decoding auth file %s: %v", c.authFile, err)
	}

	var auth types.AuthData
	if err := json.Unmarshal(decodedData, &auth); err != nil {
		return nil, fmt.Errorf("error parsing auth file %s: %v", c.authFile, err)
	}

	if auth.VehicleID == "" {
		return nil, fmt.Errorf("no vehicle ID found in auth file %s", c.authFile)
	}

	return &auth, nil
}

func (c *Client) GetVehicleState() (*types.VehicleState, error) {
	auth, err := c.GetAuthData()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{
		"operationName": "GetVehicleState",
		"variables":     map[string]string{"vehicleId": auth.VehicleID},
		"query":         "query GetVehicleState($vehicleId: ID!) { vehicleState(vehicleId: $vehicleId) { cabinClimateInteriorTemperature { value } powerState { value } driveMode { value } gearStatus { value } vehicleMileage { value } batteryLevel { value } range { value } chargerStatus { value } chargeState { value } batteryLimit { value } chargeEndTime { value } chargeElapsedTime { value } chargePower { value } chargeSession { value } batteryCapacity { value } batteryEnergyRemaining { value } } }",
	})
	if err != nil {
		return nil, fmt.Errorf("error building vehicle state request: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + auth.Token,
	}

	var vehicleState types.VehicleState
	if err := c.client.DoRequest("POST", "", string(body), headers, &vehicleState); err != nil {
		return nil, err
	}
	return &vehicleState, nil
}
