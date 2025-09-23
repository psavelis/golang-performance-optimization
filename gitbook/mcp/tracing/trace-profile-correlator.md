# Trace ↔ Profile Correlator MCP

Name: tracing.trace_profile_correlator

Problem:
Bridging per-request latency and aggregated CPU hotspots historically manual.

Inputs:
```json
{
  "trace_id": "4f8cc5b9d3a1f1e2",
  "service": "generator-otel",
  "profile_type": "cpu",
  "profile_range_minutes": 10,
  "min_flat_pct": 1.0
}
```

Algorithm:
1. Fetch trace & extract span timing + attributes
2. Query Pyroscope for overlapping window filtered by trace_id tag (if present) or service
3. Map top symbols to span phases (overlap heuristic)
4. Output symbol -> span correlation + potential optimization path

Output:
```json
{
  "trace_id":"4f8cc5b9d3a1f1e2",
  "spans": 143,
  "hotspots": [
    {"fn":"processBatch","flat_pct":22.9,"likely_span":"generator.process"}
  ],
  "coverage_pct": 41.2,
  "recommendation": "Investigate processBatch allocation spikes; link to span attributes size=batch_size"}
```

Extensions:
- Add per-symbol latency attribution weight
- Provide deep link to Pyroscope diff view using span boundary
