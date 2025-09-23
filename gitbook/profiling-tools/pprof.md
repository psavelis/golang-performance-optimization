# pprof

Master Go's primary profiling tool for comprehensive performance analysis, from basic CPU profiling to advanced memory analysis and production monitoring.

## pprof Overview

pprof is Go's built-in profiling tool that provides detailed insights into your program's runtime behavior. It can capture and analyze various performance metrics:

- CPU usage patterns and hot functions
- Memory allocation and heap usage
- Goroutine blocking and synchronization
- Mutex contention analysis
- Custom application metrics

## Basic pprof Usage

### Command-Line CPU Profiling

```go
package main

import (
    "flag"
    "log"
    "os"
    "runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
    flag.Parse()
    
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }
    
    // Your application code
    performCPUIntensiveWork()
}

func performCPUIntensiveWork() {
    // Simulate CPU-intensive operations
    for i := 0; i < 1000000; i++ {
        calculatePrime(i)
    }
}

func calculatePrime(n int) bool {
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

Run and analyze:
```bash
go run main.go -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### HTTP Server Integration

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

func main() {
    // Start pprof HTTP server on dedicated port
    go func() {
        log.Println("pprof server starting on :6060")
        log.Println(http.ListenAndServe(":6060", nil))
    }()
    
    // Application routes
    http.HandleFunc("/", heavyHandler)
    http.HandleFunc("/memory", memoryHandler)
    http.HandleFunc("/goroutines", goroutineHandler)
    
    log.Println("Application server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func heavyHandler(w http.ResponseWriter, r *http.Request) {
    // CPU-intensive work
    result := 0
    for i := 0; i < 10000000; i++ {
        result += i * i
    }
    
    fmt.Fprintf(w, "Calculation result: %d\n", result)
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
    // Memory allocation
    data := make([][]byte, 1000)
    for i := range data {
        data[i] = make([]byte, 1024*1024) // 1MB chunks
    }
    
    runtime.GC()
    fmt.Fprintf(w, "Allocated %d MB\n", len(data))
}

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
    // Create many goroutines
    for i := 0; i < 100; i++ {
        go func(id int) {
            time.Sleep(time.Duration(id) * time.Millisecond)
        }(i)
    }
    
    fmt.Fprintf(w, "Created 100 goroutines\n")
}
```

Access profiles:
```bash
# CPU profile (collect for 30 seconds)
curl "http://localhost:6060/debug/pprof/profile?seconds=30" > cpu.prof

# Heap profile
curl "http://localhost:6060/debug/pprof/heap" > heap.prof

# Goroutine profile
curl "http://localhost:6060/debug/pprof/goroutine" > goroutine.prof

# Block profile
curl "http://localhost:6060/debug/pprof/block" > block.prof
```

## Interactive pprof Analysis

### CPU Profile Analysis

```bash
go tool pprof cpu.prof

# Top CPU consumers
(pprof) top
Showing nodes accounting for 1230ms, 94.62% of 1300ms total
Dropped 15 nodes (cum <= 6.50ms)
      flat  flat%   sum%        cum   cum%
     780ms 60.00% 60.00%      780ms 60.00%  main.calculatePrime
     290ms 22.31% 82.31%      290ms 22.31%  runtime.mallocgc
     100ms  7.69% 90.00%      100ms  7.69%  main.performCPUIntensiveWork
      60ms  4.62% 94.62%       60ms  4.62%  runtime.memmove

# Show top 20
(pprof) top20

# Function details
(pprof) list main.calculatePrime
Total: 1.30s
ROUTINE ======================== main.calculatePrime in /path/to/main.go
     780ms      780ms (flat, cum) 60.00% of Total
         .          .     25:func calculatePrime(n int) bool {
         .          .     26:    if n < 2 {
         .          .     27:        return false
         .          .     28:    }
     180ms      180ms     29:    for i := 2; i*i <= n; i++ {
     600ms      600ms     30:        if n%i == 0 {
         .          .     31:            return false
         .          .     32:        }
         .          .     33:    }
         .          .     34:    return true
         .          .     35:}

# Call graph
(pprof) web

# Flame graph (if supported)
(pprof) web --flame
```

### Memory Profile Analysis

```bash
go tool pprof heap.prof

# Top memory allocators
(pprof) top
Showing nodes accounting for 512.17MB, 100% of 512.17MB total
      flat  flat%   sum%        cum   cum%
  512.17MB   100%   100%   512.17MB   100%  main.memoryHandler

# Show allocation sites
(pprof) list main.memoryHandler
Total: 512.17MB
ROUTINE ======================== main.memoryHandler
  512.17MB   512.17MB (flat, cum)   100% of Total
         .          .     35:func memoryHandler(w http.ResponseWriter, r *http.Request) {
         .          .     36:    // Memory allocation
         .          .     37:    data := make([][]byte, 1000)
         .          .     38:    for i := range data {
  512.17MB   512.17MB     39:        data[i] = make([]byte, 1024*1024) // 1MB chunks
         .          .     40:    }
         .          .     41:
         .          .     42:    runtime.GC()
         .          .     43:    fmt.Fprintf(w, "Allocated %d MB\n", len(data))
         .          .     44:}

# Sample values (shows actual allocations)
(pprof) sample_index=alloc_objects
(pprof) sample_index=alloc_space
(pprof) sample_index=inuse_objects
(pprof) sample_index=inuse_space
```

## Advanced pprof Features

### Comparative Analysis

```go
// benchmark_comparison.go
package main

import (
    "fmt"
    "os"
    "runtime/pprof"
    "time"
)

func main() {
    // Profile before optimization
    f1, _ := os.Create("before.prof")
    pprof.StartCPUProfile(f1)
    
    slowAlgorithm()
    
    pprof.StopCPUProfile()
    f1.Close()
    
    // Small delay
    time.Sleep(100 * time.Millisecond)
    
    // Profile after optimization
    f2, _ := os.Create("after.prof")
    pprof.StartCPUProfile(f2)
    
    fastAlgorithm()
    
    pprof.StopCPUProfile()
    f2.Close()
}

func slowAlgorithm() {
    var result []int
    for i := 0; i < 100000; i++ {
        result = append(result, i) // Inefficient growth
    }
    fmt.Printf("Slow algorithm processed %d items\n", len(result))
}

func fastAlgorithm() {
    result := make([]int, 0, 100000) // Pre-allocated
    for i := 0; i < 100000; i++ {
        result = append(result, i)
    }
    fmt.Printf("Fast algorithm processed %d items\n", len(result))
}
```

Compare profiles:
```bash
go run benchmark_comparison.go

# Compare the two profiles
go tool pprof -base before.prof after.prof

# Or use diff mode
go tool pprof -diff_base before.prof after.prof
```

### Labels and Tags

```go
package main

import (
    "context"
    "runtime/pprof"
    "time"
)

func main() {
    // Label different operations for detailed analysis
    processUserRequests()
}

func processUserRequests() {
    // Simulate different types of requests
    for i := 0; i < 100; i++ {
        userType := "premium"
        if i%3 == 0 {
            userType = "free"
        }
        
        operation := "read"
        if i%10 == 0 {
            operation = "write"
        }
        
        // Add labels to profile data
        pprof.Do(context.Background(), 
            pprof.Labels("user_type", userType, "operation", operation),
            func(ctx context.Context) {
                processRequest(userType, operation)
            })
    }
}

func processRequest(userType, operation string) {
    // Different processing based on user type and operation
    if userType == "premium" {
        if operation == "write" {
            time.Sleep(5 * time.Millisecond) // Expensive operation
        } else {
            time.Sleep(1 * time.Millisecond) // Fast read
        }
    } else {
        time.Sleep(10 * time.Millisecond) // Slower for free users
    }
    
    // Simulate some CPU work
    result := 0
    iterations := 10000
    if userType == "premium" {
        iterations *= 2 // More processing for premium users
    }
    
    for i := 0; i < iterations; i++ {
        result += i * i
    }
}
```

Analyze labeled profiles:
```bash
go tool pprof cpu.prof

# Show available tags
(pprof) tags
user_type: free premium
operation: read write

# Focus on specific tags
(pprof) tagfocus="user_type:premium"
(pprof) tagfocus="operation:write"
(pprof) tagfocus="user_type:premium,operation:write"

# Ignore specific tags
(pprof) tagignore="user_type:free"

# Show top functions for tagged samples
(pprof) top --tags
```

### Memory Allocation Tracking

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "strings"
)

var memprofile = flag.String("memprofile", "", "write memory profile to file")

func main() {
    flag.Parse()
    
    // Different allocation patterns
    stringOperations()
    sliceOperations()
    mapOperations()
    
    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            panic(err)
        }
        defer f.Close()
        
        runtime.GC() // Get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            panic(err)
        }
    }
}

func stringOperations() {
    var result string
    for i := 0; i < 10000; i++ {
        result += fmt.Sprintf("item_%d ", i) // Inefficient concatenation
    }
    _ = result
}

func sliceOperations() {
    var data [][]byte
    for i := 0; i < 1000; i++ {
        chunk := make([]byte, 1024*(i%10+1)) // Variable sizes
        for j := range chunk {
            chunk[j] = byte(j % 256)
        }
        data = append(data, chunk)
    }
    _ = data
}

func mapOperations() {
    cache := make(map[string][]byte)
    for i := 0; i < 5000; i++ {
        key := strings.Repeat("k", i%100+1)
        value := make([]byte, (i%50+1)*100)
        cache[key] = value
    }
    _ = cache
}
```

Analyze memory allocation:
```bash
go run main.go -memprofile=mem.prof
go tool pprof mem.prof

# Top memory allocators
(pprof) top
(pprof) top --cum

# Show allocation sites
(pprof) list stringOperations
(pprof) list sliceOperations

# Different sample types
(pprof) sample_index=alloc_space
(pprof) sample_index=inuse_space
(pprof) sample_index=alloc_objects
(pprof) sample_index=inuse_objects
```

## Web Interface

### Starting Web UI

```bash
# Start web interface
go tool pprof -http=:8080 cpu.prof

# Open specific view
go tool pprof -http=:8080 -focus=main cpu.prof

# Compare profiles in web UI
go tool pprof -http=:8080 -base=before.prof after.prof
```

Web interface features:
- **Top**: Tabular view of top functions
- **Graph**: Visual call graph
- **Flame Graph**: Flame graph visualization
- **Peek**: Function call hierarchy
- **Source**: Source code view with annotations
- **Disasm**: Assembly code view

### Custom Views and Filters

```bash
# Focus on specific functions
(pprof) focus="main\\..*"
(pprof) focus="runtime\\.gc"

# Ignore functions
(pprof) ignore="runtime\\..*"

# Show only specific sample types
(pprof) sample_index=inuse_space
(pprof) sample_index=alloc_objects

# Combine filters
(pprof) focus="main\\..*" -ignore="runtime\\..*"
```

## Production Profiling Best Practices

### Safe Production Integration

```go
package main

import (
    "context"
    "log"
    "net/http"
    "net/http/pprof"
    "runtime"
    "time"
)

func main() {
    // Configure for production
    runtime.SetBlockProfileRate(1)
    runtime.SetMutexProfileFraction(1)
    
    // Separate profiling server
    profMux := http.NewServeMux()
    profMux.HandleFunc("/debug/pprof/", pprof.Index)
    profMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    profMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    profMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    profMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    profServer := &http.Server{
        Addr:    ":6060",
        Handler: profMux,
    }
    
    go func() {
        log.Printf("Profiling server starting on %s", profServer.Addr)
        if err := profServer.ListenAndServe(); err != nil {
            log.Printf("Profiling server error: %v", err)
        }
    }()
    
    // Main application server
    appMux := http.NewServeMux()
    appMux.HandleFunc("/", handleRequest)
    
    appServer := &http.Server{
        Addr:    ":8080",
        Handler: appMux,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
    
    log.Printf("Application server starting on %s", appServer.Addr)
    log.Fatal(appServer.ListenAndServe())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Add profiling context
    ctx := pprof.WithLabels(r.Context(), pprof.Labels("endpoint", r.URL.Path))
    pprof.SetGoroutineLabels(ctx)
    
    // Process request
    processWithProfiling(ctx)
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func processWithProfiling(ctx context.Context) {
    pprof.Do(ctx, pprof.Labels("operation", "business_logic"), func(ctx context.Context) {
        time.Sleep(10 * time.Millisecond) // Simulate work
    })
}
```

### Automated Profile Collection

```go
type ProfileCollector struct {
    outputDir string
    interval  time.Duration
    retention time.Duration
}

func NewProfileCollector(outputDir string, interval, retention time.Duration) *ProfileCollector {
    return &ProfileCollector{
        outputDir: outputDir,
        interval:  interval,
        retention: retention,
    }
}

func (pc *ProfileCollector) Start(ctx context.Context) error {
    ticker := time.NewTicker(pc.interval)
    defer ticker.Stop()
    
    // Cleanup old profiles
    go pc.cleanupOldProfiles(ctx)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := pc.collectProfiles(); err != nil {
                log.Printf("Profile collection error: %v", err)
            }
        }
    }
}

func (pc *ProfileCollector) collectProfiles() error {
    timestamp := time.Now().Format("20060102-150405")
    
    profiles := []struct {
        name string
        url  string
    }{
        {"cpu", "http://localhost:6060/debug/pprof/profile?seconds=30"},
        {"heap", "http://localhost:6060/debug/pprof/heap"},
        {"goroutine", "http://localhost:6060/debug/pprof/goroutine"},
        {"block", "http://localhost:6060/debug/pprof/block"},
    }
    
    for _, profile := range profiles {
        filename := fmt.Sprintf("%s/%s-%s.prof", pc.outputDir, profile.name, timestamp)
        if err := pc.downloadProfile(profile.url, filename); err != nil {
            log.Printf("Failed to collect %s profile: %v", profile.name, err)
        } else {
            log.Printf("Collected %s profile: %s", profile.name, filename)
        }
    }
    
    return nil
}
```

pprof is the cornerstone of Go performance analysis, providing the tools and insights needed to understand, optimize, and monitor your application's runtime behavior in both development and production environments.
