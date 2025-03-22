package main

import (
	"testing"
	"time"

	"github.com/rivflux/rivian/pkg/types"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func TestInfluxDBPointCreation(t *testing.T) {
	// Create a sample vehicle state
	state := types.VehicleState{
		Data: struct {
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
		}{
			VehicleState: struct {
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
			}{
				CabinClimateInteriorTemperature: struct {
					Value float64 `json:"value"`
				}{Value: 22.5},
				PowerState: struct {
					Value string `json:"value"`
				}{Value: "ON"},
				DriveMode: struct {
					Value string `json:"value"`
				}{Value: "SPORT"},
				GearStatus: struct {
					Value string `json:"value"`
				}{Value: "PARK"},
				VehicleMileage: struct {
					Value float64 `json:"value"`
				}{Value: 1234.5},
				BatteryLevel: struct {
					Value float64 `json:"value"`
				}{Value: 85.5},
				Range: struct {
					Value float64 `json:"value"`
				}{Value: 250.0},
				ChargerStatus: struct {
					Value string `json:"value"`
				}{Value: "DISCONNECTED"},
				ChargeState: struct {
					Value string `json:"value"`
				}{Value: "NOT_CHARGING"},
				BatteryLimit: struct {
					Value float64 `json:"value"`
				}{Value: 85.0},
				ChargeEndTime: struct {
					Value string `json:"value"`
				}{Value: "2024-03-22T12:00:00Z"},
				ChargeElapsedTime: struct {
					Value float64 `json:"value"`
				}{Value: 3600.0},
				ChargePower: struct {
					Value float64 `json:"value"`
				}{Value: 11.5},
				ChargeSession: struct {
					Value float64 `json:"value"`
				}{Value: 1.0},
				BatteryCapacity: struct {
					Value float64 `json:"value"`
				}{Value: 135.0},
				BatteryEnergyRemaining: struct {
					Value float64 `json:"value"`
				}{Value: 115.0},
			},
		},
	}

	// Create point
	point := write.NewPoint("rivian",
		nil,
		map[string]interface{}{
			"cabin_temp":              state.Data.VehicleState.CabinClimateInteriorTemperature.Value,
			"power_state":             state.Data.VehicleState.PowerState.Value,
			"drive_mode":              state.Data.VehicleState.DriveMode.Value,
			"gear":                    state.Data.VehicleState.GearStatus.Value,
			"mileage":                 state.Data.VehicleState.VehicleMileage.Value,
			"battery_level":           state.Data.VehicleState.BatteryLevel.Value,
			"range":                   state.Data.VehicleState.Range.Value,
			"charger_status":          state.Data.VehicleState.ChargerStatus.Value,
			"charge_state":            state.Data.VehicleState.ChargeState.Value,
			"battery_limit":           state.Data.VehicleState.BatteryLimit.Value,
			"charge_end_time":         state.Data.VehicleState.ChargeEndTime.Value,
			"charge_elapsed_time":     state.Data.VehicleState.ChargeElapsedTime.Value,
			"charge_power":            state.Data.VehicleState.ChargePower.Value,
			"charge_session":          state.Data.VehicleState.ChargeSession.Value,
			"battery_capacity":        state.Data.VehicleState.BatteryCapacity.Value,
			"battery_energy_remaining": state.Data.VehicleState.BatteryEnergyRemaining.Value,
		},
		time.Now(),
	)

	// Verify point fields
	if point.Name() != "rivian" {
		t.Errorf("Expected point name to be 'rivian', got '%s'", point.Name())
	}

	fields := point.FieldList()
	expectedFields := map[string]interface{}{
		"cabin_temp":              22.5,
		"power_state":             "ON",
		"drive_mode":              "SPORT",
		"gear":                    "PARK",
		"mileage":                 1234.5,
		"battery_level":           85.5,
		"range":                   250.0,
		"charger_status":          "DISCONNECTED",
		"charge_state":            "NOT_CHARGING",
		"battery_limit":           85.0,
		"charge_end_time":         "2024-03-22T12:00:00Z",
		"charge_elapsed_time":     3600.0,
		"charge_power":            11.5,
		"charge_session":          1.0,
		"battery_capacity":        135.0,
		"battery_energy_remaining": 115.0,
	}

	for _, field := range fields {
		expected, ok := expectedFields[field.Key]
		if !ok {
			t.Errorf("Unexpected field: %s", field.Key)
			continue
		}
		if field.Value != expected {
			t.Errorf("Field %s = %v, want %v", field.Key, field.Value, expected)
		}
	}
} 