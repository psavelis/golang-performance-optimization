# MCP Quickstart

Purpose: Fast path to invoke core MCPs locally & in CI.

Prerequisites:
- Go 1.21+ (if implementing MCP servers in Go)
- Access tokens: Prometheus (read), Pyroscope (read), Tempo/Jaeger (read)
- Bench & profiling artifacts generated (make ci_profiles)

Minimal Invocation Pattern:
```bash
mcp invoke prometheus.query_assistant payload.json > output.json
```

Recommended Env Vars:
| Var | Purpose |
|-----|---------|
| PROM_URL | Prometheus base URL |
| PYRO_URL | Pyroscope base URL |
| TEMPO_URL | Tempo/Jaeger API |
| MCP_TIMEOUT_SEC | Global timeout per invocation |
| MCP_CACHE_DIR | Local cache for fetched profiles |

Core Workflow Examples:
1. Hotspot Regression Check
   - Run: perf.flamegraph_diff_explainer -> perf.allocation_hotspot_classifier -> cross.optimization_recommendation_orchestrator
2. Post-Deploy Latency Validation
   - Run: tracing.latency_regression_detector -> tracing.trace_summarizer (for failing operations)
3. SLO Overshoot Early Warning
   - Run: cross.slo_error_budget_monitor -> tracing.span_anomaly_detector (focus top at-risk op)

Exit Codes (suggested convention):
Code | Meaning
-----|--------
0 | Success
10 | Soft warning (non-blocking)
20 | Hard fail (regression gate)
30 | Upstream dependency failure

Next: See CI Integration Examples for wiring in pipelines.
