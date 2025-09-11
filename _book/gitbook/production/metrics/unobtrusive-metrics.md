# Unobtrusive Prometheus Metrics for Short-Lived Jobs

This guide shows how to collect Prometheus metrics from CLI-style workloads without invasive code changes. We use a tiny env-driven helper that exposes `/metrics` when enabled.

## Why unobtrusive?
- Minimal change: a single call at startup; optional hold after work to let Prometheus scrape.
- Reusable: shared helper provides Go/process collectors and HTTP handler.
- Safe defaults: disabled unless explicitly enabled via env.

## Enabling metrics

Set environment variables when running a job:

- `METRICS_ENABLED=true` to turn metrics on (or set `METRICS_ADDR` directly)
- `METRICS_ADDR=:2112` to bind the server (defaults to `:2112`)
- `METRICS_PATH=/metrics` to change the scrape path (defaults to `/metrics`)
- `METRICS_HOLD_FOR=60s` to keep the process alive so Prometheus can scrape

Example (generator):

```bash
METRICS_ENABLED=true METRICS_ADDR=:2112 METRICS_HOLD_FOR=60s ./generator 100000 .docs/artifacts/test-data/test_prom.json
```

Prometheus scrape config (already included):

```yaml
scrape_configs:
  - job_name: 'generator'
    static_configs:
      - targets: ['host.docker.internal:2112']
  - job_name: 'loader'
    static_configs:
      - targets: ['host.docker.internal:2113']
```

## How it works

- `StartFromEnv()` reads env and starts an HTTP server with `/metrics` using the standard registry and Go/process collectors.
- `HoldFromEnv()` optionally sleeps for the configured duration to keep ephemeral jobs observable.
- If disabled, both functions are no-ops.

## Pros and cons

Pros:
- Very low code footprint; easy to retrofit into CLIs and batch jobs
- Works across binaries via a shared helper
- Uses standard Prometheus tooling; zero vendor lock-in

Cons:
- Ephemeral processes need a hold window or a pull-friendly sidecar to be scraped
- Exposes a local HTTP port; ensure proper network policy in production
- Registry is shared in-process; avoid duplicate registration when mixing frameworks

## Where it's wired

- Base generator and loader
- Optimized generator (and can be added to optimized loader similarly)

These binaries start the metrics endpoint only when env is set, preserving default behavior.
