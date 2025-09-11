# Tutorial: Adding Pyroscope Correlation MCP

Objective: Implement `tracing.trace_profile_correlator` bridging traces and profiles.

Core Idea: Map trace spans to top CPU symbols within overlapping time windows to prioritize optimization efforts.

Workflow:
1. Inputs: trace_id, service, profile_type, window
2. Fetch trace; derive min/max timestamp
3. Query Pyroscope render (from, to) filtered by service + optional trace_id tag
4. Convert to pprof; parse top symbols
5. For each symbol, heuristically assign to span whose window overlaps largest sample share
6. Compute coverage_pct = sum(mapped_symbol_flat / total_flat)

Mapping Heuristic:
- OverlapWeight = durationOverlap(span, symbolTimeRange) * symbolFlat
- Choose span with max OverlapWeight

Output Fields:
- hotspots[] {fn, flat_pct, likely_span}
- coverage_pct
- recommendation

Edge Cases:
- No profile data: hotspots=[]; recommendation references enabling profiling
- Multiple spans same function: attribute to deepest in hierarchy

Security:
- Ensure trace_id validated length/charset to prevent injection in query params

Extensions:
- Add diff baseline capability
- Provide deep link to Pyroscope UI using computed window
