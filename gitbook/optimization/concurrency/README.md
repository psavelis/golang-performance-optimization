# Concurrency Optimization

Concurrency optimization in Go involves designing efficient goroutine patterns, optimizing channel operations, minimizing synchronization overhead, and building scalable concurrent systems. This comprehensive guide covers advanced concurrency optimization techniques for high-performance Go applications.

## Introduction to Concurrency Optimization

Effective concurrency optimization requires understanding:
- **Goroutine lifecycle** and scheduling behavior
- **Channel patterns** and their performance characteristics
- **Synchronization primitives** and their trade-offs
- **Lock-free programming** techniques
- **Work distribution** strategies

### Key Concurrency Principles

1. **Minimize contention** - Reduce shared state access
2. **Use appropriate patterns** - Choose the right concurrency pattern
3. **Avoid over-parallelization** - More goroutines ≠ better performance
4. **Design for scalability** - Consider workload growth
5. **Profile concurrency** - Measure actual contention and blocking

## Goroutine Optimization Patterns

### Efficient Worker Pool Pattern

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
)

// WorkItem represents a unit of work
type WorkItem struct {
    ID   int
    Data []byte
    Done chan error
}

// WorkerPool implements an efficient worker pool
type WorkerPool struct {
    numWorkers int
    jobQueue   chan WorkItem
    workers    []*Worker
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

type Worker struct {
    id       int
    pool     *WorkerPool
    jobQueue chan WorkItem
    quit     chan bool
}

func NewWorkerPool(numWorkers, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    pool := &WorkerPool{
        numWorkers: numWorkers,
        jobQueue:   make(chan WorkItem, queueSize),
        workers:    make([]*Worker, numWorkers),
        ctx:        ctx,
        cancel:     cancel,
    }
    
    // Create workers
    for i := 0; i < numWorkers; i++ {
        pool.workers[i] = &Worker{
            id:       i,
            pool:     pool,
            jobQueue: pool.jobQueue,
            quit:     make(chan bool),
        }
    }
    
    return pool
}

func (wp *WorkerPool) Start() {
    for _, worker := range wp.workers {
        wp.wg.Add(1)
        go worker.start()
    }
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    
    for _, worker := range wp.workers {
        worker.quit <- true
    }
    
    wp.wg.Wait()
    close(wp.jobQueue)
}

func (wp *WorkerPool) Submit(work WorkItem) bool {
    select {
    case wp.jobQueue <- work:
        return true
    case <-wp.ctx.Done():
        return false
    default:
        return false // Queue full
    }
}

func (wp *WorkerPool) SubmitBlocking(work WorkItem) error {
    select {
    case wp.jobQueue <- work:
        return nil
    case <-wp.ctx.Done():
        return fmt.Errorf("worker pool shutting down")
    }
}

func (w *Worker) start() {
    defer w.pool.wg.Done()
    
    for {
        select {
        case work := <-w.jobQueue:
            err := w.processWork(work)
            if work.Done != nil {
                work.Done <- err
            }
            
        case <-w.quit:
            return
            
        case <-w.pool.ctx.Done():
            return
        }
    }
}

func (w *Worker) processWork(work WorkItem) error {
    // Simulate work processing
    time.Sleep(time.Millisecond * time.Duration(len(work.Data)))
    
    // Process the data
    checksum := 0
    for _, b := range work.Data {
        checksum += int(b)
    }
    
    fmt.Printf("Worker %d processed item %d (checksum: %d)\n", 
        w.id, work.ID, checksum)
    
    return nil
}

// Adaptive Worker Pool that adjusts size based on load
type AdaptiveWorkerPool struct {
    *WorkerPool
    minWorkers    int
    maxWorkers    int
    loadThreshold float64
    scaleInterval time.Duration
    mutex         sync.RWMutex
}

func NewAdaptiveWorkerPool(minWorkers, maxWorkers, queueSize int) *AdaptiveWorkerPool {
    base := NewWorkerPool(minWorkers, queueSize)
    
    adaptive := &AdaptiveWorkerPool{
        WorkerPool:    base,
        minWorkers:    minWorkers,
        maxWorkers:    maxWorkers,
        loadThreshold: 0.8, // Scale up when 80% of queue is full
        scaleInterval: 5 * time.Second,
    }
    
    return adaptive
}

func (awp *AdaptiveWorkerPool) StartWithAutoScaling() {
    awp.Start()
    go awp.autoScale()
}

func (awp *AdaptiveWorkerPool) autoScale() {
    ticker := time.NewTicker(awp.scaleInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            awp.checkAndScale()
        case <-awp.ctx.Done():
            return
        }
    }
}

func (awp *AdaptiveWorkerPool) checkAndScale() {
    awp.mutex.Lock()
    defer awp.mutex.Unlock()
    
    queueLen := len(awp.jobQueue)
    queueCap := cap(awp.jobQueue)
    currentWorkers := awp.numWorkers
    
    loadRatio := float64(queueLen) / float64(queueCap)
    
    if loadRatio > awp.loadThreshold && currentWorkers < awp.maxWorkers {
        // Scale up
        newWorker := &Worker{
            id:       currentWorkers,
            pool:     awp.WorkerPool,
            jobQueue: awp.jobQueue,
            quit:     make(chan bool),
        }
        
        awp.workers = append(awp.workers, newWorker)
        awp.numWorkers++
        awp.wg.Add(1)
        go newWorker.start()
        
        fmt.Printf("Scaled up: %d workers (load: %.2f)\n", awp.numWorkers, loadRatio)
        
    } else if loadRatio < 0.2 && currentWorkers > awp.minWorkers {
        // Scale down
        lastWorker := awp.workers[len(awp.workers)-1]
        lastWorker.quit <- true
        
        awp.workers = awp.workers[:len(awp.workers)-1]
        awp.numWorkers--
        
        fmt.Printf("Scaled down: %d workers (load: %.2f)\n", awp.numWorkers, loadRatio)
    }
}

func demonstrateWorkerPools() {
    fmt.Println("=== WORKER POOL OPTIMIZATION ===")
    
    // Fixed worker pool
    numCPU := runtime.NumCPU()
    pool := NewWorkerPool(numCPU, 100)
    pool.Start()
    defer pool.Stop()
    
    // Generate work
    start := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            work := WorkItem{
                ID:   id,
                Data: make([]byte, 10+id%50), // Variable work size
                Done: make(chan error, 1),
            }
            
            if pool.Submit(work) {
                <-work.Done // Wait for completion
            } else {
                fmt.Printf("Failed to submit work item %d\n", id)
            }
        }(i)
    }
    
    wg.Wait()
    fixedPoolTime := time.Since(start)
    
    // Adaptive worker pool
    adaptivePool := NewAdaptiveWorkerPool(2, numCPU*2, 100)
    adaptivePool.StartWithAutoScaling()
    defer adaptivePool.Stop()
    
    start = time.Now()
    wg = sync.WaitGroup{}
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            work := WorkItem{
                ID:   id,
                Data: make([]byte, 10+id%50),
                Done: make(chan error, 1),
            }
            
            if adaptivePool.Submit(work) {
                <-work.Done
            }
        }(i)
    }
    
    wg.Wait()
    adaptivePoolTime := time.Since(start)
    
    fmt.Printf("Fixed pool (%d workers): %v\n", numCPU, fixedPoolTime)
    fmt.Printf("Adaptive pool: %v\n", adaptivePoolTime)
}
```

### Lock-Free Data Structures

```go
package main

import (
    "sync/atomic"
    "unsafe"
)

// Lock-free stack using atomic operations
type LockFreeStack struct {
    head unsafe.Pointer
}

type stackNode struct {
    next unsafe.Pointer
    data interface{}
}

func NewLockFreeStack() *LockFreeStack {
    return &LockFreeStack{}
}

func (s *LockFreeStack) Push(data interface{}) {
    newNode := &stackNode{data: data}
    
    for {
        currentHead := atomic.LoadPointer(&s.head)
        newNode.next = currentHead
        
        if atomic.CompareAndSwapPointer(&s.head, currentHead, unsafe.Pointer(newNode)) {
            break
        }
        // Retry if CAS failed
    }
}

func (s *LockFreeStack) Pop() interface{} {
    for {
        currentHead := atomic.LoadPointer(&s.head)
        if currentHead == nil {
            return nil // Stack is empty
        }
        
        headNode := (*stackNode)(currentHead)
        nextHead := atomic.LoadPointer(&headNode.next)
        
        if atomic.CompareAndSwapPointer(&s.head, currentHead, nextHead) {
            return headNode.data
        }
        // Retry if CAS failed
    }
}

// Lock-free queue using atomic operations
type LockFreeQueue struct {
    head unsafe.Pointer
    tail unsafe.Pointer
}

type queueNode struct {
    next unsafe.Pointer
    data interface{}
}

func NewLockFreeQueue() *LockFreeQueue {
    dummy := &queueNode{}
    q := &LockFreeQueue{
        head: unsafe.Pointer(dummy),
        tail: unsafe.Pointer(dummy),
    }
    return q
}

func (q *LockFreeQueue) Enqueue(data interface{}) {
    newNode := &queueNode{data: data}
    
    for {
        tail := atomic.LoadPointer(&q.tail)
        tailNode := (*queueNode)(tail)
        next := atomic.LoadPointer(&tailNode.next)
        
        if tail == atomic.LoadPointer(&q.tail) { // Ensure tail hasn't changed
            if next == nil {
                // Try to link new node at end of list
                if atomic.CompareAndSwapPointer(&tailNode.next, next, unsafe.Pointer(newNode)) {
                    atomic.CompareAndSwapPointer(&q.tail, tail, unsafe.Pointer(newNode))
                    break
                }
            } else {
                // Help advance tail
                atomic.CompareAndSwapPointer(&q.tail, tail, next)
            }
        }
    }
}

func (q *LockFreeQueue) Dequeue() interface{} {
    for {
        head := atomic.LoadPointer(&q.head)
        tail := atomic.LoadPointer(&q.tail)
        headNode := (*queueNode)(head)
        next := atomic.LoadPointer(&headNode.next)
        
        if head == atomic.LoadPointer(&q.head) { // Ensure head hasn't changed
            if head == tail {
                if next == nil {
                    return nil // Queue is empty
                }
                // Help advance tail
                atomic.CompareAndSwapPointer(&q.tail, tail, next)
            } else {
                if next == nil {
                    continue // Another thread is in the process of updating
                }
                
                nextNode := (*queueNode)(next)
                data := nextNode.data
                
                // Try to move head to next node
                if atomic.CompareAndSwapPointer(&q.head, head, next) {
                    return data
                }
            }
        }
    }
}

// Lock-free counter
type LockFreeCounter struct {
    value int64
}

func (c *LockFreeCounter) Increment() int64 {
    return atomic.AddInt64(&c.value, 1)
}

func (c *LockFreeCounter) Decrement() int64 {
    return atomic.AddInt64(&c.value, -1)
}

func (c *LockFreeCounter) Add(delta int64) int64 {
    return atomic.AddInt64(&c.value, delta)
}

func (c *LockFreeCounter) Get() int64 {
    return atomic.LoadInt64(&c.value)
}

func (c *LockFreeCounter) CompareAndSwap(old, new int64) bool {
    return atomic.CompareAndSwapInt64(&c.value, old, new)
}

func demonstrateLockFreeStructures() {
    fmt.Println("\n=== LOCK-FREE DATA STRUCTURES ===")
    
    // Test lock-free stack
    stack := NewLockFreeStack()
    var wg sync.WaitGroup
    
    // Concurrent pushes
    numGoroutines := 100
    itemsPerGoroutine := 100
    
    start := time.Now()
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(base int) {
            defer wg.Done()
            for j := 0; j < itemsPerGoroutine; j++ {
                stack.Push(base*itemsPerGoroutine + j)
            }
        }(i)
    }
    wg.Wait()
    pushTime := time.Since(start)
    
    // Concurrent pops
    start = time.Now()
    poppedCount := int64(0)
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < itemsPerGoroutine; j++ {
                if stack.Pop() != nil {
                    atomic.AddInt64(&poppedCount, 1)
                }
            }
        }()
    }
    wg.Wait()
    popTime := time.Since(start)
    
    fmt.Printf("Lock-free stack: %d pushes in %v, %d pops in %v\n", 
        numGoroutines*itemsPerGoroutine, pushTime, poppedCount, popTime)
    
    // Test lock-free queue
    queue := NewLockFreeQueue()
    
    start = time.Now()
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(base int) {
            defer wg.Done()
            for j := 0; j < itemsPerGoroutine; j++ {
                queue.Enqueue(base*itemsPerGoroutine + j)
            }
        }(i)
    }
    wg.Wait()
    enqueueTime := time.Since(start)
    
    start = time.Now()
    dequeuedCount := int64(0)
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < itemsPerGoroutine; j++ {
                if queue.Dequeue() != nil {
                    atomic.AddInt64(&dequeuedCount, 1)
                }
            }
        }()
    }
    wg.Wait()
    dequeueTime := time.Since(start)
    
    fmt.Printf("Lock-free queue: %d enqueues in %v, %d dequeues in %v\n", 
        numGoroutines*itemsPerGoroutine, enqueueTime, dequeuedCount, dequeueTime)
}
```

### Channel Optimization Patterns

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
)

// Optimized fan-out pattern
type FanOutProcessor struct {
    input       <-chan interface{}
    outputs     []chan interface{}
    numWorkers  int
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

func NewFanOutProcessor(input <-chan interface{}, numWorkers int) *FanOutProcessor {
    ctx, cancel := context.WithCancel(context.Background())
    
    outputs := make([]chan interface{}, numWorkers)
    for i := range outputs {
        outputs[i] = make(chan interface{}, 10) // Buffered channels
    }
    
    return &FanOutProcessor{
        input:      input,
        outputs:    outputs,
        numWorkers: numWorkers,
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (fop *FanOutProcessor) Start() {
    fop.wg.Add(1)
    go fop.distribute()
}

func (fop *FanOutProcessor) Stop() {
    fop.cancel()
    fop.wg.Wait()
    
    for _, output := range fop.outputs {
        close(output)
    }
}

func (fop *FanOutProcessor) GetOutput(index int) <-chan interface{} {
    if index < len(fop.outputs) {
        return fop.outputs[index]
    }
    return nil
}

func (fop *FanOutProcessor) distribute() {
    defer fop.wg.Done()
    
    index := 0
    for {
        select {
        case data, ok := <-fop.input:
            if !ok {
                return
            }
            
            // Round-robin distribution
            select {
            case fop.outputs[index] <- data:
                index = (index + 1) % fop.numWorkers
            case <-fop.ctx.Done():
                return
            }
            
        case <-fop.ctx.Done():
            return
        }
    }
}

// Optimized fan-in pattern
type FanInProcessor struct {
    inputs  []<-chan interface{}
    output  chan interface{}
    ctx     context.Context
    cancel  context.CancelFunc
    wg      sync.WaitGroup
}

func NewFanInProcessor(inputs []<-chan interface{}) *FanInProcessor {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &FanInProcessor{
        inputs: inputs,
        output: make(chan interface{}, len(inputs)*10), // Buffered
        ctx:    ctx,
        cancel: cancel,
    }
}

func (fip *FanInProcessor) Start() {
    for i, input := range fip.inputs {
        fip.wg.Add(1)
        go fip.merge(i, input)
    }
}

func (fip *FanInProcessor) Stop() {
    fip.cancel()
    fip.wg.Wait()
    close(fip.output)
}

func (fip *FanInProcessor) GetOutput() <-chan interface{} {
    return fip.output
}

func (fip *FanInProcessor) merge(id int, input <-chan interface{}) {
    defer fip.wg.Done()
    
    for {
        select {
        case data, ok := <-input:
            if !ok {
                return
            }
            
            select {
            case fip.output <- data:
            case <-fip.ctx.Done():
                return
            }
            
        case <-fip.ctx.Done():
            return
        }
    }
}

// Pipeline with backpressure
type Stage func(interface{}) interface{}

type Pipeline struct {
    stages     []Stage
    bufferSize int
    ctx        context.Context
    cancel     context.CancelFunc
}

func NewPipeline(stages []Stage, bufferSize int) *Pipeline {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Pipeline{
        stages:     stages,
        bufferSize: bufferSize,
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (p *Pipeline) Process(input <-chan interface{}) <-chan interface{} {
    current := input
    
    for i, stage := range p.stages {
        current = p.createStage(i, stage, current)
    }
    
    return current
}

func (p *Pipeline) createStage(id int, stage Stage, input <-chan interface{}) <-chan interface{} {
    output := make(chan interface{}, p.bufferSize)
    
    go func() {
        defer close(output)
        
        for {
            select {
            case data, ok := <-input:
                if !ok {
                    return
                }
                
                result := stage(data)
                
                select {
                case output <- result:
                case <-p.ctx.Done():
                    return
                }
                
            case <-p.ctx.Done():
                return
            }
        }
    }()
    
    return output
}

func (p *Pipeline) Stop() {
    p.cancel()
}

func demonstrateChannelOptimization() {
    fmt.Println("\n=== CHANNEL OPTIMIZATION ===")
    
    // Test fan-out/fan-in pattern
    input := make(chan interface{}, 100)
    numWorkers := runtime.NumCPU()
    
    // Create fan-out processor
    fanOut := NewFanOutProcessor(input, numWorkers)
    fanOut.Start()
    
    // Create workers that consume from fan-out
    var workerOutputs []<-chan interface{}
    for i := 0; i < numWorkers; i++ {
        workerOutput := make(chan interface{}, 10)
        workerOutputs = append(workerOutputs, workerOutput)
        
        go func(id int, input <-chan interface{}, output chan interface{}) {
            defer close(output)
            
            for data := range input {
                // Simulate processing
                time.Sleep(time.Microsecond * 100)
                output <- fmt.Sprintf("worker-%d: %v", id, data)
            }
        }(i, fanOut.GetOutput(i), workerOutput)
    }
    
    // Create fan-in processor
    fanIn := NewFanInProcessor(workerOutputs)
    fanIn.Start()
    
    // Generate input data
    go func() {
        defer close(input)
        for i := 0; i < 1000; i++ {
            input <- i
        }
    }()
    
    // Consume results
    start := time.Now()
    resultCount := 0
    for result := range fanIn.GetOutput() {
        resultCount++
        if resultCount%100 == 0 {
            fmt.Printf("Processed %d results: %v\n", resultCount, result)
        }
    }
    
    fanOutTime := time.Since(start)
    fanOut.Stop()
    fanIn.Stop()
    
    fmt.Printf("Fan-out/Fan-in processed %d items in %v\n", resultCount, fanOutTime)
    
    // Test pipeline pattern
    stages := []Stage{
        func(data interface{}) interface{} {
            // Stage 1: Convert to string
            return fmt.Sprintf("stage1:%v", data)
        },
        func(data interface{}) interface{} {
            // Stage 2: Add timestamp
            return fmt.Sprintf("%v:time:%d", data, time.Now().UnixNano())
        },
        func(data interface{}) interface{} {
            // Stage 3: Add prefix
            return fmt.Sprintf("final:%v", data)
        },
    }
    
    pipeline := NewPipeline(stages, 50)
    pipelineInput := make(chan interface{}, 100)
    
    // Process through pipeline
    start = time.Now()
    output := pipeline.Process(pipelineInput)
    
    // Generate input
    go func() {
        defer close(pipelineInput)
        for i := 0; i < 100; i++ {
            pipelineInput <- i
        }
    }()
    
    // Consume output
    pipelineResults := 0
    for result := range output {
        pipelineResults++
        if pipelineResults <= 5 {
            fmt.Printf("Pipeline result: %v\n", result)
        }
    }
    
    pipelineTime := time.Since(start)
    pipeline.Stop()
    
    fmt.Printf("Pipeline processed %d items in %v\n", pipelineResults, pipelineTime)
}
```

### Synchronization Optimization

```go
package main

import (
    "sync"
    "sync/atomic"
    "time"
)

// Read-Write Mutex optimization
type OptimizedCache struct {
    mu    sync.RWMutex
    data  map[string]interface{}
    stats CacheStats
}

type CacheStats struct {
    reads  int64
    writes int64
    hits   int64
    misses int64
}

func NewOptimizedCache() *OptimizedCache {
    return &OptimizedCache{
        data: make(map[string]interface{}),
    }
}

func (c *OptimizedCache) Get(key string) (interface{}, bool) {
    atomic.AddInt64(&c.stats.reads, 1)
    
    c.mu.RLock()
    value, exists := c.data[key]
    c.mu.RUnlock()
    
    if exists {
        atomic.AddInt64(&c.stats.hits, 1)
    } else {
        atomic.AddInt64(&c.stats.misses, 1)
    }
    
    return value, exists
}

func (c *OptimizedCache) Set(key string, value interface{}) {
    atomic.AddInt64(&c.stats.writes, 1)
    
    c.mu.Lock()
    c.data[key] = value
    c.mu.Unlock()
}

func (c *OptimizedCache) GetStats() CacheStats {
    return CacheStats{
        reads:  atomic.LoadInt64(&c.stats.reads),
        writes: atomic.LoadInt64(&c.stats.writes),
        hits:   atomic.LoadInt64(&c.stats.hits),
        misses: atomic.LoadInt64(&c.stats.misses),
    }
}

// Sharded map to reduce contention
type ShardedMap struct {
    shards    []*Shard
    numShards int
}

type Shard struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func NewShardedMap(numShards int) *ShardedMap {
    shards := make([]*Shard, numShards)
    for i := range shards {
        shards[i] = &Shard{
            data: make(map[string]interface{}),
        }
    }
    
    return &ShardedMap{
        shards:    shards,
        numShards: numShards,
    }
}

func (sm *ShardedMap) getShard(key string) *Shard {
    hash := fnv32(key)
    return sm.shards[hash%uint32(sm.numShards)]
}

func (sm *ShardedMap) Get(key string) (interface{}, bool) {
    shard := sm.getShard(key)
    shard.mu.RLock()
    value, exists := shard.data[key]
    shard.mu.RUnlock()
    return value, exists
}

func (sm *ShardedMap) Set(key string, value interface{}) {
    shard := sm.getShard(key)
    shard.mu.Lock()
    shard.data[key] = value
    shard.mu.Unlock()
}

func (sm *ShardedMap) Delete(key string) {
    shard := sm.getShard(key)
    shard.mu.Lock()
    delete(shard.data, key)
    shard.mu.Unlock()
}

// Simple FNV-1a hash function
func fnv32(key string) uint32 {
    hash := uint32(2166136261)
    for i := 0; i < len(key); i++ {
        hash ^= uint32(key[i])
        hash *= 16777619
    }
    return hash
}

func demonstrateSynchronizationOptimization() {
    fmt.Println("\n=== SYNCHRONIZATION OPTIMIZATION ===")
    
    // Test regular map vs sharded map
    numGoroutines := 100
    operationsPerGoroutine := 1000
    
    // Regular cache
    cache := NewOptimizedCache()
    start := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < operationsPerGoroutine; j++ {
                key := fmt.Sprintf("key_%d_%d", id, j)
                cache.Set(key, j)
                cache.Get(key)
            }
        }(i)
    }
    wg.Wait()
    cacheTime := time.Since(start)
    
    // Sharded map
    numShards := runtime.NumCPU() * 2
    shardedMap := NewShardedMap(numShards)
    start = time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < operationsPerGoroutine; j++ {
                key := fmt.Sprintf("key_%d_%d", id, j)
                shardedMap.Set(key, j)
                shardedMap.Get(key)
            }
        }(i)
    }
    wg.Wait()
    shardedTime := time.Since(start)
    
    stats := cache.GetStats()
    fmt.Printf("Regular cache: %v (reads: %d, writes: %d, hit rate: %.2f%%)\n", 
        cacheTime, stats.reads, stats.writes, 
        float64(stats.hits)/float64(stats.reads)*100)
    fmt.Printf("Sharded map (%d shards): %v (%.1fx faster)\n", 
        numShards, shardedTime, float64(cacheTime)/float64(shardedTime))
}
```

## Best Practices for Concurrency Optimization

### 1. Choose the Right Concurrency Pattern

```go
// Use channels for communication
func pipelinePattern(input <-chan Data) <-chan Result {
    output := make(chan Result)
    go func() {
        defer close(output)
        for data := range input {
            result := process(data)
            output <- result
        }
    }()
    return output
}

// Use sync primitives for protecting shared state
func sharedStatePattern() {
    var mu sync.RWMutex
    var cache map[string]interface{}
    
    get := func(key string) interface{} {
        mu.RLock()
        defer mu.RUnlock()
        return cache[key]
    }
    
    set := func(key string, value interface{}) {
        mu.Lock()
        defer mu.Unlock()
        cache[key] = value
    }
}
```

### 2. Optimize for Your Workload

```go
// CPU-bound: Use GOMAXPROCS workers
func cpuBoundPattern() {
    numWorkers := runtime.NumCPU()
    // ... implementation
}

// I/O-bound: Use more workers than CPU cores
func ioBoundPattern() {
    numWorkers := runtime.NumCPU() * 4
    // ... implementation
}
```

### 3. Profile Concurrency Issues

```bash
# Check for contention
go tool pprof http://localhost:6060/debug/pprof/mutex

# Check for blocking
go tool pprof http://localhost:6060/debug/pprof/block

# Check goroutine states
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Next Steps

- Learn [Goroutine Patterns](goroutine-patterns.md) in detail
- Study [Channel Optimization](channel-optimization.md) techniques
- Explore [Worker Pools](worker-pools.md) implementation
- Master [Lock-Free Programming](lock-free.md) patterns

## Summary

Concurrency optimization in Go requires:

1. **Understanding workload characteristics** - CPU vs I/O bound
2. **Choosing appropriate patterns** - Channels vs shared state
3. **Minimizing contention** - Sharding and lock-free techniques
4. **Optimizing for scalability** - Adaptive worker pools
5. **Profiling and measuring** - Identifying actual bottlenecks

Apply these techniques based on your specific performance requirements and always validate improvements through benchmarking and profiling.
