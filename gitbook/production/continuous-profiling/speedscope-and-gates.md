# Speedscope Visualization & Enforced Gates

This page explains the enforced performance gates and interactive flamegraph viewing using Speedscope artifacts.

## Gates (CI Enforced)
| Gate | Env Var | Threshold | Action |
|------|---------|-----------|--------|
| CPU Regression | AI_GATE_CPU_REG_MAX | 8% | Fail job if any symbol exceeds |
| Memory Regression | AI_GATE_MEM_REG_MAX | 10% | Fail job |
| Benchmark Slowdown | AI_GATE_BENCH_SLOW_MAX | 5% | Fail job |
| Overall Rating | AI_GATE_MIN_RATING | 5 | Fail job |

## Failing Example
If a regression occurs you'll see job failure with lines like:
```
Gate failure:
CPU regression 12.40% > 8.00%: main.process: 4.00% -> 16.40% (12.40% regression)
```

## Local Dry Run (No Fail)
```
make ci_profiles
go run ./cmd/ai-profiler-analyzer
```

## Forcing Gate Evaluation Locally
```
AI_GATE_ENABLE=true AI_GATE_CPU_REG_MAX=8 AI_GATE_MEM_REG_MAX=10 \
AI_GATE_BENCH_SLOW_MAX=5 AI_GATE_MIN_RATING=5 \
go run ./cmd/ai-profiler-analyzer
```
Exit code 2 indicates failure.

## Speedscope Artifacts
Generated in CI at:
```
.docs/artifacts/ci/speedscope/*.speedscope.json
```
Open via https://www.speedscope.app/ (File → Browse) for interactive zoom & differential inspection.

## Adding More Profiles
Extend Makefile target `ci_profiles` with additional `pprof -proto` conversions then run speedscope on each.

## Tuning Guidance
| Scenario | Suggested Change |
|----------|------------------|
| Frequent false positives | Raise CPU/MEM thresholds by 2–3% |
| Missed small regressions | Lower bench slowdown to 3% |
| Overhead concern | Reduce sample size (fewer events) but keep ratio comparisons |

---
Speedscope plus objective gates accelerates credible performance review—issues are caught early, explored visually, and triaged with shared quantitative standards.