# Regression Gates & Advanced Workflow Automation

This guide layers deterministic governance on top of profiling + AI-assisted analysis. It converts insights into **objective pass/fail signals** for CI.

## Gate Types
| Gate | Purpose | Typical Threshold |
|------|---------|-------------------|
| CPU Hotspot Regression | Prevent large flat% increases | > +8% flat delta |
| Allocation Regression | Control alloc_space growth | > +10% flat delta |
| Benchmark Slowdown | Maintain micro performance | > +5% time/op delta |
| Overall Rating Fail | Enforce quality bar | rating overall < 5 |
| New Hotspot Alert (soft) | Detect churn | new hotspot > 4% flat |

## Enabling Hard Gates
Set environment variables in the analysis job:
```
AI_GATE_ENABLE=true
AI_GATE_CPU_REG_MAX=8
AI_GATE_MEM_REG_MAX=10
AI_GATE_BENCH_SLOW_MAX=5
AI_GATE_MIN_RATING=5
```
(Analyzer will be extended to exit non-zero when violated.)

## Suggested PR Flow
1. Dev pushes code → profiling + bench jobs run.
2. AI analyzer produces structured report.
3. Gates evaluate JSON → pass/fail.
4. PR comment includes summarised reasons for any failure.
5. Optional: label PR with `performance-regression` automatically (GitHub Action).

## Speedscope Integration (Optional)
Add conversion step:
```
pprof -sample_index=cpu -proto generator_cpu.prof > cpu.pb
npx speedscope cpu.pb  # or use a converter library
```
Then upload `cpu.speedscope.json` artifact.

## Automation Tips
| Enhancement | Benefit |
|-------------|---------|
| Cache baseline profiles | Lower noise; compute diff vs stable artifact |
| Nightly trend job | Track historical rating & hotspots | 
| Slack webhook notification | Real-time visibility |
| Issue auto-open | Forces accountability for persistent failures |

## Tuning Strategy
- Start with soft gates (warnings) for 1–2 weeks.
- Record false positives; adjust thresholds or sampling sizes.
- Promote to hard gates once stable.

## Example JSON Snippet (report.json)
```json
{
  "cpu_improvements": ["runtime.mallocgc: 12.5% -> 7.1% (5.4% improvement)"],
  "cpu_regressions": ["main.process: 8.0% -> 16.4% (8.4% regression)"],
  "rating": {"cpu":7,"memory":5,"overall":6},
  "summary": "Overall net improvement..."
}
```
Use `jq` in a gate step:
```
CPU_REG_MAX=${AI_GATE_CPU_REG_MAX:-8}
COUNT_REG=$(jq '.cpu_regressions | length' report.json)
[ "$COUNT_REG" -gt 0 ] && echo "CPU regressions detected: $COUNT_REG" || true
```

## Future Roadmap
- Span ↔ hotspot correlation for endpoint-level gating.
- Automated root cause hypothesis generation.
- Performance budget dashboards (Grafana + JSON data source).

---
These practices turn profiling from raw data into a **governed performance lifecycle**, preventing regressions while enabling rapid, verifiable optimizations.
