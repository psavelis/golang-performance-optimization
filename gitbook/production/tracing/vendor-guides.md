# Vendor Guides (Datadog, New Relic, Splunk, Dynatrace)

This guide explains how to send traces from the local OTel Collector to popular vendors. Choose one vendor and enable its exporter in `env/otel-collector/config.yaml`. For Dynatrace, use the dedicated override flow.

## Datadog

1. Set environment variables:
   - `DD_API_KEY` (required)
   - `DD_SITE` (e.g., datadoghq.com, datadoghq.eu)
2. Uncomment the `datadog` exporter in the Collector config and add `datadog` to the `traces` pipeline exporters.
3. Restart the Collector.
4. Ensure services export to the Collector (defaults in Makefile already do).

## New Relic

1. Set environment variables:
   - `NEW_RELIC_API_KEY` (required)
   - `NEW_RELIC_REGION` (US or EU)
2. Uncomment the `newrelic` exporter and add it to the `traces` pipeline exporters.
3. Restart the Collector.

## Splunk (HEC)

1. Set environment variables:
   - `SPLUNK_HEC_TOKEN`, `SPLUNK_HEC_ENDPOINT` (e.g., https://http-inputs-<realm>.splunkcloud.com/services/collector)
   - Optional: `SPLUNK_INDEX`
2. Uncomment the `splunk_hec` exporter and add it to the `traces` pipeline exporters.
3. Restart the Collector.

## Dynatrace (OTLP HTTP)

Use the dedicated guide for an isolated override flow and environment variables:

- [Dynatrace Integration (OTLP)](dynatrace-integration.md)

## Service Metadata

The applications set resource attributes using standard environment variables:
- `OTEL_SERVICE_NAME` (service.name)
- `OTEL_SERVICE_VERSION` (service.version)
- `OTEL_ENV` (deployment.environment)

These are recognized by major vendors for service maps, filtering, and dashboards.
