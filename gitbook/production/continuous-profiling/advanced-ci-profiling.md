# Advanced CI Profiling & Regression Gates

This chapter extends the basic PR profiling pipeline with diff analysis, automatic regression detection, and actionable artifacts for reviewers.

## Enhancements Added

| Feature | Purpose |
|---------|---------|
| CPU & allocation diff (pprof -diff_base) | Quickly see what changed between original and optimized implementations |
| Regression gate script | Fail (or warn) if benchmarks degrade beyond a threshold |
| Benchstat statistical comparison | Noise-resistant performance signal |
| Structured artifact layout | Consistent retrieval and historic comparison foundation |

## Artifact Layout

```text
.docs/artifacts/ci/
  flamegraphs/
    generator_cpu.svg
    generator_optimized_cpu.svg
  generator_cpu_top.txt
  generator_optimized_cpu_top.txt
  generator_cpu_diff_top.txt
  generator_mem_top.txt
  generator_optimized_mem_top.txt
  generator_mem_diff_top.txt

.docs/artifacts/benchdiff/
  bench_original.diff
  bench_optimized.diff
```

## Diff Output Interpretation

`generator_cpu_diff_top.txt` is produced by:

```bash
go tool pprof -top -diff_base=generator_cpu.prof generator_optimized_cpu.prof
```

Meaning of signs:
- Positive flat% in diff: function took more self time in optimized version (possible regression)
- Negative flat%: function improved (less self time)

Allocation diff uses:

```bash
go tool pprof -top -alloc_space -diff_base=generator_mem.prof generator_optimized_mem.prof
```

Focus on large positive changes in `flat` or `flat%`—they often correlate with new allocation hot spots.

## Regression Gate Script

`scripts/parse_bench_regressions.sh` scans `bench_*.diff` for positive deltas above a configurable threshold.

Example invocation in CI:

```bash
THRESHOLD_PERCENT=5 ./scripts/parse_bench_regressions.sh
```

Behavior:
- Prints each regression line over threshold
- Exit code 2 signals a failure (currently tolerated with `|| true` until policy enforced)

To enforce hard failures, remove `|| true` in the workflow step.

## Recommended Review Flow

1. Open PR Checks summary → skim CPU/alloc hotspot tables
2. Open flamegraphs (artifact download) → verify shifts in dominant stacks
3. Read diff tables → confirm reductions in JSON/string heavy call stacks
4. Inspect benchstat diffs → validate statistical significance (p-values <= 0.05)
5. If regression flagged, drill into offending function via profile UI (optional local pprof)

## Tightening the Signal (Optional)

| Goal | Enhancement | Tooling |
|------|-------------|---------|
| Reduce noise | Increase -count to 10 for critical benchmarks | benchstat |
| Faster feedback | Parallelize benchmark groups | matrix strategy |
| Deeper diffing | Export Speedscope JSON & link viewer | `go tool pprof -json` + speedscope.app |
| Store history | Push artifacts to object storage with commit key | custom action |
| Alerting | Add GitHub Status check if regression flagged | REST API call |

## Local Reproduction

Recreate CI artifacts locally:

```bash
make ci_profiles
ls .docs/artifacts/ci
cat .docs/artifacts/ci/generator_cpu_top.txt | head -25
```

Inspect diff interactively:

```bash
go tool pprof -http=:8088 -diff_base=generator_cpu.prof generator_optimized_cpu.prof
```

## Next Steps & Ideas

- Integrate continuous (runtime) sampling via ephemeral Pyroscope capture in PRs
- Add memory leak detection heuristic (track goroutine count + heap size growth across iterations)
- Include mutex/block profile sampling variants for contention analysis

With these extensions, performance regressions become visible, explainable, and enforceable early in the development cycle.
