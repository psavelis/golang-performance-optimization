# Performance Analysis Artifacts

This directory contains supporting artifacts from the performance optimization process.

## Directory Structure

### [`flamegraphs/`](flamegraphs/)
**CPU and Memory Utilization Visualizations**
- Hierarchical representation of execution profiles and resource allocation patterns
- Comparative analysis of original vs. optimized implementations
- Empirical evidence of performance bottlenecks and optimization efficacy

**SVG Profile Resources:**
- Generator CPU Utilization:
  - [Original Generator CPU](flamegraphs/generator_original_cpu.svg) - Reveals `generator.String()` (41.2%), `uuid.New()` (18.7%), and `json.Marshal()` (22.1%) as primary CPU consumers
  - [Optimized Generator CPU](flamegraphs/generator_optimized_cpu.svg) - Documents reduction through `getRandomStringFromPool()` (4.2%) and efficient serialization patterns

- Generator Memory Allocation:
  - [Original Generator Memory](flamegraphs/generator_original_mem.svg) - Identifies `make([]byte)` in string generation (35.7%) and marshal buffer expansion as dominant allocations
  - [Optimized Generator Memory](flamegraphs/generator_optimized_mem.svg) - Demonstrates static allocation model with `stringPool[]` (0.4%) and pre-sized buffers

- Loader CPU Utilization:
  - [Original Loader CPU](flamegraphs/loader_original_cpu.svg) - Exposes `json.Unmarshal()` (32.4%) and `database/sql.Exec()` (37.2%) as primary bottlenecks
  - [Optimized Loader CPU](flamegraphs/loader_optimized_cpu.svg) - Illustrates concurrent execution via `processWorker()` goroutines with distributed `stmt.Exec()` (28.3%)

- Loader Memory Allocation:
  - [Original Loader Memory](flamegraphs/loader_original_mem.svg) - Demonstrates monolithic allocation pattern in `json.Unmarshal()` (78.3% of heap)
  - [Optimized Loader Memory](flamegraphs/loader_optimized_mem.svg) - Shows constant memory footprint through `processEventsStreaming()` with fixed batch sizes

**Quantitative Optimization Metrics:**
- String allocation reduction: 35.7% → 0.4% of total heap (98.9% improvement)
- Random number generation optimization: 22.8% → 3.9% CPU utilization (82.9% reduction)
- Loader memory efficiency: 470MB → 5MB peak utilization (98.9% reduction)
- Database throughput improvement: 2,305 → 15,257 events/second (6.6× increase)
- Query preparation overhead: 29.5% → 5.7% CPU utilization (80.7% reduction)

### [`profiles/`](profiles/)
**Go Profiling Data**
- `generator_cpu.prof` - CPU profile of original generator implementation
- `generator_mem.prof` - Memory profile showing allocation patterns
- `loader_cpu.prof` - CPU profile revealing database bottlenecks
- `loader_mem.prof` - Memory profile showing 590MB+ memory explosion

**Usage:**
```bash
# View CPU profile with flame graph
go tool pprof -http=:8080 profiles/generator_cpu.prof

# Analyze memory allocations
go tool pprof profiles/generator_mem.prof

# Generate SVG flamegraph
go tool pprof -svg profiles/generator_cpu.prof > flamegraphs/custom_flamegraph.svg
```

### [`test-data/`](test-data/)
**Sample JSON Output Files**
- Test datasets used during performance analysis and validation
- Various sizes used for benchmarking and scalability testing

### [`profiling/`](../assets/profiling/)
**Profiling-Enabled Source Code**
- `generator-profiling/main.go` - Generator with CPU/memory profiling instrumentation
- `loader-profiling/main.go` - Loader with profiling capabilities enabled
- Used for collecting performance profiles during bottleneck analysis

## Key Profiling Insights

### CPU Bottlenecks Identified
- **61% syscall overhead** in original loader
- **15.96% random generation** inefficiency
- **12.3% string concatenation** in generator

### Memory Issues Found
- **590MB memory explosion** for 100K events in loader
- **21.5MB string allocations** in generator
- **High GC pressure** from repeated allocations

These artifacts provide concrete evidence of the performance issues addressed through systematic optimization.
