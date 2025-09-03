# Event Processing System - Performance Optimization

This repository demonstrates performance optimization techniques for a Go-based event processing system. It showcases how systematic profiling-driven development can identify and resolve bottlenecks in data processing applications.

## Performance Results

| Metric | Baseline | Optimized | Improvement |
|--------|----------|-----------|-------------|
| Event Generation | 835ms | 139ms | 6.01× faster |
| Database Loading | 43.38s | 7.41s | 5.85× faster |
| Total Processing | 44.22s | 7.55s | 5.86× faster |
| Throughput | 2,305 events/sec | 15,257 events/sec | 6.6× higher |
| Memory Footprint | 590MB+ peak | 5MB constant | 99% reduction |

## Quick Start

```bash
# Setup environment
make start_env && sleep 3
docker exec -i postgres psql -U test -d test < env/data/postgres/00-schema.sql

# Build all versions  
make build && make build_optimized

# Generate 100K events
target/build/bin/generator-optimized 100000 events.json          # 139ms

# Load to PostgreSQL  
target/build/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' events.json  # 7.4s
```

## Performance Analysis

Performance profiles and visualizations are available in the [.docs/artifacts](.docs/artifacts/) directory.

### Key Optimizations

1. **Generator Component:**
   - Replaced `generator.String()` with static `stringPool` lookup (41.2% → 4.2% CPU)
   - Eliminated dynamic allocations through pooling (35.7% → 0.4% heap usage)
   - Optimized random generation with efficient bitwise operations

2. **Loader Component:**
   - Transformed architecture from monolithic to worker pool concurrency
   - Replaced `json.Unmarshal()` with streaming `json.Decoder` (78.3% → 10.8% memory)
   - Implemented prepared statement batching (37.2% → 28.3% CPU, distributed)
   - Configured connection pooling parameters for optimal throughput

## System Architecture

The optimized system employs a pipeline architecture:

```
┌─────────────────┐     ┌───────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  String Pooling │ ──► │ Streaming JSON │ ──► │  Worker Pool    │ ──► │ Batch Database  │
│    Generator    │     │   Processing   │     │     Loader      │     │   Operations    │
└─────────────────┘     └───────────────┘     └─────────────────┘     └─────────────────┘
      139ms/100K          Constant 5MB         15.2K events/sec        1000 rows/batch
```

## Technical Improvements

| Category | Issue | Solution |
|----------|-------|----------|
| Memory | Excessive allocations | Streaming architecture with constant memory profile |
| Security | SQL injection risks | Parameterized queries with prepared statements |
| Concurrency | Sequential processing | Worker pool with controlled parallelism |
| Resilience | Connection failures | Automatic retry logic with exponential backoff |
| Performance | Individual transactions | Batched operations with transaction pooling |

## Repository Structure

```
├── cmd/                      # Application components
│   ├── generator/           # Original event generator
│   ├── generator-optimized/ # Optimized event generator
│   ├── generator-profiling/ # Profiling-enabled generator
│   ├── loader/              # Original database loader
│   ├── loader-optimized/    # Optimized database loader
│   └── loader-profiling/    # Profiling-enabled loader
├── env/                     # Environment configuration
│   └── data/                # Database setup scripts
├── pkg/                     # Shared packages
│   ├── generator/           # String generation utilities
│   └── model/               # Data structures
└── .docs/                   # Documentation
    ├── analysis/            # Performance analysis
    ├── artifacts/           # Profiling data and visualizations
    │   ├── flamegraphs/    # CPU and memory visualizations
    │   ├── profiles/       # Raw profiling data
    │   └── test-data/      # Benchmark datasets
    ├── benchmarks/          # Performance measurements
    └── optimization/        # Implementation details
```

For comprehensive documentation, see the [detailed analysis](.docs/README.md).

## Key Optimizations

1. **Streaming Processing**: Implementation of constant memory processing regardless of input dataset size
2. **String Pooling**: Reduction of memory allocations through string interning techniques
3. **Batch Operations**: Consolidation of database operations from individual inserts to batched transactions
4. **Worker Concurrency**: Implementation of parallel processing with efficient resource management
5. **Connection Pooling**: Optimization of database connection handling for sustained throughput

## Critical Issues Addressed

| Category | Issue | Solution |
|----------|-------|----------|
| Memory | Excessive allocations | Implemented streaming architecture with constant memory profile |
| Security | SQL injection risks | Added parameterized queries and proper input validation |
| Concurrency | Sequential processing | Developed worker pool with controlled concurrency |
| Resilience | Connection failures | Implemented automatic retry logic with backoff strategy |
| Performance | Individual transactions | Consolidated operations into efficient batches |

## Repository Structure

```
├── cmd/                         # Source implementations
│   ├── generator/              # Original baseline generator
│   ├── generator-optimized/    # Optimized generator
│   ├── loader/                 # Original baseline loader
│   └── loader-optimized/       # Optimized loader
├── env/                        # Database environment setup
├── pkg/                        # Shared utilities and models
└── .docs/                      # Documentation and analysis
    ├── study.md               # Complete performance study
    ├── analysis/              # Bottleneck identification
    ├── benchmarks/            # Performance measurements
    ├── optimization/          # Implementation techniques
    └── proof-of-concept/      # Verification artifacts and test results
```

## Documentation

The documentation is organized hierarchically:

- [Complete Performance Study](.docs/study.md): Analysis of optimization techniques
- [Bottleneck Analysis](.docs/analysis/bottlenecks.md): Profiling results and identified issues
- [Performance Benchmarks](.docs/benchmarks/performance.md): Detailed measurements
- [Optimization Techniques](.docs/optimization/techniques.md): Implementation strategies
- [Performance Verification](.docs/proof-of-concept/performance_analysis_report.md): Independent verification of performance claims

## Build Commands

```bash
# Build original and optimized versions
make build
make build_optimized  

# Environment setup
make start_env

# Run database setup
docker exec -i postgres psql -U test -d test < env/data/postgres/00-schema.sql
```
