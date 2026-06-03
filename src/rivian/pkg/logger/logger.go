package logger

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/x10send/rivflux/pkg/rivian"
)

// Run polls Rivian and writes vehicle data to InfluxDB until ctx is cancelled.
func Run(ctx context.Context, authFile, influxURL, influxToken, influxOrg, influxBucket string, pollInterval int) error {
	client := influxdb2.NewClient(influxURL, influxToken)
	defer client.Close()

	health, err := client.Health(ctx)
	if err != nil {
		return fmt.Errorf("InfluxDB health check failed: %v", err)
	}
	if health.Status != domain.HealthCheckStatusPass {
		return fmt.Errorf("InfluxDB is not healthy: %v", health.Message)
	}

	writeAPI := client.WriteAPIBlocking(influxOrg, influxBucket)
	rivianClient := rivian.NewClient(authFile, false)

	for {
		state, err := rivianClient.GetVehicleState()
		if err != nil {
			log.Printf("Error getting vehicle state: %v", err)
		} else {
			point := write.NewPoint("rivian",
				nil,
				map[string]interface{}{
					"cabin_temp":               state.Data.VehicleState.CabinClimateInteriorTemperature.Value,
					"power_state":              state.Data.VehicleState.PowerState.Value,
					"drive_mode":               state.Data.VehicleState.DriveMode.Value,
					"gear":                     state.Data.VehicleState.GearStatus.Value,
					"mileage":                  state.Data.VehicleState.VehicleMileage.Value,
					"battery_level":            state.Data.VehicleState.BatteryLevel.Value,
					"range":                    state.Data.VehicleState.Range.Value,
					"charger_status":           state.Data.VehicleState.ChargerStatus.Value,
					"charge_state":             state.Data.VehicleState.ChargeState.Value,
					"battery_limit":            state.Data.VehicleState.BatteryLimit.Value,
					"charge_end_time":          state.Data.VehicleState.ChargeEndTime.Value,
					"charge_elapsed_time":      state.Data.VehicleState.ChargeElapsedTime.Value,
					"charge_power":             state.Data.VehicleState.ChargePower.Value,
					"charge_session":           state.Data.VehicleState.ChargeSession.Value,
					"battery_capacity":         state.Data.VehicleState.BatteryCapacity.Value,
					"battery_energy_remaining": state.Data.VehicleState.BatteryEnergyRemaining.Value,
				},
				time.Now(),
			)
			if err := writeAPI.WritePoint(ctx, point); err != nil {
				log.Printf("Error writing to InfluxDB: %v", err)
			} else {
				log.Printf("Data written at %v", time.Now())
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(pollInterval) * time.Second):
		}
	}
}
