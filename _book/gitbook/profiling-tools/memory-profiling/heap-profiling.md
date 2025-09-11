# Heap Profiling

Heap profiling reveals memory allocation patterns, identifies memory leaks, and guides memory optimization strategies. This comprehensive guide covers heap profiling techniques, analysis methods, and optimization approaches for Go applications.

## Introduction to Heap Profiling

Heap profiling captures snapshots of memory allocations, showing:
- **What objects** are allocated
- **Where allocations** occur in code
- **How much memory** each allocation type consumes
- **Call stacks** leading to allocations

### Key Concepts

- **Live Objects**: Currently allocated and referenced objects
- **Allocation Sites**: Code locations where objects are created
- **Object Types**: Specific types being allocated
- **Size Distribution**: Memory usage patterns across object types

## Enabling Heap Profiling

### Using net/http/pprof

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
    "time"
)

func main() {
    // Enable pprof endpoints
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Simulate application workload
    go memoryIntensiveWork()
    
    // Keep the application running
    select {}
}

func memoryIntensiveWork() {
    data := make(map[string][]byte)
    
    for i := 0; ; i++ {
        // Allocate varying sizes of data
        size := 1024 + (i%10)*1024
        key := fmt.Sprintf("data_%d", i)
        data[key] = make([]byte, size)
        
        // Periodically clean up old data
        if i%100 == 0 {
            for k := range data {
                if len(data) > 50 {
                    delete(data, k)
                    break
                }
            }
        }
        
        time.Sleep(10 * time.Millisecond)
    }
}
```

### Programmatic Heap Profiling

```go
package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

func captureHeapProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("could not create heap profile: %v", err)
    }
    defer f.Close()

    // Force garbage collection before profiling
    runtime.GC()
    
    if err := pprof.WriteHeapProfile(f); err != nil {
        return fmt.Errorf("could not write heap profile: %v", err)
    }

    return nil
}

func demonstrateHeapProfiling() {
    // Initial heap profile
    if err := captureHeapProfile("heap_initial.prof"); err != nil {
        panic(err)
    }

    // Perform memory allocations
    data := allocateTestData()
    
    // Profile after allocations
    if err := captureHeapProfile("heap_after_alloc.prof"); err != nil {
        panic(err)
    }

    // Use the data to prevent optimization
    fmt.Printf("Allocated %d items\n", len(data))
    
    // Final profile
    runtime.GC() // Force cleanup
    if err := captureHeapProfile("heap_final.prof"); err != nil {
        panic(err)
    }
}

func allocateTestData() [][]byte {
    data := make([][]byte, 1000)
    for i := range data {
        // Allocate varying sizes
        size := 1024 * (1 + i%10)
        data[i] = make([]byte, size)
    }
    return data
}
```

## Collecting Heap Profiles

### HTTP Endpoints

```bash
# Get current heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Get allocation profile (since program start)
curl http://localhost:6060/debug/pprof/allocs > allocs.prof

# Profile with specific sampling rate
curl "http://localhost:6060/debug/pprof/heap?gc=1" > heap_after_gc.prof
```

### Command Line Analysis

```bash
# Interactive analysis
go tool pprof heap.prof

# Web interface
go tool pprof -http=:8080 heap.prof

# Direct analysis
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Advanced Heap Profiling Techniques

### Continuous Heap Monitoring

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "runtime"
    "runtime/pprof"
    "sync"
    "time"
)

type HeapMonitor struct {
    mu              sync.RWMutex
    enabled         bool
    interval        time.Duration
    threshold       int64 // Memory threshold in bytes
    profileDir      string
    lastProfile     time.Time
    memoryHistory   []MemorySnapshot
    maxHistorySize  int
}

type MemorySnapshot struct {
    Timestamp    time.Time
    HeapAlloc    uint64
    HeapSys      uint64
    HeapInuse    uint64
    HeapReleased uint64
    GCCount      uint32
}

func NewHeapMonitor(interval time.Duration, threshold int64, profileDir string) *HeapMonitor {
    return &HeapMonitor{
        interval:       interval,
        threshold:      threshold,
        profileDir:     profileDir,
        maxHistorySize: 100,
        memoryHistory:  make([]MemorySnapshot, 0, 100),
    }
}

func (hm *HeapMonitor) Start(ctx context.Context) error {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    if hm.enabled {
        return fmt.Errorf("heap monitor already running")
    }

    // Ensure profile directory exists
    if err := os.MkdirAll(hm.profileDir, 0755); err != nil {
        return fmt.Errorf("failed to create profile directory: %v", err)
    }

    hm.enabled = true
    go hm.monitorLoop(ctx)

    log.Printf("Heap monitor started: interval=%v, threshold=%d bytes", hm.interval, hm.threshold)
    return nil
}

func (hm *HeapMonitor) Stop() {
    hm.mu.Lock()
    defer hm.mu.Unlock()
    hm.enabled = false
    log.Println("Heap monitor stopped")
}

func (hm *HeapMonitor) monitorLoop(ctx context.Context) {
    ticker := time.NewTicker(hm.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hm.checkMemoryUsage()
        }
    }
}

func (hm *HeapMonitor) checkMemoryUsage() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)

    // Record memory snapshot
    snapshot := MemorySnapshot{
        Timestamp:    time.Now(),
        HeapAlloc:    stats.HeapAlloc,
        HeapSys:      stats.HeapSys,
        HeapInuse:    stats.HeapInuse,
        HeapReleased: stats.HeapReleased,
        GCCount:      stats.NumGC,
    }

    hm.addMemorySnapshot(snapshot)

    // Check if memory usage exceeds threshold
    if int64(stats.HeapAlloc) > hm.threshold {
        hm.triggerProfile("threshold_exceeded")
    }

    // Check for potential memory leaks
    if hm.detectMemoryLeak() {
        hm.triggerProfile("potential_leak")
    }

    // Log memory statistics
    log.Printf("Memory: HeapAlloc=%s, HeapSys=%s, HeapInuse=%s, GC=%d",
        formatBytes(stats.HeapAlloc),
        formatBytes(stats.HeapSys),
        formatBytes(stats.HeapInuse),
        stats.NumGC)
}

func (hm *HeapMonitor) addMemorySnapshot(snapshot MemorySnapshot) {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    hm.memoryHistory = append(hm.memoryHistory, snapshot)
    
    // Trim history if too large
    if len(hm.memoryHistory) > hm.maxHistorySize {
        hm.memoryHistory = hm.memoryHistory[len(hm.memoryHistory)-hm.maxHistorySize:]
    }
}

func (hm *HeapMonitor) detectMemoryLeak() bool {
    hm.mu.RLock()
    defer hm.mu.RUnlock()

    if len(hm.memoryHistory) < 10 {
        return false
    }

    // Check if memory is consistently growing
    recentSnapshots := hm.memoryHistory[len(hm.memoryHistory)-10:]
    growthCount := 0

    for i := 1; i < len(recentSnapshots); i++ {
        if recentSnapshots[i].HeapAlloc > recentSnapshots[i-1].HeapAlloc {
            growthCount++
        }
    }

    // If memory grew in 80% of recent samples, consider it a potential leak
    return float64(growthCount)/float64(len(recentSnapshots)-1) > 0.8
}

func (hm *HeapMonitor) triggerProfile(reason string) {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    // Rate limit profiling
    if time.Since(hm.lastProfile) < 5*time.Minute {
        return
    }

    timestamp := time.Now().Format("20060102_150405")
    filename := filepath.Join(hm.profileDir, fmt.Sprintf("heap_%s_%s.prof", reason, timestamp))
    
    if err := captureHeapProfile(filename); err != nil {
        log.Printf("Failed to capture heap profile: %v", err)
        return
    }

    hm.lastProfile = time.Now()
    log.Printf("Heap profile captured: %s (reason: %s)", filename, reason)
}

func (hm *HeapMonitor) GetMemoryHistory() []MemorySnapshot {
    hm.mu.RLock()
    defer hm.mu.RUnlock()
    
    // Return a copy
    history := make([]MemorySnapshot, len(hm.memoryHistory))
    copy(history, hm.memoryHistory)
    return history
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
```

### Allocation Tracking

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

type AllocationTracker struct {
    mu            sync.RWMutex
    allocations   map[string]*AllocationInfo
    enabled       bool
    threshold     int64
}

type AllocationInfo struct {
    Count       int64
    TotalBytes  int64
    LastSeen    time.Time
    CallStack   []uintptr
}

func NewAllocationTracker(threshold int64) *AllocationTracker {
    return &AllocationTracker{
        allocations: make(map[string]*AllocationInfo),
        threshold:   threshold,
    }
}

func (at *AllocationTracker) Start() {
    at.mu.Lock()
    defer at.mu.Unlock()
    at.enabled = true
    
    // Set memory profiling rate
    runtime.MemProfileRate = 1024 // Sample every 1KB
    
    fmt.Println("Allocation tracker started")
}

func (at *AllocationTracker) Stop() {
    at.mu.Lock()
    defer at.mu.Unlock()
    at.enabled = false
    
    // Reset memory profiling rate
    runtime.MemProfileRate = 512 * 1024 // Default
    
    fmt.Println("Allocation tracker stopped")
}

func (at *AllocationTracker) TrackAllocation(size int64, location string) {
    if !at.enabled {
        return
    }

    at.mu.Lock()
    defer at.mu.Unlock()

    info, exists := at.allocations[location]
    if !exists {
        info = &AllocationInfo{
            CallStack: make([]uintptr, 10),
        }
        at.allocations[location] = info
        
        // Capture call stack
        runtime.Callers(2, info.CallStack)
    }

    info.Count++
    info.TotalBytes += size
    info.LastSeen = time.Now()

    // Log large allocations
    if size > at.threshold {
        fmt.Printf("Large allocation: %d bytes at %s\n", size, location)
    }
}

func (at *AllocationTracker) GetReport() string {
    at.mu.RLock()
    defer at.mu.RUnlock()

    var report strings.Builder
    report.WriteString("=== ALLOCATION REPORT ===\n\n")

    // Sort by total bytes
    type allocPair struct {
        location string
        info     *AllocationInfo
    }

    var pairs []allocPair
    for location, info := range at.allocations {
        pairs = append(pairs, allocPair{location, info})
    }

    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].info.TotalBytes > pairs[j].info.TotalBytes
    })

    // Generate report
    for i, pair := range pairs {
        if i >= 20 { // Top 20
            break
        }

        avgSize := pair.info.TotalBytes / pair.info.Count
        report.WriteString(fmt.Sprintf("%2d. %s\n", i+1, pair.location))
        report.WriteString(fmt.Sprintf("    Count: %d, Total: %s, Avg: %s\n",
            pair.info.Count,
            formatBytes(uint64(pair.info.TotalBytes)),
            formatBytes(uint64(avgSize))))
        report.WriteString(fmt.Sprintf("    Last seen: %s\n\n",
            pair.info.LastSeen.Format("2006-01-02 15:04:05")))
    }

    return report.String()
}

// Example usage with allocation tracking
func demonstrateAllocationTracking() {
    tracker := NewAllocationTracker(1024 * 1024) // 1MB threshold
    tracker.Start()
    defer tracker.Stop()

    // Simulate different allocation patterns
    go heavyAllocator(tracker)
    go frequentAllocator(tracker)
    go burstyAllocator(tracker)

    time.Sleep(30 * time.Second)

    fmt.Println(tracker.GetReport())
}

func heavyAllocator(tracker *AllocationTracker) {
    for i := 0; i < 100; i++ {
        size := int64(1024 * 1024 * (1 + i%5)) // 1-5MB allocations
        data := make([]byte, size)
        tracker.TrackAllocation(size, "heavyAllocator")
        
        // Use data to prevent optimization
        data[0] = byte(i)
        
        time.Sleep(100 * time.Millisecond)
    }
}

func frequentAllocator(tracker *AllocationTracker) {
    for i := 0; i < 10000; i++ {
        size := int64(1024) // 1KB allocations
        data := make([]byte, size)
        tracker.TrackAllocation(size, "frequentAllocator")
        
        data[0] = byte(i)
        
        time.Sleep(time.Millisecond)
    }
}

func burstyAllocator(tracker *AllocationTracker) {
    for burst := 0; burst < 10; burst++ {
        for i := 0; i < 100; i++ {
            size := int64(64 * 1024) // 64KB allocations
            data := make([]byte, size)
            tracker.TrackAllocation(size, "burstyAllocator")
            
            data[0] = byte(i)
        }
        time.Sleep(5 * time.Second) // Burst interval
    }
}
```

## Analyzing Heap Profiles

### Understanding Profile Output

```bash
# Basic heap analysis
go tool pprof heap.prof

# Common pprof commands for heap analysis
(pprof) top           # Top memory consumers
(pprof) top -cum      # Top cumulative allocators
(pprof) list main     # Source code with allocation info
(pprof) web           # Visual call graph
(pprof) png           # Generate PNG image
(pprof) alloc_space   # Switch to allocation space view
(pprof) alloc_objects # Switch to allocation objects view
```

### Sample Analysis Session

```bash
# Example pprof session
$ go tool pprof heap.prof
File: myapp
Type: inuse_space
Time: Jan 2, 2023 at 3:04pm (UTC)
Entering interactive mode (type "help" for commands, "o" for options)

(pprof) top
Showing nodes accounting for 512.17MB, 89.24% of 574.01MB total
Dropped 45 nodes (cum <= 2.87MB)
      flat  flat%   sum%        cum   cum%
  256.05MB 44.62% 44.62%   256.05MB 44.62%  main.allocateData
  128.02MB 22.31% 66.93%   384.07MB 66.93%  main.processRecords
   64.01MB 11.15% 78.08%    64.01MB 11.15%  encoding/json.Marshal
   32.06MB  5.59% 83.67%    96.07MB 16.74%  net/http.(*Client).Do
   16.03MB  2.79% 86.46%    16.03MB  2.79%  crypto/tls.(*Conn).Read
   16.00MB  2.79% 89.24%    16.00MB  2.79%  bufio.NewReaderSize

(pprof) list allocateData
Total: 574.01MB
ROUTINE ======================== main.allocateData in /app/main.go
  256.05MB   256.05MB (flat, cum) 44.62% of Total
         .          .     45:func allocateData(size int) []byte {
         .          .     46:    // This line allocates most memory
  256.05MB   256.05MB     47:    return make([]byte, size)
         .          .     48:}
```

### Comparative Analysis

```go
package main

import (
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

func compareHeapProfiles(beforeFile, afterFile string) error {
    // Generate comparison using pprof
    cmd := exec.Command("go", "tool", "pprof", "-base", beforeFile, "-top", afterFile)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to compare profiles: %v", err)
    }

    analysis := parseProfileComparison(string(output))
    fmt.Println(analysis)
    
    return nil
}

func parseProfileComparison(output string) string {
    lines := strings.Split(output, "\n")
    var result strings.Builder
    
    result.WriteString("=== HEAP PROFILE COMPARISON ===\n\n")
    
    improvements := []string{}
    regressions := []string{}
    
    for _, line := range lines {
        if strings.Contains(line, "MB") && !strings.Contains(line, "Showing") {
            if strings.Contains(line, "-") {
                improvements = append(improvements, line)
            } else if strings.Contains(line, "+") {
                regressions = append(regressions, line)
            }
        }
    }
    
    if len(improvements) > 0 {
        result.WriteString("MEMORY IMPROVEMENTS:\n")
        for _, improvement := range improvements {
            result.WriteString(fmt.Sprintf("  %s\n", improvement))
        }
        result.WriteString("\n")
    }
    
    if len(regressions) > 0 {
        result.WriteString("MEMORY REGRESSIONS:\n")
        for _, regression := range regressions {
            result.WriteString(fmt.Sprintf("  %s\n", regression))
        }
        result.WriteString("\n")
    }
    
    return result.String()
}
```

## Memory Leak Detection

### Automated Leak Detection

```go
package main

import (
    "context"
    "fmt"
    "log"
    "runtime"
    "time"
)

type LeakDetector struct {
    baselineSet    bool
    baselineAlloc  uint64
    baselineGC     uint32
    threshold      float64 // Growth threshold (e.g., 1.5 = 50% growth)
    checkInterval  time.Duration
}

func NewLeakDetector(threshold float64, checkInterval time.Duration) *LeakDetector {
    return &LeakDetector{
        threshold:     threshold,
        checkInterval: checkInterval,
    }
}

func (ld *LeakDetector) Start(ctx context.Context) {
    // Set baseline after initial warmup
    time.Sleep(10 * time.Second)
    ld.setBaseline()
    
    ticker := time.NewTicker(ld.checkInterval)
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

func (ld *LeakDetector) setBaseline() {
    runtime.GC() // Force garbage collection
    runtime.GC() // Run twice to ensure cleanup
    
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    ld.baselineAlloc = stats.HeapAlloc
    ld.baselineGC = stats.NumGC
    ld.baselineSet = true
    
    log.Printf("Memory baseline set: %s", formatBytes(ld.baselineAlloc))
}

func (ld *LeakDetector) checkForLeaks() {
    if !ld.baselineSet {
        return
    }
    
    runtime.GC()
    runtime.GC()
    
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    currentAlloc := stats.HeapAlloc
    growth := float64(currentAlloc) / float64(ld.baselineAlloc)
    
    if growth > ld.threshold {
        log.Printf("POTENTIAL MEMORY LEAK DETECTED!")
        log.Printf("  Baseline: %s", formatBytes(ld.baselineAlloc))
        log.Printf("  Current:  %s", formatBytes(currentAlloc))
        log.Printf("  Growth:   %.2fx (threshold: %.2fx)", growth, ld.threshold)
        log.Printf("  GC Runs:  %d → %d", ld.baselineGC, stats.NumGC)
        
        // Capture heap profile for analysis
        filename := fmt.Sprintf("leak_detected_%d.prof", time.Now().Unix())
        if err := captureHeapProfile(filename); err != nil {
            log.Printf("Failed to capture leak profile: %v", err)
        } else {
            log.Printf("Heap profile saved: %s", filename)
        }
    } else {
        log.Printf("Memory check: %s (%.2fx baseline)", 
            formatBytes(currentAlloc), growth)
    }
}
```

### Memory Growth Analysis

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

type MemoryGrowthAnalyzer struct {
    samples    []MemorySample
    maxSamples int
}

type MemorySample struct {
    Timestamp time.Time
    HeapAlloc uint64
    HeapSys   uint64
    GCCount   uint32
    Goroutines int
}

func NewMemoryGrowthAnalyzer(maxSamples int) *MemoryGrowthAnalyzer {
    return &MemoryGrowthAnalyzer{
        samples:    make([]MemorySample, 0, maxSamples),
        maxSamples: maxSamples,
    }
}

func (mga *MemoryGrowthAnalyzer) RecordSample() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    sample := MemorySample{
        Timestamp:  time.Now(),
        HeapAlloc:  stats.HeapAlloc,
        HeapSys:    stats.HeapSys,
        GCCount:    stats.NumGC,
        Goroutines: runtime.NumGoroutine(),
    }
    
    mga.samples = append(mga.samples, sample)
    
    // Trim samples if we exceed maximum
    if len(mga.samples) > mga.maxSamples {
        mga.samples = mga.samples[len(mga.samples)-mga.maxSamples:]
    }
}

func (mga *MemoryGrowthAnalyzer) AnalyzeGrowth() string {
    if len(mga.samples) < 2 {
        return "Insufficient samples for analysis"
    }
    
    var result strings.Builder
    result.WriteString("=== MEMORY GROWTH ANALYSIS ===\n\n")
    
    first := mga.samples[0]
    last := mga.samples[len(mga.samples)-1]
    duration := last.Timestamp.Sub(first.Timestamp)
    
    // Calculate growth rates
    heapGrowth := float64(last.HeapAlloc) / float64(first.HeapAlloc)
    sysGrowth := float64(last.HeapSys) / float64(first.HeapSys)
    gcIncrease := last.GCCount - first.GCCount
    goroutineGrowth := float64(last.Goroutines) / float64(first.Goroutines)
    
    result.WriteString(fmt.Sprintf("Analysis Period: %v\n", duration))
    result.WriteString(fmt.Sprintf("Sample Count: %d\n\n", len(mga.samples)))
    
    result.WriteString("MEMORY GROWTH:\n")
    result.WriteString(fmt.Sprintf("  Heap Alloc: %s → %s (%.2fx)\n",
        formatBytes(first.HeapAlloc), formatBytes(last.HeapAlloc), heapGrowth))
    result.WriteString(fmt.Sprintf("  Heap Sys:   %s → %s (%.2fx)\n",
        formatBytes(first.HeapSys), formatBytes(last.HeapSys), sysGrowth))
    result.WriteString(fmt.Sprintf("  GC Runs:    %d → %d (+%d)\n",
        first.GCCount, last.GCCount, gcIncrease))
    result.WriteString(fmt.Sprintf("  Goroutines: %d → %d (%.2fx)\n\n",
        first.Goroutines, last.Goroutines, goroutineGrowth))
    
    // Calculate trends
    if len(mga.samples) >= 10 {
        trend := mga.calculateTrend()
        result.WriteString(fmt.Sprintf("TREND ANALYSIS:\n"))
        result.WriteString(fmt.Sprintf("  Memory trend: %s\n", trend.description))
        result.WriteString(fmt.Sprintf("  Growth rate: %.2f MB/minute\n", trend.growthRate))
        
        if trend.isLeak {
            result.WriteString("  ⚠️  POTENTIAL MEMORY LEAK DETECTED\n")
        }
    }
    
    return result.String()
}

type MemoryTrend struct {
    description string
    growthRate  float64 // MB per minute
    isLeak      bool
}

func (mga *MemoryGrowthAnalyzer) calculateTrend() MemoryTrend {
    if len(mga.samples) < 10 {
        return MemoryTrend{description: "Insufficient data"}
    }
    
    // Calculate linear regression for memory growth
    n := len(mga.samples)
    var sumX, sumY, sumXY, sumX2 float64
    
    for i, sample := range mga.samples {
        x := float64(i)
        y := float64(sample.HeapAlloc) / (1024 * 1024) // Convert to MB
        
        sumX += x
        sumY += y
        sumXY += x * y
        sumX2 += x * x
    }
    
    // Calculate slope (growth rate per sample)
    slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
    
    // Convert to MB per minute
    avgInterval := mga.samples[n-1].Timestamp.Sub(mga.samples[0].Timestamp).Minutes() / float64(n-1)
    growthRate := slope / avgInterval
    
    trend := MemoryTrend{
        growthRate: growthRate,
    }
    
    switch {
    case growthRate > 10:
        trend.description = "Rapid growth"
        trend.isLeak = true
    case growthRate > 1:
        trend.description = "Moderate growth"
        trend.isLeak = growthRate > 5
    case growthRate > 0.1:
        trend.description = "Slow growth"
    case growthRate > -0.1:
        trend.description = "Stable"
    default:
        trend.description = "Decreasing"
    }
    
    return trend
}
```

## Best Practices for Heap Profiling

### 1. Profile at Steady State

```go
func profileAtSteadyState() {
    // Warm up the application
    for i := 0; i < 1000; i++ {
        performTypicalOperation()
    }
    
    // Force GC to clean up warmup allocations
    runtime.GC()
    runtime.GC()
    
    // Wait for steady state
    time.Sleep(5 * time.Second)
    
    // Now capture the profile
    captureHeapProfile("steady_state.prof")
}
```

### 2. Use Comparative Analysis

```go
func optimizationWorkflow() {
    // Baseline measurement
    captureHeapProfile("before_optimization.prof")
    
    // Apply optimization
    optimizeMemoryUsage()
    
    // Post-optimization measurement
    runtime.GC()
    runtime.GC()
    captureHeapProfile("after_optimization.prof")
    
    // Compare profiles
    compareHeapProfiles("before_optimization.prof", "after_optimization.prof")
}
```

### 3. Monitor Production Continuously

```bash
#!/bin/bash
# production_heap_monitor.sh

INTERVAL=300  # 5 minutes
PROFILE_DIR="/var/log/heap-profiles"
THRESHOLD_MB=1000

while true; do
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    PROFILE_FILE="$PROFILE_DIR/heap_$TIMESTAMP.prof"
    
    # Capture heap profile
    curl -s "http://localhost:6060/debug/pprof/heap" > "$PROFILE_FILE"
    
    # Check heap size
    HEAP_SIZE=$(go tool pprof -top "$PROFILE_FILE" | head -n 1 | grep -o '[0-9.]*MB' | sed 's/MB//')
    
    if (( $(echo "$HEAP_SIZE > $THRESHOLD_MB" | bc -l) )); then
        echo "$(date): Large heap detected: ${HEAP_SIZE}MB"
        # Send alert or take action
    fi
    
    sleep $INTERVAL
done
```

## Next Steps

- Learn [Allocation Profiling](allocation-profiling.md) techniques
- Study [Memory Leak Detection](leak-detection.md) methods
- Explore [Memory Optimization](../../optimization/memory/README.md) strategies

## Summary

Heap profiling is essential for memory optimization:

1. **Regular monitoring** prevents memory issues
2. **Comparative analysis** validates optimizations  
3. **Automated detection** catches leaks early
4. **Production profiling** ensures real-world performance
5. **Trend analysis** predicts future memory needs

Use heap profiling proactively to maintain optimal memory usage and prevent performance degradation in production systems.
