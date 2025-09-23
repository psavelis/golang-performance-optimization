# Machine Context Protocol (MCP) for Observability & Performance

This section catalogs specialized MCP services that accelerate expert workflows across observability, tracing, and performance engineering.

Goals:
- Shorten MTTI / MTTR by automating high-cognitive investigation loops
- Normalize input/output contracts for safe automation in CI / Chat / IDE agents
- Provide repeatable playbooks mapped to concrete MCP capabilities

Structure:
1. Observability MCPs – metrics, logs, profiling queries
2. Tracing MCPs – latency, critical path, anomaly detection, correlation
3. Performance MCPs – benchmarks, flamegraphs, resource pressure analytics
4. Cross-Cutting MCPs – SLOs, cost, capacity, aggregated recommendations
5. Patterns – composition, chaining, governance
6. References – repos, schemas, examples
7. Tutorials – step-by-step implementations

Each MCP spec below follows a consistent template:
Field | Description
------|------------
Name | Canonical identifier
Problem | Pain addressed / latency in traditional workflow
Inputs | Required parameters (typed)
Outputs | Structured fields (JSON) + optional artifacts
Data Sources | Systems queried (Prometheus, Pyroscope, Tempo, etc.)
Algorithm | Core heuristic / model approach
Failure Modes | Expected errors + mitigation
Security | Auth scopes, least privilege
Example | Minimal invocation
Extension | Natural evolution / advanced variant

Use these MCPs as modular building blocks in: CI pipelines, on-demand chat copilots, pre-merge performance gates, automated incident retros.
