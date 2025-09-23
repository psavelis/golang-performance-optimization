# Troubleshooting Guide

Comprehensive troubleshooting guide for common Go performance issues, diagnostic techniques, and resolution strategies.

## Common Performance Problems

### High CPU Usage

**Symptoms:**
- CPU utilization consistently >80%
- Slow response times
- High system load
- Fan noise on development machines

**Diagnostic Steps:**

1. **CPU Profile Analysis**
```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Or for applications without HTTP server
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

2. **Identify Hot Functions**
```bash
(pprof) top10
(pprof) list functionName
(pprof) web  # Visualize call graph
```

3. **Check for CPU-bound Loops**
```go
// BAD: Tight loop without yielding
func inefficientLoop() {
    for {
        // Intensive computation
        result := expensiveCalculation()
        if result > threshold {
            break
        }
    }
}

// GOOD: Yield to scheduler periodically
func efficientLoop() {
    for {
        result := expensiveCalculation()
        if result > threshold {
            break
        }
        runtime.Gosched() // Yield to other goroutines
    }
}
```

**Common Causes & Solutions:**

| Problem | Cause | Solution |
|---------|-------|----------|
| Infinite loops | Logic errors | Add proper exit conditions |
| Inefficient algorithms | O(n²) or worse complexity | Optimize algorithm complexity |
| JSON parsing overhead | Large payloads | Use streaming parsers |
| Regular expression compilation | Repeated compilation | Compile once, reuse |
| String concatenation | Repeated string building | Use strings.Builder |

### Memory Leaks

**Symptoms:**
- Memory usage grows continuously
- Out-of-memory errors
- Increased GC frequency
- Performance degradation over time

**Diagnostic Steps:**

1. **Heap Profile Comparison**
```bash
# Take baseline heap profile
curl http://localhost:6060/debug/pprof/heap > heap1.prof

# Wait and take another profile
sleep 300
curl http://localhost:6060/debug/pprof/heap > heap2.prof

# Compare profiles
go tool pprof -base heap1.prof heap2.prof
```

2. **Memory Growth Analysis**
```go
package main

import (
    "runtime"
    "time"
)

func monitorMemory() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            fmt.Printf("Alloc = %d KB", m.Alloc/1024)
            fmt.Printf("Sys = %d KB", m.Sys/1024)
            fmt.Printf("NumGC = %d\n", m.NumGC)
            
            // Alert on significant growth
            if m.Alloc > 100*1024*1024 { // 100MB
                fmt.Println("WARNING: High memory usage detected")
            }
        }
    }
}
```

**Common Memory Leak Patterns:**

1. **Goroutine Leaks**
```go
// BAD: Goroutine never exits
func leakyGoroutine() {
    go func() {
        for {
            select {
            case <-someChannel:
                // Process data
            // Missing case for shutdown signal
            }
        }
    }()
}

// GOOD: Proper cleanup
func properGoroutine(ctx context.Context) {
    go func() {
        for {
            select {
            case <-someChannel:
                // Process data
            case <-ctx.Done():
                return // Exit goroutine
            }
        }
    }()
}
```

2. **Slice Reference Leaks**
```go
// BAD: Keeps reference to entire array
func leakySlice(data []byte) []byte {
    return data[100:200] // Still references original array
}

// GOOD: Copy to new slice
func properSlice(data []byte) []byte {
    result := make([]byte, 100)
    copy(result, data[100:200])
    return result
}
```

3. **Map Growth Leaks**
```go
// BAD: Map grows without bounds
var cache = make(map[string][]byte)

func badCache(key string, value []byte) {
    cache[key] = value // Never cleaned up
}

// GOOD: LRU cache with size limit
type LRUCache struct {
    maxSize int
    data    map[string]*list.Element
    order   *list.List
}

func (c *LRUCache) Put(key string, value []byte) {
    if len(c.data) >= c.maxSize {
        c.evictOldest()
    }
    // Add new entry
}
```

### Goroutine Problems

**Symptoms:**
- Exponentially growing goroutine count
- Deadlocks
- Resource exhaustion
- Panic due to too many goroutines

**Diagnostic Commands:**
```bash
# Monitor goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Get goroutine stack traces
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Goroutine profile analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Common Issues:**

1. **Goroutine Leaks**
```go
// Diagnostic function
func checkGoroutineLeaks() {
    initial := runtime.NumGoroutine()
    
    // Run your code
    runSomeCode()
    
    // Check if goroutines increased
    final := runtime.NumGoroutine()
    if final > initial {
        fmt.Printf("Goroutine leak detected: %d -> %d\n", initial, final)
        
        // Get stack traces
        buf := make([]byte, 1<<16)
        stackSize := runtime.Stack(buf, true)
        fmt.Printf("Stack traces:\n%s\n", buf[:stackSize])
    }
}
```

2. **Deadlocks**
```go
// BAD: Potential deadlock
func deadlockExample() {
    var mu1, mu2 sync.Mutex
    
    go func() {
        mu1.Lock()
        time.Sleep(time.Millisecond)
        mu2.Lock() // Can deadlock here
        mu2.Unlock()
        mu1.Unlock()
    }()
    
    go func() {
        mu2.Lock()
        time.Sleep(time.Millisecond)
        mu1.Lock() // Can deadlock here
        mu1.Unlock()
        mu2.Unlock()
    }()
}

// GOOD: Consistent lock ordering
func deadlockFree() {
    var mu1, mu2 sync.Mutex
    
    lockPair := func() {
        mu1.Lock()
        mu2.Lock()
        defer mu2.Unlock()
        defer mu1.Unlock()
        // Critical section
    }
    
    go lockPair()
    go lockPair()
}
```

### GC Pressure Issues

**Symptoms:**
- High GC CPU usage (>5%)
- Frequent GC cycles
- Long GC pause times
- Application stalls

**Diagnostic Tools:**
```bash
# Enable GC tracing
export GODEBUG=gctrace=1

# Monitor GC behavior
go run myapp.go 2>&1 | grep "gc "

# Alternative: programmatic monitoring
go run -gcflags="-m" myapp.go  # Show escape analysis
```

**GC Optimization Strategies:**

1. **Reduce Allocation Rate**
```go
// BAD: High allocation rate
func processDataBad(items []string) []string {
    var result []string
    for _, item := range items {
        processed := strings.ToUpper(item) + "_PROCESSED"
        result = append(result, processed)
    }
    return result
}

// GOOD: Pre-allocate and reuse
func processDataGood(items []string) []string {
    result := make([]string, 0, len(items))
    var builder strings.Builder
    
    for _, item := range items {
        builder.Reset()
        builder.Grow(len(item) + 10) // Pre-allocate
        builder.WriteString(strings.ToUpper(item))
        builder.WriteString("_PROCESSED")
        result = append(result, builder.String())
    }
    return result
}
```

2. **Use Object Pools**
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processWithPool(data []byte) []byte {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer[:0]) // Reset length, keep capacity
    
    // Use buffer for processing
    buffer = append(buffer, data...)
    // ... processing logic
    
    // Return copy since buffer will be reused
    result := make([]byte, len(buffer))
    copy(result, buffer)
    return result
}
```

## Performance Anti-patterns

### Premature Optimization

**Problem:** Optimizing without measuring
```go
// BAD: Complex optimization without profiling
func prematureOptimization(data []int) int {
    // Unrolled loop for "performance"
    sum := 0
    i := 0
    for ; i < len(data)-3; i += 4 {
        sum += data[i] + data[i+1] + data[i+2] + data[i+3]
    }
    for ; i < len(data); i++ {
        sum += data[i]
    }
    return sum
}

// GOOD: Simple, clear code
func clearCode(data []int) int {
    sum := 0
    for _, value := range data {
        sum += value
    }
    return sum
}
```

**Solution:** Always measure first, optimize second.

### Interface Overuse

**Problem:** Excessive interface usage in hot paths
```go
// BAD: Unnecessary interface in hot path
type Processor interface {
    Process(data []byte) []byte
}

func hotPath(p Processor, data []byte) []byte {
    return p.Process(data) // Interface call overhead
}

// GOOD: Direct call for hot paths
type ConcreteProcessor struct{}

func (cp *ConcreteProcessor) Process(data []byte) []byte {
    // Implementation
    return data
}

func hotPath(cp *ConcreteProcessor, data []byte) []byte {
    return cp.Process(data) // Direct call, can be inlined
}
```

### Excessive Logging

**Problem:** Performance impact from logging
```go
// BAD: Expensive logging in hot path
func processRequest(req Request) {
    start := time.Now()
    defer func() {
        log.Printf("Request %s took %v", req.ID, time.Since(start))
    }()
    
    for i, item := range req.Items {
        log.Printf("Processing item %d: %+v", i, item) // Expensive!
        processItem(item)
    }
}

// GOOD: Conditional logging
func processRequest(req Request) {
    start := time.Now()
    defer func() {
        if logLevel >= InfoLevel {
            log.Printf("Request %s took %v", req.ID, time.Since(start))
        }
    }()
    
    for i, item := range req.Items {
        if logLevel >= DebugLevel {
            log.Printf("Processing item %d: %s", i, item.ID) // Lighter logging
        }
        processItem(item)
    }
}
```

## Debugging Techniques

### Race Condition Detection

```bash
# Build with race detector
go build -race myapp.go

# Run tests with race detection
go test -race ./...

# For production-like testing
go run -race myapp.go
```

**Common Race Patterns:**
```go
// BAD: Race condition
var counter int

func incrementCounter() {
    counter++ // Not atomic
}

// GOOD: Atomic operations
var counter int64

func incrementCounter() {
    atomic.AddInt64(&counter, 1)
}
```

### Escape Analysis Debugging

```bash
# Show escape analysis decisions
go build -gcflags="-m" myapp.go

# More verbose output
go build -gcflags="-m -m" myapp.go
```

**Example Analysis:**
```go
func analyzeEscape() {
    x := 42        // Does not escape
    y := &x        // May cause x to escape
    print(y)
}

// Output: ./main.go:2: moved to heap: x
//         ./main.go:3: &x escapes to heap
```

### Memory Usage Profiling

```go
func debugMemoryUsage() {
    var m1, m2 runtime.MemStats
    
    runtime.ReadMemStats(&m1)
    
    // Your code here
    performOperation()
    
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Memory used: %d bytes\n", m2.Alloc-m1.Alloc)
    fmt.Printf("Allocations: %d\n", m2.Mallocs-m1.Mallocs)
    fmt.Printf("GC cycles: %d\n", m2.NumGC-m1.NumGC)
}
```

## Production Troubleshooting

### Remote Debugging Setup

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Start pprof server on separate port
    go func() {
        log.Println("pprof server starting on :6060")
        log.Println(http.ListenAndServe(":6060", nil))
    }()
    
    // Graceful shutdown handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Main application
    runApplication()
    
    <-sigChan
    log.Println("Shutting down...")
}
```

### Health Check Integration

```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    status := map[string]interface{}{
        "status":     "healthy",
        "goroutines": runtime.NumGoroutine(),
        "memory": map[string]interface{}{
            "alloc":      m.Alloc,
            "sys":        m.Sys,
            "heap_inuse": m.HeapInuse,
        },
        "gc": map[string]interface{}{
            "num_gc":        m.NumGC,
            "pause_total":   m.PauseTotalNs,
            "gc_cpu_fraction": m.GCCPUFraction,
        },
    }
    
    // Alert thresholds
    if runtime.NumGoroutine() > 10000 {
        status["status"] = "warning"
        status["message"] = "High goroutine count"
    }
    
    if m.GCCPUFraction > 0.05 {
        status["status"] = "warning"
        status["message"] = "High GC overhead"
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}
```

### Emergency Response Procedures

**High Memory Usage:**
```bash
# 1. Capture heap profile immediately
curl http://localhost:6060/debug/pprof/heap > emergency-heap.prof

# 2. Force garbage collection
curl http://localhost:6060/debug/pprof/heap?gc=1

# 3. Check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 4. Restart service if critical
sudo systemctl restart myapp
```

**High CPU Usage:**
```bash
# 1. Capture CPU profile
curl "http://localhost:6060/debug/pprof/profile?seconds=30" > emergency-cpu.prof

# 2. Check system resources
top -p $(pgrep myapp)

# 3. Analyze immediately
go tool pprof emergency-cpu.prof
```

This troubleshooting guide provides systematic approaches to identify, diagnose, and resolve common Go performance issues in both development and production environments.
