package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/x10send/rivflux/pkg/logger"
	"github.com/x10send/rivflux/pkg/setup"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func main() {
	authFile := flag.String("auth-file", envOr("AUTH_FILE", "/data/auth.json"), "Path to auth file")
	influxURL := flag.String("influx-url", envOr("INFLUX_URL", "http://influxdb:8086"), "InfluxDB URL")
	influxToken := flag.String("influx-token", envOr("INFLUX_TOKEN", ""), "InfluxDB token")
	influxOrg := flag.String("influx-org", envOr("INFLUX_ORG", "rivflux"), "InfluxDB organization")
	influxBucket := flag.String("influx-bucket", envOr("INFLUX_BUCKET", "rivian"), "InfluxDB bucket")
	pollInterval := flag.Int("poll-interval", envIntOr("POLL_INTERVAL", 300), "Polling interval in seconds")
	setupPort := flag.Int("setup-port", envIntOr("SETUP_PORT", 8888), "Port for setup web UI")
	flag.Parse()

	if *influxToken == "" {
		log.Fatal("INFLUX_TOKEN env var (or -influx-token flag) is required")
	}

	// Ensure the directory for the auth file exists.
	if err := os.MkdirAll(filepath.Dir(*authFile), 0700); err != nil {
		log.Fatalf("Cannot create auth file directory: %v", err)
	}

	var (
		mu           sync.Mutex
		loggerCancel context.CancelFunc
	)

	startLogger := func() {
		mu.Lock()
		defer mu.Unlock()
		if _, err := os.Stat(*authFile); err != nil {
			return
		}
		if loggerCancel != nil {
			loggerCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		loggerCancel = cancel
		go func() {
			// Retry loop: if logger exits due to transient error (e.g. InfluxDB
			// temporarily unavailable), restart it after a short delay.
			for {
				if err := logger.Run(ctx, *authFile, *influxURL, *influxToken, *influxOrg, *influxBucket, *pollInterval); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						log.Printf("Logger error (will retry in 30s): %v", err)
					}
				}
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		log.Println("Logger started")
	}

	mux := http.NewServeMux()
	handler := setup.NewHandler(*authFile, startLogger)
	handler.Register(mux)

	startLogger()

	addr := fmt.Sprintf(":%d", *setupPort)
	log.Printf("Setup UI listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
