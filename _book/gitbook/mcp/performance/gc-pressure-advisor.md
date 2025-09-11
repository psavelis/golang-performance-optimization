# GC Pressure Advisor MCP

Name: perf.gc_pressure_advisor

Problem:
Unbounded heap growth + frequent GCs degrade latency.

Inputs:
```json
{
  "runtime_metrics_endpoint": "http://app:9090/metrics",
  "range_minutes": 30,
  "target_gc_pct": 20
}
```

Algorithm:
1. Pull go_memstats_* counters via PromQL
2. Compute allocation rate, GC frequency, pause percent
3. Compare live heap vs historical median
4. Recommend tuning (GOGC, pooling)

Output:
```json
{
  "alloc_rate_mb_s": 18.4,
  "gc_pause_pct": 2.7,
  "live_heap_mb": 612,
  "recommendations":["Reduce transient allocations in processor batch","Consider lowering GOGC to 80 if latency sensitive"],
  "risk_level":"medium"
}
```

Extensions:
- Add predictive model for heap exhaustion
