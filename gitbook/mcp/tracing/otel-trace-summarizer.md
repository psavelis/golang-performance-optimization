# OTel Trace Summarizer MCP

Name: tracing.trace_summarizer

Problem:
Engineers manually click expansions in UIs to derive bottlenecks & error context.

Inputs:
```json
{
  "trace_id": "4f8cc5b9d3a1f1e2",
  "include_events": true,
  "max_spans": 200,
  "collapse_internal": true
}
```

Algorithm:
1. Fetch trace from Tempo/Jaeger via API
2. Build span tree; compute inclusive/exclusive durations
3. Identify top latency contributors (exclusive time)
4. Surface error spans & logs
5. Generate high-level narrative

Output:
```json
{
  "trace_id":"4f8cc5b9d3a1f1e2",
  "spans_total": 143,
  "services": {"api":71,"db":18,"cache":12},
  "critical_path_ms": 212.4,
  "top_exclusive":[{"span":"db SELECT products","exclusive_ms":61.3,"pct":28.8}],
  "errors":[{"span":"cache PUT","status":"ERROR","message":"timeout"}],
  "narrative":"DB query dominates 29% of wall time; cache timeouts degrade latency tail."}
```

Extensions:
- Attach Pyroscope deep link tags if available
- Derive span-level CPU vs wall variance (requires profiling tags)
