# Continuous Profiling in CI (Pull Requests)

This guide shows how each Pull Request automatically generates profiling data, flamegraphs, and benchmark comparisons so reviewers can confidently assess performance impact.

## Objectives

- Capture CPU & memory profiles for a representative workload in PRs.
- Generate SVG flamegraphs for quick hotspot visualization.
- Provide top CPU and allocation hotspots directly in the PR summary.
- Compare micro-benchmarks (baseline main vs PR) using `benchstat`.
- Keep execution fast (< ~2 min) to avoid slowing developer feedback loops.

## Workflow Overview

Two GitHub Actions jobs:

1. `profiling` — Builds profiling binaries, runs them with a reduced dataset, extracts CPU/heap profiles, generates flamegraphs, and uploads artifacts.
2. `benchdiff` — Checks out `main` for baseline benchmarks, then current PR branch, runs key benchmarks multiple times, and generates statistical diffs with `benchstat`.

Artifacts exposed to reviewers:

- `ci-profiles-<sha>/flamegraphs/*.svg`
- CPU top stacks: `generator_cpu_top.txt`, `generator_optimized_cpu_top.txt`
- Allocation top stacks: `generator_mem_top.txt`, `generator_optimized_mem_top.txt`
- Benchmark diff summaries in `benchdiff-<sha>/`

A PR comment is posted on first run (optional) and rich summaries appear in the Checks tab.

## Key Makefile Target

The CI invokes a purpose-built target:

```makefile
ci_profiles: clean
	mkdir -p .docs/artifacts/ci/flamegraphs
	go build -o $(BUILD_DIR)/bin/generator-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-profiling
	go build -o $(BUILD_DIR)/bin/generator-optimized-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-optimized-profiling
	$(BUILD_DIR)/bin/generator-profiling 50000 /dev/null || true
	$(BUILD_DIR)/bin/generator-optimized-profiling 50000 /dev/null || true
	go tool pprof -svg generator_cpu.prof > .docs/artifacts/ci/flamegraphs/generator_cpu.svg 2>/dev/null || true
	go tool pprof -svg generator_optimized_cpu.prof > .docs/artifacts/ci/flamegraphs/generator_optimized_cpu.svg 2>/dev/null || true
	go tool pprof -top -nodecount=15 generator_cpu.prof > .docs/artifacts/ci/generator_cpu_top.txt 2>/dev/null || true
	go tool pprof -top -nodecount=15 generator_optimized_cpu.prof > .docs/artifacts/ci/generator_optimized_cpu_top.txt 2>/dev/null || true
	go tool pprof -top -alloc_space -nodecount=15 generator_mem.prof > .docs/artifacts/ci/generator_mem_top.txt 2>/dev/null || true
	go tool pprof -top -alloc_space -nodecount=15 generator_optimized_mem.prof > .docs/artifacts/ci/generator_optimized_mem_top.txt 2>/dev/null || true
```

## GitHub Workflow (Excerpt)

```yaml
jobs:
  profiling:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go mod download
      - run: make ci_profiles
      - uses: actions/upload-artifact@v4
        with:
          name: ci-profiles-${{ github.sha }}
          path: .docs/artifacts/ci
      - name: Summarize hotspots
        run: |
          echo '## CPU Hotspots' >> $GITHUB_STEP_SUMMARY
          sed -n '1,25p' .docs/artifacts/ci/generator_cpu_top.txt >> $GITHUB_STEP_SUMMARY || true
```

The benchmark comparison job fetches `main` as a baseline and runs `benchstat`:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat /tmp/base.txt /tmp/pr.txt > bench_original.diff
benchstat /tmp/base_opt.txt /tmp/pr_opt.txt > bench_optimized.diff
```

## Reading the Flamegraphs

Look for:

- Wide stacks high in the graph: hot cumulative paths.
- Repeated string allocation or JSON marshaling frames: tuning targets.
- Excessive time in `encoding/json` vs business logic: candidate for alternative encoders.

## Reading the Hotspot Tables (pprof -top)

Columns:

- `flat`: Time spent directly in function.
- `flat%`: Percent of total samples for `flat`.
- `sum%`: Cumulative percentage up to this row.
- `cum`: Cumulative time including callees.

Prioritize functions with high `flat%` and high `cum` if they’re on critical paths.

## Reading the Allocation Tables (-alloc_space)

- Focus on large `flat` allocators to reduce GC pressure.
- Watch for transient allocations in tight loops.

## Benchmark Diff Interpretation

A `benchstat` diff example:

```text
name                       old time/op    new time/op    delta
GeneratorOriginal-8          2.45ms ± 3%    2.10ms ± 2%  -14.3%  (p=0.002 n=5+5)
GeneratorOptimized-8         1.10ms ± 2%    1.05ms ± 2%   -4.5%  (p=0.041 n=5+5)
```

Key signals:

- `delta`: Relative improvement (negative = faster / fewer allocs).
- `p=`: Statistical significance (<= 0.05 is generally meaningful).
- Always consider variance ±%. High variance may need more iterations.

## Extending Further

| Goal | Approach |
|------|----------|
| Track regressions over time | Persist artifacts to S3 + compare last N builds |
| Alert on > X% slowdown | Add a parsing step + GitHub Status API failure |
| Deeper diff (flamegraph diff) | Integrate `speedscope` JSON export + PR link |
| Continuous profiling (runtime) | Run Pyroscope in CI ephemeral container + capture 30s sample |

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Missing SVG flamegraphs | Ensure `go tool pprof` available (Go installed) |
| Empty hotspot files | Workload too small; raise event count or remove `|| true` for visibility |
| Bench diffs empty | Baseline branch fetch failed; ensure `main` exists upstream |

## Review Checklist for PRs

- [ ] Any increase in `flat%` for JSON or string generation functions?
- [ ] Did optimized build show expected lower allocations?
- [ ] Benchmark deltas negative (improvement) or neutral?
- [ ] No large variance suggesting instability?
- [ ] Flamegraph width dominated by intended hot loops only?

This pipeline turns performance review into a first-class, automated signal—use it to prevent regressions early.
