# Optimized Implementation Profiling

This directory contains the profiling-enabled version of the optimized event loader. It includes all optimizations (worker pools, streaming processing, batch operations, etc.) with added CPU and memory profiling instrumentation.

## Usage

To build and run profiling on the optimized implementation:

```bash
# Build manually
go build -o loader-optimized-profiling

# Run with profiling enabled
./loader-optimized-profiling "postgresql://test:test@localhost:5432/test?sslmode=disable" test_optimized.json
```

## Generated Artifacts

The profiling runs produce `.prof` files in the current directory:
- `loader_optimized_cpu.prof`: CPU profile of optimized loader execution
- `loader_optimized_mem.prof`: Memory profile of optimized loader execution

These files can be converted to SVG flamegraphs:
```bash
go tool pprof -svg loader_optimized_cpu.prof > loader_optimized_cpu.svg
go tool pprof -svg loader_optimized_mem.prof > loader_optimized_mem.svg
```

## Implementation Details

This implementation combines the optimizations from the loader-optimized version with profiling instrumentation from the loader-profiling version. It provides an accurate performance profile of the optimized implementation for comparative analysis.
