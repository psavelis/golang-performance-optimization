# Go Memory Model

Understanding Go's memory model is essential for writing high-performance applications and concurrent programs. This chapter covers memory allocation strategies, garbage collection impact, and optimization techniques.

## Memory Architecture Overview

### Go Memory Layout

```
┌─────────────────────────────────────────────────────────────┐
│                     Go Memory Layout                       │
├─────────────────────────────────────────────────────────────┤
│ Text Segment    │ Global Data   │ Heap      │ Stack        │
│ (Program Code)  │ (Global Vars) │ (Dynamic) │ (Goroutines) │
└─────────────────────────────────────────────────────────────┘
```

### Memory Regions

#### 1. **Text Segment**
- Contains compiled program code
- Read-only and shared across processes
- Includes function code and constants

#### 2. **Global Data Segment**
- Global variables and static data
- Package-level variables
- String literals and compile-time constants

#### 3. **Heap**
- Dynamic memory allocation
- Garbage-collected memory
- Shared across all goroutines

#### 4. **Stack**
- Per-goroutine stack memory
- Local variables and function calls
- Automatically managed, grows/shrinks as needed

## Stack vs Heap Allocation

### Stack Allocation

Stack allocation is preferred for performance:

```go
func stackAllocation() {
    // These variables are allocated on the stack
    var x int = 42
    var arr [100]int
    var s struct {
        name string
        age  int
    }
    
    // Stack allocated: fast allocation/deallocation
    // No GC pressure
    // Limited lifetime (function scope)
}
```

**Stack Characteristics:**
- **Fast**: Simple pointer arithmetic
- **Automatic**: Cleaned up when function returns
- **Limited**: Fixed size per goroutine (default 8KB, grows to 1GB)
- **Thread-safe**: Each goroutine has its own stack

### Heap Allocation

Heap allocation occurs when data escapes:

```go
func heapAllocation() *int {
    x := 42
    return &x  // x escapes to heap via return
}

func heapAllocation2() {
    s := make([]int, 1000000)  // Large slice -> heap
    _ = s
}

func heapAllocation3() {
    var i interface{} = 42  // Interface -> heap
    _ = i
}
```

**Heap Characteristics:**
- **Flexible**: Dynamic sizing
- **Persistent**: Survives function calls
- **GC Managed**: Automatic cleanup via garbage collector
- **Slower**: Allocation and deallocation overhead

## Escape Analysis

Go's compiler performs escape analysis to determine allocation location:

### Escape Analysis Rules

```go
package main

// go build -gcflags="-m" main.go  // Shows escape analysis

func noEscape() {
    x := 42        // Does not escape: stack allocated
    _ = x
}

func escapeViaReturn() *int {
    x := 42        // Escapes via return: heap allocated
    return &x
}

func escapeViaInterface() {
    x := 42        // Escapes to interface{}: heap allocated
    var i interface{} = x
    _ = i
}

func escapeViaSlice() {
    x := 42
    s := []*int{&x}  // x escapes via slice: heap allocated
    _ = s
}

func escapeViaChannel() {
    x := 42
    ch := make(chan *int, 1)
    ch <- &x       // x escapes via channel: heap allocated
}

func escapeViaClosure() func() int {
    x := 42
    return func() int {  // x escapes via closure: heap allocated
        return x
    }
}
```

### Controlling Escape Analysis

```go
// Minimize escapes for better performance
func optimizedFunction(data []byte) {
    // Use value receivers when possible
    processor := DataProcessor{
        buffer: make([]byte, len(data)),  // May escape
    }
    
    // Copy data instead of referencing
    copy(processor.buffer, data)
    
    // Process without creating references
    result := processor.Process()  // Value return, no escape
    _ = result
}

// Pool objects to reduce heap allocations
var processorPool = sync.Pool{
    New: func() interface{} {
        return &DataProcessor{
            buffer: make([]byte, 1024),
        }
    },
}

func pooledProcessing(data []byte) {
    processor := processorPool.Get().(*DataProcessor)
    defer processorPool.Put(processor)
    
    processor.Reset()
    processor.ProcessData(data)
}
```

## Memory Allocation Patterns

### Small Object Allocation

Go optimizes small object allocation with size classes:

```go
// Size classes for small objects (< 32KB)
// 8, 16, 24, 32, 48, 64, 80, 96, 112, 128, ...

func smallObjectOptimization() {
    // These allocations use optimized size classes
    s1 := make([]byte, 16)   // Exactly fits 16-byte class
    s2 := make([]byte, 17)   // Uses 24-byte class (overhead)
    s3 := make([]byte, 24)   // Exactly fits 24-byte class
    
    _, _, _ = s1, s2, s3
}

// Prefer sizes that align with size classes
type OptimizedStruct struct {
    field1 uint64  // 8 bytes
    field2 uint64  // 8 bytes
    // Total: 16 bytes (fits 16-byte size class perfectly)
}

type SuboptimalStruct struct {
    field1 uint64  // 8 bytes
    field2 uint64  // 8 bytes
    field3 byte    // 1 byte + 7 bytes padding
    // Total: 24 bytes (uses 24-byte size class, wastes padding)
}
```

### Large Object Allocation

Large objects (>32KB) bypass size classes:

```go
func largeObjectAllocation() {
    // Large objects allocated directly from heap
    largeSlice := make([]byte, 1024*1024)  // 1MB
    largeMap := make(map[string][]byte, 10000)
    
    // Consider streaming or chunking for large data
    processInChunks(largeSlice, 4096)  // Process 4KB at a time
}

func processInChunks(data []byte, chunkSize int) {
    for i := 0; i < len(data); i += chunkSize {
        end := i + chunkSize
        if end > len(data) {
            end = len(data)
        }
        
        chunk := data[i:end]  // Slice doesn't allocate new memory
        processChunk(chunk)
    }
}
```

### Memory Pools and Reuse

Implement object pooling for frequently allocated objects:

```go
// Buffer pool for reducing allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)  // 1KB initial capacity
    },
}

func pooledProcessing(data []byte) []byte {
    buf := bufferPool.Get().([]byte)
    buf = buf[:0]  // Reset length, keep capacity
    defer bufferPool.Put(buf)
    
    // Use buffer for processing
    buf = append(buf, processData(data)...)
    
    // Return copy, not the pooled buffer
    result := make([]byte, len(buf))
    copy(result, buf)
    return result
}

// Object pool for complex structures
type ExpensiveObject struct {
    data   []byte
    cache  map[string]interface{}
    worker chan Job
}

var objectPool = sync.Pool{
    New: func() interface{} {
        return &ExpensiveObject{
            data:   make([]byte, 0, 1024),
            cache:  make(map[string]interface{}),
            worker: make(chan Job, 10),
        }
    },
}

func (o *ExpensiveObject) Reset() {
    o.data = o.data[:0]
    for k := range o.cache {
        delete(o.cache, k)
    }
    // Drain worker channel
    for len(o.worker) > 0 {
        <-o.worker
    }
}
```

## Memory Layout Optimization

### Struct Field Ordering

Optimize struct layout to minimize memory usage:

```go
// Poor layout: 32 bytes due to padding
type PoorLayout struct {
    flag1    bool     // 1 byte
    // 7 bytes padding
    value1   uint64   // 8 bytes
    flag2    bool     // 1 byte
    // 7 bytes padding
    value2   uint64   // 8 bytes
    // Total: 32 bytes
}

// Optimized layout: 24 bytes
type OptimizedLayout struct {
    value1   uint64   // 8 bytes
    value2   uint64   // 8 bytes
    flag1    bool     // 1 byte
    flag2    bool     // 1 byte
    // 6 bytes padding
    // Total: 24 bytes (25% smaller!)
}

// Even better: use bit fields for flags
type BitFieldLayout struct {
    value1   uint64   // 8 bytes
    value2   uint64   // 8 bytes
    flags    uint8    // 1 byte (can hold 8 boolean flags)
    // 7 bytes padding
    // Total: 24 bytes, but can store more flags
}
```

### Cache Line Optimization

Optimize for CPU cache performance:

```go
// Cache line size is typically 64 bytes
const CacheLineSize = 64

// Align hot data to cache line boundaries
type CacheOptimizedStruct struct {
    // Hot data: frequently accessed together
    counter   uint64
    timestamp uint64
    status    uint32
    flags     uint32
    // Total: 24 bytes - fits in single cache line with room for more
    
    // Padding to next cache line
    _ [CacheLineSize - 24]byte
    
    // Cold data: less frequently accessed
    metadata string
    config   map[string]interface{}
}

// False sharing prevention
type AtomicCounters struct {
    counter1 uint64
    _        [CacheLineSize - 8]byte  // Prevent false sharing
    counter2 uint64
    _        [CacheLineSize - 8]byte  // Prevent false sharing
    counter3 uint64
}
```

## Memory Management Patterns

### Pre-allocation Strategies

```go
// Pre-allocate slices with known capacity
func preallocationOptimization(items []Item) []Result {
    // Bad: frequent reallocations
    var results []Result
    for _, item := range items {
        results = append(results, processItem(item))
    }
    
    // Good: pre-allocate with known size
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, processItem(item))
    }
    
    return results
}

// Map pre-allocation
func mapPreallocation(items []Item) map[string]Result {
    // Good: pre-allocate map with estimated size
    results := make(map[string]Result, len(items))
    
    for _, item := range items {
        results[item.ID] = processItem(item)
    }
    
    return results
}

// Builder pre-allocation
func stringBuilderOptimization(items []string) string {
    // Calculate total size to avoid reallocations
    totalSize := 0
    for _, item := range items {
        totalSize += len(item) + 1  // +1 for separator
    }
    
    var builder strings.Builder
    builder.Grow(totalSize)  // Pre-allocate buffer
    
    for i, item := range items {
        if i > 0 {
            builder.WriteByte(',')
        }
        builder.WriteString(item)
    }
    
    return builder.String()
}
```

### Memory Streaming Patterns

```go
// Stream processing to avoid loading entire dataset
func streamProcessing(reader io.Reader, writer io.Writer) error {
    buffer := make([]byte, 64*1024)  // 64KB buffer
    
    for {
        n, err := reader.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }
        
        // Process chunk without loading everything into memory
        processed := processChunk(buffer[:n])
        
        if _, err := writer.Write(processed); err != nil {
            return err
        }
    }
    
    return nil
}

// Batch processing with bounded memory
func batchProcessing(items <-chan Item, batchSize int) {
    batch := make([]Item, 0, batchSize)
    
    for item := range items {
        batch = append(batch, item)
        
        if len(batch) == batchSize {
            processBatch(batch)
            batch = batch[:0]  // Reset slice, keep capacity
        }
    }
    
    // Process remaining items
    if len(batch) > 0 {
        processBatch(batch)
    }
}
```

## Memory Profiling and Analysis

### Memory Profile Interpretation

```go
// Example: Analyzing memory allocation hotspots
func memoryProfilingExample() {
    // Start memory profiling
    if *memProfile != "" {
        f, err := os.Create(*memProfile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        defer f.Close()
        
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
    }
    
    // Application code that allocates memory
    allocateMemory()
}

func allocateMemory() {
    // Various allocation patterns for profiling
    
    // 1. Large slice allocation
    largeSlice := make([]int, 1000000)
    
    // 2. Many small allocations
    for i := 0; i < 10000; i++ {
        _ = make([]byte, 64)
    }
    
    // 3. Map allocations
    m := make(map[string][]byte)
    for i := 0; i < 1000; i++ {
        key := fmt.Sprintf("key%d", i)
        m[key] = make([]byte, 256)
    }
    
    // Keep references to prevent GC
    _ = largeSlice
    _ = m
}
```

### Memory Leak Detection

```go
// Common memory leak patterns and prevention
func memoryLeakPrevention() {
    // 1. Goroutine leaks
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    go func() {
        select {
        case <-doWork():
            // Work completed
        case <-ctx.Done():
            // Timeout or cancellation
        }
    }()
    
    // 2. Reference cycles
    type Node struct {
        children []*Node
        parent   *Node  // Potential cycle
    }
    
    // Break cycles explicitly
    func (n *Node) Cleanup() {
        for _, child := range n.children {
            child.parent = nil
            child.Cleanup()
        }
        n.children = nil
    }
    
    // 3. Global variable accumulation
    var globalCache = make(map[string]interface{})
    var cacheMutex sync.RWMutex
    
    func addToCache(key string, value interface{}) {
        cacheMutex.Lock()
        defer cacheMutex.Unlock()
        
        // Implement cache size limit
        if len(globalCache) > 10000 {
            // Remove oldest entries
            for k := range globalCache {
                delete(globalCache, k)
                if len(globalCache) <= 5000 {
                    break
                }
            }
        }
        
        globalCache[key] = value
    }
}
```

## Advanced Memory Techniques

### Memory Mapping

```go
import (
    "os"
    "syscall"
    "unsafe"
)

// Memory mapping for large files
func memoryMappedFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    fileInfo, err := file.Stat()
    if err != nil {
        return nil, err
    }
    
    // Memory map the file
    data, err := syscall.Mmap(int(file.Fd()), 0, int(fileInfo.Size()),
        syscall.PROT_READ, syscall.MAP_SHARED)
    if err != nil {
        return nil, err
    }
    
    return data, nil
}

// Cleanup memory mapping
func unmapFile(data []byte) error {
    return syscall.Munmap(data)
}
```

### Unsafe Optimizations

```go
import "unsafe"

// Zero-copy string to byte conversion (read-only)
func stringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(
        &struct {
            string
            Cap int
        }{s, len(s)},
    ))
}

// Zero-copy byte to string conversion
func bytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

// Use with extreme caution - violates Go memory safety guarantees
func unsafeOptimization() {
    s := "hello world"
    b := stringToBytes(s)  // No allocation
    
    // DANGER: Do not modify b - it points to string data!
    _ = b
}
```

## Memory Performance Best Practices

### ✅ **Do's**

1. **Pre-allocate with known sizes**
   ```go
   slice := make([]Item, 0, expectedSize)
   m := make(map[string]Item, expectedSize)
   ```

2. **Use object pools for frequent allocations**
   ```go
   var pool sync.Pool
   obj := pool.Get()
   defer pool.Put(obj)
   ```

3. **Optimize struct layout**
   ```go
   // Group fields by size, largest first
   type Optimized struct {
       ptr    *Data   // 8 bytes
       value  uint64  // 8 bytes
       flag   bool    // 1 byte
   }
   ```

4. **Reuse buffers**
   ```go
   buf := buf[:0]  // Reset length, keep capacity
   ```

5. **Stream large data**
   ```go
   // Process in chunks instead of loading everything
   ```

### ❌ **Don'ts**

1. **Don't ignore escape analysis**
   ```go
   // Check with: go build -gcflags="-m"
   ```

2. **Don't create unnecessary interfaces**
   ```go
   // Interfaces cause heap allocation
   var i interface{} = value  // Heap allocation
   ```

3. **Don't append in loops without pre-allocation**
   ```go
   // This causes multiple reallocations
   for _, item := range items {
       result = append(result, process(item))
   }
   ```

4. **Don't ignore memory leaks**
   ```go
   // Always clean up goroutines and close channels
   ```

5. **Don't premature optimize**
   ```go
   // Profile first, then optimize based on data
   ```

## Memory Model Guarantees

### Happens-Before Relationships

Go's memory model defines when memory operations are visible across goroutines:

```go
// Channel operations provide happens-before guarantees
func channelSynchronization() {
    done := make(chan bool)
    var data string
    
    go func() {
        data = "hello"  // Happens before channel send
        done <- true
    }()
    
    <-done              // Happens after channel receive
    fmt.Println(data)   // Guaranteed to see "hello"
}

// Mutex operations provide happens-before guarantees
func mutexSynchronization() {
    var mu sync.Mutex
    var data string
    
    go func() {
        mu.Lock()
        data = "hello"  // Happens before unlock
        mu.Unlock()
    }()
    
    mu.Lock()           // Happens after lock
    fmt.Println(data)   // May see "hello" or ""
    mu.Unlock()
}
```

Understanding Go's memory model enables you to write efficient, correct concurrent programs while minimizing allocation overhead and maximizing cache performance.

---

**Next**: [Goroutine Scheduler](goroutine-scheduler.md) - Learn how Go schedules and manages concurrent execution
