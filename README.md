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

---

## Unraid Setup Guide

If you already have InfluxDB and Grafana running, skip to [step 3](#3-install-rivflux).

### 1. Install InfluxDB

In the **Apps** tab search for **InfluxDB** and install it. The default port is `8086`.

Once the container is running, open `http://[UNRAID-IP]:8086` to complete the initial setup:

1. Create an admin username and password
2. Set the **Organization** name — use `rivflux` to match the defaults, or any name you prefer
3. Set the **Bucket** name — use `rivian`, or any name you prefer
4. Click **Continue** — InfluxDB will display an **API token**. **Copy it now**, it won't be shown again. This is your `INFLUX_TOKEN`.

If you missed the token, generate a new one under **Load Data → API Tokens → Generate API Token → All Access Token**.

### 2. Install Grafana

In the **Apps** tab search for **Grafana** and install it. The default port is `3000`.

Open `http://[UNRAID-IP]:3000` and log in (default: `admin` / `admin`). Then:

**Add the InfluxDB datasource:**

1. Go to **Connections → Data Sources → Add new data source**
2. Select **InfluxDB**
3. Set **Query Language** to **Flux**
4. Set **URL** to `http://[UNRAID-IP]:8086`
5. Under **InfluxDB Details**, enter your **Organization** and **Token**
6. Set **Default Bucket** to your bucket name (e.g. `rivian`)
7. Click **Save & Test** — you should see a success message

**Import the dashboard:**

1. Go to **Dashboards → Import**
2. Click **Upload dashboard JSON file**
3. Select `config/grafana/dashboards/Rivian.json` from this repo
4. Select the InfluxDB datasource you just created
5. Click **Import**

### 3. Install Rivflux

In the **Apps** tab search for **Rivflux** and install it, or from the **Docker** tab click **Add Container** and enter `ghcr.io/x10send/rivflux:latest` as the repository.

Configure the following variables in the template:

| Variable | Value |
|---|---|
| `INFLUX_TOKEN` | The API token from step 1 |
| `INFLUX_URL` | `http://[UNRAID-IP]:8086` |
| `INFLUX_ORG` | Your organization name (default: `rivflux`) |
| `INFLUX_BUCKET` | Your bucket name (default: `rivian`) |

The host-side port defaults to **8888** — change it in the template if that port is already in use.

Start the container, then open the WebUI (`http://[UNRAID-IP]:8888`), enter your Rivian credentials, and complete MFA if prompted. The collector starts automatically once authenticated. Data will appear in Grafana within one poll interval (default: 5 minutes).

---

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

The dashboard JSON is at `config/grafana/dashboards/Rivian.json`. When using Docker Compose it is provisioned automatically. For standalone Grafana, import it manually as described in the Unraid setup guide above.

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
