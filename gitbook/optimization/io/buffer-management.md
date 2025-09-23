# Buffer Management

Comprehensive guide to efficient buffer management in Go applications. This guide covers buffer pooling, sizing strategies, zero-copy techniques, and advanced optimization patterns for I/O operations.

## Table of Contents

- [Introduction](#introduction)
- [Buffer Pool Management](#buffer-pool-management)
- [Buffer Sizing Strategies](#buffer-sizing-strategies)
- [Zero-Copy Techniques](#zero-copy-techniques)
- [Advanced Buffer Patterns](#advanced-buffer-patterns)
- [Performance Optimization](#performance-optimization)
- [Monitoring and Metrics](#monitoring-and-metrics)
- [Best Practices](#best-practices)

## Introduction

Buffer management is critical for high-performance I/O operations in Go applications. Efficient buffer usage reduces memory allocations, minimizes garbage collection pressure, and improves overall application throughput.

### Buffer Management Framework

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// BufferManager coordinates buffer allocation and reuse across the application
type BufferManager struct {
    pools       map[BufferClass]*SizedBufferPool
    factory     BufferFactory
    allocator   BufferAllocator
    monitor     *BufferMonitor
    optimizer   *BufferOptimizer
    metrics     *BufferMetrics
    config      BufferManagerConfig
    mu          sync.RWMutex
}

// BufferManagerConfig contains buffer manager configuration
type BufferManagerConfig struct {
    EnablePooling       bool
    EnableOptimization  bool
    EnableMonitoring    bool
    DefaultPoolSize     int
    MaxPoolSize         int
    BufferSizes         []int
    GCTriggerRatio      float64
    OptimizationPeriod  time.Duration
    MonitoringInterval  time.Duration
    AllocationStrategy  AllocationStrategy
}

// BufferClass represents different buffer size classes
type BufferClass int

const (
    SmallBuffer  BufferClass = iota // 1KB - 4KB
    MediumBuffer                    // 4KB - 64KB
    LargeBuffer                     // 64KB - 1MB
    HugeBuffer                      // 1MB+
)

// AllocationStrategy defines buffer allocation strategies
type AllocationStrategy int

const (
    PooledAllocation AllocationStrategy = iota
    DirectAllocation
    HybridAllocation
    ZeroCopyAllocation
)

// SizedBufferPool manages buffers of a specific size class
type SizedBufferPool struct {
    class       BufferClass
    bufferSize  int
    pool        chan *ManagedBuffer
    factory     func() *ManagedBuffer
    validator   func(*ManagedBuffer) bool
    resetter    func(*ManagedBuffer)
    stats       *PoolStatistics
    config      PoolConfig
    mu          sync.RWMutex
}

// PoolConfig contains pool-specific configuration
type PoolConfig struct {
    InitialSize     int
    MaxSize         int
    GrowthFactor    float64
    ShrinkThreshold float64
    ValidateOnGet   bool
    ValidateOnPut   bool
    ResetOnPut      bool
    MaxAge          time.Duration
}

// PoolStatistics tracks pool performance metrics
type PoolStatistics struct {
    BuffersCreated   int64
    BuffersRetrieved int64
    BuffersReturned  int64
    BuffersDestroyed int64
    HitRate          float64
    MissRate         float64
    MemoryUsage      int64
    PeakMemoryUsage  int64
    CurrentSize      int32
    PeakSize         int32
}

// ManagedBuffer represents a managed buffer with lifecycle tracking
type ManagedBuffer struct {
    Buffer      *bytes.Buffer
    Data        []byte
    Size        int
    Capacity    int
    Class       BufferClass
    CreatedAt   time.Time
    LastUsed    time.Time
    UseCount    int64
    State       BufferState
    Metadata    map[string]interface{}
    pool        *SizedBufferPool
    manager     *BufferManager
}

// BufferState defines buffer states
type BufferState int

const (
    BufferIdle BufferState = iota
    BufferInUse
    BufferReturned
    BufferExpired
    BufferCorrupted
)

// BufferFactory creates new buffers
type BufferFactory interface {
    CreateBuffer(class BufferClass, size int) *ManagedBuffer
    ValidateBuffer(buffer *ManagedBuffer) bool
    ResetBuffer(buffer *ManagedBuffer) error
    DestroyBuffer(buffer *ManagedBuffer) error
}

// BufferAllocator handles buffer allocation strategies
type BufferAllocator interface {
    AllocateBuffer(size int, options AllocOptions) (*ManagedBuffer, error)
    DeallocateBuffer(buffer *ManagedBuffer) error
    GetAllocationStrategy() AllocationStrategy
    SetAllocationStrategy(strategy AllocationStrategy)
}

// AllocOptions contains buffer allocation options
type AllocOptions struct {
    PreferredClass  BufferClass
    MinSize         int
    MaxSize         int
    ZeroCopy        bool
    Alignment       int
    Locality        MemoryLocality
    Priority        AllocationPriority
}

// MemoryLocality defines memory locality preferences
type MemoryLocality int

const (
    AnyLocality MemoryLocality = iota
    LocalLocality
    RemoteLocality
    NUMALocality
)

// AllocationPriority defines allocation priority
type AllocationPriority int

const (
    LowPriority AllocationPriority = iota
    NormalPriority
    HighPriority
    CriticalPriority
)

// BufferMonitor monitors buffer usage patterns
type BufferMonitor struct {
    events      chan BufferEvent
    collectors  []BufferCollector
    analyzer    *BufferAnalyzer
    alerting    *BufferAlerting
    running     bool
    mu          sync.RWMutex
}

// BufferEvent represents a buffer lifecycle event
type BufferEvent struct {
    EventType   BufferEventType
    BufferID    string
    Size        int
    Class       BufferClass
    Timestamp   time.Time
    Duration    time.Duration
    Success     bool
    Error       error
    Metadata    map[string]interface{}
}

// BufferEventType defines buffer event types
type BufferEventType int

const (
    BufferAllocated BufferEventType = iota
    BufferRetrieved
    BufferReturned
    BufferResized
    BufferReset
    BufferDestroyed
    BufferCorruption
    BufferLeak
)

// BufferCollector collects buffer metrics
type BufferCollector interface {
    CollectEvent(event BufferEvent)
    GetMetrics() map[string]interface{}
    Reset()
}

// BufferAnalyzer analyzes buffer usage patterns
type BufferAnalyzer struct {
    patterns    map[string]*BufferPattern
    trends      *BufferTrends
    predictions *BufferPredictions
    config      AnalyzerConfig
}

// BufferPattern represents a buffer usage pattern
type BufferPattern struct {
    Name            string
    Class           BufferClass
    AverageSize     int
    UsageFrequency  float64
    LifetimeMean    time.Duration
    LifetimeStdDev  time.Duration
    OptimalPoolSize int
    Efficiency      float64
}

// BufferTrends tracks buffer usage trends
type BufferTrends struct {
    AllocationTrend  TrendDirection
    SizeTrend        TrendDirection
    UsageTrend       TrendDirection
    EfficiencyTrend  TrendDirection
    PredictedPeak    time.Time
    PredictedLoad    float64
}

// TrendDirection defines trend directions
type TrendDirection int

const (
    TrendUnknown TrendDirection = iota
    TrendIncreasing
    TrendDecreasing
    TrendStable
    TrendVolatile
)

// BufferPredictions provides buffer usage predictions
type BufferPredictions struct {
    NextHour    BufferDemand
    NextDay     BufferDemand
    NextWeek    BufferDemand
    Confidence  float64
    UpdatedAt   time.Time
}

// BufferDemand represents predicted buffer demand
type BufferDemand struct {
    SmallBuffers  int
    MediumBuffers int
    LargeBuffers  int
    HugeBuffers   int
    TotalMemory   int64
    PeakMemory    int64
}

// AnalyzerConfig contains analyzer configuration
type AnalyzerConfig struct {
    PatternWindowSize   time.Duration
    TrendAnalysisPeriod time.Duration
    PredictionHorizon   time.Duration
    ConfidenceThreshold float64
}

// BufferAlerting provides alerting for buffer issues
type BufferAlerting struct {
    thresholds BufferThresholds
    alerts     chan BufferAlert
    handlers   []BufferAlertHandler
}

// BufferThresholds defines alerting thresholds
type BufferThresholds struct {
    MaxMemoryUsage     int64
    MinHitRate         float64
    MaxLeakRate        float64
    MaxCorruptionRate  float64
    PoolSizeThreshold  int32
}

// BufferAlert represents a buffer alert
type BufferAlert struct {
    Type        BufferAlertType
    Severity    AlertSeverity
    Message     string
    Class       BufferClass
    Metrics     map[string]interface{}
    Timestamp   time.Time
    Suggestions []string
}

// BufferAlertType defines alert types
type BufferAlertType int

const (
    HighMemoryUsageAlert BufferAlertType = iota
    LowHitRateAlert
    BufferLeakAlert
    CorruptionAlert
    PoolSizeAlert
)

// AlertSeverity defines alert severity levels
type AlertSeverity int

const (
    InfoSeverity AlertSeverity = iota
    WarningSeverity
    ErrorSeverity
    CriticalSeverity
)

// BufferAlertHandler handles buffer alerts
type BufferAlertHandler interface {
    HandleAlert(alert BufferAlert) error
}

// BufferOptimizer optimizes buffer configurations
type BufferOptimizer struct {
    strategies []OptimizationStrategy
    simulator  *BufferSimulator
    evaluator  *PerformanceEvaluator
    config     OptimizerConfig
}

// OptimizationStrategy defines buffer optimization strategies
type OptimizationStrategy interface {
    Analyze(manager *BufferManager) (*OptimizationResult, error)
    Apply(manager *BufferManager, result *OptimizationResult) error
    Validate(manager *BufferManager) error
}

// OptimizationResult contains optimization results
type OptimizationResult struct {
    StrategyName        string
    ExpectedImprovement PerformanceGain
    Changes             []ConfigChange
    Risks               []string
    Validation          ValidationResult
}

// PerformanceGain represents expected performance improvements
type PerformanceGain struct {
    MemoryReduction     float64
    ThroughputIncrease  float64
    LatencyReduction    float64
    AllocationReduction float64
    OverallScore        float64
}

// ConfigChange represents a configuration change
type ConfigChange struct {
    Parameter string
    OldValue  interface{}
    NewValue  interface{}
    Impact    string
    Risk      RiskLevel
}

// RiskLevel defines risk levels
type RiskLevel int

const (
    LowRisk RiskLevel = iota
    MediumRisk
    HighRisk
    CriticalRisk
)

// ValidationResult contains validation results
type ValidationResult struct {
    Valid       bool
    Confidence  float64
    Issues      []string
    Warnings    []string
    Metrics     map[string]float64
}

// BufferMetrics tracks overall buffer performance
type BufferMetrics struct {
    TotalBuffers        int64
    ActiveBuffers       int64
    TotalMemoryUsage    int64
    PeakMemoryUsage     int64
    AllocationRate      float64
    DeallocationRate    float64
    HitRate             float64
    EfficiencyScore     float64
    FragmentationRatio  float64
}

// NewBufferManager creates a new buffer manager
func NewBufferManager(config BufferManagerConfig) *BufferManager {
    bm := &BufferManager{
        pools:     make(map[BufferClass]*SizedBufferPool),
        factory:   NewDefaultBufferFactory(),
        allocator: NewHybridAllocator(),
        monitor:   NewBufferMonitor(),
        optimizer: NewBufferOptimizer(),
        metrics:   &BufferMetrics{},
        config:    config,
    }
    
    // Initialize buffer pools for each class
    bm.initializePools()
    
    // Start monitoring if enabled
    if config.EnableMonitoring {
        bm.monitor.Start()
    }
    
    // Start optimization if enabled
    if config.EnableOptimization {
        go bm.optimizationLoop()
    }
    
    return bm
}

// initializePools initializes buffer pools for each class
func (bm *BufferManager) initializePools() {
    poolConfig := PoolConfig{
        InitialSize:     bm.config.DefaultPoolSize,
        MaxSize:         bm.config.MaxPoolSize,
        GrowthFactor:    1.5,
        ShrinkThreshold: 0.3,
        ValidateOnGet:   true,
        ValidateOnPut:   true,
        ResetOnPut:      true,
        MaxAge:          time.Hour,
    }
    
    // Small buffers (1KB - 4KB)
    bm.pools[SmallBuffer] = NewSizedBufferPool(SmallBuffer, 4096, poolConfig, bm.factory)
    
    // Medium buffers (4KB - 64KB)
    bm.pools[MediumBuffer] = NewSizedBufferPool(MediumBuffer, 65536, poolConfig, bm.factory)
    
    // Large buffers (64KB - 1MB)
    bm.pools[LargeBuffer] = NewSizedBufferPool(LargeBuffer, 1048576, poolConfig, bm.factory)
    
    // Huge buffers (1MB+)
    bm.pools[HugeBuffer] = NewSizedBufferPool(HugeBuffer, 4194304, poolConfig, bm.factory)
}

// GetBuffer retrieves a buffer from the appropriate pool
func (bm *BufferManager) GetBuffer(size int) (*ManagedBuffer, error) {
    class := bm.classifySize(size)
    
    bm.mu.RLock()
    pool, exists := bm.pools[class]
    bm.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("no pool for buffer class %d", class)
    }
    
    buffer := pool.Get()
    if buffer == nil {
        return nil, fmt.Errorf("failed to get buffer from pool")
    }
    
    // Resize buffer if needed
    if buffer.Size < size {
        if err := bm.resizeBuffer(buffer, size); err != nil {
            pool.Put(buffer)
            return nil, err
        }
    }
    
    buffer.State = BufferInUse
    buffer.LastUsed = time.Now()
    atomic.AddInt64(&buffer.UseCount, 1)
    atomic.AddInt64(&bm.metrics.ActiveBuffers, 1)
    
    // Record event
    if bm.config.EnableMonitoring {
        event := BufferEvent{
            EventType: BufferRetrieved,
            BufferID:  fmt.Sprintf("%p", buffer),
            Size:      size,
            Class:     class,
            Timestamp: time.Now(),
            Success:   true,
        }
        bm.monitor.RecordEvent(event)
    }
    
    return buffer, nil
}

// ReturnBuffer returns a buffer to its pool
func (bm *BufferManager) ReturnBuffer(buffer *ManagedBuffer) error {
    if buffer == nil {
        return fmt.Errorf("cannot return nil buffer")
    }
    
    buffer.State = BufferReturned
    buffer.LastUsed = time.Now()
    atomic.AddInt64(&bm.metrics.ActiveBuffers, -1)
    
    bm.mu.RLock()
    pool, exists := bm.pools[buffer.Class]
    bm.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("no pool for buffer class %d", buffer.Class)
    }
    
    pool.Put(buffer)
    
    // Record event
    if bm.config.EnableMonitoring {
        event := BufferEvent{
            EventType: BufferReturned,
            BufferID:  fmt.Sprintf("%p", buffer),
            Size:      buffer.Size,
            Class:     buffer.Class,
            Timestamp: time.Now(),
            Success:   true,
        }
        bm.monitor.RecordEvent(event)
    }
    
    return nil
}

// classifySize classifies buffer size into appropriate class
func (bm *BufferManager) classifySize(size int) BufferClass {
    switch {
    case size <= 4096:
        return SmallBuffer
    case size <= 65536:
        return MediumBuffer
    case size <= 1048576:
        return LargeBuffer
    default:
        return HugeBuffer
    }
}

// resizeBuffer resizes a buffer to the specified size
func (bm *BufferManager) resizeBuffer(buffer *ManagedBuffer, newSize int) error {
    if buffer.Capacity >= newSize {
        buffer.Size = newSize
        return nil
    }
    
    // Grow buffer capacity
    newCapacity := max(newSize, buffer.Capacity*2)
    newData := make([]byte, newSize, newCapacity)
    
    if buffer.Data != nil {
        copy(newData, buffer.Data[:min(len(buffer.Data), newSize)])
    }
    
    buffer.Data = newData
    buffer.Size = newSize
    buffer.Capacity = newCapacity
    
    // Update buffer in underlying bytes.Buffer
    buffer.Buffer.Reset()
    buffer.Buffer.Write(newData)
    
    // Record resize event
    if bm.config.EnableMonitoring {
        event := BufferEvent{
            EventType: BufferResized,
            BufferID:  fmt.Sprintf("%p", buffer),
            Size:      newSize,
            Class:     buffer.Class,
            Timestamp: time.Now(),
            Success:   true,
        }
        bm.monitor.RecordEvent(event)
    }
    
    return nil
}

// optimizationLoop runs periodic buffer optimization
func (bm *BufferManager) optimizationLoop() {
    ticker := time.NewTicker(bm.config.OptimizationPeriod)
    defer ticker.Stop()
    
    for range ticker.C {
        if bm.config.EnableOptimization {
            bm.optimizeBuffers()
        }
    }
}

// optimizeBuffers optimizes buffer configurations
func (bm *BufferManager) optimizeBuffers() {
    // Analyze current performance
    analysis := bm.analyzer.AnalyzePerformance(bm)
    
    // Generate optimization recommendations
    optimizations := bm.optimizer.GenerateOptimizations(analysis)
    
    // Apply safe optimizations
    for _, opt := range optimizations {
        if opt.Validation.Valid && opt.Validation.Confidence > 0.8 {
            bm.optimizer.ApplyOptimization(bm, opt)
        }
    }
}

// NewSizedBufferPool creates a new sized buffer pool
func NewSizedBufferPool(class BufferClass, bufferSize int, config PoolConfig, factory BufferFactory) *SizedBufferPool {
    pool := &SizedBufferPool{
        class:      class,
        bufferSize: bufferSize,
        pool:       make(chan *ManagedBuffer, config.MaxSize),
        config:     config,
        stats:      &PoolStatistics{},
    }
    
    // Set up factory function
    pool.factory = func() *ManagedBuffer {
        return factory.CreateBuffer(class, bufferSize)
    }
    
    // Set up validator
    pool.validator = factory.ValidateBuffer
    
    // Set up resetter
    pool.resetter = func(buffer *ManagedBuffer) {
        factory.ResetBuffer(buffer)
    }
    
    // Pre-populate pool
    for i := 0; i < config.InitialSize; i++ {
        buffer := pool.factory()
        pool.pool <- buffer
        atomic.AddInt32(&pool.stats.CurrentSize, 1)
        atomic.AddInt64(&pool.stats.BuffersCreated, 1)
    }
    
    return pool
}

// Get retrieves a buffer from the pool
func (sbp *SizedBufferPool) Get() *ManagedBuffer {
    atomic.AddInt64(&sbp.stats.BuffersRetrieved, 1)
    
    select {
    case buffer := <-sbp.pool:
        atomic.AddInt32(&sbp.stats.CurrentSize, -1)
        
        // Validate buffer if enabled
        if sbp.config.ValidateOnGet && sbp.validator != nil {
            if !sbp.validator(buffer) {
                // Buffer is invalid, create new one
                sbp.destroyBuffer(buffer)
                buffer = sbp.factory()
                atomic.AddInt64(&sbp.stats.BuffersCreated, 1)
            }
        }
        
        sbp.updateHitRate(true)
        return buffer
        
    default:
        // Pool is empty, create new buffer
        buffer := sbp.factory()
        atomic.AddInt64(&sbp.stats.BuffersCreated, 1)
        sbp.updateHitRate(false)
        return buffer
    }
}

// Put returns a buffer to the pool
func (sbp *SizedBufferPool) Put(buffer *ManagedBuffer) {
    if buffer == nil {
        return
    }
    
    atomic.AddInt64(&sbp.stats.BuffersReturned, 1)
    
    // Validate buffer if enabled
    if sbp.config.ValidateOnPut && sbp.validator != nil {
        if !sbp.validator(buffer) {
            sbp.destroyBuffer(buffer)
            return
        }
    }
    
    // Reset buffer if enabled
    if sbp.config.ResetOnPut && sbp.resetter != nil {
        sbp.resetter(buffer)
    }
    
    // Check buffer age
    if sbp.config.MaxAge > 0 && time.Since(buffer.CreatedAt) > sbp.config.MaxAge {
        sbp.destroyBuffer(buffer)
        return
    }
    
    // Try to return to pool
    select {
    case sbp.pool <- buffer:
        atomic.AddInt32(&sbp.stats.CurrentSize, 1)
        
        // Update peak size
        currentSize := atomic.LoadInt32(&sbp.stats.CurrentSize)
        for {
            peak := atomic.LoadInt32(&sbp.stats.PeakSize)
            if currentSize <= peak || atomic.CompareAndSwapInt32(&sbp.stats.PeakSize, peak, currentSize) {
                break
            }
        }
        
    default:
        // Pool is full, destroy buffer
        sbp.destroyBuffer(buffer)
    }
}

// destroyBuffer destroys a buffer
func (sbp *SizedBufferPool) destroyBuffer(buffer *ManagedBuffer) {
    buffer.State = BufferExpired
    atomic.AddInt64(&sbp.stats.BuffersDestroyed, 1)
    
    // Perform cleanup if needed
    if buffer.manager != nil && buffer.manager.factory != nil {
        buffer.manager.factory.DestroyBuffer(buffer)
    }
}

// updateHitRate updates pool hit rate
func (sbp *SizedBufferPool) updateHitRate(hit bool) {
    retrieved := atomic.LoadInt64(&sbp.stats.BuffersRetrieved)
    if retrieved == 0 {
        return
    }
    
    // Simplified hit rate calculation
    if hit {
        sbp.stats.HitRate = float64(retrieved-1) / float64(retrieved)
    } else {
        sbp.stats.MissRate = float64(1) / float64(retrieved)
    }
}

// GetStatistics returns pool statistics
func (sbp *SizedBufferPool) GetStatistics() PoolStatistics {
    sbp.mu.RLock()
    defer sbp.mu.RUnlock()
    
    stats := *sbp.stats
    
    // Calculate derived metrics
    total := stats.BuffersRetrieved + stats.BuffersCreated
    if total > 0 {
        stats.HitRate = float64(stats.BuffersRetrieved) / float64(total)
        stats.MissRate = float64(stats.BuffersCreated) / float64(total)
    }
    
    stats.MemoryUsage = int64(atomic.LoadInt32(&stats.CurrentSize)) * int64(sbp.bufferSize)
    
    return stats
}

// DefaultBufferFactory implements BufferFactory
type DefaultBufferFactory struct{}

// NewDefaultBufferFactory creates a new default buffer factory
func NewDefaultBufferFactory() *DefaultBufferFactory {
    return &DefaultBufferFactory{}
}

// CreateBuffer creates a new managed buffer
func (dbf *DefaultBufferFactory) CreateBuffer(class BufferClass, size int) *ManagedBuffer {
    data := make([]byte, 0, size)
    buffer := bytes.NewBuffer(data)
    
    return &ManagedBuffer{
        Buffer:    buffer,
        Data:      data,
        Size:      0,
        Capacity:  size,
        Class:     class,
        CreatedAt: time.Now(),
        LastUsed:  time.Now(),
        UseCount:  0,
        State:     BufferIdle,
        Metadata:  make(map[string]interface{}),
    }
}

// ValidateBuffer validates a buffer
func (dbf *DefaultBufferFactory) ValidateBuffer(buffer *ManagedBuffer) bool {
    if buffer == nil || buffer.Buffer == nil || buffer.Data == nil {
        return false
    }
    
    // Check for corruption
    if buffer.State == BufferCorrupted {
        return false
    }
    
    // Check capacity consistency
    if cap(buffer.Data) != buffer.Capacity {
        return false
    }
    
    return true
}

// ResetBuffer resets a buffer to its initial state
func (dbf *DefaultBufferFactory) ResetBuffer(buffer *ManagedBuffer) error {
    if buffer == nil {
        return fmt.Errorf("cannot reset nil buffer")
    }
    
    buffer.Buffer.Reset()
    buffer.Size = 0
    buffer.State = BufferIdle
    
    // Clear metadata
    for key := range buffer.Metadata {
        delete(buffer.Metadata, key)
    }
    
    return nil
}

// DestroyBuffer destroys a buffer
func (dbf *DefaultBufferFactory) DestroyBuffer(buffer *ManagedBuffer) error {
    if buffer == nil {
        return nil
    }
    
    buffer.State = BufferExpired
    buffer.Buffer = nil
    buffer.Data = nil
    buffer.Metadata = nil
    
    return nil
}

// HybridAllocator implements BufferAllocator with hybrid allocation strategy
type HybridAllocator struct {
    strategy       AllocationStrategy
    poolThreshold  int
    directSizes    map[int]bool
    stats          AllocationStats
    mu             sync.RWMutex
}

// AllocationStats tracks allocation statistics
type AllocationStats struct {
    PooledAllocations   int64
    DirectAllocations   int64
    ZeroCopyAllocations int64
    TotalMemory         int64
    PeakMemory          int64
}

// NewHybridAllocator creates a new hybrid allocator
func NewHybridAllocator() *HybridAllocator {
    return &HybridAllocator{
        strategy:      HybridAllocation,
        poolThreshold: 1048576, // 1MB threshold
        directSizes:   make(map[int]bool),
        stats:         AllocationStats{},
    }
}

// AllocateBuffer allocates a buffer using the configured strategy
func (ha *HybridAllocator) AllocateBuffer(size int, options AllocOptions) (*ManagedBuffer, error) {
    var buffer *ManagedBuffer
    var err error
    
    switch {
    case options.ZeroCopy:
        buffer, err = ha.allocateZeroCopy(size, options)
        atomic.AddInt64(&ha.stats.ZeroCopyAllocations, 1)
        
    case size > ha.poolThreshold:
        buffer, err = ha.allocateDirect(size, options)
        atomic.AddInt64(&ha.stats.DirectAllocations, 1)
        
    default:
        buffer, err = ha.allocatePooled(size, options)
        atomic.AddInt64(&ha.stats.PooledAllocations, 1)
    }
    
    if err == nil && buffer != nil {
        atomic.AddInt64(&ha.stats.TotalMemory, int64(buffer.Capacity))
        
        // Update peak memory
        for {
            peak := atomic.LoadInt64(&ha.stats.PeakMemory)
            total := atomic.LoadInt64(&ha.stats.TotalMemory)
            if total <= peak || atomic.CompareAndSwapInt64(&ha.stats.PeakMemory, peak, total) {
                break
            }
        }
    }
    
    return buffer, err
}

// allocatePooled allocates a buffer from a pool
func (ha *HybridAllocator) allocatePooled(size int, options AllocOptions) (*ManagedBuffer, error) {
    // This would integrate with the buffer manager's pool system
    return nil, fmt.Errorf("pooled allocation not implemented")
}

// allocateDirect allocates a buffer directly
func (ha *HybridAllocator) allocateDirect(size int, options AllocOptions) (*ManagedBuffer, error) {
    class := SmallBuffer
    if size > 4096 {
        class = MediumBuffer
    }
    if size > 65536 {
        class = LargeBuffer
    }
    if size > 1048576 {
        class = HugeBuffer
    }
    
    factory := NewDefaultBufferFactory()
    return factory.CreateBuffer(class, size), nil
}

// allocateZeroCopy allocates a zero-copy buffer
func (ha *HybridAllocator) allocateZeroCopy(size int, options AllocOptions) (*ManagedBuffer, error) {
    // Zero-copy allocation would use memory mapping or shared memory
    return ha.allocateDirect(size, options)
}

// DeallocateBuffer deallocates a buffer
func (ha *HybridAllocator) DeallocateBuffer(buffer *ManagedBuffer) error {
    if buffer == nil {
        return nil
    }
    
    atomic.AddInt64(&ha.stats.TotalMemory, -int64(buffer.Capacity))
    
    return nil
}

// GetAllocationStrategy returns current allocation strategy
func (ha *HybridAllocator) GetAllocationStrategy() AllocationStrategy {
    ha.mu.RLock()
    defer ha.mu.RUnlock()
    return ha.strategy
}

// SetAllocationStrategy sets allocation strategy
func (ha *HybridAllocator) SetAllocationStrategy(strategy AllocationStrategy) {
    ha.mu.Lock()
    defer ha.mu.Unlock()
    ha.strategy = strategy
}

// NewBufferMonitor creates a new buffer monitor
func NewBufferMonitor() *BufferMonitor {
    return &BufferMonitor{
        events:     make(chan BufferEvent, 10000),
        collectors: make([]BufferCollector, 0),
        analyzer:   NewBufferAnalyzer(),
        alerting:   NewBufferAlerting(),
    }
}

// Start starts the buffer monitor
func (bm *BufferMonitor) Start() error {
    bm.mu.Lock()
    defer bm.mu.Unlock()
    
    if bm.running {
        return fmt.Errorf("monitor already running")
    }
    
    bm.running = true
    go bm.monitorLoop()
    
    return nil
}

// RecordEvent records a buffer event
func (bm *BufferMonitor) RecordEvent(event BufferEvent) {
    if !bm.running {
        return
    }
    
    select {
    case bm.events <- event:
    default:
        // Event queue full
    }
}

// monitorLoop processes buffer events
func (bm *BufferMonitor) monitorLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for bm.running {
        select {
        case event := <-bm.events:
            bm.processEvent(event)
            
        case <-ticker.C:
            bm.performPeriodicTasks()
        }
    }
}

// processEvent processes a single buffer event
func (bm *BufferMonitor) processEvent(event BufferEvent) {
    // Notify collectors
    for _, collector := range bm.collectors {
        collector.CollectEvent(event)
    }
    
    // Analyze event
    if bm.analyzer != nil {
        bm.analyzer.ProcessEvent(event)
    }
    
    // Check for alerts
    bm.checkAlerts(event)
}

// checkAlerts checks for alert conditions
func (bm *BufferMonitor) checkAlerts(event BufferEvent) {
    // Implementation would check various alert conditions
    // and send alerts through the alerting system
}

// performPeriodicTasks performs periodic monitoring tasks
func (bm *BufferMonitor) performPeriodicTasks() {
    // Aggregate metrics, update trends, etc.
}

// NewBufferAnalyzer creates a new buffer analyzer
func NewBufferAnalyzer() *BufferAnalyzer {
    return &BufferAnalyzer{
        patterns:    make(map[string]*BufferPattern),
        trends:      &BufferTrends{},
        predictions: &BufferPredictions{},
        config: AnalyzerConfig{
            PatternWindowSize:   time.Hour,
            TrendAnalysisPeriod: time.Hour * 24,
            PredictionHorizon:   time.Hour * 24,
            ConfidenceThreshold: 0.8,
        },
    }
}

// ProcessEvent processes a buffer event for analysis
func (ba *BufferAnalyzer) ProcessEvent(event BufferEvent) {
    // Update patterns, trends, and predictions based on the event
}

// AnalyzePerformance analyzes buffer manager performance
func (ba *BufferAnalyzer) AnalyzePerformance(manager *BufferManager) map[string]interface{} {
    return make(map[string]interface{})
}

// NewBufferAlerting creates a new buffer alerting system
func NewBufferAlerting() *BufferAlerting {
    return &BufferAlerting{
        thresholds: BufferThresholds{
            MaxMemoryUsage:    1024 * 1024 * 1024, // 1GB
            MinHitRate:        0.8,
            MaxLeakRate:       0.01,
            MaxCorruptionRate: 0.001,
            PoolSizeThreshold: 10000,
        },
        alerts:   make(chan BufferAlert, 1000),
        handlers: make([]BufferAlertHandler, 0),
    }
}

// NewBufferOptimizer creates a new buffer optimizer
func NewBufferOptimizer() *BufferOptimizer {
    return &BufferOptimizer{
        strategies: []OptimizationStrategy{
            &PoolSizeStrategy{},
            &BufferSizeStrategy{},
            &AllocationStrategy{},
        },
        simulator: NewBufferSimulator(),
        evaluator: NewPerformanceEvaluator(),
    }
}

// GenerateOptimizations generates optimization recommendations
func (bo *BufferOptimizer) GenerateOptimizations(analysis map[string]interface{}) []*OptimizationResult {
    return make([]*OptimizationResult, 0)
}

// ApplyOptimization applies an optimization
func (bo *BufferOptimizer) ApplyOptimization(manager *BufferManager, result *OptimizationResult) error {
    return nil
}

// Strategy implementations
type PoolSizeStrategy struct{}
type BufferSizeStrategy struct{}
type AllocationStrategyImpl struct{}

// BufferSimulator simulates buffer performance
type BufferSimulator struct{}

func NewBufferSimulator() *BufferSimulator {
    return &BufferSimulator{}
}

// PerformanceEvaluator evaluates buffer performance
type PerformanceEvaluator struct{}

func NewPerformanceEvaluator() *PerformanceEvaluator {
    return &PerformanceEvaluator{}
}

// Utility functions
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// Example usage
func ExampleBufferManagement() {
    // Create buffer manager
    config := BufferManagerConfig{
        EnablePooling:       true,
        EnableOptimization:  true,
        EnableMonitoring:    true,
        DefaultPoolSize:     100,
        MaxPoolSize:         1000,
        BufferSizes:         []int{4096, 65536, 1048576},
        GCTriggerRatio:      0.8,
        OptimizationPeriod:  time.Minute,
        MonitoringInterval:  time.Second,
        AllocationStrategy:  HybridAllocation,
    }
    
    manager := NewBufferManager(config)
    
    // Example: Reading and processing data
    for i := 0; i < 1000; i++ {
        // Get a buffer for reading data
        buffer, err := manager.GetBuffer(8192)
        if err != nil {
            fmt.Printf("Failed to get buffer: %v\n", err)
            continue
        }
        
        // Simulate reading data
        data := fmt.Sprintf("Data chunk %d", i)
        buffer.Buffer.WriteString(data)
        
        // Process the data
        processData(buffer.Buffer.Bytes())
        
        // Return buffer to pool
        manager.ReturnBuffer(buffer)
    }
    
    // Get performance metrics
    metrics := manager.GetMetrics()
    fmt.Printf("Total buffers: %d\n", metrics.TotalBuffers)
    fmt.Printf("Memory usage: %d bytes\n", metrics.TotalMemoryUsage)
    fmt.Printf("Hit rate: %.2f%%\n", metrics.HitRate*100)
    fmt.Printf("Efficiency score: %.2f\n", metrics.EfficiencyScore)
}

func processData(data []byte) {
    // Simulate data processing
    _ = len(data)
}

// GetMetrics returns buffer manager metrics
func (bm *BufferManager) GetMetrics() *BufferMetrics {
    bm.mu.RLock()
    defer bm.mu.RUnlock()
    
    metrics := *bm.metrics
    
    // Aggregate metrics from all pools
    totalMemory := int64(0)
    totalHits := int64(0)
    totalRequests := int64(0)
    
    for _, pool := range bm.pools {
        stats := pool.GetStatistics()
        totalMemory += stats.MemoryUsage
        totalHits += stats.BuffersRetrieved
        totalRequests += stats.BuffersRetrieved + stats.BuffersCreated
    }
    
    metrics.TotalMemoryUsage = totalMemory
    if totalRequests > 0 {
        metrics.HitRate = float64(totalHits) / float64(totalRequests)
    }
    
    // Calculate efficiency score
    metrics.EfficiencyScore = bm.calculateEfficiencyScore()
    
    return &metrics
}

// calculateEfficiencyScore calculates overall efficiency score
func (bm *BufferManager) calculateEfficiencyScore() float64 {
    score := 100.0
    
    // Factor in hit rate
    score *= bm.metrics.HitRate
    
    // Factor in memory efficiency
    if bm.metrics.PeakMemoryUsage > 0 {
        memoryEfficiency := float64(bm.metrics.TotalMemoryUsage) / float64(bm.metrics.PeakMemoryUsage)
        score *= memoryEfficiency
    }
    
    // Factor in fragmentation
    score *= (1.0 - bm.metrics.FragmentationRatio)
    
    return score
}
```

## Zero-Copy Techniques

Advanced zero-copy techniques for minimizing memory operations and improving performance.

### Memory Mapping

Using memory-mapped files for efficient I/O operations without buffer copying.

### Shared Memory Buffers

Implementing shared memory buffers for inter-process communication.

### Direct I/O Operations

Optimizing I/O operations to minimize memory copying between kernel and user space.

## Advanced Buffer Patterns

Specialized buffer patterns for different application scenarios.

### Ring Buffers

Implementing efficient ring buffers for streaming applications.

### Scatter-Gather Buffers

Using scatter-gather techniques for efficient network I/O.

### Hierarchical Buffering

Implementing multi-level buffering strategies for complex data flows.

## Best Practices

1. **Pool Management**: Use appropriate buffer pools for different size classes
2. **Size Classification**: Classify buffers by size for optimal pooling
3. **Lifecycle Tracking**: Track buffer lifecycle to prevent leaks
4. **Performance Monitoring**: Monitor buffer usage patterns and efficiency
5. **Zero-Copy**: Use zero-copy techniques when possible
6. **Alignment**: Consider memory alignment for better cache performance
7. **Validation**: Validate buffers before reuse
8. **Optimization**: Continuously optimize buffer configurations

## Summary

Efficient buffer management is essential for high-performance Go applications:

1. **Pooling**: Implement sophisticated buffer pooling strategies
2. **Monitoring**: Monitor buffer usage and performance metrics
3. **Optimization**: Apply dynamic optimization techniques
4. **Zero-Copy**: Leverage zero-copy techniques for best performance
5. **Patterns**: Use appropriate buffer patterns for different scenarios

These techniques enable developers to minimize memory allocation overhead and maximize I/O performance through efficient buffer management.
