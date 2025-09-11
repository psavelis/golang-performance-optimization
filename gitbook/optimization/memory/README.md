# Memory Optimization

Memory optimization is crucial for building efficient, scalable Go applications. This comprehensive guide covers memory management strategies, allocation patterns, optimization techniques, and tools for achieving optimal memory performance.

## Introduction to Memory Optimization

Effective memory optimization in Go involves understanding:
- **Allocation patterns** and their performance implications
- **Garbage collection** behavior and optimization strategies
- **Memory pools** and object reuse techniques
- **Stack vs heap** allocation trade-offs
- **Memory layout** optimization for cache efficiency

### Key Memory Optimization Principles

1. **Minimize allocations** - Reduce GC pressure
2. **Reuse objects** - Avoid repeated allocations
3. **Choose appropriate data structures** - Optimize for access patterns
4. **Understand escape analysis** - Keep allocations on stack when possible
5. **Profile regularly** - Measure to validate optimizations

## Memory Allocation Patterns

### Understanding Go Memory Model

```go
package main

import (
    "fmt"
    "runtime"
    "unsafe"
)

func demonstrateAllocationPatterns() {
    fmt.Println("=== MEMORY ALLOCATION PATTERNS ===")
    
    // Stack allocation (efficient)
    stackAllocation()
    
    // Heap allocation (requires GC)
    heapAllocation()
    
    // String allocation patterns
    stringAllocation()
    
    // Slice allocation patterns
    sliceAllocation()
    
    // Map allocation patterns
    mapAllocation()
}

func stackAllocation() {
    fmt.Println("\n--- Stack Allocation ---")
    
    // These typically stay on stack
    var a int = 42
    var b [100]int
    var c struct{ x, y int }
    
    fmt.Printf("Stack variables: a=%d, b[0]=%d, c.x=%d\n", a, b[0], c.x)
    
    // Function-local variables with known size
    processLocalData()
}

func processLocalData() {
    buffer := [1024]byte{} // Stack allocated
    counter := 0
    
    for i := range buffer {
        buffer[i] = byte(i % 256)
        counter++
    }
    
    fmt.Printf("Processed %d bytes on stack\n", counter)
}

func heapAllocation() {
    fmt.Println("\n--- Heap Allocation ---")
    
    // These escape to heap
    slice := make([]int, 1000)    // Dynamic size
    ptr := &slice[0]              // Taking address causes escape
    dynamic := make([]int, getValue()) // Size not known at compile time
    
    fmt.Printf("Heap allocations: slice len=%d, ptr=%p, dynamic len=%d\n", 
        len(slice), ptr, len(dynamic))
}

func getValue() int {
    return 500 // Dynamic value
}

func stringAllocation() {
    fmt.Println("\n--- String Allocation Patterns ---")
    
    // Efficient: String literals (in read-only memory)
    literal := "Hello, World!"
    
    // Inefficient: String concatenation (creates new allocations)
    var concatenated string
    for i := 0; i < 5; i++ {
        concatenated += fmt.Sprintf("Item %d ", i) // Each += allocates
    }
    
    // Efficient: Using strings.Builder
    var builder strings.Builder
    for i := 0; i < 5; i++ {
        builder.WriteString(fmt.Sprintf("Item %d ", i))
    }
    efficient := builder.String()
    
    fmt.Printf("Literal: %s\n", literal)
    fmt.Printf("Concatenated: %s\n", concatenated)
    fmt.Printf("Builder: %s\n", efficient)
}

func sliceAllocation() {
    fmt.Println("\n--- Slice Allocation Patterns ---")
    
    // Inefficient: Growing slice without capacity
    var inefficient []int
    for i := 0; i < 1000; i++ {
        inefficient = append(inefficient, i) // Multiple reallocations
    }
    
    // Efficient: Pre-allocating capacity
    efficient := make([]int, 0, 1000)
    for i := 0; i < 1000; i++ {
        efficient = append(efficient, i) // No reallocations
    }
    
    fmt.Printf("Inefficient slice len=%d, cap=%d\n", len(inefficient), cap(inefficient))
    fmt.Printf("Efficient slice len=%d, cap=%d\n", len(efficient), cap(efficient))
}

func mapAllocation() {
    fmt.Println("\n--- Map Allocation Patterns ---")
    
    // Inefficient: Default map growth
    inefficient := make(map[int]string)
    for i := 0; i < 1000; i++ {
        inefficient[i] = fmt.Sprintf("value_%d", i)
    }
    
    // Efficient: Pre-sizing map
    efficient := make(map[int]string, 1000)
    for i := 0; i < 1000; i++ {
        efficient[i] = fmt.Sprintf("value_%d", i)
    }
    
    fmt.Printf("Maps created with %d entries each\n", len(inefficient))
}
```

### Memory Layout Optimization

```go
package main

import (
    "fmt"
    "unsafe"
)

// Inefficient struct layout (24 bytes due to padding)
type InefficientStruct struct {
    A bool   // 1 byte
    B int64  // 8 bytes (but starts at offset 8 due to alignment)
    C bool   // 1 byte  
    D int32  // 4 bytes (but starts at offset 20 due to alignment)
}

// Efficient struct layout (16 bytes, reordered fields)
type EfficientStruct struct {
    B int64  // 8 bytes
    D int32  // 4 bytes
    A bool   // 1 byte
    C bool   // 1 byte
    // 2 bytes padding at end
}

// Very efficient with bit packing for booleans
type VeryEfficientStruct struct {
    B int64  // 8 bytes
    D int32  // 4 bytes
    Flags uint8 // Pack multiple booleans into single byte
}

func demonstrateMemoryLayout() {
    fmt.Println("=== MEMORY LAYOUT OPTIMIZATION ===")
    
    fmt.Printf("InefficientStruct size: %d bytes\n", unsafe.Sizeof(InefficientStruct{}))
    fmt.Printf("EfficientStruct size: %d bytes\n", unsafe.Sizeof(EfficientStruct{}))
    fmt.Printf("VeryEfficientStruct size: %d bytes\n", unsafe.Sizeof(VeryEfficientStruct{}))
    
    // Demonstrate array impact
    inefficientArray := make([]InefficientStruct, 1000)
    efficientArray := make([]EfficientStruct, 1000)
    
    fmt.Printf("1000 inefficient structs: %d bytes\n", 
        len(inefficientArray) * int(unsafe.Sizeof(InefficientStruct{})))
    fmt.Printf("1000 efficient structs: %d bytes\n", 
        len(efficientArray) * int(unsafe.Sizeof(EfficientStruct{})))
    
    memorySaved := len(inefficientArray) * (int(unsafe.Sizeof(InefficientStruct{})) - 
        int(unsafe.Sizeof(EfficientStruct{})))
    fmt.Printf("Memory saved: %d bytes (%.1f%%)\n", memorySaved, 
        float64(memorySaved)/float64(len(inefficientArray)*int(unsafe.Sizeof(InefficientStruct{})))*100)
}

// Cache-friendly data layout
type CacheFriendlyProcessor struct {
    // Hot data (frequently accessed together)
    ID       uint64
    Status   uint32
    Counter  uint32
    
    // Cold data (less frequently accessed)
    Name        string
    Description string
    Metadata    map[string]interface{}
}

// Cache-unfriendly layout
type CacheUnfriendlyProcessor struct {
    ID          uint64
    Name        string
    Status      uint32
    Description string
    Counter     uint32
    Metadata    map[string]interface{}
}

func demonstrateCacheEfficiency() {
    fmt.Println("\n=== CACHE EFFICIENCY ===")
    
    // Process hot path operations on cache-friendly layout
    processors := make([]CacheFriendlyProcessor, 10000)
    
    // Initialize
    for i := range processors {
        processors[i].ID = uint64(i)
        processors[i].Status = uint32(i % 4)
        processors[i].Counter = 0
    }
    
    // Hot path: Only access ID, Status, Counter (first 16 bytes)
    sum := uint64(0)
    for i := range processors {
        if processors[i].Status == 1 {
            processors[i].Counter++
            sum += processors[i].ID
        }
    }
    
    fmt.Printf("Processed %d cache-friendly structs, sum=%d\n", len(processors), sum)
}
```

## Advanced Memory Optimization Techniques

### Object Pooling

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Example: Buffer pool for frequent allocations
type BufferPool struct {
    pool sync.Pool
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                // Create new buffer when pool is empty
                return make([]byte, 0, 1024)
            },
        },
    }
}

func (bp *BufferPool) Get() []byte {
    return bp.pool.Get().([]byte)
}

func (bp *BufferPool) Put(buf []byte) {
    // Reset buffer but keep capacity
    buf = buf[:0]
    bp.pool.Put(buf)
}

// Example: Object pool for complex structures
type WorkerPool struct {
    pool sync.Pool
}

type Worker struct {
    ID       int
    Buffer   []byte
    Results  map[string]interface{}
    Status   string
}

func NewWorkerPool() *WorkerPool {
    return &WorkerPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &Worker{
                    Buffer:  make([]byte, 0, 4096),
                    Results: make(map[string]interface{}),
                }
            },
        },
    }
}

func (wp *WorkerPool) Get() *Worker {
    worker := wp.pool.Get().(*Worker)
    worker.Reset()
    return worker
}

func (wp *WorkerPool) Put(worker *Worker) {
    wp.pool.Put(worker)
}

func (w *Worker) Reset() {
    w.ID = 0
    w.Buffer = w.Buffer[:0]
    // Clear map but reuse it
    for k := range w.Results {
        delete(w.Results, k)
    }
    w.Status = ""
}

func (w *Worker) Process(data []byte) error {
    // Simulate processing
    w.Buffer = append(w.Buffer, data...)
    w.Results["processed_bytes"] = len(data)
    w.Status = "completed"
    return nil
}

func demonstrateObjectPooling() {
    fmt.Println("=== OBJECT POOLING ===")
    
    bufferPool := NewBufferPool()
    workerPool := NewWorkerPool()
    
    // Simulate high-frequency operations
    const iterations = 10000
    
    // Without pooling (creates new objects each time)
    start := time.Now()
    for i := 0; i < iterations; i++ {
        buf := make([]byte, 0, 1024)
        buf = append(buf, []byte("test data")...)
        // buf would be garbage collected
    }
    withoutPooling := time.Since(start)
    
    // With buffer pooling
    start = time.Now()
    for i := 0; i < iterations; i++ {
        buf := bufferPool.Get()
        buf = append(buf, []byte("test data")...)
        bufferPool.Put(buf)
    }
    withPooling := time.Since(start)
    
    // Complex object pooling
    start = time.Now()
    for i := 0; i < 1000; i++ {
        worker := workerPool.Get()
        worker.ID = i
        worker.Process([]byte("complex data"))
        workerPool.Put(worker)
    }
    complexPooling := time.Since(start)
    
    fmt.Printf("Without pooling: %v\n", withoutPooling)
    fmt.Printf("With buffer pooling: %v (%.1fx faster)\n", withPooling, 
        float64(withoutPooling)/float64(withPooling))
    fmt.Printf("Complex object pooling: %v for 1000 operations\n", complexPooling)
}
```

### Memory-Efficient Data Structures

```go
package main

import (
    "fmt"
    "unsafe"
)

// Memory-efficient boolean slice using bit packing
type BitSet struct {
    bits []uint64
    size int
}

func NewBitSet(size int) *BitSet {
    numWords := (size + 63) / 64
    return &BitSet{
        bits: make([]uint64, numWords),
        size: size,
    }
}

func (bs *BitSet) Set(index int) {
    if index >= bs.size {
        return
    }
    wordIndex := index / 64
    bitIndex := index % 64
    bs.bits[wordIndex] |= 1 << bitIndex
}

func (bs *BitSet) Get(index int) bool {
    if index >= bs.size {
        return false
    }
    wordIndex := index / 64
    bitIndex := index % 64
    return (bs.bits[wordIndex] & (1 << bitIndex)) != 0
}

func (bs *BitSet) Clear(index int) {
    if index >= bs.size {
        return
    }
    wordIndex := index / 64
    bitIndex := index % 64
    bs.bits[wordIndex] &^= 1 << bitIndex
}

func (bs *BitSet) MemoryUsage() int {
    return len(bs.bits) * 8 // 8 bytes per uint64
}

// Compare with regular boolean slice
func compareBitSetMemoryUsage() {
    fmt.Println("\n=== BIT SET MEMORY EFFICIENCY ===")
    
    const size = 100000
    
    // Regular boolean slice
    boolSlice := make([]bool, size)
    boolMemory := len(boolSlice) * int(unsafe.Sizeof(bool(true)))
    
    // BitSet
    bitSet := NewBitSet(size)
    bitSetMemory := bitSet.MemoryUsage()
    
    fmt.Printf("100k booleans as []bool: %d bytes\n", boolMemory)
    fmt.Printf("100k booleans as BitSet: %d bytes\n", bitSetMemory)
    fmt.Printf("Memory saved: %d bytes (%.1fx reduction)\n", 
        boolMemory-bitSetMemory, float64(boolMemory)/float64(bitSetMemory))
    
    // Performance test
    start := time.Now()
    for i := 0; i < size; i++ {
        boolSlice[i] = i%2 == 0
    }
    boolSliceTime := time.Since(start)
    
    start = time.Now()
    for i := 0; i < size; i++ {
        if i%2 == 0 {
            bitSet.Set(i)
        }
    }
    bitSetTime := time.Since(start)
    
    fmt.Printf("[]bool write time: %v\n", boolSliceTime)
    fmt.Printf("BitSet write time: %v\n", bitSetTime)
}

// Memory-efficient string interning
type StringInterner struct {
    strings map[string]string
    mutex   sync.RWMutex
}

func NewStringInterner() *StringInterner {
    return &StringInterner{
        strings: make(map[string]string),
    }
}

func (si *StringInterner) Intern(s string) string {
    si.mutex.RLock()
    if interned, exists := si.strings[s]; exists {
        si.mutex.RUnlock()
        return interned
    }
    si.mutex.RUnlock()
    
    si.mutex.Lock()
    defer si.mutex.Unlock()
    
    // Double-check pattern
    if interned, exists := si.strings[s]; exists {
        return interned
    }
    
    // Copy string to ensure we own the memory
    interned := string([]byte(s))
    si.strings[interned] = interned
    return interned
}

func (si *StringInterner) Stats() (int, int) {
    si.mutex.RLock()
    defer si.mutex.RUnlock()
    
    count := len(si.strings)
    totalBytes := 0
    for s := range si.strings {
        totalBytes += len(s)
    }
    
    return count, totalBytes
}

func demonstrateStringInterning() {
    fmt.Println("\n=== STRING INTERNING ===")
    
    interner := NewStringInterner()
    
    // Common strings that would be duplicated
    commonStrings := []string{
        "user_id", "session_id", "request_id", "timestamp",
        "status", "method", "path", "response_time",
    }
    
    // Simulate processing many records with duplicate strings
    const numRecords = 100000
    
    // Without interning
    var withoutInterning []map[string]string
    start := time.Now()
    for i := 0; i < numRecords; i++ {
        record := make(map[string]string)
        for _, key := range commonStrings {
            record[key] = fmt.Sprintf("%s_%d", key, i%1000) // Many duplicates
        }
        withoutInterning = append(withoutInterning, record)
    }
    withoutInterningTime := time.Since(start)
    
    // With interning
    var withInterning []map[string]string
    start = time.Now()
    for i := 0; i < numRecords; i++ {
        record := make(map[string]string)
        for _, key := range commonStrings {
            internedKey := interner.Intern(key)
            internedValue := interner.Intern(fmt.Sprintf("%s_%d", key, i%1000))
            record[internedKey] = internedValue
        }
        withInterning = append(withInterning, record)
    }
    withInterningTime := time.Since(start)
    
    internedCount, internedBytes := interner.Stats()
    
    fmt.Printf("Processed %d records\n", numRecords)
    fmt.Printf("Without interning time: %v\n", withoutInterningTime)
    fmt.Printf("With interning time: %v\n", withInterningTime)
    fmt.Printf("Interned strings: %d unique strings, %d bytes\n", internedCount, internedBytes)
    
    // Estimate memory savings (rough calculation)
    avgStringLength := 20
    estimatedWithoutInterning := numRecords * len(commonStrings) * 2 * avgStringLength
    estimatedWithInterning := internedBytes
    
    fmt.Printf("Estimated memory usage without interning: %d bytes\n", estimatedWithoutInterning)
    fmt.Printf("Estimated memory usage with interning: %d bytes\n", estimatedWithInterning)
    fmt.Printf("Estimated memory saved: %.1fx reduction\n", 
        float64(estimatedWithoutInterning)/float64(estimatedWithInterning))
}
```

### Garbage Collection Optimization

```go
package main

import (
    "fmt"
    "runtime"
    "runtime/debug"
    "time"
)

type GCOptimizer struct {
    originalGOGC int
    targetGC     time.Duration
}

func NewGCOptimizer() *GCOptimizer {
    return &GCOptimizer{
        originalGOGC: debug.SetGCPercent(-1), // Get current setting
        targetGC:     10 * time.Millisecond,
    }
}

func (gco *GCOptimizer) OptimizeForLatency() {
    // Optimize for low latency (more frequent, shorter GC pauses)
    debug.SetGCPercent(50) // GC when heap grows 50% (vs default 100%)
    
    fmt.Println("GC optimized for latency (frequent, short collections)")
}

func (gco *GCOptimizer) OptimizeForThroughput() {
    // Optimize for throughput (less frequent, longer GC pauses)
    debug.SetGCPercent(200) // GC when heap grows 200%
    
    fmt.Println("GC optimized for throughput (infrequent, longer collections)")
}

func (gco *GCOptimizer) RestoreDefaults() {
    debug.SetGCPercent(gco.originalGOGC)
    fmt.Println("GC settings restored to defaults")
}

func (gco *GCOptimizer) MonitorGC(duration time.Duration) {
    start := time.Now()
    var stats runtime.MemStats
    
    runtime.ReadMemStats(&stats)
    initialGC := stats.NumGC
    initialPause := stats.PauseTotalNs
    
    fmt.Printf("Starting GC monitoring for %v...\n", duration)
    
    time.Sleep(duration)
    
    runtime.ReadMemStats(&stats)
    finalGC := stats.NumGC
    finalPause := stats.PauseTotalNs
    
    gcCount := finalGC - initialGC
    totalPause := time.Duration(finalPause - initialPause)
    avgPause := time.Duration(0)
    if gcCount > 0 {
        avgPause = totalPause / time.Duration(gcCount)
    }
    
    fmt.Printf("GC Statistics over %v:\n", duration)
    fmt.Printf("  GC runs: %d\n", gcCount)
    fmt.Printf("  Total pause time: %v\n", totalPause)
    fmt.Printf("  Average pause: %v\n", avgPause)
    fmt.Printf("  GC frequency: %.2f/sec\n", float64(gcCount)/duration.Seconds())
    fmt.Printf("  Heap size: %s\n", formatBytes(stats.HeapInuse))
    fmt.Printf("  Allocated: %s\n", formatBytes(stats.TotalAlloc))
}

func demonstrateGCOptimization() {
    fmt.Println("=== GARBAGE COLLECTION OPTIMIZATION ===")
    
    optimizer := NewGCOptimizer()
    defer optimizer.RestoreDefaults()
    
    // Create allocation workload
    workload := func() {
        for i := 0; i < 1000; i++ {
            // Create temporary allocations
            data := make([]byte, 1024*10) // 10KB
            data[0] = byte(i)
            
            // Some long-lived allocations
            if i%100 == 0 {
                longLived := make([]byte, 1024*100) // 100KB
                longLived[0] = byte(i)
                // Simulate keeping some data
                time.Sleep(time.Microsecond)
            }
            
            time.Sleep(time.Microsecond * 100)
        }
    }
    
    fmt.Println("\n--- Default GC Settings ---")
    runtime.GC() // Clear state
    optimizer.MonitorGC(2 * time.Second)
    go workload()
    time.Sleep(3 * time.Second)
    
    fmt.Println("\n--- Latency-Optimized GC ---")
    optimizer.OptimizeForLatency()
    runtime.GC() // Clear state
    go workload()
    optimizer.MonitorGC(2 * time.Second)
    time.Sleep(3 * time.Second)
    
    fmt.Println("\n--- Throughput-Optimized GC ---")
    optimizer.OptimizeForThroughput()
    runtime.GC() // Clear state
    go workload()
    optimizer.MonitorGC(2 * time.Second)
    time.Sleep(3 * time.Second)
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

## Memory Optimization Patterns

### Zero-Allocation Patterns

```go
package main

import (
    "fmt"
    "strconv"
    "strings"
    "time"
)

// Zero-allocation integer to string conversion
type IntBuffer struct {
    buf [32]byte // Enough for largest int64
}

func (ib *IntBuffer) FormatInt(n int64) string {
    if n == 0 {
        return "0"
    }
    
    negative := n < 0
    if negative {
        n = -n
    }
    
    i := len(ib.buf)
    for n > 0 {
        i--
        ib.buf[i] = '0' + byte(n%10)
        n /= 10
    }
    
    if negative {
        i--
        ib.buf[i] = '-'
    }
    
    return string(ib.buf[i:])
}

// Zero-allocation string building
type StringBuilder struct {
    buf []byte
}

func NewStringBuilder(initialCap int) *StringBuilder {
    return &StringBuilder{
        buf: make([]byte, 0, initialCap),
    }
}

func (sb *StringBuilder) WriteString(s string) {
    sb.buf = append(sb.buf, s...)
}

func (sb *StringBuilder) WriteByte(b byte) {
    sb.buf = append(sb.buf, b)
}

func (sb *StringBuilder) String() string {
    return string(sb.buf)
}

func (sb *StringBuilder) Reset() {
    sb.buf = sb.buf[:0]
}

func (sb *StringBuilder) Len() int {
    return len(sb.buf)
}

// Zero-allocation JSON-like formatting
func formatRecord(sb *StringBuilder, id int, name string, active bool) {
    sb.Reset()
    sb.WriteString(`{"id":`)
    
    // Zero-allocation int conversion
    var intBuf IntBuffer
    sb.WriteString(intBuf.FormatInt(int64(id)))
    
    sb.WriteString(`,"name":"`)
    sb.WriteString(name)
    sb.WriteString(`","active":`)
    
    if active {
        sb.WriteString("true")
    } else {
        sb.WriteString("false")
    }
    
    sb.WriteByte('}')
}

func demonstrateZeroAllocation() {
    fmt.Println("\n=== ZERO-ALLOCATION PATTERNS ===")
    
    const iterations = 100000
    
    // With allocations (standard approach)
    start := time.Now()
    for i := 0; i < iterations; i++ {
        result := fmt.Sprintf(`{"id":%d,"name":"%s","active":%t}`, 
            i, "test_user", i%2 == 0)
        _ = result
    }
    withAllocations := time.Since(start)
    
    // Zero-allocation approach
    sb := NewStringBuilder(64)
    start = time.Now()
    for i := 0; i < iterations; i++ {
        formatRecord(sb, i, "test_user", i%2 == 0)
        result := sb.String()
        _ = result
    }
    zeroAllocation := time.Since(start)
    
    fmt.Printf("With allocations: %v\n", withAllocations)
    fmt.Printf("Zero allocation: %v\n", zeroAllocation)
    fmt.Printf("Improvement: %.1fx faster\n", float64(withAllocations)/float64(zeroAllocation))
    
    // Show example output
    formatRecord(sb, 123, "example_user", true)
    fmt.Printf("Example output: %s\n", sb.String())
}
```

### Memory-Efficient Algorithms

```go
package main

import (
    "fmt"
    "sort"
    "time"
)

// Memory-efficient sorting for large datasets
func inPlaceQuickSort(data []int, low, high int) {
    if low < high {
        pi := partition(data, low, high)
        inPlaceQuickSort(data, low, pi-1)
        inPlaceQuickSort(data, pi+1, high)
    }
}

func partition(data []int, low, high int) int {
    pivot := data[high]
    i := low - 1
    
    for j := low; j < high; j++ {
        if data[j] < pivot {
            i++
            data[i], data[j] = data[j], data[i]
        }
    }
    
    data[i+1], data[high] = data[high], data[i+1]
    return i + 1
}

// Memory-efficient duplicate removal
func removeDuplicatesInPlace(data []int) []int {
    if len(data) == 0 {
        return data
    }
    
    // Sort first (in-place)
    sort.Ints(data)
    
    // Remove duplicates in-place
    j := 0
    for i := 1; i < len(data); i++ {
        if data[i] != data[j] {
            j++
            data[j] = data[i]
        }
    }
    
    return data[:j+1]
}

// Memory-efficient merging of sorted slices
func mergeSortedInPlace(a, b []int) []int {
    result := make([]int, 0, len(a)+len(b))
    i, j := 0, 0
    
    for i < len(a) && j < len(b) {
        if a[i] <= b[j] {
            result = append(result, a[i])
            i++
        } else {
            result = append(result, b[j])
            j++
        }
    }
    
    // Append remaining elements
    result = append(result, a[i:]...)
    result = append(result, b[j:]...)
    
    return result
}

func demonstrateMemoryEfficientAlgorithms() {
    fmt.Println("\n=== MEMORY-EFFICIENT ALGORITHMS ===")
    
    // Generate test data
    data1 := make([]int, 10000)
    for i := range data1 {
        data1[i] = (i * 7) % 1000 // Create some duplicates
    }
    
    data2 := make([]int, len(data1))
    copy(data2, data1)
    
    // Standard sort vs in-place sort (memory comparison)
    start := time.Now()
    sort.Ints(data1)
    standardSort := time.Since(start)
    
    start = time.Now()
    inPlaceQuickSort(data2, 0, len(data2)-1)
    inPlaceSort := time.Since(start)
    
    fmt.Printf("Standard sort: %v\n", standardSort)
    fmt.Printf("In-place sort: %v\n", inPlaceSort)
    
    // Duplicate removal
    duplicateData := make([]int, 0, 1000)
    for i := 0; i < 1000; i++ {
        duplicateData = append(duplicateData, i%100) // Many duplicates
    }
    
    originalLen := len(duplicateData)
    uniqueData := removeDuplicatesInPlace(duplicateData)
    
    fmt.Printf("Original length: %d\n", originalLen)
    fmt.Printf("After duplicate removal: %d\n", len(uniqueData))
    fmt.Printf("Memory saved: %d elements\n", originalLen-len(uniqueData))
}
```

## Best Practices for Memory Optimization

### 1. Profile-Driven Optimization

```go
func profileMemoryUsage() {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Code to profile
    performMemoryIntensiveOperation()
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Memory used: %d bytes\n", m2.TotalAlloc-m1.TotalAlloc)
    fmt.Printf("Heap objects: %d\n", m2.HeapObjects)
    fmt.Printf("GC cycles: %d\n", m2.NumGC-m1.NumGC)
}
```

### 2. Benchmark Memory Allocations

```go
func BenchmarkMemoryOptimization(b *testing.B) {
    b.ReportAllocs() // Report allocations
    
    for i := 0; i < b.N; i++ {
        optimizedFunction()
    }
}
```

### 3. Use Build Tags for Memory Profiling

```go
// +build memprofile

package main

import _ "net/http/pprof"
```

## Next Steps

- Explore [Memory Pools](memory-pools.md) implementation
- Learn [Object Reuse](object-reuse.md) patterns
- Study [Stack vs Heap](stack-vs-heap.md) allocation strategies
- Master [Memory Layout](memory-layout.md) optimization

## Summary

Memory optimization in Go requires understanding:

1. **Allocation patterns** and their performance impact
2. **Object pooling** for high-frequency allocations
3. **Memory layout** optimization for cache efficiency
4. **Garbage collection** tuning for workload characteristics
5. **Zero-allocation** patterns for critical paths

Apply these techniques systematically, always measuring the impact to ensure optimizations provide real benefits in your specific use case.
