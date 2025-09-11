# Benchmark Trend Analyzer MCP

Name: perf.benchmark_trend_analyzer

Problem:
Single PR comparisons miss gradual regressions accumulating over weeks.

Inputs:
```json
{
  "bench_history_path": ".ci/bench/history/*.json",
  "min_points": 8,
  "slope_window": 5,
  "alpha": 0.05,
  "metrics": ["ns_per_op","allocs_per_op"],
  "seasonality_days": 7
}
```

Algorithm:
1. Load historical benchmark JSON snapshots
2. For each test+metric compute rolling slope (OLS) over slope_window
3. Perform significance test vs 0 slope
4. Flag monotonic regressions or volatility spikes

Output:
```json
{
  "tests": [
    {"name":"BenchmarkProcess/size=1k","ns_per_op":{"trend":"regression","slope":+182.4,"pct_change":12.1}},
    {"name":"BenchmarkProcess/size=1k","allocs_per_op":{"trend":"stable"}}
  ],
  "regressions":1,
  "improvements":0,
  "recommendations":["Investigate code changes last 5 commits affecting batch serialization"]
}
```

Extensions:
- Add Holt-Winters seasonal adjustment
- Export sparkline SVGs per test
