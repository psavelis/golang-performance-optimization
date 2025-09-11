# Profiling Tools Overview

Go provides a comprehensive suite of profiling tools that enable developers to analyze application performance, identify bottlenecks, and optimize resource usage. This chapter introduces the profiling ecosystem and when to use each tool.

## Go Profiling Ecosystem

### Core Profiling Tools

```
┌─────────────────────────────────────────────────────────────┐
│                   Go Profiling Stack                       │
├─────────────────────────────────────────────────────────────┤
│ Analysis Tools  │ Collection     │ Profile Types           │
│                 │                │                         │
│ • go tool pprof │ • net/http/    │ • CPU profiling         │
│ • go tool trace │   pprof        │ • Memory profiling      │
│ • benchstat     │ • runtime/     │ • Goroutine profiling   │
│ • pprof web UI  │   pprof        │ • Block profiling       │
│ • Flame graphs  │ • testing      │ • Mutex profiling       │
│                 │ • Manual       │ • Trace profiling       │
└─────────────────────────────────────────────────────────────┘
```

### Profile Types Comparison

| Profile Type | What it Measures | When to Use | Collection Overhead |
|--------------|------------------|-------------|-------------------|
| **CPU** | CPU time spent in functions | High CPU usage, slow functions | Low (1-5%) |
| **Memory** | Heap allocations and usage | Memory leaks, high allocation | Medium (5-10%) |
| **Goroutine** | Goroutine states and stacks | Goroutine leaks, blocking | Low (1-2%) |
| **Block** | Blocking on synchronization | Lock contention, channel blocks | Medium (10-20%) |
| **Mutex** | Lock contention events | Mutex bottlenecks | High (20-40%) |
| **Trace** | Complete execution timeline | Complex concurrency issues | High (20-50%) |

## Profile Collection Methods

### 1. HTTP Endpoint (Production-Ready)

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"  // Registers /debug/pprof endpoints
    "time"
)

func main() {
    // Start profiling server
    go func() {
        // Only expose on localhost in production
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your application code
    runApplication()
}

// Access profiles at:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/profile?seconds=30
// http://localhost:6060/debug/pprof/heap
// http://localhost:6060/debug/pprof/goroutine
```

### 2. Direct File Output

```go
package main

import (
    "os"
    "runtime/pprof"
    "runtime/trace"
)

func profileToFile() {
    // CPU profiling
    cpuFile, err := os.Create("cpu.prof")
    if err != nil {
        panic(err)
    }
    defer cpuFile.Close()
    
    pprof.StartCPUProfile(cpuFile)
    defer pprof.StopCPUProfile()
    
    // Memory profiling
    memFile, err := os.Create("mem.prof")
    if err != nil {
        panic(err)
    }
    defer func() {
        runtime.GC() // Get up-to-date statistics
        pprof.WriteHeapProfile(memFile)
        memFile.Close()
    }()
    
    // Execution trace
    traceFile, err := os.Create("trace.out")
    if err != nil {
        panic(err)
    }
    defer traceFile.Close()
    
    trace.Start(traceFile)
    defer trace.Stop()
    
    // Run your application code
    runApplication()
}
```

### 3. Test-Integrated Profiling

```go
package main

import (
    "flag"
    "os"
    "runtime/pprof"
    "testing"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to file")

func TestWithProfiling(t *testing.T) {
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            t.Fatal(err)
        }
        defer f.Close()
        
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    
    // Run your test
    runBenchmarkWorkload()
    
    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            t.Fatal(err)
        }
        defer f.Close()
        
        runtime.GC()
        pprof.WriteHeapProfile(f)
    }
}

// Run with: go test -cpuprofile=cpu.prof -memprofile=mem.prof
```

### 4. Conditional Profiling

```go
package main

import (
    "context"
    "os"
    "runtime/pprof"
    "sync"
    "time"
)

type Profiler struct {
    mu       sync.Mutex
    active   bool
    cpuFile  *os.File
    profiles map[string]*os.File
}

func NewProfiler() *Profiler {
    return &Profiler{
        profiles: make(map[string]*os.File),
    }
}

func (p *Profiler) StartCPU(filename string) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.active {
        return nil // Already profiling
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    
    p.cpuFile = f
    p.active = true
    return pprof.StartCPUProfile(f)
}

func (p *Profiler) StopCPU() {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if !p.active {
        return
    }
    
    pprof.StopCPUProfile()
    p.cpuFile.Close()
    p.active = false
}

func (p *Profiler) WriteProfile(name, filename string) error {
    profile := pprof.Lookup(name)
    if profile == nil {
        return fmt.Errorf("profile %s not found", name)
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    return profile.WriteTo(f, 0)
}

// Usage with signal-based profiling
func signalBasedProfiling() {
    profiler := NewProfiler()
    
    // Start profiling on SIGUSR1
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGUSR1)
    
    go func() {
        for range c {
            if profiler.active {
                profiler.StopCPU()
                profiler.WriteProfile("heap", "heap.prof")
                profiler.WriteProfile("goroutine", "goroutine.prof")
                log.Println("Profiling stopped and written to files")
            } else {
                timestamp := time.Now().Format("20060102_150405")
                profiler.StartCPU(fmt.Sprintf("cpu_%s.prof", timestamp))
                log.Println("Profiling started")
            }
        }
    }()
    
    // Your application code
    runApplication()
}
```

## Analysis Workflow

### Basic Analysis Commands

```bash
# CPU profile analysis
go tool pprof cpu.prof
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile analysis
go tool pprof mem.prof
go tool pprof http://localhost:6060/debug/pprof/heap

# Interactive commands in pprof:
# (pprof) top10          # Show top 10 functions
# (pprof) list main.main # Show source code with annotations
# (pprof) web            # Generate web-based visualization
# (pprof) pdf            # Generate PDF call graph
# (pprof) help           # Show all commands

# Trace analysis
go tool trace trace.out

# Compare profiles
go tool pprof -base=baseline.prof current.prof
```

### Advanced Analysis Techniques

```go
// Custom profile collection for analysis
func advancedProfiling() {
    // Collect baseline
    baseline := collectProfile("heap")
    
    // Run test workload
    runWorkload()
    
    // Collect after workload
    after := collectProfile("heap")
    
    // Programmatic analysis
    analyzeProfileDifference(baseline, after)
}

func collectProfile(profileType string) *profile.Profile {
    var buf bytes.Buffer
    
    switch profileType {
    case "heap":
        runtime.GC()
        pprof.WriteHeapProfile(&buf)
    case "cpu":
        // CPU profiling requires time duration
        pprof.StartCPUProfile(&buf)
        time.Sleep(10 * time.Second)
        pprof.StopCPUProfile()
    case "goroutine":
        pprof.Lookup("goroutine").WriteTo(&buf, 1)
    }
    
    p, err := profile.Parse(&buf)
    if err != nil {
        log.Fatal(err)
    }
    return p
}

func analyzeProfileDifference(baseline, current *profile.Profile) {
    // This is a simplified example
    // Real analysis would use the profile package
    
    baselineAllocs := getTotalAllocations(baseline)
    currentAllocs := getTotalAllocations(current)
    
    difference := currentAllocs - baselineAllocs
    
    fmt.Printf("Allocation difference: %d bytes\n", difference)
    fmt.Printf("Growth factor: %.2fx\n", float64(currentAllocs)/float64(baselineAllocs))
}
```

## Profiling in Different Environments

### Development Environment

```go
// Development profiling setup
func developmentProfiling() {
    // Always enable pprof endpoints in development
    go func() {
        log.Println("Starting profiling server on :6060")
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Add profiling hooks to critical paths
    profileCriticalPath := func(name string, fn func()) {
        start := time.Now()
        fn()
        duration := time.Since(start)
        
        if duration > 100*time.Millisecond {
            log.Printf("Slow operation %s: %v", name, duration)
        }
    }
    
    // Use profiling wrapper
    profileCriticalPath("database_query", func() {
        // Database operation
    })
}
```

### Production Environment

```go
// Production-safe profiling
func productionProfiling() {
    // Only enable profiling with environment variable
    if os.Getenv("ENABLE_PROFILING") == "true" {
        // Bind to localhost only for security
        go func() {
            log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
        }()
    }
    
    // Implement sampling for low overhead
    profileSampler := &ProfileSampler{
        SampleRate: 0.01, // 1% of requests
    }
    
    http.HandleFunc("/api/endpoint", func(w http.ResponseWriter, r *http.Request) {
        if profileSampler.ShouldProfile() {
            // Profile this request
            ctx := profileSampler.StartProfiling(r.Context())
            defer profileSampler.StopProfiling(ctx)
        }
        
        // Handle request normally
        handleRequest(w, r)
    })
}

type ProfileSampler struct {
    SampleRate float64
    mu         sync.Mutex
    counter    uint64
}

func (ps *ProfileSampler) ShouldProfile() bool {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    
    ps.counter++
    return (float64(ps.counter)*ps.SampleRate) >= 1.0
}
```

### Testing Environment

```go
// Benchmark with profiling
func BenchmarkWithProfiling(b *testing.B) {
    // CPU profiling
    if testing.Short() {
        b.Skip("Skipping profiling in short mode")
    }
    
    // Memory profiling
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark code
        runBenchmarkCode()
    }
    b.StopTimer()
    
    runtime.ReadMemStats(&m2)
    
    // Report allocations
    allocations := m2.TotalAlloc - m1.TotalAlloc
    b.ReportMetric(float64(allocations)/float64(b.N), "allocs/op")
    
    mallocs := m2.Mallocs - m1.Mallocs
    b.ReportMetric(float64(mallocs)/float64(b.N), "B/op")
}

// Integration test profiling
func TestIntegrationWithProfiling(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test profiling")
    }
    
    // Start background profiling
    profiler := startBackgroundProfiler()
    defer profiler.Stop()
    
    // Run integration test
    runIntegrationTest(t)
    
    // Analyze results
    results := profiler.GetResults()
    if results.MaxMemoryUsage > threshold {
        t.Errorf("Memory usage too high: %d bytes", results.MaxMemoryUsage)
    }
}
```

## Tool Selection Guide

### When to Use Each Tool

#### **CPU Profiling** 🔥
- **Use when**: High CPU usage, slow response times
- **Best for**: Algorithm optimization, hot path identification
- **Overhead**: 1-5%
- **Duration**: 10-60 seconds

```go
// CPU profiling example
func cpuProfilingExample() {
    // Start CPU profiling
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // CPU-intensive work
    for i := 0; i < 1000000; i++ {
        processItem(i)
    }
}
```

#### **Memory Profiling** 🧠
- **Use when**: High memory usage, suspected leaks
- **Best for**: Allocation optimization, memory leak detection
- **Overhead**: 5-10%
- **Duration**: Point-in-time snapshot

```go
// Memory profiling example
func memoryProfilingExample() {
    // Allocate memory
    data := make([][]byte, 1000)
    for i := range data {
        data[i] = make([]byte, 1024)
    }
    
    // Take memory snapshot
    f, _ := os.Create("mem.prof")
    defer f.Close()
    runtime.GC()
    pprof.WriteHeapProfile(f)
    
    // Keep reference to prevent GC
    _ = data
}
```

#### **Goroutine Profiling** 🔄
- **Use when**: Suspected goroutine leaks, deadlocks
- **Best for**: Concurrency issue diagnosis
- **Overhead**: 1-2%
- **Duration**: Point-in-time snapshot

```go
// Goroutine profiling example
func goroutineProfilingExample() {
    // Create many goroutines
    for i := 0; i < 1000; i++ {
        go func(id int) {
            time.Sleep(time.Hour) // Simulate long-running
        }(i)
    }
    
    // Profile goroutines
    f, _ := os.Create("goroutine.prof")
    defer f.Close()
    pprof.Lookup("goroutine").WriteTo(f, 1)
}
```

#### **Block Profiling** 🚫
- **Use when**: Suspected blocking on channels/mutexes
- **Best for**: Synchronization bottleneck identification
- **Overhead**: 10-20%
- **Duration**: Enable before running load

```go
// Block profiling example
func blockProfilingExample() {
    // Enable block profiling
    runtime.SetBlockProfileRate(1)
    defer runtime.SetBlockProfileRate(0)
    
    // Create blocking scenario
    ch := make(chan int)
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch <- 1
    }()
    
    // This will block and be recorded
    <-ch
    
    // Profile blocks
    f, _ := os.Create("block.prof")
    defer f.Close()
    pprof.Lookup("block").WriteTo(f, 0)
}
```

#### **Mutex Profiling** 🔒
- **Use when**: Lock contention suspected
- **Best for**: Mutex bottleneck analysis
- **Overhead**: 20-40%
- **Duration**: Enable before running concurrent load

```go
// Mutex profiling example
func mutexProfilingExample() {
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
    defer runtime.SetMutexProfileFraction(0)
    
    var mu sync.Mutex
    var counter int
    
    // Create contention
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                mu.Lock()
                counter++
                time.Sleep(time.Microsecond) // Increase contention
                mu.Unlock()
            }
        }()
    }
    
    wg.Wait()
    
    // Profile mutex contention
    f, _ := os.Create("mutex.prof")
    defer f.Close()
    pprof.Lookup("mutex").WriteTo(f, 0)
}
```

#### **Execution Tracing** 📊
- **Use when**: Complex concurrency issues, scheduling problems
- **Best for**: Complete system behavior analysis
- **Overhead**: 20-50%
- **Duration**: Short bursts (1-10 seconds)

```go
// Trace profiling example
func traceProfilingExample() {
    f, _ := os.Create("trace.out")
    defer f.Close()
    
    trace.Start(f)
    defer trace.Stop()
    
    // Complex concurrent workload
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                processTask(id, j)
                if j%10 == 0 {
                    runtime.Gosched() // Yield
                }
            }
        }(i)
    }
    
    wg.Wait()
}
```

## Best Practices for Profiling

### ✅ **Do's**

1. **Profile in realistic conditions**
   ```go
   // Use production-like data and load
   ```

2. **Collect multiple samples**
   ```go
   // Take several profiles to identify patterns
   ```

3. **Focus on the biggest opportunities**
   ```go
   // Optimize the top 10% of hot spots first
   ```

4. **Measure before and after changes**
   ```go
   // Always benchmark your optimizations
   ```

5. **Use the right tool for the problem**
   ```go
   // CPU for speed, memory for allocations, etc.
   ```

### ❌ **Don'ts**

1. **Don't profile in debug mode**
   ```go
   // Always use optimized builds: go build -o app
   ```

2. **Don't enable all profiling types simultaneously**
   ```go
   // High overhead can distort results
   ```

3. **Don't profile for too short a duration**
   ```go
   // CPU profiles need 10+ seconds for accuracy
   ```

4. **Don't ignore the baseline**
   ```go
   // Always compare against expected performance
   ```

5. **Don't optimize without profiling**
   ```go
   // Profile first, then optimize based on data
   ```

## Integration with CI/CD

```go
// Automated performance regression detection
func benchmarkRegression() {
    // Run benchmark with profiling
    cmd := exec.Command("go", "test", "-bench=.", "-cpuprofile=cpu.prof", "-memprofile=mem.prof")
    output, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }
    
    // Parse benchmark results
    results := parseBenchmarkOutput(string(output))
    
    // Compare with baseline
    baseline := loadBaseline()
    if results.NsPerOp > baseline.NsPerOp*1.1 { // 10% regression threshold
        log.Fatalf("Performance regression detected: %v vs %v", results.NsPerOp, baseline.NsPerOp)
    }
    
    // Update baseline if improvement
    if results.NsPerOp < baseline.NsPerOp*0.95 { // 5% improvement
        saveBaseline(results)
        log.Printf("Performance improvement: %v -> %v", baseline.NsPerOp, results.NsPerOp)
    }
}
```

Understanding the profiling tools and when to use them is essential for effective performance optimization. Each tool provides unique insights into different aspects of your application's behavior.

---

**Next**: [CPU Profiling](cpu-profiling.md) - Deep dive into CPU performance analysis
