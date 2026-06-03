package main

import (
	"context"
	"flag"
	"log"

	"github.com/x10send/rivflux/pkg/logger"
)

func main() {
	authFile := flag.String("auth-file", "auth.json", "Path to auth file")
	influxURL := flag.String("influx-url", "http://influxdb:8086", "InfluxDB URL")
	influxToken := flag.String("influx-token", "", "InfluxDB token")
	influxOrg := flag.String("influx-org", "rivflux", "InfluxDB organization")
	influxBucket := flag.String("influx-bucket", "rivian", "InfluxDB bucket")
	pollInterval := flag.Int("poll-interval", 300, "Polling interval in seconds")
	flag.Parse()

	if *influxToken == "" {
		log.Fatal("InfluxDB token is required")
	}

	if err := logger.Run(context.Background(), *authFile, *influxURL, *influxToken, *influxOrg, *influxBucket, *pollInterval); err != nil {
		log.Fatal(err)
	}
}
