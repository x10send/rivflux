package main

import (
	"context"

	"github.com/x10send/rivflux/pkg/logger"
)

func runLogger(authFile, influxURL, influxToken, influxOrg, influxBucket string, pollInterval int) error {
	return logger.Run(context.Background(), authFile, influxURL, influxToken, influxOrg, influxBucket, pollInterval)
}
