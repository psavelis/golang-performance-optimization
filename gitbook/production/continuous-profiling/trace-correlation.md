# Trace ↔ Profile Correlation

This tutorial shows how to correlate OpenTelemetry trace spans with Pyroscope continuous profiles by injecting the root trace and span identifiers into Pyroscope tags.

## Why Correlate?

When a latency spike appears in tracing (Tempo), you want to jump directly into the CPU or allocation profile that was active during that span. By adding `trace_id` (and an optional `root_span_id`) tags to Pyroscope profiles you gain:

- Direct filtering in Pyroscope for a problematic trace
- Time-aligned hotspot analysis for that execution
- Unified incident narrative (single IDs across logs, traces, profiles)

## Enabling Correlation

Set the environment variable before running the Pyroscope-enabled binaries:

```bash
PYROSCOPE_TRACE_CORRELATION=true PYROSCOPE_SERVER_ADDRESS=http://localhost:4040 \
  PYROSCOPE_TAGS="env=dev,service=generator" \
  ./generator-grafana-pyroscope 50000 out.json
```

When enabled the binary:
1. Starts a lightweight root span (`generator.run` or `loader.run`)
2. Extracts `trace_id` and `span_id`
3. Appends them to `PYROSCOPE_TAGS` (once) as `trace_id=<id>,root_span_id=<id>`
4. Profiles uploaded to Pyroscope now carry those IDs

You’ll see the enhanced tags in Pyroscope’s label selector.

## Querying in Pyroscope

1. Open Pyroscope UI → Select the application name (e.g. `interviewing-golang.generator`)
2. Add a tag filter: `trace_id=<value>`
3. Compare flamegraph of that trace vs baseline (remove `trace_id` filter)

## Trace to Profile Workflow

1. In Grafana Tempo, locate a slow trace
2. Copy its `Trace ID`
3. Paste as a tag filter in Pyroscope
4. Analyze top frames unique to that trace

## Overhead Considerations

- Only a single root span is created (no per-event spans) to keep correlation overhead negligible.
- Tag cardinality impact: one additional value per execution. Safe for short-lived batch jobs and low-frequency tasks.
- Disable by unsetting `PYROSCOPE_TRACE_CORRELATION`.

## Advanced: Enforcing Tag Structure

If you already use `PYROSCOPE_TAGS`, the correlation logic appends missing keys without overwriting existing ones. Example final tag string:

```
env=dev,service=generator,version=1.0.0,trace_id=abcd1234...,root_span_id=ef567890...
```

## Future Extensions

| Goal | Approach |
|------|----------|
| Span-level profiling windows | Embed span timestamps; query narrower ranges |
| Multi-span correlation | Export span events referencing profile slice IDs |
| UI deep link | Construct Grafana Tempo URL from `trace_id` & add link in docs |

Correlating traces and profiles shortens mean time to insight by combining structural (trace) and cost (profile) perspectives of the same execution.
