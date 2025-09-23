# Pyroscope Diff Analyzer MCP

Name: pyroscope.diff_analyzer

Problem:
Comparing two time windows manually wastes time; need structured symbol delta classification.

Inputs:
```json
{
  "service": "generator-otel",
  "profile_type": "cpu",
  "baseline_range": {"from":"-30m","to":"-20m"},
  "compare_range": {"from":"-10m","to":"now"},
  "min_abs_flat_pct": 0.5,
  "min_delta_pct": 2.0
}
```

Algorithm:
1. Fetch and diff flamegraphs (Pyroscope diff endpoint or local pprof diff)
2. Build map: symbol -> {baseline_flat, compare_flat}
3. Classify improvements, regressions, new, removed based on thresholds

Output:
```json
{
  "service":"generator-otel",
  "profile_type":"cpu",
  "regressions":[{"fn":"compressPayload","delta_flat_pct":+4.1,"from":3.2,"to":7.3}],
  "improvements":[{"fn":"marshalEvent","delta_flat_pct":-3.5,"from":5.1,"to":1.6}],
  "new":[{"fn":"optimizeBatch","to":2.2}],
  "removed":[{"fn":"legacyPath","from":1.1}],
  "summary":{"regression_count":1,"improvement_count":1},
  "rating":7,
  "recommendations":["Investigate compressPayload changes (recent commit?)"]
}
```

Extensions:
- Include cumulative deltas
- Provide Speedscope diff artifact
