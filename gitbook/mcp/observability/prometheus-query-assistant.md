# Prometheus Query Assistant MCP

Name: prometheus.query_assistant

Problem:
Teams waste cycles iterating raw PromQL. This MCP synthesizes validated queries + immediate numeric context.

Inputs (JSON):
```json
{
  "expression": "rate(http_requests_total{service='api'}[5m])",
  "range_minutes": 30,
  "step_seconds": 15,
  "expand": ["_sum", "_p95"],
  "optimize": true
}
```

Behavior:
1. Validates expression via Prometheus /api/v1/series match
2. Executes range query
3. If optimize=true, suggests cardinality reduction (drops low-variance labels)
4. If expand contains suffix tokens (e.g. _p95) generates derived quantile queries when histograms exist

Outputs:
```json
{
  "original": "rate(http_requests_total{service='api'}[5m])",
  "optimized": "sum by (status) (rate(http_requests_total{service='api'}[5m]))",
  "series_count": 24,
  "samples_total": 2880,
  "stats": {"min":12.4, "max":18.9, "p95":17.7, "avg":15.2},
  "cardinality": {"before":6, "after":3},
  "diagnostics": ["High label spread: handler (12)", "Consider rollup window 10m for smoother p95"],
  "expansions": [
     {"kind":"p95","query":"histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{service='api'}[5m])))","value":0.182}
  ]
}
```

Failure Modes:
- 400 invalid expression -> return {"error_code":"VALIDATION"}
- Empty series -> still success with stats empty
- Timeout -> partial=false, suggestion shorten window

Security:
- Needs read-only Prometheus HTTP token; deny any label regex outside sanctioned prefixes.

Example Invocation:
```bash
mcp invoke prometheus.query_assistant payload.json
```

Extensions:
- Add anomaly detection (EWMA band breach)
- Provide cost hint (sample density * retention)
