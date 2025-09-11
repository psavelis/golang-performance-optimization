# MCP Patterns

Common composition & governance blueprints.

Pattern | Description | Example Chain
--------|-------------|--------------
Incident Triage | Compress logs + trace summary + hotspot profile | log slicer -> trace summarizer -> profile explorer
Regression Gate | Bench diff + flame diff + gatekeeper decision | trend analyzer -> diff explainer -> gatekeeper
Capacity Plan | Forecast + cost efficiency + SLO burn | forecasting -> cost analyzer -> error budget
Optimization Backlog | Aggregate top N improvements | diff explainer + alloc classifier + recommendation orchestrator

Governance:
- Enforce versioned MCP contracts (JSON Schema)
- Provide circuit breakers (max runtime, max bytes)
- Log structured audit events (who invoked, scope, latency)
