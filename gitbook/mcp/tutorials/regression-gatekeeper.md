# Tutorial: Implementing a Regression Gatekeeper MCP

Objective: Enforce performance budgets via automated benchmark diff classification.

Steps:
1. Inputs: baseline_report, candidate_report, thresholds{metric->pct}, hard_fail_pct, min_sample
2. Load JSON benchmark diffs (ns/op, allocs/op, B/op)
3. For each case:
   - If samples < min_sample: mark neutral
   - delta_pct = (candidate - baseline) / baseline * 100
   - Status: fail if delta_pct > hard_fail_pct; warn if > threshold; pass otherwise
4. Aggregate summary & decision (fail > warn > pass)
5. Emit JSON for CI + markdown snippet for PR comment

Go Pseudocode:
```go
for _, c := range cases {
  if c.Baseline.Samples < minSample { c.Status = Neutral; continue }
  pct := (c.Candidate.Value - c.Baseline.Value)/c.Baseline.Value*100
  switch {
    case pct > hardFail: c.Status = Fail
    case pct > warn: c.Status = Warn
    default: c.Status = Pass
  }
}
```

Edge Cases:
- Zero baseline value -> skip or treat as neutral
- Negative deltas (improvements) -> record improvement_count

Extensions:
- Add statistical significance (Welch t-test) gate
- Export SARIF for inline IDE surfacing
