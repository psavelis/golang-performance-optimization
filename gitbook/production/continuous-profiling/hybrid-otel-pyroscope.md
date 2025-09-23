# Hybrid OpenTelemetry + Pyroscope Profiling

This guide shows how to run a binary with BOTH:

- OpenTelemetry tracing (exported to Tempo / any OTLP backend)
- Continuous profiling via Grafana Pyroscope
- Optional trace ↔ profile correlation tags (trace_id, root_span_id)

## When To Use
Use the hybrid setup when you need to answer both:

1. "What distributed request / span is slow?" (tracing)
2. "Why is it slow at the code level?" (profiling)

## Binaries Supporting Hybrid Mode
- `cmd/generator-otel` (now: optional Pyroscope)
- `cmd/loader-otel` (now: optional Pyroscope)

Enable by environment flags instead of separate builds.

## Environment Flags
| Variable | Purpose | Default |
|----------|---------|---------|
| PYROSCOPE_ENABLE | Turns on in-process profiler | false |
| PYROSCOPE_TRACE_CORRELATION | Adds trace_id/root_span_id tags | false |
| PYROSCOPE_SERVER_ADDRESS | Pyroscope server URL | http://localhost:4040 |
| PYROSCOPE_APPLICATION_NAME | Override app name | <binary-derived> |
| PYROSCOPE_TAGS | Extra static tags (k=v,comma) | service=<auto> |
| OTEL_EXPORTER_OTLP_ENDPOINT | OTLP GRPC endpoint | localhost:4317 |
| OTEL_RESOURCE_ATTRIBUTES | service.name / env / etc | unset |

## Minimal Local Run
```bash
PYROSCOPE_ENABLE=true \
PYROSCOPE_TRACE_CORRELATION=true \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
OTEL_RESOURCE_ATTRIBUTES="service.name=generator-otel,env=dev" \
go run ./cmd/generator-otel 5000 /tmp/out.json
```
Then load them:
```bash
PYROSCOPE_ENABLE=true \
PYROSCOPE_TRACE_CORRELATION=true \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
OTEL_RESOURCE_ATTRIBUTES="service.name=loader-otel,env=dev" \
go run ./cmd/loader-otel postgres://user:pass@localhost:5432/db?sslmode=disable /tmp/out.json
```

## Correlation Mechanics
1. Root span started early in execution.
2. If correlation enabled and span context valid, profiler tags mutated to include:
   - trace_id=<hex>
   - root_span_id=<hex>
3. In Pyroscope you can filter: `trace_id=abcd1234...`
4. In Tempo you can search for the trace_id and manually pivot back to Pyroscope.

## Building Deep Links (Manual)
Pyroscope UI supports query params. Example pattern:

```text
http://localhost:4040/?query=service%3Dgenerator-otel%20trace_id%3D<TRACE>
```
Tempo trace URL (varies by stack) often ends with `/trace/<TRACE>`. Store that ID.

## CI Integration Suggestions
- Keep profiling disabled in fast unit jobs (leave PYROSCOPE_ENABLE unset).
- Enable only in a dedicated performance or profiling job.
- Optionally add a matrix job with and without correlation to compare tag cardinality impact.

## Operational Guardrails
| Concern | Mitigation |
|---------|------------|
| Tag Cardinality | Only root IDs added once; avoid per-request mutation. |
| Overhead | Pyroscope sampling is low; still isolate perf-sensitive benchmarks without it. |
| Trace Absence | If no valid span context, tags omitted gracefully. |
| Multi-Tenancy | Use PYROSCOPE_TENANT_ID for multi-team clusters. |

## Troubleshooting
| Symptom | Cause | Fix |
|---------|-------|-----|
| No profiles in UI | Wrong PYROSCOPE_SERVER_ADDRESS | Verify port 4040 reachable |
| No trace_id tag | PYROSCOPE_TRACE_CORRELATION not true OR span invalid | Ensure flag + OTel init earlier |
| High tag cardinality warning | Added dynamic IDs elsewhere | Remove per-request tag logic |
| Build fails on pyroscope-go import | Module not downloaded | `go mod tidy` or ensure network access |

## Next Evolution
- Auto-inject trace links into logs (add trace_id field in structured logger).
- Emit span event containing Pyroscope query URL for easy pivot.
- Add speedscope JSON artifact export in CI for flamegraph diffing UX.

## Quick Validation Checklist
- Run generator-otel with both flags → confirm profiles appear tagged with trace_id.
- Search Tempo for same trace_id → validate time range overlaps workload window.
- Run without PYROSCOPE_TRACE_CORRELATION → tags disappear (expected).

---
Hybrid mode gives staff+ engineers a single workflow to move from high-level distributed latency to CPU/object allocation root causes without context switching across differently built binaries.
