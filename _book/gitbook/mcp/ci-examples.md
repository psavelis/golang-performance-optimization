# MCP CI Integration Examples

Scenarios demonstrating practical chaining inside GitHub Actions or other CI systems.

## Example 1: Benchmark Regression Gate + Flamegraph Diff
```yaml
jobs:
  perf-guard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Benchmarks
        run: go test -bench=. -benchmem ./... | tee bench.txt
      - name: Convert Benchmarks to JSON
        run: ./scripts/bench_to_json.sh bench.txt > baseline.json
      - name: Run Regression Gatekeeper MCP
        run: mcp invoke perf.regression_gatekeeper gate-input.json > gate-output.json
      - name: CPU Profiles
        run: make ci_profiles
      - name: Flamegraph Diff
        run: mcp invoke perf.flamegraph_diff_explainer diff-input.json > diff-output.json
      - name: Orchestrate Recommendations
        run: mcp invoke cross.optimization_recommendation_orchestrator rec-input.json > rec-output.json
      - name: Fail on Hard Regression
        run: jq -e '.decision=="fail"' gate-output.json || exit 0
```

## Example 2: Post-Deploy Latency Verification
```yaml
jobs:
  latency-verify:
    steps:
      - name: Detect Regression
        run: mcp invoke tracing.latency_regression_detector input.json > out.json
      - name: Summarize Trace
        if: always()
        run: mcp invoke tracing.trace_summarizer trace-input.json > trace-out.json
```

## Example 3: Nightly Trend & Capacity Forecast
```yaml
jobs:
  nightly-trend:
    schedule: ['0 2 * * *']
    steps:
      - run: mcp invoke perf.benchmark_trend_analyzer trend-input.json > trend.json
      - run: mcp invoke cross.capacity_forecasting_engine cap-input.json > cap.json
      - run: mcp invoke cross.slo_error_budget_monitor slo-input.json > slo.json
      - run: mcp invoke cross.performance_risk_scoring risk-input.json > risk.json
      - name: Publish Report
        run: ./scripts/publish_perf_report.sh
```

## Governance Recommendations
- Enforce max runtime via MCP_TIMEOUT_SEC
- Cache profile & metrics responses to reduce repeated cost
- Store artifacts under `.docs/artifacts/ci/mcp/` namespaced by job

## Failure Strategy
Classification | Action
--------------|-------
Hard Fail | Block merge / deployment
Soft Warn | Add PR comment with guidance
Upstream Error | Retry once with backoff

## Artifact Conventions
File | Purpose
----|--------
`gate-output.json` | Regression decision
`diff-output.json` | Flamegraph diff summary
`trend.json` | Historical trend evaluation
`risk.json` | Composite risk score

Extend by adding SARIF for IDE surfacing or Slack notifications.
