# Flamegraph Diff Explainer MCP

Name: perf.flamegraph_diff_explainer

Problem:
Engineers struggle to interpret red/green diff quadrants quickly.

Inputs:
```json
{
  "baseline_pprof": ".docs/artifacts/ci/profiles/baseline_cpu.pprof",
  "candidate_pprof": ".docs/artifacts/ci/profiles/candidate_cpu.pprof",
  "min_delta_flat_pct": 1.0,
  "focus_package_prefix": "github.com/org/project/pkg/"
}
```

Algorithm:
1. Run pprof diff (or internal parser) -> symbol deltas
2. Group by package; compute package-level net flat delta
3. Classify top expansion candidates

Output:
```json
{
  "package_summary":[{"pkg":"pkg/processor","delta_flat_pct":+5.1}],
  "regressions":[{"fn":"processBatch","flat_delta_pct":+3.9}],
  "improvements":[{"fn":"encodeEvent","flat_delta_pct":-2.2}],
  "recommendations":["Investigate processBatch; inspect alloc profile too"]
}
```

Extensions:
- Integrate allocation profile correlation
- Provide Speedscope diff link
