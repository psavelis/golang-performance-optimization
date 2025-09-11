# Latency Regression Detector MCP

Name: tracing.latency_regression_detector

Problem:
Post-deploy latency regressions need near-real-time validation vs pre-deploy baseline.

Inputs:
```json
{
  "service": "api",
  "operation": "POST /checkout",
  "deploy_sha": "abc1234",
  "baseline_sha": "def5678",
  "metric": "traces_span_duration_milliseconds_bucket",
  "stat": "p95",
  "lookback_minutes": 20
}
```

Algorithm:
1. Map SHAs -> deployment timestamps via release metadata
2. Define baseline window pre baseline_sha deploy & compare window after deploy_sha
3. Query histogram buckets & compute quantiles
4. Perform relative delta & significance (Welch t or bootstrap) if sample size adequate

Output:
```json
{
  "operation":"POST /checkout",
  "baseline_p95_ms": 142.1,
  "current_p95_ms": 171.4,
  "delta_pct": 20.6,
  "significant": true,
  "samples_baseline": 1824,
  "samples_current": 1962,
  "recommendation": "Rollback or optimize DB write path (validate index usage)."
}
```

Extensions:
- Integrate with regression gate to block promotion
