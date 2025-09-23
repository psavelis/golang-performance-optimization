# Cost Efficiency Analyzer MCP

Name: cross.cost_efficiency_analyzer

Problem:
Costs rise faster than delivered capacity; need normalized efficiency KPIs.

Inputs:
```json
{
  "cost_export_path": "finops/costs.csv",
  "throughput_metric": "rate(http_requests_total[5m])",
  "range_hours": 24,
  "normalize_by": "region"
}
```

Algorithm:
1. Load cost allocation (CSV or API)
2. Aggregate throughput per dimension
3. Compute cost_per_1k_requests & trend vs prior day

Output:
```json
{
  "regions":[{"region":"us-east","cost_per_1k":0.042,"delta_pct":+8.1}],
  "overall_delta_pct": 5.4,
  "recommendations":["Investigate autoscale min replicas in us-east"]
}
```

Extensions:
- Blend with latency SLO to compute cost/performance frontier
