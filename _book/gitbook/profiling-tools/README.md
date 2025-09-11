# Profiling Tools Overview

This section covers the complete ecosystem of Go profiling tools, from built-in runtime profilers to advanced third-party solutions. Master these tools to diagnose and optimize any performance issue.

## Tool Categories

### 🔧 Built-in Go Tools
- **`go tool pprof`** - Primary profiling interface
- **`go tool trace`** - Execution tracing and analysis
- **`go test -bench`** - Benchmarking framework
- **Runtime diagnostics** - Built-in performance monitoring

### 📊 Profile Types
- **[CPU Profiling](cpu-profiling/README.md)** - Function execution time analysis
- **[Memory Profiling](memory-profiling/README.md)** - Heap and allocation analysis  
- **[Goroutine Profiling](goroutine-profiling/README.md)** - Concurrency analysis
- **[Mutex & Block Profiling](concurrency-profiling/README.md)** - Synchronization analysis
- **[Custom Profiling](custom-profiling/README.md)** - Application-specific metrics

### 🌐 Third-party Tools
- **Flamegraph tools** - Visual profile analysis
- **Continuous profiling platforms** - Production monitoring
- **Performance testing frameworks** - Load and stress testing

## Quick Reference Guide

### Essential Commands

```bash
# Live profiling from running application
go tool pprof http://localhost:6060/debug/pprof/profile     # CPU
go tool pprof http://localhost:6060/debug/pprof/heap        # Memory
go tool pprof http://localhost:6060/debug/pprof/goroutine   # Goroutines

# Benchmark profiling
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

# Profile analysis
go tool pprof -http=:8080 cpu.prof                         # Web interface
go tool pprof cpu.prof                                      # Interactive CLI

# Execution tracing
go test -trace=trace.out
go tool trace trace.out
```

### Profile Collection Methods

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **Live HTTP** | Production monitoring | Real traffic, no restart | Network overhead |
| **Benchmark** | Development optimization | Repeatable, controlled | Synthetic workload |
| **Programmatic** | Custom integration | Full control | Requires code changes |
| **Signal-based** | Emergency debugging | No restart needed | Limited control |

## Tool Selection Guide

### Choose CPU Profiling When:
- ✅ Functions consuming excessive CPU time
- ✅ Algorithm optimization needed
- ✅ Hot path identification required
- ✅ Baseline performance measurement

### Choose Memory Profiling When:
- ✅ High memory usage or leaks suspected
- ✅ Garbage collection pressure
- ✅ Allocation pattern analysis needed
- ✅ Memory optimization opportunities

### Choose Goroutine Profiling When:
- ✅ Concurrency issues or deadlocks
- ✅ Goroutine leaks suspected
- ✅ Scheduler analysis needed
- ✅ Channel operation debugging

### Choose Blocking Profiling When:
- ✅ Lock contention suspected
- ✅ I/O bottlenecks present
- ✅ Synchronization issues
- ✅ Channel blocking analysis

## Advanced Profiling Workflows

### Multi-Profile Analysis
```bash
# Collect comprehensive profile set
go test -bench=BenchmarkCritical \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  -blockprofile=block.prof \
  -mutexprofile=mutex.prof \
  -trace=trace.out

# Analyze relationships between profiles
go tool pprof -http=:8080 cpu.prof &
go tool pprof -http=:8081 mem.prof &
go tool trace trace.out
```

### Production Profiling Pipeline
```bash
#!/bin/bash
# production-profile.sh

SERVICE_URL="http://production-service:6060"
PROFILE_DIR="profiles/$(date +%Y%m%d-%H%M%S)"

mkdir -p "$PROFILE_DIR"

# Collect all profile types
go tool pprof -seconds=30 -output="$PROFILE_DIR/cpu.prof" "$SERVICE_URL/debug/pprof/profile"
go tool pprof -output="$PROFILE_DIR/heap.prof" "$SERVICE_URL/debug/pprof/heap"
go tool pprof -output="$PROFILE_DIR/goroutine.prof" "$SERVICE_URL/debug/pprof/goroutine"
go tool pprof -output="$PROFILE_DIR/mutex.prof" "$SERVICE_URL/debug/pprof/mutex"

# Generate reports
go tool pprof -top -output="$PROFILE_DIR/cpu-top.txt" "$PROFILE_DIR/cpu.prof"
go tool pprof -top -output="$PROFILE_DIR/heap-top.txt" "$PROFILE_DIR/heap.prof"

echo "Profiles collected in $PROFILE_DIR"
```

## Tool Integration Patterns

### Development Workflow
```go
// Integrated profiling in development
func main() {
    if *profileFlag {
        defer profile.Start(
            profile.CPUProfile,
            profile.MemProfile,
            profile.ProfilePath("."),
        ).Stop()
    }
    
    // Application logic
    runApplication()
}
```

### Continuous Integration
```yaml
# .github/workflows/performance.yml
name: Performance Testing
on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.24
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem \
            -cpuprofile=cpu.prof \
            -memprofile=mem.prof \
            ./...
      
      - name: Analyze profiles
        run: |
          go tool pprof -top cpu.prof > cpu-analysis.txt
          go tool pprof -top mem.prof > mem-analysis.txt
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v2
        with:
          name: performance-profiles
          path: |
            *.prof
            *-analysis.txt
```

## Best Practices Summary

### ✅ Do's
- **Profile first, optimize second** - Measure before making changes
- **Use multiple profile types** - Get complete performance picture
- **Profile production workloads** - Real data reveals real issues
- **Automate profile collection** - Make profiling part of your workflow
- **Compare before/after** - Validate optimization effectiveness

### ❌ Don'ts
- **Don't guess at bottlenecks** - Always measure and verify
- **Don't profile toy examples** - Use realistic workloads
- **Don't ignore low-hanging fruit** - Address obvious issues first
- **Don't over-optimize** - Balance development time vs. performance gains
- **Don't forget production impact** - Consider profiling overhead

## Learning Path

### Beginner (2-4 hours)
1. **[Basic CPU Profiling](cpu-profiling/basic-cpu-profiling.md)**
2. **[Memory Profiling Basics](memory-profiling/heap-profiling.md)**
3. **Practice with sample applications**

### Intermediate (6-8 hours)
1. **[Advanced CPU Techniques](cpu-profiling/advanced-techniques.md)**
2. **[Goroutine Analysis](goroutine-profiling/goroutine-analysis.md)**
3. **[Blocking Operations](concurrency-profiling/mutex-contention.md)**
4. **Multi-profile correlation**

### Advanced (10+ hours)
1. **[Custom Profiling](custom-profiling/custom-profiles.md)**
2. **[Production Monitoring](custom-profiling/runtime-metrics.md)**
3. **[Flamegraph Analysis](cpu-profiling/flamegraph-analysis.md)**
4. **Performance testing integration**

## Tools Ecosystem

### Core Go Tools
```bash
# Built into Go toolchain
go tool pprof     # Profile analysis
go tool trace     # Execution tracing
go tool compile   # Compiler diagnostics
go tool objdump   # Assembly analysis
```

### Essential Extensions
```bash
# Install additional tools
go install github.com/google/pprof@latest
go install github.com/pkg/profile@latest
go install golang.org/x/tools/cmd/stress@latest

# Flamegraph generation
git clone https://github.com/brendangregg/FlameGraph.git
export PATH=$PATH:$PWD/FlameGraph
```

### Recommended Third-party
- **Grafana Pyroscope** - Continuous profiling platform
- **DataDog Profiler** - Commercial profiling service  
- **Go-torch** - Flamegraph generation (legacy)
- **Hey/AB** - HTTP load testing with profiling

## Next Steps

Ready to dive into specific profiling techniques? Choose your path:

- **🚀 Start with [CPU Profiling](cpu-profiling/README.md)** for immediate performance wins
- **🧠 Explore [Memory Profiling](memory-profiling/README.md)** for allocation optimization
- **⚡ Master [Goroutine Profiling](goroutine-profiling/README.md)** for concurrency issues
- **🔒 Learn [Concurrency Profiling](concurrency-profiling/README.md)** for synchronization problems

Each section builds comprehensive expertise in its profiling domain, with practical examples and production-ready techniques.

---

**Remember**: The best profiling strategy combines multiple tools and techniques to get a complete performance picture. Start simple, then layer on complexity as needed.
