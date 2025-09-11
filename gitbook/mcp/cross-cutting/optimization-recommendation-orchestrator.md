# Optimization Recommendation Orchestrator MCP

Name: cross.optimization_recommendation_orchestrator

Problem:
Fragmented recommendations across tools slow strategic planning.

Inputs:
```json
{
  "sources": ["perf.flamegraph_diff_explainer","perf.allocation_hotspot_classifier","tracing.span_anomaly_detector"],
  "prioritize": ["latency","cpu","cost"],
  "max_recommendations": 10
}
```

Algorithm:
1. Collect recommendations JSON from source MCP outputs
2. Tag each with dimension (latency/cpu/memory/cost/reliability)
3. Deduplicate semantically (hash stemmed key terms)
4. Score by (impact_weight * recency * confidence)

Output:
```json
{
  "items":[
    {"action":"Optimize processBatch CPU hotspot","dimension":"cpu","score":0.91},
    {"action":"Shard processor lock","dimension":"latency","score":0.77}
  ],
  "next_top_3":["processBatch optimization","DB index creation","Cache timeout tuning"],
  "method":"weighted_merge_v1"
}
```

Extensions:
- Add backlog export (Jira API)
- Add ROI estimation (cost saved vs effort)
