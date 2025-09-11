# Memory Pools Mastery

## 🎯 Learning Objectives

After completing this tutorial, you will be able to:

- **Understand memory pool fundamentals** and when to use them
- **Implement sync.Pool efficiently** for common use cases
- **Build custom pool solutions** for specialized requirements
- **Measure pool performance** and optimize for your workload
- **Avoid common pooling pitfalls** that reduce effectiveness
- **Deploy production-ready pooling** with monitoring and metrics

## 📚 What You'll Build

Throughout this tutorial, you'll create:

1. **Basic Object Pool** - Simple reusable object management
2. **High-Performance Buffer Pool** - Optimized for byte operations
3. **Custom Pool Manager** - Multi-pool coordination system
4. **Performance Monitor** - Real-time pool efficiency tracking

### 🔍 Prerequisites

- Understanding of Go memory allocation
- Basic knowledge of garbage collection
- Familiarity with Go's sync package
- Performance optimization concepts

## 🚀 Why Memory Pools Matter

Memory pools reduce garbage collection pressure by reusing objects instead of constantly allocating new ones. This is especially important for high-throughput applications.

### The Problem: Allocation Pressure

```go
// ❌ Poor: Creates new objects constantly
func processRequests() {
    for request := range requests {
        buffer := make([]byte, 4096)    // New allocation
        response := &Response{}         // New allocation
        
        process(request, buffer, response)
        
        // Objects become garbage after function returns
    }
}
```

### The Solution: Object Reuse

```go
// ✅ Good: Reuses objects via pools
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{}
    },
}

func processRequestsWithPools() {
    for request := range requests {
        buffer := bufferPool.Get().([]byte)
        response := responsePool.Get().(*Response)
        
        // Reset for reuse
        buffer = buffer[:0]
        response.Reset()
        
        process(request, buffer, response)
        
        // Return to pools for reuse
        bufferPool.Put(buffer)
        responsePool.Put(response)
    }
}
```

## 🛠️ Hands-on: Your First Memory Pool

Let's start with a practical example that demonstrates immediate performance benefits:

### Step 1: Basic Buffer Pool

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// BufferPool demonstrates basic pooling concepts
type BufferPool struct {
    pool sync.Pool
    size int
}

// NewBufferPool creates a buffer pool with fixed-size buffers
func NewBufferPool(bufferSize int) *BufferPool {
    return &BufferPool{
        size: bufferSize,
        pool: sync.Pool{
            New: func() interface{} {
                // This function is called when pool is empty
                fmt.Printf("🔨 Creating new buffer of size %d\n", bufferSize)
                return make([]byte, bufferSize)
            },
        },
    }
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() []byte {
    return bp.pool.Get().([]byte)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
    // Reset buffer length but keep capacity
    buf = buf[:0]
    bp.pool.Put(buf)
}

// Demo function showing pool effectiveness
func demoBufferPool() {
    fmt.Println("🎯 Buffer Pool Demo")
    fmt.Println("==================")
    
    pool := NewBufferPool(1024)
    
    // First use - will create new buffers
    fmt.Println("\n📝 First batch (creates new buffers):")
    buffers := make([][]byte, 3)
    for i := 0; i < 3; i++ {
        buffers[i] = pool.Get()
        fmt.Printf("   Got buffer %d\n", i+1)
    }
    
    // Return buffers to pool
    fmt.Println("\n♻️  Returning buffers to pool:")
    for i, buf := range buffers {
        pool.Put(buf)
        fmt.Printf("   Returned buffer %d\n", i+1)
    }
    
    // Second use - will reuse existing buffers
    fmt.Println("\n🔄 Second batch (reuses buffers):")
    for i := 0; i < 3; i++ {
        buf := pool.Get()
        fmt.Printf("   Got buffer %d (reused!)\n", i+1)
        pool.Put(buf)
    }
}

func main() {
    demoBufferPool()
}
```

> 💡 **Try This**: Run the demo and observe how only the first batch creates new buffers. The second batch reuses existing ones, avoiding allocations.

### Step 2: Performance Comparison

Let's measure the performance impact of pooling:

```go
package main

import (
    "sync"
    "testing"
)

// Benchmark without pooling
func BenchmarkWithoutPool(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        // Simulate typical workload
        buffer := make([]byte, 4096)
        
        // Use buffer (prevents optimization)
        buffer[0] = byte(i)
        _ = buffer
    }
}

// Benchmark with sync.Pool
func BenchmarkWithPool(b *testing.B) {
    b.ReportAllocs()
    
    pool := sync.Pool{
        New: func() interface{} {
            return make([]byte, 4096)
        },
    }
    
    for i := 0; i < b.N; i++ {
        buffer := pool.Get().([]byte)
        
        // Use buffer
        buffer[0] = byte(i)
        
        // Return to pool
        pool.Put(buffer)
    }
}

// Benchmark with pre-allocated slice
func BenchmarkWithPrealloc(b *testing.B) {
    b.ReportAllocs()
    
    buffer := make([]byte, 4096)
    
    for i := 0; i < b.N; i++ {
        // Reset and reuse same buffer
        buffer = buffer[:0]
        buffer = append(buffer, byte(i))
    }
}
```

**Expected Results:**
- **Without Pool**: High allocation count, high memory usage
- **With Pool**: Reduced allocations, lower GC pressure  
- **Pre-allocated**: Minimal allocations, best performance for single-threaded use

## 🔧 Advanced Pool Implementations

Now let's build more sophisticated pooling solutions:
    ObjectsInUse     int32
    AllocationsSaved int64
    MemorySaved      int64
    LastAccess       time.Time
}

// PoolFactory creates pool instances
type PoolFactory interface {
    CreatePool(name string, config PoolConfig) Pool
    CreateTypedPool(name string, factory func() interface{}, config PoolConfig) Pool
}

// PoolMonitor monitors pool performance
type PoolMonitor struct {
    events     chan PoolEvent
    collectors []PoolCollector
    alerting   *PoolAlerting
    running    bool
    mu         sync.RWMutex
}

// PoolEvent represents a pool operation event
type PoolEvent struct {
    PoolName    string
    EventType   PoolEventType
    Size        int64
    Duration    time.Duration
    Timestamp   time.Time
    ObjectType  string
}

// PoolEventType defines pool event types
type PoolEventType int

const (
    PoolGet PoolEventType = iota
    PoolPut
    PoolGrow
    PoolShrink
    PoolClear
    PoolResize
)

// PoolCollector collects pool metrics
type PoolCollector interface {
    CollectEvent(event PoolEvent)
    GetMetrics() map[string]interface{}
    Reset()
}

// PoolAlerting provides alerting for pool issues
type PoolAlerting struct {
    thresholds   PoolThresholds
    alerts       chan PoolAlert
    handlers     []PoolAlertHandler
}

// PoolThresholds defines alerting thresholds
type PoolThresholds struct {
    MinHitRate       float64
    MaxMissRate      float64
    MaxPoolSize      int32
    MaxMemoryUsage   int64
}

// PoolAlert represents a pool alert
type PoolAlert struct {
    Type        PoolAlertType
    Severity    AlertSeverity
    Message     string
    PoolName    string
    Metrics     map[string]interface{}
    Timestamp   time.Time
}

// PoolAlertType defines alert types
type PoolAlertType int

const (
    HitRateAlert PoolAlertType = iota
    MissRateAlert
    MemoryAlert
    SizeAlert
)

// AlertSeverity defines alert severity
type AlertSeverity int

const (
    InfoSeverity AlertSeverity = iota
    WarningSeverity
    ErrorSeverity
    CriticalSeverity
)

// PoolAlertHandler handles pool alerts
type PoolAlertHandler interface {
    HandleAlert(alert PoolAlert) error
}

// NewPoolManager creates a new pool manager
func NewPoolManager(config PoolConfig) *PoolManager {
    return &PoolManager{
        pools:   make(map[string]Pool),
        factory: NewDefaultPoolFactory(),
        metrics: &PoolMetrics{},
        monitor: NewPoolMonitor(),
        config:  config,
    }
}

// GetPool gets or creates a pool
func (pm *PoolManager) GetPool(name string) Pool {
    pm.mu.RLock()
    if pool, exists := pm.pools[name]; exists {
        pm.mu.RUnlock()
        return pool
    }
    pm.mu.RUnlock()
    
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    // Double-check after acquiring write lock
    if pool, exists := pm.pools[name]; exists {
        return pool
    }
    
    // Create new pool
    pool := pm.factory.CreatePool(name, pm.config)
    pm.pools[name] = pool
    atomic.AddInt32(&pm.metrics.PoolCount, 1)
    
    return pool
}

// GetTypedPool gets or creates a typed pool
func (pm *PoolManager) GetTypedPool(name string, factory func() interface{}) Pool {
    pm.mu.RLock()
    if pool, exists := pm.pools[name]; exists {
        pm.mu.RUnlock()
        return pool
    }
    pm.mu.RUnlock()
    
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    if pool, exists := pm.pools[name]; exists {
        return pool
    }
    
    pool := pm.factory.CreateTypedPool(name, factory, pm.config)
    pm.pools[name] = pool
    atomic.AddInt32(&pm.metrics.PoolCount, 1)
    
    return pool
}

// OptimizedSyncPool wraps sync.Pool with metrics and optimization
type OptimizedSyncPool struct {
    pool        *sync.Pool
    name        string
    config      PoolConfig
    stats       PoolStats
    factory     func() interface{}
    validator   func(interface{}) bool
    resetter    func(interface{})
    monitor     *PoolMonitor
    mu          sync.RWMutex
}

// NewOptimizedSyncPool creates a new optimized sync.Pool
func NewOptimizedSyncPool(name string, factory func() interface{}, config PoolConfig) *OptimizedSyncPool {
    pool := &OptimizedSyncPool{
        name:    name,
        config:  config,
        factory: factory,
        stats: PoolStats{
            PoolName:   name,
            MaxSize:    int32(config.MaxPoolSize),
            LastAccess: time.Now(),
        },
    }
    
    pool.pool = &sync.Pool{
        New: func() interface{} {
            atomic.AddInt64(&pool.stats.Misses, 1)
            atomic.AddInt32(&pool.stats.CurrentSize, 1)
            obj := pool.factory()
            
            if pool.monitor != nil {
                event := PoolEvent{
                    PoolName:   pool.name,
                    EventType:  PoolGet,
                    Size:       int64(getObjectSize(obj)),
                    Timestamp:  time.Now(),
                    ObjectType: getObjectType(obj),
                }
                pool.monitor.RecordEvent(event)
            }
            
            return obj
        },
    }
    
    return pool
}

// Get retrieves an object from the pool
func (osp *OptimizedSyncPool) Get() interface{} {
    atomic.AddInt64(&osp.stats.Gets, 1)
    osp.stats.LastAccess = time.Now()
    
    obj := osp.pool.Get()
    
    // Validate object if validator is set
    if osp.validator != nil && !osp.validator(obj) {
        // Object is invalid, create new one
        atomic.AddInt64(&osp.stats.Misses, 1)
        return osp.factory()
    }
    
    atomic.AddInt64(&osp.stats.Hits, 1)
    
    if osp.monitor != nil {
        event := PoolEvent{
            PoolName:   osp.name,
            EventType:  PoolGet,
            Size:       int64(getObjectSize(obj)),
            Timestamp:  time.Now(),
            ObjectType: getObjectType(obj),
        }
        osp.monitor.RecordEvent(event)
    }
    
    return obj
}

// Put returns an object to the pool
func (osp *OptimizedSyncPool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&osp.stats.Puts, 1)
    osp.stats.LastAccess = time.Now()
    
    // Reset object if resetter is provided
    if osp.resetter != nil {
        osp.resetter(obj)
    }
    
    // Validate object before putting back
    if osp.validator != nil && !osp.validator(obj) {
        return // Don't put invalid objects back
    }
    
    osp.pool.Put(obj)
    
    if osp.monitor != nil {
        event := PoolEvent{
            PoolName:   osp.name,
            EventType:  PoolPut,
            Size:       int64(getObjectSize(obj)),
            Timestamp:  time.Now(),
            ObjectType: getObjectType(obj),
        }
        osp.monitor.RecordEvent(event)
    }
}

// Size returns the current pool size (estimated)
func (osp *OptimizedSyncPool) Size() int {
    return int(atomic.LoadInt32(&osp.stats.CurrentSize))
}

// Clear clears the pool
func (osp *OptimizedSyncPool) Clear() {
    // sync.Pool doesn't support clearing, so we recreate it
    osp.mu.Lock()
    defer osp.mu.Unlock()
    
    oldPool := osp.pool
    osp.pool = &sync.Pool{New: oldPool.New}
    atomic.StoreInt32(&osp.stats.CurrentSize, 0)
    
    if osp.monitor != nil {
        event := PoolEvent{
            PoolName:  osp.name,
            EventType: PoolClear,
            Timestamp: time.Now(),
        }
        osp.monitor.RecordEvent(event)
    }
}

// GetMetrics returns pool statistics
func (osp *OptimizedSyncPool) GetMetrics() PoolStats {
    osp.mu.RLock()
    defer osp.mu.RUnlock()
    
    stats := osp.stats
    
    // Calculate hit rate
    totalRequests := stats.Gets
    if totalRequests > 0 {
        stats.AllocationsSaved = stats.Hits
        stats.MemorySaved = stats.Hits * int64(getObjectSize(osp.factory()))
    }
    
    return stats
}

// Resize is not applicable for sync.Pool
func (osp *OptimizedSyncPool) Resize(newSize int) error {
    return fmt.Errorf("resize not supported for sync.Pool")
}

// SetValidator sets object validator
func (osp *OptimizedSyncPool) SetValidator(validator func(interface{}) bool) {
    osp.validator = validator
}

// SetResetter sets object resetter
func (osp *OptimizedSyncPool) SetResetter(resetter func(interface{})) {
    osp.resetter = resetter
}

// ChannelPool implements a channel-based object pool
type ChannelPool struct {
    name      string
    objects   chan interface{}
    factory   func() interface{}
    config    PoolConfig
    stats     PoolStats
    monitor   *PoolMonitor
    mu        sync.RWMutex
}

// NewChannelPool creates a new channel-based pool
func NewChannelPool(name string, factory func() interface{}, config PoolConfig) *ChannelPool {
    pool := &ChannelPool{
        name:    name,
        objects: make(chan interface{}, config.MaxPoolSize),
        factory: factory,
        config:  config,
        stats: PoolStats{
            PoolName:   name,
            MaxSize:    int32(config.MaxPoolSize),
            LastAccess: time.Now(),
        },
    }
    
    // Pre-populate pool if initial size is specified
    for i := 0; i < config.InitialSize; i++ {
        pool.objects <- factory()
        atomic.AddInt32(&pool.stats.CurrentSize, 1)
    }
    
    return pool
}

// Get retrieves an object from the channel pool
func (cp *ChannelPool) Get() interface{} {
    atomic.AddInt64(&cp.stats.Gets, 1)
    cp.stats.LastAccess = time.Now()
    
    select {
    case obj := <-cp.objects:
        atomic.AddInt64(&cp.stats.Hits, 1)
        atomic.AddInt32(&cp.stats.CurrentSize, -1)
        
        if cp.monitor != nil {
            event := PoolEvent{
                PoolName:   cp.name,
                EventType:  PoolGet,
                Size:       int64(getObjectSize(obj)),
                Timestamp:  time.Now(),
                ObjectType: getObjectType(obj),
            }
            cp.monitor.RecordEvent(event)
        }
        
        return obj
        
    default:
        // Pool is empty, create new object
        atomic.AddInt64(&cp.stats.Misses, 1)
        obj := cp.factory()
        
        if cp.monitor != nil {
            event := PoolEvent{
                PoolName:   cp.name,
                EventType:  PoolGet,
                Size:       int64(getObjectSize(obj)),
                Timestamp:  time.Now(),
                ObjectType: getObjectType(obj),
            }
            cp.monitor.RecordEvent(event)
        }
        
        return obj
    }
}

// Put returns an object to the channel pool
func (cp *ChannelPool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&cp.stats.Puts, 1)
    cp.stats.LastAccess = time.Now()
    
    select {
    case cp.objects <- obj:
        atomic.AddInt32(&cp.stats.CurrentSize, 1)
        
        if cp.monitor != nil {
            event := PoolEvent{
                PoolName:   cp.name,
                EventType:  PoolPut,
                Size:       int64(getObjectSize(obj)),
                Timestamp:  time.Now(),
                ObjectType: getObjectType(obj),
            }
            cp.monitor.RecordEvent(event)
        }
        
    default:
        // Pool is full, discard object
        // Optionally trigger pool resize or alert
    }
}

// Size returns the current pool size
func (cp *ChannelPool) Size() int {
    return int(atomic.LoadInt32(&cp.stats.CurrentSize))
}

// Clear clears the channel pool
func (cp *ChannelPool) Clear() {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    // Drain the channel
    for {
        select {
        case <-cp.objects:
            atomic.AddInt32(&cp.stats.CurrentSize, -1)
        default:
            atomic.StoreInt32(&cp.stats.CurrentSize, 0)
            
            if cp.monitor != nil {
                event := PoolEvent{
                    PoolName:  cp.name,
                    EventType: PoolClear,
                    Timestamp: time.Now(),
                }
                cp.monitor.RecordEvent(event)
            }
            return
        }
    }
}

// GetMetrics returns channel pool statistics
func (cp *ChannelPool) GetMetrics() PoolStats {
    cp.mu.RLock()
    defer cp.mu.RUnlock()
    
    stats := cp.stats
    stats.ObjectsInUse = int32(cap(cp.objects)) - stats.CurrentSize
    
    totalRequests := stats.Gets
    if totalRequests > 0 {
        stats.AllocationsSaved = stats.Hits
        stats.MemorySaved = stats.Hits * int64(getObjectSize(cp.factory()))
    }
    
    return stats
}

// Resize resizes the channel pool
func (cp *ChannelPool) Resize(newSize int) error {
    if newSize <= 0 {
        return fmt.Errorf("invalid pool size: %d", newSize)
    }
    
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    oldObjects := cp.objects
    cp.objects = make(chan interface{}, newSize)
    cp.config.MaxPoolSize = newSize
    cp.stats.MaxSize = int32(newSize)
    
    // Transfer existing objects
    transferred := 0
    for transferred < newSize {
        select {
        case obj := <-oldObjects:
            cp.objects <- obj
            transferred++
        default:
            break
        }
    }
    
    atomic.StoreInt32(&cp.stats.CurrentSize, int32(transferred))
    
    if cp.monitor != nil {
        event := PoolEvent{
            PoolName:  cp.name,
            EventType: PoolResize,
            Size:      int64(newSize),
            Timestamp: time.Now(),
        }
        cp.monitor.RecordEvent(event)
    }
    
    return nil
}

// LockFreePool implements a lock-free object pool using atomic operations
type LockFreePool struct {
    name      string
    head      unsafe.Pointer // *poolNode
    factory   func() interface{}
    config    PoolConfig
    stats     PoolStats
    monitor   *PoolMonitor
    size      int64
}

// poolNode represents a node in the lock-free pool
type poolNode struct {
    next   unsafe.Pointer // *poolNode
    object interface{}
}

// NewLockFreePool creates a new lock-free pool
func NewLockFreePool(name string, factory func() interface{}, config PoolConfig) *LockFreePool {
    return &LockFreePool{
        name:    name,
        factory: factory,
        config:  config,
        stats: PoolStats{
            PoolName:   name,
            MaxSize:    int32(config.MaxPoolSize),
            LastAccess: time.Now(),
        },
    }
}

// Get retrieves an object from the lock-free pool
func (lfp *LockFreePool) Get() interface{} {
    atomic.AddInt64(&lfp.stats.Gets, 1)
    lfp.stats.LastAccess = time.Now()
    
    for {
        head := atomic.LoadPointer(&lfp.head)
        if head == nil {
            // Pool is empty, create new object
            atomic.AddInt64(&lfp.stats.Misses, 1)
            obj := lfp.factory()
            
            if lfp.monitor != nil {
                event := PoolEvent{
                    PoolName:   lfp.name,
                    EventType:  PoolGet,
                    Size:       int64(getObjectSize(obj)),
                    Timestamp:  time.Now(),
                    ObjectType: getObjectType(obj),
                }
                lfp.monitor.RecordEvent(event)
            }
            
            return obj
        }
        
        node := (*poolNode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&lfp.head, head, next) {
            atomic.AddInt64(&lfp.stats.Hits, 1)
            atomic.AddInt64(&lfp.size, -1)
            
            obj := node.object
            
            if lfp.monitor != nil {
                event := PoolEvent{
                    PoolName:   lfp.name,
                    EventType:  PoolGet,
                    Size:       int64(getObjectSize(obj)),
                    Timestamp:  time.Now(),
                    ObjectType: getObjectType(obj),
                }
                lfp.monitor.RecordEvent(event)
            }
            
            return obj
        }
        // CAS failed, retry
    }
}

// Put returns an object to the lock-free pool
func (lfp *LockFreePool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&lfp.stats.Puts, 1)
    lfp.stats.LastAccess = time.Now()
    
    // Check pool size limit
    if atomic.LoadInt64(&lfp.size) >= int64(lfp.config.MaxPoolSize) {
        return // Pool is full
    }
    
    node := &poolNode{object: obj}
    
    for {
        head := atomic.LoadPointer(&lfp.head)
        atomic.StorePointer(&node.next, head)
        
        if atomic.CompareAndSwapPointer(&lfp.head, head, unsafe.Pointer(node)) {
            atomic.AddInt64(&lfp.size, 1)
            
            if lfp.monitor != nil {
                event := PoolEvent{
                    PoolName:   lfp.name,
                    EventType:  PoolPut,
                    Size:       int64(getObjectSize(obj)),
                    Timestamp:  time.Now(),
                    ObjectType: getObjectType(obj),
                }
                lfp.monitor.RecordEvent(event)
            }
            
            return
        }
        // CAS failed, retry
    }
}

// Size returns the current pool size
func (lfp *LockFreePool) Size() int {
    return int(atomic.LoadInt64(&lfp.size))
}

// Clear clears the lock-free pool
func (lfp *LockFreePool) Clear() {
    for {
        head := atomic.LoadPointer(&lfp.head)
        if head == nil {
            break
        }
        
        node := (*poolNode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&lfp.head, head, next) {
            atomic.AddInt64(&lfp.size, -1)
        }
    }
    
    if lfp.monitor != nil {
        event := PoolEvent{
            PoolName:  lfp.name,
            EventType: PoolClear,
            Timestamp: time.Now(),
        }
        lfp.monitor.RecordEvent(event)
    }
}

// GetMetrics returns lock-free pool statistics
func (lfp *LockFreePool) GetMetrics() PoolStats {
    stats := lfp.stats
    stats.CurrentSize = int32(atomic.LoadInt64(&lfp.size))
    
    totalRequests := stats.Gets
    if totalRequests > 0 {
        stats.AllocationsSaved = stats.Hits
        stats.MemorySaved = stats.Hits * int64(getObjectSize(lfp.factory()))
    }
    
    return stats
}

// Resize is not easily supported in lock-free pools
func (lfp *LockFreePool) Resize(newSize int) error {
    lfp.config.MaxPoolSize = newSize
    lfp.stats.MaxSize = int32(newSize)
    return nil
}

// ShardedPool implements a sharded pool for reduced contention
type ShardedPool struct {
    name      string
    shards    []Pool
    shardMask uint64
    config    PoolConfig
    stats     PoolStats
    monitor   *PoolMonitor
}

// NewShardedPool creates a new sharded pool
func NewShardedPool(name string, factory func() interface{}, config PoolConfig, shardCount int) *ShardedPool {
    if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
        shardCount = 16 // Default to 16 shards (power of 2)
    }
    
    shards := make([]Pool, shardCount)
    for i := 0; i < shardCount; i++ {
        shardConfig := config
        shardConfig.MaxPoolSize = config.MaxPoolSize / shardCount
        shards[i] = NewChannelPool(fmt.Sprintf("%s-shard-%d", name, i), factory, shardConfig)
    }
    
    return &ShardedPool{
        name:      name,
        shards:    shards,
        shardMask: uint64(shardCount - 1),
        config:    config,
        stats: PoolStats{
            PoolName:   name,
            MaxSize:    int32(config.MaxPoolSize),
            LastAccess: time.Now(),
        },
    }
}

// getShard returns the appropriate shard for the current goroutine
func (sp *ShardedPool) getShard() Pool {
    // Use goroutine ID for sharding (simplified)
    gid := getGoroutineID()
    shardIndex := gid & sp.shardMask
    return sp.shards[shardIndex]
}

// Get retrieves an object from the sharded pool
func (sp *ShardedPool) Get() interface{} {
    atomic.AddInt64(&sp.stats.Gets, 1)
    sp.stats.LastAccess = time.Now()
    
    shard := sp.getShard()
    obj := shard.Get()
    
    // Aggregate shard stats
    shardStats := shard.GetMetrics()
    atomic.AddInt64(&sp.stats.Hits, shardStats.Hits)
    atomic.AddInt64(&sp.stats.Misses, shardStats.Misses)
    
    if sp.monitor != nil {
        event := PoolEvent{
            PoolName:   sp.name,
            EventType:  PoolGet,
            Size:       int64(getObjectSize(obj)),
            Timestamp:  time.Now(),
            ObjectType: getObjectType(obj),
        }
        sp.monitor.RecordEvent(event)
    }
    
    return obj
}

// Put returns an object to the sharded pool
func (sp *ShardedPool) Put(obj interface{}) {
    if obj == nil {
        return
    }
    
    atomic.AddInt64(&sp.stats.Puts, 1)
    sp.stats.LastAccess = time.Now()
    
    shard := sp.getShard()
    shard.Put(obj)
    
    if sp.monitor != nil {
        event := PoolEvent{
            PoolName:   sp.name,
            EventType:  PoolPut,
            Size:       int64(getObjectSize(obj)),
            Timestamp:  time.Now(),
            ObjectType: getObjectType(obj),
        }
        sp.monitor.RecordEvent(event)
    }
}

// Size returns the total size across all shards
func (sp *ShardedPool) Size() int {
    total := 0
    for _, shard := range sp.shards {
        total += shard.Size()
    }
    return total
}

// Clear clears all shards
func (sp *ShardedPool) Clear() {
    for _, shard := range sp.shards {
        shard.Clear()
    }
    
    if sp.monitor != nil {
        event := PoolEvent{
            PoolName:  sp.name,
            EventType: PoolClear,
            Timestamp: time.Now(),
        }
        sp.monitor.RecordEvent(event)
    }
}

// GetMetrics aggregates metrics from all shards
func (sp *ShardedPool) GetMetrics() PoolStats {
    stats := sp.stats
    stats.CurrentSize = int32(sp.Size())
    
    totalHits := int64(0)
    totalMisses := int64(0)
    totalGets := int64(0)
    totalPuts := int64(0)
    
    for _, shard := range sp.shards {
        shardStats := shard.GetMetrics()
        totalHits += shardStats.Hits
        totalMisses += shardStats.Misses
        totalGets += shardStats.Gets
        totalPuts += shardStats.Puts
    }
    
    stats.Hits = totalHits
    stats.Misses = totalMisses
    stats.Gets = totalGets
    stats.Puts = totalPuts
    
    if totalGets > 0 {
        stats.AllocationsSaved = totalHits
    }
    
    return stats
}

// Resize resizes all shards proportionally
func (sp *ShardedPool) Resize(newSize int) error {
    shardSize := newSize / len(sp.shards)
    
    for _, shard := range sp.shards {
        if err := shard.Resize(shardSize); err != nil {
            return fmt.Errorf("failed to resize shard: %w", err)
        }
    }
    
    sp.config.MaxPoolSize = newSize
    sp.stats.MaxSize = int32(newSize)
    
    return nil
}

// DefaultPoolFactory implements the PoolFactory interface
type DefaultPoolFactory struct{}

// NewDefaultPoolFactory creates a new default pool factory
func NewDefaultPoolFactory() *DefaultPoolFactory {
    return &DefaultPoolFactory{}
}

// CreatePool creates a pool based on configuration
func (dpf *DefaultPoolFactory) CreatePool(name string, config PoolConfig) Pool {
    factory := func() interface{} {
        return make([]byte, 1024) // Default byte slice factory
    }
    
    return dpf.CreateTypedPool(name, factory, config)
}

// CreateTypedPool creates a typed pool based on configuration
func (dpf *DefaultPoolFactory) CreateTypedPool(name string, factory func() interface{}, config PoolConfig) Pool {
    switch config.PoolType {
    case SyncPool:
        return NewOptimizedSyncPool(name, factory, config)
    case ChannelPool:
        return NewChannelPool(name, factory, config)
    case LockFreePool:
        return NewLockFreePool(name, factory, config)
    case ShardedPool:
        return NewShardedPool(name, factory, config, 16)
    default:
        return NewOptimizedSyncPool(name, factory, config)
    }
}

// NewPoolMonitor creates a new pool monitor
func NewPoolMonitor() *PoolMonitor {
    return &PoolMonitor{
        events:     make(chan PoolEvent, 10000),
        collectors: make([]PoolCollector, 0),
        alerting:   NewPoolAlerting(),
    }
}

// RecordEvent records a pool event
func (pm *PoolMonitor) RecordEvent(event PoolEvent) {
    if !pm.running {
        return
    }
    
    select {
    case pm.events <- event:
    default:
        // Channel full, drop event
    }
}

// Start starts the pool monitor
func (pm *PoolMonitor) Start() error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    if pm.running {
        return fmt.Errorf("monitor already running")
    }
    
    pm.running = true
    go pm.monitorLoop()
    
    return nil
}

// monitorLoop processes pool events
func (pm *PoolMonitor) monitorLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for pm.running {
        select {
        case event := <-pm.events:
            pm.processEvent(event)
            
        case <-ticker.C:
            pm.checkAlerts()
        }
    }
}

// processEvent processes a single pool event
func (pm *PoolMonitor) processEvent(event PoolEvent) {
    for _, collector := range pm.collectors {
        collector.CollectEvent(event)
    }
}

// checkAlerts checks for alert conditions
func (pm *PoolMonitor) checkAlerts() {
    // Simplified alert checking
    for _, collector := range pm.collectors {
        metrics := collector.GetMetrics()
        
        if hitRate, ok := metrics["hit_rate"].(float64); ok {
            if hitRate < pm.alerting.thresholds.MinHitRate {
                alert := PoolAlert{
                    Type:      HitRateAlert,
                    Severity:  WarningSeverity,
                    Message:   fmt.Sprintf("Low hit rate: %.2f%%", hitRate*100),
                    Timestamp: time.Now(),
                }
                pm.alerting.SendAlert(alert)
            }
        }
    }
}

// NewPoolAlerting creates a new pool alerting system
func NewPoolAlerting() *PoolAlerting {
    return &PoolAlerting{
        thresholds: PoolThresholds{
            MinHitRate:     0.7,  // 70%
            MaxMissRate:    0.3,  // 30%
            MaxPoolSize:    10000,
            MaxMemoryUsage: 100 * 1024 * 1024, // 100MB
        },
        alerts:   make(chan PoolAlert, 1000),
        handlers: make([]PoolAlertHandler, 0),
    }
}

// SendAlert sends a pool alert
func (pa *PoolAlerting) SendAlert(alert PoolAlert) {
    select {
    case pa.alerts <- alert:
    default:
        // Alert channel full
    }
    
    for _, handler := range pa.handlers {
        go handler.HandleAlert(alert)
    }
}

// GetMetrics returns aggregated pool metrics
func (pm *PoolManager) GetMetrics() PoolMetrics {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    metrics := *pm.metrics
    
    totalGets := int64(0)
    totalPuts := int64(0)
    totalHits := int64(0)
    totalMisses := int64(0)
    
    for _, pool := range pm.pools {
        stats := pool.GetMetrics()
        totalGets += stats.Gets
        totalPuts += stats.Puts
        totalHits += stats.Hits
        totalMisses += stats.Misses
    }
    
    metrics.TotalGets = totalGets
    metrics.TotalPuts = totalPuts
    metrics.TotalHits = totalHits
    metrics.TotalMisses = totalMisses
    
    if totalGets > 0 {
        metrics.HitRate = float64(totalHits) / float64(totalGets)
    }
    
    return metrics
}

// Helper functions

// getObjectSize estimates object size (simplified)
func getObjectSize(obj interface{}) uintptr {
    if obj == nil {
        return 0
    }
    
    switch obj.(type) {
    case []byte:
        return uintptr(len(obj.([]byte)))
    case string:
        return uintptr(len(obj.(string)))
    default:
        return unsafe.Sizeof(obj)
    }
}

// getObjectType returns object type name
func getObjectType(obj interface{}) string {
    if obj == nil {
        return "nil"
    }
    return fmt.Sprintf("%T", obj)
}

// getGoroutineID returns current goroutine ID (simplified)
func getGoroutineID() uint64 {
    // This is a simplified implementation
    // In practice, you might use runtime.Stack() or other methods
    return uint64(time.Now().UnixNano()) % 1000
}

// Example usage and benchmarking
func ExamplePoolUsage() {
    // Create pool manager
    config := PoolConfig{
        MaxPoolSize:      1000,
        InitialSize:      10,
        GrowthFactor:     1.5,
        ShrinkThreshold:  0.3,
        CleanupInterval:  time.Minute,
        EnableMetrics:    true,
        EnableMonitoring: true,
        PoolType:         SyncPool,
    }
    
    manager := NewPoolManager(config)
    
    // Get a typed pool for byte slices
    bytePool := manager.GetTypedPool("bytes", func() interface{} {
        return make([]byte, 1024)
    })
    
    // Use the pool
    buffer := bytePool.Get().([]byte)
    
    // Use buffer...
    
    // Return to pool
    bytePool.Put(buffer)
    
    // Get metrics
    metrics := bytePool.GetMetrics()
    fmt.Printf("Pool hit rate: %.2f%%\n", float64(metrics.Hits)/float64(metrics.Gets)*100)
}
```

## Performance Analysis

Advanced techniques for analyzing and optimizing memory pool performance.

### Pool Benchmarking

Comprehensive benchmarking methodologies for pool performance.

### Memory Efficiency Analysis

Analyzing memory efficiency and allocation patterns.

### Contention Analysis

Understanding and reducing pool contention.

## Best Practices

1. **Choose Appropriate Pool Type**: Match pool type to usage patterns
2. **Size Pools Correctly**: Balance memory usage and hit rates
3. **Reset Objects**: Clear object state before returning to pool
4. **Validate Objects**: Ensure object integrity before reuse
5. **Monitor Performance**: Track hit rates and allocation patterns
6. **Use Typed Pools**: Create type-specific pools for better performance
7. **Consider Sharding**: Use sharded pools for high-contention scenarios
8. **Profile Regularly**: Continuously monitor pool effectiveness

## Summary

Memory pools are essential for high-performance Go applications:

1. **sync.Pool**: Best for temporary objects with unpredictable lifetimes
2. **Channel Pools**: Good for bounded object sets with predictable usage
3. **Lock-Free Pools**: Optimal for high-contention scenarios
4. **Sharded Pools**: Excellent for reducing contention across goroutines
5. **Monitoring**: Essential for optimizing pool configuration and usage

These techniques enable developers to minimize garbage collection pressure and maximize application performance through efficient object reuse.
