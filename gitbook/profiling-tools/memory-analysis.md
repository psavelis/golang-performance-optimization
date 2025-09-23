# Memory Analysis

Master comprehensive memory analysis techniques to identify leaks, optimize allocation patterns, and ensure efficient memory usage in Go applications.

## Memory Profiling Fundamentals

Go provides powerful tools for analyzing memory usage patterns, from basic heap profiling to advanced allocation tracking and leak detection.

### Types of Memory Profiles

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

var (
    memprofile = flag.String("memprofile", "", "write memory profile to file")
    allocprofile = flag.String("allocprofile", "", "write allocation profile to file")
)

func main() {
    flag.Parse()
    
    // Demonstrate different allocation patterns
    demonstrateHeapAllocations()
    demonstrateStackAllocations()
    demonstrateMemoryLeaks()
    
    // Capture heap profile (current in-use memory)
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
    
    // Capture allocation profile (all allocations since start)
    if *allocprofile != "" {
        f, err := os.Create(*allocprofile)
        if err != nil {
            panic(err)
        }
        defer f.Close()
        
        if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
            panic(err)
        }
    }
}

func demonstrateHeapAllocations() {
    // Large slice allocation
    data := make([]byte, 10*1024*1024) // 10MB
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    // Keep reference to prevent GC
    _ = data
    
    // Multiple small allocations
    var smallSlices [][]byte
    for i := 0; i < 1000; i++ {
        slice := make([]byte, 1024) // 1KB each
        smallSlices = append(smallSlices, slice)
    }
    
    _ = smallSlices
}

func demonstrateStackAllocations() {
    // This might be allocated on stack if escape analysis determines it doesn't escape
    localData := make([]int, 100)
    processLocalData(localData)
}

func processLocalData(data []int) {
    for i := range data {
        data[i] = i * i
    }
}

func demonstrateMemoryLeaks() {
    // Simulate a memory leak scenario
    leakyMap := make(map[string][]byte)
    
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        value := make([]byte, 1024)
        leakyMap[key] = value
        
        // Simulate cleanup that misses some entries
        if i%100 == 0 && i > 0 {
            delete(leakyMap, fmt.Sprintf("key_%d", i-50))
        }
    }
    
    // Keep reference to demonstrate the leak
    _ = leakyMap
}
```

## Advanced Memory Analysis

### Real-time Memory Monitoring

```go
package main

import (
    "context"
    "runtime"
    "time"
)

type MemoryMonitor struct {
    interval time.Duration
    callback func(MemoryStats)
}

type MemoryStats struct {
    Timestamp    time.Time
    Alloc        uint64 // Currently allocated bytes
    TotalAlloc   uint64 // Total allocated bytes (cumulative)
    Sys          uint64 // System memory obtained from OS
    Lookups      uint64 // Number of pointer lookups
    Mallocs      uint64 // Number of malloc calls
    Frees        uint64 // Number of free calls
    HeapAlloc    uint64 // Heap allocated bytes
    HeapSys      uint64 // Heap system bytes
    HeapIdle     uint64 // Heap idle bytes
    HeapInuse    uint64 // Heap in-use bytes
    HeapReleased uint64 // Heap released bytes
    HeapObjects  uint64 // Number of allocated heap objects
    StackInuse   uint64 // Stack in-use bytes
    StackSys     uint64 // Stack system bytes
    GCCPUFraction float64 // Fraction of CPU time used by GC
    NumGC        uint32  // Number of completed GC cycles
    LastGC       time.Time // Time of last GC
    PauseTotalNs uint64  // Total GC pause time in nanoseconds
}

func NewMemoryMonitor(interval time.Duration, callback func(MemoryStats)) *MemoryMonitor {
    return &MemoryMonitor{
        interval: interval,
        callback: callback,
    }
}

func (mm *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(mm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            stats := mm.collectMemoryStats()
            mm.callback(stats)
        }
    }
}

func (mm *MemoryMonitor) collectMemoryStats() MemoryStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return MemoryStats{
        Timestamp:     time.Now(),
        Alloc:         m.Alloc,
        TotalAlloc:    m.TotalAlloc,
        Sys:           m.Sys,
        Lookups:       m.Lookups,
        Mallocs:       m.Mallocs,
        Frees:         m.Frees,
        HeapAlloc:     m.HeapAlloc,
        HeapSys:       m.HeapSys,
        HeapIdle:      m.HeapIdle,
        HeapInuse:     m.HeapInuse,
        HeapReleased:  m.HeapReleased,
        HeapObjects:   m.HeapObjects,
        StackInuse:    m.StackInuse,
        StackSys:      m.StackSys,
        GCCPUFraction: m.GCCPUFraction,
        NumGC:         m.NumGC,
        LastGC:        time.Unix(0, int64(m.LastGC)),
        PauseTotalNs:  m.PauseTotalNs,
    }
}

func (ms MemoryStats) String() string {
    return fmt.Sprintf(
        "Memory Stats [%s]:\n"+
            "  Allocated: %s\n"+
            "  System: %s\n"+
            "  Heap In-use: %s\n"+
            "  Heap Objects: %d\n"+
            "  GC Cycles: %d\n"+
            "  GC CPU%%: %.2f\n"+
            "  Total GC Pause: %v\n",
        ms.Timestamp.Format(time.RFC3339),
        formatBytes(ms.Alloc),
        formatBytes(ms.Sys),
        formatBytes(ms.HeapInuse),
        ms.HeapObjects,
        ms.NumGC,
        ms.GCCPUFraction*100,
        time.Duration(ms.PauseTotalNs),
    )
}

func formatBytes(bytes uint64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Example usage
func demonstrateMemoryMonitoring() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    monitor := NewMemoryMonitor(time.Second, func(stats MemoryStats) {
        fmt.Println(stats)
        
        // Alert on high memory usage
        if stats.HeapInuse > 100*1024*1024 { // 100MB
            fmt.Printf("⚠️  High memory usage detected: %s\n", formatBytes(stats.HeapInuse))
        }
        
        // Alert on frequent GC
        if stats.GCCPUFraction > 0.1 { // >10% CPU spent on GC
            fmt.Printf("⚠️  High GC overhead: %.2f%%\n", stats.GCCPUFraction*100)
        }
    })
    
    go monitor.Start(ctx)
    
    // Simulate memory allocation patterns
    simulateMemoryWorkload()
}

func simulateMemoryWorkload() {
    var data [][]byte
    
    for i := 0; i < 1000; i++ {
        // Allocate chunks of varying sizes
        size := 1024 * (1 + i%100)
        chunk := make([]byte, size)
        data = append(data, chunk)
        
        // Occasionally free some memory
        if i%50 == 0 && len(data) > 10 {
            data = data[10:] // Remove old chunks
        }
        
        time.Sleep(10 * time.Millisecond)
    }
}
```

### Memory Leak Detection

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
)

type LeakDetector struct {
    baseline     MemoryStats
    threshold    uint64
    checkCount   int
    alertChannel chan LeakAlert
}

type LeakAlert struct {
    Timestamp   time.Time
    Growth      uint64
    Type        string
    Description string
}

func NewLeakDetector(threshold uint64) *LeakDetector {
    return &LeakDetector{
        threshold:    threshold,
        alertChannel: make(chan LeakAlert, 10),
    }
}

func (ld *LeakDetector) StartMonitoring(ctx context.Context, interval time.Duration) {
    // Establish baseline
    ld.baseline = collectMemoryStats()
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ld.checkForLeaks()
        }
    }
}

func (ld *LeakDetector) checkForLeaks() {
    current := collectMemoryStats()
    ld.checkCount++
    
    // Check heap growth
    heapGrowth := current.HeapInuse - ld.baseline.HeapInuse
    if heapGrowth > ld.threshold {
        alert := LeakAlert{
            Timestamp:   time.Now(),
            Growth:      heapGrowth,
            Type:        "heap_growth",
            Description: fmt.Sprintf("Heap grew by %s since baseline", formatBytes(heapGrowth)),
        }
        
        select {
        case ld.alertChannel <- alert:
        default:
            // Channel full, drop alert
        }
    }
    
    // Check for memory not being freed
    allocDelta := current.Mallocs - current.Frees
    baselineAllocDelta := ld.baseline.Mallocs - ld.baseline.Frees
    
    if allocDelta > baselineAllocDelta+10000 { // 10k objects not freed
        alert := LeakAlert{
            Timestamp:   time.Now(),
            Growth:      allocDelta - baselineAllocDelta,
            Type:        "object_leak",
            Description: fmt.Sprintf("%d objects allocated but not freed", allocDelta-baselineAllocDelta),
        }
        
        select {
        case ld.alertChannel <- alert:
        default:
        }
    }
    
    // Update baseline periodically
    if ld.checkCount%10 == 0 {
        ld.baseline = current
    }
}

func (ld *LeakDetector) GetAlerts() <-chan LeakAlert {
    return ld.alertChannel
}

func collectMemoryStats() MemoryStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return MemoryStats{
        Timestamp:   time.Now(),
        Alloc:       m.Alloc,
        TotalAlloc:  m.TotalAlloc,
        Sys:         m.Sys,
        Mallocs:     m.Mallocs,
        Frees:       m.Frees,
        HeapAlloc:   m.HeapAlloc,
        HeapInuse:   m.HeapInuse,
        HeapObjects: m.HeapObjects,
    }
}

// Example of potential memory leak patterns
func demonstrateLeakPatterns() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    
    detector := NewLeakDetector(10 * 1024 * 1024) // 10MB threshold
    
    go detector.StartMonitoring(ctx, 5*time.Second)
    
    // Monitor alerts
    go func() {
        for alert := range detector.GetAlerts() {
            fmt.Printf("🚨 MEMORY LEAK ALERT: %s - %s\n", alert.Type, alert.Description)
        }
    }()
    
    // Simulate different leak patterns
    go simulateSlowLeak()
    go simulateEventualGrowth()
    go simulateOscillatingMemory()
    
    <-ctx.Done()
}

func simulateSlowLeak() {
    var leakySlice [][]byte
    
    for i := 0; i < 1000; i++ {
        // Gradually accumulating memory that's never freed
        chunk := make([]byte, 1024*10) // 10KB
        leakySlice = append(leakySlice, chunk)
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Keep reference to prevent GC
    _ = leakySlice
}

func simulateEventualGrowth() {
    cache := make(map[string][]byte)
    
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        value := make([]byte, 512)
        cache[key] = value
        
        // Only occasionally clean up
        if i%1000 == 0 && i > 0 {
            // Clean only half
            for j := i - 500; j < i; j++ {
                delete(cache, fmt.Sprintf("key_%d", j))
            }
        }
        
        time.Sleep(5 * time.Millisecond)
    }
    
    _ = cache
}

func simulateOscillatingMemory() {
    for cycle := 0; cycle < 10; cycle++ {
        var temp [][]byte
        
        // Allocate phase
        for i := 0; i < 100; i++ {
            chunk := make([]byte, 1024*100) // 100KB
            temp = append(temp, chunk)
        }
        
        time.Sleep(5 * time.Second)
        
        // Free phase
        temp = nil
        runtime.GC()
        
        time.Sleep(2 * time.Second)
    }
}
```

### Garbage Collection Analysis

```go
package main

import (
    "fmt"
    "runtime"
    "runtime/debug"
    "time"
)

type GCAnalyzer struct {
    history []GCStats
}

type GCStats struct {
    Timestamp     time.Time
    NumGC         uint32
    PauseTotal    time.Duration
    PauseRecent   time.Duration
    GCCPUFraction float64
    HeapSize      uint64
    AllocRate     float64 // bytes/second
}

func NewGCAnalyzer() *GCAnalyzer {
    return &GCAnalyzer{
        history: make([]GCStats, 0),
    }
}

func (gca *GCAnalyzer) CollectStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    stats := GCStats{
        Timestamp:     time.Now(),
        NumGC:         m.NumGC,
        PauseTotal:    time.Duration(m.PauseTotalNs),
        GCCPUFraction: m.GCCPUFraction,
        HeapSize:      m.HeapInuse,
    }
    
    // Calculate recent pause time
    if len(gca.history) > 0 {
        last := gca.history[len(gca.history)-1]
        stats.PauseRecent = stats.PauseTotal - last.PauseTotal
        
        // Calculate allocation rate
        timeDelta := stats.Timestamp.Sub(last.Timestamp).Seconds()
        bytesDelta := float64(m.TotalAlloc - uint64(len(gca.history))*1000000) // Rough estimate
        stats.AllocRate = bytesDelta / timeDelta
    }
    
    gca.history = append(gca.history, stats)
    
    // Keep only last 100 entries
    if len(gca.history) > 100 {
        gca.history = gca.history[1:]
    }
}

func (gca *GCAnalyzer) AnalyzePatterns() GCAnalysis {
    if len(gca.history) < 2 {
        return GCAnalysis{}
    }
    
    analysis := GCAnalysis{
        TotalSamples: len(gca.history),
    }
    
    var totalPause time.Duration
    var maxPause time.Duration
    var totalGCTime float64
    var heapSizes []uint64
    
    for i, stats := range gca.history {
        if i == 0 {
            continue
        }
        
        pause := stats.PauseRecent
        totalPause += pause
        
        if pause > maxPause {
            maxPause = pause
        }
        
        totalGCTime += stats.GCCPUFraction
        heapSizes = append(heapSizes, stats.HeapSize)
    }
    
    analysis.AveragePause = totalPause / time.Duration(len(gca.history)-1)
    analysis.MaxPause = maxPause
    analysis.AverageGCCPU = totalGCTime / float64(len(gca.history)-1)
    
    // Calculate heap growth trend
    if len(heapSizes) > 1 {
        first := heapSizes[0]
        last := heapSizes[len(heapSizes)-1]
        analysis.HeapGrowthTrend = float64(int64(last)-int64(first)) / float64(first) * 100
    }
    
    return analysis
}

type GCAnalysis struct {
    TotalSamples     int
    AveragePause     time.Duration
    MaxPause         time.Duration
    AverageGCCPU     float64
    HeapGrowthTrend  float64 // percentage
}

func (gca GCAnalysis) String() string {
    return fmt.Sprintf(
        "GC Analysis:\n"+
            "  Samples: %d\n"+
            "  Average Pause: %v\n"+
            "  Max Pause: %v\n"+
            "  Average GC CPU: %.2f%%\n"+
            "  Heap Growth: %.2f%%\n",
        gca.TotalSamples,
        gca.AveragePause,
        gca.MaxPause,
        gca.AverageGCCPU*100,
        gca.HeapGrowthTrend,
    )
}

func demonstrateGCAnalysis() {
    analyzer := NewGCAnalyzer()
    
    // Collect stats every second
    go func() {
        for i := 0; i < 60; i++ {
            analyzer.CollectStats()
            time.Sleep(time.Second)
        }
    }()
    
    // Simulate memory allocation patterns
    simulateGCWorkload()
    
    // Print analysis
    time.Sleep(time.Second)
    analysis := analyzer.AnalyzePatterns()
    fmt.Println(analysis)
    
    // Recommendations based on analysis
    if analysis.AveragePause > 10*time.Millisecond {
        fmt.Println("⚠️  Recommendation: Consider optimizing allocation patterns to reduce GC pressure")
    }
    
    if analysis.AverageGCCPU > 0.05 {
        fmt.Println("⚠️  Recommendation: GC overhead is high, review memory allocation frequency")
    }
    
    if analysis.HeapGrowthTrend > 50 {
        fmt.Println("⚠️  Recommendation: Heap is growing rapidly, check for memory leaks")
    }
}

func simulateGCWorkload() {
    // Different allocation patterns to stress GC
    
    // Pattern 1: Many small allocations
    for i := 0; i < 10000; i++ {
        _ = make([]byte, 100)
    }
    
    // Pattern 2: Few large allocations
    var large [][]byte
    for i := 0; i < 10; i++ {
        large = append(large, make([]byte, 1024*1024)) // 1MB
    }
    
    // Pattern 3: Oscillating allocations
    for cycle := 0; cycle < 5; cycle++ {
        var temp [][]byte
        for i := 0; i < 1000; i++ {
            temp = append(temp, make([]byte, 1024))
        }
        temp = nil // Release for GC
        runtime.GC()
        time.Sleep(2 * time.Second)
    }
}

// GC Tuning utilities
func optimizeGCSettings() {
    // Set GC target percentage (default is 100)
    debug.SetGCPercent(50) // More aggressive GC
    
    // Force GC to understand current behavior
    runtime.GC()
    
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    fmt.Printf("GC Settings Applied:\n")
    fmt.Printf("  GC Target: %d%%\n", debug.SetGCPercent(-1)) // Returns current value
    fmt.Printf("  Current Heap: %s\n", formatBytes(m.HeapInuse))
    fmt.Printf("  Next GC Target: %s\n", formatBytes(m.NextGC))
}
```

Memory analysis in Go provides comprehensive insights into allocation patterns, garbage collection behavior, and potential memory leaks, enabling you to optimize memory usage and maintain efficient application performance.
