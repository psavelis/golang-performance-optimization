# Critical Path Extractor MCP

Name: tracing.critical_path_extractor

Problem:
Identifying true latency contributors hidden by concurrency or fan-out is tedious.

Inputs:
```json
{
  "trace_id": "4f8cc5b9d3a1f1e2",
  "threshold_ms": 5,
  "include_attributes": ["db.statement","http.url"],
  "collapse_noise_pct": 1.5
}
```

Algorithm:
1. Build DAG of spans with start/finish times
2. Compute critical path (longest path by wall time)
3. Collapse spans below collapse_noise_pct of total wall
4. Emit ordered list + cumulative %

Output:
```json
{
  "trace_id":"4f8cc5b9d3a1f1e2",
  "critical_path_ms": 212.4,
  "segments":[
    {"span":"api handler","exclusive_ms":34.1,"cum_pct":16.0},
    {"span":"db SELECT products","exclusive_ms":61.3,"cum_pct":44.8},
    {"span":"cache PUT","exclusive_ms":27.9,"cum_pct":57.9}
  ],
  "recommendation":"Optimize DB query (29% exclusive); evaluate caching strategy."}
```

Extensions:
- Add what-if simulation (predict savings removing span)
