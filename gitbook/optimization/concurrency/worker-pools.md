# Worker Pool Patterns

Advanced worker pool implementations for optimal resource utilization and performance in Go applications. This guide covers sophisticated worker pool patterns, from basic implementations to enterprise-grade solutions.

## Table of Contents

- [Introduction](#introduction)
- [Basic Worker Pool](#basic-worker-pool)
- [Dynamic Scaling Pools](#dynamic-scaling-pools)
- [Priority-Based Pools](#priority-based-pools)
- [Specialized Pool Types](#specialized-pool-types)
- [Pool Management](#pool-management)
- [Performance Optimization](#performance-optimization)
- [Monitoring and Metrics](#monitoring-and-metrics)
- [Best Practices](#best-practices)

## Introduction

Worker pools manage a fixed number of goroutines to process work items efficiently, providing controlled resource usage and optimal throughput for concurrent applications.

### Core Components

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
)

// WorkerPool defines the interface for worker pool implementations
type WorkerPool interface {
    Submit(work Work) error
    SubmitWithTimeout(work Work, timeout time.Duration) error
    Close() error
    GetMetrics() PoolMetrics
    Resize(newSize int) error
}

// Work represents a unit of work to be processed
type Work struct {
    ID       string
    Task     func() error
    Priority int
    Deadline time.Time
    Context  context.Context
    Result   chan error
}

// PoolMetrics provides comprehensive metrics for worker pools
type PoolMetrics struct {
    PoolSize          int32
    ActiveWorkers     int32
    QueueSize         int32
    QueueCapacity     int32
    TasksSubmitted    int64
    TasksCompleted    int64
    TasksFailed       int64
    AverageLatency    time.Duration
    ThroughputPerSec  float64
    QueueUtilization  float64
    WorkerUtilization float64
}

// WorkerPoolConfig contains configuration for worker pools
type WorkerPoolConfig struct {
    MinWorkers       int
    MaxWorkers       int
    QueueSize        int
    IdleTimeout      time.Duration
    MaxTaskDuration  time.Duration
    EnableMetrics    bool
    EnableProfiling  bool
}
```

## Basic Worker Pool

Implementation of a fundamental worker pool with essential features.

### Simple Worker Pool

```go
// SimpleWorkerPool implements a basic worker pool
type SimpleWorkerPool struct {
    workers     []*Worker
    workQueue   chan Work
    quit        chan struct{}
    wg          sync.WaitGroup
    config      WorkerPoolConfig
    metrics     *PoolMetrics
    state       int32 // 0: stopped, 1: running, 2: closing
}

// Worker represents a single worker goroutine
type Worker struct {
    id          int
    workQueue   chan Work
    quit        chan struct{}
    pool        *SimpleWorkerPool
    currentWork *Work
    startTime   time.Time
    tasksDone   int64
}

// NewSimpleWorkerPool creates a new simple worker pool
func NewSimpleWorkerPool(config WorkerPoolConfig) *SimpleWorkerPool {
    pool := &SimpleWorkerPool{
        workQueue: make(chan Work, config.QueueSize),
        quit:      make(chan struct{}),
        config:    config,
        metrics:   &PoolMetrics{QueueCapacity: int32(config.QueueSize)},
    }
    
    // Create workers
    pool.workers = make([]*Worker, config.MinWorkers)
    for i := 0; i < config.MinWorkers; i++ {
        pool.workers[i] = &Worker{
            id:        i,
            workQueue: pool.workQueue,
            quit:      pool.quit,
            pool:      pool,
            startTime: time.Now(),
        }
    }
    
    atomic.StoreInt32(&pool.metrics.PoolSize, int32(config.MinWorkers))
    return pool
}

// Start starts the worker pool
func (swp *SimpleWorkerPool) Start() error {
    if !atomic.CompareAndSwapInt32(&swp.state, 0, 1) {
        return fmt.Errorf("pool already running")
    }
    
    // Start all workers
    for _, worker := range swp.workers {
        swp.wg.Add(1)
        go worker.run()
    }
    
    // Start metrics collector
    if swp.config.EnableMetrics {
        go swp.metricsCollector()
    }
    
    return nil
}

// Submit submits work to the pool
func (swp *SimpleWorkerPool) Submit(work Work) error {
    if atomic.LoadInt32(&swp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    select {
    case swp.workQueue <- work:
        atomic.AddInt64(&swp.metrics.TasksSubmitted, 1)
        atomic.StoreInt32(&swp.metrics.QueueSize, int32(len(swp.workQueue)))
        return nil
    default:
        return fmt.Errorf("work queue full")
    }
}

// run executes the worker main loop
func (w *Worker) run() {
    defer w.pool.wg.Done()
    
    for {
        select {
        case work := <-w.workQueue:
            w.processWork(work)
        case <-w.quit:
            return
        }
    }
}

// processWork processes a single work item
func (w *Worker) processWork(work Work) {
    atomic.AddInt32(&w.pool.metrics.ActiveWorkers, 1)
    defer atomic.AddInt32(&w.pool.metrics.ActiveWorkers, -1)
    
    w.currentWork = &work
    startTime := time.Now()
    
    // Execute work with timeout protection
    err := w.executeWithTimeout(work)
    
    duration := time.Since(startTime)
    w.updateMetrics(err, duration)
    
    w.currentWork = nil
    atomic.AddInt64(&w.tasksDone, 1)
    
    // Send result if channel provided
    if work.Result != nil {
        select {
        case work.Result <- err:
        default:
            // Non-blocking send
        }
    }
}

// executeWithTimeout executes work with timeout protection
func (w *Worker) executeWithTimeout(work Work) error {
    done := make(chan error, 1)
    
    go func() {
        defer func() {
            if r := recover(); r != nil {
                done <- fmt.Errorf("panic: %v", r)
            }
        }()
        
        done <- work.Task()
    }()
    
    timeout := w.pool.config.MaxTaskDuration
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    
    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        return fmt.Errorf("task timeout")
    case <-work.Context.Done():
        return work.Context.Err()
    }
}

// Close gracefully shuts down the worker pool
func (swp *SimpleWorkerPool) Close() error {
    if !atomic.CompareAndSwapInt32(&swp.state, 1, 2) {
        return fmt.Errorf("pool not running")
    }
    
    close(swp.workQueue)
    swp.wg.Wait()
    close(swp.quit)
    
    atomic.StoreInt32(&swp.state, 0)
    return nil
}

// GetMetrics returns current pool metrics
func (swp *SimpleWorkerPool) GetMetrics() PoolMetrics {
    metrics := *swp.metrics
    metrics.QueueSize = int32(len(swp.workQueue))
    metrics.QueueUtilization = float64(metrics.QueueSize) / float64(swp.config.QueueSize)
    return metrics
}

// Resize changes the pool size (not implemented in simple pool)
func (swp *SimpleWorkerPool) Resize(newSize int) error {
    return fmt.Errorf("resize not supported in simple worker pool")
}

// SubmitWithTimeout submits work with timeout
func (swp *SimpleWorkerPool) SubmitWithTimeout(work Work, timeout time.Duration) error {
    if atomic.LoadInt32(&swp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    select {
    case swp.workQueue <- work:
        atomic.AddInt64(&swp.metrics.TasksSubmitted, 1)
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("submit timeout")
    }
}

// updateMetrics updates worker and pool metrics
func (w *Worker) updateMetrics(err error, duration time.Duration) {
    if err != nil {
        atomic.AddInt64(&w.pool.metrics.TasksFailed, 1)
    } else {
        atomic.AddInt64(&w.pool.metrics.TasksCompleted, 1)
    }
    
    // Update average latency
    currentAvg := w.pool.metrics.AverageLatency
    newAvg := time.Duration((currentAvg.Nanoseconds() + duration.Nanoseconds()) / 2)
    w.pool.metrics.AverageLatency = newAvg
}

// metricsCollector periodically updates metrics
func (swp *SimpleWorkerPool) metricsCollector() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    lastCompleted := int64(0)
    
    for {
        select {
        case <-ticker.C:
            completed := atomic.LoadInt64(&swp.metrics.TasksCompleted)
            throughput := float64(completed - lastCompleted)
            swp.metrics.ThroughputPerSec = throughput
            lastCompleted = completed
            
        case <-swp.quit:
            return
        }
    }
}
```

## Priority-Based Pools

Worker pools that handle tasks based on priority levels.

### Priority Queue Pool

```go
// PriorityWorkerPool implements a priority-based worker pool
type PriorityWorkerPool struct {
    priorityQueues map[int]chan Work // Priority level -> queue
    workers        []*PriorityWorker
    quit           chan struct{}
    wg             sync.WaitGroup
    config         WorkerPoolConfig
    metrics        *PriorityPoolMetrics
    state          int32
    scheduler      *PriorityScheduler
}

// PriorityPoolMetrics extends PoolMetrics with priority-specific metrics
type PriorityPoolMetrics struct {
    PoolMetrics
    QueueSizesByPriority   map[int]int32
    TasksByPriority        map[int]int64
    LatencyByPriority      map[int]time.Duration
    StarvationEvents       int64
    PriorityInversions     int64
}

// PriorityWorker handles work items based on priority
type PriorityWorker struct {
    id            int
    pool          *PriorityWorkerPool
    currentWork   *Work
    tasksDone     int64
    lastWorkTime  time.Time
}

// PriorityScheduler manages priority-based work distribution
type PriorityScheduler struct {
    pool                *PriorityWorkerPool
    starvationThreshold time.Duration
    lastScheduleTime    map[int]time.Time
    priorityWeights     map[int]float64
}

// NewPriorityWorkerPool creates a new priority-based worker pool
func NewPriorityWorkerPool(config WorkerPoolConfig, priorities []int) *PriorityWorkerPool {
    pool := &PriorityWorkerPool{
        priorityQueues: make(map[int]chan Work),
        quit:          make(chan struct{}),
        config:        config,
        metrics: &PriorityPoolMetrics{
            PoolMetrics: PoolMetrics{
                QueueCapacity: int32(config.QueueSize),
            },
            QueueSizesByPriority: make(map[int]int32),
            TasksByPriority:      make(map[int]int64),
            LatencyByPriority:    make(map[int]time.Duration),
        },
    }
    
    // Create priority queues
    queueSizePerPriority := config.QueueSize / len(priorities)
    for _, priority := range priorities {
        pool.priorityQueues[priority] = make(chan Work, queueSizePerPriority)
        pool.metrics.QueueSizesByPriority[priority] = 0
        pool.metrics.TasksByPriority[priority] = 0
        pool.metrics.LatencyByPriority[priority] = 0
    }
    
    // Create workers
    pool.workers = make([]*PriorityWorker, config.MinWorkers)
    for i := 0; i < config.MinWorkers; i++ {
        pool.workers[i] = &PriorityWorker{
            id:   i,
            pool: pool,
        }
    }
    
    // Create scheduler
    pool.scheduler = &PriorityScheduler{
        pool:                pool,
        starvationThreshold: 5 * time.Second,
        lastScheduleTime:    make(map[int]time.Time),
        priorityWeights:     make(map[int]float64),
    }
    
    // Initialize priority weights (higher priority = higher weight)
    for _, priority := range priorities {
        pool.scheduler.priorityWeights[priority] = float64(priority + 10) // Avoid zero weights
    }
    
    atomic.StoreInt32(&pool.metrics.PoolSize, int32(config.MinWorkers))
    return pool
}

// Start starts the priority worker pool
func (pwp *PriorityWorkerPool) Start() error {
    if !atomic.CompareAndSwapInt32(&pwp.state, 0, 1) {
        return fmt.Errorf("pool already running")
    }
    
    // Start workers
    for _, worker := range pwp.workers {
        pwp.wg.Add(1)
        go worker.run()
    }
    
    // Start metrics collector
    if pwp.config.EnableMetrics {
        go pwp.metricsCollector()
    }
    
    return nil
}

// Submit submits work to the appropriate priority queue
func (pwp *PriorityWorkerPool) Submit(work Work) error {
    if atomic.LoadInt32(&pwp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    queue, exists := pwp.priorityQueues[work.Priority]
    if !exists {
        return fmt.Errorf("priority %d not supported", work.Priority)
    }
    
    select {
    case queue <- work:
        atomic.AddInt64(&pwp.metrics.TasksSubmitted, 1)
        atomic.AddInt64(&pwp.metrics.TasksByPriority[work.Priority], 1)
        pwp.updateQueueMetrics()
        return nil
    default:
        return fmt.Errorf("priority queue %d full", work.Priority)
    }
}

// SubmitWithTimeout submits work with timeout
func (pwp *PriorityWorkerPool) SubmitWithTimeout(work Work, timeout time.Duration) error {
    if atomic.LoadInt32(&pwp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    queue, exists := pwp.priorityQueues[work.Priority]
    if !exists {
        return fmt.Errorf("priority %d not supported", work.Priority)
    }
    
    select {
    case queue <- work:
        atomic.AddInt64(&pwp.metrics.TasksSubmitted, 1)
        atomic.AddInt64(&pwp.metrics.TasksByPriority[work.Priority], 1)
        pwp.updateQueueMetrics()
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("submit timeout for priority %d", work.Priority)
    }
}

// run executes the priority worker main loop
func (pw *PriorityWorker) run() {
    defer pw.pool.wg.Done()
    
    for {
        select {
        case <-pw.pool.quit:
            return
        default:
            work := pw.pool.scheduler.getNextWork()
            if work != nil {
                pw.processWork(*work)
            } else {
                // No work available, yield
                runtime.Gosched()
            }
        }
    }
}

// getNextWork selects the next work item based on priority scheduling
func (ps *PriorityScheduler) getNextWork() *Work {
    // Check for starvation first
    if work := ps.checkStarvation(); work != nil {
        return work
    }
    
    // Normal priority-based selection
    priorities := ps.getSortedPriorities()
    
    for _, priority := range priorities {
        queue := ps.pool.priorityQueues[priority]
        select {
        case work := <-queue:
            ps.lastScheduleTime[priority] = time.Now()
            return &work
        default:
            continue
        }
    }
    
    return nil
}

// checkStarvation checks for and handles priority starvation
func (ps *PriorityScheduler) checkStarvation() *Work {
    now := time.Now()
    
    for priority, queue := range ps.pool.priorityQueues {
        lastScheduled, exists := ps.lastScheduleTime[priority]
        if !exists {
            lastScheduled = now
            ps.lastScheduleTime[priority] = now
        }
        
        // Check if this priority has been starved
        if now.Sub(lastScheduled) > ps.starvationThreshold && len(queue) > 0 {
            select {
            case work := <-queue:
                atomic.AddInt64(&ps.pool.metrics.StarvationEvents, 1)
                ps.lastScheduleTime[priority] = now
                return &work
            default:
            }
        }
    }
    
    return nil
}

// getSortedPriorities returns priorities sorted by weight (highest first)
func (ps *PriorityScheduler) getSortedPriorities() []int {
    type priorityWeight struct {
        priority int
        weight   float64
        queueLen int
    }
    
    weights := make([]priorityWeight, 0, len(ps.priorityWeights))
    for priority, weight := range ps.priorityWeights {
        queueLen := len(ps.pool.priorityQueues[priority])
        if queueLen > 0 { // Only consider non-empty queues
            weights = append(weights, priorityWeight{
                priority: priority,
                weight:   weight,
                queueLen: queueLen,
            })
        }
    }
    
    // Sort by weight (descending)
    for i := 0; i < len(weights)-1; i++ {
        for j := i + 1; j < len(weights); j++ {
            if weights[i].weight < weights[j].weight {
                weights[i], weights[j] = weights[j], weights[i]
            }
        }
    }
    
    priorities := make([]int, len(weights))
    for i, pw := range weights {
        priorities[i] = pw.priority
    }
    
    return priorities
}

// processWork processes a work item with priority tracking
func (pw *PriorityWorker) processWork(work Work) {
    atomic.AddInt32(&pw.pool.metrics.ActiveWorkers, 1)
    defer atomic.AddInt32(&pw.pool.metrics.ActiveWorkers, -1)
    
    pw.currentWork = &work
    startTime := time.Now()
    
    // Execute work
    err := pw.executeWork(work)
    
    duration := time.Since(startTime)
    pw.updatePriorityMetrics(work.Priority, err, duration)
    
    pw.currentWork = nil
    pw.lastWorkTime = time.Now()
    atomic.AddInt64(&pw.tasksDone, 1)
    
    // Send result
    if work.Result != nil {
        select {
        case work.Result <- err:
        default:
        }
    }
}

// executeWork executes a work item
func (pw *PriorityWorker) executeWork(work Work) error {
    defer func() {
        if r := recover(); r != nil {
            // Handle panic
        }
    }()
    
    return work.Task()
}

// updatePriorityMetrics updates metrics for specific priority
func (pw *PriorityWorker) updatePriorityMetrics(priority int, err error, duration time.Duration) {
    if err != nil {
        atomic.AddInt64(&pw.pool.metrics.TasksFailed, 1)
    } else {
        atomic.AddInt64(&pw.pool.metrics.TasksCompleted, 1)
    }
    
    // Update priority-specific latency
    currentLatency := pw.pool.metrics.LatencyByPriority[priority]
    newLatency := time.Duration((currentLatency.Nanoseconds() + duration.Nanoseconds()) / 2)
    pw.pool.metrics.LatencyByPriority[priority] = newLatency
}

// updateQueueMetrics updates queue size metrics
func (pwp *PriorityWorkerPool) updateQueueMetrics() {
    totalQueueSize := int32(0)
    
    for priority, queue := range pwp.priorityQueues {
        queueSize := int32(len(queue))
        pwp.metrics.QueueSizesByPriority[priority] = queueSize
        totalQueueSize += queueSize
    }
    
    atomic.StoreInt32(&pwp.metrics.QueueSize, totalQueueSize)
}

// Close gracefully shuts down the priority pool
func (pwp *PriorityWorkerPool) Close() error {
    if !atomic.CompareAndSwapInt32(&pwp.state, 1, 2) {
        return fmt.Errorf("pool not running")
    }
    
    // Close all priority queues
    for _, queue := range pwp.priorityQueues {
        close(queue)
    }
    
    close(pwp.quit)
    pwp.wg.Wait()
    
    atomic.StoreInt32(&pwp.state, 0)
    return nil
}

// GetMetrics returns current pool metrics
func (pwp *PriorityWorkerPool) GetMetrics() PoolMetrics {
    pwp.updateQueueMetrics()
    
    metrics := pwp.metrics.PoolMetrics
    metrics.PoolSize = int32(len(pwp.workers))
    
    if metrics.QueueCapacity > 0 {
        metrics.QueueUtilization = float64(metrics.QueueSize) / float64(metrics.QueueCapacity)
    }
    
    return metrics
}

// Resize changes the pool size
func (pwp *PriorityWorkerPool) Resize(newSize int) error {
    return fmt.Errorf("resize not supported in priority worker pool")
}

// metricsCollector collects metrics for priority pool
func (pwp *PriorityWorkerPool) metricsCollector() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    lastCompleted := int64(0)
    
    for {
        select {
        case <-ticker.C:
            completed := atomic.LoadInt64(&pwp.metrics.TasksCompleted)
            throughput := float64(completed - lastCompleted)
            pwp.metrics.ThroughputPerSec = throughput
            lastCompleted = completed
            
            pwp.updateQueueMetrics()
            
        case <-pwp.quit:
            return
        }
    }
}
```

## Specialized Pool Types

Specialized worker pool implementations for specific use cases.

### Batch Processing Pool

```go
// BatchWorkerPool processes work items in batches for efficiency
type BatchWorkerPool struct {
    batchSize     int
    flushInterval time.Duration
    workers       []*BatchWorker
    workQueue     chan Work
    quit          chan struct{}
    wg            sync.WaitGroup
    config        WorkerPoolConfig
    metrics       *PoolMetrics
    state         int32
}

// BatchWorker processes work items in batches
type BatchWorker struct {
    id           int
    pool         *BatchWorkerPool
    batch        []Work
    lastFlush    time.Time
    batchProcessor func([]Work) error
}

// NewBatchWorkerPool creates a new batch processing pool
func NewBatchWorkerPool(config WorkerPoolConfig, batchSize int, flushInterval time.Duration) *BatchWorkerPool {
    pool := &BatchWorkerPool{
        batchSize:     batchSize,
        flushInterval: flushInterval,
        workQueue:     make(chan Work, config.QueueSize),
        quit:          make(chan struct{}),
        config:        config,
        metrics:       &PoolMetrics{QueueCapacity: int32(config.QueueSize)},
    }
    
    // Create workers
    pool.workers = make([]*BatchWorker, config.MinWorkers)
    for i := 0; i < config.MinWorkers; i++ {
        pool.workers[i] = &BatchWorker{
            id:        i,
            pool:      pool,
            batch:     make([]Work, 0, batchSize),
            lastFlush: time.Now(),
        }
    }
    
    return pool
}

// Start starts the batch worker pool
func (bwp *BatchWorkerPool) Start() error {
    if !atomic.CompareAndSwapInt32(&bwp.state, 0, 1) {
        return fmt.Errorf("pool already running")
    }
    
    for _, worker := range bwp.workers {
        bwp.wg.Add(1)
        go worker.run()
    }
    
    return nil
}

// run executes the batch worker main loop
func (bw *BatchWorker) run() {
    defer bw.pool.wg.Done()
    
    flushTicker := time.NewTicker(bw.pool.flushInterval)
    defer flushTicker.Stop()
    
    for {
        select {
        case work := <-bw.pool.workQueue:
            bw.addToBatch(work)
            
            if len(bw.batch) >= bw.pool.batchSize {
                bw.flushBatch()
            }
            
        case <-flushTicker.C:
            if len(bw.batch) > 0 {
                bw.flushBatch()
            }
            
        case <-bw.pool.quit:
            // Flush remaining work before stopping
            if len(bw.batch) > 0 {
                bw.flushBatch()
            }
            return
        }
    }
}

// addToBatch adds a work item to the current batch
func (bw *BatchWorker) addToBatch(work Work) {
    bw.batch = append(bw.batch, work)
}

// flushBatch processes the current batch
func (bw *BatchWorker) flushBatch() {
    if len(bw.batch) == 0 {
        return
    }
    
    atomic.AddInt32(&bw.pool.metrics.ActiveWorkers, 1)
    defer atomic.AddInt32(&bw.pool.metrics.ActiveWorkers, -1)
    
    startTime := time.Now()
    
    // Process batch
    var err error
    if bw.batchProcessor != nil {
        err = bw.batchProcessor(bw.batch)
    } else {
        err = bw.defaultBatchProcessor(bw.batch)
    }
    
    duration := time.Since(startTime)
    bw.updateBatchMetrics(len(bw.batch), err, duration)
    
    // Clear batch
    bw.batch = bw.batch[:0]
    bw.lastFlush = time.Now()
}

// defaultBatchProcessor processes batch items individually
func (bw *BatchWorker) defaultBatchProcessor(batch []Work) error {
    var lastError error
    
    for _, work := range batch {
        if err := work.Task(); err != nil {
            lastError = err
            
            // Send individual result
            if work.Result != nil {
                select {
                case work.Result <- err:
                default:
                }
            }
        } else {
            if work.Result != nil {
                select {
                case work.Result <- nil:
                default:
                }
            }
        }
    }
    
    return lastError
}

// updateBatchMetrics updates metrics after batch processing
func (bw *BatchWorker) updateBatchMetrics(batchSize int, err error, duration time.Duration) {
    if err != nil {
        atomic.AddInt64(&bw.pool.metrics.TasksFailed, int64(batchSize))
    } else {
        atomic.AddInt64(&bw.pool.metrics.TasksCompleted, int64(batchSize))
    }
    
    // Update average latency
    avgDuration := duration / time.Duration(batchSize)
    currentAvg := bw.pool.metrics.AverageLatency
    newAvg := time.Duration((currentAvg.Nanoseconds() + avgDuration.Nanoseconds()) / 2)
    bw.pool.metrics.AverageLatency = newAvg
}

// Submit submits work to the batch pool
func (bwp *BatchWorkerPool) Submit(work Work) error {
    if atomic.LoadInt32(&bwp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    select {
    case bwp.workQueue <- work:
        atomic.AddInt64(&bwp.metrics.TasksSubmitted, 1)
        return nil
    default:
        return fmt.Errorf("work queue full")
    }
}

// SubmitWithTimeout submits work with timeout
func (bwp *BatchWorkerPool) SubmitWithTimeout(work Work, timeout time.Duration) error {
    if atomic.LoadInt32(&bwp.state) != 1 {
        return fmt.Errorf("pool not running")
    }
    
    select {
    case bwp.workQueue <- work:
        atomic.AddInt64(&bwp.metrics.TasksSubmitted, 1)
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("submit timeout")
    }
}

// Close gracefully shuts down the batch pool
func (bwp *BatchWorkerPool) Close() error {
    if !atomic.CompareAndSwapInt32(&bwp.state, 1, 2) {
        return fmt.Errorf("pool not running")
    }
    
    close(bwp.workQueue)
    bwp.wg.Wait()
    close(bwp.quit)
    
    atomic.StoreInt32(&bwp.state, 0)
    return nil
}

// GetMetrics returns current batch pool metrics
func (bwp *BatchWorkerPool) GetMetrics() PoolMetrics {
    metrics := *bwp.metrics
    metrics.QueueSize = int32(len(bwp.workQueue))
    metrics.PoolSize = int32(len(bwp.workers))
    metrics.QueueUtilization = float64(metrics.QueueSize) / float64(bwp.config.QueueSize)
    return metrics
}

// Resize changes the batch pool size
func (bwp *BatchWorkerPool) Resize(newSize int) error {
    return fmt.Errorf("resize not supported in batch worker pool")
}
```

## Load Balancing

Load balancing strategies for distributing work across pools.

### Load Balancer Implementation

```go
// LoadBalancer defines interface for load balancing strategies
type LoadBalancer interface {
    SelectPool(pools []WorkerPool) WorkerPool
    UpdateMetrics(pool WorkerPool, metrics PoolMetrics)
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
    current int64
}

// NewRoundRobinBalancer creates a new round-robin balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
    return &RoundRobinBalancer{}
}

// SelectPool selects a pool using round-robin strategy
func (rrb *RoundRobinBalancer) SelectPool(pools []WorkerPool) WorkerPool {
    if len(pools) == 0 {
        return nil
    }
    
    index := atomic.AddInt64(&rrb.current, 1) % int64(len(pools))
    return pools[index]
}

// UpdateMetrics is not used in round-robin balancing
func (rrb *RoundRobinBalancer) UpdateMetrics(pool WorkerPool, metrics PoolMetrics) {
    // No-op for round-robin
}

// WeightedBalancer implements weighted load balancing
type WeightedBalancer struct {
    poolMetrics map[WorkerPool]PoolMetrics
    weights     map[WorkerPool]float64
    mu          sync.RWMutex
}

// NewWeightedBalancer creates a new weighted balancer
func NewWeightedBalancer() *WeightedBalancer {
    return &WeightedBalancer{
        poolMetrics: make(map[WorkerPool]PoolMetrics),
        weights:     make(map[WorkerPool]float64),
    }
}

// SelectPool selects a pool based on weighted metrics
func (wb *WeightedBalancer) SelectPool(pools []WorkerPool) WorkerPool {
    if len(pools) == 0 {
        return nil
    }
    
    wb.mu.RLock()
    defer wb.mu.RUnlock()
    
    bestPool := pools[0]
    bestScore := wb.calculateScore(bestPool)
    
    for _, pool := range pools[1:] {
        score := wb.calculateScore(pool)
        if score > bestScore {
            bestScore = score
            bestPool = pool
        }
    }
    
    return bestPool
}

// calculateScore calculates a score for pool selection
func (wb *WeightedBalancer) calculateScore(pool WorkerPool) float64 {
    metrics, exists := wb.poolMetrics[pool]
    if !exists {
        return 0.5 // Default score for unknown pools
    }
    
    // Score based on queue utilization (lower is better)
    queueScore := 1.0 - metrics.QueueUtilization
    
    // Score based on worker utilization (balanced is better)
    utilizationScore := 1.0 - abs(metrics.WorkerUtilization-0.7) // Target 70% utilization
    
    // Score based on throughput (higher is better)
    throughputScore := min(metrics.ThroughputPerSec/1000.0, 1.0)
    
    // Weighted combination
    return queueScore*0.4 + utilizationScore*0.3 + throughputScore*0.3
}

// UpdateMetrics updates metrics for load balancing decisions
func (wb *WeightedBalancer) UpdateMetrics(pool WorkerPool, metrics PoolMetrics) {
    wb.mu.Lock()
    defer wb.mu.Unlock()
    
    wb.poolMetrics[pool] = metrics
}

// Helper functions
func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

// PoolHealthChecker monitors pool health
type PoolHealthChecker struct {
    pools     map[string]WorkerPool
    health    map[string]HealthStatus
    thresholds HealthThresholds
    mu        sync.RWMutex
}

// HealthStatus represents pool health status
type HealthStatus struct {
    IsHealthy        bool
    LastHealthCheck  time.Time
    FailureCount     int
    ErrorRate        float64
    ResponseTime     time.Duration
}

// HealthThresholds defines health check thresholds
type HealthThresholds struct {
    MaxErrorRate      float64
    MaxResponseTime   time.Duration
    MaxFailureCount   int
    CheckInterval     time.Duration
}

// NewPoolHealthChecker creates a new pool health checker
func NewPoolHealthChecker() *PoolHealthChecker {
    return &PoolHealthChecker{
        pools:  make(map[string]WorkerPool),
        health: make(map[string]HealthStatus),
        thresholds: HealthThresholds{
            MaxErrorRate:    0.1,  // 10% error rate
            MaxResponseTime: 5 * time.Second,
            MaxFailureCount: 5,
            CheckInterval:   30 * time.Second,
        },
    }
}

// AddPool adds a pool to health monitoring
func (phc *PoolHealthChecker) AddPool(name string, pool WorkerPool) {
    phc.mu.Lock()
    defer phc.mu.Unlock()
    
    phc.pools[name] = pool
    phc.health[name] = HealthStatus{
        IsHealthy:       true,
        LastHealthCheck: time.Now(),
    }
    
    go phc.monitorPool(name)
}

// monitorPool continuously monitors a pool's health
func (phc *PoolHealthChecker) monitorPool(name string) {
    ticker := time.NewTicker(phc.thresholds.CheckInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        phc.checkPoolHealth(name)
    }
}

// checkPoolHealth performs health check on a pool
func (phc *PoolHealthChecker) checkPoolHealth(name string) {
    phc.mu.Lock()
    defer phc.mu.Unlock()
    
    pool, exists := phc.pools[name]
    if !exists {
        return
    }
    
    metrics := pool.GetMetrics()
    status := phc.health[name]
    
    // Check error rate
    errorRate := float64(metrics.TasksFailed) / float64(metrics.TasksSubmitted)
    if errorRate > phc.thresholds.MaxErrorRate {
        status.FailureCount++
        status.IsHealthy = false
    }
    
    // Check response time
    if metrics.AverageLatency > phc.thresholds.MaxResponseTime {
        status.FailureCount++
        status.IsHealthy = false
    }
    
    // Reset health if under thresholds
    if errorRate <= phc.thresholds.MaxErrorRate && 
       metrics.AverageLatency <= phc.thresholds.MaxResponseTime {
        if status.FailureCount > 0 {
            status.FailureCount--
        }
        if status.FailureCount == 0 {
            status.IsHealthy = true
        }
    }
    
    // Mark unhealthy if too many failures
    if status.FailureCount >= phc.thresholds.MaxFailureCount {
        status.IsHealthy = false
    }
    
    status.LastHealthCheck = time.Now()
    status.ErrorRate = errorRate
    status.ResponseTime = metrics.AverageLatency
    
    phc.health[name] = status
}

// IsHealthy checks if a pool is healthy
func (phc *PoolHealthChecker) IsHealthy(name string) bool {
    phc.mu.RLock()
    defer phc.mu.RUnlock()
    
    if status, exists := phc.health[name]; exists {
        return status.IsHealthy
    }
    return false
}

// GetHealthStatus returns health status for a pool
func (phc *PoolHealthChecker) GetHealthStatus(name string) (HealthStatus, bool) {
    phc.mu.RLock()
    defer phc.mu.RUnlock()
    
    status, exists := phc.health[name]
    return status, exists
}

// AggregatedMetrics provides aggregated metrics across pools
type AggregatedMetrics struct {
    totalPools      int32
    totalTasks      int64
    totalFailures   int64
    avgThroughput   float64
    avgLatency      time.Duration
    mu              sync.RWMutex
}

// NewAggregatedMetrics creates new aggregated metrics
func NewAggregatedMetrics() *AggregatedMetrics {
    return &AggregatedMetrics{}
}

// UpdateMetrics updates aggregated metrics
func (am *AggregatedMetrics) UpdateMetrics(metrics PoolMetrics) {
    am.mu.Lock()
    defer am.mu.Unlock()
    
    am.totalTasks += metrics.TasksCompleted
    am.totalFailures += metrics.TasksFailed
    
    // Update averages (simplified)
    am.avgThroughput = (am.avgThroughput + metrics.ThroughputPerSec) / 2
    am.avgLatency = time.Duration((am.avgLatency.Nanoseconds() + metrics.AverageLatency.Nanoseconds()) / 2)
}

// GetAggregatedMetrics returns current aggregated metrics
func (am *AggregatedMetrics) GetAggregatedMetrics() (int64, int64, float64, time.Duration) {
    am.mu.RLock()
    defer am.mu.RUnlock()
    
    return am.totalTasks, am.totalFailures, am.avgThroughput, am.avgLatency
}
```

## Summary

Worker pool patterns provide essential building blocks for concurrent Go applications:

1. **Simple Pools**: Fixed-size pools for predictable workloads
2. **Dynamic Pools**: Auto-scaling pools for variable workloads  
3. **Priority Pools**: Priority-based task processing with starvation prevention
4. **Batch Pools**: Efficient batch processing for related tasks
5. **Load Balancing**: Intelligent work distribution across pools
6. **Health Monitoring**: Continuous pool health assessment

Choose the appropriate pattern based on your specific requirements for performance, scalability, and resource utilization.
