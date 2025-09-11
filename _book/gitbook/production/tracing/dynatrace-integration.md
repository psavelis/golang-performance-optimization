# Dynatrace Integration (OTLP)

This guide shows how to export OpenTelemetry traces to Dynatrace from the local stack, alongside Tempo.

## Prerequisites

- Dynatrace environment with API access
- Two values:
  - DT_OTLP_ENDPOINT: e.g., `https://{your-env-id}.live.dynatrace.com/api/v2/otlp`
  - DT_API_TOKEN: API token with `ingest` scopes

## Configure the Collector

We provide an alternate Collector config that adds an OTLP HTTP exporter for Dynatrace and keeps local Tempo:

- `env/otel-collector/config.dynatrace.yaml`

Key exporter settings:

```yaml
exporters:
  otlphttp/dynatrace:
    endpoint: "${DT_OTLP_ENDPOINT}"
    headers:
      Authorization: "Api-Token ${DT_API_TOKEN}"
    compression: gzip
    timeout: 10s
```

## Start the stack with Dynatrace enabled

Set environment variables and use the compose override:

```bash
export DT_OTLP_ENDPOINT="https://YOUR_ENV.live.dynatrace.com/api/v2/otlp"
export DT_API_TOKEN="<redacted>"
make start_observability_stack_dynatrace
```

This runs the same stack (Postgres, Pyroscope, Prometheus, Grafana, Tempo) with the Collector exporting to Dynatrace as well.

## Run instrumented apps

Use the existing OTel binaries; they export to the local Collector which fans out to Tempo and Dynatrace:

```bash
make run_otel_generator_dynatrace
make run_otel_loader_dynatrace
```

## Validation

- Collector logs will show outgoing requests to the Dynatrace endpoint (and still log/export to Tempo).
- In Dynatrace, check Traces/Services to see spans with your service names (e.g., `interviewing-golang.generator-otel`).

## Notes & troubleshooting

- Ensure your token has correct scopes for OTLP ingest.
- Corporate proxies/firewalls may block outbound HTTPS to Dynatrace; configure networking as needed.
- You can revert to the standard config by using `make start_observability_stack`.
- This approach keeps vendor config isolated to an override file and environment variables, matching our professional standards.
