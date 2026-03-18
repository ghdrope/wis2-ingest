# WIS2 Ingest

> WIS2 Data Ingest

WIS2 Ingest is a lightweight command-line tool that consumes and persists data distributed through the WMO Information System 2.0 (WIS2). Its purpose is primarily to receive real-time data notifications and transforms them into locally accessible files.

By continuously listening to configured data streams/topics, the ingestor automatically retrieves and stores relevant datasets in an organized directories structure (<YEAR>/<MONTH>/<DAY>/<HOUR>/<FILENAME>).

## Quick Access

[![Documentation](https://img.shields.io/badge/Documentation-white?plastic&logo=docusaurus&logoColor=blue&color=yellow&link=https%3A%2F%2Fwww.notion.so%2FSIMUStack-1f471badea8b80719cc8f94476e146a0%3Fpvs%3D4
)](https://www.notion.so/WIS2-Ingest-32771badea8b80f399aed688c7788522?source=copy_link)

## Project Structure

```txt
cmd/        - CLI entrypoints and command definitions
internal/   - Core application logic
pkg/        - Shared utilities
go.mod      - Go module definition and dependencies

Makefile    - Build and development commands
.github/    - CI/CD workflows and repository automation (GitHub Actions)

Dockerfile  - Container image definition

charts/     - Helm charts for deployment in Kubernetes environments
```

## Setup

To run WIS2 Ingest, you only need to configure a few environmnet variables:

Optional:

- WIS2_MQTT_HOST – MQTT broker host (default: globalbroker.meteo.fr)

- WIS2_MQTT_PORT – MQTT broker port (default: 8883)

- WIS2_MQTT_USERNAME – MQTT username (default: everyone)

- WIS2_MQTT_PASSWORD – MQTT password (default: everyone)

**Note**: The default credentials (`everyone` / `everyone`) are publicly shared within the WIS2 ecosystem and are intended for accessing open, non-restricted data streams. They are safe to use and do not grant access to secured or sensitive topics.

Mandatory:

- `WIS2_MQTT_TOPIC` – MQTT topic(s) to subscribe to (required, ; separated)

- `WIS2_OUTPUT_DIRECTORY` – Directory where data will be stored (required)

Example:

```bash
export WIS2_MQTT_TOPIC="origin/a/wis2/#"
export WIS2_OUTPUT_DIRECTORY="data"
```

## Usage

### Run Locally (Go)

First, ensure dependencies are installed and build the binary:

```bash
go mod tidy
make build
```

Then run the ingest command:

```bash
./.bin/wis2-ingest ingestor
```

### Run in Kubernetes

A Helm chart is provided under the `charts/` directory for deployment in a Kubernetes environment.

Basic example:

```bash
helm install wis2-ingest ./charts \
  --set env.WIS2_MQTT_TOPIC="origin/a/wis2/#" \
  --set env.WIS2_OUTPUT_DIRECTORY="data"
```

This deploys the ingest service as a container that continuously consumes and stores data inside the cluster.
