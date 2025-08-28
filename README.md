# Watchdog

[![Coverage Status](https://coveralls.io/repos/github/OneBusAway/watchdog/badge.svg?branch=main)](https://coveralls.io/github/OneBusAway/watchdog?branch=main)

**Watchdog** is a Go-based service that monitors [OneBusAway (OBA)](https://onebusaway.org/) REST API servers.  
It exposes a comprehensive set of **Prometheus metrics** for monitoring:

- GTFS Static and GTFS-RT data integrity
- Vehicle telemetry
- Agency and stop coverage
- Overall operational health
  See the full list of metrics and interpretation guide [here](./docs/METRICS.md)

## Table of Contents

- [Requirements](#requirements)
- [Setup](#setup)
  - [Configuration](#configuration)
    - [Example config.json](#example-configjson)
    - [Ways to Provide the Config File](#ways-to-provide-the-config-file)
      - [Local Configuration (recommended for development)](#1-local-configuration-recommended-for-development)
      - [Remote Configuration (recommended for production)](#2-remote-configuration-recommended-for-production)
  - [Application Options](#application-options)
  - [Environment Variables](#environment-variables)
- [Running](#running)
  - [Docker Compose (recommended)](#1-docker-compose-recommended)
  - [Watchdog Only](#2-watchdog-only)
    - [Local Config](#local-config)
    - [Remote Config (with auth)](#remote-config-with-auth)
  - [Docker (single container)](#3-docker-single-container)
    - [Build image](#build-image)
    - [Run with local config](#run-with-local-config)
    - [Run with remote config](#run-with-remote-config)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [Integration Tests](#integration-tests)

## Requirements

- **Go 1.23+**

## Setup

### Configuration

Watchdog requires a configuration file (`config.json`) before running. Even placeholder data is necessary to start the service.

#### Example `config.json`

```json
[
  {
    "name": "Test Server 1",
    "id": 1,
    "oba_base_url": "https://test1.example.com",
    "oba_api_key": "test-key-1",
    "gtfs_url": "https://gtfs1.example.com",
    "trip_update_url": "https://trip1.example.com",
    "vehicle_position_url": "https://vehicle1.example.com",
    "gtfs_rt_api_key": "api-key-1",
    "gtfs_rt_api_value": "api-value-1",
    "agency_id": "agency-1"
  }
]
```

#### Ways to Provide the Config File

#### 1. Local Configuration (recommended for development)

- Copy or rename `config.json.template` → `config.json`
- Fill in your server values
- Run with:

```bash
go run ./cmd/watchdog/ --config-file path/to/config.json
```

Note:

- ⚠️The file **must** be named `config.json`
- `config.json` is Git-ignored (to protect secrets)

#### 2. Remote Configuration (recommended for production)

- Prepare `config.json` as above
- Host it publicly (or on a private server)
- Run with:

```bash
go run ./cmd/watchdog/ --config-url http://example.com/config.json
```

If authentication is required, set:

```bash
export CONFIG_AUTH_USER="username"
export CONFIG_AUTH_PASS="password"
```

### Application Options

- **Fetch Interval** → default `30s` (`--fetch-interval <seconds>`)
- **Environment** → `development` (default), `staging`, `production` (`--env <value>`)
- **Port** → default `4000` (`--port <number>`)

⚠️ If running with **Docker Compose**, Prometheus runs on `9090` and Grafana on `3000`. Don’t use those ports.

### Environment Variables

- **Sentry DSN**

```bash
    export SENTRY_DSN="your_sentry_dsn"
```

- **Config Auth (for remote configs)**

```bash
    export CONFIG_AUTH_USER="username"
    export CONFIG_AUTH_PASS="password"
```

## Running

### 1. Docker Compose (recommended)

Run Watchdog with **Prometheus + Grafana**:

```bash
docker compose up --build
```

Services:

- Watchdog → `4000`
- Prometheus → `9090`
- Grafana → `3000`

Stop services:

```bash
docker compose down
```

Restart services:

```bash
docker compose restart
```

Grafana auto-loads a Go runtime dashboard. Prometheus is pre-configured to scrape Watchdog.

### 2. Watchdog Only

#### Local Config

```bash
go run ./cmd/watchdog/ --config-file path/to/config.json
```

#### Remote Config (with auth)

```bash
go run ./cmd/watchdog/ \
  --config-url http://example.com/config.json
```

### 3. Docker (single container)

#### Build image

```bash
docker build -t watchdog .
```

#### Run with local config

```bash
docker run -d \
  --name watchdog \
  -v ./config.json:/app/config.json \
  -p 4000:4000 \
  watchdog \
  --config-file /app/config.json
```

#### Run with remote config

```bash
docker run -d \
  --name watchdog \
  -e CONFIG_AUTH_USER=admin \
  -e CONFIG_AUTH_PASS=password \
  -p 4000:4000 \
  watchdog \
  --config-url http://example.com/config.json
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

- Copy `integration_config.json.template` → `integration_config.json`
- Fill in OBA server values
- Run:

```bash
go test -tags=integration ./internal/integration \
  -integration-config path/to/integration_config.json
```

Note:

- ⚠️ the file **must** be named `integration_config.json`
- It’s Git-ignored for safety
