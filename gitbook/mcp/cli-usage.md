# MCP CLI Usage

This page documents how to discover and validate Machine Capability Protocol (MCP) capabilities locally using the consolidated `mcp-cli` binary.

## Build
The standard `make build` target now produces `target/build/bin/mcp-cli`.

## List Capabilities
```
make mcp_list
```
Outputs: `id\tdomain\tmaturity`.

Include schema paths:
```
make mcp_schemas
```
Outputs: `id\tdomain\tmaturity\tinputSchema\toutputSchema`.

## Validate Input Against Schema
Provide a JSON file (or pipe via stdin) and run:
```
make mcp_validate CAP=prometheus.query_assistant FILE=./examples/prom-query.json
```
If validation passes you see `INPUT OK (schema validated)`.

## Contract Test (CI Ready)
A lightweight contract test target asserts that each capability schema rejects an empty object (ensuring required fields are enforced) and that schemas are reachable:
```
make mcp_contract_test
```
Failure indicates either a schema not found or overly-permissive schema.

## Direct Invocation Examples
Validate only:
```
./target/build/bin/mcp-cli -cap perf.benchmark_trend_analyzer -in bench_payload.json -validate
```
Execute stub (echo metadata):
```
./target/build/bin/mcp-cli -cap tracing.trace_summarizer -in trace_sample.json
```
Provide input via stdin:
```
cat trace_sample.json | ./target/build/bin/mcp-cli -cap tracing.trace_summarizer
```

## Exit Codes
- 0: Success / validation passed.
- 1: Any error (schema violation, manifest parse failure, missing capability, IO issue).

## Adding New Capabilities
1. Add input/output JSON Schemas under `mcp/schemas/{input,output}`.
2. Append new capability block to `mcp/manifest.yaml` (keep sorted by domain).
3. Run `make build && make mcp_validate CAP=<new.id> FILE=example.json`.
4. Update GitBook docs (this page or related tutorials) if needed.

## Future Enhancements
- Rich YAML parsing (replace minimal fallback with full parser).
- Automatic example payload generation from schema.
- Real handler dispatch mapping ID -> implementation.
- CI job publishing rendered API reference from schemas.

---
This CLI ensures MCP specs are executable and contract-tested, enabling automated integration and regression detection across the observability and performance lifecycle.
