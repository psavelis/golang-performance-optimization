# Continuous Optimization Playbook (Senior / Staff)

A pragmatic lifecycle for sustaining performance excellence.

## 1. Hypothesis Intake
Source | Example | Action
-------|---------|-------
SLO Alert | p95 latency ↑ 20% | Trace sample + hotspot tag filter
Bench Diff | time/op +6% | Reproduce locally with -count=15
Profile Drift | New hotspot >4% flat | Code archeology + changelog review
Cost Spike | CPU hours +12% | Check allocator churn & GC cycles

## 2. Triage Matrix
Impact vs Effort grid:
- High Impact / Low Effort → immediate PR
- High Impact / High Effort → design review
- Low Impact / Low Effort → batch fixes
- Low Impact / High Effort → defer/backlog

## 3. Investigation Toolkit
Signal | Primary Tool | Escalate To
-------|--------------|------------
Latency | Tempo traces | Trace + profile correlation
CPU | pprof top/diff | Block/Mutex profile
Memory | alloc_space diff | Escape analysis (-gcflags -m)
Alloc churn | alloc_objects | Object pooling experiment
Goroutine leak | goroutine profile | blocking ops instrumentation

## 4. Root Cause Patterns
Pattern | Signature | Strategy
--------|----------|---------
Algorithmic O(n^2) | Hotspot grows super-linearly with input | Redesign data structure
Excess allocations | High alloc_space without flat CPU | Reuse buffers / sync.Pool
Lock contention | Mutex profile spikes | Shard lock / reduce critical section
I/O Bound | Low CPU, high wall latency | Parallelize / pipeline
GC Pressure | Frequent short GC cycles | Reduce transient allocations

## 5. Remediation Workflow
1. Profile baseline (commit A)
2. Implement change (branch)
3. Run `ci_profiles` locally + AI analyzer
4. Confirm improvements > regressions
5. Add micro-bench if gap previously unmeasured
6. Merge behind feature flag if risky

## 6. Definition of Done (Performance PR)
- Baseline vs optimized profiles archived
- Benchstat delta ≤ +2% for unrelated benchmarks
- No new hotspots >5% without justification
- Gates pass (or waiver documented)
- Follow-up monitoring dashboard updated

## 7. Weekly Rituals
Ritual | Outcome
-------|--------
Hotspot Review | Rotate top 5 persistent CPU offenders
Allocation Audit | Focus on top alloc_space regressions
Benchmark Trend Scan | Detect slow drifts early
Gate Failure Postmortems | Improve thresholds / detection logic

## 8. Backlog Taxonomy
Category | Examples | KPI
---------|----------|----
Preventive | Pre-warm caches, pooling | Reduced p95 latency
Corrective | Remove quadratic join | Flat% decrease
Hygiene | Update benchmarks | Coverage % of critical paths
Strategic | Async pipeline redesign | Throughput gain

## 9. Metrics & KPIs
KPI | Target
----|-------
Net Performance Win Rate | >70% PRs with positive rating
Mean Time to Detect Regression | <1 day
Hotspot Churn Rate | <15% weekly
Benchmark Coverage (critical funcs) | >80%

## 10. Escalation Criteria
Escalate to arch review if:
- Any single regression >15% persists 3 PRs
- Overall rating <4 twice in a sprint
- Hotspot churn >25% (instability signal)

## 11. Tooling Enhancements Queue
Rank | Idea | Leverage
-----|------|--------
1 | Span-linked profile URLs | Pyroscope + Tempo
2 | Selective dynamic profiling | Agent API
3 | Historical baseline server | Object store + diff API
4 | SARIF export | GitHub code scanning UI
5 | Performance budget dashboard | Grafana JSON datasource

---
Treat performance as a product: observable state, feedback loops, user centric outcomes, and continuous iteration. This playbook institutionalizes that mindset.
