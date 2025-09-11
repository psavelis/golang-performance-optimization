# Local Setup (OTel Collector)

The repository includes an OpenTelemetry Collector composed alongside Prometheus, Grafana, Pyroscope, and Postgres.

Quick start:

```bash
make start_observability_stack
make build
make run_otel_generator
make run_otel_loader
```

Endpoints:
- OTLP gRPC: localhost:4317
- OTLP HTTP: http://localhost:4318

Environment variables used:
- OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
- OTEL_EXPORTER_OTLP_INSECURE=true
- OTEL_SERVICE_NAME=interviewing-golang.<component>-otel
- OTEL_SERVICE_VERSION (optional)
- OTEL_ENV (optional)

Collector config: `env/otel-collector/config.yaml`
- Default exporter is `logging` so you can see spans in the collector logs.
- Enable vendor exporters by uncommenting and providing credentials via environment variables.

Generated test data is stored under `.docs/artifacts/test-data/` to avoid polluting the repository root.
