# MCP Schemas

Provide machine-validated interfaces (JSON Schema) for consistent automation & contract testing.

Example: Regression Gatekeeper Input Schema
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "RegressionGatekeeperInput",
  "type": "object",
  "required": ["baseline_report", "candidate_report", "thresholds"],
  "properties": {
    "baseline_report": {"type": "string", "minLength": 1},
    "candidate_report": {"type": "string", "minLength": 1},
    "thresholds": {
      "type": "object",
      "properties": {
        "ns_per_op_pct": {"type": "number", "minimum": 0},
        "allocs_per_op_pct": {"type": "number", "minimum": 0}
      },
      "required": ["ns_per_op_pct", "allocs_per_op_pct"]
    },
    "hard_fail_pct": {"type": "number", "minimum": 0},
    "min_sample": {"type": "integer", "minimum": 1}
  }
}
```

Example: Trace Summarizer Output Schema (excerpt)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "TraceSummarizerOutput",
  "type": "object",
  "required": ["trace_id", "spans_total", "critical_path_ms"],
  "properties": {
    "trace_id": {"type": "string", "pattern": "^[a-f0-9]{16,32}$"},
    "spans_total": {"type": "integer", "minimum": 1},
    "critical_path_ms": {"type": "number", "minimum": 0},
    "top_exclusive": {"type": "array", "items": {"type": "object", "required":["span","exclusive_ms"], "properties": {"span":{"type":"string"}, "exclusive_ms":{"type":"number"}}}},
    "errors": {"type": "array"},
    "narrative": {"type": "string"}
  }
}
```

Schema Governance:
- Version fields (schema_version) for backward compatibility
- CI contract tests: validate fixture payloads
- Reject unknown properties (additionalProperties=false) for strictness in critical MCPs

Tooling:
- Use ajv (Node), gojsonschema (Go), or jsonschema crate (Rust)

Next Steps:
- Add full catalog of schemas per MCP (deferred)
