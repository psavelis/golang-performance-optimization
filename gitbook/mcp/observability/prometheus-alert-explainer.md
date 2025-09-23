# Prometheus Alert Explainer MCP

Name: prometheus.alert_explainer

Problem:
On-call load increases when alerts lack immediate causal & impact context.

Inputs:
```json
{
  "alert_name": "HighErrorRate",
  "fingerprint": "ab12cd34",
  "lookback_minutes": 60,
  "correlate_labels": ["service", "cluster"],
  "include_runs": 3
}
```

Algorithm:
1. Fetch active alert via Alertmanager API
2. Pull firing expression & evaluate for lookback range
3. Compute deviation vs last include_runs resolved windows
4. Correlate with related alerts sharing correlate_labels
5. Surface top contributing label combinations (Pareto 80%)

Output:
```json
{
  "alert": {"name":"HighErrorRate","state":"firing","severity":"page"},
  "expression": "increase(http_requests_total{code=~'5..'}[5m]) / increase(http_requests_total[5m]) > 0.02",
  "current_value": 0.037,
  "baseline_p95": 0.011,
  "deviation_ratio": 3.36,
  "correlated_alerts": ["LatencySLODegradation"],
  "top_contributors": [
    {"service":"checkout","cluster":"prod-a","error_ratio":0.061,"share":0.44},
    {"service":"cart","cluster":"prod-a","error_ratio":0.049,"share":0.31}
  ],
  "impact_assessment": "Approx 3.4x baseline. Two services drive 75% of errors.",
  "recommended_actions": [
    "Check recent deploys for checkout,cart in prod-a",
    "Examine trace error spans correlated to 5xx",
    "Enable temporary adaptive throttling if saturation rises"
  ]
}
```

Failure Modes:
- Alert not found -> 404 style output with nulls
- Expression no longer valid -> mark stale=true

Security:
- Alertmanager read token; optional redaction for labels containing pii=true annotation

Extensions:
- Pull recent code changes (Git provider) for owning teams
- Predict time-to-breach SLO based on derivative
