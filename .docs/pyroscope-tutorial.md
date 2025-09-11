# Continuous Profiling with Grafana Pyroscope (PoC)

This guide shows a production-ready setup for continuous profiling using Grafana Pyroscope with the event generator and database loader applications.

## Overview

Components:
- Pyroscope server (ingestion + UI)
- Grafana with pre-provisioned Pyroscope datasource
- Go services with embedded Pyroscope profiler

Ports:
- Pyroscope: http://localhost:4040
- Grafana: http://localhost:3000 (admin/admin)

## Start the stack

```bash
make start_profiling_stack
```

## Build & run the profiled services

```bash
# Build all binaries
make build

# Generate 100k events with Pyroscope profiling enabled
make run_pyroscope_generator

# Load the generated events into Postgres with profiling
make run_pyroscope_loader
```

You should see two applications in Pyroscope labeled:
- interviewing-golang.generator
- interviewing-golang.loader

## Environment variables

- PYROSCOPE_SERVER_ADDRESS (default: http://localhost:4040)
- PYROSCOPE_APPLICATION_NAME (default set by binary)
- PYROSCOPE_TENANT_ID (optional for multi-tenancy)
- PYROSCOPE_TAGS (comma-separated key=value)

Examples:
```
PYROSCOPE_TAGS=env=staging,region=eu-west-1
```

## Production notes

- Run Pyroscope and Grafana with persistent volumes (configured).
- Tag profiles with env, version, region for filtering.
- Keep Pyroscope and Grafana behind authentication in production.
- Export profiles or scrape remote_write to central systems if needed.

## Troubleshooting

- If profiles don’t show up, verify Pyroscope UI and that the server address is reachable from your host.
- Ensure time sync (NTP) on hosts; large clock skews can affect timelines.
- Check firewall rules between services and Pyroscope.
