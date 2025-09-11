# Memory Profiling

Memory profiling is essential for identifying memory leaks, excessive allocations, and optimizing memory usage patterns. This chapter covers comprehensive memory profiling techniques and optimization strategies.

## Memory Profiling Fundamentals

### Types of Memory Profiles

Go provides several memory-related profiles:

```
┌─────────────────────────────────────────────────────────────┐
│                   Memory Profile Types                     │
├─────────────────────────────────────────────────────────────┤
│  Profile Type   │  What It Measures    │  Use Case          │
│                 │                      │                    │
│  heap           │  Live heap objects   │  Memory leaks      │
│  allocs         │  All allocations     │  Allocation rate   │
│  goroutine      │  Goroutine stacks    │  Stack memory      │
│  threadcreate   │  OS thread creation  │  Thread overhead   │
└─────────────────────────────────────────────────────────────┘
```

### Memory Profile Collection

```go
package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

// Basic memory profiling
func basicMemoryProfiling() {
    // Allocate memory for profiling
    data := allocateTestData()
    
    // Force garbage collection for accurate snapshot
    runtime.GC()
    runtime.GC() // Call twice to ensure cleanup
    
    // Create memory profile
    f, err := os.Create("mem.prof")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    // Write heap profile
    if err := pprof.WriteHeapProfile(f); err != nil {
        panic(err)
    }
    
    fmt.Printf("Memory profile written to mem.prof\n")
    fmt.Printf("Analyze with: go tool pprof mem.prof\n")
    
    // Keep reference to prevent premature GC
    _ = data
}

func allocateTestData() [][]byte {
    // Allocate various sizes to see allocation patterns
    data := make([][]byte, 1000)
    
    for i := range data {
        size := (i%10 + 1) * 1024 // 1KB to 10KB
        data[i] = make([]byte, size)
        
        // Fill with data to prevent optimization
        for j := range data[i] {
            data[i][j] = byte(i % 256)
        }
    }
    
    return data
}
```

### HTTP Endpoint Memory Profiling

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

func memoryProfilingServer() {
    // Start memory-intensive application
    go memoryIntensiveApplication()
    
    fmt.Println("Memory profiling server running on :6060")
    fmt.Println("Available endpoints:")
    fmt.Println("  /debug/pprof/heap       - Current heap profile")
    fmt.Println("  /debug/pprof/allocs     - All allocations profile")
    fmt.Println("  /debug/pprof/goroutine  - Goroutine profile")
    fmt.Println()
    fmt.Println("Usage examples:")
    fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
    fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/allocs")
    
    log.Fatal(http.ListenAndServe("localhost:6060", nil))
}

func memoryIntensiveApplication() {
    // Simulate various allocation patterns
    cache := make(map[string][]byte)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    counter := 0
    for range ticker.C {
        counter++
        
        // Different allocation patterns
        switch counter % 4 {
        case 0:
            // Small frequent allocations
            for i := 0; i < 100; i++ {
                key := fmt.Sprintf("small_%d_%d", counter, i)
                cache[key] = make([]byte, 64)
            }
            
        case 1:
            // Medium allocations
            for i := 0; i < 10; i++ {
                key := fmt.Sprintf("medium_%d_%d", counter, i)
                cache[key] = make([]byte, 4096)
            }
            
        case 2:
            // Large allocations
            key := fmt.Sprintf("large_%d", counter)
            cache[key] = make([]byte, 1024*1024) // 1MB
            
        case 3:
            // Cleanup some entries to simulate memory churn
            for k := range cache {
                delete(cache, k)
                counter++
                if counter%100 == 0 {
                    break
                }
            }
        }
        
        // Trigger GC occasionally
        if counter%50 == 0 {
            runtime.GC()
        }
        
        // Print memory stats periodically
        if counter%100 == 0 {
            printMemoryStats()
        }
    }
}

func printMemoryStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    fmt.Printf("=== Memory Stats ===\n")
    fmt.Printf("Heap Alloc: %d KB\n", bToKb(m.HeapAlloc))
    fmt.Printf("Heap Sys: %d KB\n", bToKb(m.HeapSys))
    fmt.Printf("Heap Objects: %d\n", m.HeapObjects)
    fmt.Printf("GC Cycles: %d\n", m.NumGC)
    fmt.Printf("Next GC: %d KB\n", bToKb(m.NextGC))
    fmt.Printf("Pause Total: %v\n", time.Duration(m.PauseTotalNs))
    fmt.Println()
}

func bToKb(b uint64) uint64 {
    return b / 1024
}
```

## Memory Profile Analysis

### Command Line Analysis

```bash
# Heap profile analysis
go tool pprof mem.prof
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocation profile analysis  
go tool pprof http://localhost:6060/debug/pprof/allocs

# Interactive pprof commands
(pprof) top10          # Top 10 memory consumers
(pprof) top10 -cum     # Top 10 by cumulative allocation
(pprof) list main.allocateData  # Show function source with allocations
(pprof) web            # Generate web visualization
(pprof) png            # Generate PNG diagram
(pprof) sample_index   # Show available sample types

# Compare memory profiles
go tool pprof -base=baseline_mem.prof current_mem.prof

# Focus analysis
go tool pprof -focus=".*allocate.*" mem.prof
go tool pprof -ignore="runtime.*" mem.prof

# Different sample types
go tool pprof -sample_index=alloc_space mem.prof   # Total allocated
go tool pprof -sample_index=alloc_objects mem.prof # Object count
go tool pprof -sample_index=inuse_space mem.prof   # Currently in use
go tool pprof -sample_index=inuse_objects mem.prof # Objects in use
```

### Programmatic Memory Analysis

```go
package main

import (
    "bytes"
    "fmt"
    "log"
    "runtime"
    "runtime/pprof"
    "sort"
    "time"
    
    "github.com/google/pprof/profile"
)

type MemoryAnalysis struct {
    TotalAllocBytes   int64
    TotalAllocObjects int64
    InUseBytes        int64
    InUseObjects      int64
    Functions         []MemoryFunction
    Timestamp         time.Time
}

type MemoryFunction struct {
    Name              string
    AllocBytes        int64
    AllocObjects      int64
    InUseBytes        int64
    InUseObjects      int64
    AllocBytesPercent float64
    InUseBytesPercent float64
}

func analyzeMemoryProfile() {
    // Collect current memory profile
    var buf bytes.Buffer
    runtime.GC() // Ensure clean snapshot
    
    if err := pprof.WriteHeapProfile(&buf); err != nil {
        log.Fatal("Failed to collect memory profile:", err)
    }
    
    // Parse profile
    p, err := profile.Parse(&buf)
    if err != nil {
        log.Fatal("Failed to parse profile:", err)
    }
    
    // Analyze profile
    analysis := analyzeMemoryProfileData(p)
    printMemoryAnalysis(analysis)
}

func analyzeMemoryProfileData(p *profile.Profile) MemoryAnalysis {
    if len(p.SampleType) < 4 {
        log.Fatal("Insufficient sample types in profile")
    }
    
    // Sample types for heap profile:
    // 0: alloc_objects/count
    // 1: alloc_space/bytes  
    // 2: inuse_objects/count
    // 3: inuse_space/bytes
    
    functionStats := make(map[string]*MemoryFunction)
    var totalAllocBytes, totalAllocObjects int64
    var totalInUseBytes, totalInUseObjects int64
    
    for _, sample := range p.Sample {
        if len(sample.Value) < 4 {
            continue
        }
        
        allocObjects := sample.Value[0]
        allocBytes := sample.Value[1]
        inUseObjects := sample.Value[2]
        inUseBytes := sample.Value[3]
        
        totalAllocObjects += allocObjects
        totalAllocBytes += allocBytes
        totalInUseObjects += inUseObjects
        totalInUseBytes += inUseBytes
        
        // Extract function information from stack
        for _, location := range sample.Location {
            for _, line := range location.Line {
                funcName := line.Function.Name
                
                stats, exists := functionStats[funcName]
                if !exists {
                    stats = &MemoryFunction{Name: funcName}
                    functionStats[funcName] = stats
                }
                
                stats.AllocObjects += allocObjects
                stats.AllocBytes += allocBytes
                stats.InUseObjects += inUseObjects
                stats.InUseBytes += inUseBytes
            }
        }
    }
    
    // Convert to slice and calculate percentages
    functions := make([]MemoryFunction, 0, len(functionStats))
    for _, stats := range functionStats {
        if totalAllocBytes > 0 {
            stats.AllocBytesPercent = float64(stats.AllocBytes) / float64(totalAllocBytes) * 100
        }
        if totalInUseBytes > 0 {
            stats.InUseBytesPercent = float64(stats.InUseBytes) / float64(totalInUseBytes) * 100
        }
        functions = append(functions, *stats)
    }
    
    // Sort by allocated bytes
    sort.Slice(functions, func(i, j int) bool {
        return functions[i].AllocBytes > functions[j].AllocBytes
    })
    
    return MemoryAnalysis{
        TotalAllocBytes:   totalAllocBytes,
        TotalAllocObjects: totalAllocObjects,
        InUseBytes:        totalInUseBytes,
        InUseObjects:      totalInUseObjects,
        Functions:         functions,
        Timestamp:         time.Now(),
    }
}

func printMemoryAnalysis(analysis MemoryAnalysis) {
    fmt.Printf("=== Memory Profile Analysis ===\n")
    fmt.Printf("Timestamp: %v\n", analysis.Timestamp.Format(time.RFC3339))
    fmt.Printf("Total Allocated: %s (%d objects)\n", 
        formatBytes(analysis.TotalAllocBytes), analysis.TotalAllocObjects)
    fmt.Printf("Currently In Use: %s (%d objects)\n", 
        formatBytes(analysis.InUseBytes), analysis.InUseObjects)
    fmt.Printf("Memory Efficiency: %.1f%% (in-use/allocated)\n",
        float64(analysis.InUseBytes)/float64(analysis.TotalAllocBytes)*100)
    fmt.Printf("\n")
    
    fmt.Printf("Top Memory Consumers:\n")
    fmt.Printf("%-50s %12s %12s %8s %8s\n", 
        "Function", "Allocated", "In Use", "Alloc%", "InUse%")
    fmt.Printf("%s\n", strings.Repeat("-", 95))
    
    for i, fn := range analysis.Functions {
        if i >= 20 { // Show top 20
            break
        }
        
        fmt.Printf("%-50s %12s %12s %7.2f%% %7.2f%%\n",
            truncateString(fn.Name, 50),
            formatBytes(fn.AllocBytes),
            formatBytes(fn.InUseBytes),
            fn.AllocBytesPercent,
            fn.InUseBytesPercent)
    }
}

func formatBytes(bytes int64) string {
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

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

### Memory Leak Detection

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

// Memory leak detector
type MemoryLeakDetector struct {
    mu              sync.Mutex
    baselineSet     bool
    baseline        runtime.MemStats
    threshold       uint64  // Bytes
    checkInterval   time.Duration
    alertCallback   func(runtime.MemStats, runtime.MemStats)
    running         bool
    stopCh          chan struct{}
}

func NewMemoryLeakDetector(threshold uint64) *MemoryLeakDetector {
    return &MemoryLeakDetector{
        threshold:     threshold,
        checkInterval: 30 * time.Second,
        stopCh:        make(chan struct{}),
        alertCallback: defaultAlertCallback,
    }
}

func (mld *MemoryLeakDetector) SetAlertCallback(callback func(runtime.MemStats, runtime.MemStats)) {
    mld.mu.Lock()
    defer mld.mu.Unlock()
    mld.alertCallback = callback
}

func (mld *MemoryLeakDetector) Start() {
    mld.mu.Lock()
    if mld.running {
        mld.mu.Unlock()
        return
    }
    mld.running = true
    mld.mu.Unlock()
    
    go mld.monitorLoop()
}

func (mld *MemoryLeakDetector) Stop() {
    mld.mu.Lock()
    if !mld.running {
        mld.mu.Unlock()
        return
    }
    mld.running = false
    mld.mu.Unlock()
    
    close(mld.stopCh)
}

func (mld *MemoryLeakDetector) SetBaseline() {
    mld.mu.Lock()
    defer mld.mu.Unlock()
    
    runtime.GC()
    runtime.GC() // Double GC for clean baseline
    runtime.ReadMemStats(&mld.baseline)
    mld.baselineSet = true
    
    fmt.Printf("Memory leak detector baseline set: %s heap\n", 
        formatBytes(int64(mld.baseline.HeapAlloc)))
}

func (mld *MemoryLeakDetector) monitorLoop() {
    ticker := time.NewTicker(mld.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            mld.checkForLeaks()
        case <-mld.stopCh:
            return
        }
    }
}

func (mld *MemoryLeakDetector) checkForLeaks() {
    mld.mu.Lock()
    if !mld.baselineSet {
        mld.mu.Unlock()
        return
    }
    
    baseline := mld.baseline
    callback := mld.alertCallback
    threshold := mld.threshold
    mld.mu.Unlock()
    
    var current runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&current)
    
    // Check for memory growth
    growth := current.HeapAlloc - baseline.HeapAlloc
    if growth > threshold {
        if callback != nil {
            callback(baseline, current)
        }
    }
}

func defaultAlertCallback(baseline, current runtime.MemStats) {
    growth := current.HeapAlloc - baseline.HeapAlloc
    growthPercent := float64(growth) / float64(baseline.HeapAlloc) * 100
    
    fmt.Printf("=== MEMORY LEAK ALERT ===\n")
    fmt.Printf("Baseline heap: %s\n", formatBytes(int64(baseline.HeapAlloc)))
    fmt.Printf("Current heap: %s\n", formatBytes(int64(current.HeapAlloc)))
    fmt.Printf("Growth: %s (%.1f%%)\n", formatBytes(int64(growth)), growthPercent)
    fmt.Printf("Objects: %d -> %d (+%d)\n", 
        baseline.HeapObjects, current.HeapObjects, 
        current.HeapObjects-baseline.HeapObjects)
    fmt.Printf("GC cycles: %d -> %d (+%d)\n",
        baseline.NumGC, current.NumGC, current.NumGC-baseline.NumGC)
    fmt.Printf("Timestamp: %v\n", time.Now().Format(time.RFC3339))
    fmt.Printf("========================\n")
}

// Example usage with leak simulation
func simulateMemoryLeak() {
    detector := NewMemoryLeakDetector(50 * 1024 * 1024) // 50MB threshold
    detector.Start()
    defer detector.Stop()
    
    // Set baseline after initialization
    time.Sleep(time.Second)
    detector.SetBaseline()
    
    // Simulate memory leak
    leakyData := make([][]byte, 0)
    
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for i := 0; i < 100; i++ {
        select {
        case <-ticker.C:
            // Allocate 1MB every second
            chunk := make([]byte, 1024*1024)
            leakyData = append(leakyData, chunk)
            
            fmt.Printf("Allocated chunk %d (total: %d MB)\n", 
                i+1, (i+1)*1)
        }
    }
    
    // Keep reference to prevent GC
    _ = leakyData
}
```

## Memory Optimization Strategies

### Allocation Reduction

```go
package main

import (
    "sync"
    "strings"
    "time"
)

// Before: Excessive allocations
func inefficientStringProcessing(items []string) []string {
    var results []string
    for _, item := range items {
        // Each operation creates new strings
        processed := strings.ToUpper(item)
        processed = strings.TrimSpace(processed)
        processed = strings.ReplaceAll(processed, "_", "-")
        results = append(results, processed)
    }
    return results
}

// After: Reduced allocations with pre-allocation and pooling
func efficientStringProcessing(items []string) []string {
    // Pre-allocate result slice
    results := make([]string, 0, len(items))
    
    // Use string builder pool
    for _, item := range items {
        builder := stringBuilderPool.Get().(*strings.Builder)
        builder.Reset()
        defer stringBuilderPool.Put(builder)
        
        // Build string efficiently
        for _, r := range item {
            if r == '_' {
                builder.WriteRune('-')
            } else if r >= 'a' && r <= 'z' {
                builder.WriteRune(r - 32) // Convert to uppercase
            } else if r != ' ' && r != '\t' && r != '\n' {
                builder.WriteRune(r)
            }
        }
        
        results = append(results, builder.String())
    }
    
    return results
}

var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

// Object pooling for complex structures
type ExpensiveObject struct {
    Buffer   []byte
    Metadata map[string]interface{}
    Counters []int64
}

func (e *ExpensiveObject) Reset() {
    e.Buffer = e.Buffer[:0]
    for k := range e.Metadata {
        delete(e.Metadata, k)
    }
    for i := range e.Counters {
        e.Counters[i] = 0
    }
}

var expensiveObjectPool = sync.Pool{
    New: func() interface{} {
        return &ExpensiveObject{
            Buffer:   make([]byte, 0, 1024),
            Metadata: make(map[string]interface{}),
            Counters: make([]int64, 10),
        }
    },
}

func processExpensiveObjects(data [][]byte) {
    for _, item := range data {
        obj := expensiveObjectPool.Get().(*ExpensiveObject)
        obj.Reset()
        
        // Use the pooled object
        obj.Buffer = append(obj.Buffer, item...)
        obj.Metadata["size"] = len(item)
        obj.Counters[0]++
        
        // Process object
        processObject(obj)
        
        // Return to pool
        expensiveObjectPool.Put(obj)
    }
}

func processObject(obj *ExpensiveObject) {
    // Simulate processing
    time.Sleep(time.Microsecond)
}
```

### Memory Layout Optimization

```go
// Cache-friendly struct layout
type OptimizedStruct struct {
    // Group pointers together (8 bytes each on 64-bit)
    data     *[]byte
    metadata *Metadata
    parent   *OptimizedStruct
    
    // Group 8-byte values
    timestamp  int64
    counter    uint64
    
    // Group 4-byte values  
    flags      uint32
    version    uint32
    
    // Group smaller values
    status     uint16
    priority   uint8
    active     bool
    // Compiler adds 4 bytes padding here to align to 8-byte boundary
    
    // Total: 64 bytes (fits exactly in one cache line)
}

// Avoid this layout - causes more memory overhead
type SuboptimalStruct struct {
    active    bool      // 1 byte + 7 bytes padding
    timestamp int64     // 8 bytes  
    status    uint16    // 2 bytes + 6 bytes padding
    counter   uint64    // 8 bytes
    priority  uint8     // 1 byte + 7 bytes padding
    version   uint32    // 4 bytes + 4 bytes padding
    // Total: 48 bytes, but with 24 bytes of wasted padding!
}

// Memory-efficient slice operations
func efficientSliceOperations() {
    // Pre-allocate with known capacity
    items := make([]Item, 0, 1000)
    
    // Batch append to reduce reallocations
    batch := make([]Item, 100)
    for i := range batch {
        batch[i] = Item{ID: i}
    }
    items = append(items, batch...)
    
    // Reuse slices by resetting length
    processedItems := items[:0] // Keep capacity, reset length
    
    for _, item := range items {
        if item.ShouldProcess() {
            processedItems = append(processedItems, processItem(item))
        }
    }
    
    // Clear references for GC if slice will live long
    for i := len(processedItems); i < cap(processedItems); i++ {
        items[i] = Item{} // Zero value
    }
}

// Memory-efficient map operations  
func efficientMapOperations() {
    // Pre-size maps when possible
    cache := make(map[string]Data, 1000)
    
    // Use string keys instead of pointer keys for better GC performance
    // Strings are immutable and don't need pointer scanning
    
    // Implement map shrinking for long-lived maps
    if len(cache) > 10000 {
        // Create new map and copy essential entries
        newCache := make(map[string]Data, 5000)
        
        cutoff := time.Now().Add(-time.Hour)
        for k, v := range cache {
            if v.LastAccessed.After(cutoff) {
                newCache[k] = v
            }
        }
        
        cache = newCache
        runtime.GC() // Suggest GC to clean up old map
    }
}
```

### Streaming and Chunking

```go
package main

import (
    "bufio"
    "io"
    "os"
)

// Memory-efficient file processing
func processLargeFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Use buffered reader to avoid loading entire file
    scanner := bufio.NewScanner(file)
    scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB buffer, 1MB max
    
    lineCount := 0
    for scanner.Scan() {
        line := scanner.Text()
        
        // Process line without storing in memory
        if err := processLine(line); err != nil {
            return err
        }
        
        lineCount++
        
        // Periodic progress without accumulating data
        if lineCount%10000 == 0 {
            fmt.Printf("Processed %d lines\n", lineCount)
        }
    }
    
    return scanner.Err()
}

// Stream processing with bounded memory
func streamProcessor(input <-chan []byte, output chan<- ProcessedData) {
    buffer := make([]byte, 0, 64*1024) // 64KB working buffer
    
    for data := range input {
        // Reset buffer, keep capacity
        buffer = buffer[:0]
        
        // Process in chunks to limit memory usage
        for len(data) > 0 {
            chunkSize := min(len(data), cap(buffer))
            chunk := data[:chunkSize]
            data = data[chunkSize:]
            
            // Process chunk
            result := processChunk(chunk)
            
            select {
            case output <- result:
            case <-time.After(time.Second):
                // Handle backpressure
                fmt.Println("Output channel blocked, dropping data")
            }
        }
    }
}

// Batch processing with memory limits
func batchProcessor(items <-chan Item) {
    const maxBatchSize = 1000
    const maxMemoryUsage = 100 * 1024 * 1024 // 100MB
    
    batch := make([]Item, 0, maxBatchSize)
    var batchMemory int64
    
    for item := range items {
        batch = append(batch, item)
        batchMemory += item.MemorySize()
        
        // Process batch when size or memory limit reached
        if len(batch) >= maxBatchSize || batchMemory >= maxMemoryUsage {
            processBatch(batch)
            
            // Reset batch, keep capacity
            batch = batch[:0]
            batchMemory = 0
        }
    }
    
    // Process remaining items
    if len(batch) > 0 {
        processBatch(batch)
    }
}
```

## Advanced Memory Profiling

### Custom Memory Tracking

```go
package main

import (
    "runtime"
    "sync"
    "sync/atomic"
    "time"
)

// Custom memory tracker for specific allocations
type MemoryTracker struct {
    mu           sync.RWMutex
    allocations  map[string]*AllocationStats
    totalBytes   int64
    totalObjects int64
}

type AllocationStats struct {
    TotalBytes   int64
    TotalObjects int64
    ActiveBytes  int64
    ActiveObjects int64
    PeakBytes    int64
    PeakObjects  int64
}

func NewMemoryTracker() *MemoryTracker {
    return &MemoryTracker{
        allocations: make(map[string]*AllocationStats),
    }
}

func (mt *MemoryTracker) TrackAllocation(category string, size int64) {
    atomic.AddInt64(&mt.totalBytes, size)
    atomic.AddInt64(&mt.totalObjects, 1)
    
    mt.mu.Lock()
    defer mt.mu.Unlock()
    
    stats, exists := mt.allocations[category]
    if !exists {
        stats = &AllocationStats{}
        mt.allocations[category] = stats
    }
    
    stats.TotalBytes += size
    stats.TotalObjects++
    stats.ActiveBytes += size
    stats.ActiveObjects++
    
    if stats.ActiveBytes > stats.PeakBytes {
        stats.PeakBytes = stats.ActiveBytes
    }
    if stats.ActiveObjects > stats.PeakObjects {
        stats.PeakObjects = stats.ActiveObjects
    }
}

func (mt *MemoryTracker) TrackDeallocation(category string, size int64) {
    atomic.AddInt64(&mt.totalBytes, -size)
    atomic.AddInt64(&mt.totalObjects, -1)
    
    mt.mu.Lock()
    defer mt.mu.Unlock()
    
    if stats, exists := mt.allocations[category]; exists {
        stats.ActiveBytes -= size
        stats.ActiveObjects--
    }
}

func (mt *MemoryTracker) GetStats() map[string]AllocationStats {
    mt.mu.RLock()
    defer mt.mu.RUnlock()
    
    result := make(map[string]AllocationStats)
    for category, stats := range mt.allocations {
        result[category] = *stats
    }
    return result
}

func (mt *MemoryTracker) PrintReport() {
    fmt.Printf("=== Memory Tracker Report ===\n")
    fmt.Printf("Total: %s allocated, %d objects\n", 
        formatBytes(atomic.LoadInt64(&mt.totalBytes)),
        atomic.LoadInt64(&mt.totalObjects))
    fmt.Printf("\n")
    
    fmt.Printf("%-20s %12s %12s %12s %10s %10s\n",
        "Category", "Total", "Active", "Peak", "Objects", "PeakObjs")
    fmt.Printf("%s\n", strings.Repeat("-", 85))
    
    stats := mt.GetStats()
    for category, stat := range stats {
        fmt.Printf("%-20s %12s %12s %12s %10d %10d\n",
            category,
            formatBytes(stat.TotalBytes),
            formatBytes(stat.ActiveBytes),
            formatBytes(stat.PeakBytes),
            stat.ActiveObjects,
            stat.PeakObjects)
    }
}

// Usage example with tracked allocations
var globalTracker = NewMemoryTracker()

type TrackedBuffer struct {
    data     []byte
    category string
}

func NewTrackedBuffer(size int, category string) *TrackedBuffer {
    data := make([]byte, size)
    globalTracker.TrackAllocation(category, int64(size))
    
    return &TrackedBuffer{
        data:     data,
        category: category,
    }
}

func (tb *TrackedBuffer) Free() {
    if tb.data != nil {
        globalTracker.TrackDeallocation(tb.category, int64(len(tb.data)))
        tb.data = nil
    }
}

func (tb *TrackedBuffer) Resize(newSize int) {
    if tb.data == nil {
        return
    }
    
    oldSize := len(tb.data)
    tb.data = make([]byte, newSize)
    
    sizeDiff := int64(newSize - oldSize)
    if sizeDiff > 0 {
        globalTracker.TrackAllocation(tb.category, sizeDiff)
    } else {
        globalTracker.TrackDeallocation(tb.category, -sizeDiff)
    }
}
```

### Memory Pressure Monitoring

```go
// Monitor system memory pressure
type MemoryPressureMonitor struct {
    thresholds    map[string]float64
    callbacks     map[string]func(float64)
    checkInterval time.Duration
    running       bool
    stopCh        chan struct{}
}

func NewMemoryPressureMonitor() *MemoryPressureMonitor {
    return &MemoryPressureMonitor{
        thresholds: map[string]float64{
            "warning":  0.80, // 80% memory usage
            "critical": 0.95, // 95% memory usage
        },
        callbacks:     make(map[string]func(float64)),
        checkInterval: 10 * time.Second,
        stopCh:        make(chan struct{}),
    }
}

func (mpm *MemoryPressureMonitor) SetThreshold(level string, threshold float64, callback func(float64)) {
    mpm.thresholds[level] = threshold
    mpm.callbacks[level] = callback
}

func (mpm *MemoryPressureMonitor) Start() {
    if mpm.running {
        return
    }
    mpm.running = true
    
    go mpm.monitorLoop()
}

func (mpm *MemoryPressureMonitor) Stop() {
    if !mpm.running {
        return
    }
    mpm.running = false
    close(mpm.stopCh)
}

func (mpm *MemoryPressureMonitor) monitorLoop() {
    ticker := time.NewTicker(mpm.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            mpm.checkMemoryPressure()
        case <-mpm.stopCh:
            return
        }
    }
}

func (mpm *MemoryPressureMonitor) checkMemoryPressure() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Calculate memory pressure ratio
    pressure := float64(m.HeapAlloc) / float64(m.HeapSys)
    
    // Check thresholds
    for level, threshold := range mpm.thresholds {
        if pressure >= threshold {
            if callback, exists := mpm.callbacks[level]; exists {
                callback(pressure)
            }
        }
    }
}

// Example usage with automatic optimization
func setupMemoryPressureHandling() {
    monitor := NewMemoryPressureMonitor()
    
    // Warning level: start gentle cleanup
    monitor.SetThreshold("warning", 0.80, func(pressure float64) {
        fmt.Printf("Memory pressure warning: %.1f%%\n", pressure*100)
        
        // Trigger garbage collection
        runtime.GC()
        
        // Clear caches
        clearOptionalCaches()
    })
    
    // Critical level: aggressive cleanup
    monitor.SetThreshold("critical", 0.95, func(pressure float64) {
        fmt.Printf("CRITICAL memory pressure: %.1f%%\n", pressure*100)
        
        // Force GC
        runtime.GC()
        runtime.GC()
        
        // Clear all caches
        clearAllCaches()
        
        // Reduce buffer sizes
        shrinkBuffers()
        
        // Alert operations
        alertOperations("Critical memory pressure detected")
    })
    
    monitor.Start()
}
```

Memory profiling is essential for building scalable, memory-efficient Go applications. Master these techniques to identify and resolve memory-related performance issues.

---

**Next**: [Goroutine Profiling](goroutine-profiling.md) - Analyze concurrent execution and goroutine behavior
