# Object Reuse Mastery

## 🎯 Learning Objectives

After completing this tutorial, you will be able to:

- **Understand object reuse fundamentals** and when to apply them
- **Implement efficient object pooling** for common scenarios
- **Design lifecycle management systems** for long-lived objects
- **Build custom reuse patterns** for specific use cases
- **Monitor reuse effectiveness** with metrics and profiling
- **Optimize memory usage** through strategic object reuse

## 📚 What You'll Build

Throughout this tutorial, you'll create:

1. **Smart Object Pool** - Self-managing pool with lifecycle tracking
2. **Buffer Reuse System** - High-performance byte buffer recycling
3. **Connection Pool Manager** - Database/network connection reuse
4. **Reuse Metrics Dashboard** - Real-time effectiveness monitoring

### 🔍 Prerequisites

- Understanding of Go memory management
- Knowledge of garbage collection concepts
- Experience with sync.Pool and basic pooling
- Familiarity with performance measurement

## 🚀 Understanding Object Reuse

Object reuse is about minimizing allocation and deallocation costs by recycling objects. This is particularly important in high-throughput applications where allocation pressure can impact performance.

### The Allocation Problem

```go
// ❌ Poor: Constant allocation in hot path
func processRequests(requests <-chan Request) {
    for req := range requests {
        // Creates new objects for every request
        buffer := make([]byte, 4096)
        response := &Response{
            ID:     req.ID,
            Buffer: buffer,
            Headers: make(map[string]string),
        }
        
        processRequest(req, response)
        
        // Objects become garbage after function ends
    }
}
```

### The Reuse Solution

```go
// ✅ Good: Object reuse with pooling
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{
            Headers: make(map[string]string, 10),
        }
    },
}

func processRequestsWithReuse(requests <-chan Request) {
    for req := range requests {
        // Reuse objects from pools
        buffer := bufferPool.Get().([]byte)
        response := responsePool.Get().(*Response)
        
        // Reset for reuse
        buffer = buffer[:0]
        response.Reset()
        response.ID = req.ID
        response.Buffer = buffer
        
        processRequest(req, response)
        
        // Return to pools for reuse
        bufferPool.Put(buffer)
        responsePool.Put(response)
    }
}

// Reset method prepares object for reuse
func (r *Response) Reset() {
    r.ID = 0
    r.Buffer = nil
    r.Status = 0
    r.Timestamp = time.Time{}
    
    // Clear map without reallocating
    for k := range r.Headers {
        delete(r.Headers, k)
    }
}
```

## 🛠️ Building a Smart Object Pool

Let's create an advanced object pool that tracks usage patterns and optimizes itself:

```go
package main

import (
    "fmt"
    "reflect"
    "sync"
    "sync/atomic"
    "time"
)

// SmartPool provides intelligent object reuse with monitoring
type SmartPool struct {
    objectType    reflect.Type
    newFunc       func() interface{}
    resetFunc     func(interface{})
    pool          chan interface{}
    maxSize       int
    created       int64
    gets          int64
    puts          int64
    hits          int64
    misses        int64
    mu            sync.RWMutex
}

// NewSmartPool creates a new smart object pool
func NewSmartPool(maxSize int, newFunc func() interface{}, resetFunc func(interface{})) *SmartPool {
    return &SmartPool{
        objectType: reflect.TypeOf(newFunc()),
        newFunc:    newFunc,
        resetFunc:  resetFunc,
        pool:       make(chan interface{}, maxSize),
        maxSize:    maxSize,
    }
}

// Get retrieves an object from the pool
func (sp *SmartPool) Get() interface{} {
    atomic.AddInt64(&sp.gets, 1)
    
    select {
    case obj := <-sp.pool:
        atomic.AddInt64(&sp.hits, 1)
        return obj
    default:
        // Pool empty, create new object
        atomic.AddInt64(&sp.misses, 1)
        atomic.AddInt64(&sp.created, 1)
        return sp.newFunc()
    }
}

// Put returns an object to the pool
func (sp *SmartPool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&sp.puts, 1)
    
    // Reset object for reuse
    if sp.resetFunc != nil {
        sp.resetFunc(obj)
    }
    
    select {
    case sp.pool <- obj:
        // Successfully returned to pool
    default:
        // Pool full, let GC handle it
    }
}

// Stats returns pool statistics
func (sp *SmartPool) Stats() PoolStats {
    gets := atomic.LoadInt64(&sp.gets)
    puts := atomic.LoadInt64(&sp.puts)
    hits := atomic.LoadInt64(&sp.hits)
    misses := atomic.LoadInt64(&sp.misses)
    created := atomic.LoadInt64(&sp.created)
    
    var hitRate float64
    if gets > 0 {
        hitRate = float64(hits) / float64(gets)
    }
    
    return PoolStats{
        ObjectType:  sp.objectType.String(),
        MaxSize:     sp.maxSize,
        CurrentSize: len(sp.pool),
        Gets:        gets,
        Puts:        puts,
        Hits:        hits,
        Misses:      misses,
        Created:     created,
        HitRate:     hitRate,
    }
}

// PoolStats contains pool performance statistics
type PoolStats struct {
    ObjectType  string
    MaxSize     int
    CurrentSize int
    Gets        int64
    Puts        int64
    Hits        int64
    Misses      int64
    Created     int64
    HitRate     float64
}

// String returns human-readable stats
func (ps PoolStats) String() string {
    return fmt.Sprintf(
        "Pool[%s]: Size=%d/%d, Gets=%d, Puts=%d, Hit Rate=%.2f%%, Created=%d",
        ps.ObjectType, ps.CurrentSize, ps.MaxSize, ps.Gets, ps.Puts, ps.HitRate*100, ps.Created,
    )
}
```

### 🎯 Practical Exercise: Buffer Pool System

Let's implement a real-world buffer reuse system:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Buffer represents a reusable byte buffer
type Buffer struct {
    Data      []byte
    usageCount int64
    lastUsed   time.Time
}

// Reset prepares buffer for reuse
func (b *Buffer) Reset() {
    b.Data = b.Data[:0]  // Reset length but keep capacity
    b.usageCount++
    b.lastUsed = time.Now()
}

// BufferPool manages reusable buffers of different sizes
type BufferPool struct {
    pools map[int]*SmartPool  // Size -> Pool mapping
    mu    sync.RWMutex
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
    bp := &BufferPool{
        pools: make(map[int]*SmartPool),
    }
    
    // Create pools for common buffer sizes
    commonSizes := []int{256, 1024, 4096, 16384, 65536}
    for _, size := range commonSizes {
        bp.pools[size] = NewSmartPool(
            50, // max pool size
            func() interface{} {
                return &Buffer{
                    Data:     make([]byte, 0, size),
                    lastUsed: time.Now(),
                }
            },
            func(obj interface{}) {
                obj.(*Buffer).Reset()
            },
        )
    }
    
    return bp
}

// GetBuffer gets a buffer of at least the specified size
func (bp *BufferPool) GetBuffer(minSize int) *Buffer {
    // Find the smallest pool that can accommodate the request
    targetSize := bp.findTargetSize(minSize)
    
    bp.mu.RLock()
    pool, exists := bp.pools[targetSize]
    bp.mu.RUnlock()
    
    if !exists {
        // Create buffer directly if no suitable pool
        return &Buffer{
            Data:     make([]byte, 0, minSize),
            lastUsed: time.Now(),
        }
    }
    
    return pool.Get().(*Buffer)
}

// PutBuffer returns a buffer to the appropriate pool
func (bp *BufferPool) PutBuffer(buffer *Buffer) {
    if buffer == nil {
        return
    }
    
    capacity := cap(buffer.Data)
    targetSize := bp.findTargetSize(capacity)
    
    bp.mu.RLock()
    pool, exists := bp.pools[targetSize]
    bp.mu.RUnlock()
    
    if exists && capacity == targetSize {
        pool.Put(buffer)
    }
    // If no suitable pool, let GC handle it
}

// findTargetSize finds the best pool size for a given capacity
func (bp *BufferPool) findTargetSize(minSize int) int {
    bp.mu.RLock()
    defer bp.mu.RUnlock()
    
    bestSize := minSize
    for size := range bp.pools {
        if size >= minSize && size < bestSize {
            bestSize = size
        }
    }
    
    return bestSize
}

// GetStats returns statistics for all pools
func (bp *BufferPool) GetStats() map[int]PoolStats {
    bp.mu.RLock()
    defer bp.mu.RUnlock()
    
    stats := make(map[int]PoolStats)
    for size, pool := range bp.pools {
        stats[size] = pool.Stats()
    }
    
    return stats
}

// Demo function showing buffer pool in action
func demoBufferPool() {
    fmt.Println("🎯 Buffer Pool Demo")
    fmt.Println("==================")
    
    pool := NewBufferPool()
    
    // Simulate various buffer usage patterns
    fmt.Println("\n📝 Testing different buffer sizes:")
    
    sizes := []int{100, 500, 2000, 8000, 32000}
    buffers := make([]*Buffer, len(sizes))
    
    // Get buffers
    for i, size := range sizes {
        buffers[i] = pool.GetBuffer(size)
        fmt.Printf("   Got buffer for %d bytes (capacity: %d)\n", 
            size, cap(buffers[i].Data))
    }
    
    // Use buffers
    for i, buffer := range buffers {
        // Simulate usage
        for j := 0; j < sizes[i]; j++ {
            buffer.Data = append(buffer.Data, byte(j%256))
        }
        fmt.Printf("   Used buffer %d: wrote %d bytes\n", i, len(buffer.Data))
    }
    
    // Return buffers
    fmt.Println("\n♻️  Returning buffers to pools:")
    for i, buffer := range buffers {
        pool.PutBuffer(buffer)
        fmt.Printf("   Returned buffer %d\n", i)
    }
    
    // Show pool statistics
    fmt.Println("\n📊 Pool Statistics:")
    for size, stats := range pool.GetStats() {
        fmt.Printf("   %s\n", stats)
    }
    
    // Test reuse
    fmt.Println("\n🔄 Testing reuse:")
    for i, size := range sizes[:3] {
        buffer := pool.GetBuffer(size)
        fmt.Printf("   Reused buffer %d (capacity: %d)\n", i, cap(buffer.Data))
        pool.PutBuffer(buffer)
    }
    
    // Final statistics
    fmt.Println("\n📊 Final Pool Statistics:")
    for size, stats := range pool.GetStats() {
        fmt.Printf("   %s\n", stats)
    }
}

func main() {
    demoBufferPool()
}
```

> 💡 **Key Benefits**: This system automatically selects the right pool size, tracks usage patterns, and provides detailed metrics for optimization.
    MaxLifetime       time.Duration
    HealthCheckPeriod time.Duration
}

// PoolStatistics tracks pool performance metrics
type PoolStatistics struct {
    Created         int64
    Retrieved       int64
    Returned        int64
    Destroyed       int64
    ValidationFails int64
    ResetFails      int64
    CurrentSize     int32
    PeakSize        int32
    HitRate         float64
    MissRate        float64
    AvgLifetime     time.Duration
}

// ObjectFactory creates new objects
type ObjectFactory interface {
    CreateObject(objectType reflect.Type) interface{}
    ValidateObject(obj interface{}) bool
    ResetObject(obj interface{}) error
    DestroyObject(obj interface{}) error
}

// LifecycleManager manages object lifecycles
type LifecycleManager struct {
    tracking    map[uintptr]*ObjectLifecycle
    cleanup     chan *ObjectLifecycle
    config      LifecycleConfig
    running     bool
    mu          sync.RWMutex
}

// LifecycleConfig contains lifecycle management configuration
type LifecycleConfig struct {
    TrackObjects      bool
    MaxAge            time.Duration
    MaxUseCount       int
    CleanupInterval   time.Duration
    EnableProfiling   bool
}

// ObjectLifecycle tracks individual object lifecycle
type ObjectLifecycle struct {
    ObjectID      uintptr
    ObjectType    reflect.Type
    CreatedAt     time.Time
    LastUsed      time.Time
    UseCount      int64
    TotalLifetime time.Duration
    State         ObjectState
    Metadata      map[string]interface{}
}

// ObjectState defines object states
type ObjectState int

const (
    ObjectCreated ObjectState = iota
    ObjectInUse
    ObjectReturned
    ObjectExpired
    ObjectDestroyed
)

// ReuseMonitor monitors object reuse patterns
type ReuseMonitor struct {
    events     chan ReuseEvent
    patterns   map[string]*ReusePattern
    alerts     *ReuseAlerting
    collectors []ReuseCollector
    running    bool
    mu         sync.RWMutex
}

// ReuseEvent represents an object reuse event
type ReuseEvent struct {
    EventType    ReuseEventType
    ObjectType   reflect.Type
    ObjectID     uintptr
    PoolName     string
    Duration     time.Duration
    Size         int64
    Timestamp    time.Time
    Metadata     map[string]interface{}
}

// ReuseEventType defines reuse event types
type ReuseEventType int

const (
    ObjectCreatedEvent ReuseEventType = iota
    ObjectRetrievedEvent
    ObjectReturnedEvent
    ObjectResetEvent
    ObjectExpiredEvent
    ObjectDestroyedEvent
    PoolGrowEvent
    PoolShrinkEvent
)

// ReusePattern identifies object reuse patterns
type ReusePattern struct {
    Name          string
    ObjectType    reflect.Type
    Frequency     int64
    AvgLifetime   time.Duration
    ReuseRate     float64
    Efficiency    float64
    Optimization  string
}

// ReuseAlerting provides alerting for reuse issues
type ReuseAlerting struct {
    thresholds   ReuseThresholds
    alerts       chan ReuseAlert
    handlers     []ReuseAlertHandler
}

// ReuseThresholds defines alerting thresholds
type ReuseThresholds struct {
    MinReuseRate     float64
    MaxMissRate      float64
    MaxPoolSize      int32
    MaxObjectAge     time.Duration
    MinEfficiency    float64
}

// ReuseAlert represents a reuse alert
type ReuseAlert struct {
    Type        ReuseAlertType
    Severity    AlertSeverity
    Message     string
    ObjectType  reflect.Type
    PoolName    string
    Metrics     map[string]interface{}
    Timestamp   time.Time
}

// ReuseAlertType defines alert types
type ReuseAlertType int

const (
    LowReuseRateAlert ReuseAlertType = iota
    HighMissRateAlert
    PoolSizeAlert
    ObjectAgeAlert
    EfficiencyAlert
)

// AlertSeverity defines alert severity levels
type AlertSeverity int

const (
    InfoAlert AlertSeverity = iota
    WarningAlert
    ErrorAlert
    CriticalAlert
)

// ReuseAlertHandler handles reuse alerts
type ReuseAlertHandler interface {
    HandleAlert(alert ReuseAlert) error
}

// ReuseCollector collects reuse metrics
type ReuseCollector interface {
    CollectEvent(event ReuseEvent)
    GetMetrics() map[string]interface{}
    Reset()
}

// ReuseMetrics tracks overall reuse performance
type ReuseMetrics struct {
    TotalPools          int32
    TotalObjects        int64
    TotalAllocations    int64
    TotalReuses         int64
    TotalDestroyals     int64
    GlobalReuseRate     float64
    MemorySaved         int64
    AllocationsSaved    int64
    AverageLifetime     time.Duration
    EfficiencyScore     float64
}

// NewReuseManager creates a new reuse manager
func NewReuseManager(config ReuseConfig) *ReuseManager {
    return &ReuseManager{
        pools:     make(map[reflect.Type]*TypedObjectPool),
        factory:   NewDefaultObjectFactory(),
        lifecycle: NewLifecycleManager(config),
        monitor:   NewReuseMonitor(),
        config:    config,
        metrics:   &ReuseMetrics{},
    }
}

// GetPool returns a typed object pool
func (rm *ReuseManager) GetPool(objectType reflect.Type) *TypedObjectPool {
    rm.mu.RLock()
    if pool, exists := rm.pools[objectType]; exists {
        rm.mu.RUnlock()
        return pool
    }
    rm.mu.RUnlock()
    
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    // Double-check after acquiring write lock
    if pool, exists := rm.pools[objectType]; exists {
        return pool
    }
    
    // Create new pool
    poolConfig := PoolConfig{
        MaxSize:           rm.config.MaxPoolSize,
        MinSize:           rm.config.InitialPoolSize,
        GrowthFactor:      1.5,
        ShrinkThreshold:   0.3,
        IdleTimeout:       5 * time.Minute,
        MaxLifetime:       rm.config.MaxObjectAge,
        HealthCheckPeriod: time.Minute,
    }
    
    pool := NewTypedObjectPool(objectType, poolConfig, rm.factory)
    rm.pools[objectType] = pool
    atomic.AddInt32(&rm.metrics.TotalPools, 1)
    
    return pool
}

// GetObject retrieves an object from the appropriate pool
func (rm *ReuseManager) GetObject(objectType reflect.Type) interface{} {
    pool := rm.GetPool(objectType)
    obj := pool.Get()
    
    atomic.AddInt64(&rm.metrics.TotalObjects, 1)
    
    if rm.config.EnableLifecycle {
        rm.lifecycle.TrackObject(obj)
    }
    
    if rm.config.EnableMonitoring {
        event := ReuseEvent{
            EventType:  ObjectRetrievedEvent,
            ObjectType: objectType,
            ObjectID:   uintptr(unsafe.Pointer(&obj)),
            Timestamp:  time.Now(),
        }
        rm.monitor.RecordEvent(event)
    }
    
    return obj
}

// ReturnObject returns an object to its pool
func (rm *ReuseManager) ReturnObject(obj interface{}) {
    if obj == nil {
        return
    }
    
    objectType := reflect.TypeOf(obj)
    pool := rm.GetPool(objectType)
    
    if rm.config.EnableLifecycle {
        rm.lifecycle.UpdateObject(obj)
    }
    
    pool.Put(obj)
    atomic.AddInt64(&rm.metrics.TotalReuses, 1)
    
    if rm.config.EnableMonitoring {
        event := ReuseEvent{
            EventType:  ObjectReturnedEvent,
            ObjectType: objectType,
            ObjectID:   uintptr(unsafe.Pointer(&obj)),
            Timestamp:  time.Now(),
        }
        rm.monitor.RecordEvent(event)
    }
}

// NewTypedObjectPool creates a new typed object pool
func NewTypedObjectPool(objectType reflect.Type, config PoolConfig, factory ObjectFactory) *TypedObjectPool {
    pool := &TypedObjectPool{
        objectType: objectType,
        pool:       make(chan interface{}, config.MaxSize),
        config:     config,
        stats:      &PoolStatistics{},
        lifecycle:  &ObjectLifecycle{ObjectType: objectType},
    }
    
    // Set up factory function
    pool.factory = func() interface{} {
        return factory.CreateObject(objectType)
    }
    
    // Set up validator
    pool.validator = factory.ValidateObject
    
    // Set up resetter
    pool.resetter = func(obj interface{}) {
        factory.ResetObject(obj)
    }
    
    // Set up destructor
    pool.destructor = factory.DestroyObject
    
    // Pre-populate pool if needed
    for i := 0; i < config.MinSize; i++ {
        obj := pool.factory()
        pool.pool <- obj
        atomic.AddInt32(&pool.stats.CurrentSize, 1)
        atomic.AddInt64(&pool.stats.Created, 1)
    }
    
    // Start health check routine
    go pool.healthCheck()
    
    return pool
}

// Get retrieves an object from the pool
func (top *TypedObjectPool) Get() interface{} {
    atomic.AddInt64(&top.stats.Retrieved, 1)
    
    select {
    case obj := <-top.pool:
        atomic.AddInt32(&top.stats.CurrentSize, -1)
        
        // Validate object before use
        if top.validator != nil && !top.validator(obj) {
            atomic.AddInt64(&top.stats.ValidationFails, 1)
            // Create new object if validation fails
            obj = top.factory()
            atomic.AddInt64(&top.stats.Created, 1)
        }
        
        atomic.AddInt64(&top.stats.Retrieved, 1)
        top.updateHitRate(true)
        
        return obj
        
    default:
        // Pool is empty, create new object
        obj := top.factory()
        atomic.AddInt64(&top.stats.Created, 1)
        top.updateHitRate(false)
        
        return obj
    }
}

// Put returns an object to the pool
func (top *TypedObjectPool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&top.stats.Returned, 1)
    
    // Validate object before returning to pool
    if top.validator != nil && !top.validator(obj) {
        atomic.AddInt64(&top.stats.ValidationFails, 1)
        top.destroyObject(obj)
        return
    }
    
    // Reset object state
    if top.resetter != nil {
        if err := top.resetObject(obj); err != nil {
            atomic.AddInt64(&top.stats.ResetFails, 1)
            top.destroyObject(obj)
            return
        }
    }
    
    // Try to return to pool
    select {
    case top.pool <- obj:
        currentSize := atomic.AddInt32(&top.stats.CurrentSize, 1)
        
        // Update peak size
        for {
            peak := atomic.LoadInt32(&top.stats.PeakSize)
            if currentSize <= peak || atomic.CompareAndSwapInt32(&top.stats.PeakSize, peak, currentSize) {
                break
            }
        }
        
    default:
        // Pool is full, destroy object
        top.destroyObject(obj)
    }
}

// resetObject resets object state
func (top *TypedObjectPool) resetObject(obj interface{}) error {
    defer func() {
        if r := recover(); r != nil {
            // Reset panicked, object is invalid
        }
    }()
    
    top.resetter(obj)
    return nil
}

// destroyObject destroys an object
func (top *TypedObjectPool) destroyObject(obj interface{}) {
    if top.destructor != nil {
        top.destructor(obj)
    }
    atomic.AddInt64(&top.stats.Destroyed, 1)
}

// updateHitRate updates pool hit rate
func (top *TypedObjectPool) updateHitRate(hit bool) {
    retrieved := atomic.LoadInt64(&top.stats.Retrieved)
    if retrieved == 0 {
        return
    }
    
    if hit {
        // Simplified hit rate calculation
        top.stats.HitRate = float64(retrieved-1) / float64(retrieved)
    } else {
        top.stats.MissRate = float64(1) / float64(retrieved)
    }
}

// healthCheck performs periodic pool health checks
func (top *TypedObjectPool) healthCheck() {
    ticker := time.NewTicker(top.config.HealthCheckPeriod)
    defer ticker.Stop()
    
    for range ticker.C {
        top.performHealthCheck()
    }
}

// performHealthCheck performs a health check on the pool
func (top *TypedObjectPool) performHealthCheck() {
    currentSize := atomic.LoadInt32(&top.stats.CurrentSize)
    
    // Shrink pool if it's too large and usage is low
    if currentSize > int32(top.config.MinSize) {
        hitRate := top.stats.HitRate
        if hitRate < top.config.ShrinkThreshold {
            top.shrinkPool()
        }
    }
    
    // Check for expired objects (simplified)
    top.removeExpiredObjects()
}

// shrinkPool reduces pool size
func (top *TypedObjectPool) shrinkPool() {
    targetReduction := int32(float64(atomic.LoadInt32(&top.stats.CurrentSize)) * 0.2) // Reduce by 20%
    
    for i := int32(0); i < targetReduction; i++ {
        select {
        case obj := <-top.pool:
            top.destroyObject(obj)
            atomic.AddInt32(&top.stats.CurrentSize, -1)
        default:
            return // No more objects to remove
        }
    }
}

// removeExpiredObjects removes expired objects from the pool
func (top *TypedObjectPool) removeExpiredObjects() {
    // This is a simplified implementation
    // In practice, you'd track object creation times
    if time.Since(top.lifecycle.CreatedAt) > top.config.MaxLifetime {
        // Object has expired, remove from pool
        select {
        case obj := <-top.pool:
            top.destroyObject(obj)
            atomic.AddInt32(&top.stats.CurrentSize, -1)
        default:
            return
        }
    }
}

// GetStatistics returns pool statistics
func (top *TypedObjectPool) GetStatistics() PoolStatistics {
    top.mu.RLock()
    defer top.mu.RUnlock()
    
    stats := *top.stats
    
    // Calculate derived metrics
    total := stats.Retrieved + stats.Created
    if total > 0 {
        stats.HitRate = float64(stats.Retrieved) / float64(total)
        stats.MissRate = float64(stats.Created) / float64(total)
    }
    
    return stats
}

// Clear clears the pool
func (top *TypedObjectPool) Clear() {
    top.mu.Lock()
    defer top.mu.Unlock()
    
    // Drain and destroy all objects
    for {
        select {
        case obj := <-top.pool:
            top.destroyObject(obj)
            atomic.AddInt32(&top.stats.CurrentSize, -1)
        default:
            return
        }
    }
}

// DefaultObjectFactory implements ObjectFactory
type DefaultObjectFactory struct{}

// NewDefaultObjectFactory creates a new default object factory
func NewDefaultObjectFactory() *DefaultObjectFactory {
    return &DefaultObjectFactory{}
}

// CreateObject creates a new object of the given type
func (dof *DefaultObjectFactory) CreateObject(objectType reflect.Type) interface{} {
    // Handle pointer types
    if objectType.Kind() == reflect.Ptr {
        elem := objectType.Elem()
        value := reflect.New(elem)
        return value.Interface()
    }
    
    // Handle non-pointer types
    value := reflect.New(objectType).Elem()
    return value.Interface()
}

// ValidateObject validates an object
func (dof *DefaultObjectFactory) ValidateObject(obj interface{}) bool {
    return obj != nil
}

// ResetObject resets an object to its initial state
func (dof *DefaultObjectFactory) ResetObject(obj interface{}) error {
    if obj == nil {
        return fmt.Errorf("cannot reset nil object")
    }
    
    value := reflect.ValueOf(obj)
    
    // Handle pointer types
    if value.Kind() == reflect.Ptr {
        if value.IsNil() {
            return fmt.Errorf("cannot reset nil pointer")
        }
        value = value.Elem()
    }
    
    // Reset fields to zero values
    return dof.resetValue(value)
}

// resetValue recursively resets a reflect.Value
func (dof *DefaultObjectFactory) resetValue(value reflect.Value) error {
    if !value.CanSet() {
        return nil
    }
    
    switch value.Kind() {
    case reflect.Struct:
        for i := 0; i < value.NumField(); i++ {
            field := value.Field(i)
            if field.CanSet() {
                dof.resetValue(field)
            }
        }
    case reflect.Slice:
        if !value.IsNil() {
            value.SetLen(0)
        }
    case reflect.Map:
        if !value.IsNil() {
            value.Set(reflect.MakeMap(value.Type()))
        }
    case reflect.Chan:
        if !value.IsNil() {
            // Cannot reset channels safely
        }
    case reflect.Ptr:
        value.Set(reflect.Zero(value.Type()))
    default:
        value.Set(reflect.Zero(value.Type()))
    }
    
    return nil
}

// DestroyObject destroys an object
func (dof *DefaultObjectFactory) DestroyObject(obj interface{}) error {
    // Perform any necessary cleanup
    if closer, ok := obj.(interface{ Close() error }); ok {
        return closer.Close()
    }
    
    return nil
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(config ReuseConfig) *LifecycleManager {
    lifecycleConfig := LifecycleConfig{
        TrackObjects:    config.EnableLifecycle,
        MaxAge:          config.MaxObjectAge,
        MaxUseCount:     config.MaxObjectUseCount,
        CleanupInterval: time.Minute,
        EnableProfiling: true,
    }
    
    lm := &LifecycleManager{
        tracking: make(map[uintptr]*ObjectLifecycle),
        cleanup:  make(chan *ObjectLifecycle, 1000),
        config:   lifecycleConfig,
    }
    
    if lifecycleConfig.TrackObjects {
        go lm.cleanupLoop()
        lm.running = true
    }
    
    return lm
}

// TrackObject starts tracking an object
func (lm *LifecycleManager) TrackObject(obj interface{}) {
    if !lm.config.TrackObjects || obj == nil {
        return
    }
    
    objectID := uintptr(unsafe.Pointer(&obj))
    
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    lifecycle := &ObjectLifecycle{
        ObjectID:   objectID,
        ObjectType: reflect.TypeOf(obj),
        CreatedAt:  time.Now(),
        LastUsed:   time.Now(),
        UseCount:   1,
        State:      ObjectInUse,
        Metadata:   make(map[string]interface{}),
    }
    
    lm.tracking[objectID] = lifecycle
}

// UpdateObject updates object tracking information
func (lm *LifecycleManager) UpdateObject(obj interface{}) {
    if !lm.config.TrackObjects || obj == nil {
        return
    }
    
    objectID := uintptr(unsafe.Pointer(&obj))
    
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    if lifecycle, exists := lm.tracking[objectID]; exists {
        lifecycle.LastUsed = time.Now()
        lifecycle.UseCount++
        lifecycle.State = ObjectReturned
        lifecycle.TotalLifetime = time.Since(lifecycle.CreatedAt)
        
        // Check for expiration
        if lm.shouldExpire(lifecycle) {
            lifecycle.State = ObjectExpired
            select {
            case lm.cleanup <- lifecycle:
            default:
                // Cleanup channel full
            }
        }
    }
}

// shouldExpire checks if an object should expire
func (lm *LifecycleManager) shouldExpire(lifecycle *ObjectLifecycle) bool {
    if lm.config.MaxAge > 0 && time.Since(lifecycle.CreatedAt) > lm.config.MaxAge {
        return true
    }
    
    if lm.config.MaxUseCount > 0 && lifecycle.UseCount >= int64(lm.config.MaxUseCount) {
        return true
    }
    
    return false
}

// cleanupLoop handles lifecycle cleanup
func (lm *LifecycleManager) cleanupLoop() {
    ticker := time.NewTicker(lm.config.CleanupInterval)
    defer ticker.Stop()
    
    for lm.running {
        select {
        case lifecycle := <-lm.cleanup:
            lm.processExpiredObject(lifecycle)
            
        case <-ticker.C:
            lm.performPeriodicCleanup()
        }
    }
}

// processExpiredObject processes an expired object
func (lm *LifecycleManager) processExpiredObject(lifecycle *ObjectLifecycle) {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    delete(lm.tracking, lifecycle.ObjectID)
    lifecycle.State = ObjectDestroyed
}

// performPeriodicCleanup performs periodic cleanup of tracked objects
func (lm *LifecycleManager) performPeriodicCleanup() {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    now := time.Now()
    for objectID, lifecycle := range lm.tracking {
        if lm.shouldExpire(lifecycle) {
            lifecycle.State = ObjectExpired
            delete(lm.tracking, objectID)
            
            select {
            case lm.cleanup <- lifecycle:
            default:
                // Cleanup channel full
            }
        }
    }
}

// GetLifecycleStats returns lifecycle statistics
func (lm *LifecycleManager) GetLifecycleStats() map[string]interface{} {
    lm.mu.RLock()
    defer lm.mu.RUnlock()
    
    stats := map[string]interface{}{
        "total_tracked":    len(lm.tracking),
        "active_objects":   0,
        "expired_objects":  0,
        "average_lifetime": time.Duration(0),
        "average_use_count": float64(0),
    }
    
    totalLifetime := time.Duration(0)
    totalUseCount := int64(0)
    activeCount := 0
    expiredCount := 0
    
    for _, lifecycle := range lm.tracking {
        totalLifetime += lifecycle.TotalLifetime
        totalUseCount += lifecycle.UseCount
        
        switch lifecycle.State {
        case ObjectInUse, ObjectReturned:
            activeCount++
        case ObjectExpired, ObjectDestroyed:
            expiredCount++
        }
    }
    
    if len(lm.tracking) > 0 {
        stats["average_lifetime"] = totalLifetime / time.Duration(len(lm.tracking))
        stats["average_use_count"] = float64(totalUseCount) / float64(len(lm.tracking))
    }
    
    stats["active_objects"] = activeCount
    stats["expired_objects"] = expiredCount
    
    return stats
}

// NewReuseMonitor creates a new reuse monitor
func NewReuseMonitor() *ReuseMonitor {
    return &ReuseMonitor{
        events:     make(chan ReuseEvent, 10000),
        patterns:   make(map[string]*ReusePattern),
        alerts:     NewReuseAlerting(),
        collectors: make([]ReuseCollector, 0),
    }
}

// RecordEvent records a reuse event
func (rm *ReuseMonitor) RecordEvent(event ReuseEvent) {
    if !rm.running {
        return
    }
    
    select {
    case rm.events <- event:
    default:
        // Event channel full, drop event
    }
}

// Start starts the reuse monitor
func (rm *ReuseMonitor) Start() error {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    if rm.running {
        return fmt.Errorf("monitor already running")
    }
    
    rm.running = true
    go rm.monitorLoop()
    
    return nil
}

// monitorLoop processes reuse events
func (rm *ReuseMonitor) monitorLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for rm.running {
        select {
        case event := <-rm.events:
            rm.processEvent(event)
            
        case <-ticker.C:
            rm.analyzePatterns()
            rm.checkAlerts()
        }
    }
}

// processEvent processes a single reuse event
func (rm *ReuseMonitor) processEvent(event ReuseEvent) {
    // Update patterns
    patternKey := event.ObjectType.String()
    
    rm.mu.Lock()
    pattern, exists := rm.patterns[patternKey]
    if !exists {
        pattern = &ReusePattern{
            Name:       patternKey,
            ObjectType: event.ObjectType,
            Frequency:  0,
        }
        rm.patterns[patternKey] = pattern
    }
    pattern.Frequency++
    rm.mu.Unlock()
    
    // Notify collectors
    for _, collector := range rm.collectors {
        collector.CollectEvent(event)
    }
}

// analyzePatterns analyzes reuse patterns
func (rm *ReuseMonitor) analyzePatterns() {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    for _, pattern := range rm.patterns {
        // Calculate reuse efficiency
        if pattern.Frequency > 0 {
            pattern.Efficiency = calculateReuseEfficiency(pattern)
            pattern.Optimization = suggestOptimization(pattern)
        }
    }
}

// calculateReuseEfficiency calculates reuse efficiency for a pattern
func calculateReuseEfficiency(pattern *ReusePattern) float64 {
    // Simplified efficiency calculation
    if pattern.Frequency > 100 {
        return 0.8 // High efficiency for frequently reused objects
    } else if pattern.Frequency > 10 {
        return 0.5 // Medium efficiency
    }
    return 0.2 // Low efficiency
}

// suggestOptimization suggests optimization for a pattern
func suggestOptimization(pattern *ReusePattern) string {
    if pattern.Efficiency < 0.3 {
        return "Consider increasing pool size or adjusting object lifecycle"
    } else if pattern.Efficiency < 0.6 {
        return "Monitor object usage patterns and optimize pool configuration"
    }
    return "Pattern shows good reuse efficiency"
}

// checkAlerts checks for alert conditions
func (rm *ReuseMonitor) checkAlerts() {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    for _, pattern := range rm.patterns {
        if pattern.Efficiency < rm.alerts.thresholds.MinEfficiency {
            alert := ReuseAlert{
                Type:       EfficiencyAlert,
                Severity:   WarningAlert,
                Message:    fmt.Sprintf("Low reuse efficiency for %s: %.2f", pattern.Name, pattern.Efficiency),
                ObjectType: pattern.ObjectType,
                Timestamp:  time.Now(),
            }
            rm.alerts.SendAlert(alert)
        }
    }
}

// NewReuseAlerting creates a new reuse alerting system
func NewReuseAlerting() *ReuseAlerting {
    return &ReuseAlerting{
        thresholds: ReuseThresholds{
            MinReuseRate:  0.7,
            MaxMissRate:   0.3,
            MaxPoolSize:   10000,
            MaxObjectAge:  time.Hour,
            MinEfficiency: 0.5,
        },
        alerts:   make(chan ReuseAlert, 1000),
        handlers: make([]ReuseAlertHandler, 0),
    }
}

// SendAlert sends a reuse alert
func (ra *ReuseAlerting) SendAlert(alert ReuseAlert) {
    select {
    case ra.alerts <- alert:
    default:
        // Alert channel full
    }
    
    for _, handler := range ra.handlers {
        go handler.HandleAlert(alert)
    }
}

// GetMetrics returns overall reuse metrics
func (rm *ReuseManager) GetMetrics() ReuseMetrics {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    metrics := *rm.metrics
    
    // Aggregate metrics from all pools
    totalRetrieved := int64(0)
    totalCreated := int64(0)
    totalReturned := int64(0)
    
    for _, pool := range rm.pools {
        stats := pool.GetStatistics()
        totalRetrieved += stats.Retrieved
        totalCreated += stats.Created
        totalReturned += stats.Returned
    }
    
    if totalRetrieved+totalCreated > 0 {
        metrics.GlobalReuseRate = float64(totalRetrieved) / float64(totalRetrieved+totalCreated)
    }
    
    metrics.TotalReuses = totalReturned
    metrics.AllocationsSaved = totalRetrieved
    
    return metrics
}

// Example usage and specialized object pools
func ExampleObjectReuse() {
    // Create reuse manager
    config := ReuseConfig{
        MaxPoolSize:        1000,
        InitialPoolSize:    10,
        MaxObjectAge:       time.Hour,
        MaxObjectUseCount:  100,
        EnableLifecycle:    true,
        EnableMonitoring:   true,
        GCTriggerThreshold: 0.8,
        ValidationEnabled:  true,
        ResetOnReturn:      true,
    }
    
    manager := NewReuseManager(config)
    
    // Example: Reusing byte buffers
    type Buffer struct {
        Data []byte
    }
    
    bufferType := reflect.TypeOf(Buffer{})
    
    // Get a buffer
    buffer := manager.GetObject(bufferType).(*Buffer)
    
    // Use buffer...
    buffer.Data = make([]byte, 1024)
    
    // Return buffer to pool
    manager.ReturnObject(buffer)
    
    // Get metrics
    metrics := manager.GetMetrics()
    fmt.Printf("Global reuse rate: %.2f%%\n", metrics.GlobalReuseRate*100)
    fmt.Printf("Total pools: %d\n", metrics.TotalPools)
    fmt.Printf("Allocations saved: %d\n", metrics.AllocationsSaved)
}

// SpecializedStringBuilderPool example
type StringBuilderPool struct {
    pool sync.Pool
}

func NewStringBuilderPool() *StringBuilderPool {
    return &StringBuilderPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &strings.Builder{}
            },
        },
    }
}

func (sbp *StringBuilderPool) Get() *strings.Builder {
    return sbp.pool.Get().(*strings.Builder)
}

func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
    sb.Reset()
    sbp.pool.Put(sb)
}
```

## Advanced Pooling Strategies

Specialized pooling strategies for different object types and usage patterns.

### Type-Specific Pools

Optimized pools for specific object types with custom reset and validation logic.

### Hierarchical Pooling

Multi-level pooling strategies for complex object hierarchies.

### Adaptive Pooling

Dynamic pool sizing based on usage patterns and performance metrics.

## Performance Optimization

Advanced techniques for optimizing object reuse performance.

### Memory Layout Optimization

Optimizing object layout for better cache performance during reuse.

### Batch Operations

Batching object allocation and deallocation for improved performance.

### Lock-Free Techniques

Implementing lock-free object pools for high-concurrency scenarios.

## Best Practices

1. **Reset Objects**: Always reset object state before reuse
2. **Validate Objects**: Validate objects before returning to pools
3. **Monitor Usage**: Track pool efficiency and usage patterns
4. **Size Pools Appropriately**: Balance memory usage and hit rates
5. **Handle Failures**: Gracefully handle reset and validation failures
6. **Lifecycle Management**: Track object lifecycles to prevent memory leaks
7. **Type Safety**: Use type-safe pools to prevent runtime errors
8. **Profile Performance**: Regularly measure reuse effectiveness

## Summary

Object reuse is essential for high-performance Go applications:

1. **Pooling**: Use appropriate pooling strategies for different object types
2. **Lifecycle Management**: Track object lifecycles to ensure proper cleanup
3. **Monitoring**: Monitor reuse patterns and efficiency
4. **Optimization**: Apply specialized optimization techniques
5. **Validation**: Ensure object integrity throughout the reuse cycle

These techniques enable developers to minimize allocation pressure and maximize application performance through efficient object reuse patterns.
