# Advanced Benchmarking

Explore sophisticated benchmarking techniques, custom metrics, and enterprise-grade performance testing strategies for production Go applications.

## Custom Benchmark Metrics

Beyond the standard ns/op, B/op, and allocs/op metrics, you can implement custom measurements to capture domain-specific performance characteristics.

### Custom Metrics Implementation

```go
type BenchmarkMetrics struct {
    Operations    int64
    TotalDuration time.Duration
    Throughput    float64 // ops/sec
    Latency       LatencyStats
    Resources     ResourceUsage
}

type LatencyStats struct {
    Min, Max, Mean time.Duration
    P50, P90, P95, P99 time.Duration
    Samples []time.Duration
}

type ResourceUsage struct {
    CPUTime     time.Duration
    MemoryPeak  int64
    GCPause     time.Duration
    Goroutines  int
}

func BenchmarkWithCustomMetrics(b *testing.B) {
    metrics := &BenchmarkMetrics{
        Latency: LatencyStats{
            Samples: make([]time.Duration, 0, b.N),
        },
    }
    
    // Capture initial state
    var memStats1, memStats2 runtime.MemStats
    runtime.ReadMemStats(&memStats1)
    startGoroutines := runtime.NumGoroutine()
    
    b.ResetTimer()
    start := time.Now()
    
    for i := 0; i < b.N; i++ {
        opStart := time.Now()
        result := complexOperation()
        opDuration := time.Since(opStart)
        
        metrics.Latency.Samples = append(metrics.Latency.Samples, opDuration)
        _ = result
    }
    
    metrics.TotalDuration = time.Since(start)
    metrics.Operations = int64(b.N)
    metrics.Throughput = float64(b.N) / metrics.TotalDuration.Seconds()
    
    // Capture final state
    runtime.ReadMemStats(&memStats2)
    metrics.Resources.MemoryPeak = int64(memStats2.Sys - memStats1.Sys)
    metrics.Resources.GCPause = time.Duration(memStats2.PauseTotalNs - memStats1.PauseTotalNs)
    metrics.Resources.Goroutines = runtime.NumGoroutine() - startGoroutines
    
    // Calculate latency percentiles
    calculateLatencyPercentiles(metrics)
    
    // Report custom metrics
    b.ReportMetric(metrics.Throughput, "ops/sec")
    b.ReportMetric(float64(metrics.Latency.P95.Nanoseconds()), "p95-ns")
    b.ReportMetric(float64(metrics.Resources.MemoryPeak), "peak-bytes")
    
    b.Logf("Custom Metrics Report:")
    b.Logf("  Throughput: %.2f ops/sec", metrics.Throughput)
    b.Logf("  Latency P95: %v", metrics.Latency.P95)
    b.Logf("  Memory Peak: %d bytes", metrics.Resources.MemoryPeak)
    b.Logf("  GC Pause: %v", metrics.Resources.GCPause)
}
```

### Network-Aware Benchmarks

```go
func BenchmarkNetworkLatency(b *testing.B) {
    // Simulate different network conditions
    networkConditions := []struct {
        name    string
        latency time.Duration
        jitter  time.Duration
    }{
        {"LAN", 1 * time.Millisecond, 100 * time.Microsecond},
        {"WAN", 50 * time.Millisecond, 10 * time.Millisecond},
        {"Mobile", 200 * time.Millisecond, 50 * time.Millisecond},
    }
    
    for _, condition := range networkConditions {
        b.Run(condition.name, func(b *testing.B) {
            simulator := NewNetworkSimulator(condition.latency, condition.jitter)
            defer simulator.Close()
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                result := networkAwareOperation(simulator)
                _ = result
            }
        })
    }
}

type NetworkSimulator struct {
    baseLatency time.Duration
    jitter      time.Duration
    rand        *rand.Rand
}

func NewNetworkSimulator(latency, jitter time.Duration) *NetworkSimulator {
    return &NetworkSimulator{
        baseLatency: latency,
        jitter:      jitter,
        rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (ns *NetworkSimulator) SimulateDelay() {
    jitterAmount := time.Duration(ns.rand.Float64() * float64(ns.jitter))
    totalDelay := ns.baseLatency + jitterAmount
    time.Sleep(totalDelay)
}
```

### Database Performance Benchmarks

```go
func BenchmarkDatabaseOperations(b *testing.B) {
    db := setupTestDatabase()
    defer db.Close()
    
    // Test different connection pool sizes
    poolSizes := []int{1, 5, 10, 25, 50}
    
    for _, poolSize := range poolSizes {
        db.SetMaxOpenConns(poolSize)
        db.SetMaxIdleConns(poolSize / 2)
        
        b.Run(fmt.Sprintf("Pool_%d", poolSize), func(b *testing.B) {
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    // Simulate realistic database workload
                    tx, err := db.Begin()
                    if err != nil {
                        b.Fatal(err)
                    }
                    
                    // Read operation
                    var count int
                    err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE active = true").Scan(&count)
                    if err != nil {
                        tx.Rollback()
                        b.Fatal(err)
                    }
                    
                    // Write operation
                    _, err = tx.Exec("UPDATE users SET last_seen = NOW() WHERE id = ?", 
                        rand.Intn(10000))
                    if err != nil {
                        tx.Rollback()
                        b.Fatal(err)
                    }
                    
                    if err = tx.Commit(); err != nil {
                        b.Fatal(err)
                    }
                }
            })
        })
    }
}
```

### Cache Performance Analysis

```go
func BenchmarkCacheEfficiency(b *testing.B) {
    cacheConfigs := []struct {
        name     string
        size     int
        strategy string
    }{
        {"LRU_1K", 1000, "LRU"},
        {"LRU_10K", 10000, "LRU"},
        {"LFU_1K", 1000, "LFU"},
        {"Random_1K", 1000, "Random"},
    }
    
    for _, config := range cacheConfigs {
        b.Run(config.name, func(b *testing.B) {
            cache := NewCache(config.size, config.strategy)
            
            // Warm up cache
            for i := 0; i < config.size/2; i++ {
                cache.Set(fmt.Sprintf("key_%d", i), i)
            }
            
            hitCount := int64(0)
            missCount := int64(0)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                key := fmt.Sprintf("key_%d", rand.Intn(config.size*2))
                
                if _, found := cache.Get(key); found {
                    atomic.AddInt64(&hitCount, 1)
                } else {
                    atomic.AddInt64(&missCount, 1)
                    cache.Set(key, i)
                }
            }
            
            hitRatio := float64(hitCount) / float64(hitCount+missCount) * 100
            b.ReportMetric(hitRatio, "hit-ratio-%")
            b.Logf("Cache hit ratio: %.2f%%", hitRatio)
        })
    }
}
```

## Load Testing Integration

### HTTP Server Load Testing

```go
func BenchmarkHTTPServerLoad(b *testing.B) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Simulate varying response times
        processingTime := time.Duration(rand.Intn(100)) * time.Millisecond
        time.Sleep(processingTime)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "timestamp": time.Now(),
            "processed": true,
        })
    })
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Test different concurrency levels
    concurrencyLevels := []int{1, 10, 50, 100, 200}
    
    for _, concurrency := range concurrencyLevels {
        b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
            client := &http.Client{
                Transport: &http.Transport{
                    MaxIdleConnsPerHost: concurrency,
                    IdleConnTimeout:     30 * time.Second,
                },
                Timeout: 30 * time.Second,
            }
            
            // Track response times
            var responseTimes []time.Duration
            var mutex sync.Mutex
            
            b.SetParallelism(concurrency)
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    start := time.Now()
                    
                    resp, err := client.Get(server.URL)
                    if err != nil {
                        b.Error(err)
                        continue
                    }
                    
                    responseTime := time.Since(start)
                    
                    mutex.Lock()
                    responseTimes = append(responseTimes, responseTime)
                    mutex.Unlock()
                    
                    resp.Body.Close()
                }
            })
            
            // Calculate and report response time percentiles
            sort.Slice(responseTimes, func(i, j int) bool {
                return responseTimes[i] < responseTimes[j]
            })
            
            if len(responseTimes) > 0 {
                p50 := responseTimes[len(responseTimes)*50/100]
                p95 := responseTimes[len(responseTimes)*95/100]
                p99 := responseTimes[len(responseTimes)*99/100]
                
                b.ReportMetric(float64(p50.Nanoseconds()), "p50-ns")
                b.ReportMetric(float64(p95.Nanoseconds()), "p95-ns")
                b.ReportMetric(float64(p99.Nanoseconds()), "p99-ns")
            }
        })
    }
}
```

### Message Queue Benchmarks

```go
func BenchmarkMessageQueue(b *testing.B) {
    queueSizes := []int{100, 1000, 10000}
    
    for _, queueSize := range queueSizes {
        b.Run(fmt.Sprintf("QueueSize_%d", queueSize), func(b *testing.B) {
            queue := make(chan Message, queueSize)
            
            // Start consumers
            numConsumers := runtime.NumCPU()
            var wg sync.WaitGroup
            
            for i := 0; i < numConsumers; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    for msg := range queue {
                        processMessage(msg)
                    }
                }()
            }
            
            b.ResetTimer()
            
            // Producer benchmark
            for i := 0; i < b.N; i++ {
                msg := Message{
                    ID:      i,
                    Payload: fmt.Sprintf("message_%d", i),
                    Time:    time.Now(),
                }
                
                select {
                case queue <- msg:
                default:
                    b.Error("Queue full")
                }
            }
            
            close(queue)
            wg.Wait()
        })
    }
}
```

## Microbenchmark Suites

### Algorithm Comparison Framework

```go
type AlgorithmBenchmark struct {
    Name      string
    Algorithm func([]int) int
    Setup     func() []int
}

func BenchmarkAlgorithmSuite(b *testing.B) {
    algorithms := []AlgorithmBenchmark{
        {
            Name:      "BubbleSort",
            Algorithm: bubbleSort,
            Setup:     func() []int { return generateRandomSlice(1000) },
        },
        {
            Name:      "QuickSort",
            Algorithm: quickSort,
            Setup:     func() []int { return generateRandomSlice(1000) },
        },
        {
            Name:      "MergeSort",
            Algorithm: mergeSort,
            Setup:     func() []int { return generateRandomSlice(1000) },
        },
    }
    
    inputSizes := []int{100, 1000, 10000}
    
    for _, size := range inputSizes {
        for _, algo := range algorithms {
            b.Run(fmt.Sprintf("%s_Size_%d", algo.Name, size), func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    b.StopTimer()
                    data := generateRandomSlice(size)
                    b.StartTimer()
                    
                    result := algo.Algorithm(data)
                    _ = result
                }
            })
        }
    }
}
```

### Data Structure Performance

```go
func BenchmarkDataStructures(b *testing.B) {
    operations := []string{"Insert", "Lookup", "Delete"}
    sizes := []int{1000, 10000, 100000}
    
    for _, size := range sizes {
        for _, op := range operations {
            // Map benchmark
            b.Run(fmt.Sprintf("Map_%s_Size_%d", op, size), func(b *testing.B) {
                m := make(map[int]bool)
                
                // Pre-populate for lookup/delete tests
                if op != "Insert" {
                    for i := 0; i < size; i++ {
                        m[i] = true
                    }
                }
                
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    key := rand.Intn(size)
                    switch op {
                    case "Insert":
                        m[key] = true
                    case "Lookup":
                        _ = m[key]
                    case "Delete":
                        delete(m, key)
                    }
                }
            })
            
            // Slice benchmark (for comparison)
            b.Run(fmt.Sprintf("Slice_%s_Size_%d", op, size), func(b *testing.B) {
                slice := make([]int, 0, size)
                
                // Pre-populate for lookup/delete tests
                if op != "Insert" {
                    for i := 0; i < size; i++ {
                        slice = append(slice, i)
                    }
                }
                
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    value := rand.Intn(size)
                    switch op {
                    case "Insert":
                        slice = append(slice, value)
                    case "Lookup":
                        for _, v := range slice {
                            if v == value {
                                break
                            }
                        }
                    case "Delete":
                        for i, v := range slice {
                            if v == value {
                                slice = append(slice[:i], slice[i+1:]...)
                                break
                            }
                        }
                    }
                }
            })
        }
    }
}
```

## Performance Regression Detection

### Automated Threshold Checking

```go
type PerformanceThreshold struct {
    MaxLatency      time.Duration
    MaxMemoryUsage  int64
    MinThroughput   float64
    MaxAllocations  int64
}

func BenchmarkWithThresholds(b *testing.B) {
    thresholds := PerformanceThreshold{
        MaxLatency:     10 * time.Millisecond,
        MaxMemoryUsage: 1024 * 1024, // 1MB
        MinThroughput:  1000,         // ops/sec
        MaxAllocations: 10,
    }
    
    var totalLatency time.Duration
    var maxMemory int64
    var totalAllocs int64
    
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        start := time.Now()
        
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)
        
        result := functionUnderTest()
        
        runtime.ReadMemStats(&m2)
        latency := time.Since(start)
        memUsed := int64(m2.Alloc - m1.Alloc)
        allocs := int64(m2.Mallocs - m1.Mallocs)
        
        totalLatency += latency
        if memUsed > maxMemory {
            maxMemory = memUsed
        }
        totalAllocs += allocs
        
        _ = result
    }
    
    avgLatency := totalLatency / time.Duration(b.N)
    throughput := float64(b.N) / totalLatency.Seconds()
    avgAllocs := totalAllocs / int64(b.N)
    
    // Check thresholds
    if avgLatency > thresholds.MaxLatency {
        b.Errorf("Latency threshold exceeded: %v > %v", avgLatency, thresholds.MaxLatency)
    }
    
    if maxMemory > thresholds.MaxMemoryUsage {
        b.Errorf("Memory threshold exceeded: %d > %d bytes", maxMemory, thresholds.MaxMemoryUsage)
    }
    
    if throughput < thresholds.MinThroughput {
        b.Errorf("Throughput threshold not met: %.2f < %.2f ops/sec", throughput, thresholds.MinThroughput)
    }
    
    if avgAllocs > thresholds.MaxAllocations {
        b.Errorf("Allocation threshold exceeded: %d > %d allocs/op", avgAllocs, thresholds.MaxAllocations)
    }
    
    b.ReportMetric(float64(avgLatency.Nanoseconds()), "avg-latency-ns")
    b.ReportMetric(throughput, "throughput-ops/sec")
    b.ReportMetric(float64(maxMemory), "peak-memory-bytes")
}
```

### Continuous Performance Monitoring

```go
func BenchmarkContinuousMonitoring(b *testing.B) {
    // Performance baseline from previous runs
    baseline := PerformanceBaseline{
        LatencyP95:  5 * time.Millisecond,
        Throughput:  2000,
        MemoryPeak:  512 * 1024,
    }
    
    // Allow 10% degradation
    tolerance := 0.10
    
    metrics := measurePerformance(b)
    
    // Compare against baseline
    latencyDelta := (metrics.LatencyP95.Seconds() - baseline.LatencyP95.Seconds()) / baseline.LatencyP95.Seconds()
    throughputDelta := (baseline.Throughput - metrics.Throughput) / baseline.Throughput
    memoryDelta := (float64(metrics.MemoryPeak - baseline.MemoryPeak)) / float64(baseline.MemoryPeak)
    
    if latencyDelta > tolerance {
        b.Errorf("Latency regression: %.2f%% increase", latencyDelta*100)
    }
    
    if throughputDelta > tolerance {
        b.Errorf("Throughput regression: %.2f%% decrease", throughputDelta*100)
    }
    
    if memoryDelta > tolerance {
        b.Errorf("Memory regression: %.2f%% increase", memoryDelta*100)
    }
    
    // Update baseline if performance improved
    if latencyDelta < -0.05 && throughputDelta < -0.05 {
        updatePerformanceBaseline(metrics)
        b.Logf("Performance baseline updated")
    }
}
```

Advanced benchmarking transforms performance testing from a simple measurement tool into a comprehensive performance engineering platform, enabling sophisticated analysis, regression detection, and continuous performance optimization in production Go applications.
