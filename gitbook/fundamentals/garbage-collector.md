# Garbage Collector

Go's garbage collector is a crucial component for automatic memory management. Understanding its behavior and optimization techniques is essential for high-performance applications.

## GC Overview

### Tri-Color Marking Algorithm

Go uses a **concurrent, tri-color, mark-and-sweep** garbage collector:

```
┌─────────────────────────────────────────────────────────────┐
│                   Tri-Color Marking                        │
├─────────────────────────────────────────────────────────────┤
│  White Objects  │  Gray Objects   │  Black Objects         │
│                 │                 │                        │
│  • Unreachable  │  • Reachable    │  • Reachable           │
│  • To be freed  │  • Not scanned  │  • Fully scanned       │
│  • Initial state│  • Work queue   │  • Safe to use         │
└─────────────────────────────────────────────────────────────┘
```

#### Marking Phases

1. **Mark Setup**: All objects start white, roots become gray
2. **Mark**: Gray objects are scanned, children become gray, object becomes black
3. **Mark Termination**: No gray objects remain
4. **Sweep**: White objects are freed

### GC Trigger Conditions

```go
// GC is triggered when:
// 1. Heap size doubles since last GC
// 2. 2 minutes have passed since last GC
// 3. runtime.GC() is called explicitly

func gcTriggerConditions() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    fmt.Printf("Next GC target: %d bytes\n", m.NextGC)
    fmt.Printf("Current heap: %d bytes\n", m.HeapAlloc)
    fmt.Printf("GC percentage: %d%%\n", debug.SetGCPercent(-1))
    
    // Default: GC triggers when heap grows 100% (doubles)
    // Lower values = more frequent GC, less memory usage
    // Higher values = less frequent GC, more memory usage
}
```

## GC Phases and Performance

### GC Cycle Breakdown

```go
import (
    "runtime"
    "runtime/debug"
    "time"
)

// Monitor GC performance
func monitorGCPerformance() {
    var m1, m2 runtime.MemStats
    
    // Before allocation
    runtime.ReadMemStats(&m1)
    runtime.GC()
    
    // Allocate memory
    data := make([][]byte, 1000)
    for i := range data {
        data[i] = make([]byte, 1024)
    }
    
    // After allocation
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Heap before: %d bytes\n", m1.HeapAlloc)
    fmt.Printf("Heap after: %d bytes\n", m2.HeapAlloc)
    fmt.Printf("GC cycles: %d\n", m2.NumGC-m1.NumGC)
    fmt.Printf("Total pause: %v\n", time.Duration(m2.PauseTotalNs-m1.PauseTotalNs))
    
    // Keep reference to prevent premature collection
    _ = data
}

// Detailed GC statistics
func detailedGCStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    fmt.Printf("=== Heap Statistics ===\n")
    fmt.Printf("Heap Alloc: %d bytes\n", m.HeapAlloc)
    fmt.Printf("Heap Sys: %d bytes\n", m.HeapSys)
    fmt.Printf("Heap Objects: %d\n", m.HeapObjects)
    fmt.Printf("Next GC: %d bytes\n", m.NextGC)
    
    fmt.Printf("=== GC Statistics ===\n")
    fmt.Printf("GC Cycles: %d\n", m.NumGC)
    fmt.Printf("Forced GC: %d\n", m.NumForcedGC)
    fmt.Printf("Total Pause: %v\n", time.Duration(m.PauseTotalNs))
    fmt.Printf("Last Pause: %v\n", time.Duration(m.PauseNs[(m.NumGC+255)%256]))
    
    // Recent pause times
    fmt.Printf("Recent Pauses: ")
    for i := 0; i < 5 && i < int(m.NumGC); i++ {
        idx := (m.NumGC - uint32(i) + 255) % 256
        fmt.Printf("%v ", time.Duration(m.PauseNs[idx]))
    }
    fmt.Println()
}
```

### Write Barriers

Write barriers ensure GC correctness during concurrent marking:

```go
// Write barriers are automatically inserted by the compiler
func writeBarrierExample() {
    type Node struct {
        data string
        next *Node
    }
    
    head := &Node{data: "first"}
    
    // This assignment triggers a write barrier
    head.next = &Node{data: "second"}
    
    // The write barrier ensures the new object is marked
    // if GC is running concurrently
}

// Write barriers have performance implications
func minimizeWriteBarriers() {
    // 1. Bulk initialization reduces write barriers
    nodes := make([]*Node, 1000)
    for i := range nodes {
        nodes[i] = &Node{ID: i}  // One write barrier per assignment
    }
    
    // 2. Batch pointer updates
    type Container struct {
        items []*Item
    }
    
    container := &Container{
        items: make([]*Item, 0, 1000),  // Pre-allocate
    }
    
    // Batch updates to reduce write barrier overhead
    newItems := make([]*Item, 100)
    for i := range newItems {
        newItems[i] = &Item{ID: i}
    }
    
    // Single append reduces write barriers compared to individual appends
    container.items = append(container.items, newItems...)
}
```

## GC Optimization Techniques

### Reducing Allocation Pressure

```go
// 1. Object pooling to reduce GC pressure
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func pooledProcessing(data []byte) []byte {
    buffer := bufferPool.Get().([]byte)
    buffer = buffer[:0]  // Reset length
    defer bufferPool.Put(buffer)
    
    // Use pooled buffer for processing
    buffer = append(buffer, processData(data)...)
    
    // Return copy, not pooled buffer
    result := make([]byte, len(buffer))
    copy(result, buffer)
    return result
}

// 2. Reuse composite literals
type Config struct {
    Settings map[string]interface{}
    Flags    []bool
}

var configPool = sync.Pool{
    New: func() interface{} {
        return &Config{
            Settings: make(map[string]interface{}),
            Flags:    make([]bool, 10),
        }
    },
}

func (c *Config) Reset() {
    // Clear map without reallocating
    for k := range c.Settings {
        delete(c.Settings, k)
    }
    
    // Reset slice
    for i := range c.Flags {
        c.Flags[i] = false
    }
}

// 3. Stack allocation preference
func preferStackAllocation() {
    // Stack allocated (preferred)
    var buffer [1024]byte
    processBuffer(buffer[:])
    
    // Avoid heap allocation when possible
    result := processData()  // Return value, not pointer
    _ = result
}
```

### Memory Layout Optimization

```go
// Optimize for GC scanning efficiency
type GCOptimizedStruct struct {
    // Group pointer fields together
    name     *string    // Pointer
    metadata *Metadata  // Pointer
    parent   *Node      // Pointer
    
    // Group non-pointer fields together
    id       uint64     // Value
    flags    uint32     // Value
    count    int32      // Value
    active   bool       // Value
}

// Avoid mixed pointer/non-pointer layouts
type SuboptimalStruct struct {
    id       uint64     // Value
    name     *string    // Pointer - causes more GC scanning
    count    int32      // Value
    metadata *Metadata  // Pointer - scattered pointers
    active   bool       // Value
}

// Use value semantics when appropriate
type ValueOptimized struct {
    // Embed small structs as values
    Config   ConfigData  // Value - no pointer indirection
    Metadata MetaData    // Value - no GC scanning overhead
    
    // Only use pointers when necessary
    LargeData *[]byte    // Pointer to large data
}
```

### GC-Friendly Data Structures

```go
// 1. Slice optimization
func sliceOptimization() {
    // Pre-allocate to avoid growth
    items := make([]Item, 0, 1000)
    
    // Process in batches to limit live objects
    const batchSize = 100
    for len(items) < 1000 {
        batch := items[len(items):len(items)+batchSize]
        processBatch(batch)
        
        // Clear processed items to help GC
        for i := range batch {
            batch[i] = Item{}  // Zero value
        }
    }
}

// 2. Map optimization
func mapOptimization() {
    // Pre-size maps to avoid rehashing
    cache := make(map[string]*Data, 1000)
    
    // Implement cache eviction to limit size
    if len(cache) > 10000 {
        // Remove oldest entries
        for k := range cache {
            delete(cache, k)
            if len(cache) <= 5000 {
                break
            }
        }
    }
    
    // Use string keys instead of pointer keys when possible
    // Strings don't need pointer scanning during GC
}

// 3. Channel optimization
func channelOptimization() {
    // Buffered channels reduce goroutine blocking and GC pressure
    ch := make(chan WorkItem, 100)
    
    // Process in batches
    var batch []WorkItem
    for {
        select {
        case item := <-ch:
            batch = append(batch, item)
            if len(batch) >= 10 {
                processBatch(batch)
                batch = batch[:0]  // Reset, keep capacity
            }
            
        case <-time.After(100 * time.Millisecond):
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0]
            }
        }
    }
}
```

## GC Tuning Parameters

### GOGC Environment Variable

```go
import "os"

// Control GC frequency with GOGC
func tuneGCFrequency() {
    // GOGC=100 (default): GC when heap doubles
    // GOGC=50: GC when heap grows 50% (more frequent)
    // GOGC=200: GC when heap grows 200% (less frequent)
    // GOGC=off: Disable automatic GC
    
    // Set programmatically
    oldGCPercent := debug.SetGCPercent(50)  // More aggressive GC
    defer debug.SetGCPercent(oldGCPercent)
    
    // Or via environment
    os.Setenv("GOGC", "50")
    
    // Monitor impact
    monitorGCPerformance()
}

// Adaptive GC tuning based on memory pressure
func adaptiveGCTuning() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Calculate memory pressure
    memPressure := float64(m.HeapAlloc) / float64(m.HeapSys)
    
    if memPressure > 0.8 {
        // High memory pressure: more aggressive GC
        debug.SetGCPercent(50)
    } else if memPressure < 0.3 {
        // Low memory pressure: less frequent GC
        debug.SetGCPercent(200)
    } else {
        // Normal: default GC
        debug.SetGCPercent(100)
    }
}
```

### Memory Limit (Go 1.19+)

```go
import "runtime/debug"

// Set soft memory limit (Go 1.19+)
func setMemoryLimit() {
    // Set 500MB limit
    limit := int64(500 * 1024 * 1024)
    debug.SetMemoryLimit(limit)
    
    // GC will be more aggressive when approaching limit
    
    // Monitor memory usage
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    usage := float64(m.HeapSys) / float64(limit)
    fmt.Printf("Memory usage: %.2f%%\n", usage*100)
}
```

## GC Debugging and Monitoring

### GC Trace Analysis

```go
// Enable GC tracing
func enableGCTrace() {
    // Set GODEBUG=gctrace=1
    os.Setenv("GODEBUG", "gctrace=1")
    
    // Sample output:
    // gc 1 @0.004s 11%: 0.018+1.3+0.076 ms clock, 0.14+0.35/1.2/3.0+0.61 ms cpu, 5->6->1 MB, 6 MB goal, 8 P
    //   │    │      │           │                        │                                │        │         │
    //   │    │      │           │                        │                                │        │         └─ Number of processors
    //   │    │      │           │                        │                                │        └─ Target heap size
    //   │    │      │           │                        │                                └─ Heap size: start->peak->end
    //   │    │      │           │                        └─ CPU time: STW+mark+STW
    //   │    │      │           └─ Wall clock time: STW+mark+STW
    //   │    │      └─ Percentage of time in GC since program start
    //   │    └─ Time since program start
    //   └─ GC cycle number
}

// Parse GC trace programmatically
func parseGCTrace() {
    // Custom GC monitoring
    var lastGC uint32
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        if m.NumGC > lastGC {
            gcCount := m.NumGC - lastGC
            lastGC = m.NumGC
            
            recentPause := m.PauseNs[(m.NumGC+255)%256]
            
            fmt.Printf("GC: %d cycles, last pause: %v\n", 
                gcCount, time.Duration(recentPause))
        }
    }
}
```

### Custom GC Metrics

```go
type GCMetrics struct {
    PauseTotal     time.Duration
    PauseAvg       time.Duration
    PauseMax       time.Duration
    CyclesPerSec   float64
    HeapSize       uint64
    ObjectCount    uint64
    GCPressure     float64
}

func collectGCMetrics() GCMetrics {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Calculate average pause time
    var totalPause time.Duration
    var maxPause time.Duration
    sampleSize := 10
    if int(m.NumGC) < sampleSize {
        sampleSize = int(m.NumGC)
    }
    
    for i := 0; i < sampleSize; i++ {
        idx := (m.NumGC - uint32(i) + 255) % 256
        pause := time.Duration(m.PauseNs[idx])
        totalPause += pause
        if pause > maxPause {
            maxPause = pause
        }
    }
    
    avgPause := totalPause / time.Duration(sampleSize)
    
    // Calculate GC pressure (time spent in GC)
    gcPressure := float64(m.PauseTotalNs) / float64(time.Now().UnixNano())
    
    return GCMetrics{
        PauseTotal:   time.Duration(m.PauseTotalNs),
        PauseAvg:     avgPause,
        PauseMax:     maxPause,
        CyclesPerSec: float64(m.NumGC) / time.Since(startTime).Seconds(),
        HeapSize:     m.HeapAlloc,
        ObjectCount:  m.HeapObjects,
        GCPressure:   gcPressure,
    }
}

// Alert on GC anomalies
func monitorGCAnomalies() {
    threshold := GCMetrics{
        PauseMax:     10 * time.Millisecond,
        GCPressure:   0.05,  // 5% of time in GC
        CyclesPerSec: 10,    // More than 10 GC/sec
    }
    
    metrics := collectGCMetrics()
    
    if metrics.PauseMax > threshold.PauseMax {
        log.Printf("WARNING: High GC pause: %v", metrics.PauseMax)
    }
    
    if metrics.GCPressure > threshold.GCPressure {
        log.Printf("WARNING: High GC pressure: %.2f%%", metrics.GCPressure*100)
    }
    
    if metrics.CyclesPerSec > threshold.CyclesPerSec {
        log.Printf("WARNING: High GC frequency: %.2f cycles/sec", metrics.CyclesPerSec)
    }
}
```

## Advanced GC Patterns

### Generational Collection Simulation

```go
// Simulate generational collection with pools
type GenerationalPool struct {
    young sync.Pool  // Short-lived objects
    old   sync.Pool  // Long-lived objects
}

func NewGenerationalPool() *GenerationalPool {
    return &GenerationalPool{
        young: sync.Pool{
            New: func() interface{} {
                return &ShortLivedObject{}
            },
        },
        old: sync.Pool{
            New: func() interface{} {
                return &LongLivedObject{}
            },
        },
    }
}

func (gp *GenerationalPool) GetShortLived() *ShortLivedObject {
    return gp.young.Get().(*ShortLivedObject)
}

func (gp *GenerationalPool) PutShortLived(obj *ShortLivedObject) {
    obj.Reset()
    gp.young.Put(obj)
}

func (gp *GenerationalPool) GetLongLived() *LongLivedObject {
    return gp.old.Get().(*LongLivedObject)
}

func (gp *GenerationalPool) PutLongLived(obj *LongLivedObject) {
    obj.Reset()
    gp.old.Put(obj)
}
```

### GC-Aware Caching

```go
// Cache that respects GC pressure
type GCAwareCache struct {
    mu       sync.RWMutex
    items    map[string]*CacheItem
    maxSize  int
    lastGC   uint32
    gcPauses []time.Duration
}

type CacheItem struct {
    Value    interface{}
    LastUsed time.Time
    Cost     int
}

func (c *GCAwareCache) Get(key string) (interface{}, bool) {
    c.maybeEvict()
    
    c.mu.RLock()
    item, exists := c.items[key]
    c.mu.RUnlock()
    
    if exists {
        item.LastUsed = time.Now()
        return item.Value, true
    }
    
    return nil, false
}

func (c *GCAwareCache) Put(key string, value interface{}, cost int) {
    c.maybeEvict()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.items[key] = &CacheItem{
        Value:    value,
        LastUsed: time.Now(),
        Cost:     cost,
    }
}

func (c *GCAwareCache) maybeEvict() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Check if GC occurred
    if m.NumGC > c.lastGC {
        c.lastGC = m.NumGC
        
        // Collect recent pause times
        recentPause := time.Duration(m.PauseNs[(m.NumGC+255)%256])
        c.gcPauses = append(c.gcPauses, recentPause)
        
        if len(c.gcPauses) > 10 {
            c.gcPauses = c.gcPauses[1:]
        }
        
        // Calculate average pause
        var total time.Duration
        for _, pause := range c.gcPauses {
            total += pause
        }
        avgPause := total / time.Duration(len(c.gcPauses))
        
        // Evict more aggressively if GC pressure is high
        if avgPause > 5*time.Millisecond {
            c.aggressiveEvict()
        }
    }
}

func (c *GCAwareCache) aggressiveEvict() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Remove oldest 50% of items when GC pressure is high
    targetSize := len(c.items) / 2
    
    type kv struct {
        key      string
        lastUsed time.Time
    }
    
    items := make([]kv, 0, len(c.items))
    for k, v := range c.items {
        items = append(items, kv{k, v.LastUsed})
    }
    
    // Sort by last used time
    sort.Slice(items, func(i, j int) bool {
        return items[i].lastUsed.Before(items[j].lastUsed)
    })
    
    // Remove oldest items
    for i := 0; i < len(items)-targetSize; i++ {
        delete(c.items, items[i].key)
    }
}
```

## GC Performance Best Practices

### ✅ **Do's**

1. **Minimize allocations**
   ```go
   // Use object pools for frequent allocations
   var pool sync.Pool
   obj := pool.Get()
   defer pool.Put(obj)
   ```

2. **Prefer stack allocation**
   ```go
   // Return values, not pointers when possible
   func process() Result { return Result{} }
   ```

3. **Group pointer fields**
   ```go
   type Optimized struct {
       ptr1, ptr2 *Data  // Pointers together
       val1, val2 int64  // Values together
   }
   ```

4. **Pre-allocate collections**
   ```go
   slice := make([]Item, 0, expectedSize)
   ```

5. **Monitor GC metrics**
   ```go
   runtime.ReadMemStats(&m)
   ```

### ❌ **Don'ts**

1. **Don't ignore allocation patterns**
   ```go
   // Profile allocation hotspots
   go tool pprof alloc.profile
   ```

2. **Don't mix pointers and values unnecessarily**
   ```go
   // This causes more GC scanning
   type Mixed struct {
       val int64
       ptr *Data
       val2 int64
   }
   ```

3. **Don't disable GC without careful consideration**
   ```go
   // GOGC=off - only for batch processing
   ```

4. **Don't ignore write barriers**
   ```go
   // Minimize pointer assignments in hot paths
   ```

5. **Don't create excessive object graphs**
   ```go
   // Deep object hierarchies slow GC scanning
   ```

## GC and Memory Safety

### Finalizers

```go
import "runtime"

// Use finalizers sparingly for cleanup
type Resource struct {
    handle unsafe.Pointer
}

func NewResource() *Resource {
    r := &Resource{
        handle: allocateResource(),
    }
    
    // Set finalizer for cleanup
    runtime.SetFinalizer(r, (*Resource).cleanup)
    return r
}

func (r *Resource) Close() {
    if r.handle != nil {
        freeResource(r.handle)
        r.handle = nil
        
        // Clear finalizer since we cleaned up explicitly
        runtime.SetFinalizer(r, nil)
    }
}

func (r *Resource) cleanup() {
    if r.handle != nil {
        freeResource(r.handle)
    }
}
```

Understanding Go's garbage collector enables you to write memory-efficient applications with predictable latency characteristics and optimal resource utilization.

---

**Next**: [Profiling Tools Overview](../profiling-tools/overview.md) - Learn about Go's profiling ecosystem
