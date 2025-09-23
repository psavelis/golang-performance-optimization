# Tutorial: Multi-Source Observability Flow MCP

Objective: Chain multiple MCPs to accelerate incident triage from alert -> root cause hypothesis.

Chain:
Alert Explainer -> Trace Summarizer -> Pyroscope Diff Analyzer -> Log Slicer -> Recommendation Orchestrator

Steps:
1. Invoke Alert Explainer with alert fingerprint
2. Use top contributing service(s) to fetch recent representative trace (Trace Summarizer)
3. Extract trace time window; feed into Pyroscope Diff (baseline vs current)
4. Run Log Slicer focusing on same service + error patterns
5. Feed resulting recommendations into Orchestrator for dedup & prioritization

Data Contract Tips:
- Pass correlation_id across steps for auditing
- Enforce max cumulative latency budget (e.g. 4s)

Output (orchestrated):
```json
{
  "correlation_id":"inc-2025-09-09-01",
  "actions":[
    {"action":"Optimize processBatch CPU hotspot","dimension":"cpu","score":0.89},
    {"action":"Mitigate DB latency spike (orders-primary)","dimension":"latency","score":0.84}
  ],
  "source_chain":["alert_explainer","trace_summarizer","pyroscope_diff","log_slicer"],
  "elapsed_ms": 1870
}
```

Governance:
- Abort chain if any step risk > defined threshold
- Log each MCP invocation to audit index

Extensions:
- Add SLO burn gating to escalate severity
- Provide Slack interactive card for action approval
