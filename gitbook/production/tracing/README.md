# Distributed Tracing with OpenTelemetry

This section covers instrumenting Go services with OpenTelemetry and exporting traces to the OTel Collector, compatible with vendors like Datadog, New Relic, and Splunk.

- Binaries: `generator-otel`, `loader-otel`
- Local Collector: OTLP gRPC on 4317, OTLP HTTP on 4318
- Make targets: `make start_observability_stack`, `make run_otel_generator`, `make run_otel_loader`

## Guides

- Local setup: [OTel Collector](local-setup.md)
- Vendor guides: [Datadog, New Relic, Splunk](vendor-guides.md)
- Vendor guide: [Dynatrace Integration (OTLP)](dynatrace-integration.md)

Notes:
- Dynatrace uses an OTLP HTTP exporter with an explicit compose override and environment variables, keeping the base stack vendor-neutral.
- All apps export traces to the local Collector; the Collector handles fan-out to backends (Tempo and optionally Dynatrace).
