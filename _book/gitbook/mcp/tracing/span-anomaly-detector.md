# Span Anomaly Detector MCP

Name: tracing.span_anomaly_detector

Problem:
Latency shifts hide until user SLOs degrade; early detection at span granularity reduces blast radius.

Inputs:
```json
{
  "service": "api",
  "operation": "GET /orders",
  "window_minutes": 30,
  "baseline_hours": 6,
  "method": "zscore",
  "z_threshold": 3.0
}
```

Algorithm:
1. Pull aggregated span duration metrics (ex: otelcol spanmetrics or PromQL)
2. Build baseline distribution over baseline_hours
3. Compute z-score for recent window p95
4. Flag anomaly if |z| >= threshold

Output:
```json
{
  "service":"api","operation":"GET /orders",
  "p95_current_ms": 183.2,
  "p95_baseline_ms": 121.7,
  "z_score": 3.41,
  "anomalous": true,
  "contributors":[{"label":"region=us-east","delta_ms":47.8}],
  "recommendation":"Investigate regional DB latency; consider routing shift."}
```

Extensions:
- Replace z-score with seasonal hybrid ESD
- Add error-rate co-correlation gating
