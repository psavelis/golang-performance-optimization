# MCP Integration Architecture

High-level flow for chained observability & performance MCPs.

Mermaid (conceptual):
```mermaid
graph TD
  A[Invoker (CI / Chat / IDE)] --> B[Gateway / Broker]
  B --> C1[Observability MCPs]
  B --> C2[Tracing MCPs]
  B --> C3[Performance MCPs]
  B --> C4[Cross-Cutting MCPs]
  C1 --> D[Recommendation Orchestrator]
  C2 --> D
  C3 --> D
  C4 --> D
  D --> E[Backlog / Ticketing]
  D --> F[PR Comment / Report]
```

Components:
- Gateway: AuthN/Z, rate limits, audit log, fan-out
- MCP Worker: Executes domain logic; runtime isolation (process / container)
- Cache Layer: Short-lived profile & metrics snapshots
- Artifact Store: Persist diff outputs / regression evidence (.docs/artifacts)

Governance Controls:
Control | Purpose
--------|--------
Rate limiting | Prevent abuse / cost overruns
Schema validation | Contract safety
Timeout budget | Guarantee responsiveness
Retry policy (idempotent ops) | Resilience
Structured audit events | Compliance & forensics

Security:
- Principle of least privilege per MCP (scoped API tokens)
- Secrets injection via runtime environment only
- Disallow arbitrary shell exec in payloads (no code injection)

Observability of MCP Layer:
- Per-MCP latency histogram
- Error categorization (validation, upstream timeout, internal)
- Success vs degraded vs failed invocation ratio

Failure Handling:
- Partial chain fallback (skip failed MCP with warning)
- Circuit break noisy upstream (e.g., Pyroscope outage)

Pluggability:
- Add new MCP by registering schema + handler manifest (YAML)
- Broker dynamically discovers & publishes capability list
