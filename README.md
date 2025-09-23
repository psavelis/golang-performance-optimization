# Go Event Processing - Performance Optimization

Event generation and database loading system with documented performance improvements through systematic optimization.

## Performance Results

| Implementation | Time (100K events) | Memory | Allocations | Improvement |
|---------------|-------------------|---------|-------------|-------------|
| **Original** | 798ms | 4.85MB | 187,602 | 1.0x |
| **Optimized** | 149ms | 1.52MB | 3,033 | **5.35x** |
| **Streaming** | 199ms | 3.95MB | 189,541 | 4.01x |

**Primary Optimization**: String pool elimination of concatenation overhead

## Quick Start

```bash
# Build implementations
make build_enhanced_profiling

# Performance comparison
target/build/bin/generator-profiling 100000 baseline.json        # ~800ms
target/build/bin/generator-enhanced-profiling 100000 optimal.json # ~150ms

# Benchmark validation
go test -bench=BenchmarkGenerator -benchmem ./pkg/benchmarks/
```

## Core Optimizations

### String Pool Implementation
```go
// Pre-allocated string array eliminates runtime allocation
var stringPool = [48]string{"ABC123", "XYZ456", "LOC789", /*...*/}

func getRandomStringFromPool() string {
    return stringPool[rand.Intn(len(stringPool))] // O(1), zero allocation
}
```

**Measured Impact**: 86x faster string generation (674ns → 7.8ns per operation)

### Streaming JSON Architecture  
```go
// Constant memory usage regardless of dataset size
func streamEventsToFile(file *os.File, count int) error {
    writer := bufio.NewWriter(file)
    defer writer.Flush()
    
    writer.WriteString("[")
    for i := 0; i < count; i++ {
        if i > 0 { writer.WriteString(",") }
        event := createEvent()
        data, _ := json.Marshal(event)
        writer.Write(data)
        if i%100 == 0 { writer.Flush() }
    }
    writer.WriteString("]")
    return nil
}
```

**Measured Impact**: Constant 3.95MB memory usage vs. linear growth

## Available Commands

```bash
make benchmark_compare               # Statistical benchmark comparison
make profile_enhanced               # CPU, memory, block profiling
make generate_enhanced_flamegraphs  # Generate SVG visualizations
make start_env                      # Start PostgreSQL environment
```

## Documentation

- **[Technical specification](.docs/technical-specification-v1.md)**: Implementation details
- **[Flamegraph analysis](.docs/flamegraph-analysis-summary.md)**: Visual profiling analysis
- **[Documentation index](.docs/documentation-index.md)**: Complete documentation guide
- **[Pyroscope tutorial](.docs/pyroscope-tutorial.md)**: Continuous profiling with Grafana Pyroscope

## System Requirements

- Go 1.24+
- 8GB RAM (recommended for large datasets)
- PostgreSQL (for database loader testing)
- Docker (for local Pyroscope + Grafana stack)

## Architecture

### Event Generator
- Generates random events with configurable count
- Event types distributed as: 15% type 1, 20% type 2, 20% type 3, 45% type 5
- Outputs JSON format compatible with database schema

### Database Loader
- Reads JSON event files
- Inserts events using parameterized queries
- Supports PostgreSQL with proper schema

### Profiling Infrastructure
- CPU profiling with `runtime/pprof`
- Memory allocation tracking
- Block and mutex profiling for bottleneck detection
- Flamegraph generation for visual analysis

### Continuous Profiling (Grafana Pyroscope)
- Pyroscope server and Grafana provisioned via Docker Compose
- Go services instrumented with `github.com/grafana/pyroscope-go`
- Make targets to start stack and run instrumented binaries

Quick start:

```bash
make start_profiling_stack
make build
make run_pyroscope_generator
make run_pyroscope_loader
# Pyroscope: http://localhost:4040, Grafana: http://localhost:3000
```

### Metrics (Prometheus)

```bash
make start_observability_stack
make build
METRICS_ADDR=:2112 METRICS_HOLD_FOR=60s target/build/bin/generator-prometheus 100000 test_prom.json
METRICS_ADDR=:2113 METRICS_HOLD_FOR=60s target/build/bin/loader-prometheus 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_prom.json
# Prometheus: http://localhost:9090, Grafana: http://localhost:3000
```

### Distributed Tracing (OpenTelemetry + Tempo)

```bash
make start_observability_stack   # includes otel-collector and tempo
make build
make run_otel_generator
make run_otel_loader
# Tempo API: http://localhost:3200, Grafana: http://localhost:3000
```

---

**Test Environment**: Apple M1 Pro, Go 1.24, macOS
**Validation**: Statistical significance with 95% confidence intervals
