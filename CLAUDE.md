# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Purpose

Fork of [bttnns/rivflux](https://github.com/bttnns/rivflux). Collects Rivian vehicle telemetry via the Rivian GraphQL API and writes it to InfluxDB. Adds a built-in web setup UI so authentication can be done through the browser rather than the CLI, making it suitable for Unraid Community Apps.

## Repository Structure

```
rivflux/
├── .github/workflows/release.yml   # Multi-arch GHCR build on v*.*.*-* tags
├── docker/rivian/Dockerfile         # Builds the server binary (amd64 + arm64)
├── src/rivian/                      # Go source (module: github.com/x10send/rivflux)
│   ├── cmd/
│   │   ├── auth/        # Standalone CLI auth tool (kept for debugging)
│   │   ├── logger/      # Standalone CLI logger (kept for debugging)
│   │   └── server/      # Combined server: web UI + logger goroutine (Docker entrypoint)
│   └── pkg/
│       ├── auth/        # Rivian auth flow (InitialLogin, CompleteMFA)
│       ├── httpclient/  # Shared HTTP client
│       ├── logger/      # Reusable logger logic (extracted from cmd/logger)
│       ├── models/      # Vehicle model
│       ├── rivian/      # Rivian API client (GetVehicleState, token refresh)
│       ├── setup/       # HTTP handler for the setup web UI
│       └── types/       # Shared types and constants
├── config/grafana/      # Grafana datasource + dashboard provisioning
├── unraid/rivflux.xml   # Unraid Community Apps template
└── docker-compose.yaml  # Full-stack compose (InfluxDB + collector + Grafana)
```

## Versioning Convention

Tags follow `vX.Y.Z-N` where `X.Y.Z` is the upstream rivflux version and `N` is the packaging iteration (starting at 1). Bump `N` for Dockerfile or workflow changes without a source change. Published Docker tags: `X.Y.Z-N` (full) and `latest`.

## Build & Run Commands

```bash
export PATH="/usr/local/go/bin:$PATH"

# Build all packages (from src/rivian/)
go build ./...
go vet ./...

# Build Docker image locally (from repo root)
docker build -f docker/rivian/Dockerfile -t rivflux .

# Run locally
docker run --rm \
  -e INFLUX_TOKEN=<token> \
  -e INFLUX_URL=http://192.168.1.x:8086 \
  -e INFLUX_ORG=rivflux \
  -e INFLUX_BUCKET=rivian \
  -v /tmp/rivflux-data:/data \
  -p 8888:8888 \
  rivflux
```

## Setup Web UI

The `cmd/server` binary always starts an HTTP server on port 8888 (configurable via `SETUP_PORT`). On first run with no `auth.json` present, it shows an auth form. On submit it calls `pkg/auth.InitialLogin`; if MFA is required it shows a one-time code form and calls `pkg/auth.CompleteMFA`. Once auth.json is written the logger goroutine starts automatically. Navigating to `/` after auth shows a status page with a re-auth form.

Credentials (username/password) are never stored — only the resulting session token written to `/data/auth.json`.

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `INFLUX_TOKEN` | Yes | — | InfluxDB write token |
| `INFLUX_URL` | Yes | `http://influxdb:8086` | InfluxDB base URL |
| `INFLUX_ORG` | No | `rivflux` | InfluxDB organization |
| `INFLUX_BUCKET` | No | `rivian` | InfluxDB bucket |
| `POLL_INTERVAL` | No | `300` | Poll interval in seconds |
| `SETUP_PORT` | No | `8888` | Web UI port |
| `AUTH_FILE` | No | `/data/auth.json` | Path to session token file |

## GitHub Actions: release.yml

Same pattern as `x10send/influxdb-mcp-server`. Triggers on `v*.*.*-*` tags. Native amd64 + arm64 runners (no QEMU). Merges digests into a manifest list and attests provenance. Required permissions: `contents: read`, `packages: write`, `attestations: write`, `id-token: write`.

## Dockerfile

- Context: repo root (Dockerfile at `docker/rivian/Dockerfile`)
- Builder: `golang:1.21-alpine` — builds `server` and `auth` binaries
- Runtime: `alpine:latest` — non-root user `rivflux`, exposes port 8888
- Entrypoint: `./server`

## Unraid Template (unraid/rivflux.xml)

- Repository: `ghcr.io/x10send/rivflux:latest`
- Port: container `8888` → host `8888` (WebUI)
- Volume: `/data` → `/mnt/user/appdata/rivflux` (auth token persistence)
- Env vars: `INFLUX_TOKEN`, `INFLUX_URL`, `INFLUX_ORG`, `INFLUX_BUCKET`, `POLL_INTERVAL`

## Quality Bar

- `go build ./...` and `go vet ./...` must pass clean
- Image runs as non-root
- Multi-arch (amd64 + arm64)
- SBOM + provenance attestation on release
