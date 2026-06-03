# Rivflux

Collects Rivian vehicle telemetry via the Rivian GraphQL API and writes it to InfluxDB. Ships with a built-in browser-based setup UI so authentication requires no CLI access — making it suitable for Unraid Community Apps and other container environments.

Fork of [bttnns/rivflux](https://github.com/bttnns/rivflux).

## Features

- Browser-based authentication (including MFA) — no CLI required
- Polls Rivian's API on a configurable interval and writes to InfluxDB
- Grafana dashboard for visualization
- Multi-arch Docker image (`linux/amd64` + `linux/arm64`)
- Unraid Community Apps compatible

## Collected Data Points

Every poll (default: 5 minutes):

- Battery level, capacity, and energy remaining
- Estimated range
- Charger status, charge state, charge power, and elapsed charge time
- Cabin temperature
- Power state, drive mode, and gear status
- Vehicle mileage

## Quick Start: Unraid

Install via Community Apps (search **Rivflux**) or add the template manually from `unraid/rivflux.xml`. Required variables: `INFLUX_TOKEN`, `INFLUX_URL`.

The container exposes a setup UI on port **8888**. Open it in your browser, enter your Rivian credentials, and complete MFA if prompted. Once authenticated the collector starts automatically.

## Quick Start: Docker Compose

```bash
git clone https://github.com/x10send/rivflux.git
cd rivflux
```

Copy and edit the environment file:

```bash
cp .env.example .env
# Set INFLUX_TOKEN, INFLUX_URL, GRAFANA_PASSWORD
```

Start the stack:

```bash
docker compose up -d
```

Open `http://localhost:8888` and authenticate with your Rivian credentials.

## Docker Image

```
ghcr.io/x10send/rivflux:latest
```

| Variable | Required | Default | Description |
|---|---|---|---|
| `INFLUX_TOKEN` | Yes | — | InfluxDB write token |
| `INFLUX_URL` | Yes | `http://influxdb:8086` | InfluxDB base URL |
| `INFLUX_ORG` | No | `rivflux` | InfluxDB organization |
| `INFLUX_BUCKET` | No | `rivian` | InfluxDB bucket |
| `POLL_INTERVAL` | No | `300` | Seconds between polls |
| `SETUP_PORT` | No | `8888` | Web UI port |
| `AUTH_FILE` | No | `/data/auth.json` | Session token path |

Mount a persistent volume at `/data` to survive container restarts without re-authenticating.

## Dashboard

The included Grafana dashboard (provisioned automatically via `config/grafana/`) shows:

![Vehicle Overview](https://github.com/rivflux/rivflux/assets/155249827/8e26413f-7e95-4726-a5bb-c280eeb36ac0)

![Charging Details](https://github.com/rivflux/rivflux/assets/155249827/cfea97ed-038d-4872-8275-a836e5ba79fd)

![Historical Data](https://github.com/rivflux/rivflux/assets/155249827/412949af-9a7d-44ac-9f7e-2d396d63b7b4)

## Security Notes

- **Use a secondary Rivian account** with access to only the vehicles you want to monitor. This limits exposure if the session token is ever compromised.
- Credentials (username/password) are never stored. Only the resulting session token is written to disk (`/data/auth.json`), with permissions set to `0600`.
- Never commit your auth file or `.env` to version control.

## Building from Source

Requires Go 1.21+.

```bash
cd src/rivian
go build ./...
go test ./...
```

Docker (from repo root):

```bash
docker build -f docker/rivian/Dockerfile -t rivflux .
```

## License

MIT License
