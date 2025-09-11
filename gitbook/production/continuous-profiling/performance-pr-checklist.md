# Performance PR Checklist

Use this standardized checklist (mirrors pull request template) to ensure each performance-related change is measurable, reversible, and governed.

## Required Evidence
| Evidence | Description | Tool/Source |
|----------|-------------|-------------|
| Baseline Profile | Pre-change CPU & alloc_space | `make ci_profiles` before change (stash) |
| Post-change Profile | After modifications | `make ci_profiles` |
| Hotspot Diff | pprof diff + AI report | AI analyzer output |
| Benchmark Diff | benchstat delta | Bench diff artifacts |
| Trace ↔ Profile Link | Span event `pyroscope.link` present | Tempo span event |
| Version Metadata | build.git_commit + build.version attributes | OTel resource / span attrs |

## Pass / Fail Gates (Recommended)
Gate | Threshold | Rationale
-----|-----------|----------
CPU Regression | < +8% flat per symbol | Protects top-line latency contributors
Alloc Regression | < +10% flat alloc_space | Cost + GC pressure control
Benchmark Slowdown | < +5% time/op | Sustains micro-perf health
Overall Rating | ≥ 5/10 | Ensures net win or neutral

## Workflow
1. Run baseline profiles on main (or fetch saved artifact).
2. Implement change.
3. Run profiling + analyzer locally: `make ci_profiles && make ai_analysis_local`.
4. Inspect `report.md` for regressions.
5. Adjust code / add benchmarks until gates satisfied.
6. Open PR; attach links & artifact excerpts.
7. If gates fail in CI, iterate or request waiver with justification.

## Waiver Criteria
Accept a temporary regression only if:
- Enables a larger architectural simplification delivering net future gain.
- Unblocks critical feature under a time constraint (record in risk log).
- Regression isolated behind a feature flag with rollback path.

## Common Pitfalls
Pitfall | Avoidance
--------|----------
Comparing noisy single-run profiles | Increase sample size or repeat runs
Ignoring allocation spikes | Monitor alloc_space even if CPU improves
Missing version linkage | Always build via Makefile (ldflags applied)
Trace event not found | Ensure correlation env flags set

## Example PR Summary (Template)
```
Intent: Reduce allocations in event generation path.
Change: Reuse internal buffers, switch JSON encoder.
Result: -18% alloc_space in top hotspot, -6% CPU flat in generate_batch.
Gates: Pass (overall rating 7/10).
Trace Link: https://tempo.local/trace/abcd1234 (Pyroscope link in span event).
```

---
Following this checklist institutionalizes a culture of measurable, reversible, and continuously improving performance.