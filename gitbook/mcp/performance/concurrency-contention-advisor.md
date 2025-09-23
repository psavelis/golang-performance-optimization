# Concurrency Contention Advisor MCP

Name: perf.concurrency_contention_advisor

Problem:
Lock contention silently limits parallel throughput.

Inputs:
```json
{
  "mutex_profile": ".docs/artifacts/ci/profiles/mutex.pprof",
  "block_profile": ".docs/artifacts/ci/profiles/block.pprof",
  "min_wait_pct": 1.0,
  "max_symbols": 30
}
```

Algorithm:
1. Parse mutex & block profiles
2. Rank symbols by wait time flat%
3. Identify patterns (coarse locks, channel misuse)

Output:
```json
{
  "mutex_hotspots":[{"fn":"processor.Process","wait_flat_pct":12.3}],
  "block_hotspots":[{"fn":"<-time.After","wait_flat_pct":4.1}],
  "recommendations":["Shard processor lock","Replace time.After in hot loop with timer reuse"]
}
```

Extensions:
- Integrate goroutine dump classification
