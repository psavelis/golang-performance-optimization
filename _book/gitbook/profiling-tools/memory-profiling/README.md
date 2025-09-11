# Memory Profiling

Master comprehensive memory profiling techniques in Go to identify memory leaks, optimize allocations, and improve garbage collection performance.

## Memory Profiling Overview

Memory profiling in Go provides insights into:
- **Heap allocation patterns** - Where and how memory is allocated
- **Memory leaks** - Objects that should be garbage collected but aren't
- **Allocation frequency** - Functions causing frequent allocations
- **Memory usage trends** - How memory usage changes over time

## Heap Profiling

### Basic Heap Profiling

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

func main() {
    // Enable pprof endpoint
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Memory-intensive application simulation
    go simulateMemoryUsage()
    
    // Take heap snapshot programmatically
    takeHeapSnapshot("before")
    
    // Let application run
    time.Sleep(30 * time.Second)
    
    takeHeapSnapshot("after")
    
    // Keep server running
    select {}
}

func takeHeapSnapshot(name string) {
    runtime.GC() // Force garbage collection before snapshot
    
    f, err := os.Create(fmt.Sprintf("heap_%s.prof", name))
    if err != nil {
        log.Printf("Could not create heap profile: %v", err)
        return
    }
    defer f.Close()
    
    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Printf("Could not write heap profile: %v", err)
    }
    
    log.Printf("Heap profile saved: heap_%s.prof", name)
}

func simulateMemoryUsage() {
    // Simulate different allocation patterns
    var data [][]byte
    
    for i := 0; ; i++ {
        // Allocate different sizes
        size := 1024 * (i%100 + 1)
        chunk := make([]byte, size)
        
        // Fill with data to prevent optimization
        for j := range chunk {
            chunk[j] = byte(j % 256)
        }
        
        data = append(data, chunk)
        
        // Occasionally clean up some data
        if i%50 == 0 && len(data) > 25 {
            data = data[25:] // Remove old allocations
        }
        
        time.Sleep(100 * time.Millisecond)
    }
}
```

### Advanced Heap Analysis

```go
// Memory-intensive data structures for profiling
type MemoryAnalyzer struct {
    cache       map[string]*CacheEntry
    bufferPool  sync.Pool
    activeConns map[int]*Connection
    metrics     *MemoryMetrics
    mu          sync.RWMutex
}

type CacheEntry struct {
    Key        string
    Value      []byte
    Timestamp  time.Time
    AccessCount int64
}

type Connection struct {
    ID       int
    Buffer   []byte
    Metadata map[string]interface{}
    Created  time.Time
}

type MemoryMetrics struct {
    AllocationsPerSecond int64
    TotalAllocatedBytes  int64
    ActiveObjects        int64
    GCCount             int64
}

func NewMemoryAnalyzer() *MemoryAnalyzer {
    ma := &MemoryAnalyzer{
        cache:       make(map[string]*CacheEntry),
        activeConns: make(map[int]*Connection),
        metrics:     &MemoryMetrics{},
        bufferPool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 4096)
            },
        },
    }
    
    // Start memory monitoring
    go ma.monitorMemory()
    
    return ma
}

func (ma *MemoryAnalyzer) AddCacheEntry(key string, data []byte) {
    ma.mu.Lock()
    defer ma.mu.Unlock()
    
    entry := &CacheEntry{
        Key:       key,
        Value:     make([]byte, len(data)),
        Timestamp: time.Now(),
    }
    copy(entry.Value, data)
    
    ma.cache[key] = entry
    atomic.AddInt64(&ma.metrics.AllocationsPerSecond, 1)
}

func (ma *MemoryAnalyzer) GetCacheEntry(key string) *CacheEntry {
    ma.mu.RLock()
    entry, exists := ma.cache[key]
    ma.mu.RUnlock()
    
    if exists {
        atomic.AddInt64(&entry.AccessCount, 1)
    }
    
    return entry
}

func (ma *MemoryAnalyzer) AddConnection(id int) {
    ma.mu.Lock()
    defer ma.mu.Unlock()
    
    conn := &Connection{
        ID:       id,
        Buffer:   make([]byte, 8192),
        Metadata: make(map[string]interface{}),
        Created:  time.Now(),
    }
    
    // Add some metadata
    conn.Metadata["user_id"] = fmt.Sprintf("user_%d", id)
    conn.Metadata["session"] = fmt.Sprintf("session_%d_%d", id, time.Now().Unix())
    
    ma.activeConns[id] = conn
    atomic.AddInt64(&ma.metrics.ActiveObjects, 1)
}

func (ma *MemoryAnalyzer) RemoveConnection(id int) {
    ma.mu.Lock()
    defer ma.mu.Unlock()
    
    if _, exists := ma.activeConns[id]; exists {
        delete(ma.activeConns, id)
        atomic.AddInt64(&ma.metrics.ActiveObjects, -1)
    }
}

func (ma *MemoryAnalyzer) ProcessData(data []byte) []byte {
    // Get buffer from pool
    buffer := ma.bufferPool.Get().([]byte)
    defer ma.bufferPool.Put(buffer)
    
    // Process data (simulate work)
    processed := make([]byte, len(data)*2)
    for i, b := range data {
        processed[i*2] = b
        processed[i*2+1] = b ^ 0xFF
    }
    
    return processed
}

func (ma *MemoryAnalyzer) monitorMemory() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        log.Printf("Memory Stats:")
        log.Printf("  Alloc: %d KB", m.Alloc/1024)
        log.Printf("  TotalAlloc: %d KB", m.TotalAlloc/1024)
        log.Printf("  Sys: %d KB", m.Sys/1024)
        log.Printf("  NumGC: %d", m.NumGC)
        log.Printf("  HeapObjects: %d", m.HeapObjects)
        log.Printf("  Cache entries: %d", len(ma.cache))
        log.Printf("  Active connections: %d", len(ma.activeConns))
    }
}

// Simulate memory leak scenario
func (ma *MemoryAnalyzer) SimulateMemoryLeak() {
    // This creates a memory leak by keeping references
    leakedData := make(map[int][]byte)
    
    for i := 0; ; i++ {
        // Allocate but never clean up
        data := make([]byte, 1024*1024) // 1MB allocation
        leakedData[i] = data
        
        time.Sleep(time.Second)
        
        // Simulate some cleanup (but not enough)
        if i%100 == 0 && len(leakedData) > 50 {
            // Only remove a few entries
            for j := 0; j < 5; j++ {
                delete(leakedData, i-50+j)
            }
        }
    }
}
```

### Memory Allocation Profiling

```go
// Allocation profiling for specific functions
func BenchmarkMemoryAllocations(b *testing.B) {
    b.Run("SliceAppend", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var slice []int
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
    
    b.Run("SlicePrealloc", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            slice := make([]int, 0, 1000)
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
    
    b.Run("StringConcat", func(b *testing.B) {
        b.ReportAllocs()
        parts := []string{"hello", "world", "golang"}
        for i := 0; i < b.N; i++ {
            result := ""
            for _, part := range parts {
                result += part
            }
            _ = result
        }
    })
    
    b.Run("StringBuilder", func(b *testing.B) {
        b.ReportAllocs()
        parts := []string{"hello", "world", "golang"}
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for _, part := range parts {
                builder.WriteString(part)
            }
            _ = builder.String()
        }
    })
}

// Custom allocation tracking
type AllocationTracker struct {
    allocations map[string]*AllocationInfo
    mu          sync.RWMutex
}

type AllocationInfo struct {
    Count     int64
    TotalSize int64
    LastSeen  time.Time
}

var globalTracker = &AllocationTracker{
    allocations: make(map[string]*AllocationInfo),
}

func TrackAllocation(category string, size int64) {
    globalTracker.mu.Lock()
    defer globalTracker.mu.Unlock()
    
    info, exists := globalTracker.allocations[category]
    if !exists {
        info = &AllocationInfo{}
        globalTracker.allocations[category] = info
    }
    
    info.Count++
    info.TotalSize += size
    info.LastSeen = time.Now()
}

func GetAllocationStats() map[string]*AllocationInfo {
    globalTracker.mu.RLock()
    defer globalTracker.mu.RUnlock()
    
    result := make(map[string]*AllocationInfo)
    for k, v := range globalTracker.allocations {
        result[k] = &AllocationInfo{
            Count:     v.Count,
            TotalSize: v.TotalSize,
            LastSeen:  v.LastSeen,
        }
    }
    
    return result
}

// Instrumented allocation functions
func TrackedMakeSlice(category string, size int) []byte {
    data := make([]byte, size)
    TrackAllocation(category, int64(size))
    return data
}

func TrackedMakeMap(category string, size int) map[string]interface{} {
    m := make(map[string]interface{}, size)
    TrackAllocation(category, int64(size*32)) // Estimate map overhead
    return m
}
```

## Memory Leak Detection

### Leak Detection Patterns

```go
// Common memory leak scenarios and detection
type LeakDetector struct {
    snapshots []MemorySnapshot
    interval  time.Duration
    threshold float64 // Growth rate threshold
}

type MemorySnapshot struct {
    Timestamp   time.Time
    HeapAlloc   uint64
    HeapSys     uint64
    HeapObjects uint64
    NumGC       uint32
}

func NewLeakDetector(interval time.Duration, threshold float64) *LeakDetector {
    ld := &LeakDetector{
        interval:  interval,
        threshold: threshold,
    }
    
    go ld.monitor()
    return ld
}

func (ld *LeakDetector) monitor() {
    ticker := time.NewTicker(ld.interval)
    defer ticker.Stop()
    
    for range ticker.C {
        ld.takeSnapshot()
        ld.analyzeGrowth()
    }
}

func (ld *LeakDetector) takeSnapshot() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    snapshot := MemorySnapshot{
        Timestamp:   time.Now(),
        HeapAlloc:   m.Alloc,
        HeapSys:     m.Sys,
        HeapObjects: m.HeapObjects,
        NumGC:       m.NumGC,
    }
    
    ld.snapshots = append(ld.snapshots, snapshot)
    
    // Keep only last 100 snapshots
    if len(ld.snapshots) > 100 {
        ld.snapshots = ld.snapshots[1:]
    }
}

func (ld *LeakDetector) analyzeGrowth() {
    if len(ld.snapshots) < 10 {
        return
    }
    
    recent := ld.snapshots[len(ld.snapshots)-10:]
    growthRate := ld.calculateGrowthRate(recent)
    
    if growthRate > ld.threshold {
        log.Printf("MEMORY LEAK DETECTED: Growth rate %.2f%% exceeds threshold %.2f%%",
            growthRate*100, ld.threshold*100)
        
        ld.generateLeakReport()
    }
}

func (ld *LeakDetector) calculateGrowthRate(snapshots []MemorySnapshot) float64 {
    if len(snapshots) < 2 {
        return 0
    }
    
    first := snapshots[0]
    last := snapshots[len(snapshots)-1]
    
    if first.HeapAlloc == 0 {
        return 0
    }
    
    return float64(last.HeapAlloc-first.HeapAlloc) / float64(first.HeapAlloc)
}

func (ld *LeakDetector) generateLeakReport() {
    log.Println("=== MEMORY LEAK REPORT ===")
    
    if len(ld.snapshots) > 0 {
        latest := ld.snapshots[len(ld.snapshots)-1]
        log.Printf("Current heap allocation: %d bytes", latest.HeapAlloc)
        log.Printf("Current heap objects: %d", latest.HeapObjects)
        log.Printf("Total GC runs: %d", latest.NumGC)
    }
    
    // Take heap dump
    takeHeapDump()
    
    // Print allocation stats
    stats := GetAllocationStats()
    log.Println("Allocation breakdown:")
    for category, info := range stats {
        log.Printf("  %s: %d allocations, %d bytes total", 
            category, info.Count, info.TotalSize)
    }
}

func takeHeapDump() {
    runtime.GC()
    
    filename := fmt.Sprintf("leak_dump_%d.prof", time.Now().Unix())
    f, err := os.Create(filename)
    if err != nil {
        log.Printf("Could not create heap dump: %v", err)
        return
    }
    defer f.Close()
    
    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Printf("Could not write heap dump: %v", err)
    } else {
        log.Printf("Heap dump saved: %s", filename)
    }
}
```

### Goroutine Leak Detection

```go
// Goroutine leak detection and monitoring
type GoroutineMonitor struct {
    baseline    int
    threshold   int
    checkInterval time.Duration
}

func NewGoroutineMonitor(threshold int, interval time.Duration) *GoroutineMonitor {
    gm := &GoroutineMonitor{
        baseline:     runtime.NumGoroutine(),
        threshold:    threshold,
        checkInterval: interval,
    }
    
    go gm.monitor()
    return gm
}

func (gm *GoroutineMonitor) monitor() {
    ticker := time.NewTicker(gm.checkInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        current := runtime.NumGoroutine()
        growth := current - gm.baseline
        
        if growth > gm.threshold {
            log.Printf("GOROUTINE LEAK DETECTED: %d goroutines (baseline: %d, growth: %d)",
                current, gm.baseline, growth)
            
            gm.dumpGoroutines()
        }
    }
}

func (gm *GoroutineMonitor) dumpGoroutines() {
    filename := fmt.Sprintf("goroutines_%d.prof", time.Now().Unix())
    f, err := os.Create(filename)
    if err != nil {
        log.Printf("Could not create goroutine dump: %v", err)
        return
    }
    defer f.Close()
    
    if profile := pprof.Lookup("goroutine"); profile != nil {
        profile.WriteTo(f, 2) // 2 = debug level with stack traces
        log.Printf("Goroutine dump saved: %s", filename)
    }
}

// Simulate goroutine leak
func SimulateGoroutineLeak() {
    for i := 0; ; i++ {
        go func(id int) {
            // Goroutine that never exits
            ch := make(chan bool)
            <-ch // Blocks forever
        }(i)
        
        time.Sleep(100 * time.Millisecond)
    }
}
```

## Memory Usage Optimization

### Object Pool Implementation

```go
// Advanced object pooling for memory optimization
type ObjectPool struct {
    pool     sync.Pool
    maxSize  int
    created  int64
    borrowed int64
    returned int64
}

func NewObjectPool(factory func() interface{}, maxSize int) *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{New: factory},
        maxSize: maxSize,
    }
}

func (op *ObjectPool) Get() interface{} {
    obj := op.pool.Get()
    atomic.AddInt64(&op.borrowed, 1)
    return obj
}

func (op *ObjectPool) Put(obj interface{}) {
    if atomic.LoadInt64(&op.created) < int64(op.maxSize) {
        op.pool.Put(obj)
        atomic.AddInt64(&op.returned, 1)
    }
}

func (op *ObjectPool) Stats() (created, borrowed, returned int64) {
    return atomic.LoadInt64(&op.created),
           atomic.LoadInt64(&op.borrowed),
           atomic.LoadInt64(&op.returned)
}

// Usage example: Buffer pool
var bufferPool = NewObjectPool(func() interface{} {
    return make([]byte, 4096)
}, 1000)

func ProcessWithPool(data []byte) []byte {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // Reset buffer
    buffer = buffer[:0]
    
    // Process data
    buffer = append(buffer, data...)
    buffer = append(buffer, []byte("_processed")...)
    
    // Return copy since buffer goes back to pool
    result := make([]byte, len(buffer))
    copy(result, buffer)
    return result
}
```

### Memory-Efficient Data Structures

```go
// Memory-optimized data structures
type CompactSlice struct {
    data     []uint32
    bitWidth int
    mask     uint32
    length   int
}

func NewCompactSlice(maxValue uint32, capacity int) *CompactSlice {
    // Calculate minimum bits needed
    bitWidth := 32 - bits.LeadingZeros32(maxValue)
    if bitWidth == 0 {
        bitWidth = 1
    }
    
    valuesPerUint32 := 32 / bitWidth
    arraySize := (capacity + valuesPerUint32 - 1) / valuesPerUint32
    
    return &CompactSlice{
        data:     make([]uint32, arraySize),
        bitWidth: bitWidth,
        mask:     (1 << bitWidth) - 1,
    }
}

func (cs *CompactSlice) Set(index int, value uint32) {
    if index >= cs.length {
        cs.length = index + 1
    }
    
    valuesPerUint32 := 32 / cs.bitWidth
    arrayIndex := index / valuesPerUint32
    bitOffset := (index % valuesPerUint32) * cs.bitWidth
    
    // Clear existing value
    cs.data[arrayIndex] &^= cs.mask << bitOffset
    
    // Set new value
    cs.data[arrayIndex] |= (value & cs.mask) << bitOffset
}

func (cs *CompactSlice) Get(index int) uint32 {
    if index >= cs.length {
        return 0
    }
    
    valuesPerUint32 := 32 / cs.bitWidth
    arrayIndex := index / valuesPerUint32
    bitOffset := (index % valuesPerUint32) * cs.bitWidth
    
    return (cs.data[arrayIndex] >> bitOffset) & cs.mask
}

func (cs *CompactSlice) MemoryUsage() int {
    return len(cs.data) * 4 // 4 bytes per uint32
}

// Regular slice for comparison
func (cs *CompactSlice) RegularSliceMemoryUsage() int {
    return cs.length * 4 // 4 bytes per uint32
}
```

## Memory Profiling Best Practices

### 1. Profile Realistic Workloads
```go
// Generate realistic memory load for profiling
func GenerateRealisticLoad() {
    // Simulate web server with caching
    cache := make(map[string][]byte)
    
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        value := make([]byte, rand.Intn(1024)+512) // 512-1536 bytes
        
        // Fill with realistic data
        for j := range value {
            value[j] = byte(rand.Intn(256))
        }
        
        cache[key] = value
        
        // Simulate cache cleanup
        if i%1000 == 0 && len(cache) > 5000 {
            // Remove random entries
            keysToRemove := make([]string, 0, 100)
            count := 0
            for k := range cache {
                keysToRemove = append(keysToRemove, k)
                count++
                if count >= 100 {
                    break
                }
            }
            
            for _, k := range keysToRemove {
                delete(cache, k)
            }
        }
    }
}
```

### 2. Continuous Monitoring
```go
// Production memory monitoring
func StartProductionMemoryMonitoring() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            // Log key metrics
            log.Printf("Memory: Alloc=%dKB Sys=%dKB NumGC=%d HeapObjects=%d",
                m.Alloc/1024, m.Sys/1024, m.NumGC, m.HeapObjects)
            
            // Alert on memory growth
            if m.Alloc > 500*1024*1024 { // 500MB
                log.Printf("HIGH MEMORY USAGE ALERT: %d MB allocated", m.Alloc/(1024*1024))
            }
        }
    }()
}
```

### 3. Automated Leak Detection
```go
// Automated leak detection in tests
func TestMemoryLeaks(t *testing.T) {
    // Get baseline
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    // Run test code
    for i := 0; i < 1000; i++ {
        // Code that might leak
        data := make([]byte, 1024)
        _ = data
    }
    
    // Force GC and check for leaks
    runtime.GC()
    runtime.GC() // Run twice to ensure cleanup
    
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    
    growth := m2.Alloc - m1.Alloc
    if growth > 100*1024 { // 100KB threshold
        t.Errorf("Potential memory leak: %d bytes not freed", growth)
    }
}
```

Memory profiling is essential for building efficient Go applications and preventing memory-related performance issues in production environments.
