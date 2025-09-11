# Tutorial: Schema Validation & Contracts

Objective: Add JSON Schema validation to enforce strict MCP input/output contracts.

Steps:
1. Author JSON Schema for each MCP input & output
2. Implement validation layer in MCP gateway (pre-handler)
3. Reject payloads with unknown properties (additionalProperties=false) for critical MCPs
4. Maintain schema versions (schema_version) & support backwards compatibility window

Go Example (gojsonschema):
```go
schemaLoader := gojsonschema.NewReferenceLoader("file://schemas/reg_gate_input.json")
bodyLoader := gojsonschema.NewBytesLoader(reqBody)
result, _ := gojsonschema.Validate(schemaLoader, bodyLoader)
if !result.Valid() { /* collect errs */ }
```

CI Contract Tests:
- Store canonical fixtures under `schemas/fixtures/`
- Validate fixtures against schema each pipeline run

Drift Detection:
- Hash schema content; alert if runtime MCP advertises hash mismatch

Extensions:
- Generate typed structs from schema (quicktype / jsonschema2go)
- Embed schema in MCP handshake for dynamic discovery
