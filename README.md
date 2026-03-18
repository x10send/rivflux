# Rivflux

A tool to collect and visualize Rivian vehicle data using InfluxDB and Grafana.

## Features

- Collects vehicle data directly from Rivian's API
- Stores data in InfluxDB
- Visualizes data using Grafana dashboards
- Supports authentication with Rivian's API (including MFA)
- Configurable polling interval

## Prerequisites

- Go 1.19 or later
- Docker and Docker Compose (optional)
- Rivian account credentials
- InfluxDB token (for writing data)

## Setup

### Option 1: Running with Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone git@github.com:rivflux/rivflux.git
cd rivflux
```

2. Create a `.env` file with your InfluxDB and Grafana credentials:
```bash
cp .env.example .env
# Edit .env with your actual credentials
```
**Note:** Before starting the data collector, you must complete the authentication process as described in the "Authentication Process" section.

3. Start all services:
```bash
docker-compose up -d
```

### Option 2: Running Locally

1. Clone the repository:
```bash
git clone git@github.com:rivflux/rivflux.git
cd rivflux
```

2. Build the binaries:
```bash
cd src/rivian
go build -o bin/auth cmd/auth/main.go
go build -o bin/logger cmd/logger/main.go
```

3. Set up InfluxDB:
   - Install InfluxDB locally or use a cloud instance
   - Create an organization and bucket
   - Generate an API token with write permissions

4. Create a `.env` file with your InfluxDB credentials:
```bash
cp .env.example .env
# Edit .env with your actual credentials
```
**Note:** Before starting the data collector, you must complete the authentication process as described in the "Authentication Process" section.

5. Start the data collector:
```bash
./bin/logger \
  -auth-file=data/auth/auth.json \
  -influx-url=http://localhost:8086 \
  -influx-token=your_influx_token \
  -influx-org=rivflux \
  -influx-bucket=rivian \
  -poll-interval=300
```

## Authentication Process

The authentication process works in two steps:

1. **Generate Auth Token** (`auth` command):
   ```bash
   # Without MFA:
   ./bin/auth -username=your.email@example.com -password=your_password -output=data/auth/auth.json

   # With MFA (you'll receive an OTP code via email or your MFA tool):
   ./bin/auth -username=your.email@example.com -password=your_password -otp=123456 -output=data/auth/auth.json
   ```
   - The command will authenticate with Rivian's API
   - If MFA is enabled, you'll need to provide the OTP code
   - You'll be prompted to select your vehicle
   - The auth token and vehicle ID will be saved to the specified file

2. **Use Auth Token** (`logger` command):
   - The logger reads the auth token from the file
   - Makes authenticated requests to Rivian's API
   - Automatically refreshes the token when needed

## Data Collection

The collector polls the following data points every 5 minutes (configurable):
- Cabin temperature
- Power state
- Drive mode
- Gear status
- Vehicle mileage
- Battery level and capacity
- Range
- Charger status and power
- Charging session details

## Dashboard Features

The Grafana dashboard provides comprehensive visualization of your Rivian vehicle data:

![Vehicle Overview](https://github.com/rivflux/rivflux/assets/155249827/8e26413f-7e95-4726-a5bb-c280eeb36ac0)
- Real-time vehicle status and metrics
- Battery level and charging information
- Drive mode and gear status
- Temperature and climate control data

![Charging Details](https://github.com/rivflux/rivflux/assets/155249827/cfea97ed-038d-4872-8275-a836e5ba79fd)
- Detailed charging session information
- Power delivery and efficiency metrics
- Charging rate and duration
- Cost and energy consumption

![Historical Data](https://github.com/rivflux/rivflux/assets/155249827/412949af-9a7d-44ac-9f7e-2d396d63b7b4)
- Historical trends and patterns
- Range and efficiency analysis
- Usage statistics over time
- Customizable time ranges and views

## Accessing the Dashboard

1. Open Grafana at http://localhost:3000
2. Login with:
   - Username: admin
   - Password: (your GRAFANA_PASSWORD)

## Configuration Options

### Auth Generator
```
-username    Rivian account email (required)
-password    Rivian account password (required)
-otp         OTP code for MFA (if enabled)
-output      Output file path (default: auth.json)
```

### Data Logger
```
-auth-file      Path to auth file (default: auth.json)
-influx-url     InfluxDB URL (default: http://influxdb:8086)
-influx-token   InfluxDB token (required)
-influx-org     InfluxDB organization (default: rivflux)
-influx-bucket  InfluxDB bucket (default: rivian)
-poll-interval  Polling interval in seconds (default: 300)
```

## Troubleshooting

1. Auth Token Issues:
```bash
# Regenerate auth token
./bin/auth -username=your.email -password=your_password -output=data/auth/auth.json
```

2. Check collector logs:
```bash
# If running in Docker:
docker-compose logs rivian-collector

# If running locally:
./bin/logger -auth-file=data/auth/auth.json [other options] 2>&1 | tee collector.log
```

3. Verify InfluxDB connection:
```bash
curl -I http://localhost:8086/health
```

4. Check Grafana logs:
```bash
docker-compose logs grafana
```

## Security Notes

- **Account Security**:
  - Create a secondary Rivian account specifically for this tool
  - Do not use your primary Rivian account to avoid potential security risks
  - The secondary account should have limited access to only the necessary vehicle data
  - This follows the principle of least privilege and minimizes potential impact if credentials are compromised

- **Credential Management**:
  - The auth token file is base64 encoded and should be kept secure
  - File permissions are automatically set to 600 (user read/write only)
  - Never commit your auth token file to version control
  - Store your `.env` file securely and never commit it
  - Regularly rotate your InfluxDB tokens and Rivian account password

- **Network Security**:
  - When running locally, ensure InfluxDB is properly secured
  - Use HTTPS for all external connections
  - Consider using a VPN if accessing the dashboard remotely
  - Implement proper firewall rules to restrict access to services

- **Data Protection**:
  - Regularly backup your InfluxDB data
  - Monitor for unauthorized access attempts
  - Review access logs periodically
  - Consider implementing rate limiting on your services

## License

MIT License
