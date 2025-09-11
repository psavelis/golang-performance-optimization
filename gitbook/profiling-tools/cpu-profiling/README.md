# CPU Profiling

CPU profiling is your primary tool for identifying performance bottlenecks and optimizing execution time. This section covers everything from basic profiling to advanced analysis techniques.

## Overview

CPU profiling samples your program's execution to identify where time is spent. By understanding which functions consume the most CPU cycles, you can focus optimization efforts where they'll have the greatest impact.

### What CPU Profiling Reveals

- **Hot functions**: Code consuming the most execution time
- **Call patterns**: How functions call each other
- **Execution paths**: Which code paths are most frequently executed
- **Algorithm efficiency**: Relative performance of different approaches
- **System overhead**: Runtime, GC, and system call costs

## Quick Start

### Enable Profiling in Your Application

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"  // Import for side effects
)

func main() {
    // Start HTTP server for profiling
    go func() {
        log.Println("Profiling server at http://localhost:6060/debug/pprof/")
        log.Fatal(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your application code
    runYourApplication()
}
```

### Collect CPU Profile

```bash
# Live profiling (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# From benchmark
go test -bench=. -cpuprofile=cpu.prof

# Programmatic profiling
go run main.go  # With profiling code embedded
```

### Analyze Profile

```bash
# Interactive analysis
go tool pprof cpu.prof

# Web interface (recommended)
go tool pprof -http=:8080 cpu.prof

# Generate report
go tool pprof -top cpu.prof
```

## Profiling Techniques

### 1. Basic CPU Profiling

The foundation of CPU analysis:

**[→ Basic CPU Profiling](basic-cpu-profiling.md)**
- Setting up profiling endpoints
- Collecting your first CPU profile
- Understanding basic output
- Identifying obvious bottlenecks

### 2. Advanced CPU Techniques

Sophisticated analysis methods:

**[→ Advanced Techniques](advanced-techniques.md)**
- Differential profiling (before/after)
- Focus and ignore patterns
- Custom sampling rates
- Production profiling strategies

### 3. Flamegraph Analysis

Visual profiling for complex applications:

**[→ Flamegraph Analysis](flamegraph-analysis.md)**
- Generating interactive flamegraphs
- Reading flamegraph visualizations
- Identifying optimization opportunities
- Comparing multiple profiles

## Common CPU Optimization Patterns

### String Operations

String manipulation often appears in CPU profiles:

```go
// ❌ Inefficient: Multiple allocations
func buildStringBad(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","  // Allocates new string each time
    }
    return result
}

// ✅ Efficient: Single allocation
func buildStringGood(items []string) string {
    var builder strings.Builder
    builder.Grow(len(items) * 10)  // Pre-allocate estimated size
    
    for i, item := range items {
        if i > 0 {
            builder.WriteByte(',')
        }
        builder.WriteString(item)
    }
    return builder.String()
}
```

### Algorithm Optimization

CPU profiles reveal algorithmic inefficiencies:

```go
// ❌ O(n²) lookup performance
func findItemsBad(items []Item, targets []string) []Item {
    var results []Item
    for _, target := range targets {
        for _, item := range items {  // Linear search for each target
            if item.Name == target {
                results = append(results, item)
                break
            }
        }
    }
    return results
}

// ✅ O(n) lookup performance  
func findItemsGood(items []Item, targets []string) []Item {
    // Pre-build map for O(1) lookups
    itemMap := make(map[string]Item, len(items))
    for _, item := range items {
        itemMap[item.Name] = item
    }
    
    results := make([]Item, 0, len(targets))
    for _, target := range targets {
        if item, exists := itemMap[target]; exists {
            results = append(results, item)
        }
    }
    return results
}
```

## CPU Profile Analysis Workflow

### 1. Initial Collection
```bash
# Collect profile during realistic workload
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=60
```

### 2. High-Level Analysis
```bash
# Identify top CPU consumers
(pprof) top10
(pprof) top10 -cum  # Cumulative time including called functions
```

### 3. Function-Level Investigation
```bash
# Examine specific function
(pprof) list functionName
(pprof) disasm functionName  # Assembly-level analysis
```

### 4. Call Graph Analysis
```bash
# Understand call relationships
(pprof) web
(pprof) png > profile.png
```

### 5. Optimization and Validation
```bash
# Compare before/after profiles
go tool pprof -base=old.prof new.prof
(pprof) top10
```

## Production CPU Profiling

### Safe Production Profiling

```go
// Production-safe profiling setup
func setupProduction() {
    // Only enable profiling in debug mode
    if os.Getenv("DEBUG_PROFILING") == "true" {
        go func() {
            // Bind to localhost only for security
            log.Fatal(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}

// Rate-limited profiling
func conditionalProfiling() {
    // Profile 1% of requests
    if rand.Float64() < 0.01 {
        f, _ := os.Create(fmt.Sprintf("profile-%d.prof", time.Now().Unix()))
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
        defer f.Close()
        
        // Execute request
        handleRequest()
    }
}
```

### Automated Profile Collection

```bash
#!/bin/bash
# collect-cpu-profiles.sh

SERVICE_URL="http://localhost:6060"
DURATION=30
INTERVAL=300  # 5 minutes

while true; do
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    echo "Collecting CPU profile at $TIMESTAMP"
    
    go tool pprof -seconds=$DURATION -output="cpu-$TIMESTAMP.prof" \
        "$SERVICE_URL/debug/pprof/profile"
    
    # Generate quick analysis
    go tool pprof -top -output="analysis-$TIMESTAMP.txt" "cpu-$TIMESTAMP.prof"
    
    sleep $INTERVAL
done
```

## CPU Profiling Best Practices

### ✅ Do's

1. **Profile realistic workloads** - Use production-like data and traffic patterns
2. **Profile for sufficient duration** - 30-60 seconds minimum for statistical significance
3. **Compare before/after** - Always measure optimization impact
4. **Focus on the biggest gains** - Optimize functions consuming >5% of CPU time
5. **Consider the call stack** - Sometimes the issue is in the caller, not the called function

### ❌ Don'ts

1. **Don't profile toy examples** - Micro-benchmarks may not reflect real performance
2. **Don't ignore cumulative time** - Functions that call expensive operations matter
3. **Don't optimize prematurely** - Profile first, then optimize based on data
4. **Don't forget about memory** - CPU optimization that increases allocations may hurt overall performance
5. **Don't profile development builds** - Use optimized builds (`go build -ldflags="-s -w"`)

## CPU Profile Interpretation Guide

### Understanding pprof Output

```
(pprof) top10
Showing nodes accounting for 2840ms, 94.67% of 3000ms total
Dropped 23 nodes (cum <= 15ms)
Showing top 10 nodes out of 45
      flat  flat%   sum%        cum   cum%
    1200ms 40.00% 40.00%     1200ms 40.00%  main.expensiveFunction
     890ms 29.67% 69.67%      890ms 29.67%  main.stringManipulation
     350ms 11.67% 81.33%      350ms 11.67%  runtime.mallocgc
     210ms  7.00% 88.33%      250ms  8.33%  strings.(*Builder).Grow
     190ms  6.33% 94.67%      190ms  6.33%  runtime.memmove
```

**Column Meanings:**
- **flat**: Time spent in this function only (excluding calls to other functions)
- **flat%**: Percentage of total time spent in this function
- **sum%**: Cumulative percentage including all functions above
- **cum**: Cumulative time including this function and all functions it calls
- **cum%**: Cumulative percentage including called functions

### Red Flags in CPU Profiles

🚨 **High string manipulation costs**
```
strings.Join: 25% of CPU time
Solution: Use strings.Builder or bytes.Buffer
```

🚨 **Excessive reflection**
```
reflect.*: 20% of CPU time
Solution: Use code generation or type-specific methods
```

🚨 **Memory allocation overhead**
```
runtime.mallocgc: 15% of CPU time
Solution: Reduce allocations, use object pools
```

🚨 **JSON serialization bottlenecks**
```
encoding/json.(*encodeState).marshal: 30% of CPU time
Solution: Use faster JSON libraries or custom serialization
```

## Learning Path

### Beginner Level
- **[Basic CPU Profiling](basic-cpu-profiling.md)** - Essential concepts and first profile
- Practice with simple applications
- Learn to read top10 output

### Intermediate Level  
- **[Advanced Techniques](advanced-techniques.md)** - Differential analysis and filtering
- Production profiling setup
- Call graph analysis

### Advanced Level
- **[Flamegraph Analysis](flamegraph-analysis.md)** - Visual profiling mastery
- Custom sampling and automation
- Performance regression detection

## Tools and Resources

### Essential Commands
```bash
# Collection
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go test -bench=. -cpuprofile=cpu.prof

# Analysis  
go tool pprof -http=:8080 cpu.prof
go tool pprof -top cpu.prof
go tool pprof -list=functionName cpu.prof

# Comparison
go tool pprof -base=baseline.prof optimized.prof

# Export
go tool pprof -png cpu.prof > cpu-profile.png
go tool pprof -svg cpu.prof > cpu-profile.svg
```

### Integration Examples
```go
// Benchmark with CPU profiling
func BenchmarkFunction(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        functionToOptimize()
    }
}

// Conditional profiling in production
func profiledHandler(w http.ResponseWriter, r *http.Request) {
    if shouldProfile() {
        f, _ := os.Create("handler-profile.prof")
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
        defer f.Close()
    }
    
    // Handle request
    actualHandler(w, r)
}
```

Ready to master CPU profiling? Start with **[Basic CPU Profiling](basic-cpu-profiling.md)** and work your way through each technique systematically.

---

**Next Steps**: [Basic CPU Profiling](basic-cpu-profiling.md) → [Advanced Techniques](advanced-techniques.md) → [Flamegraph Analysis](flamegraph-analysis.md)
