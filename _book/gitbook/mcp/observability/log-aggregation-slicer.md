# Log Aggregation Slicer MCP

Name: logs.aggregation_slicer

Problem:
During incidents, reducing millions of log lines into high-signal clusters is slow.

Inputs:
```json
{
  "query": "{service='loader', level='error'} |= 'timeout'",
  "range_minutes": 20,
  "cluster_method": "fingerprint",
  "max_clusters": 12,
  "include_samples": 2
}
```

Algorithm:
1. Query Loki (or Elasticsearch) for lines
2. Normalize (strip numbers/UUIDs)
3. Fingerprint or minhash cluster
4. Rank by count & recency

Output:
```json
{
  "total_lines": 8421,
  "clusters": [
    {"pattern":"request timeout upstream service=inventory","count":3114,"pct":0.37,"samples":["...","..."]},
    {"pattern":"db retry exceeded attempts=5","count":1542,"pct":0.18}
  ],
  "recommendations":["Investigate inventory upstream latency","Adjust DB retry backoff policy"],
  "method":"fingerprint"
}
```

Failure Modes:
- Query exceeds limit -> advise narrower time slice

Extensions:
- Add anomaly score per cluster
- Provide diff vs previous 24h window
