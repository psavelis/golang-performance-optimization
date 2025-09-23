# AI-Assisted Continuous Profiling & Optimization Pipeline

This advanced pipeline layer adds an automated “AI-style” analysis phase that fuses:

1. Flamegraph / pprof top outputs (CPU + allocation)
2. Benchmark statistical diffs (benchstat)
3. Hotspot churn (added / removed functions)
4. Rating & recommendations

It produces structured JSON + human Markdown summarizing: improvements, regressions, new hotspots, removed hotspots, benchmark deltas, and actionable optimization guidance.

> Out-of-the-box this repository uses a deterministic heuristic engine (no external API). You can seamlessly upgrade to a real LLM by setting `AI_ANALYSIS_ENDPOINT` to an internal service that returns enriched recommendations.

## Artifacts Generated
Location: `.docs/artifacts/ci/ai_analysis/`

| File | Description |
|------|-------------|
| `report.json` | Machine-consumable structured analysis output |
| `report.md` | Human-readable summary (ratings, deltas, hotspots, actions) |

## Ratings Rubric (Heuristic)
| Dimension | Basis | Notes |
|-----------|-------|-------|
| CPU | Improvement vs regression count | +/-2 adjustments |
| Memory | Same heuristic as CPU | Alloc-space focus |
| Overall | Average of CPU + Memory | Clamped 1–10 |

## Change Classification Logic
| Classification | Rule |
|----------------|-----|
| Improvement | Flat% drop > 5% for same symbol |
| Regression | Flat% increase > 5% |
| New Hotspot | Appears only in optimized/top with Flat% > 1% |
| Removed Hotspot | Present in original, absent in optimized |
| Benchmark Regression | benchstat delta > +2% |

Thresholds are conservative; tune in `cmd/ai-profiler-analyzer/main.go` if your workload is noisier.

## Enabling Real LLM Integration

1. Deploy an internal microservice exposing `POST /analyze` returning JSON like:

   ```json
   { "summary_override": "High-level narrative...", "recommendations": ["...","..."] }
   ```

2. Export `AI_ANALYSIS_ENDPOINT=https://perf-ai.internal/analyze` in the workflow job environment.
3. (Optional) Sign requests with a short-lived token (extend analyzer code where `maybeCallExternal` is invoked).

## Local Trial

Run the existing CI profiling locally, then invoke analysis:

```bash
make ci_profiles
go run ./cmd/ai-profiler-analyzer
cat .docs/artifacts/ci/ai_analysis/report.md
```

## Interpreting Output

Section | How to Use
--------|-----------
CPU/Memory Improvements | Confirm expected optimizations landed.
Regressions | Prioritize largest Flat% or alloc-space increases first.
New Hotspots | Validate they are intentional shifts (e.g., moved work earlier) or unintended complexity.
Benchmark Deltas | Correlate micro-benchmark slowdowns with hotspot symbol changes.
Recommendations | Immediate next actions; feed into a performance backlog.

## Extending Precision

Enhancement | Benefit
------------|--------
Speedscope JSON integration | Rich differential flamegraph UIs.
Statistically significant change detection (Mann–Whitney) | Reduce false positives.
pprof diff quadrants classification | Highlight structural shifts (refactor vs micro-change).
Symbol semantic grouping (packages) | Attribute ownership and team accountability.
JIT tap into tracing spans | Tie hotspot deltas directly to SLO-impacting endpoints.

## Governance Suggestions

Level | Gate Example
------|--------------
Warning | Any regression >5% Flat% or alloc-space.
Soft Fail | Overall rating <5.
Hard Fail | Any single benchmark slowdown >15% or CPU regression >10% Flat%.

Wire these gates in the analyzer (currently it only reports). Add exit codes once signal-to-noise is tuned.

## Roadmap Ideas

1. Export SARIF for IDE inline performance hints.
2. Embed links to Pyroscope filtered views (trace_id + package filter).
3. Auto-open GitHub issues for persistent regressions across 3 consecutive PRs.
4. Incorporate energy/CO₂ estimation using CPU time deltas.

---
This AI-assisted layer elevates profiling from raw data capture to guided optimization cycles—accelerating expert feedback loops while remaining transparent and controllable.
