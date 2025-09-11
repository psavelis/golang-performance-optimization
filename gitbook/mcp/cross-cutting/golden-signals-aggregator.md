# Golden Signals Aggregator MCP

Name: cross.golden_signals_aggregator

Problem:
Distributed dashboards fragment the canonical health view.

Inputs:
```json
{
  "services": ["api","generator","loader"],
  "window_minutes": 10,
  "latency_metric": "histogram_quantile(0.95, sum by (le,service) (rate(http_request_duration_seconds_bucket[5m])))",
  "traffic_metric": "sum(rate(http_requests_total[5m])) by (service)",
  "error_metric": "sum(rate(http_requests_total{code=~'5..'}[5m])) by (service)",
  "saturation_metric": "sum(rate(process_cpu_seconds_total[1m])) by (service)"
}
```

Algorithm:
1. Execute provided PromQL templates per service
2. Normalize & assemble golden signals table
3. Flag outliers (z > 2) per signal

Output:
```json
{
  "services":[
    {"service":"api","latency_p95_ms":181,"errors_per_s":0.7,"cpu_pct":0.74,"outliers":["latency"]}
  ],
  "recommendations":["Investigate api latency outlier vs peer mean"]
}
```

Extensions:
- Add SLO burn integration
