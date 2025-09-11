# Tutorial: Golden Signals Aggregator MCP

Objective: Implement `cross.golden_signals_aggregator` to unify latency, traffic, errors, saturation per service.

Steps:
1. Inputs: services[], latency_metric, traffic_metric, error_metric, saturation_metric
2. For each metric template substitute service label if required
3. Execute PromQL queries in parallel (fan-out concurrency limit)
4. Normalize units (latency -> ms, saturation -> fraction)
5. Compute outlier detection (z-score vs peer mean)
6. Build table; annotate services with outlier dimensions

Concurrency Pattern (Go):
```go
sem := make(chan struct{}, 4)
for _, m := range metrics {
  sem <- struct{}{}
  go func(m metricQuery){ defer func(){<-sem}(); execProm(m) }(m)
}
```

Edge Cases:
- Missing metric for a service -> mark null & continue
- High cardinality risk -> enforce max services (e.g., 30)

Output Tips:
- Keep numeric precision limited (round to 2–3 decimals)
- Provide recommendation only for top 1–2 critical outliers

Extensions:
- Add SLO burn integration
- Export HTML summary widget
