# SLO Error Budget Monitor MCP

Name: cross.slo_error_budget_monitor

Problem:
Manual spreadsheet SLO burn tracking delays intervention.

Inputs:
```json
{
  "slo_target": 99.9,
  "window_hours": 24,
  "error_metric": "rate(http_requests_total{code=~'5..'}[5m])",
  "total_metric": "rate(http_requests_total[5m])",
  "history_days": 7
}
```

Algorithm:
1. Compute error ratio over window
2. Project burn rate vs remaining budget
3. Classify risk tiers (normal / watch / urgent)

Output:
```json
{
  "target":99.9,
  "current_error_ratio":0.0007,
  "current_availability":99.93,
  "remaining_budget_pct": 63.2,
  "burn_rate": 2.1,
  "risk":"watch",
  "time_to_exhaust_hours": 19.4,
  "recommendations":["Throttle write endpoints if burn >3 for 2 more hours"]
}
```

Extensions:
- Multi-window multi-burn gating (4h/30d)
