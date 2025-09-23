# Regression Gatekeeper MCP

Name: perf.regression_gatekeeper

Problem:
Performance regressions slip through without structured gating.

Inputs:
```json
{
  "baseline_report": ".docs/artifacts/ci/benchdiff/baseline.json",
  "candidate_report": ".docs/artifacts/ci/benchdiff/candidate.json",
  "thresholds": {"ns_per_op_pct": 5.0, "allocs_per_op_pct": 5.0},
  "hard_fail_pct": 12.0,
  "min_sample": 5
}
```

Algorithm:
1. Load benchstat diff outputs
2. For each metric compute delta%; classify pass/warn/fail
3. Aggregate severity & produce decision

Output:
```json
{
  "cases": [
    {"name":"BenchmarkProcess/size=1k","metric":"ns_per_op","delta_pct":6.2,"status":"warn"}
  ],
  "decision":"warn",
  "summary":{"warn":1,"fail":0},
  "recommendation":"Review hotspot diff before merging."}
```

Extensions:
- Integrate statistical significance per case
- Emit SARIF for PR annotations
