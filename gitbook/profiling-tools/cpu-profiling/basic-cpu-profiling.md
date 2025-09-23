# Basic CPU Profiling

CPU profiling is the foundation of performance optimization in Go. This guide covers the essential techniques for identifying CPU bottlenecks and understanding where your application spends its computational resources.

## Introduction to CPU Profiling

CPU profiling captures samples of your program's execution to identify which functions consume the most CPU time. Go's built-in profiler uses statistical sampling to provide insights without significantly impacting performance.

### Key Concepts

- **Sampling Rate**: Go samples the stack approximately 100 times per second by default
- **Statistical Accuracy**: Longer profiling sessions provide more accurate results
- **Overhead**: CPU profiling adds minimal overhead (typically <5%)
- **Resolution**: Can identify hotspots down to individual lines of code

## Enabling CPU Profiling

### Using pprof Package

```go
package main

import (
    "fmt"
    "os"
    "runtime/pprof"
    "time"
)

func main() {
    // Create CPU profile file
    f, err := os.Create("cpu.prof")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    // Start CPU profiling
    if err := pprof.StartCPUProfile(f); err != nil {
        panic(err)
    }
    defer pprof.StopCPUProfile()

    // Your application code here
    performWork()
}

func performWork() {
    // Simulate CPU-intensive work
    for i := 0; i < 1000000; i++ {
        _ = fibonacci(30)
    }
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
```

### Using net/http/pprof

For web applications, use the HTTP profiling endpoint:

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Add profiling endpoints to default mux
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Your web application code
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Simulate work
    result := expensiveOperation()
    fmt.Fprintf(w, "Result: %d", result)
}

func expensiveOperation() int {
    sum := 0
    for i := 0; i < 100000; i++ {
        sum += i * i
    }
    return sum
}
```

## Collecting CPU Profiles

### Command Line Profiling

```bash
# Profile a running application for 30 seconds
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Profile from saved file
go tool pprof cpu.prof

# Save profile for later analysis
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
```

### Programmatic Collection

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "runtime/pprof"
    "time"
)

func collectCPUProfile(duration time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), duration)
    defer cancel()

    // Get profile from HTTP endpoint
    resp, err := http.Get("http://localhost:6060/debug/pprof/profile?seconds=30")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Parse profile
    profile, err := pprof.Parse(resp.Body)
    if err != nil {
        return err
    }

    fmt.Printf("Profile contains %d samples\n", len(profile.Sample))
    return nil
}
```

## Analyzing CPU Profiles

### Interactive Analysis

```bash
# Start interactive pprof session
go tool pprof cpu.prof

# Commands within pprof:
(pprof) top           # Show top functions by CPU usage
(pprof) top10         # Show top 10 functions
(pprof) list main     # Show source code for main function
(pprof) web           # Generate web visualization
(pprof) svg           # Generate SVG flamegraph
(pprof) traces        # Show sample traces
```

### Common pprof Commands

#### Top Command
```bash
(pprof) top
Showing nodes accounting for 2.34s, 78.52% of 2.98s total
Dropped 15 nodes (cum <= 0.01s)
      flat  flat%   sum%        cum   cum%
     1.20s 40.27% 40.27%      1.20s 40.27%  main.fibonacci
     0.64s 21.48% 61.75%      1.84s 61.75%  main.performWork
     0.32s 10.74% 72.49%      0.32s 10.74%  runtime.usleep
     0.18s  6.04% 78.52%      0.18s  6.04%  runtime.pthread_cond_signal
```

#### List Command
```bash
(pprof) list fibonacci
Total: 2.98s
ROUTINE ======================== main.fibonacci in /main.go
     1.20s      1.20s (flat, cum) 40.27% of Total
         .          .     23:func fibonacci(n int) int {
         .          .     24:    if n <= 1 {
      0.12s      0.12s     25:        return n
         .          .     26:    }
     1.08s      1.08s     27:    return fibonacci(n-1) + fibonacci(n-2)
         .          .     28:}
```

## Understanding Profile Output

### Metrics Explanation

- **flat**: CPU time spent in this function only
- **flat%**: Percentage of total CPU time spent in this function
- **sum%**: Cumulative percentage up to this line
- **cum**: CPU time spent in this function and its callees
- **cum%**: Percentage of total CPU time including callees

### Sample Profile Analysis

```go
package main

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "os"
    "runtime/pprof"
)

func main() {
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Different workload patterns
    cpuIntensive()
    memoryIntensive()
    mixedWorkload()
}

func cpuIntensive() {
    // Pure computation
    for i := 0; i < 1000000; i++ {
        _ = isPrime(i)
    }
}

func memoryIntensive() {
    // Allocation-heavy
    data := make([][]int, 1000)
    for i := range data {
        data[i] = make([]int, 1000)
        for j := range data[i] {
            data[i][j] = i * j
        }
    }
}

func mixedWorkload() {
    // Mixed CPU and memory operations
    for i := 0; i < 100; i++ {
        n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
        if isPrime(int(n.Int64())) {
            data := make([]int, 1000)
            for j := range data {
                data[j] = int(n.Int64()) * j
            }
        }
    }
}

func isPrime(n int) bool {
    if n < 2 {
        return false
    }
    for i := 2; i*i <= n; i++ {
        if n%i == 0 {
            return false
        }
    }
    return true
}
```

## Best Practices

### 1. Profile Production-Like Conditions

```go
// Good: Profile under realistic load
func benchmarkWithLoad() {
    // Simulate production traffic patterns
    for i := 0; i < 1000; i++ {
        go handleRequest()
    }
    
    // Profile during steady state
    time.Sleep(10 * time.Second)
}

// Bad: Profile with artificial workload
func benchmarkArtificial() {
    for i := 0; i < 1000000; i++ {
        // Unrealistic tight loop
    }
}
```

### 2. Sufficient Profiling Duration

```go
// Good: Long enough for statistical significance
duration := 30 * time.Second
if isProduction {
    duration = 5 * time.Minute  // Longer for production
}

// Bad: Too short for meaningful results
duration := 1 * time.Second
```

### 3. Profile Comparison

```go
func compareOptimizations() {
    // Profile before optimization
    profileBefore := collectProfile("before.prof")
    
    // Apply optimization
    optimizeCode()
    
    // Profile after optimization
    profileAfter := collectProfile("after.prof")
    
    // Compare profiles
    // go tool pprof -base before.prof after.prof
}
```

## Common CPU Profiling Patterns

### Hot Loop Detection

```go
func detectHotLoops() {
    // This will show up prominently in CPU profile
    for i := 0; i < 1000000; i++ {
        expensiveFunction()  // Hot spot
    }
}
```

### Recursive Function Analysis

```go
func recursiveAnalysis(n int) int {
    // CPU profiler will show recursive call patterns
    if n <= 1 {
        return n
    }
    return recursiveAnalysis(n-1) + recursiveAnalysis(n-2)
}
```

### Goroutine CPU Usage

```go
func goroutineCPUUsage() {
    // Profile will aggregate CPU usage across goroutines
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 100000; j++ {
                _ = j * j
            }
        }()
    }
}
```

## Troubleshooting CPU Profiles

### Empty or Minimal Profiles

```go
// Ensure sufficient CPU work
func enoughWork() {
    // Too little work won't show in profile
    for i := 0; i < 10; i++ {
        _ = i * 2
    }
    
    // Sufficient work for profiling
    for i := 0; i < 1000000; i++ {
        _ = complexCalculation(i)
    }
}
```

### Profile Accuracy Issues

```bash
# Increase sampling rate for more accuracy
CPUPROFILE_HZ=500 go run main.go

# Default is 100 Hz, higher values give more accuracy but more overhead
```

### Missing Stack Information

```go
// Ensure debug symbols are available
go build -ldflags="-s -w" main.go  // Bad: strips symbols
go build main.go                   // Good: keeps symbols
```

## Integration with Testing

```go
func BenchmarkWithCPUProfile(b *testing.B) {
    f, _ := os.Create("bench.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        functionToOptimize()
    }
}
```

## Next Steps

- Learn [Advanced CPU Profiling Techniques](advanced-techniques.md)
- Explore [Flamegraph Analysis](flamegraph-analysis.md)
- Understand [Memory Profiling](../memory-profiling/README.md)

## Summary

Basic CPU profiling provides essential insights into where your Go application spends its computational resources. Key takeaways:

1. Use `runtime/pprof` for programmatic profiling
2. Use `net/http/pprof` for web applications
3. Profile for sufficient duration under realistic load
4. Focus on functions with high flat CPU usage
5. Compare profiles before and after optimizations
6. Integrate profiling into your development workflow

Master these fundamentals before moving to advanced profiling techniques and specialized analysis tools.
