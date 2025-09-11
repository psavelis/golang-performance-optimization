# Concurrency Optimization

Concurrency optimization in Go involves maximizing performance through effective use of goroutines, channels, and synchronization primitives while avoiding common pitfalls like race conditions, deadlocks, and excessive contention. This chapter explores advanced concurrency patterns, performance optimization techniques, and monitoring strategies.

## Goroutine Management and Optimization

### Advanced Goroutine Pooling
Implement sophisticated goroutine pools that adapt to workload characteristics:

```go
package concurrency_optimization

import (
    "context"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
)

// Adaptive worker pool with dynamic scaling
type AdaptiveWorkerPool struct {
    minWorkers    int32
    maxWorkers    int32
    currentWorkers int32
    activeWorkers  int32
    
    workQueue     chan WorkItem
    workerQueue   chan chan WorkItem
    quitChans     []chan bool
    
    metrics       PoolMetrics
    scaler        *PoolScaler
    
    mu sync.RWMutex
}

type WorkItem struct {
    ID       uint64
    Task     func() error
    Priority int
    Deadline time.Time
    Result   chan<- error
}

type PoolMetrics struct {
    TasksSubmitted   int64 `json:"tasks_submitted"`
    TasksCompleted   int64 `json:"tasks_completed"`
    TasksTimedOut    int64 `json:"tasks_timed_out"`
    TotalProcessTime int64 `json:"total_process_time_ns"`
    AverageWaitTime  int64 `json:"average_wait_time_ns"`
    QueueDepth       int64 `json:"queue_depth"`
    WorkerUtilization float64 `json:"worker_utilization"`
}

type PoolScaler struct {
    pool            *AdaptiveWorkerPool
    scaleUpThreshold   int
    scaleDownThreshold int
    scaleUpCooldown    time.Duration
    scaleDownCooldown  time.Duration
    lastScaleUp        time.Time
    lastScaleDown      time.Time
    enabled            bool
}

func NewAdaptiveWorkerPool(minWorkers, maxWorkers int, queueSize int) *AdaptiveWorkerPool {
    pool := &AdaptiveWorkerPool{
        minWorkers:  int32(minWorkers),
        maxWorkers:  int32(maxWorkers),
        workQueue:   make(chan WorkItem, queueSize),
        workerQueue: make(chan chan WorkItem, maxWorkers),
        quitChans:   make([]chan bool, 0, maxWorkers),
    }
    
    pool.scaler = &PoolScaler{
        pool:               pool,
        scaleUpThreshold:   queueSize / 2,
        scaleDownThreshold: queueSize / 10,
        scaleUpCooldown:    30 * time.Second,
        scaleDownCooldown:  60 * time.Second,
        enabled:            true,
    }
    
    // Start minimum workers
    for i := 0; i < minWorkers; i++ {
        pool.addWorker()
    }
    
    // Start scaling monitor
    go pool.scaler.monitor()
    
    return pool
}

func (pool *AdaptiveWorkerPool) Submit(task func() error, priority int, timeout time.Duration) error {
    atomic.AddInt64(&pool.metrics.TasksSubmitted, 1)
    
    deadline := time.Now().Add(timeout)
    result := make(chan error, 1)
    
    workItem := WorkItem{
        ID:       atomic.AddUint64(new(uint64), 1),
        Task:     task,
        Priority: priority,
        Deadline: deadline,
        Result:   result,
    }
    
    select {
    case pool.workQueue <- workItem:
        atomic.AddInt64(&pool.metrics.QueueDepth, 1)
        
        // Wait for result or timeout
        select {
        case err := <-result:
            return err
        case <-time.After(timeout):
            atomic.AddInt64(&pool.metrics.TasksTimedOut, 1)
            return fmt.Errorf("task timed out after %v", timeout)
        }
        
    case <-time.After(100 * time.Millisecond):
        return fmt.Errorf("work queue full, cannot submit task")
    }
}

func (pool *AdaptiveWorkerPool) addWorker() {
    currentWorkers := atomic.AddInt32(&pool.currentWorkers, 1)
    
    if currentWorkers > pool.maxWorkers {
        atomic.AddInt32(&pool.currentWorkers, -1)
        return
    }
    
    quit := make(chan bool)
    
    pool.mu.Lock()
    pool.quitChans = append(pool.quitChans, quit)
    pool.mu.Unlock()
    
    go pool.worker(quit)
}

func (pool *AdaptiveWorkerPool) removeWorker() {
    currentWorkers := atomic.LoadInt32(&pool.currentWorkers)
    
    if currentWorkers <= pool.minWorkers {
        return
    }
    
    pool.mu.Lock()
    if len(pool.quitChans) > 0 {
        quit := pool.quitChans[len(pool.quitChans)-1]
        pool.quitChans = pool.quitChans[:len(pool.quitChans)-1]
        
        select {
        case quit <- true:
            atomic.AddInt32(&pool.currentWorkers, -1)
        default:
            // Worker already quit
        }
    }
    pool.mu.Unlock()
}

func (pool *AdaptiveWorkerPool) worker(quit <-chan bool) {
    workerChan := make(chan WorkItem)
    
    for {
        // Register worker as available
        select {
        case pool.workerQueue <- workerChan:
            // Worker registered
        case <-quit:
            return
        }
        
        // Wait for work or quit signal
        select {
        case workItem := <-workerChan:
            pool.processWorkItem(workItem)
            
        case <-quit:
            return
        }
    }
}

func (pool *AdaptiveWorkerPool) processWorkItem(workItem WorkItem) {
    atomic.AddInt32(&pool.activeWorkers, 1)
    defer atomic.AddInt32(&pool.activeWorkers, -1)
    
    start := time.Now()
    
    // Check if task has already expired
    if time.Now().After(workItem.Deadline) {
        atomic.AddInt64(&pool.metrics.TasksTimedOut, 1)
        workItem.Result <- fmt.Errorf("task expired before processing")
        return
    }
    
    // Execute task with timeout
    done := make(chan error, 1)
    go func() {
        done <- workItem.Task()
    }()
    
    select {
    case err := <-done:
        workItem.Result <- err
        atomic.AddInt64(&pool.metrics.TasksCompleted, 1)
        
    case <-time.After(time.Until(workItem.Deadline)):
        atomic.AddInt64(&pool.metrics.TasksTimedOut, 1)
        workItem.Result <- fmt.Errorf("task execution timed out")
    }
    
    // Update metrics
    processingTime := time.Since(start).Nanoseconds()
    atomic.AddInt64(&pool.metrics.TotalProcessTime, processingTime)
    atomic.AddInt64(&pool.metrics.QueueDepth, -1)
}

func (scaler *PoolScaler) monitor() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if !scaler.enabled {
            continue
        }
        
        scaler.evaluateScaling()
    }
}

func (scaler *PoolScaler) evaluateScaling() {
    queueDepth := int(atomic.LoadInt64(&scaler.pool.metrics.QueueDepth))
    currentWorkers := atomic.LoadInt32(&scaler.pool.currentWorkers)
    activeWorkers := atomic.LoadInt32(&scaler.pool.activeWorkers)
    
    // Calculate utilization
    utilization := float64(activeWorkers) / float64(currentWorkers)
    atomic.StoreUint64((*uint64)(unsafe.Pointer(&scaler.pool.metrics.WorkerUtilization)), 
                      math.Float64bits(utilization))
    
    now := time.Now()
    
    // Scale up conditions
    if queueDepth > scaler.scaleUpThreshold && 
       utilization > 0.8 && 
       currentWorkers < scaler.pool.maxWorkers &&
       now.Sub(scaler.lastScaleUp) > scaler.scaleUpCooldown {
        
        scaleCount := min(int(scaler.pool.maxWorkers-currentWorkers), 
                         (queueDepth/scaler.scaleUpThreshold)+1)
        
        for i := 0; i < scaleCount; i++ {
            scaler.pool.addWorker()
        }
        
        scaler.lastScaleUp = now
        fmt.Printf("Scaled up by %d workers (total: %d)\n", 
                  scaleCount, atomic.LoadInt32(&scaler.pool.currentWorkers))
    }
    
    // Scale down conditions
    if queueDepth < scaler.scaleDownThreshold && 
       utilization < 0.3 && 
       currentWorkers > scaler.pool.minWorkers &&
       now.Sub(scaler.lastScaleDown) > scaler.scaleDownCooldown {
        
        scaleCount := min(int(currentWorkers-scaler.pool.minWorkers), 
                         int(currentWorkers/4)+1)
        
        for i := 0; i < scaleCount; i++ {
            scaler.pool.removeWorker()
        }
        
        scaler.lastScaleDown = now
        fmt.Printf("Scaled down by %d workers (total: %d)\n", 
                  scaleCount, atomic.LoadInt32(&scaler.pool.currentWorkers))
    }
}

func (pool *AdaptiveWorkerPool) GetMetrics() PoolMetrics {
    return PoolMetrics{
        TasksSubmitted:    atomic.LoadInt64(&pool.metrics.TasksSubmitted),
        TasksCompleted:    atomic.LoadInt64(&pool.metrics.TasksCompleted),
        TasksTimedOut:     atomic.LoadInt64(&pool.metrics.TasksTimedOut),
        TotalProcessTime:  atomic.LoadInt64(&pool.metrics.TotalProcessTime),
        QueueDepth:        atomic.LoadInt64(&pool.metrics.QueueDepth),
        WorkerUtilization: math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(&pool.metrics.WorkerUtilization)))),
    }
}

func (pool *AdaptiveWorkerPool) Shutdown(timeout time.Duration) {
    pool.scaler.enabled = false
    
    pool.mu.Lock()
    defer pool.mu.Unlock()
    
    // Signal all workers to quit
    for _, quit := range pool.quitChans {
        select {
        case quit <- true:
        case <-time.After(1 * time.Second):
            // Worker didn't respond to quit signal
        }
    }
    
    // Wait for graceful shutdown or timeout
    deadline := time.Now().Add(timeout)
    for atomic.LoadInt32(&pool.currentWorkers) > 0 && time.Now().Before(deadline) {
        time.Sleep(100 * time.Millisecond)
    }
}

// Work-stealing scheduler
type WorkStealingScheduler struct {
    workers     []*WorkStealingWorker
    globalQueue *LockFreeQueue
    workerCount int
    running     int32
}

type WorkStealingWorker struct {
    id          int
    localQueue  *LockFreeQueue
    scheduler   *WorkStealingScheduler
    running     int32
    stealAttempts int64
    stealSuccesses int64
}

type LockFreeQueue struct {
    head unsafe.Pointer // *QueueNode
    tail unsafe.Pointer // *QueueNode
    size int64
}

type QueueNode struct {
    data unsafe.Pointer // *WorkItem
    next unsafe.Pointer // *QueueNode
}

func NewWorkStealingScheduler(workerCount int) *WorkStealingScheduler {
    if workerCount <= 0 {
        workerCount = runtime.NumCPU()
    }
    
    scheduler := &WorkStealingScheduler{
        workers:     make([]*WorkStealingWorker, workerCount),
        globalQueue: NewLockFreeQueue(),
        workerCount: workerCount,
        running:     1,
    }
    
    // Initialize workers
    for i := 0; i < workerCount; i++ {
        scheduler.workers[i] = &WorkStealingWorker{
            id:          i,
            localQueue:  NewLockFreeQueue(),
            scheduler:   scheduler,
            running:     1,
        }
        
        go scheduler.workers[i].run()
    }
    
    return scheduler
}

func NewLockFreeQueue() *LockFreeQueue {
    dummy := &QueueNode{}
    queue := &LockFreeQueue{
        head: unsafe.Pointer(dummy),
        tail: unsafe.Pointer(dummy),
    }
    return queue
}

func (q *LockFreeQueue) Enqueue(item *WorkItem) {
    newNode := &QueueNode{
        data: unsafe.Pointer(item),
    }
    
    for {
        tail := (*QueueNode)(atomic.LoadPointer(&q.tail))
        next := (*QueueNode)(atomic.LoadPointer(&tail.next))
        
        if tail == (*QueueNode)(atomic.LoadPointer(&q.tail)) {
            if next == nil {
                if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(newNode)) {
                    atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(newNode))
                    atomic.AddInt64(&q.size, 1)
                    break
                }
            } else {
                atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            }
        }
    }
}

func (q *LockFreeQueue) Dequeue() *WorkItem {
    for {
        head := (*QueueNode)(atomic.LoadPointer(&q.head))
        tail := (*QueueNode)(atomic.LoadPointer(&q.tail))
        next := (*QueueNode)(atomic.LoadPointer(&head.next))
        
        if head == (*QueueNode)(atomic.LoadPointer(&q.head)) {
            if head == tail {
                if next == nil {
                    return nil // Queue is empty
                }
                atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            } else {
                if next == nil {
                    continue
                }
                
                data := (*WorkItem)(atomic.LoadPointer(&next.data))
                if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
                    atomic.AddInt64(&q.size, -1)
                    return data
                }
            }
        }
    }
}

func (q *LockFreeQueue) Size() int64 {
    return atomic.LoadInt64(&q.size)
}

func (scheduler *WorkStealingScheduler) Submit(task func() error) {
    workItem := &WorkItem{
        ID:   atomic.AddUint64(new(uint64), 1),
        Task: task,
    }
    
    // Try to add to a worker's local queue first
    workerID := int(workItem.ID) % scheduler.workerCount
    worker := scheduler.workers[workerID]
    
    if worker.localQueue.Size() < 100 { // Local queue threshold
        worker.localQueue.Enqueue(workItem)
    } else {
        // Local queue full, add to global queue
        scheduler.globalQueue.Enqueue(workItem)
    }
}

func (worker *WorkStealingWorker) run() {
    for atomic.LoadInt32(&worker.running) == 1 {
        var workItem *WorkItem
        
        // 1. Try local queue first
        workItem = worker.localQueue.Dequeue()
        
        // 2. Try global queue
        if workItem == nil {
            workItem = worker.scheduler.globalQueue.Dequeue()
        }
        
        // 3. Try work stealing
        if workItem == nil {
            workItem = worker.stealWork()
        }
        
        if workItem != nil {
            // Execute the work item
            workItem.Task()
        } else {
            // No work available, sleep briefly
            time.Sleep(1 * time.Millisecond)
        }
    }
}

func (worker *WorkStealingWorker) stealWork() *WorkItem {
    atomic.AddInt64(&worker.stealAttempts, 1)
    
    // Try to steal from other workers
    for i := 0; i < worker.scheduler.workerCount; i++ {
        if i == worker.id {
            continue
        }
        
        victim := worker.scheduler.workers[i]
        if victim.localQueue.Size() > 1 {
            workItem := victim.localQueue.Dequeue()
            if workItem != nil {
                atomic.AddInt64(&worker.stealSuccesses, 1)
                return workItem
            }
        }
    }
    
    return nil
}

func (scheduler *WorkStealingScheduler) Shutdown() {
    atomic.StoreInt32(&scheduler.running, 0)
    
    for _, worker := range scheduler.workers {
        atomic.StoreInt32(&worker.running, 0)
    }
}

func (scheduler *WorkStealingScheduler) GetStats() map[string]interface{} {
    stats := make(map[string]interface{})
    
    var totalStealAttempts, totalStealSuccesses int64
    for _, worker := range scheduler.workers {
        totalStealAttempts += atomic.LoadInt64(&worker.stealAttempts)
        totalStealSuccesses += atomic.LoadInt64(&worker.stealSuccesses)
    }
    
    stealRate := float64(0)
    if totalStealAttempts > 0 {
        stealRate = float64(totalStealSuccesses) / float64(totalStealAttempts)
    }
    
    stats["total_steal_attempts"] = totalStealAttempts
    stats["total_steal_successes"] = totalStealSuccesses
    stats["steal_success_rate"] = stealRate
    stats["global_queue_size"] = scheduler.globalQueue.Size()
    
    return stats
}
```

## Channel Optimization and Patterns

### High-Performance Channel Patterns
Implement optimized channel usage patterns for different scenarios:

```go
// Buffered channel with adaptive sizing
type AdaptiveChannel struct {
    ch          chan interface{}
    capacity    int
    utilization float64
    resizing    int32
    metrics     ChannelMetrics
    mu          sync.RWMutex
}

type ChannelMetrics struct {
    Sends       int64   `json:"sends"`
    Receives    int64   `json:"receives"`
    Blocks      int64   `json:"blocks"`
    Utilization float64 `json:"utilization"`
    Resizes     int64   `json:"resizes"`
}

func NewAdaptiveChannel(initialCapacity int) *AdaptiveChannel {
    return &AdaptiveChannel{
        ch:       make(chan interface{}, initialCapacity),
        capacity: initialCapacity,
    }
}

func (ac *AdaptiveChannel) Send(value interface{}) bool {
    atomic.AddInt64(&ac.metrics.Sends, 1)
    
    select {
    case ac.ch <- value:
        ac.updateUtilization()
        return true
        
    default:
        atomic.AddInt64(&ac.metrics.Blocks, 1)
        ac.considerResize()
        
        // Try again after potential resize
        select {
        case ac.ch <- value:
            return true
        case <-time.After(100 * time.Millisecond):
            return false
        }
    }
}

func (ac *AdaptiveChannel) Receive() (interface{}, bool) {
    atomic.AddInt64(&ac.metrics.Receives, 1)
    
    select {
    case value := <-ac.ch:
        ac.updateUtilization()
        return value, true
        
    case <-time.After(100 * time.Millisecond):
        return nil, false
    }
}

func (ac *AdaptiveChannel) updateUtilization() {
    currentLength := len(ac.ch)
    utilization := float64(currentLength) / float64(ac.capacity)
    
    // Exponential moving average
    alpha := 0.1
    ac.utilization = alpha*utilization + (1-alpha)*ac.utilization
    
    atomic.StoreUint64((*uint64)(unsafe.Pointer(&ac.metrics.Utilization)), 
                      math.Float64bits(ac.utilization))
}

func (ac *AdaptiveChannel) considerResize() {
    if atomic.LoadInt32(&ac.resizing) == 1 {
        return
    }
    
    // Resize if utilization is consistently high or low
    if ac.utilization > 0.8 && ac.capacity < 10000 {
        go ac.resize(ac.capacity * 2)
    } else if ac.utilization < 0.2 && ac.capacity > 10 {
        go ac.resize(ac.capacity / 2)
    }
}

func (ac *AdaptiveChannel) resize(newCapacity int) {
    if !atomic.CompareAndSwapInt32(&ac.resizing, 0, 1) {
        return
    }
    defer atomic.StoreInt32(&ac.resizing, 0)
    
    ac.mu.Lock()
    defer ac.mu.Unlock()
    
    // Create new channel
    newCh := make(chan interface{}, newCapacity)
    
    // Transfer existing items
    close(ac.ch)
    for item := range ac.ch {
        select {
        case newCh <- item:
        default:
            // New channel full, drop items
            break
        }
    }
    
    ac.ch = newCh
    ac.capacity = newCapacity
    atomic.AddInt64(&ac.metrics.Resizes, 1)
}

func (ac *AdaptiveChannel) GetMetrics() ChannelMetrics {
    return ChannelMetrics{
        Sends:       atomic.LoadInt64(&ac.metrics.Sends),
        Receives:    atomic.LoadInt64(&ac.metrics.Receives),
        Blocks:      atomic.LoadInt64(&ac.metrics.Blocks),
        Utilization: math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(&ac.metrics.Utilization)))),
        Resizes:     atomic.LoadInt64(&ac.metrics.Resizes),
    }
}

// Fan-out/Fan-in pattern with load balancing
type LoadBalancedFanOut struct {
    workers     []chan<- WorkUnit
    workerCount int
    selector    WorkerSelector
    metrics     FanOutMetrics
}

type WorkUnit struct {
    ID       uint64
    Data     interface{}
    Priority int
    Result   chan<- interface{}
}

type WorkerSelector interface {
    SelectWorker(workers []chan<- WorkUnit, workUnit WorkUnit) int
}

type FanOutMetrics struct {
    UnitsDispatched int64            `json:"units_dispatched"`
    WorkerLoads     map[int]int64    `json:"worker_loads"`
    mu              sync.RWMutex
}

// Round-robin selector
type RoundRobinSelector struct {
    counter uint64
}

func (rr *RoundRobinSelector) SelectWorker(workers []chan<- WorkUnit, workUnit WorkUnit) int {
    return int(atomic.AddUint64(&rr.counter, 1) % uint64(len(workers)))
}

// Least-loaded selector
type LeastLoadedSelector struct {
    loads []int64
}

func NewLeastLoadedSelector(workerCount int) *LeastLoadedSelector {
    return &LeastLoadedSelector{
        loads: make([]int64, workerCount),
    }
}

func (ll *LeastLoadedSelector) SelectWorker(workers []chan<- WorkUnit, workUnit WorkUnit) int {
    minLoad := atomic.LoadInt64(&ll.loads[0])
    minIndex := 0
    
    for i := 1; i < len(ll.loads); i++ {
        load := atomic.LoadInt64(&ll.loads[i])
        if load < minLoad {
            minLoad = load
            minIndex = i
        }
    }
    
    atomic.AddInt64(&ll.loads[minIndex], 1)
    return minIndex
}

func (ll *LeastLoadedSelector) DecrementLoad(workerIndex int) {
    atomic.AddInt64(&ll.loads[workerIndex], -1)
}

// Priority-based selector
type PrioritySelector struct {
    highPriorityWorkers []int
    normalWorkers       []int
    counter             uint64
}

func NewPrioritySelector(totalWorkers int, highPriorityCount int) *PrioritySelector {
    ps := &PrioritySelector{
        highPriorityWorkers: make([]int, highPriorityCount),
        normalWorkers:       make([]int, totalWorkers-highPriorityCount),
    }
    
    for i := 0; i < highPriorityCount; i++ {
        ps.highPriorityWorkers[i] = i
    }
    
    for i := 0; i < totalWorkers-highPriorityCount; i++ {
        ps.normalWorkers[i] = highPriorityCount + i
    }
    
    return ps
}

func (ps *PrioritySelector) SelectWorker(workers []chan<- WorkUnit, workUnit WorkUnit) int {
    if workUnit.Priority > 5 && len(ps.highPriorityWorkers) > 0 {
        // High priority work goes to dedicated workers
        index := atomic.AddUint64(&ps.counter, 1) % uint64(len(ps.highPriorityWorkers))
        return ps.highPriorityWorkers[index]
    }
    
    // Normal priority work
    index := atomic.AddUint64(&ps.counter, 1) % uint64(len(ps.normalWorkers))
    return ps.normalWorkers[index]
}

func NewLoadBalancedFanOut(workerCount int, selector WorkerSelector) *LoadBalancedFanOut {
    workers := make([]chan<- WorkUnit, workerCount)
    
    fanOut := &LoadBalancedFanOut{
        workers:     workers,
        workerCount: workerCount,
        selector:    selector,
        metrics: FanOutMetrics{
            WorkerLoads: make(map[int]int64),
        },
    }
    
    // Initialize workers
    for i := 0; i < workerCount; i++ {
        workerChan := make(chan WorkUnit, 100)
        workers[i] = workerChan
        fanOut.metrics.WorkerLoads[i] = 0
        
        go fanOut.worker(i, workerChan)
    }
    
    return fanOut
}

func (lbf *LoadBalancedFanOut) Submit(workUnit WorkUnit) error {
    workerIndex := lbf.selector.SelectWorker(lbf.workers, workUnit)
    
    select {
    case lbf.workers[workerIndex] <- workUnit:
        atomic.AddInt64(&lbf.metrics.UnitsDispatched, 1)
        
        lbf.metrics.mu.Lock()
        lbf.metrics.WorkerLoads[workerIndex]++
        lbf.metrics.mu.Unlock()
        
        return nil
        
    case <-time.After(1 * time.Second):
        return fmt.Errorf("worker %d queue full", workerIndex)
    }
}

func (lbf *LoadBalancedFanOut) worker(workerID int, workChan <-chan WorkUnit) {
    for workUnit := range workChan {
        // Process work unit
        result := lbf.processWork(workUnit)
        
        // Send result if channel provided
        if workUnit.Result != nil {
            select {
            case workUnit.Result <- result:
            case <-time.After(1 * time.Second):
                // Result channel blocked or closed
            }
        }
        
        // Update metrics
        lbf.metrics.mu.Lock()
        lbf.metrics.WorkerLoads[workerID]--
        lbf.metrics.mu.Unlock()
        
        // Notify selector if it tracks load
        if ll, ok := lbf.selector.(*LeastLoadedSelector); ok {
            ll.DecrementLoad(workerID)
        }
    }
}

func (lbf *LoadBalancedFanOut) processWork(workUnit WorkUnit) interface{} {
    // Simulate work processing
    time.Sleep(time.Duration(workUnit.Priority) * time.Millisecond)
    return fmt.Sprintf("Processed work unit %d", workUnit.ID)
}

func (lbf *LoadBalancedFanOut) GetMetrics() FanOutMetrics {
    lbf.metrics.mu.RLock()
    defer lbf.metrics.mu.RUnlock()
    
    // Create a copy of the metrics
    loads := make(map[int]int64)
    for k, v := range lbf.metrics.WorkerLoads {
        loads[k] = v
    }
    
    return FanOutMetrics{
        UnitsDispatched: atomic.LoadInt64(&lbf.metrics.UnitsDispatched),
        WorkerLoads:     loads,
    }
}
```

## Lock-Free Data Structures

### Advanced Lock-Free Implementations
Implement sophisticated lock-free data structures for high-concurrency scenarios:

```go
// Lock-free hash map with linear probing
type LockFreeHashMap struct {
    buckets []unsafe.Pointer // Array of *HashBucket
    size    int64
    mask    uint64
}

type HashBucket struct {
    key   unsafe.Pointer // *string
    value unsafe.Pointer // *interface{}
    hash  uint64
    state int32 // 0: empty, 1: occupied, 2: deleted
}

const (
    BucketEmpty = iota
    BucketOccupied
    BucketDeleted
)

func NewLockFreeHashMap(capacity int) *LockFreeHashMap {
    // Round up to power of 2
    cap := 1
    for cap < capacity {
        cap <<= 1
    }
    
    hm := &LockFreeHashMap{
        buckets: make([]unsafe.Pointer, cap),
        mask:    uint64(cap - 1),
    }
    
    // Initialize buckets
    for i := range hm.buckets {
        bucket := &HashBucket{}
        hm.buckets[i] = unsafe.Pointer(bucket)
    }
    
    return hm
}

func (hm *LockFreeHashMap) Put(key string, value interface{}) bool {
    hash := hm.hash(key)
    
    for probe := uint64(0); probe < uint64(len(hm.buckets)); probe++ {
        index := (hash + probe) & hm.mask
        bucket := (*HashBucket)(atomic.LoadPointer(&hm.buckets[index]))
        
        for {
            state := atomic.LoadInt32(&bucket.state)
            
            switch state {
            case BucketEmpty:
                // Try to claim empty bucket
                if atomic.CompareAndSwapInt32(&bucket.state, BucketEmpty, BucketOccupied) {
                    atomic.StorePointer(&bucket.key, unsafe.Pointer(&key))
                    atomic.StorePointer(&bucket.value, unsafe.Pointer(&value))
                    atomic.StoreUint64(&bucket.hash, hash)
                    atomic.AddInt64(&hm.size, 1)
                    return true
                }
                
            case BucketOccupied:
                if bucket.hash == hash {
                    bucketKey := (*string)(atomic.LoadPointer(&bucket.key))
                    if *bucketKey == key {
                        // Update existing key
                        atomic.StorePointer(&bucket.value, unsafe.Pointer(&value))
                        return false // Key already existed
                    }
                }
                // Key doesn't match, continue probing
                break
                
            case BucketDeleted:
                // Try to reuse deleted bucket
                if atomic.CompareAndSwapInt32(&bucket.state, BucketDeleted, BucketOccupied) {
                    atomic.StorePointer(&bucket.key, unsafe.Pointer(&key))
                    atomic.StorePointer(&bucket.value, unsafe.Pointer(&value))
                    atomic.StoreUint64(&bucket.hash, hash)
                    atomic.AddInt64(&hm.size, 1)
                    return true
                }
            }
            
            // Bucket state changed, retry
            if atomic.LoadInt32(&bucket.state) != state {
                continue
            }
            
            break
        }
    }
    
    return false // Hash map full
}

func (hm *LockFreeHashMap) Get(key string) (interface{}, bool) {
    hash := hm.hash(key)
    
    for probe := uint64(0); probe < uint64(len(hm.buckets)); probe++ {
        index := (hash + probe) & hm.mask
        bucket := (*HashBucket)(atomic.LoadPointer(&hm.buckets[index]))
        
        state := atomic.LoadInt32(&bucket.state)
        
        if state == BucketEmpty {
            return nil, false
        }
        
        if state == BucketOccupied && bucket.hash == hash {
            bucketKey := (*string)(atomic.LoadPointer(&bucket.key))
            if *bucketKey == key {
                value := (*interface{})(atomic.LoadPointer(&bucket.value))
                return *value, true
            }
        }
    }
    
    return nil, false
}

func (hm *LockFreeHashMap) Delete(key string) bool {
    hash := hm.hash(key)
    
    for probe := uint64(0); probe < uint64(len(hm.buckets)); probe++ {
        index := (hash + probe) & hm.mask
        bucket := (*HashBucket)(atomic.LoadPointer(&hm.buckets[index]))
        
        state := atomic.LoadInt32(&bucket.state)
        
        if state == BucketEmpty {
            return false
        }
        
        if state == BucketOccupied && bucket.hash == hash {
            bucketKey := (*string)(atomic.LoadPointer(&bucket.key))
            if *bucketKey == key {
                if atomic.CompareAndSwapInt32(&bucket.state, BucketOccupied, BucketDeleted) {
                    atomic.AddInt64(&hm.size, -1)
                    return true
                }
            }
        }
    }
    
    return false
}

func (hm *LockFreeHashMap) Size() int64 {
    return atomic.LoadInt64(&hm.size)
}

func (hm *LockFreeHashMap) hash(key string) uint64 {
    // FNV-1a hash
    hash := uint64(14695981039346656037)
    for _, b := range []byte(key) {
        hash ^= uint64(b)
        hash *= 1099511628211
    }
    return hash
}

// Lock-free stack
type LockFreeStack struct {
    head unsafe.Pointer // *StackNode
    size int64
}

type StackNode struct {
    data interface{}
    next unsafe.Pointer // *StackNode
}

func NewLockFreeStack() *LockFreeStack {
    return &LockFreeStack{}
}

func (s *LockFreeStack) Push(data interface{}) {
    newNode := &StackNode{data: data}
    
    for {
        head := atomic.LoadPointer(&s.head)
        newNode.next = head
        
        if atomic.CompareAndSwapPointer(&s.head, head, unsafe.Pointer(newNode)) {
            atomic.AddInt64(&s.size, 1)
            break
        }
    }
}

func (s *LockFreeStack) Pop() (interface{}, bool) {
    for {
        head := atomic.LoadPointer(&s.head)
        if head == nil {
            return nil, false
        }
        
        node := (*StackNode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&s.head, head, next) {
            atomic.AddInt64(&s.size, -1)
            return node.data, true
        }
    }
}

func (s *LockFreeStack) Size() int64 {
    return atomic.LoadInt64(&s.size)
}

func (s *LockFreeStack) IsEmpty() bool {
    return atomic.LoadPointer(&s.head) == nil
}

// Lock-free ring buffer
type LockFreeRingBuffer struct {
    buffer []unsafe.Pointer
    head   int64
    tail   int64
    mask   int64
}

func NewLockFreeRingBuffer(capacity int) *LockFreeRingBuffer {
    // Round up to power of 2
    cap := 1
    for cap < capacity {
        cap <<= 1
    }
    
    return &LockFreeRingBuffer{
        buffer: make([]unsafe.Pointer, cap),
        mask:   int64(cap - 1),
    }
}

func (rb *LockFreeRingBuffer) Put(item interface{}) bool {
    for {
        head := atomic.LoadInt64(&rb.head)
        tail := atomic.LoadInt64(&rb.tail)
        
        if head-tail >= int64(len(rb.buffer)) {
            return false // Buffer full
        }
        
        index := head & rb.mask
        
        if atomic.CompareAndSwapPointer(&rb.buffer[index], nil, unsafe.Pointer(&item)) {
            atomic.AddInt64(&rb.head, 1)
            return true
        }
    }
}

func (rb *LockFreeRingBuffer) Get() (interface{}, bool) {
    for {
        tail := atomic.LoadInt64(&rb.tail)
        head := atomic.LoadInt64(&rb.head)
        
        if tail >= head {
            return nil, false // Buffer empty
        }
        
        index := tail & rb.mask
        item := atomic.LoadPointer(&rb.buffer[index])
        
        if item != nil {
            if atomic.CompareAndSwapPointer(&rb.buffer[index], item, nil) {
                atomic.AddInt64(&rb.tail, 1)
                return *(*interface{})(item), true
            }
        }
    }
}

func (rb *LockFreeRingBuffer) Size() int64 {
    head := atomic.LoadInt64(&rb.head)
    tail := atomic.LoadInt64(&rb.tail)
    return head - tail
}

func (rb *LockFreeRingBuffer) Capacity() int {
    return len(rb.buffer)
}
```

## Deadlock Detection and Prevention

### Advanced Deadlock Prevention
Implement comprehensive deadlock detection and prevention mechanisms:

```go
package deadlock_prevention

import (
    "fmt"
    "sync"
    "time"
)

// Deadlock detector using resource allocation graph
type DeadlockDetector struct {
    resources map[string]*Resource
    processes map[string]*Process
    graph     *AllocationGraph
    mu        sync.RWMutex
    enabled   bool
}

type Resource struct {
    ID        string
    Owner     string
    Waiters   []string
    Timestamp time.Time
    mu        sync.Mutex
}

type Process struct {
    ID        string
    Holding   []string
    Waiting   []string
    Timestamp time.Time
}

type AllocationGraph struct {
    edges map[string][]string // process -> resources
    mu    sync.RWMutex
}

func NewDeadlockDetector() *DeadlockDetector {
    return &DeadlockDetector{
        resources: make(map[string]*Resource),
        processes: make(map[string]*Process),
        graph:     &AllocationGraph{edges: make(map[string][]string)},
        enabled:   true,
    }
}

func (dd *DeadlockDetector) AcquireResource(processID, resourceID string, timeout time.Duration) error {
    if !dd.enabled {
        return nil
    }
    
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    // Create process if not exists
    if _, exists := dd.processes[processID]; !exists {
        dd.processes[processID] = &Process{
            ID:        processID,
            Holding:   make([]string, 0),
            Waiting:   make([]string, 0),
            Timestamp: time.Now(),
        }
    }
    
    // Create resource if not exists
    if _, exists := dd.resources[resourceID]; !exists {
        dd.resources[resourceID] = &Resource{
            ID:        resourceID,
            Waiters:   make([]string, 0),
            Timestamp: time.Now(),
        }
    }
    
    resource := dd.resources[resourceID]
    process := dd.processes[processID]
    
    // Check if resource is available
    if resource.Owner == "" {
        // Acquire immediately
        resource.Owner = processID
        resource.Timestamp = time.Now()
        process.Holding = append(process.Holding, resourceID)
        dd.graph.addEdge(processID, resourceID)
        return nil
    }
    
    // Resource is busy, check for potential deadlock
    if dd.wouldCauseDeadlock(processID, resourceID) {
        return fmt.Errorf("acquiring resource %s would cause deadlock for process %s", resourceID, processID)
    }
    
    // Add to waiters
    resource.Waiters = append(resource.Waiters, processID)
    process.Waiting = append(process.Waiting, resourceID)
    
    // Wait for resource with timeout
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        dd.mu.Unlock()
        time.Sleep(10 * time.Millisecond)
        dd.mu.Lock()
        
        if resource.Owner == processID {
            // Successfully acquired
            process.Holding = append(process.Holding, resourceID)
            dd.removeFromWaiters(process, resourceID)
            dd.graph.addEdge(processID, resourceID)
            return nil
        }
    }
    
    // Timeout
    dd.removeFromWaiters(process, resourceID)
    return fmt.Errorf("timeout acquiring resource %s for process %s", resourceID, processID)
}

func (dd *DeadlockDetector) ReleaseResource(processID, resourceID string) error {
    if !dd.enabled {
        return nil
    }
    
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    resource, exists := dd.resources[resourceID]
    if !exists {
        return fmt.Errorf("resource %s does not exist", resourceID)
    }
    
    if resource.Owner != processID {
        return fmt.Errorf("process %s does not own resource %s", processID, resourceID)
    }
    
    // Release resource
    resource.Owner = ""
    resource.Timestamp = time.Now()
    dd.graph.removeEdge(processID, resourceID)
    
    // Remove from process holdings
    process := dd.processes[processID]
    process.Holding = dd.removeFromSlice(process.Holding, resourceID)
    
    // Assign to next waiter
    if len(resource.Waiters) > 0 {
        nextOwner := resource.Waiters[0]
        resource.Owner = nextOwner
        resource.Waiters = resource.Waiters[1:]
        
        nextProcess := dd.processes[nextOwner]
        nextProcess.Holding = append(nextProcess.Holding, resourceID)
        nextProcess.Waiting = dd.removeFromSlice(nextProcess.Waiting, resourceID)
        dd.graph.addEdge(nextOwner, resourceID)
    }
    
    return nil
}

func (dd *DeadlockDetector) wouldCauseDeadlock(processID, resourceID string) bool {
    // Check if adding this wait edge would create a cycle
    resourceOwner := dd.resources[resourceID].Owner
    if resourceOwner == "" {
        return false
    }
    
    // Would create edge: processID -> resourceOwner
    // Check if resourceOwner can reach processID
    return dd.graph.hasPath(resourceOwner, processID)
}

func (dd *DeadlockDetector) removeFromWaiters(process *Process, resourceID string) {
    resource := dd.resources[resourceID]
    for i, waiter := range resource.Waiters {
        if waiter == process.ID {
            resource.Waiters = append(resource.Waiters[:i], resource.Waiters[i+1:]...)
            break
        }
    }
    process.Waiting = dd.removeFromSlice(process.Waiting, resourceID)
}

func (dd *DeadlockDetector) removeFromSlice(slice []string, item string) []string {
    for i, v := range slice {
        if v == item {
            return append(slice[:i], slice[i+1:]...)
        }
    }
    return slice
}

func (ag *AllocationGraph) addEdge(from, to string) {
    ag.mu.Lock()
    defer ag.mu.Unlock()
    
    if ag.edges[from] == nil {
        ag.edges[from] = make([]string, 0)
    }
    
    // Check if edge already exists
    for _, edge := range ag.edges[from] {
        if edge == to {
            return
        }
    }
    
    ag.edges[from] = append(ag.edges[from], to)
}

func (ag *AllocationGraph) removeEdge(from, to string) {
    ag.mu.Lock()
    defer ag.mu.Unlock()
    
    edges := ag.edges[from]
    for i, edge := range edges {
        if edge == to {
            ag.edges[from] = append(edges[:i], edges[i+1:]...)
            break
        }
    }
    
    if len(ag.edges[from]) == 0 {
        delete(ag.edges, from)
    }
}

func (ag *AllocationGraph) hasPath(from, to string) bool {
    ag.mu.RLock()
    defer ag.mu.RUnlock()
    
    visited := make(map[string]bool)
    return ag.dfs(from, to, visited)
}

func (ag *AllocationGraph) dfs(current, target string, visited map[string]bool) bool {
    if current == target {
        return true
    }
    
    if visited[current] {
        return false
    }
    
    visited[current] = true
    
    for _, neighbor := range ag.edges[current] {
        if ag.dfs(neighbor, target, visited) {
            return true
        }
    }
    
    return false
}

// Ordered lock acquisition to prevent deadlocks
type OrderedLockManager struct {
    locks   map[string]*OrderedMutex
    order   map[string]int
    counter int64
    mu      sync.RWMutex
}

type OrderedMutex struct {
    mu    sync.Mutex
    id    string
    order int
}

func NewOrderedLockManager() *OrderedLockManager {
    return &OrderedLockManager{
        locks: make(map[string]*OrderedMutex),
        order: make(map[string]int),
    }
}

func (olm *OrderedLockManager) GetLock(id string) *OrderedMutex {
    olm.mu.Lock()
    defer olm.mu.Unlock()
    
    if lock, exists := olm.locks[id]; exists {
        return lock
    }
    
    // Create new lock with ordering
    order := int(atomic.AddInt64(&olm.counter, 1))
    lock := &OrderedMutex{
        id:    id,
        order: order,
    }
    
    olm.locks[id] = lock
    olm.order[id] = order
    
    return lock
}

func (olm *OrderedLockManager) AcquireMultiple(lockIDs []string) []*OrderedMutex {
    // Sort locks by order to ensure consistent acquisition order
    locks := make([]*OrderedMutex, len(lockIDs))
    for i, id := range lockIDs {
        locks[i] = olm.GetLock(id)
    }
    
    // Sort by order
    sort.Slice(locks, func(i, j int) bool {
        return locks[i].order < locks[j].order
    })
    
    // Acquire in order
    for _, lock := range locks {
        lock.mu.Lock()
    }
    
    return locks
}

func (olm *OrderedLockManager) ReleaseMultiple(locks []*OrderedMutex) {
    // Release in reverse order
    for i := len(locks) - 1; i >= 0; i-- {
        locks[i].mu.Unlock()
    }
}

// Timeout-based lock to prevent infinite waiting
type TimeoutMutex struct {
    ch chan struct{}
}

func NewTimeoutMutex() *TimeoutMutex {
    return &TimeoutMutex{
        ch: make(chan struct{}, 1),
    }
}

func (tm *TimeoutMutex) Lock(timeout time.Duration) error {
    select {
    case tm.ch <- struct{}{}:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("lock acquisition timed out after %v", timeout)
    }
}

func (tm *TimeoutMutex) Unlock() {
    select {
    case <-tm.ch:
    default:
        panic("unlock of unlocked mutex")
    }
}

func (tm *TimeoutMutex) TryLock() bool {
    select {
    case tm.ch <- struct{}{}:
        return true
    default:
        return false
    }
}
```

Concurrency optimization requires careful balance between performance, correctness, and maintainability. By implementing adaptive pooling strategies, lock-free data structures, and deadlock prevention mechanisms, Go applications can achieve high scalability while maintaining safety guarantees.

## Key Takeaways

1. **Implement adaptive pooling** - adjust worker pool size based on workload
2. **Use work-stealing schedulers** - distribute work efficiently across cores
3. **Optimize channel usage** - choose appropriate buffer sizes and patterns
4. **Design lock-free structures** - eliminate contention for high-performance paths
5. **Prevent deadlocks systematically** - use ordered acquisition and detection
6. **Monitor concurrency metrics** - track utilization and contention
7. **Balance safety and performance** - avoid over-synchronization
8. **Test under load** - validate concurrency behavior with realistic workloads

Effective concurrency optimization enables Go applications to scale linearly with available CPU cores while maintaining correctness and avoiding common concurrency pitfalls.
