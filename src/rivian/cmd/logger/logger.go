package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rivflux/rivian/pkg/rivian"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func runLogger(authFile, influxURL, influxToken, influxOrg, influxBucket string, pollInterval int) error {
	// Create InfluxDB client
	client := influxdb2.NewClient(influxURL, influxToken)
	defer client.Close()

	// Check InfluxDB connection
	health, err := client.Health(context.Background())
	if err != nil {
		return fmt.Errorf("InfluxDB health check failed: %v", err)
	}
	if health.Status != domain.HealthCheckStatusPass {
		return fmt.Errorf("InfluxDB is not healthy: %v", health.Message)
	}

	// Get write API
	writeAPI := client.WriteAPIBlocking(influxOrg, influxBucket)

	// Create Rivian client
	rivianClient := rivian.NewClient(authFile, true)

	// Main loop
	for {
		// Get vehicle state
		state, err := rivianClient.GetVehicleState()
		if err != nil {
			log.Printf("Error getting vehicle state: %v", err)
			time.Sleep(time.Duration(pollInterval) * time.Second)
			continue
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

		// Write to InfluxDB
		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Printf("Error writing to InfluxDB: %v", err)
		} else {
			log.Printf("Successfully wrote data at %v", time.Now())
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
} 