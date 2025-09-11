# Tutorial: Building a Trace Summarizer MCP

Objective: Implement `tracing.trace_summarizer` to convert raw trace data into actionable latency narrative.

Steps:
1. Define JSON Schema (inputs/outputs) ensuring required: trace_id
2. Implement client for Tempo/Jaeger gRPC or HTTP
3. Parse spans -> build parent/child edges
4. Compute exclusive duration: exclusive = span.duration - sum(child.durations overlapping)
5. Rank top N spans by exclusive_ms
6. Detect errors: status != OK or events containing exception
7. Generate narrative template (critical span + % of total)

Example Go Skeleton:
```go
// pseudo
trace := fetchTrace(ctx, traceID)
root := buildTree(trace.Spans)
calcExclusive(root)
critical := rankExclusive(root, 5)
errors := collectErrors(root)
summary := renderNarrative(critical, errors)
```

Edge Cases:
- Orphan spans: attach to synthetic root
- Large traces > max_spans: truncate & warn

Testing:
- Use fixture trace with known critical path; assert exclusive ordering
- Inject error span; ensure detection

Hardening:
- Add timeout + circuit breaker
- Limit output size (cap spans_total)

Extension Path:
- Add CPU vs wall correlation (needs profiling tags)
