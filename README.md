# Rivflux

**note, v2 is wip, dont use?**

A tool to collect and visualize Rivian vehicle data using InfluxDB and Grafana.

## Features

- Collects vehicle data directly from Rivian's API
- Stores data in InfluxDB
- Visualizes data using Grafana dashboards
- Supports authentication with Rivian's API (including MFA)
- Configurable polling interval

## Prerequisites

- Go 1.19 or later
- Docker and Docker Compose
- Rivian account credentials
- InfluxDB token (for writing data)

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/rivflux.git
cd rivflux
```

2. Build the binaries:
```bash
cd src/rivian
go build -o bin/auth cmd/auth/main.go
go build -o bin/logger cmd/logger/main.go
```

3. Generate the authentication token:
```bash
# Without MFA:
./bin/auth -username=your.email@example.com -password=your_password -output=data/auth/auth.json

# With MFA (you'll receive an OTP code via email or your MFA tool):
./bin/auth -username=your.email@example.com -password=your_password -otp=123456 -output=data/auth/auth.json
```

4. Create a `.env` file with your InfluxDB and Grafana credentials:
```bash
cp .env.example .env
# Edit .env with your actual credentials
```

5. Start the services:
```bash
docker-compose up -d influxdb grafana
```

6. Start the data collector:
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
   - Authenticates with Rivian's API
   - Handles MFA if enabled on your account
   - Lists your vehicles and lets you select one
   - Saves the auth token and vehicle ID to an encrypted file

2. **Use Auth Token** (`logger` command):
   - Reads the auth token from the file
   - Uses it to make authenticated requests to Rivian's API
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

- The auth token file is base64 encoded and should be kept secure
- File permissions are set to 600 (user read/write only)
- Never commit your auth token file to version control
- Store your `.env` file securely and never commit it

## License

MIT License
