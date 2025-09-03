# Optimized Implementation Profiling

This directory contains the profiling-enabled version of the optimized event generator. It includes all optimizations (string pooling, efficient random generation, etc.) with added CPU and memory profiling instrumentation.

## Usage

To build and run profiling on the optimized implementation:

```bash
# Build manually
go build -o generator-optimized-profiling

# Run with profiling enabled
./generator-optimized-profiling 100000 test_optimized.json
```

## Generated Artifacts

The profiling runs produce `.prof` files in the current directory:
- `generator_optimized_cpu.prof`: CPU profile of optimized generator execution
- `generator_optimized_mem.prof`: Memory profile of optimized generator execution

These files can be converted to SVG flamegraphs:
```bash
go tool pprof -svg generator_optimized_cpu.prof > generator_optimized_cpu.svg
go tool pprof -svg generator_optimized_mem.prof > generator_optimized_mem.svg
```

## Implementation Details

This implementation combines the optimizations from the generator-optimized version with profiling instrumentation from the generator-profiling version. It provides an accurate performance profile of the optimized implementation for comparative analysis.
