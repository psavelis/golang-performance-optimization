# Documentation index

## Core Documentation

| Document | Purpose |
|----------|---------|
| **[Technical specification v1](technical-specification-v1.md)** | Implementation architecture and patterns |
| **[Flamegraph analysis summary](flamegraph-analysis-summary.md)** | Visual profiling insights and analysis |

## Key Performance Results

**Measured Improvement**: 5.35x execution time reduction through string pool optimization

| Metric | Original | Optimized | Improvement |
|--------|----------|-----------|-------------|
| **Execution Time** | 798ms | 149ms | 5.35x faster |
| **Memory Usage** | 4.85MB | 1.52MB | 68.7% reduction |
| **Allocations** | 187,602 | 3,033 | 98.4% reduction |
| **String Generation** | 674ns/op | 7.8ns/op | 86x faster |

## Navigation

### Documentation
- **Setup & Usage**: `../README.md`
- **Build Commands**: `../Makefile`
- **Technical details**: `technical-specification-v1.md`
- **Visual analysis**: `flamegraph-analysis-summary.md`

### Code Structure
- **Benchmarks**: `../pkg/benchmarks/`
- **Implementations**: `../cmd/`
- **Test Generator**: `../cmd/generator/`
- **Database Loader**: `../cmd/loader/`

## Artifacts Structure

```
.docs/artifacts/
├── benchmarks/                     # Historical benchmark data
├── flamegraphs/                    # SVG flamegraph visualizations
├── latest-flamegraphs/            # Current optimization comparisons
├── latest-profiles/               # CPU, memory, block, mutex profiles
├── profiles/                      # Historical profile data
├── test-data/                     # Generated test datasets
├── traces/                        # Execution traces
├── generator_bench.txt            # Generator performance benchmarks
├── latest_generator_benchmark.txt # Current benchmark results
├── memory_bench.txt               # Memory allocation analysis
└── string_bench.txt               # String operation benchmarks
```

## Quick Command Reference

### Benchmark and Validation
```bash
# Performance comparison
make benchmark_compare

# Quick benchmark validation  
make bench_quick

# Complete profiling analysis
make profiling_study
```

### Profiling Tools
```bash
# Interactive profiling server
make pprof_server

# Generate flamegraphs
make generate_enhanced_flamegraphs

# Statistical benchmark validation
go test -bench=BenchmarkGenerator -count=3 -benchmem ./pkg/benchmarks/
```

## Implementation Analysis

### Core Optimization: String Pool
The primary performance improvement comes from eliminating string concatenation through pre-allocated string pools.

**Original**: Dynamic string generation with concatenation  
**Optimized**: Array lookup from pre-allocated strings  
**Result**: 86x faster string operations (674ns → 7.8ns per operation)

### Implementations Compared
1. **Original**: Baseline implementation with string concatenation
2. **Optimized**: String pool implementation with enhanced profiling
3. **Streaming**: Constant memory usage for arbitrary dataset sizes

---

**Documentation Status**: Current, validated measurements  
**Last Updated**: September 2025  
**Validation**: 95% statistical confidence
