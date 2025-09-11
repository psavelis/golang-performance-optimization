# I/O Latency Bottleneck Finder MCP

Name: perf.io_latency_bottleneck_finder

Problem:
Identifying slow external dependencies across network & disk layers is time-consuming.

Inputs:
```json
{
  "span_metrics_prefix": "traces_span_duration_milliseconds_bucket",
  "io_labels": ["db.instance","peer.service"],
  "range_minutes": 15,
  "quantile": 0.95
}
```

Algorithm:
1. Query spanmetrics aggregated by io_labels
2. Compute quantile per label value
3. Rank deltas vs historical (optional)

Output:
```json
{
  "bottlenecks":[{"label":"db.instance=orders-primary","p95_ms":182.1},{"label":"peer.service=inventory","p95_ms":141.5}],
  "recommendations":["Add index to orders query path","Investigate inventory upstream latency"]
}
```

Extensions:
- Join with error ratio metric to prioritize
