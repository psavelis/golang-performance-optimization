# Goroutine Patterns Mastery

## 🎯 Learning Objectives

After completing this tutorial, you will be able to:

- **Master core goroutine patterns** for optimal concurrency design
- **Implement producer-consumer systems** with high throughput
- **Build scalable worker pools** for parallel processing
- **Design efficient pipelines** for data transformation
- **Apply advanced patterns** like scatter-gather and circuit breakers
- **Optimize pattern performance** using metrics and monitoring

## 📚 What You'll Build

Throughout this tutorial, you'll create:

1. **Producer-Consumer Pipeline** - High-throughput data processing
2. **Adaptive Worker Pool** - Dynamic scaling based on load
3. **Fan-Out/Fan-In System** - Parallel processing with aggregation
4. **Circuit Breaker Pattern** - Resilient service communication

### 🔍 Prerequisites

- Solid understanding of Go channels and goroutines
- Experience with context package for cancellation
- Basic knowledge of sync package primitives
- Understanding of concurrent programming concepts

## 🚀 Why Goroutine Patterns Matter

Goroutine patterns provide proven solutions for common concurrency challenges. They help you build scalable, maintainable systems while avoiding common pitfalls like goroutine leaks and deadlocks.

### The Problem: Ad-hoc Concurrency

```go
// ❌ Poor: Ad-hoc goroutine usage
func badConcurrentProcessing(items []Item) {
    for _, item := range items {
        go func(i Item) {
            // No coordination, no limits, potential goroutine explosion
            process(i)
        }(item)
    }
    // No way to know when processing is complete
}
```

### The Solution: Structured Patterns

```go
// ✅ Good: Worker pool pattern with coordination
func goodConcurrentProcessing(items []Item) error {
    const numWorkers = 4
    
    // Input channel for work distribution
    workCh := make(chan Item, len(items))
    
    // Error handling
    var wg sync.WaitGroup
    errCh := make(chan error, numWorkers)
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go worker(workCh, errCh, &wg)
    }
    
    // Send work
    for _, item := range items {
        workCh <- item
    }
    close(workCh)
    
    // Wait for completion
    wg.Wait()
    close(errCh)
    
    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    
    return nil
}

func worker(workCh <-chan Item, errCh chan<- error, wg *sync.WaitGroup) {
    defer wg.Done()
    for item := range workCh {
        if err := process(item); err != nil {
            errCh <- err
            return
        }
    }
}
```

## 🛠️ Pattern Foundation

Let's establish the core types and interfaces we'll use throughout this tutorial:

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

// PatternType represents different goroutine pattern types
type PatternType int

const (
    ProducerConsumer PatternType = iota
    FanOut
    FanIn
    Pipeline
    WorkerPool
    RateLimiter
    CircuitBreaker
    ScatterGather
)

// GoroutinePattern defines the interface for goroutine patterns
type GoroutinePattern interface {
    Start(ctx context.Context) error
    Stop() error
    GetMetrics() PatternMetrics
    GetType() PatternType
    Configure(config PatternConfig) error
}

// PatternMetrics provides performance metrics for patterns
type PatternMetrics struct {
    GoroutinesActive   int64
    MessagesProcessed  int64
    MessagesPending    int64
    AvgLatency        time.Duration
    ThroughputPerSec  float64
    ErrorRate         float64
    MemoryUsage       int64
}

// PatternConfig contains configuration for patterns
type PatternConfig struct {
    MaxGoroutines     int
    BufferSize        int
    TimeoutDuration   time.Duration
    BatchSize         int
    RetryAttempts     int
    BackoffStrategy   string
}

// PatternManager manages multiple goroutine patterns
type PatternManager struct {
    patterns        map[string]GoroutinePattern
    metrics         map[string]*PatternMetrics
    monitor         *PatternMonitor
    optimizer       *PatternOptimizer
    balancer        *LoadBalancer
}

// NewPatternManager creates a new pattern manager
func NewPatternManager() *PatternManager {
    return &PatternManager{
        patterns:  make(map[string]GoroutinePattern),
        metrics:   make(map[string]*PatternMetrics),
        monitor:   NewPatternMonitor(),
        optimizer: NewPatternOptimizer(),
        balancer:  NewLoadBalancer(),
    }
}

// RegisterPattern registers a new pattern
func (pm *PatternManager) RegisterPattern(name string, pattern GoroutinePattern) {
    pm.patterns[name] = pattern
    pm.metrics[name] = &PatternMetrics{}
}

// StartPattern starts a registered pattern
func (pm *PatternManager) StartPattern(name string, ctx context.Context) error {
    pattern, exists := pm.patterns[name]
    if !exists {
        return fmt.Errorf("pattern %s not found", name)
    }
    
    go pm.monitor.MonitorPattern(name, pattern)
    return pattern.Start(ctx)
}

// GetPatternMetrics returns metrics for a pattern
func (pm *PatternManager) GetPatternMetrics(name string) *PatternMetrics {
    return pm.metrics[name]
}
```

## Producer-Consumer Patterns

Efficient producer-consumer implementations for high-throughput data processing.

### Single Producer, Single Consumer

```go
// SPSCQueue implements a lock-free single producer, single consumer queue
type SPSCQueue struct {
    buffer   []interface{}
    head     int64
    tail     int64
    capacity int64
    mask     int64
}

// NewSPSCQueue creates a new SPSC queue with the given capacity (must be power of 2)
func NewSPSCQueue(capacity int) *SPSCQueue {
    if capacity&(capacity-1) != 0 {
        panic("capacity must be power of 2")
    }
    
    return &SPSCQueue{
        buffer:   make([]interface{}, capacity),
        capacity: int64(capacity),
        mask:     int64(capacity - 1),
    }
}

// Push adds an item to the queue (producer only)
func (q *SPSCQueue) Push(item interface{}) bool {
    head := atomic.LoadInt64(&q.head)
    tail := atomic.LoadInt64(&q.tail)
    
    if head-tail >= q.capacity {
        return false // Queue full
    }
    
    q.buffer[head&q.mask] = item
    atomic.StoreInt64(&q.head, head+1)
    return true
}

// Pop removes an item from the queue (consumer only)
func (q *SPSCQueue) Pop() (interface{}, bool) {
    head := atomic.LoadInt64(&q.head)
    tail := atomic.LoadInt64(&q.tail)
    
    if head == tail {
        return nil, false // Queue empty
    }
    
    item := q.buffer[tail&q.mask]
    atomic.StoreInt64(&q.tail, tail+1)
    return item, true
}

// SPSCPattern implements single producer, single consumer pattern
type SPSCPattern struct {
    queue     *SPSCQueue
    producer  func() interface{}
    consumer  func(interface{}) error
    running   int64
    metrics   *PatternMetrics
    stopCh    chan struct{}
}

// NewSPSCPattern creates a new SPSC pattern
func NewSPSCPattern(capacity int, producer func() interface{}, consumer func(interface{}) error) *SPSCPattern {
    return &SPSCPattern{
        queue:    NewSPSCQueue(capacity),
        producer: producer,
        consumer: consumer,
        metrics:  &PatternMetrics{},
        stopCh:   make(chan struct{}),
    }
}

// Start starts the SPSC pattern
func (sp *SPSCPattern) Start(ctx context.Context) error {
    if !atomic.CompareAndSwapInt64(&sp.running, 0, 1) {
        return fmt.Errorf("pattern already running")
    }
    
    // Start producer goroutine
    go sp.producerLoop(ctx)
    
    // Start consumer goroutine
    go sp.consumerLoop(ctx)
    
    return nil
}

// producerLoop runs the producer goroutine
func (sp *SPSCPattern) producerLoop(ctx context.Context) {
    defer atomic.StoreInt64(&sp.running, 0)
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-sp.stopCh:
            return
        default:
            item := sp.producer()
            if item != nil {
                for !sp.queue.Push(item) {
                    runtime.Gosched() // Yield if queue full
                }
                atomic.AddInt64(&sp.metrics.MessagesProcessed, 1)
            }
        }
    }
}

// consumerLoop runs the consumer goroutine
func (sp *SPSCPattern) consumerLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-sp.stopCh:
            return
        default:
            if item, ok := sp.queue.Pop(); ok {
                if err := sp.consumer(item); err != nil {
                    // Handle error
                }
            } else {
                runtime.Gosched() // Yield if queue empty
            }
        }
    }
}

// Stop stops the SPSC pattern
func (sp *SPSCPattern) Stop() error {
    close(sp.stopCh)
    return nil
}

// GetMetrics returns pattern metrics
func (sp *SPSCPattern) GetMetrics() PatternMetrics {
    return *sp.metrics
}

// GetType returns pattern type
func (sp *SPSCPattern) GetType() PatternType {
    return ProducerConsumer
}

// Configure configures the pattern
func (sp *SPSCPattern) Configure(config PatternConfig) error {
    // Configuration implementation
    return nil
}
```

### Multiple Producer, Multiple Consumer

```go
// MPMCPattern implements multiple producer, multiple consumer pattern
type MPMCPattern struct {
    queue       chan interface{}
    producers   []func() interface{}
    consumers   []func(interface{}) error
    numProducers int
    numConsumers int
    running     int64
    metrics     *PatternMetrics
    stopCh      chan struct{}
    wg          sync.WaitGroup
}

// NewMPMCPattern creates a new MPMC pattern
func NewMPMCPattern(bufferSize, numProducers, numConsumers int) *MPMCPattern {
    return &MPMCPattern{
        queue:        make(chan interface{}, bufferSize),
        producers:    make([]func() interface{}, numProducers),
        consumers:    make([]func(interface{}) error, numConsumers),
        numProducers: numProducers,
        numConsumers: numConsumers,
        metrics:      &PatternMetrics{},
        stopCh:       make(chan struct{}),
    }
}

// SetProducer sets a producer function for a specific index
func (mp *MPMCPattern) SetProducer(index int, producer func() interface{}) {
    if index < len(mp.producers) {
        mp.producers[index] = producer
    }
}

// SetConsumer sets a consumer function for a specific index
func (mp *MPMCPattern) SetConsumer(index int, consumer func(interface{}) error) {
    if index < len(mp.consumers) {
        mp.consumers[index] = consumer
    }
}

// Start starts the MPMC pattern
func (mp *MPMCPattern) Start(ctx context.Context) error {
    if !atomic.CompareAndSwapInt64(&mp.running, 0, 1) {
        return fmt.Errorf("pattern already running")
    }
    
    // Start producer goroutines
    for i, producer := range mp.producers {
        if producer != nil {
            mp.wg.Add(1)
            go mp.producerWorker(ctx, i, producer)
        }
    }
    
    // Start consumer goroutines
    for i, consumer := range mp.consumers {
        if consumer != nil {
            mp.wg.Add(1)
            go mp.consumerWorker(ctx, i, consumer)
        }
    }
    
    return nil
}

// producerWorker runs a producer worker
func (mp *MPMCPattern) producerWorker(ctx context.Context, id int, producer func() interface{}) {
    defer mp.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-mp.stopCh:
            return
        default:
            item := producer()
            if item != nil {
                select {
                case mp.queue <- item:
                    atomic.AddInt64(&mp.metrics.MessagesProcessed, 1)
                case <-ctx.Done():
                    return
                case <-mp.stopCh:
                    return
                }
            }
        }
    }
}

// consumerWorker runs a consumer worker
func (mp *MPMCPattern) consumerWorker(ctx context.Context, id int, consumer func(interface{}) error) {
    defer mp.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-mp.stopCh:
            return
        case item := <-mp.queue:
            if err := consumer(item); err != nil {
                // Handle error
            }
        }
    }
}

// Stop stops the MPMC pattern
func (mp *MPMCPattern) Stop() error {
    close(mp.stopCh)
    mp.wg.Wait()
    close(mp.queue)
    return nil
}

// GetMetrics returns pattern metrics
func (mp *MPMCPattern) GetMetrics() PatternMetrics {
    return *mp.metrics
}

// GetType returns pattern type
func (mp *MPMCPattern) GetType() PatternType {
    return ProducerConsumer
}

// Configure configures the pattern
func (mp *MPMCPattern) Configure(config PatternConfig) error {
    // Configuration implementation
    return nil
}
```

## Fan-Out/Fan-In Patterns

Distribute work across multiple goroutines and collect results efficiently.

### Fan-Out Pattern

```go
// FanOutPattern distributes work to multiple goroutines
type FanOutPattern struct {
    input       <-chan interface{}
    outputs     []chan interface{}
    numWorkers  int
    distributor DistributionStrategy
    running     int64
    metrics     *PatternMetrics
    stopCh      chan struct{}
    wg          sync.WaitGroup
}

// DistributionStrategy defines how work is distributed
type DistributionStrategy interface {
    Distribute(item interface{}, outputs []chan interface{}) error
}

// RoundRobinDistribution distributes work in round-robin fashion
type RoundRobinDistribution struct {
    current int64
}

// Distribute implements round-robin distribution
func (rr *RoundRobinDistribution) Distribute(item interface{}, outputs []chan interface{}) error {
    index := atomic.AddInt64(&rr.current, 1) % int64(len(outputs))
    outputs[index] <- item
    return nil
}

// LoadBasedDistribution distributes based on queue load
type LoadBasedDistribution struct{}

// Distribute implements load-based distribution
func (lb *LoadBasedDistribution) Distribute(item interface{}, outputs []chan interface{}) error {
    // Find the output channel with the smallest queue
    minLoad := len(outputs[0])
    selectedIndex := 0
    
    for i, output := range outputs {
        if len(output) < minLoad {
            minLoad = len(output)
            selectedIndex = i
        }
    }
    
    outputs[selectedIndex] <- item
    return nil
}

// NewFanOutPattern creates a new fan-out pattern
func NewFanOutPattern(input <-chan interface{}, numWorkers int, bufferSize int, strategy DistributionStrategy) *FanOutPattern {
    outputs := make([]chan interface{}, numWorkers)
    for i := range outputs {
        outputs[i] = make(chan interface{}, bufferSize)
    }
    
    return &FanOutPattern{
        input:       input,
        outputs:     outputs,
        numWorkers:  numWorkers,
        distributor: strategy,
        metrics:     &PatternMetrics{},
        stopCh:      make(chan struct{}),
    }
}

// Start starts the fan-out pattern
func (fo *FanOutPattern) Start(ctx context.Context) error {
    if !atomic.CompareAndSwapInt64(&fo.running, 0, 1) {
        return fmt.Errorf("pattern already running")
    }
    
    fo.wg.Add(1)
    go fo.distributorLoop(ctx)
    
    return nil
}

// distributorLoop runs the main distribution loop
func (fo *FanOutPattern) distributorLoop(ctx context.Context) {
    defer func() {
        fo.wg.Done()
        for _, output := range fo.outputs {
            close(output)
        }
    }()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-fo.stopCh:
            return
        case item := <-fo.input:
            if err := fo.distributor.Distribute(item, fo.outputs); err != nil {
                // Handle distribution error
            } else {
                atomic.AddInt64(&fo.metrics.MessagesProcessed, 1)
            }
        }
    }
}

// GetOutputs returns output channels for workers
func (fo *FanOutPattern) GetOutputs() []<-chan interface{} {
    result := make([]<-chan interface{}, len(fo.outputs))
    for i, ch := range fo.outputs {
        result[i] = ch
    }
    return result
}

// Stop stops the fan-out pattern
func (fo *FanOutPattern) Stop() error {
    close(fo.stopCh)
    fo.wg.Wait()
    return nil
}

// GetMetrics returns pattern metrics
func (fo *FanOutPattern) GetMetrics() PatternMetrics {
    return *fo.metrics
}

// GetType returns pattern type
func (fo *FanOutPattern) GetType() PatternType {
    return FanOut
}

// Configure configures the pattern
func (fo *FanOutPattern) Configure(config PatternConfig) error {
    return nil
}
```

### Fan-In Pattern

```go
// FanInPattern collects results from multiple goroutines
type FanInPattern struct {
    inputs        []<-chan interface{}
    output        chan interface{}
    merger        MergeStrategy
    numInputs     int
    running       int64
    metrics       *PatternMetrics
    stopCh        chan struct{}
    wg            sync.WaitGroup
}

// MergeStrategy defines how results are merged
type MergeStrategy interface {
    Merge(inputs []<-chan interface{}, output chan<- interface{}, ctx context.Context)
}

// FirstAvailableMerge merges by taking first available result
type FirstAvailableMerge struct{}

// Merge implements first-available merging
func (fa *FirstAvailableMerge) Merge(inputs []<-chan interface{}, output chan<- interface{}, ctx context.Context) {
    cases := make([]reflect.SelectCase, len(inputs)+2)
    
    // Add context cancellation case
    cases[0] = reflect.SelectCase{
        Dir:  reflect.SelectRecv,
        Chan: reflect.ValueOf(ctx.Done()),
    }
    
    // Add stop channel case
    cases[1] = reflect.SelectCase{
        Dir:  reflect.SelectRecv,
        Chan: reflect.ValueOf(make(chan struct{})),
    }
    
    // Add input cases
    for i, input := range inputs {
        cases[i+2] = reflect.SelectCase{
            Dir:  reflect.SelectRecv,
            Chan: reflect.ValueOf(input),
        }
    }
    
    for {
        chosen, value, ok := reflect.Select(cases)
        
        if chosen == 0 { // Context cancelled
            return
        }
        
        if chosen == 1 { // Stop signal
            return
        }
        
        if ok {
            output <- value.Interface()
        } else {
            // Channel closed, remove from cases
            cases = append(cases[:chosen], cases[chosen+1:]...)
            if len(cases) <= 2 { // Only context and stop cases left
                return
            }
        }
    }
}

// PriorityMerge merges based on priority
type PriorityMerge struct {
    priorities []int
}

// Merge implements priority-based merging
func (pm *PriorityMerge) Merge(inputs []<-chan interface{}, output chan<- interface{}, ctx context.Context) {
    // Sort inputs by priority and process in order
    for {
        for i, priority := range pm.priorities {
            select {
            case <-ctx.Done():
                return
            case item, ok := <-inputs[i]:
                if ok {
                    output <- item
                } else {
                    return // Channel closed
                }
            default:
                // Continue to next priority
            }
        }
        
        // Yield if no items available
        runtime.Gosched()
    }
}

// NewFanInPattern creates a new fan-in pattern
func NewFanInPattern(inputs []<-chan interface{}, bufferSize int, strategy MergeStrategy) *FanInPattern {
    return &FanInPattern{
        inputs:    inputs,
        output:    make(chan interface{}, bufferSize),
        merger:    strategy,
        numInputs: len(inputs),
        metrics:   &PatternMetrics{},
        stopCh:    make(chan struct{}),
    }
}

// Start starts the fan-in pattern
func (fi *FanInPattern) Start(ctx context.Context) error {
    if !atomic.CompareAndSwapInt64(&fi.running, 0, 1) {
        return fmt.Errorf("pattern already running")
    }
    
    fi.wg.Add(1)
    go fi.mergeLoop(ctx)
    
    return nil
}

// mergeLoop runs the main merge loop
func (fi *FanInPattern) mergeLoop(ctx context.Context) {
    defer func() {
        fi.wg.Done()
        close(fi.output)
    }()
    
    fi.merger.Merge(fi.inputs, fi.output, ctx)
}

// GetOutput returns the merged output channel
func (fi *FanInPattern) GetOutput() <-chan interface{} {
    return fi.output
}

// Stop stops the fan-in pattern
func (fi *FanInPattern) Stop() error {
    close(fi.stopCh)
    fi.wg.Wait()
    return nil
}

// GetMetrics returns pattern metrics
func (fi *FanInPattern) GetMetrics() PatternMetrics {
    return *fi.metrics
}

// GetType returns pattern type
func (fi *FanInPattern) GetType() PatternType {
    return FanIn
}

// Configure configures the pattern
func (fi *FanInPattern) Configure(config PatternConfig) error {
    return nil
}

// Helper function - would need to import reflect package
var reflect interface{} // Placeholder for reflect package
```

## Pipeline Patterns

Build efficient processing pipelines with staged operations.

### Stage-Based Pipeline

```go
// PipelineStage represents a single processing stage
type PipelineStage struct {
    Name        string
    Processor   func(interface{}) interface{}
    Parallel    int
    BufferSize  int
    input       <-chan interface{}
    output      chan interface{}
    metrics     *StageMetrics
}

// StageMetrics tracks performance metrics for a pipeline stage
type StageMetrics struct {
    ItemsProcessed int64
    ProcessingTime time.Duration
    QueueDepth     int64
    ErrorCount     int64
}

// Pipeline represents a multi-stage processing pipeline
type Pipeline struct {
    stages      []*PipelineStage
    input       chan interface{}
    output      <-chan interface{}
    running     int64
    metrics     *PatternMetrics
    stopCh      chan struct{}
    wg          sync.WaitGroup
}

// NewPipeline creates a new pipeline
func NewPipeline(bufferSize int) *Pipeline {
    input := make(chan interface{}, bufferSize)
    
    return &Pipeline{
        stages:  make([]*PipelineStage, 0),
        input:   input,
        metrics: &PatternMetrics{},
        stopCh:  make(chan struct{}),
    }
}

// AddStage adds a new stage to the pipeline
func (p *Pipeline) AddStage(name string, processor func(interface{}) interface{}, parallel int, bufferSize int) *Pipeline {
    stage := &PipelineStage{
        Name:       name,
        Processor:  processor,
        Parallel:   parallel,
        BufferSize: bufferSize,
        metrics:    &StageMetrics{},
    }
    
    // Connect to previous stage or pipeline input
    if len(p.stages) == 0 {
        stage.input = p.input
    } else {
        stage.input = p.stages[len(p.stages)-1].output
    }
    
    stage.output = make(chan interface{}, bufferSize)
    p.stages = append(p.stages, stage)
    
    // Update pipeline output
    p.output = stage.output
    
    return p
}

// Start starts the pipeline
func (p *Pipeline) Start(ctx context.Context) error {
    if !atomic.CompareAndSwapInt64(&p.running, 0, 1) {
        return fmt.Errorf("pipeline already running")
    }
    
    // Start all stages
    for _, stage := range p.stages {
        for i := 0; i < stage.Parallel; i++ {
            p.wg.Add(1)
            go p.stageWorker(ctx, stage, i)
        }
    }
    
    return nil
}

// stageWorker runs a worker for a pipeline stage
func (p *Pipeline) stageWorker(ctx context.Context, stage *PipelineStage, workerID int) {
    defer p.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-p.stopCh:
            return
        case item, ok := <-stage.input:
            if !ok {
                return // Input channel closed
            }
            
            startTime := time.Now()
            result := stage.Processor(item)
            processingTime := time.Since(startTime)
            
            // Update metrics
            atomic.AddInt64(&stage.metrics.ItemsProcessed, 1)
            stage.metrics.ProcessingTime += processingTime
            
            // Send result to next stage
            select {
            case stage.output <- result:
            case <-ctx.Done():
                return
            case <-p.stopCh:
                return
            }
        }
    }
}

// Send sends an item to the pipeline
func (p *Pipeline) Send(item interface{}) error {
    select {
    case p.input <- item:
        return nil
    default:
        return fmt.Errorf("pipeline input buffer full")
    }
}

// GetOutput returns the pipeline output channel
func (p *Pipeline) GetOutput() <-chan interface{} {
    return p.output
}

// Stop stops the pipeline
func (p *Pipeline) Stop() error {
    close(p.stopCh)
    close(p.input)
    p.wg.Wait()
    
    // Close all stage outputs
    for _, stage := range p.stages {
        close(stage.output)
    }
    
    return nil
}

// GetStageMetrics returns metrics for a specific stage
func (p *Pipeline) GetStageMetrics(stageName string) *StageMetrics {
    for _, stage := range p.stages {
        if stage.Name == stageName {
            return stage.metrics
        }
    }
    return nil
}

// GetMetrics returns pipeline metrics
func (p *Pipeline) GetMetrics() PatternMetrics {
    return *p.metrics
}

// GetType returns pattern type
func (p *Pipeline) GetType() PatternType {
    return Pipeline
}

// Configure configures the pipeline
func (p *Pipeline) Configure(config PatternConfig) error {
    return nil
}
```

## Performance Optimization

Techniques for optimizing goroutine pattern performance.

### Pattern Monitor

```go
// PatternMonitor monitors pattern performance
type PatternMonitor struct {
    patterns     map[string]GoroutinePattern
    metrics      map[string]*MonitoringMetrics
    alerts       chan Alert
    thresholds   map[string]Threshold
}

// MonitoringMetrics extends basic metrics with monitoring data
type MonitoringMetrics struct {
    PatternMetrics
    CPUUsage           float64
    MemoryUsage        int64
    GoroutineCount     int
    ChannelUtilization float64
    LastUpdated        time.Time
}

// Alert represents a performance alert
type Alert struct {
    Pattern     string
    Type        string
    Severity    string
    Message     string
    Timestamp   time.Time
    Metrics     MonitoringMetrics
}

// Threshold defines performance thresholds
type Threshold struct {
    MaxLatency         time.Duration
    MaxMemoryUsage     int64
    MaxErrorRate       float64
    MinThroughput      float64
    MaxGoroutines      int
}

// NewPatternMonitor creates a new pattern monitor
func NewPatternMonitor() *PatternMonitor {
    return &PatternMonitor{
        patterns:   make(map[string]GoroutinePattern),
        metrics:    make(map[string]*MonitoringMetrics),
        alerts:     make(chan Alert, 100),
        thresholds: make(map[string]Threshold),
    }
}

// MonitorPattern starts monitoring a pattern
func (pm *PatternMonitor) MonitorPattern(name string, pattern GoroutinePattern) {
    pm.patterns[name] = pattern
    pm.metrics[name] = &MonitoringMetrics{}
    
    go pm.monitoringLoop(name)
}

// monitoringLoop runs the monitoring loop for a pattern
func (pm *PatternMonitor) monitoringLoop(name string) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        pm.updateMetrics(name)
        pm.checkThresholds(name)
    }
}

// updateMetrics updates metrics for a pattern
func (pm *PatternMonitor) updateMetrics(name string) {
    pattern := pm.patterns[name]
    if pattern == nil {
        return
    }
    
    baseMetrics := pattern.GetMetrics()
    monitoringMetrics := pm.metrics[name]
    
    // Copy base metrics
    monitoringMetrics.PatternMetrics = baseMetrics
    
    // Add monitoring-specific metrics
    monitoringMetrics.CPUUsage = pm.getCPUUsage()
    monitoringMetrics.MemoryUsage = pm.getMemoryUsage()
    monitoringMetrics.GoroutineCount = runtime.NumGoroutine()
    monitoringMetrics.LastUpdated = time.Now()
}

// getCPUUsage gets current CPU usage
func (pm *PatternMonitor) getCPUUsage() float64 {
    // Simplified CPU usage calculation
    return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU())
}

// getMemoryUsage gets current memory usage
func (pm *PatternMonitor) getMemoryUsage() int64 {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return int64(m.Alloc)
}

// checkThresholds checks if any thresholds are exceeded
func (pm *PatternMonitor) checkThresholds(name string) {
    threshold, exists := pm.thresholds[name]
    if !exists {
        return
    }
    
    metrics := pm.metrics[name]
    
    // Check latency threshold
    if metrics.AvgLatency > threshold.MaxLatency {
        pm.sendAlert(Alert{
            Pattern:   name,
            Type:      "latency",
            Severity:  "warning",
            Message:   fmt.Sprintf("Latency %v exceeds threshold %v", metrics.AvgLatency, threshold.MaxLatency),
            Timestamp: time.Now(),
            Metrics:   *metrics,
        })
    }
    
    // Check memory threshold
    if metrics.MemoryUsage > threshold.MaxMemoryUsage {
        pm.sendAlert(Alert{
            Pattern:   name,
            Type:      "memory",
            Severity:  "critical",
            Message:   fmt.Sprintf("Memory usage %d exceeds threshold %d", metrics.MemoryUsage, threshold.MaxMemoryUsage),
            Timestamp: time.Now(),
            Metrics:   *metrics,
        })
    }
    
    // Check error rate threshold
    if metrics.ErrorRate > threshold.MaxErrorRate {
        pm.sendAlert(Alert{
            Pattern:   name,
            Type:      "error_rate",
            Severity:  "critical",
            Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", metrics.ErrorRate*100, threshold.MaxErrorRate*100),
            Timestamp: time.Now(),
            Metrics:   *metrics,
        })
    }
}

// sendAlert sends an alert
func (pm *PatternMonitor) sendAlert(alert Alert) {
    select {
    case pm.alerts <- alert:
    default:
        // Alert channel full, drop alert
    }
}

// SetThreshold sets performance thresholds for a pattern
func (pm *PatternMonitor) SetThreshold(name string, threshold Threshold) {
    pm.thresholds[name] = threshold
}

// GetAlerts returns the alerts channel
func (pm *PatternMonitor) GetAlerts() <-chan Alert {
    return pm.alerts
}

// GetMetrics returns current metrics for a pattern
func (pm *PatternMonitor) GetMetrics(name string) *MonitoringMetrics {
    return pm.metrics[name]
}
```

### Pattern Optimizer

```go
// PatternOptimizer optimizes pattern performance
type PatternOptimizer struct {
    recommendations map[string][]Recommendation
    optimizer       map[PatternType]OptimizerFunc
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
    Type        string
    Priority    int
    Description string
    Action      func() error
    Impact      string
}

// OptimizerFunc optimizes a specific pattern type
type OptimizerFunc func(pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation

// NewPatternOptimizer creates a new pattern optimizer
func NewPatternOptimizer() *PatternOptimizer {
    optimizer := &PatternOptimizer{
        recommendations: make(map[string][]Recommendation),
        optimizer:       make(map[PatternType]OptimizerFunc),
    }
    
    // Register optimizers for different pattern types
    optimizer.optimizer[ProducerConsumer] = optimizer.optimizeProducerConsumer
    optimizer.optimizer[FanOut] = optimizer.optimizeFanOut
    optimizer.optimizer[FanIn] = optimizer.optimizeFanIn
    optimizer.optimizer[Pipeline] = optimizer.optimizePipeline
    
    return optimizer
}

// OptimizePattern generates optimization recommendations for a pattern
func (po *PatternOptimizer) OptimizePattern(name string, pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation {
    optimizer, exists := po.optimizer[pattern.GetType()]
    if !exists {
        return nil
    }
    
    recommendations := optimizer(pattern, metrics)
    po.recommendations[name] = recommendations
    
    return recommendations
}

// optimizeProducerConsumer optimizes producer-consumer patterns
func (po *PatternOptimizer) optimizeProducerConsumer(pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation {
    recommendations := make([]Recommendation, 0)
    
    // Check for high latency
    if metrics.AvgLatency > 100*time.Millisecond {
        recommendations = append(recommendations, Recommendation{
            Type:        "buffer_size",
            Priority:    8,
            Description: "Increase buffer size to reduce latency",
            Impact:      "May increase memory usage but reduce blocking",
        })
    }
    
    // Check for low throughput
    if metrics.ThroughputPerSec < 1000 {
        recommendations = append(recommendations, Recommendation{
            Type:        "parallelism",
            Priority:    7,
            Description: "Increase number of consumer goroutines",
            Impact:      "Higher CPU usage but better throughput",
        })
    }
    
    // Check for high memory usage
    if metrics.MemoryUsage > 100*1024*1024 { // 100MB
        recommendations = append(recommendations, Recommendation{
            Type:        "memory_optimization",
            Priority:    6,
            Description: "Optimize memory usage with object pooling",
            Impact:      "Lower memory usage, potential GC pressure reduction",
        })
    }
    
    return recommendations
}

// optimizeFanOut optimizes fan-out patterns
func (po *PatternOptimizer) optimizeFanOut(pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation {
    recommendations := make([]Recommendation, 0)
    
    // Check for uneven load distribution
    if metrics.ThroughputPerSec < 500 {
        recommendations = append(recommendations, Recommendation{
            Type:        "load_balancing",
            Priority:    9,
            Description: "Implement better load balancing strategy",
            Impact:      "More even work distribution",
        })
    }
    
    return recommendations
}

// optimizeFanIn optimizes fan-in patterns
func (po *PatternOptimizer) optimizeFanIn(pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation {
    recommendations := make([]Recommendation, 0)
    
    // Check for merge bottlenecks
    if metrics.AvgLatency > 50*time.Millisecond {
        recommendations = append(recommendations, Recommendation{
            Type:        "merge_strategy",
            Priority:    8,
            Description: "Optimize merge strategy for better performance",
            Impact:      "Reduced merge latency",
        })
    }
    
    return recommendations
}

// optimizePipeline optimizes pipeline patterns
func (po *PatternOptimizer) optimizePipeline(pattern GoroutinePattern, metrics *MonitoringMetrics) []Recommendation {
    recommendations := make([]Recommendation, 0)
    
    // Check for pipeline bottlenecks
    if metrics.MessagesPending > 1000 {
        recommendations = append(recommendations, Recommendation{
            Type:        "stage_optimization",
            Priority:    9,
            Description: "Optimize slow pipeline stages",
            Impact:      "Better pipeline throughput",
        })
    }
    
    return recommendations
}

// GetRecommendations returns optimization recommendations for a pattern
func (po *PatternOptimizer) GetRecommendations(name string) []Recommendation {
    return po.recommendations[name]
}

// LoadBalancer balances load across pattern instances
type LoadBalancer struct {
    patterns map[string][]GoroutinePattern
    strategy LoadBalancingStrategy
}

// LoadBalancingStrategy defines load balancing behavior
type LoadBalancingStrategy interface {
    SelectPattern(patterns []GoroutinePattern) GoroutinePattern
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
    return &LoadBalancer{
        patterns: make(map[string][]GoroutinePattern),
        strategy: &RoundRobinStrategy{},
    }
}

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
    current int64
}

// SelectPattern selects a pattern using round-robin
func (rr *RoundRobinStrategy) SelectPattern(patterns []GoroutinePattern) GoroutinePattern {
    if len(patterns) == 0 {
        return nil
    }
    
    index := atomic.AddInt64(&rr.current, 1) % int64(len(patterns))
    return patterns[index]
}
```

## Summary

Goroutine patterns provide powerful abstractions for building concurrent Go applications. Key principles:

1. **Choose the Right Pattern**: Select patterns based on your specific use case
2. **Monitor Performance**: Continuously monitor pattern metrics
3. **Optimize Based on Data**: Use metrics to guide optimization decisions
4. **Balance Trade-offs**: Consider memory, CPU, and latency trade-offs
5. **Scale Appropriately**: Adjust parallelism based on workload characteristics

Effective use of these patterns can significantly improve application performance and scalability.
