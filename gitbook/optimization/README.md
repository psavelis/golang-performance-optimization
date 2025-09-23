# Optimization Strategies

Master systematic approaches to Go performance optimization, from identifying bottlenecks to implementing targeted improvements across algorithms, data structures, and system architecture.

## Performance Optimization Methodology

### The Optimization Process

```go
// 1. Measure First - Establish Baseline
func BenchmarkBaseline(b *testing.B) {
    data := generateTestData(10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := originalAlgorithm(data)
        _ = result
    }
}

// 2. Profile to Identify Bottlenecks
func ProfileBottlenecks() {
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Run your application
    runApplicationWorkload()
}

// 3. Optimize Targeted Areas
func BenchmarkOptimized(b *testing.B) {
    data := generateTestData(10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := optimizedAlgorithm(data)
        _ = result
    }
}

// 4. Validate Improvements
func BenchmarkComparison(b *testing.B) {
    data := generateTestData(10000)
    
    b.Run("Original", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := originalAlgorithm(data)
            _ = result
        }
    })
    
    b.Run("Optimized", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := optimizedAlgorithm(data)
            _ = result
        }
    })
}
```

## Algorithm Optimization

### Complexity Reduction

```go
package main

import (
    "fmt"
    "sort"
    "time"
)

// O(n²) - Inefficient nested loops
func findDuplicatesNaive(data []int) []int {
    var duplicates []int
    
    for i := 0; i < len(data); i++ {
        for j := i + 1; j < len(data); j++ {
            if data[i] == data[j] {
                duplicates = append(duplicates, data[i])
                break
            }
        }
    }
    
    return duplicates
}

// O(n log n) - Sort-based approach
func findDuplicatesSorted(data []int) []int {
    if len(data) < 2 {
        return nil
    }
    
    sorted := make([]int, len(data))
    copy(sorted, data)
    sort.Ints(sorted)
    
    var duplicates []int
    prev := sorted[0]
    
    for i := 1; i < len(sorted); i++ {
        if sorted[i] == prev {
            duplicates = append(duplicates, sorted[i])
            // Skip additional duplicates
            for i+1 < len(sorted) && sorted[i+1] == sorted[i] {
                i++
            }
        }
        prev = sorted[i]
    }
    
    return duplicates
}

// O(n) - Hash map approach
func findDuplicatesHash(data []int) []int {
    seen := make(map[int]bool)
    duplicateSet := make(map[int]bool)
    
    for _, value := range data {
        if seen[value] {
            duplicateSet[value] = true
        } else {
            seen[value] = true
        }
    }
    
    var duplicates []int
    for value := range duplicateSet {
        duplicates = append(duplicates, value)
    }
    
    return duplicates
}

// Benchmark different approaches
func BenchmarkDuplicateDetection(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        data := generateDataWithDuplicates(size)
        
        b.Run(fmt.Sprintf("Naive_Size_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                result := findDuplicatesNaive(data)
                _ = result
            }
        })
        
        b.Run(fmt.Sprintf("Sorted_Size_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                result := findDuplicatesSorted(data)
                _ = result
            }
        })
        
        b.Run(fmt.Sprintf("Hash_Size_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                result := findDuplicatesHash(data)
                _ = result
            }
        })
    }
}

func generateDataWithDuplicates(size int) []int {
    data := make([]int, size)
    for i := range data {
        data[i] = i % (size / 10) // Create duplicates
    }
    return data
}
```

### Caching and Memoization

```go
package main

import (
    "sync"
    "time"
)

// Expensive computation without caching
func expensiveComputation(n int) int {
    time.Sleep(time.Millisecond) // Simulate expensive operation
    if n <= 1 {
        return n
    }
    return expensiveComputation(n-1) + expensiveComputation(n-2)
}

// Memoized version with cache
type MemoizedComputer struct {
    cache map[int]int
    mu    sync.RWMutex
}

func NewMemoizedComputer() *MemoizedComputer {
    return &MemoizedComputer{
        cache: make(map[int]int),
    }
}

func (mc *MemoizedComputer) Compute(n int) int {
    // Check cache first
    mc.mu.RLock()
    if result, exists := mc.cache[n]; exists {
        mc.mu.RUnlock()
        return result
    }
    mc.mu.RUnlock()
    
    // Compute and cache result
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    // Double-check pattern
    if result, exists := mc.cache[n]; exists {
        return result
    }
    
    var result int
    if n <= 1 {
        result = n
    } else {
        result = mc.Compute(n-1) + mc.Compute(n-2)
    }
    
    mc.cache[n] = result
    return result
}

// LRU Cache implementation for bounded memory usage
type LRUCache struct {
    capacity int
    cache    map[int]*Node
    head     *Node
    tail     *Node
}

type Node struct {
    key   int
    value int
    prev  *Node
    next  *Node
}

func NewLRUCache(capacity int) *LRUCache {
    head := &Node{}
    tail := &Node{}
    head.next = tail
    tail.prev = head
    
    return &LRUCache{
        capacity: capacity,
        cache:    make(map[int]*Node),
        head:     head,
        tail:     tail,
    }
}

func (lru *LRUCache) Get(key int) (int, bool) {
    if node, exists := lru.cache[key]; exists {
        lru.moveToHead(node)
        return node.value, true
    }
    return 0, false
}

func (lru *LRUCache) Put(key, value int) {
    if node, exists := lru.cache[key]; exists {
        node.value = value
        lru.moveToHead(node)
    } else {
        newNode := &Node{key: key, value: value}
        
        if len(lru.cache) >= lru.capacity {
            // Remove least recently used
            tail := lru.removeTail()
            delete(lru.cache, tail.key)
        }
        
        lru.cache[key] = newNode
        lru.addToHead(newNode)
    }
}

func (lru *LRUCache) moveToHead(node *Node) {
    lru.removeNode(node)
    lru.addToHead(node)
}

func (lru *LRUCache) removeNode(node *Node) {
    node.prev.next = node.next
    node.next.prev = node.prev
}

func (lru *LRUCache) addToHead(node *Node) {
    node.prev = lru.head
    node.next = lru.head.next
    lru.head.next.prev = node
    lru.head.next = node
}

func (lru *LRUCache) removeTail() *Node {
    last := lru.tail.prev
    lru.removeNode(last)
    return last
}

// Cached computation with LRU eviction
type CachedComputer struct {
    cache *LRUCache
}

func NewCachedComputer(cacheSize int) *CachedComputer {
    return &CachedComputer{
        cache: NewLRUCache(cacheSize),
    }
}

func (cc *CachedComputer) Compute(n int) int {
    if result, exists := cc.cache.Get(n); exists {
        return result
    }
    
    var result int
    if n <= 1 {
        result = n
    } else {
        result = cc.Compute(n-1) + cc.Compute(n-2)
    }
    
    cc.cache.Put(n, result)
    return result
}
```

## Data Structure Optimization

### Efficient Data Access Patterns

```go
package main

import (
    "sort"
    "strings"
)

// Inefficient string operations
func buildStringNaive(words []string) string {
    var result string
    for _, word := range words {
        result += word + " "
    }
    return strings.TrimSpace(result)
}

// Optimized with strings.Builder
func buildStringOptimized(words []string) string {
    var builder strings.Builder
    
    // Pre-allocate capacity if known
    totalLen := 0
    for _, word := range words {
        totalLen += len(word) + 1 // +1 for space
    }
    builder.Grow(totalLen)
    
    for i, word := range words {
        if i > 0 {
            builder.WriteByte(' ')
        }
        builder.WriteString(word)
    }
    
    return builder.String()
}

// Slice optimization: pre-allocation vs growth
func processDataNaive(input []int) []int {
    var result []int
    
    for _, value := range input {
        if value%2 == 0 {
            result = append(result, value*2)
        }
    }
    
    return result
}

func processDataOptimized(input []int) []int {
    // Pre-allocate with estimated capacity
    result := make([]int, 0, len(input)/2)
    
    for _, value := range input {
        if value%2 == 0 {
            result = append(result, value*2)
        }
    }
    
    return result
}

// Map optimization: pre-sizing and key types
func buildMapNaive() map[string]int {
    m := make(map[string]int)
    
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        m[key] = i
    }
    
    return m
}

func buildMapOptimized() map[string]int {
    // Pre-size map to avoid rehashing
    m := make(map[string]int, 10000)
    
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        m[key] = i
    }
    
    return m
}

// Struct optimization: field ordering and packing
type UnoptimizedStruct struct {
    flag1  bool    // 1 byte + 7 bytes padding
    value1 int64   // 8 bytes
    flag2  bool    // 1 byte + 7 bytes padding
    value2 int64   // 8 bytes
    // Total: 32 bytes
}

type OptimizedStruct struct {
    value1 int64   // 8 bytes
    value2 int64   // 8 bytes
    flag1  bool    // 1 byte
    flag2  bool    // 1 byte + 6 bytes padding
    // Total: 24 bytes
}

// Interface vs concrete types
type Processor interface {
    Process(data []int) int
}

type ConcreteProcessor struct {
    multiplier int
}

func (cp *ConcreteProcessor) Process(data []int) int {
    sum := 0
    for _, value := range data {
        sum += value * cp.multiplier
    }
    return sum
}

func BenchmarkInterfaceVsConcrete(b *testing.B) {
    data := make([]int, 1000)
    for i := range data {
        data[i] = i
    }
    
    processor := &ConcreteProcessor{multiplier: 2}
    
    b.Run("Interface", func(b *testing.B) {
        var p Processor = processor
        for i := 0; i < b.N; i++ {
            result := p.Process(data)
            _ = result
        }
    })
    
    b.Run("Concrete", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := processor.Process(data)
            _ = result
        }
    })
}
```

### Memory Pool Optimization

```go
package main

import (
    "sync"
)

// Object pooling for expensive allocations
type Buffer struct {
    data []byte
}

func (b *Buffer) Reset() {
    b.data = b.data[:0]
}

type BufferPool struct {
    pool sync.Pool
}

func NewBufferPool(initialSize int) *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &Buffer{
                    data: make([]byte, 0, initialSize),
                }
            },
        },
    }
}

func (bp *BufferPool) Get() *Buffer {
    return bp.pool.Get().(*Buffer)
}

func (bp *BufferPool) Put(b *Buffer) {
    b.Reset()
    bp.pool.Put(b)
}

// Usage example
func processDataWithPool(pool *BufferPool, input []byte) []byte {
    buffer := pool.Get()
    defer pool.Put(buffer)
    
    // Process data using the pooled buffer
    for _, b := range input {
        if b > 128 {
            buffer.data = append(buffer.data, b)
        }
    }
    
    // Return copy since buffer will be reused
    result := make([]byte, len(buffer.data))
    copy(result, buffer.data)
    return result
}

// Without pool (allocates new buffer each time)
func processDataWithoutPool(input []byte) []byte {
    var buffer []byte
    
    for _, b := range input {
        if b > 128 {
            buffer = append(buffer, b)
        }
    }
    
    return buffer
}

func BenchmarkBufferPool(b *testing.B) {
    input := make([]byte, 10000)
    for i := range input {
        input[i] = byte(i % 256)
    }
    
    pool := NewBufferPool(1024)
    
    b.Run("WithPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := processDataWithPool(pool, input)
            _ = result
        }
    })
    
    b.Run("WithoutPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := processDataWithoutPool(input)
            _ = result
        }
    })
}
```

## Concurrency Optimization

### Goroutine Pool Pattern

```go
package main

import (
    "context"
    "runtime"
    "sync"
    "time"
)

type WorkerPool struct {
    workerCount int
    taskQueue   chan Task
    wg          sync.WaitGroup
    ctx         context.Context
    cancel      context.CancelFunc
}

type Task func() error

func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        workerCount: workerCount,
        taskQueue:   make(chan Task, queueSize),
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case task := <-wp.taskQueue:
            if err := task(); err != nil {
                // Handle error (log, retry, etc.)
                _ = err
            }
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(task Task) error {
    select {
    case wp.taskQueue <- task:
        return nil
    case <-wp.ctx.Done():
        return wp.ctx.Err()
    default:
        return fmt.Errorf("task queue full")
    }
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    close(wp.taskQueue)
    wp.wg.Wait()
}

// Adaptive worker pool that adjusts size based on load
type AdaptiveWorkerPool struct {
    minWorkers  int
    maxWorkers  int
    currentWorkers int
    taskQueue   chan Task
    workerWg    sync.WaitGroup
    ctx         context.Context
    cancel      context.CancelFunc
    mu          sync.RWMutex
}

func NewAdaptiveWorkerPool(minWorkers, maxWorkers, queueSize int) *AdaptiveWorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    awp := &AdaptiveWorkerPool{
        minWorkers:     minWorkers,
        maxWorkers:     maxWorkers,
        currentWorkers: minWorkers,
        taskQueue:      make(chan Task, queueSize),
        ctx:            ctx,
        cancel:         cancel,
    }
    
    return awp
}

func (awp *AdaptiveWorkerPool) Start() {
    // Start minimum workers
    for i := 0; i < awp.minWorkers; i++ {
        awp.workerWg.Add(1)
        go awp.worker()
    }
    
    // Start load monitor
    go awp.monitorLoad()
}

func (awp *AdaptiveWorkerPool) worker() {
    defer awp.workerWg.Done()
    
    idleTimer := time.NewTimer(30 * time.Second)
    defer idleTimer.Stop()
    
    for {
        select {
        case task := <-awp.taskQueue:
            idleTimer.Reset(30 * time.Second)
            if err := task(); err != nil {
                _ = err
            }
        case <-idleTimer.C:
            // Worker idle too long, consider shutdown
            awp.mu.Lock()
            if awp.currentWorkers > awp.minWorkers {
                awp.currentWorkers--
                awp.mu.Unlock()
                return
            }
            awp.mu.Unlock()
            idleTimer.Reset(30 * time.Second)
        case <-awp.ctx.Done():
            return
        }
    }
}

func (awp *AdaptiveWorkerPool) monitorLoad() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            queueLen := len(awp.taskQueue)
            queueCap := cap(awp.taskQueue)
            
            awp.mu.Lock()
            currentWorkers := awp.currentWorkers
            
            // Scale up if queue is getting full
            if queueLen > queueCap*3/4 && currentWorkers < awp.maxWorkers {
                awp.currentWorkers++
                awp.workerWg.Add(1)
                go awp.worker()
            }
            
            awp.mu.Unlock()
        case <-awp.ctx.Done():
            return
        }
    }
}

func (awp *AdaptiveWorkerPool) Submit(task Task) error {
    select {
    case awp.taskQueue <- task:
        return nil
    case <-awp.ctx.Done():
        return awp.ctx.Err()
    default:
        return fmt.Errorf("task queue full")
    }
}

func (awp *AdaptiveWorkerPool) Stop() {
    awp.cancel()
    close(awp.taskQueue)
    awp.workerWg.Wait()
}

// Benchmark different concurrency patterns
func BenchmarkConcurrencyPatterns(b *testing.B) {
    work := func() error {
        time.Sleep(time.Microsecond * 100)
        return nil
    }
    
    b.Run("Goroutine_Per_Task", func(b *testing.B) {
        var wg sync.WaitGroup
        for i := 0; i < b.N; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                work()
            }()
        }
        wg.Wait()
    })
    
    b.Run("Worker_Pool", func(b *testing.B) {
        pool := NewWorkerPool(runtime.NumCPU(), 1000)
        pool.Start()
        defer pool.Stop()
        
        var wg sync.WaitGroup
        for i := 0; i < b.N; i++ {
            wg.Add(1)
            pool.Submit(func() error {
                defer wg.Done()
                return work()
            })
        }
        wg.Wait()
    })
    
    b.Run("Adaptive_Pool", func(b *testing.B) {
        pool := NewAdaptiveWorkerPool(2, runtime.NumCPU()*2, 1000)
        pool.Start()
        defer pool.Stop()
        
        var wg sync.WaitGroup
        for i := 0; i < b.N; i++ {
            wg.Add(1)
            pool.Submit(func() error {
                defer wg.Done()
                return work()
            })
        }
        wg.Wait()
    })
}
```

## System-Level Optimization

### I/O Optimization

```go
package main

import (
    "bufio"
    "io"
    "os"
)

// Inefficient byte-by-byte reading
func readFileNaive(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    var data []byte
    buf := make([]byte, 1)
    
    for {
        n, err := file.Read(buf)
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        if n > 0 {
            data = append(data, buf[0])
        }
    }
    
    return data, nil
}

// Optimized buffered reading
func readFileOptimized(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // Get file size for pre-allocation
    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }
    
    data := make([]byte, stat.Size())
    _, err = io.ReadFull(file, data)
    return data, err
}

// Streaming approach for large files
func processLargeFileStreaming(filename string, processor func([]byte) error) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    reader := bufio.NewReader(file)
    buffer := make([]byte, 64*1024) // 64KB buffer
    
    for {
        n, err := reader.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := processor(buffer[:n]); err != nil {
            return err
        }
    }
    
    return nil
}

// Memory-mapped file access for random access patterns
func readFileMapped(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }
    
    // For large files, consider using mmap
    // This is a simplified version - real implementation would use syscalls
    data := make([]byte, stat.Size())
    _, err = io.ReadFull(file, data)
    return data, err
}
```

Systematic optimization strategies transform Go applications from functional to high-performance, addressing bottlenecks at every level from algorithms to system architecture while maintaining code clarity and maintainability.
