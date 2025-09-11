# Query Plan Regression Analyzer MCP

Name: perf.query_plan_regression_analyzer

Problem:
Database query plan shifts increase latency/cost unexpectedly after schema or data distribution changes.

Inputs:
```json
{
  "db_type": "postgres",
  "query_fingerprint": "select orders where status=? and created_at>?",
  "baseline_plan_id": "abc123",
  "current_plan_id": "def456",
  "metrics": ["total_cost","rows","width"],
  "threshold_pct": 10
}
```

Algorithm:
1. Fetch baseline & current plan JSON (EXPLAIN ANALYZE cached)
2. Diff nodes: cost, estimated vs actual rows
3. Flag nodes with > threshold_pct increase
4. Suggest index or rewrite

Output:
```json
{
  "plan_change": true,
  "regressions":[{"node":"Seq Scan orders","cost_delta_pct":+34.2,"rows_mismatch_pct":+220}],
  "recommendations":["Add composite index (status, created_at)","Update statistics"],
  "risk":"high"
}
```

Extensions:
- Integrate row-level sampling for skew detection
