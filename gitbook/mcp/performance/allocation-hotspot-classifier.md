# Allocation Hotspot Classifier MCP

Name: perf.allocation_hotspot_classifier

Problem:
Excessive allocations silently erode throughput & increase GC overhead.

Inputs:
```json
{
  "alloc_profile": ".docs/artifacts/ci/profiles/alloc.pprof",
  "min_flat_bytes_pct": 1.0,
  "classify_patterns": ["bytes.Buffer", "json.Marshal"],
  "max_symbols": 40
}
```

Algorithm:
1. Parse allocation profile (space)
2. Rank symbols by flat bytes%
3. Pattern match classify_patterns for targeted recommendations

Output:
```json
{
  "top":[{"fn":"bytes.makeSlice","flat_bytes_pct":17.2}],
  "patterns":[{"pattern":"json.Marshal","matches":3}],
  "recommendations":["Pre-size buffers","Reuse encoder via sync.Pool"]
}
```

Extensions:
- Add alloc objects profile cross-reference
