# Tutorial: CI Orchestration & Chaining MCPs

Objective: Compose multiple MCP invocations into resilient CI performance workflows.

Principles:
- Isolate concerns per step (bench -> diff -> gate -> recs)
- Fail fast on contract errors, not on soft regressions
- Emit machine & human artifacts simultaneously

Pattern:
1. Preparation: build binaries, generate profiles
2. Benchmark stage
3. Analysis stage (diffs, trends)
4. Decision stage (gatekeeper)
5. Recommendation aggregation
6. Publish artifacts / comments

Resilience Techniques:
- Add retry wrapper only for upstream 5xx
- Cache heavy downloads (profiles) keyed by commit SHA
- Parallelize independent MCPs (jobs.matrix or background steps)

Example Guard Script:
```bash
set -euo pipefail
run() { echo "> $@"; "$@"; }
run mcp invoke perf.regression_gatekeeper gate.json > gate-out.json || exit 20
if jq -e '.decision=="fail"' gate-out.json; then
  echo "Hard regression fail"; exit 20
fi
```

Observability of CI Flow:
- Capture per-MCP duration -> time series (custom metric)
- Tag artifacts with commit, branch, build number

Extensions:
- Add dynamic threshold scaling based on trend volatility
- Integrate approval workflow for overriding hard fails
