# Mutex Contention Analysis

Mutex contention analysis identifies synchronization bottlenecks where multiple goroutines compete for the same locks. This comprehensive guide covers advanced techniques for detecting, analyzing, and resolving mutex contention in Go applications.

## Understanding Mutex Contention

Mutex contention occurs when:
- **Multiple goroutines** attempt to acquire the same mutex simultaneously
- **Long hold times** keep locks unavailable for extended periods
- **High frequency locking** creates competition for shared resources
- **Poor lock granularity** causes unnecessary serialization
- **Lock ordering issues** lead to potential deadlocks

### Contention Metrics and Analysis

```go
package main

import (
    "fmt"
    "runtime"
    "sort"
    "sync"
    "sync/atomic"
    "time"
)

// ContentionAnalyzer provides detailed mutex contention analysis
type ContentionAnalyzer struct {
    mutexTrackers map[string]*TrackedMutex
    globalStats   *GlobalContentionStats
    mu            sync.RWMutex
    enabled       bool
}

type GlobalContentionStats struct {
    TotalContentions   int64
    TotalWaitTime      int64
    TotalLockTime      int64
    ActiveGoroutines   int32
    PeakContention     int32
    ContentionRate     float64 // Contentions per second
}

func NewContentionAnalyzer() *ContentionAnalyzer {
    return &ContentionAnalyzer{
        mutexTrackers: make(map[string]*TrackedMutex),
        globalStats:   &GlobalContentionStats{},
    }
}

func (ca *ContentionAnalyzer) Enable() {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    if ca.enabled {
        return
    }
    
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
    ca.enabled = true
}

func (ca *ContentionAnalyzer) Disable() {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    runtime.SetMutexProfileFraction(0)
    ca.enabled = false
}

func (ca *ContentionAnalyzer) CreateTrackedMutex(name string) *TrackedMutex {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    if tracker, exists := ca.mutexTrackers[name]; exists {
        return tracker
    }
    
    tracker := &TrackedMutex{
        name:     name,
        analyzer: ca,
        stats:    &MutexContentionStats{},
    }
    
    ca.mutexTrackers[name] = tracker
    return tracker
}

func (ca *ContentionAnalyzer) recordContention(mutexName string, waitTime, holdTime time.Duration, goroutineID int) {
    atomic.AddInt64(&ca.globalStats.TotalContentions, 1)
    atomic.AddInt64(&ca.globalStats.TotalWaitTime, int64(waitTime))
    atomic.AddInt64(&ca.globalStats.TotalLockTime, int64(holdTime))
    
    // Update peak contention tracking
    active := atomic.AddInt32(&ca.globalStats.ActiveGoroutines, 1)
    for {
        peak := atomic.LoadInt32(&ca.globalStats.PeakContention)
        if active <= peak || atomic.CompareAndSwapInt32(&ca.globalStats.PeakContention, peak, active) {
            break
        }
    }
}

func (ca *ContentionAnalyzer) recordLockRelease(mutexName string) {
    atomic.AddInt32(&ca.globalStats.ActiveGoroutines, -1)
}

func (ca *ContentionAnalyzer) GetContentionReport() ContentionReport {
    ca.mu.RLock()
    defer ca.mu.RUnlock()
    
    var mutexReports []MutexContentionReport
    totalContentions := atomic.LoadInt64(&ca.globalStats.TotalContentions)
    totalWaitTime := atomic.LoadInt64(&ca.globalStats.TotalWaitTime)
    totalLockTime := atomic.LoadInt64(&ca.globalStats.TotalLockTime)
    peakContention := atomic.LoadInt32(&ca.globalStats.PeakContention)
    
    for name, tracker := range ca.mutexTrackers {
        report := tracker.GetReport()
        report.Name = name
        mutexReports = append(mutexReports, report)
    }
    
    // Sort by contention severity
    sort.Slice(mutexReports, func(i, j int) bool {
        return mutexReports[i].ContentionSeverity() > mutexReports[j].ContentionSeverity()
    })
    
    return ContentionReport{
        TotalMutexes:     len(mutexReports),
        TotalContentions: totalContentions,
        TotalWaitTime:    time.Duration(totalWaitTime),
        TotalLockTime:    time.Duration(totalLockTime),
        PeakContention:   peakContention,
        MutexReports:     mutexReports,
    }
}

// TrackedMutex wraps sync.Mutex with contention tracking
type TrackedMutex struct {
    sync.Mutex
    name     string
    analyzer *ContentionAnalyzer
    stats    *MutexContentionStats
}

type MutexContentionStats struct {
    LockCount        int64
    ContentionCount  int64
    TotalWaitTime    int64
    TotalHoldTime    int64
    MaxWaitTime      int64
    MaxHoldTime      int64
    WaitTimeHist     *TimeHistogram
    HoldTimeHist     *TimeHistogram
    FirstContention  time.Time
    LastContention   time.Time
    GoroutineIDs     map[int]int64 // Goroutine ID -> contention count
    mu               sync.RWMutex
}

func (tm *TrackedMutex) Lock() {
    goroutineID := getGoroutineID()
    start := time.Now()
    
    // Try to acquire without blocking
    if tm.Mutex.TryLock() {
        // Got lock immediately - no contention
        atomic.AddInt64(&tm.stats.LockCount, 1)
        tm.recordLockAcquired(start, false, 0, goroutineID)
        return
    }
    
    // Lock is contended - record wait time
    tm.Mutex.Lock()
    waitTime := time.Since(start)
    
    atomic.AddInt64(&tm.stats.LockCount, 1)
    atomic.AddInt64(&tm.stats.ContentionCount, 1)
    atomic.AddInt64(&tm.stats.TotalWaitTime, int64(waitTime))
    
    // Update max wait time
    for {
        maxWait := atomic.LoadInt64(&tm.stats.MaxWaitTime)
        if int64(waitTime) <= maxWait || atomic.CompareAndSwapInt64(&tm.stats.MaxWaitTime, maxWait, int64(waitTime)) {
            break
        }
    }
    
    tm.recordLockAcquired(start, true, waitTime, goroutineID)
    tm.analyzer.recordContention(tm.name, waitTime, 0, goroutineID)
}

func (tm *TrackedMutex) Unlock() {
    lockStart := time.Now() // This should be stored from Lock(), simplified here
    tm.Mutex.Unlock()
    
    holdTime := time.Since(lockStart) // Simplified - would need proper tracking
    atomic.AddInt64(&tm.stats.TotalHoldTime, int64(holdTime))
    
    // Update max hold time
    for {
        maxHold := atomic.LoadInt64(&tm.stats.MaxHoldTime)
        if int64(holdTime) <= maxHold || atomic.CompareAndSwapInt64(&tm.stats.MaxHoldTime, maxHold, int64(holdTime)) {
            break
        }
    }
    
    tm.analyzer.recordLockRelease(tm.name)
}

func (tm *TrackedMutex) TryLock() bool {
    if tm.Mutex.TryLock() {
        atomic.AddInt64(&tm.stats.LockCount, 1)
        return true
    }
    return false
}

func (tm *TrackedMutex) recordLockAcquired(start time.Time, contended bool, waitTime time.Duration, goroutineID int) {
    now := time.Now()
    
    tm.stats.mu.Lock()
    defer tm.stats.mu.Unlock()
    
    if contended {
        if tm.stats.FirstContention.IsZero() {
            tm.stats.FirstContention = now
        }
        tm.stats.LastContention = now
        
        // Track per-goroutine contention
        if tm.stats.GoroutineIDs == nil {
            tm.stats.GoroutineIDs = make(map[int]int64)
        }
        tm.stats.GoroutineIDs[goroutineID]++
        
        // Update histograms
        if tm.stats.WaitTimeHist == nil {
            tm.stats.WaitTimeHist = NewTimeHistogram()
        }
        tm.stats.WaitTimeHist.Record(waitTime)
    }
}

func (tm *TrackedMutex) GetReport() MutexContentionReport {
    tm.stats.mu.RLock()
    defer tm.stats.mu.RUnlock()
    
    lockCount := atomic.LoadInt64(&tm.stats.LockCount)
    contentionCount := atomic.LoadInt64(&tm.stats.ContentionCount)
    totalWaitTime := atomic.LoadInt64(&tm.stats.TotalWaitTime)
    totalHoldTime := atomic.LoadInt64(&tm.stats.TotalHoldTime)
    maxWaitTime := atomic.LoadInt64(&tm.stats.MaxWaitTime)
    maxHoldTime := atomic.LoadInt64(&tm.stats.MaxHoldTime)
    
    var contentionRate float64
    if lockCount > 0 {
        contentionRate = float64(contentionCount) / float64(lockCount) * 100
    }
    
    var avgWaitTime, avgHoldTime time.Duration
    if contentionCount > 0 {
        avgWaitTime = time.Duration(totalWaitTime / contentionCount)
    }
    if lockCount > 0 {
        avgHoldTime = time.Duration(totalHoldTime / lockCount)
    }
    
    // Get top contending goroutines
    var topGoroutines []GoroutineContention
    for gid, count := range tm.stats.GoroutineIDs {
        topGoroutines = append(topGoroutines, GoroutineContention{
            GoroutineID: gid,
            Contentions: count,
        })
    }
    
    sort.Slice(topGoroutines, func(i, j int) bool {
        return topGoroutines[i].Contentions > topGoroutines[j].Contentions
    })
    
    if len(topGoroutines) > 5 {
        topGoroutines = topGoroutines[:5] // Top 5
    }
    
    return MutexContentionReport{
        Name:              tm.name,
        LockCount:         lockCount,
        ContentionCount:   contentionCount,
        ContentionRate:    contentionRate,
        TotalWaitTime:     time.Duration(totalWaitTime),
        TotalHoldTime:     time.Duration(totalHoldTime),
        AvgWaitTime:       avgWaitTime,
        AvgHoldTime:       avgHoldTime,
        MaxWaitTime:       time.Duration(maxWaitTime),
        MaxHoldTime:       time.Duration(maxHoldTime),
        FirstContention:   tm.stats.FirstContention,
        LastContention:    tm.stats.LastContention,
        TopGoroutines:     topGoroutines,
        WaitTimeHistogram: tm.stats.WaitTimeHist,
    }
}

type MutexContentionReport struct {
    Name              string
    LockCount         int64
    ContentionCount   int64
    ContentionRate    float64
    TotalWaitTime     time.Duration
    TotalHoldTime     time.Duration
    AvgWaitTime       time.Duration
    AvgHoldTime       time.Duration
    MaxWaitTime       time.Duration
    MaxHoldTime       time.Duration
    FirstContention   time.Time
    LastContention    time.Time
    TopGoroutines     []GoroutineContention
    WaitTimeHistogram *TimeHistogram
}

func (mcr MutexContentionReport) ContentionSeverity() float64 {
    // Calculate contention severity score
    severity := mcr.ContentionRate * float64(mcr.TotalWaitTime/time.Millisecond)
    return severity
}

func (mcr MutexContentionReport) String() string {
    result := fmt.Sprintf(`Mutex: %s
  Lock Count: %d
  Contention Count: %d (%.1f%%)
  Total Wait Time: %v
  Average Wait Time: %v
  Max Wait Time: %v
  Total Hold Time: %v
  Average Hold Time: %v
  Max Hold Time: %v
  Contention Severity: %.2f`,
        mcr.Name,
        mcr.LockCount,
        mcr.ContentionCount, mcr.ContentionRate,
        mcr.TotalWaitTime,
        mcr.AvgWaitTime,
        mcr.MaxWaitTime,
        mcr.TotalHoldTime,
        mcr.AvgHoldTime,
        mcr.MaxHoldTime,
        mcr.ContentionSeverity())
    
    if len(mcr.TopGoroutines) > 0 {
        result += "\n  Top Contending Goroutines:"
        for i, gc := range mcr.TopGoroutines {
            result += fmt.Sprintf("\n    %d. Goroutine %d: %d contentions", i+1, gc.GoroutineID, gc.Contentions)
        }
    }
    
    if mcr.WaitTimeHistogram != nil {
        result += "\n  Wait Time Distribution:"
        result += mcr.WaitTimeHistogram.String()
    }
    
    return result
}

type GoroutineContention struct {
    GoroutineID int
    Contentions int64
}

type ContentionReport struct {
    TotalMutexes     int
    TotalContentions int64
    TotalWaitTime    time.Duration
    TotalLockTime    time.Duration
    PeakContention   int32
    MutexReports     []MutexContentionReport
}

func (cr ContentionReport) String() string {
    result := fmt.Sprintf(`Global Contention Analysis:
  Total Mutexes: %d
  Total Contentions: %d
  Total Wait Time: %v
  Total Lock Time: %v
  Peak Concurrent Waiters: %d`,
        cr.TotalMutexes,
        cr.TotalContentions,
        cr.TotalWaitTime,
        cr.TotalLockTime,
        cr.PeakContention)
    
    if len(cr.MutexReports) > 0 {
        result += "\n\nMutex Contention Details:"
        for i, report := range cr.MutexReports {
            if i >= 5 { // Show top 5 most contended
                break
            }
            result += fmt.Sprintf("\n\n%d. %s", i+1, report.String())
        }
    }
    
    return result
}

// TimeHistogram tracks time distribution
type TimeHistogram struct {
    buckets []HistogramBucket
    mu      sync.RWMutex
}

type HistogramBucket struct {
    LowerBound time.Duration
    UpperBound time.Duration
    Count      int64
}

func NewTimeHistogram() *TimeHistogram {
    return &TimeHistogram{
        buckets: []HistogramBucket{
            {0, time.Microsecond, 0},
            {time.Microsecond, 10 * time.Microsecond, 0},
            {10 * time.Microsecond, 100 * time.Microsecond, 0},
            {100 * time.Microsecond, time.Millisecond, 0},
            {time.Millisecond, 10 * time.Millisecond, 0},
            {10 * time.Millisecond, 100 * time.Millisecond, 0},
            {100 * time.Millisecond, time.Second, 0},
            {time.Second, 10 * time.Second, 0},
        },
    }
}

func (th *TimeHistogram) Record(duration time.Duration) {
    th.mu.Lock()
    defer th.mu.Unlock()
    
    for i := range th.buckets {
        if duration >= th.buckets[i].LowerBound && duration < th.buckets[i].UpperBound {
            th.buckets[i].Count++
            return
        }
    }
    
    // Duration exceeds all buckets
    if len(th.buckets) > 0 {
        th.buckets[len(th.buckets)-1].Count++
    }
}

func (th *TimeHistogram) String() string {
    th.mu.RLock()
    defer th.mu.RUnlock()
    
    var result string
    for i, bucket := range th.buckets {
        result += fmt.Sprintf("\n    %v - %v: %d", bucket.LowerBound, bucket.UpperBound, bucket.Count)
    }
    return result
}

// Utility function to get goroutine ID (simplified)
func getGoroutineID() int {
    // In real implementation, you would extract this from runtime stack
    // This is a simplified placeholder
    return int(time.Now().UnixNano() % 10000)
}

func demonstrateContentionAnalysis() {
    fmt.Println("=== MUTEX CONTENTION ANALYSIS ===")
    
    analyzer := NewContentionAnalyzer()
    analyzer.Enable()
    defer analyzer.Disable()
    
    // Create tracked mutexes for different scenarios
    highContentionMutex := analyzer.CreateTrackedMutex("high_contention")
    mediumContentionMutex := analyzer.CreateTrackedMutex("medium_contention")
    lowContentionMutex := analyzer.CreateTrackedMutex("low_contention")
    
    var wg sync.WaitGroup
    
    // High contention scenario - many goroutines, long hold times
    fmt.Println("Testing high contention scenario...")
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 50; j++ {
                highContentionMutex.Lock()
                time.Sleep(time.Millisecond * 5) // Long hold time
                highContentionMutex.Unlock()
                time.Sleep(time.Microsecond * 100) // Short gap
            }
        }(i)
    }
    
    // Medium contention scenario - moderate competition
    fmt.Println("Testing medium contention scenario...")
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                mediumContentionMutex.Lock()
                time.Sleep(time.Millisecond * 2) // Medium hold time
                mediumContentionMutex.Unlock()
                time.Sleep(time.Millisecond * 1) // Medium gap
            }
        }(i)
    }
    
    // Low contention scenario - minimal competition
    fmt.Println("Testing low contention scenario...")
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 200; j++ {
                lowContentionMutex.Lock()
                time.Sleep(time.Microsecond * 100) // Short hold time
                lowContentionMutex.Unlock()
                time.Sleep(time.Millisecond * 5) // Long gap
            }
        }(i)
    }
    
    wg.Wait()
    
    // Generate comprehensive report
    report := analyzer.GetContentionReport()
    fmt.Printf("\n%s\n", report)
}
```

## Advanced Contention Detection

### Lock-Free Alternatives Analysis

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// ContentionComparator compares different synchronization approaches
type ContentionComparator struct {
    mutexResults    BenchmarkResult
    rwMutexResults  BenchmarkResult
    atomicResults   BenchmarkResult
    channelResults  BenchmarkResult
}

type BenchmarkResult struct {
    Name           string
    Operations     int64
    Duration       time.Duration
    Contentions    int64
    Throughput     float64 // ops/sec
    Efficiency     float64 // relative to atomic baseline
}

func (cc *ContentionComparator) CompareSynchronization(workload int, goroutines int) {
    fmt.Printf("Comparing synchronization methods (workload: %d, goroutines: %d)\n", workload, goroutines)
    
    // Test mutex-based approach
    cc.mutexResults = cc.testMutex(workload, goroutines)
    
    // Test RWMutex-based approach
    cc.rwMutexResults = cc.testRWMutex(workload, goroutines)
    
    // Test atomic-based approach
    cc.atomicResults = cc.testAtomic(workload, goroutines)
    
    // Test channel-based approach
    cc.channelResults = cc.testChannel(workload, goroutines)
    
    cc.printComparisonReport()
}

func (cc *ContentionComparator) testMutex(workload, goroutines int) BenchmarkResult {
    var counter int64
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < workload; j++ {
                mu.Lock()
                counter++
                // Simulate some work
                time.Sleep(time.Nanosecond * 100)
                mu.Unlock()
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := int64(workload * goroutines)
    throughput := float64(totalOps) / duration.Seconds()
    
    return BenchmarkResult{
        Name:       "Mutex",
        Operations: totalOps,
        Duration:   duration,
        Throughput: throughput,
    }
}

func (cc *ContentionComparator) testRWMutex(workload, goroutines int) BenchmarkResult {
    var counter int64
    var mu sync.RWMutex
    var wg sync.WaitGroup
    
    start := time.Now()
    
    // Mix of readers and writers
    writers := goroutines / 4
    readers := goroutines - writers
    
    // Writers
    for i := 0; i < writers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < workload; j++ {
                mu.Lock()
                counter++
                time.Sleep(time.Nanosecond * 100)
                mu.Unlock()
            }
        }()
    }
    
    // Readers
    for i := 0; i < readers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < workload; j++ {
                mu.RLock()
                _ = counter
                time.Sleep(time.Nanosecond * 50)
                mu.RUnlock()
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := int64(workload * goroutines)
    throughput := float64(totalOps) / duration.Seconds()
    
    return BenchmarkResult{
        Name:       "RWMutex",
        Operations: totalOps,
        Duration:   duration,
        Throughput: throughput,
    }
}

func (cc *ContentionComparator) testAtomic(workload, goroutines int) BenchmarkResult {
    var counter int64
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < workload; j++ {
                atomic.AddInt64(&counter, 1)
                // Simulate some work
                time.Sleep(time.Nanosecond * 100)
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := int64(workload * goroutines)
    throughput := float64(totalOps) / duration.Seconds()
    
    return BenchmarkResult{
        Name:       "Atomic",
        Operations: totalOps,
        Duration:   duration,
        Throughput: throughput,
    }
}

func (cc *ContentionComparator) testChannel(workload, goroutines int) BenchmarkResult {
    counterCh := make(chan int64, 1)
    counterCh <- 0 // Initialize counter
    
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < workload; j++ {
                // Get current value
                current := <-counterCh
                // Increment and put back
                counterCh <- current + 1
                // Simulate some work
                time.Sleep(time.Nanosecond * 100)
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := int64(workload * goroutines)
    throughput := float64(totalOps) / duration.Seconds()
    
    return BenchmarkResult{
        Name:       "Channel",
        Operations: totalOps,
        Duration:   duration,
        Throughput: throughput,
    }
}

func (cc *ContentionComparator) printComparisonReport() {
    results := []BenchmarkResult{
        cc.atomicResults,   // Baseline
        cc.mutexResults,
        cc.rwMutexResults,
        cc.channelResults,
    }
    
    // Calculate efficiency relative to atomic
    baseline := cc.atomicResults.Throughput
    for i := range results {
        if baseline > 0 {
            results[i].Efficiency = results[i].Throughput / baseline
        }
    }
    
    fmt.Printf("\nSynchronization Method Comparison:\n")
    fmt.Printf("%-10s %12s %12s %12s %10s\n", "Method", "Duration", "Ops/Sec", "Efficiency", "Relative")
    fmt.Printf("%s\n", strings.Repeat("-", 65))
    
    for _, result := range results {
        relative := ""
        if result.Name != "Atomic" {
            speedup := cc.atomicResults.Duration.Seconds() / result.Duration.Seconds()
            if speedup > 1 {
                relative = fmt.Sprintf("%.2fx slower", speedup)
            } else {
                relative = fmt.Sprintf("%.2fx faster", 1/speedup)
            }
        } else {
            relative = "baseline"
        }
        
        fmt.Printf("%-10s %12v %12.0f %12.2f %10s\n",
            result.Name,
            result.Duration,
            result.Throughput,
            result.Efficiency,
            relative)
    }
}

func demonstrateContentionComparison() {
    fmt.Println("\n=== SYNCHRONIZATION METHOD COMPARISON ===")
    
    comparator := &ContentionComparator{}
    
    // Test different scenarios
    scenarios := []struct {
        workload   int
        goroutines int
    }{
        {1000, 2},   // Low contention
        {1000, 10},  // Medium contention
        {1000, 50},  // High contention
    }
    
    for _, scenario := range scenarios {
        fmt.Printf("\n")
        comparator.CompareSynchronization(scenario.workload, scenario.goroutines)
    }
}
```

## Optimization Strategies

### 1. Lock Granularity Optimization

```go
// Fine-grained locking
type ShardedCounter struct {
    shards []struct {
        mu    sync.Mutex
        count int64
    }
    numShards int
}

func NewShardedCounter(numShards int) *ShardedCounter {
    sc := &ShardedCounter{
        shards:    make([]struct{mu sync.Mutex; count int64}, numShards),
        numShards: numShards,
    }
    return sc
}

func (sc *ShardedCounter) Increment(key string) {
    shard := sc.getShard(key)
    shard.mu.Lock()
    shard.count++
    shard.mu.Unlock()
}

func (sc *ShardedCounter) getShard(key string) *struct{mu sync.Mutex; count int64} {
    hash := fnv32(key)
    return &sc.shards[hash%uint32(sc.numShards)]
}
```

### 2. Read-Write Separation

```go
// Use RWMutex for read-heavy workloads
type CacheWithRWLock struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func (c *CacheWithRWLock) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    value, exists := c.data[key]
    return value, exists
}

func (c *CacheWithRWLock) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

### 3. Lock-Free Alternatives

```go
// Atomic operations for simple cases
type AtomicCounter struct {
    value int64
}

func (ac *AtomicCounter) Increment() int64 {
    return atomic.AddInt64(&ac.value, 1)
}

func (ac *AtomicCounter) Get() int64 {
    return atomic.LoadInt64(&ac.value)
}
```

## Best Practices

### 1. Minimize Lock Scope

```go
// Bad: Long lock scope
func badExample(mu *sync.Mutex, data map[string]int) {
    mu.Lock()
    defer mu.Unlock()
    
    // Long computation under lock
    result := expensiveComputation()
    data["result"] = result
}

// Good: Minimal lock scope
func goodExample(mu *sync.Mutex, data map[string]int) {
    // Compute outside lock
    result := expensiveComputation()
    
    mu.Lock()
    data["result"] = result
    mu.Unlock()
}
```

### 2. Consistent Lock Ordering

```go
// Avoid deadlocks with consistent ordering
func transferFunds(from, to *Account, amount int) {
    // Always lock accounts in consistent order (e.g., by ID)
    first, second := from, to
    if from.ID > to.ID {
        first, second = to, from
    }
    
    first.mu.Lock()
    defer first.mu.Unlock()
    
    second.mu.Lock()
    defer second.mu.Unlock()
    
    // Transfer logic
}
```

### 3. Profile-Guided Optimization

```bash
# Enable mutex profiling
export GOMAXPROCS=4
go tool pprof http://localhost:6060/debug/pprof/mutex

# Analyze contention
(pprof) top10
(pprof) tree
(pprof) web
```

## Next Steps

- Study [Block Profiling](block-profiling.md) analysis
- Learn [Channel Analysis](channel-analysis.md) techniques
- Explore [Lock-Free Programming](../../optimization/concurrency/lock-free.md)
- Master [Goroutine Patterns](../../optimization/concurrency/goroutine-patterns.md)

## Summary

Mutex contention analysis enables building efficient concurrent applications by:

1. **Identifying bottlenecks** - Finding high-contention locks
2. **Measuring impact** - Quantifying contention costs
3. **Comparing alternatives** - Evaluating different synchronization methods
4. **Optimizing granularity** - Right-sizing lock scope and frequency
5. **Validating improvements** - Confirming optimization benefits

Use these techniques to eliminate synchronization bottlenecks and maximize concurrent performance.
