# Production Profiling Setup (Grafana Pyroscope)

This guide shows how to run the Pyroscope stack locally and instrument Go services in this repository. It mirrors a production-ready approach: persistent storage, provisioning, and env-driven config.

## Quick Start

```bash
# 1) Start the stack (Pyroscope, Grafana, Postgres)
make start_profiling_stack

# 2) Build binaries (including Pyroscope-enabled apps)
make build

# 3) Generate 100k events with continuous profiling enabled
make run_pyroscope_generator

# 4) Load events into Postgres with continuous profiling enabled
make run_pyroscope_loader

# UIs
# Pyroscope: http://localhost:4040
# Grafana:  http://localhost:3000  (admin/admin)
```

## Stack Startup

- Requirements: Docker Desktop, Make, Go 1.24+
- Start Pyroscope, Grafana, and Postgres:

```bash
make start_profiling_stack
```

- Services:
  - Pyroscope: http://localhost:4040
  - Grafana: http://localhost:3000 (admin/admin)

## Build and Run the Examples

- Build all binaries:

```bash
make build
```

- Run Pyroscope-enabled generator (writes JSON file):

```bash
make run_pyroscope_generator
```

- Run Pyroscope-enabled loader (loads JSON into Postgres):

```bash
make run_pyroscope_loader
```

Both apps push profiles to Pyroscope automatically. In the UI you’ll see:

- interviewing-golang.generator
- interviewing-golang.loader

## Configuration

The profilers are configured via environment variables:

- PYROSCOPE_SERVER_ADDRESS (default http://localhost:4040)
- PYROSCOPE_APPLICATION_NAME (default set by binary)
- PYROSCOPE_TENANT_ID (optional)
- PYROSCOPE_TAGS (e.g., env=dev,service=generator,version=1.0.0)

## Troubleshooting

- Profiles not visible:
  - Ensure Pyroscope is reachable at PYROSCOPE_SERVER_ADDRESS
  - Check tags and app names in the UI filters
- Grafana datasource missing: provisioning is mounted in docker-compose; restart Grafana container
- Docker daemon issues: start Docker Desktop first
