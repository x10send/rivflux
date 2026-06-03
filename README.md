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

Install via Community Apps (search **Rivflux**), or from the **Docker** tab click **Add Container** and enter `ghcr.io/x10send/rivflux:latest` as the repository.

Required: `INFLUX_TOKEN` and `INFLUX_URL` (your InfluxDB address). Org, bucket, poll interval, and the host-side port (default **8888**) are all configurable in the template.

Open the WebUI after starting the container, enter your Rivian credentials, and complete MFA if prompted. The collector starts automatically once authenticated.

To visualize data, import `config/grafana/dashboards/Rivian.json` into your existing Grafana instance and point it at your InfluxDB datasource.

## Quick Start: Docker Compose

The compose file brings up InfluxDB, Rivflux, and Grafana together with the dashboard pre-provisioned. If you already have Grafana, you can remove the Grafana service and skip `GRAFANA_PASSWORD`.

```bash
git clone https://github.com/x10send/rivflux.git
cd rivflux
```

Copy and edit the environment file:

```bash
cp .env.example .env
# Required: INFLUX_TOKEN
# Required if using bundled InfluxDB: INFLUX_PASSWORD
# Required if using bundled Grafana: GRAFANA_PASSWORD
# Optional: INFLUX_ORG, INFLUX_BUCKET, POLL_INTERVAL, SETUP_PORT
```

Start the stack:

```bash
docker compose up -d
```

Open `http://localhost:8888` (or your configured `SETUP_PORT`) and authenticate with your Rivian credentials. Grafana is available at `http://localhost:3000`.

## Docker Image

```
ghcr.io/x10send/rivflux:latest
```

| Variable | Required | Default | Description |
|---|---|---|---|
| `INFLUX_TOKEN` | Yes | — | InfluxDB write token |
| `INFLUX_URL` | Yes | — | InfluxDB base URL (e.g. `http://192.168.1.x:8086`) |
| `INFLUX_ORG` | No | `rivflux` | InfluxDB organization |
| `INFLUX_BUCKET` | No | `rivian` | InfluxDB bucket |
| `POLL_INTERVAL` | No | `300` | Seconds between polls |
| `SETUP_PORT` | No | `8888` | Web UI port — change if 8888 is already in use |

Mount a persistent volume at `/data` to survive container restarts without re-authenticating.

## Dashboard

![Vehicle Overview](https://github.com/bttnns/rivflux/assets/155249827/8e26413f-7e95-4726-a5bb-c280eeb36ac0)

![Charging Details](https://github.com/bttnns/rivflux/assets/155249827/cfea97ed-038d-4872-8275-a836e5ba79fd)

![Historical Data](https://github.com/bttnns/rivflux/assets/155249827/412949af-9a7d-44ac-9f7e-2d396d63b7b4)

The dashboard JSON is at `config/grafana/dashboards/Rivian.json`. When using Docker Compose it is provisioned automatically. For standalone Grafana, import it manually and configure an InfluxDB datasource pointing at your instance.

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
