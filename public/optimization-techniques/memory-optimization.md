# Memory Optimization

Memory optimization is crucial for building high-performance Go applications that scale efficiently. This chapter covers advanced memory management techniques, allocation patterns, and strategies for minimizing garbage collection overhead while maintaining code clarity and maintainability.

## Understanding Go's Memory Model

### Memory Allocation Patterns
Go uses a sophisticated memory allocator that manages heap and stack allocations automatically. Understanding these patterns is essential for effective optimization:

```go
package memory_optimization

import (
    "runtime"
    "sync"
    "unsafe"
)

// Demonstrate allocation patterns and optimization strategies
type MemoryAnalyzer struct {
    allocStats    AllocStats
    poolManager   *PoolManager
    cacheManager  *CacheManager
}

type AllocStats struct {
    TotalAllocs   uint64
    HeapAllocs    uint64
    StackAllocs   uint64
    PoolHits      uint64
    CacheMisses   uint64
}

// Stack vs Heap allocation analysis
func AnalyzeAllocationPattern() {
    var m1, m2 runtime.MemStats
    
    // Measure stack allocation
    runtime.ReadMemStats(&m1)
    stackValue := createStackValue()
    runtime.ReadMemStats(&m2)
    
    stackAllocs := m2.Mallocs - m1.Mallocs
    
    // Measure heap allocation
    runtime.ReadMemStats(&m1)
    heapValue := createHeapValue()
    runtime.ReadMemStats(&m2)
    
    heapAllocs := m2.Mallocs - m1.Mallocs
    
    fmt.Printf("Stack allocations: %d\n", stackAllocs)
    fmt.Printf("Heap allocations: %d\n", heapAllocs)
    
    // Prevent optimization
    runtime.KeepAlive(stackValue)
    runtime.KeepAlive(heapValue)
}

// This typically allocates on stack (small, fixed size)
func createStackValue() [10]int {
    return [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
}

// This forces heap allocation (escape analysis)
func createHeapValue() *[10]int {
    return &[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
}

// Advanced escape analysis understanding
func demonstrateEscapeAnalysis() {
    // These will likely stay on stack
    localSlice := make([]int, 10)
    localMap := make(map[string]int)
    
    // This escapes to heap (returned pointer)
    heapSlice := createSliceOnHeap()
    
    // This escapes to heap (interface conversion)
    var interface{} = localSlice
    
    // This might escape (large size)
    largeArray := make([]int, 100000)
    
    processData(localSlice, localMap, heapSlice, largeArray)
}

func createSliceOnHeap() *[]int {
    slice := make([]int, 10)
    return &slice // Forces heap allocation
}

func processData(args ...interface{}) {
    // Process data without additional allocations
    for _, arg := range args {
        _ = arg
    }
}
```

### Memory Pool Management
Implement sophisticated memory pools to reduce allocation overhead:

```go
type PoolManager struct {
    bufferPools   map[int]*sync.Pool
    objectPools   map[string]*sync.Pool
    slicePools    map[int]*sync.Pool
    mu           sync.RWMutex
    stats        PoolStats
}

type PoolStats struct {
    Gets        int64
    Puts        int64
    Hits        int64
    Misses      int64
    Created     int64
    MaxSize     int64
}

func NewPoolManager() *PoolManager {
    return &PoolManager{
        bufferPools: make(map[int]*sync.Pool),
        objectPools: make(map[string]*sync.Pool),
        slicePools:  make(map[int]*sync.Pool),
    }
}

// Size-specific buffer pools
func (pm *PoolManager) GetBuffer(size int) []byte {
    pm.mu.RLock()
    pool, exists := pm.bufferPools[size]
    pm.mu.RUnlock()
    
    if !exists {
        pm.mu.Lock()
        // Double-check after acquiring write lock
        if pool, exists = pm.bufferPools[size]; !exists {
            pool = &sync.Pool{
                New: func() interface{} {
                    atomic.AddInt64(&pm.stats.Created, 1)
                    return make([]byte, size)
                },
            }
            pm.bufferPools[size] = pool
        }
        pm.mu.Unlock()
    }
    
    atomic.AddInt64(&pm.stats.Gets, 1)
    buffer := pool.Get().([]byte)
    
    // Reset buffer
    for i := range buffer {
        buffer[i] = 0
    }
    
    atomic.AddInt64(&pm.stats.Hits, 1)
    return buffer
}

func (pm *PoolManager) PutBuffer(buffer []byte) {
    if len(buffer) == 0 {
        return
    }
    
    size := len(buffer)
    pm.mu.RLock()
    pool, exists := pm.bufferPools[size]
    pm.mu.RUnlock()
    
    if exists {
        atomic.AddInt64(&pm.stats.Puts, 1)
        pool.Put(buffer)
    }
}

// Generic slice pool with capacity management
func (pm *PoolManager) GetSlice(capacity int) []interface{} {
    // Round up to nearest power of 2 for better pool utilization
    poolSize := nextPowerOf2(capacity)
    
    pm.mu.RLock()
    pool, exists := pm.slicePools[poolSize]
    pm.mu.RUnlock()
    
    if !exists {
        pm.mu.Lock()
        if pool, exists = pm.slicePools[poolSize]; !exists {
            pool = &sync.Pool{
                New: func() interface{} {
                    atomic.AddInt64(&pm.stats.Created, 1)
                    return make([]interface{}, 0, poolSize)
                },
            }
            pm.slicePools[poolSize] = pool
        }
        pm.mu.Unlock()
    }
    
    atomic.AddInt64(&pm.stats.Gets, 1)
    slice := pool.Get().([]interface{})
    
    // Reset slice length but keep capacity
    slice = slice[:0]
    
    atomic.AddInt64(&pm.stats.Hits, 1)
    return slice
}

func (pm *PoolManager) PutSlice(slice []interface{}) {
    if cap(slice) == 0 {
        return
    }
    
    poolSize := cap(slice)
    pm.mu.RLock()
    pool, exists := pm.slicePools[poolSize]
    pm.mu.RUnlock()
    
    if exists {
        atomic.AddInt64(&pm.stats.Puts, 1)
        // Clear references to prevent memory leaks
        for i := range slice {
            slice[i] = nil
        }
        pool.Put(slice[:0])
    }
}

// Typed object pools for complex structures
type ReusableObject interface {
    Reset()
}

func (pm *PoolManager) RegisterObjectPool(typeName string, newFunc func() ReusableObject) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.objectPools[typeName] = &sync.Pool{
        New: func() interface{} {
            atomic.AddInt64(&pm.stats.Created, 1)
            return newFunc()
        },
    }
}

func (pm *PoolManager) GetObject(typeName string) ReusableObject {
    pm.mu.RLock()
    pool, exists := pm.objectPools[typeName]
    pm.mu.RUnlock()
    
    if !exists {
        atomic.AddInt64(&pm.stats.Misses, 1)
        return nil
    }
    
    atomic.AddInt64(&pm.stats.Gets, 1)
    obj := pool.Get().(ReusableObject)
    obj.Reset()
    
    atomic.AddInt64(&pm.stats.Hits, 1)
    return obj
}

func (pm *PoolManager) PutObject(typeName string, obj ReusableObject) {
    pm.mu.RLock()
    pool, exists := pm.objectPools[typeName]
    pm.mu.RUnlock()
    
    if exists {
        atomic.AddInt64(&pm.stats.Puts, 1)
        pool.Put(obj)
    }
}

func nextPowerOf2(n int) int {
    if n <= 1 {
        return 1
    }
    
    power := 1
    for power < n {
        power <<= 1
    }
    return power
}

// Example reusable object
type ProcessingContext struct {
    InputBuffer  []byte
    OutputBuffer []byte
    TempData     map[string]interface{}
    Counters     map[string]int64
}

func (pc *ProcessingContext) Reset() {
    pc.InputBuffer = pc.InputBuffer[:0]
    pc.OutputBuffer = pc.OutputBuffer[:0]
    
    // Clear maps without reallocating
    for k := range pc.TempData {
        delete(pc.TempData, k)
    }
    
    for k := range pc.Counters {
        delete(pc.Counters, k)
    }
}

func NewProcessingContext() ReusableObject {
    return &ProcessingContext{
        InputBuffer:  make([]byte, 0, 1024),
        OutputBuffer: make([]byte, 0, 1024),
        TempData:     make(map[string]interface{}),
        Counters:     make(map[string]int64),
    }
}

// Pool usage example
func ExamplePoolUsage() {
    poolManager := NewPoolManager()
    
    // Register object pool
    poolManager.RegisterObjectPool("ProcessingContext", NewProcessingContext)
    
    // Use buffer pool
    buffer := poolManager.GetBuffer(1024)
    defer poolManager.PutBuffer(buffer)
    
    // Use object pool
    ctx := poolManager.GetObject("ProcessingContext").(*ProcessingContext)
    defer poolManager.PutObject("ProcessingContext", ctx)
    
    // Use slice pool
    slice := poolManager.GetSlice(100)
    defer poolManager.PutSlice(slice)
    
    // Process data using pooled resources
    processWithPools(buffer, ctx, slice)
}

func processWithPools(buffer []byte, ctx *ProcessingContext, slice []interface{}) {
    // Example processing using pooled resources
    copy(ctx.InputBuffer, buffer[:100])
    ctx.TempData["processing_time"] = time.Now()
    ctx.Counters["operations"] = 1
    
    for i := 0; i < 50; i++ {
        slice = append(slice, fmt.Sprintf("item_%d", i))
    }
}
```

## String and Slice Optimization

### Advanced String Building Techniques
Optimize string operations for minimal allocations:

```go
package string_optimization

import (
    "strings"
    "unsafe"
)

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

// Zero-allocation string building for known patterns
type ZeroAllocStringBuilder struct {
    buffer []byte
    offset int
}

func NewZeroAllocStringBuilder(capacity int) *ZeroAllocStringBuilder {
    return &ZeroAllocStringBuilder{
        buffer: make([]byte, capacity),
        offset: 0,
    }
}

func (zasb *ZeroAllocStringBuilder) WriteString(s string) {
    if zasb.offset+len(s) <= len(zasb.buffer) {
        copy(zasb.buffer[zasb.offset:], s)
        zasb.offset += len(s)
    }
}

func (zasb *ZeroAllocStringBuilder) WriteByte(b byte) {
    if zasb.offset < len(zasb.buffer) {
        zasb.buffer[zasb.offset] = b
        zasb.offset++
    }
}

func (zasb *ZeroAllocStringBuilder) String() string {
    // Zero-allocation conversion using unsafe
    return *(*string)(unsafe.Pointer(&zasb.buffer[:zasb.offset]))
}

func (zasb *ZeroAllocStringBuilder) Reset() {
    zasb.offset = 0
}

func (zasb *ZeroAllocStringBuilder) Len() int {
    return zasb.offset
}

func (zasb *ZeroAllocStringBuilder) Cap() int {
    return len(zasb.buffer)
}

// String interning for reducing memory usage
type StringInterner struct {
    intern map[string]string
    mu     sync.RWMutex
    stats  InternStats
}

type InternStats struct {
    Lookups    int64
    Hits       int64
    Inserts    int64
    MemorySaved int64
}

func NewStringInterner() *StringInterner {
    return &StringInterner{
        intern: make(map[string]string),
    }
}

func (si *StringInterner) Intern(s string) string {
    atomic.AddInt64(&si.stats.Lookups, 1)
    
    si.mu.RLock()
    if interned, exists := si.intern[s]; exists {
        si.mu.RUnlock()
        atomic.AddInt64(&si.stats.Hits, 1)
        atomic.AddInt64(&si.stats.MemorySaved, int64(len(s)))
        return interned
    }
    si.mu.RUnlock()
    
    si.mu.Lock()
    defer si.mu.Unlock()
    
    // Double-check pattern
    if interned, exists := si.intern[s]; exists {
        atomic.AddInt64(&si.stats.Hits, 1)
        return interned
    }
    
    // Create canonical string
    canonical := string([]byte(s))
    si.intern[canonical] = canonical
    atomic.AddInt64(&si.stats.Inserts, 1)
    
    return canonical
}

func (si *StringInterner) Stats() InternStats {
    return InternStats{
        Lookups:     atomic.LoadInt64(&si.stats.Lookups),
        Hits:        atomic.LoadInt64(&si.stats.Hits),
        Inserts:     atomic.LoadInt64(&si.stats.Inserts),
        MemorySaved: atomic.LoadInt64(&si.stats.MemorySaved),
    }
}

// Advanced slice operations with minimal allocations
type SliceOptimizer struct {
    pools map[reflect.Type]*sync.Pool
    mu    sync.RWMutex
}

func NewSliceOptimizer() *SliceOptimizer {
    return &SliceOptimizer{
        pools: make(map[reflect.Type]*sync.Pool),
    }
}

// Generic slice pooling with type safety
func (so *SliceOptimizer) GetSlice(elementType reflect.Type, capacity int) interface{} {
    so.mu.RLock()
    pool, exists := so.pools[elementType]
    so.mu.RUnlock()
    
    if !exists {
        so.mu.Lock()
        if pool, exists = so.pools[elementType]; !exists {
            pool = &sync.Pool{
                New: func() interface{} {
                    return reflect.MakeSlice(
                        reflect.SliceOf(elementType),
                        0,
                        capacity,
                    ).Interface()
                },
            }
            so.pools[elementType] = pool
        }
        so.mu.Unlock()
    }
    
    slice := pool.Get()
    
    // Reset slice length
    sliceValue := reflect.ValueOf(slice)
    sliceValue.SetLen(0)
    
    return slice
}

func (so *SliceOptimizer) PutSlice(slice interface{}) {
    sliceValue := reflect.ValueOf(slice)
    if sliceValue.Kind() != reflect.Slice {
        return
    }
    
    elementType := sliceValue.Type().Elem()
    
    so.mu.RLock()
    pool, exists := so.pools[elementType]
    so.mu.RUnlock()
    
    if exists {
        // Clear slice elements to prevent memory leaks
        for i := 0; i < sliceValue.Len(); i++ {
            sliceValue.Index(i).Set(reflect.Zero(elementType))
        }
        sliceValue.SetLen(0)
        
        pool.Put(slice)
    }
}

// In-place slice operations
func AppendInPlace(slice []interface{}, elements ...interface{}) []interface{} {
    // Check if we have enough capacity
    needed := len(slice) + len(elements)
    if cap(slice) >= needed {
        // Extend slice length and copy elements
        extended := slice[:needed]
        copy(extended[len(slice):], elements)
        return extended
    }
    
    // Need to allocate new slice
    return append(slice, elements...)
}

func RemoveInPlace(slice []interface{}, index int) []interface{} {
    if index < 0 || index >= len(slice) {
        return slice
    }
    
    // Move elements left
    copy(slice[index:], slice[index+1:])
    
    // Clear last element and shrink
    slice[len(slice)-1] = nil
    return slice[:len(slice)-1]
}

func FilterInPlace(slice []interface{}, predicate func(interface{}) bool) []interface{} {
    writeIndex := 0
    
    for readIndex, element := range slice {
        if predicate(element) {
            if writeIndex != readIndex {
                slice[writeIndex] = element
            }
            writeIndex++
        }
    }
    
    // Clear remaining elements
    for i := writeIndex; i < len(slice); i++ {
        slice[i] = nil
    }
    
    return slice[:writeIndex]
}
```

## Cache-Friendly Data Structures

### Implementing Cache-Aware Algorithms
Design data structures that optimize CPU cache utilization:

```go
package cache_optimization

import (
    "math/bits"
    "unsafe"
)

// Cache-friendly array-based data structures
const (
    CacheLineSize = 64 // bytes
    L1CacheSize   = 32 * 1024 // 32KB typical L1 cache
    L2CacheSize   = 256 * 1024 // 256KB typical L2 cache
    L3CacheSize   = 8 * 1024 * 1024 // 8MB typical L3 cache
)

// Cache-aligned structure to prevent false sharing
type CacheAligned struct {
    data [CacheLineSize]byte
}

// Hot/Cold data separation
type HotColdSeparatedStruct struct {
    // Hot data: frequently accessed fields
    ID        uint64
    Counter   int64
    Active    bool
    _         [CacheLineSize - 17]byte // Padding to cache line boundary
    
    // Cold data: infrequently accessed fields
    Description string
    Metadata    map[string]interface{}
    History     []Event
}

type Event struct {
    Timestamp time.Time
    Type      string
    Data      interface{}
}

// Array of Structures (AoS) vs Structure of Arrays (SoA)
type ParticleAoS struct {
    X, Y, Z    float64
    VX, VY, VZ float64
    Mass       float64
    Active     bool
}

type ParticlesSoA struct {
    Count      int
    X, Y, Z    []float64
    VX, VY, VZ []float64
    Mass       []float64
    Active     []bool
}

func NewParticlesSoA(capacity int) *ParticlesSoA {
    return &ParticlesSoA{
        Count:  0,
        X:      make([]float64, capacity),
        Y:      make([]float64, capacity),
        Z:      make([]float64, capacity),
        VX:     make([]float64, capacity),
        VY:     make([]float64, capacity),
        VZ:     make([]float64, capacity),
        Mass:   make([]float64, capacity),
        Active: make([]bool, capacity),
    }
}

// Cache-friendly iteration patterns
func UpdateParticlesAoS(particles []ParticleAoS, dt float64) {
    // Poor cache utilization - jumping between distant memory locations
    for i := range particles {
        if particles[i].Active {
            particles[i].X += particles[i].VX * dt
            particles[i].Y += particles[i].VY * dt
            particles[i].Z += particles[i].VZ * dt
        }
    }
}

func UpdateParticlesSoA(particles *ParticlesSoA, dt float64) {
    // Excellent cache utilization - sequential memory access
    for i := 0; i < particles.Count; i++ {
        if particles.Active[i] {
            particles.X[i] += particles.VX[i] * dt
            particles.Y[i] += particles.VY[i] * dt
            particles.Z[i] += particles.VZ[i] * dt
        }
    }
}

// Memory layout optimization
type OptimizedStruct struct {
    // Group fields by size and access pattern
    // 8-byte fields first
    BigCounter1 int64
    BigCounter2 int64
    Timestamp   time.Time
    
    // 4-byte fields
    MediumCounter1 int32
    MediumCounter2 int32
    
    // 2-byte fields
    SmallCounter1 int16
    SmallCounter2 int16
    
    // 1-byte fields
    Flag1 bool
    Flag2 bool
    Flag3 bool
    Flag4 bool
    
    // String and slice fields last (they contain pointers)
    Name    string
    Data    []byte
    Metadata map[string]interface{}
}

// Bit packing for memory efficiency
type PackedFlags struct {
    flags uint64
}

func (pf *PackedFlags) SetFlag(position uint, value bool) {
    if position >= 64 {
        return
    }
    
    if value {
        pf.flags |= 1 << position
    } else {
        pf.flags &^= 1 << position
    }
}

func (pf *PackedFlags) GetFlag(position uint) bool {
    if position >= 64 {
        return false
    }
    
    return (pf.flags & (1 << position)) != 0
}

func (pf *PackedFlags) CountSetBits() int {
    return bits.OnesCount64(pf.flags)
}

// Cache-aware hash table implementation
type CacheAwareHashMap struct {
    buckets    []CacheLineBucket
    bucketMask uint64
    size       int64
    tombstones int64
}

type CacheLineBucket struct {
    // Pack multiple key-value pairs in a cache line
    entries [4]Entry
    metadata uint64 // 16 bits per entry: 15 bits hash + 1 bit occupied
}

type Entry struct {
    key   uint32
    value uint32
}

func NewCacheAwareHashMap(capacity int) *CacheAwareHashMap {
    // Round up to power of 2
    bucketCount := nextPowerOf2(capacity / 4)
    
    return &CacheAwareHashMap{
        buckets:    make([]CacheLineBucket, bucketCount),
        bucketMask: uint64(bucketCount - 1),
    }
}

func (cam *CacheAwareHashMap) Put(key, value uint32) bool {
    hash := cam.hash(key)
    bucketIndex := hash & cam.bucketMask
    
    bucket := &cam.buckets[bucketIndex]
    
    // Linear probing within the cache line
    for i := 0; i < 4; i++ {
        entryMeta := (bucket.metadata >> (i * 16)) & 0xFFFF
        
        if entryMeta == 0 { // Empty slot
            bucket.entries[i] = Entry{key: key, value: value}
            bucket.metadata |= (uint64(hash&0x7FFF) | 0x8000) << (i * 16)
            atomic.AddInt64(&cam.size, 1)
            return true
        }
        
        if bucket.entries[i].key == key { // Update existing
            bucket.entries[i].value = value
            return true
        }
    }
    
    // Cache line full, use quadratic probing
    return cam.putWithProbing(key, value, hash)
}

func (cam *CacheAwareHashMap) Get(key uint32) (uint32, bool) {
    hash := cam.hash(key)
    bucketIndex := hash & cam.bucketMask
    
    bucket := &cam.buckets[bucketIndex]
    
    // Check cache line first
    for i := 0; i < 4; i++ {
        entryMeta := (bucket.metadata >> (i * 16)) & 0xFFFF
        
        if entryMeta == 0 {
            return 0, false // Empty slot, key not found
        }
        
        if (entryMeta&0x8000) != 0 && bucket.entries[i].key == key {
            return bucket.entries[i].value, true
        }
    }
    
    // Search with probing
    return cam.getWithProbing(key, hash)
}

func (cam *CacheAwareHashMap) hash(key uint32) uint64 {
    // FNV-1a hash
    hash := uint64(2166136261)
    bytes := (*[4]byte)(unsafe.Pointer(&key))[:]
    
    for _, b := range bytes {
        hash ^= uint64(b)
        hash *= 16777619
    }
    
    return hash
}

func (cam *CacheAwareHashMap) putWithProbing(key, value uint32, hash uint64) bool {
    // Quadratic probing with triangular numbers
    for probe := uint64(1); probe < uint64(len(cam.buckets)); probe++ {
        bucketIndex := (hash + probe*probe) & cam.bucketMask
        bucket := &cam.buckets[bucketIndex]
        
        for i := 0; i < 4; i++ {
            entryMeta := (bucket.metadata >> (i * 16)) & 0xFFFF
            
            if entryMeta == 0 {
                bucket.entries[i] = Entry{key: key, value: value}
                bucket.metadata |= (uint64(hash&0x7FFF) | 0x8000) << (i * 16)
                atomic.AddInt64(&cam.size, 1)
                return true
            }
            
            if bucket.entries[i].key == key {
                bucket.entries[i].value = value
                return true
            }
        }
    }
    
    return false // Hash table full
}

func (cam *CacheAwareHashMap) getWithProbing(key uint32, hash uint64) (uint32, bool) {
    for probe := uint64(1); probe < uint64(len(cam.buckets)); probe++ {
        bucketIndex := (hash + probe*probe) & cam.bucketMask
        bucket := &cam.buckets[bucketIndex]
        
        for i := 0; i < 4; i++ {
            entryMeta := (bucket.metadata >> (i * 16)) & 0xFFFF
            
            if entryMeta == 0 {
                return 0, false
            }
            
            if (entryMeta&0x8000) != 0 && bucket.entries[i].key == key {
                return bucket.entries[i].value, true
            }
        }
    }
    
    return 0, false
}

// Cache-aware sorting algorithms
func CacheAwareMergeSort(data []int) {
    if len(data) <= 1 {
        return
    }
    
    // Use insertion sort for small arrays (better cache utilization)
    if len(data) <= 32 {
        insertionSort(data)
        return
    }
    
    // Cache-aware merge sort
    mid := len(data) / 2
    CacheAwareMergeSort(data[:mid])
    CacheAwareMergeSort(data[mid:])
    
    merge(data, mid)
}

func insertionSort(data []int) {
    for i := 1; i < len(data); i++ {
        key := data[i]
        j := i - 1
        
        for j >= 0 && data[j] > key {
            data[j+1] = data[j]
            j--
        }
        
        data[j+1] = key
    }
}

func merge(data []int, mid int) {
    // Use a temporary buffer for merging
    temp := make([]int, len(data))
    copy(temp, data)
    
    i, j, k := 0, mid, 0
    
    for i < mid && j < len(data) {
        if temp[i] <= temp[j] {
            data[k] = temp[i]
            i++
        } else {
            data[k] = temp[j]
            j++
        }
        k++
    }
    
    for i < mid {
        data[k] = temp[i]
        i++
        k++
    }
    
    for j < len(data) {
        data[k] = temp[j]
        j++
        k++
    }
}
```

## Garbage Collection Optimization

### GC-Aware Programming Patterns
Implement strategies to minimize garbage collection impact:

```go
package gc_optimization

import (
    "runtime"
    "runtime/debug"
    "time"
)

// GC monitoring and tuning
type GCOptimizer struct {
    config      GCConfig
    stats       GCStats
    lastGCStats debug.GCStats
    tuner       *GCTuner
}

type GCConfig struct {
    TargetHeapSize   int64   `json:"target_heap_size"`
    GCPercent       int     `json:"gc_percent"`
    MaxGCPause      time.Duration `json:"max_gc_pause"`
    AllocationRate  int64   `json:"allocation_rate"`
    TuningEnabled   bool    `json:"tuning_enabled"`
}

type GCStats struct {
    NumGC           uint32        `json:"num_gc"`
    PauseTotal      time.Duration `json:"pause_total"`
    PauseAvg        time.Duration `json:"pause_avg"`
    PauseMax        time.Duration `json:"pause_max"`
    HeapSize        uint64        `json:"heap_size"`
    AllocRate       float64       `json:"alloc_rate"`
    GCOverhead      float64       `json:"gc_overhead"`
}

func NewGCOptimizer(config GCConfig) *GCOptimizer {
    optimizer := &GCOptimizer{
        config: config,
        tuner:  NewGCTuner(config),
    }
    
    if config.TuningEnabled {
        go optimizer.startTuningLoop()
    }
    
    return optimizer
}

func (gco *GCOptimizer) startTuningLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        gco.updateStats()
        gco.tuner.Tune(gco.stats)
    }
}

func (gco *GCOptimizer) updateStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    var gcStats debug.GCStats
    debug.ReadGCStats(&gcStats)
    
    // Calculate GC overhead
    totalTime := time.Since(time.Unix(0, gcStats.LastGC))
    gcTime := time.Duration(0)
    for _, pause := range gcStats.Pause {
        gcTime += pause
    }
    
    gcOverhead := float64(gcTime) / float64(totalTime) * 100
    
    // Calculate allocation rate
    timeDiff := float64(gcStats.LastGC - gco.lastGCStats.LastGC) / 1e9 // Convert to seconds
    allocDiff := float64(m.TotalAlloc - m.Mallocs)
    allocRate := allocDiff / timeDiff
    
    gco.stats = GCStats{
        NumGC:      m.NumGC,
        PauseTotal: time.Duration(m.PauseTotalNs),
        HeapSize:   m.HeapSys,
        AllocRate:  allocRate,
        GCOverhead: gcOverhead,
    }
    
    // Calculate pause statistics
    if len(gcStats.Pause) > 0 {
        var total, max time.Duration
        for _, pause := range gcStats.Pause {
            total += pause
            if pause > max {
                max = pause
            }
        }
        
        gco.stats.PauseAvg = total / time.Duration(len(gcStats.Pause))
        gco.stats.PauseMax = max
    }
    
    gco.lastGCStats = gcStats
}

// Adaptive GC tuning
type GCTuner struct {
    config        GCConfig
    currentGCPercent int
    history       []GCMeasurement
    maxHistory    int
}

type GCMeasurement struct {
    Timestamp   time.Time     `json:"timestamp"`
    GCPercent   int          `json:"gc_percent"`
    PauseTime   time.Duration `json:"pause_time"`
    HeapSize    uint64       `json:"heap_size"`
    AllocRate   float64      `json:"alloc_rate"`
}

func NewGCTuner(config GCConfig) *GCTuner {
    return &GCTuner{
        config:         config,
        currentGCPercent: config.GCPercent,
        maxHistory:     50,
        history:        make([]GCMeasurement, 0, 50),
    }
}

func (gct *GCTuner) Tune(stats GCStats) {
    measurement := GCMeasurement{
        Timestamp: time.Now(),
        GCPercent: gct.currentGCPercent,
        PauseTime: stats.PauseAvg,
        HeapSize:  stats.HeapSize,
        AllocRate: stats.AllocRate,
    }
    
    gct.addMeasurement(measurement)
    
    // Tune GC based on current performance
    newGCPercent := gct.calculateOptimalGCPercent(stats)
    
    if newGCPercent != gct.currentGCPercent {
        fmt.Printf("Tuning GC: %d%% -> %d%%\n", gct.currentGCPercent, newGCPercent)
        debug.SetGCPercent(newGCPercent)
        gct.currentGCPercent = newGCPercent
    }
}

func (gct *GCTuner) addMeasurement(measurement GCMeasurement) {
    gct.history = append(gct.history, measurement)
    
    if len(gct.history) > gct.maxHistory {
        gct.history = gct.history[1:]
    }
}

func (gct *GCTuner) calculateOptimalGCPercent(stats GCStats) int {
    // If pauses are too long, increase GC frequency (lower GC percent)
    if stats.PauseAvg > gct.config.MaxGCPause {
        return max(50, gct.currentGCPercent-25)
    }
    
    // If GC overhead is too high, decrease GC frequency (higher GC percent)
    if stats.GCOverhead > 10.0 { // More than 10% overhead
        return min(800, gct.currentGCPercent+50)
    }
    
    // If heap is growing too large, increase GC frequency
    if int64(stats.HeapSize) > gct.config.TargetHeapSize {
        return max(50, gct.currentGCPercent-10)
    }
    
    // Use historical data to predict optimal setting
    return gct.predictOptimalGCPercent()
}

func (gct *GCTuner) predictOptimalGCPercent() int {
    if len(gct.history) < 10 {
        return gct.currentGCPercent
    }
    
    // Find the GC percent setting with the best balance of pause time and overhead
    bestScore := float64(-1)
    bestGCPercent := gct.currentGCPercent
    
    // Group measurements by GC percent
    groups := make(map[int][]GCMeasurement)
    for _, measurement := range gct.history {
        groups[measurement.GCPercent] = append(groups[measurement.GCPercent], measurement)
    }
    
    for gcPercent, measurements := range groups {
        if len(measurements) < 3 {
            continue
        }
        
        // Calculate average performance metrics
        var avgPause time.Duration
        var avgHeapSize uint64
        
        for _, m := range measurements {
            avgPause += m.PauseTime
            avgHeapSize += m.HeapSize
        }
        
        avgPause /= time.Duration(len(measurements))
        avgHeapSize /= uint64(len(measurements))
        
        // Score based on pause time and memory usage
        pauseScore := 1.0 - float64(avgPause)/float64(gct.config.MaxGCPause)
        memoryScore := 1.0 - float64(avgHeapSize)/float64(gct.config.TargetHeapSize)
        
        score := (pauseScore + memoryScore) / 2.0
        
        if score > bestScore {
            bestScore = score
            bestGCPercent = gcPercent
        }
    }
    
    return bestGCPercent
}

// GC-friendly data structures
type GCFriendlyMap struct {
    buckets    [][]KeyValue
    bucketMask int
    size       int
    gcPressure int64
}

type KeyValue struct {
    Key   string
    Value interface{}
}

func NewGCFriendlyMap(capacity int) *GCFriendlyMap {
    bucketCount := nextPowerOf2(capacity / 8)
    
    return &GCFriendlyMap{
        buckets:    make([][]KeyValue, bucketCount),
        bucketMask: bucketCount - 1,
    }
}

func (gfm *GCFriendlyMap) Put(key string, value interface{}) {
    hash := gfm.hash(key)
    bucketIndex := hash & gfm.bucketMask
    
    bucket := gfm.buckets[bucketIndex]
    
    // Check if key exists
    for i, kv := range bucket {
        if kv.Key == key {
            bucket[i].Value = value
            return
        }
    }
    
    // Add new key-value pair
    gfm.buckets[bucketIndex] = append(bucket, KeyValue{Key: key, Value: value})
    gfm.size++
    
    // Monitor GC pressure
    atomic.AddInt64(&gfm.gcPressure, 1)
    
    // Trigger cleanup if needed
    if atomic.LoadInt64(&gfm.gcPressure) > int64(gfm.size)*2 {
        gfm.cleanup()
    }
}

func (gfm *GCFriendlyMap) Get(key string) (interface{}, bool) {
    hash := gfm.hash(key)
    bucketIndex := hash & gfm.bucketMask
    
    bucket := gfm.buckets[bucketIndex]
    
    for _, kv := range bucket {
        if kv.Key == key {
            return kv.Value, true
        }
    }
    
    return nil, false
}

func (gfm *GCFriendlyMap) cleanup() {
    // Force garbage collection to clean up unreferenced objects
    runtime.GC()
    atomic.StoreInt64(&gfm.gcPressure, 0)
}

func (gfm *GCFriendlyMap) hash(key string) int {
    hash := 0
    for _, b := range []byte(key) {
        hash = hash*31 + int(b)
    }
    return hash
}

// Object recycling to reduce GC pressure
type ObjectRecycler struct {
    pools map[string]*sync.Pool
    mu    sync.RWMutex
    stats RecyclerStats
}

type RecyclerStats struct {
    ObjectsCreated  int64 `json:"objects_created"`
    ObjectsRecycled int64 `json:"objects_recycled"`
    PoolHits       int64 `json:"pool_hits"`
    PoolMisses     int64 `json:"pool_misses"`
}

func NewObjectRecycler() *ObjectRecycler {
    return &ObjectRecycler{
        pools: make(map[string]*sync.Pool),
    }
}

func (or *ObjectRecycler) RegisterType(typeName string, factory func() interface{}) {
    or.mu.Lock()
    defer or.mu.Unlock()
    
    or.pools[typeName] = &sync.Pool{
        New: func() interface{} {
            atomic.AddInt64(&or.stats.ObjectsCreated, 1)
            return factory()
        },
    }
}

func (or *ObjectRecycler) Get(typeName string) interface{} {
    or.mu.RLock()
    pool, exists := or.pools[typeName]
    or.mu.RUnlock()
    
    if !exists {
        atomic.AddInt64(&or.stats.PoolMisses, 1)
        return nil
    }
    
    atomic.AddInt64(&or.stats.PoolHits, 1)
    return pool.Get()
}

func (or *ObjectRecycler) Put(typeName string, obj interface{}) {
    or.mu.RLock()
    pool, exists := or.pools[typeName]
    or.mu.RUnlock()
    
    if exists {
        // Reset object if it implements Resetter interface
        if resetter, ok := obj.(interface{ Reset() }); ok {
            resetter.Reset()
        }
        
        atomic.AddInt64(&or.stats.ObjectsRecycled, 1)
        pool.Put(obj)
    }
}

// Helper functions
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

Memory optimization is a continuous process that requires understanding Go's memory model, implementing appropriate pooling strategies, designing cache-friendly data structures, and managing garbage collection effectively. By applying these techniques systematically, you can build applications that scale efficiently and maintain consistent performance under load.

## Key Takeaways

1. **Understand allocation patterns** - distinguish between stack and heap allocations
2. **Implement sophisticated pooling** - use typed pools for different object types
3. **Design cache-friendly structures** - consider CPU cache behavior in data layout
4. **Separate hot and cold data** - group frequently accessed fields together
5. **Minimize GC pressure** - reduce allocation rate and object lifetime
6. **Use in-place operations** - modify data structures without additional allocations
7. **Monitor and tune GC** - adapt garbage collection parameters to workload characteristics
8. **Implement object recycling** - reuse expensive objects to reduce allocation overhead

Effective memory optimization enables Go applications to handle larger workloads with lower resource consumption and more predictable performance characteristics.
