# Best Practices

This section distills years of Go performance engineering experience into actionable guidelines, checklists, and proven patterns for building high-performance Go applications.

## Core Principles

### 🎯 Performance Engineering Mindset

**1. Measure First, Optimize Second**
```go
// ❌ Don't optimize based on assumptions
func assumedOptimization() {
    // "I think this will be faster"
    complexOptimization()
}

// ✅ Always measure before optimizing
func measuredOptimization() {
    // Profile, benchmark, then optimize
    if profileData.showsBottleneck() {
        targetedOptimization()
    }
}
```

**2. Understand Your Runtime**
- Know how the Go scheduler works
- Understand memory allocation patterns
- Learn garbage collection behavior
- Master escape analysis implications

**3. Think in Systems**
- Optimize the whole, not just parts
- Consider network, I/O, and external dependencies
- Balance CPU, memory, and latency trade-offs
- Account for production constraints

## Performance Guidelines

### Memory Management

#### Allocation Patterns
```go
// ✅ Pre-allocate slices with known capacity
func efficientSliceUsage(n int) []Item {
    items := make([]Item, 0, n)  // Capacity hint prevents reallocations
    for i := 0; i < n; i++ {
        items = append(items, generateItem(i))
    }
    return items
}

// ✅ Reuse buffers to reduce allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processWithPool(data []byte) []byte {
    buf := bufferPool.Get().([]byte)
    buf = buf[:0] // Reset length, keep capacity
    defer bufferPool.Put(buf)
    
    // Process data using buffer
    return processData(data, buf)
}
```

#### Memory Layout Optimization
```go
// ❌ Poor struct layout (uses more memory)
type BadStruct struct {
    flag1    bool    // 1 byte + 7 bytes padding
    value    uint64  // 8 bytes
    flag2    bool    // 1 byte + 7 bytes padding  
    name     string  // 16 bytes
} // Total: 32 bytes

// ✅ Optimized struct layout (better packing)
type GoodStruct struct {
    value    uint64  // 8 bytes
    name     string  // 16 bytes
    flag1    bool    // 1 byte
    flag2    bool    // 1 byte + 6 bytes padding
} // Total: 24 bytes (25% smaller)
```

### Algorithm Selection

#### Data Structure Choice
```go
// ✅ Choose appropriate data structures
func dataStructureSelection() {
    // For frequent lookups: map
    userMap := make(map[string]*User)
    
    // For ordered iteration: slice
    userList := make([]*User, 0)
    
    // For unique items: map[T]struct{}
    uniqueItems := make(map[string]struct{})
    
    // For priority queues: container/heap
    var priorityQueue PriorityQueue
    heap.Init(&priorityQueue)
}
```

#### Algorithm Complexity Awareness
```go
// ❌ O(n²) nested loops
func findDuplicatesSlow(items []string) map[string]int {
    counts := make(map[string]int)
    for i, item := range items {
        for j := i + 1; j < len(items); j++ {
            if items[j] == item {
                counts[item]++
            }
        }
    }
    return counts
}

// ✅ O(n) single pass
func findDuplicatesFast(items []string) map[string]int {
    counts := make(map[string]int)
    for _, item := range items {
        counts[item]++
    }
    return counts
}
```

### Concurrency Patterns

#### Goroutine Management
```go
// ✅ Use worker pools for controlled concurrency
func workerPoolPattern(jobs <-chan Job, results chan<- Result) {
    const numWorkers = 8
    
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- processJob(job)
            }
        }()
    }
    
    wg.Wait()
    close(results)
}

// ✅ Use context for cancellation and timeouts
func contextAwareOperation(ctx context.Context) error {
    select {
    case result := <-performOperation():
        return handleResult(result)
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

#### Channel Optimization
```go
// ✅ Buffer channels appropriately
func channelBuffering() {
    // Unbuffered: synchronous communication
    sync := make(chan Message)
    
    // Buffered: async with known capacity
    async := make(chan Message, 100)
    
    // Size buffer based on expected load
    burst := make(chan Message, expectedBurstSize)
}
```

### I/O and Networking

#### Buffer Management
```go
// ✅ Use appropriate buffer sizes
func efficientIO() {
    // For file I/O: typically 64KB
    fileBuffer := make([]byte, 64*1024)
    
    // For network I/O: typically 8-32KB
    netBuffer := make([]byte, 32*1024)
    
    // Use bufio for small, frequent operations
    reader := bufio.NewReaderSize(conn, 8192)
    writer := bufio.NewWriterSize(conn, 8192)
}
```

#### Connection Pooling
```go
// ✅ Reuse connections
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

## Code Review Checklist

### Performance Review Points

#### ✅ Memory Allocation Review
- [ ] Pre-allocate slices and maps with expected capacity
- [ ] Use `strings.Builder` for string concatenation
- [ ] Implement object pooling for frequently allocated objects
- [ ] Avoid unnecessary boxing/unboxing of interfaces
- [ ] Check struct field ordering for optimal packing

#### ✅ Algorithm Efficiency Review  
- [ ] Verify time complexity of algorithms (avoid O(n²) where possible)
- [ ] Choose appropriate data structures for use case
- [ ] Consider caching for expensive computations
- [ ] Eliminate redundant work in loops
- [ ] Use appropriate sorting algorithms

#### ✅ Concurrency Review
- [ ] Limit goroutine creation (use worker pools)
- [ ] Check for race conditions and data races
- [ ] Ensure proper channel usage and sizing
- [ ] Verify context usage for cancellation
- [ ] Review lock contention potential

#### ✅ I/O and Resource Review
- [ ] Use connection pooling for external services
- [ ] Implement proper timeouts for all I/O operations
- [ ] Buffer I/O operations appropriately
- [ ] Close resources in defer statements
- [ ] Handle errors appropriately

### Code Quality Checklist

```go
// ✅ Performance-conscious code template
func performantFunction(ctx context.Context, input []Item) ([]Result, error) {
    // 1. Input validation
    if len(input) == 0 {
        return nil, nil
    }
    
    // 2. Pre-allocate with capacity
    results := make([]Result, 0, len(input))
    
    // 3. Use buffer pool for intermediate data
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // 4. Process with context awareness
    for _, item := range input {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            // Process item efficiently
            result, err := processItem(item, buf)
            if err != nil {
                return nil, fmt.Errorf("processing item %v: %w", item, err)
            }
            results = append(results, result)
        }
    }
    
    return results, nil
}
```

## Production Deployment

### Performance Monitoring

#### Essential Metrics
```go
// ✅ Comprehensive performance monitoring
type PerformanceMetrics struct {
    // Throughput metrics
    RequestsPerSecond   float64 `json:"requests_per_second"`
    EventsProcessed     int64   `json:"events_processed"`
    
    // Latency metrics  
    ResponseTimeP50     time.Duration `json:"response_time_p50"`
    ResponseTimeP95     time.Duration `json:"response_time_p95"`
    ResponseTimeP99     time.Duration `json:"response_time_p99"`
    
    // Resource metrics
    CPUUsagePercent     float64 `json:"cpu_usage_percent"`
    MemoryUsageBytes    uint64  `json:"memory_usage_bytes"`
    GoroutineCount      int     `json:"goroutine_count"`
    
    // Go runtime metrics
    GCCycles            uint32        `json:"gc_cycles"`
    GCPauseTime         time.Duration `json:"gc_pause_time"`
    HeapObjects         uint64        `json:"heap_objects"`
    AllocationRate      uint64        `json:"allocation_rate"`
}

func collectMetrics() PerformanceMetrics {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return PerformanceMetrics{
        RequestsPerSecond:   requestCounter.Rate(),
        ResponseTimeP95:     responseTimeHistogram.Percentile(0.95),
        CPUUsagePercent:     getCPUUsage(),
        MemoryUsageBytes:    m.HeapAlloc,
        GoroutineCount:      runtime.NumGoroutine(),
        GCCycles:            m.NumGC,
        GCPauseTime:         time.Duration(m.PauseNs[(m.NumGC+255)%256]),
        HeapObjects:         m.HeapObjects,
        AllocationRate:      m.Mallocs - m.Frees,
    }
}
```

#### Runtime Tuning
```bash
# ✅ Production environment tuning
export GOMAXPROCS=8              # Match container CPU limits
export GOGC=100                  # Default GC target (adjust based on memory pressure)
export GOMEMLIMIT=2GiB           # Hard memory limit (Go 1.19+)
export GODEBUG=gctrace=1         # GC monitoring in production

# Application-specific tuning
export GODEBUG=schedtrace=1000   # Scheduler monitoring (debugging only)
export GODEBUG=allocfreetrace=1  # Allocation tracing (debugging only)
```

### Deployment Strategy

#### Gradual Rollout
```yaml
# ✅ Performance-aware deployment pipeline
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: performance-optimized-service
spec:
  strategy:
    canary:
      steps:
      - setWeight: 5      # Start with 5% traffic
      - pause: {duration: 10m}
      - analysis:          # Automated performance validation
          templates:
          - templateName: success-rate
          - templateName: response-time
          - templateName: error-rate
      - setWeight: 25
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 10m}
      - setWeight: 100
```

#### Performance Validation
```go
// ✅ Automated performance regression detection
func validatePerformanceInProduction() error {
    // Collect current metrics
    current := collectMetrics()
    
    // Compare against baseline
    baseline := loadBaselineMetrics()
    
    // Define acceptable thresholds
    thresholds := PerformanceThresholds{
        MaxResponseTimeIncrease: 1.2,  // 20% increase max
        MaxCPUIncrease:         1.15, // 15% increase max  
        MaxMemoryIncrease:      1.1,  // 10% increase max
        MinThroughputRatio:     0.95, // 5% decrease max
    }
    
    // Validate against thresholds
    if current.ResponseTimeP95 > baseline.ResponseTimeP95*thresholds.MaxResponseTimeIncrease {
        return fmt.Errorf("response time regression detected")
    }
    
    // Additional validations...
    
    return nil
}
```

## Performance Testing

### Benchmark Design

#### Effective Benchmarking
```go
// ✅ Comprehensive benchmark suite
func BenchmarkCriticalPath(b *testing.B) {
    // Setup test data
    testData := generateRealisticTestData(10000)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result := criticalPathFunction(testData)
        _ = result // Prevent optimization
    }
}

// ✅ Sub-benchmarks for different scenarios
func BenchmarkVariousInputSizes(b *testing.B) {
    sizes := []int{100, 1000, 10000, 100000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            data := generateTestData(size)
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                _ = processData(data)
            }
        })
    }
}
```

#### Load Testing Integration
```go
// ✅ Production load testing
func TestProductionLoad(t *testing.T) {
    server := startTestServer()
    defer server.Close()
    
    // Configure load test parameters
    config := LoadTestConfig{
        Concurrency:     50,
        RequestsPerSec:  1000,
        Duration:        5 * time.Minute,
        TargetURL:       server.URL,
    }
    
    // Run load test
    results := runLoadTest(config)
    
    // Validate performance requirements
    assert.Less(t, results.MedianResponseTime, 100*time.Millisecond)
    assert.Less(t, results.P95ResponseTime, 500*time.Millisecond)
    assert.Greater(t, results.SuccessRate, 0.999)
    assert.Less(t, results.ErrorRate, 0.001)
}
```

## Anti-Patterns to Avoid

### ❌ Common Performance Mistakes

#### Premature Optimization
```go
// ❌ Don't optimize without profiling
func prematureOptimization() {
    // Complex optimization for unclear benefit
    useComplexDataStructure()
}

// ✅ Profile-driven optimization
func measuredOptimization() {
    // 1. Profile current implementation
    // 2. Identify actual bottlenecks
    // 3. Optimize based on data
    // 4. Measure improvement
}
```

#### Micro-optimizations
```go
// ❌ Micro-optimizing non-critical paths
func microOptimization() {
    // Optimizing a function that uses 0.1% of CPU time
}

// ✅ Focus on high-impact optimizations  
func macroOptimization() {
    // Optimize functions using >5% of CPU time
}
```

#### Ignoring Memory Allocation
```go
// ❌ Ignoring allocation patterns
func allocationHeavy() string {
    result := ""
    for i := 0; i < 1000; i++ {
        result += fmt.Sprintf("item%d,", i) // Many allocations
    }
    return result
}

// ✅ Allocation-conscious implementation
func allocationLight() string {
    var builder strings.Builder
    builder.Grow(10000) // Pre-allocate
    
    for i := 0; i < 1000; i++ {
        builder.WriteString(fmt.Sprintf("item%d,", i))
    }
    return builder.String()
}
```

## Summary Guidelines

### Development Process
1. **Design for performance** from the beginning
2. **Profile early and often** during development
3. **Set performance budgets** and monitor against them
4. **Automate performance testing** in CI/CD pipelines
5. **Monitor production performance** continuously

### Optimization Strategy
1. **Measure first** - Always profile before optimizing
2. **Focus on impact** - Optimize high-CPU/high-allocation functions
3. **Validate improvements** - Benchmark before/after changes
4. **Consider trade-offs** - Balance readability, maintainability, and performance
5. **Monitor regressions** - Continuously validate performance in production

### Production Readiness
1. **Comprehensive monitoring** - Track all key performance metrics
2. **Gradual rollouts** - Use canary deployments for performance validation
3. **Automated alerts** - Set up alerts for performance regressions
4. **Runbook procedures** - Document performance troubleshooting steps
5. **Regular reviews** - Conduct periodic performance reviews

By following these best practices, you'll build Go applications that perform excellently from development through production, with the monitoring and processes needed to maintain that performance over time.

---

**Next Steps**: Apply these best practices to your projects and explore the **[Tools & Resources](../tools-resources/README.md)** section for additional performance engineering tools and references.
