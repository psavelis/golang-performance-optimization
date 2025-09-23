# Pyroscope Profile Explorer MCP

Name: pyroscope.profile_explorer

Problem:
Manual UI navigation slows hotspot triage, especially under incident pressure.

Inputs:
```json
{
  "service": "generator-otel",
  "profile_type": "cpu",
  "range_minutes": 15,
  "group_by": ["function"],
  "min_flat_pct": 1.0,
  "max_symbols": 30,
  "tag_filters": {"trace_id": "abc123"}
}
```

Process:
1. Query Pyroscope /render for time range
2. Convert profile to pprof & parse top table
3. Filter symbols by min_flat_pct & limit
4. Emit enriched JSON + optional markdown summary

Output:
```json
{
  "service":"generator-otel",
  "profile_type":"cpu",
  "window":"2025-09-09T11:00Z..2025-09-09T11:15Z",
  "symbols":[
    {"fn":"processBatch","flat_pct":23.4,"cum_pct":55.1},
    {"fn":"compressPayload","flat_pct":11.2,"cum_pct":18.7}
  ],
  "total_samples": 132245,
  "filters":{"trace_id":"abc123"},
  "recommendation":"Focus on processBatch; >20% flat and >50% cumulative chain."}
```

Failure Modes:
- Empty profile -> symbols=[]
- Pyroscope timeout -> include retry_after suggestion

Extensions:
- Add diff baseline param
- Export Speedscope URL deep link
