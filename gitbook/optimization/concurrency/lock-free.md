# Lock-Free Programming

Lock-free programming in Go enables building high-performance concurrent systems without traditional synchronization primitives. This comprehensive guide covers lock-free algorithms, data structures, and optimization techniques for Go applications.

## Table of Contents

- [Introduction to Lock-Free Programming](#introduction-to-lock-free-programming)
- [Atomic Operations](#atomic-operations)
- [Memory Models and Ordering](#memory-models-and-ordering)
- [Lock-Free Data Structures](#lock-free-data-structures)
- [ABA Problem and Solutions](#aba-problem-and-solutions)
- [Performance Benchmarking](#performance-benchmarking)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)

## Introduction to Lock-Free Programming

Lock-free programming uses atomic operations and memory barriers to achieve thread safety without locks, eliminating deadlocks, priority inversion, and lock contention.

### Benefits and Trade-offs

```go
package main

import (
    "runtime"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// LockFreeSystem provides a framework for lock-free programming
type LockFreeSystem struct {
    atomicCounter    int64
    memoryOrdering   MemoryOrderingManager
    hazardPointer    *HazardPointerManager
    performanceMetrics *LockFreeMetrics
}

// LockFreeMetrics tracks performance metrics for lock-free operations
type LockFreeMetrics struct {
    OperationsPerformed  int64
    CASFailures         int64
    RetryAttempts       int64
    AverageRetries      float64
    ThroughputPerSecond float64
    MemoryAccesses      int64
}

// MemoryOrderingManager handles memory ordering constraints
type MemoryOrderingManager struct {
    sequentialConsistency bool
    acquireRelease       bool
    relaxedOrdering      bool
}

// HazardPointerManager manages hazard pointers for safe memory reclamation
type HazardPointerManager struct {
    hazardPointers  []unsafe.Pointer
    retiredPointers chan unsafe.Pointer
    epoch          int64
    participants   int32
}

// NewLockFreeSystem creates a new lock-free system
func NewLockFreeSystem() *LockFreeSystem {
    return &LockFreeSystem{
        memoryOrdering:     MemoryOrderingManager{sequentialConsistency: true},
        hazardPointer:      NewHazardPointerManager(),
        performanceMetrics: &LockFreeMetrics{},
    }
}

// NewHazardPointerManager creates a new hazard pointer manager
func NewHazardPointerManager() *HazardPointerManager {
    return &HazardPointerManager{
        hazardPointers:  make([]unsafe.Pointer, runtime.NumCPU()*2),
        retiredPointers: make(chan unsafe.Pointer, 1000),
    }
}

// CompareAndSwapPointer performs atomic CAS operation on pointers
func (lfs *LockFreeSystem) CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) bool {
    atomic.AddInt64(&lfs.performanceMetrics.OperationsPerformed, 1)
    
    success := atomic.CompareAndSwapPointer(addr, old, new)
    if !success {
        atomic.AddInt64(&lfs.performanceMetrics.CASFailures, 1)
    }
    
    return success
}

// LoadPointer performs atomic load with memory ordering
func (lfs *LockFreeSystem) LoadPointer(addr *unsafe.Pointer) unsafe.Pointer {
    atomic.AddInt64(&lfs.performanceMetrics.MemoryAccesses, 1)
    return atomic.LoadPointer(addr)
}

// StorePointer performs atomic store with memory ordering
func (lfs *LockFreeSystem) StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {
    atomic.AddInt64(&lfs.performanceMetrics.MemoryAccesses, 1)
    atomic.StorePointer(addr, val)
}
```

## Atomic Operations

Comprehensive coverage of Go's atomic operations for lock-free programming.

### Basic Atomic Operations

```go
// AtomicOperationsDemo demonstrates various atomic operations
type AtomicOperationsDemo struct {
    counter     int64
    flag        int32
    pointer     unsafe.Pointer
    value       atomic.Value
}

// CounterOperations demonstrates atomic counter operations
func (aod *AtomicOperationsDemo) CounterOperations() {
    // Basic increment/decrement
    atomic.AddInt64(&aod.counter, 1)
    atomic.AddInt64(&aod.counter, -1)
    
    // Compare and swap
    for {
        old := atomic.LoadInt64(&aod.counter)
        new := old * 2
        if atomic.CompareAndSwapInt64(&aod.counter, old, new) {
            break
        }
        // Retry on failure
    }
    
    // Swap operation
    oldValue := atomic.SwapInt64(&aod.counter, 42)
    _ = oldValue
}

// FlagOperations demonstrates atomic flag operations
func (aod *AtomicOperationsDemo) FlagOperations() {
    // Set flag
    atomic.StoreInt32(&aod.flag, 1)
    
    // Check flag
    if atomic.LoadInt32(&aod.flag) == 1 {
        // Flag is set
    }
    
    // Test and set
    if atomic.CompareAndSwapInt32(&aod.flag, 0, 1) {
        // Successfully acquired flag
        defer atomic.StoreInt32(&aod.flag, 0) // Release
    }
}

// PointerOperations demonstrates atomic pointer operations
func (aod *AtomicOperationsDemo) PointerOperations() {
    type Node struct {
        value int
        next  *Node
    }
    
    // Create new node
    newNode := &Node{value: 42}
    
    // Atomic pointer store
    atomic.StorePointer(&aod.pointer, unsafe.Pointer(newNode))
    
    // Atomic pointer load
    ptr := atomic.LoadPointer(&aod.pointer)
    if ptr != nil {
        node := (*Node)(ptr)
        _ = node.value
    }
    
    // Compare and swap pointer
    oldPtr := atomic.LoadPointer(&aod.pointer)
    newPtr := unsafe.Pointer(&Node{value: 100})
    atomic.CompareAndSwapPointer(&aod.pointer, oldPtr, newPtr)
}

// ValueOperations demonstrates atomic.Value operations
func (aod *AtomicOperationsDemo) ValueOperations() {
    type Config struct {
        timeout time.Duration
        retries int
    }
    
    // Store value
    config := &Config{timeout: time.Second, retries: 3}
    aod.value.Store(config)
    
    // Load value
    if loaded := aod.value.Load(); loaded != nil {
        if cfg, ok := loaded.(*Config); ok {
            _ = cfg.timeout
        }
    }
    
    // Conditional store (no CAS for atomic.Value)
    newConfig := &Config{timeout: 2 * time.Second, retries: 5}
    aod.value.Store(newConfig)
}
```

### Advanced Atomic Patterns

```go
// AtomicPatterns demonstrates advanced atomic programming patterns
type AtomicPatterns struct {
    refCount    int64
    state       int64
    epochTime   int64
}

// ReferenceCountingPattern implements atomic reference counting
func (ap *AtomicPatterns) ReferenceCountingPattern() {
    // Increment reference count
    atomic.AddInt64(&ap.refCount, 1)
    
    // Decrement and check if zero
    if atomic.AddInt64(&ap.refCount, -1) == 0 {
        // Last reference released, cleanup
        ap.cleanup()
    }
}

// cleanup performs cleanup when reference count reaches zero
func (ap *AtomicPatterns) cleanup() {
    // Cleanup logic
}

// StateTransitionPattern implements atomic state transitions
func (ap *AtomicPatterns) StateTransitionPattern() {
    const (
        StateIdle    = 0
        StateRunning = 1
        StateStopped = 2
    )
    
    // Try to transition from idle to running
    if atomic.CompareAndSwapInt64(&ap.state, StateIdle, StateRunning) {
        // Successfully transitioned to running
        defer atomic.StoreInt64(&ap.state, StateStopped)
        
        // Do work
        ap.doWork()
    }
}

// doWork performs some work
func (ap *AtomicPatterns) doWork() {
    time.Sleep(100 * time.Millisecond)
}

// TimestampPattern implements atomic timestamp management
func (ap *AtomicPatterns) TimestampPattern() {
    now := time.Now().UnixNano()
    
    // Update timestamp if newer
    for {
        current := atomic.LoadInt64(&ap.epochTime)
        if now <= current {
            break // Current timestamp is newer or equal
        }
        
        if atomic.CompareAndSwapInt64(&ap.epochTime, current, now) {
            break // Successfully updated
        }
        // Retry if CAS failed
    }
}

// GetLatestTimestamp returns the latest timestamp
func (ap *AtomicPatterns) GetLatestTimestamp() time.Time {
    nano := atomic.LoadInt64(&ap.epochTime)
    return time.Unix(0, nano)
}
```

## Memory Models and Ordering

Understanding Go's memory model and ensuring correct ordering in lock-free code.

### Memory Ordering Guarantees

```go
// MemoryOrderingExample demonstrates memory ordering concepts
type MemoryOrderingExample struct {
    data  int64
    ready int32
    result int64
}

// SequentialConsistencyExample demonstrates sequential consistency
func (moe *MemoryOrderingExample) SequentialConsistencyExample() {
    // Writer goroutine
    go func() {
        // All operations are sequentially consistent in Go
        atomic.StoreInt64(&moe.data, 42)
        atomic.StoreInt32(&moe.ready, 1) // Release
    }()
    
    // Reader goroutine
    go func() {
        for atomic.LoadInt32(&moe.ready) == 0 { // Acquire
            runtime.Gosched()
        }
        value := atomic.LoadInt64(&moe.data)
        atomic.StoreInt64(&moe.result, value)
    }()
}

// AcquireReleaseExample demonstrates acquire-release semantics
func (moe *MemoryOrderingExample) AcquireReleaseExample() {
    // In Go, all atomic operations have acquire-release semantics
    // This ensures that:
    // 1. No memory operation can be reordered before an acquire
    // 2. No memory operation can be reordered after a release
    
    // Producer
    go func() {
        atomic.StoreInt64(&moe.data, 100)    // Regular store
        atomic.StoreInt32(&moe.ready, 1)     // Release store
    }()
    
    // Consumer
    go func() {
        if atomic.LoadInt32(&moe.ready) == 1 { // Acquire load
            value := atomic.LoadInt64(&moe.data) // Will see the store
            atomic.StoreInt64(&moe.result, value)
        }
    }()
}

// MemoryBarrierExample demonstrates explicit memory barriers
func (moe *MemoryOrderingExample) MemoryBarrierExample() {
    // Go doesn't expose explicit memory barriers, but we can achieve
    // similar effects through atomic operations
    
    var x, y int64
    var barrier int32
    
    // Writer
    go func() {
        atomic.StoreInt64(&x, 1)
        atomic.StoreInt32(&barrier, 1) // Memory barrier effect
        atomic.StoreInt64(&y, 1)
    }()
    
    // Reader
    go func() {
        if atomic.LoadInt64(&y) == 1 {
            atomic.LoadInt32(&barrier) // Memory barrier effect
            value := atomic.LoadInt64(&x) // Guaranteed to see x = 1
            atomic.StoreInt64(&moe.result, value)
        }
    }()
}
```

### Publication and Initialization

```go
// PublicationPattern demonstrates safe object publication
type PublicationPattern struct {
    instance unsafe.Pointer
    once     sync.Once
}

// SharedObject represents an object to be safely published
type SharedObject struct {
    value int
    data  []byte
}

// NewSharedObject creates a new shared object
func NewSharedObject(value int) *SharedObject {
    return &SharedObject{
        value: value,
        data:  make([]byte, 1024),
    }
}

// SafePublication demonstrates safe object publication
func (pp *PublicationPattern) SafePublication(value int) *SharedObject {
    // Double-checked locking pattern
    ptr := atomic.LoadPointer(&pp.instance)
    if ptr != nil {
        return (*SharedObject)(ptr)
    }
    
    // Create new object
    obj := NewSharedObject(value)
    
    // Try to publish
    if atomic.CompareAndSwapPointer(&pp.instance, nil, unsafe.Pointer(obj)) {
        return obj // Successfully published
    }
    
    // Another goroutine published first, return their object
    return (*SharedObject)(atomic.LoadPointer(&pp.instance))
}

// LazyInitialization demonstrates lock-free lazy initialization
func (pp *PublicationPattern) LazyInitialization() *SharedObject {
    pp.once.Do(func() {
        obj := NewSharedObject(42)
        atomic.StorePointer(&pp.instance, unsafe.Pointer(obj))
    })
    
    return (*SharedObject)(atomic.LoadPointer(&pp.instance))
}

// VolatilePublication demonstrates volatile-like publication
func (pp *PublicationPattern) VolatilePublication(value int) {
    // In Go, atomic operations provide sufficient guarantees
    // for publication without explicit volatile semantics
    
    obj := NewSharedObject(value)
    atomic.StorePointer(&pp.instance, unsafe.Pointer(obj))
}

// SafeRead safely reads the published object
func (pp *PublicationPattern) SafeRead() *SharedObject {
    ptr := atomic.LoadPointer(&pp.instance)
    if ptr == nil {
        return nil
    }
    return (*SharedObject)(ptr)
}
```

## Lock-Free Data Structures

Implementation of common lock-free data structures.

### Lock-Free Stack

```go
// LockFreeStack implements a lock-free stack using atomic operations
type LockFreeStack struct {
    head unsafe.Pointer
    size int64
}

// StackNode represents a node in the lock-free stack
type StackNode struct {
    value interface{}
    next  unsafe.Pointer
}

// NewLockFreeStack creates a new lock-free stack
func NewLockFreeStack() *LockFreeStack {
    return &LockFreeStack{}
}

// Push adds an element to the stack
func (s *LockFreeStack) Push(value interface{}) {
    newNode := &StackNode{value: value}
    
    for {
        head := atomic.LoadPointer(&s.head)
        newNode.next = head
        
        if atomic.CompareAndSwapPointer(&s.head, head, unsafe.Pointer(newNode)) {
            atomic.AddInt64(&s.size, 1)
            break
        }
        // Retry on CAS failure
    }
}

// Pop removes and returns the top element from the stack
func (s *LockFreeStack) Pop() (interface{}, bool) {
    for {
        head := atomic.LoadPointer(&s.head)
        if head == nil {
            return nil, false // Stack is empty
        }
        
        node := (*StackNode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&s.head, head, next) {
            atomic.AddInt64(&s.size, -1)
            return node.value, true
        }
        // Retry on CAS failure
    }
}

// Size returns the approximate size of the stack
func (s *LockFreeStack) Size() int64 {
    return atomic.LoadInt64(&s.size)
}

// IsEmpty checks if the stack is empty
func (s *LockFreeStack) IsEmpty() bool {
    return atomic.LoadPointer(&s.head) == nil
}
```

### Lock-Free Queue

```go
// LockFreeQueue implements a lock-free queue using the Michael & Scott algorithm
type LockFreeQueue struct {
    head unsafe.Pointer
    tail unsafe.Pointer
    size int64
}

// QueueNode represents a node in the lock-free queue
type QueueNode struct {
    value interface{}
    next  unsafe.Pointer
}

// NewLockFreeQueue creates a new lock-free queue
func NewLockFreeQueue() *LockFreeQueue {
    dummy := &QueueNode{}
    queue := &LockFreeQueue{
        head: unsafe.Pointer(dummy),
        tail: unsafe.Pointer(dummy),
    }
    return queue
}

// Enqueue adds an element to the queue
func (q *LockFreeQueue) Enqueue(value interface{}) {
    newNode := &QueueNode{value: value}
    newNodePtr := unsafe.Pointer(newNode)
    
    for {
        tail := atomic.LoadPointer(&q.tail)
        tailNode := (*QueueNode)(tail)
        next := atomic.LoadPointer(&tailNode.next)
        
        // Check if tail is still the last node
        if tail == atomic.LoadPointer(&q.tail) {
            if next == nil {
                // Try to link new node at the end of the list
                if atomic.CompareAndSwapPointer(&tailNode.next, nil, newNodePtr) {
                    // Successfully added, now try to move tail
                    atomic.CompareAndSwapPointer(&q.tail, tail, newNodePtr)
                    atomic.AddInt64(&q.size, 1)
                    break
                }
            } else {
                // Help advance tail pointer
                atomic.CompareAndSwapPointer(&q.tail, tail, next)
            }
        }
    }
}

// Dequeue removes and returns an element from the queue
func (q *LockFreeQueue) Dequeue() (interface{}, bool) {
    for {
        head := atomic.LoadPointer(&q.head)
        tail := atomic.LoadPointer(&q.tail)
        headNode := (*QueueNode)(head)
        next := atomic.LoadPointer(&headNode.next)
        
        // Check if head is still the first node
        if head == atomic.LoadPointer(&q.head) {
            if head == tail {
                if next == nil {
                    // Queue is empty
                    return nil, false
                }
                // Help advance tail pointer
                atomic.CompareAndSwapPointer(&q.tail, tail, next)
            } else {
                if next == nil {
                    // This shouldn't happen in a correct implementation
                    continue
                }
                
                nextNode := (*QueueNode)(next)
                value := nextNode.value
                
                // Try to move head pointer
                if atomic.CompareAndSwapPointer(&q.head, head, next) {
                    atomic.AddInt64(&q.size, -1)
                    return value, true
                }
            }
        }
    }
}

// Size returns the approximate size of the queue
func (q *LockFreeQueue) Size() int64 {
    return atomic.LoadInt64(&q.size)
}

// IsEmpty checks if the queue is empty
func (q *LockFreeQueue) IsEmpty() bool {
    head := atomic.LoadPointer(&q.head)
    tail := atomic.LoadPointer(&q.tail)
    headNode := (*QueueNode)(head)
    next := atomic.LoadPointer(&headNode.next)
    
    return head == tail && next == nil
}
```

### Lock-Free Hash Map

```go
// LockFreeHashMap implements a lock-free hash map
type LockFreeHashMap struct {
    buckets []unsafe.Pointer
    size    int64
    mask    uint64
}

// HashMapNode represents a node in the hash map
type HashMapNode struct {
    key   string
    value interface{}
    hash  uint64
    next  unsafe.Pointer
}

// NewLockFreeHashMap creates a new lock-free hash map
func NewLockFreeHashMap(capacity int) *LockFreeHashMap {
    // Ensure capacity is power of 2
    if capacity&(capacity-1) != 0 {
        capacity = nextPowerOfTwo(capacity)
    }
    
    return &LockFreeHashMap{
        buckets: make([]unsafe.Pointer, capacity),
        mask:    uint64(capacity - 1),
    }
}

// nextPowerOfTwo returns the next power of 2 greater than or equal to n
func nextPowerOfTwo(n int) int {
    if n <= 1 {
        return 1
    }
    n--
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    n |= n >> 32
    return n + 1
}

// hash computes hash for a string key
func (hm *LockFreeHashMap) hash(key string) uint64 {
    // Simple FNV-1a hash
    hash := uint64(14695981039346656037)
    for _, b := range []byte(key) {
        hash ^= uint64(b)
        hash *= 1099511628211
    }
    return hash
}

// Put inserts or updates a key-value pair
func (hm *LockFreeHashMap) Put(key string, value interface{}) {
    hash := hm.hash(key)
    bucketIndex := hash & hm.mask
    newNode := &HashMapNode{
        key:   key,
        value: value,
        hash:  hash,
    }
    
    for {
        head := atomic.LoadPointer(&hm.buckets[bucketIndex])
        
        // Search for existing key
        current := head
        for current != nil {
            node := (*HashMapNode)(current)
            if node.hash == hash && node.key == key {
                // Update existing node's value
                // Note: This is a simplified version; a full implementation
                // would need to handle concurrent updates properly
                node.value = value
                return
            }
            current = atomic.LoadPointer(&node.next)
        }
        
        // Key not found, insert new node at head
        newNode.next = head
        if atomic.CompareAndSwapPointer(&hm.buckets[bucketIndex], head, unsafe.Pointer(newNode)) {
            atomic.AddInt64(&hm.size, 1)
            break
        }
        // Retry on CAS failure
    }
}

// Get retrieves a value by key
func (hm *LockFreeHashMap) Get(key string) (interface{}, bool) {
    hash := hm.hash(key)
    bucketIndex := hash & hm.mask
    
    head := atomic.LoadPointer(&hm.buckets[bucketIndex])
    current := head
    
    for current != nil {
        node := (*HashMapNode)(current)
        if node.hash == hash && node.key == key {
            return node.value, true
        }
        current = atomic.LoadPointer(&node.next)
    }
    
    return nil, false
}

// Delete removes a key-value pair
func (hm *LockFreeHashMap) Delete(key string) bool {
    hash := hm.hash(key)
    bucketIndex := hash & hm.mask
    
    for {
        head := atomic.LoadPointer(&hm.buckets[bucketIndex])
        if head == nil {
            return false // Key not found
        }
        
        headNode := (*HashMapNode)(head)
        if headNode.hash == hash && headNode.key == key {
            // Remove head node
            next := atomic.LoadPointer(&headNode.next)
            if atomic.CompareAndSwapPointer(&hm.buckets[bucketIndex], head, next) {
                atomic.AddInt64(&hm.size, -1)
                return true
            }
            // Retry on CAS failure
            continue
        }
        
        // Search in the chain
        prev := head
        current := atomic.LoadPointer(&headNode.next)
        
        for current != nil {
            node := (*HashMapNode)(current)
            if node.hash == hash && node.key == key {
                // Remove current node
                next := atomic.LoadPointer(&node.next)
                prevNode := (*HashMapNode)(prev)
                if atomic.CompareAndSwapPointer(&prevNode.next, current, next) {
                    atomic.AddInt64(&hm.size, -1)
                    return true
                }
                // Chain changed, restart
                break
            }
            prev = current
            current = atomic.LoadPointer(&node.next)
        }
        
        // Key not found in current snapshot, but chain might have changed
        // Check if head is still the same
        if head != atomic.LoadPointer(&hm.buckets[bucketIndex]) {
            continue // Restart if head changed
        }
        
        return false // Key not found
    }
}

// Size returns the approximate size of the hash map
func (hm *LockFreeHashMap) Size() int64 {
    return atomic.LoadInt64(&hm.size)
}
```

## ABA Problem and Solutions

Understanding and solving the ABA problem in lock-free programming.

### ABA Problem Demonstration

```go
// ABAProblematic demonstrates the ABA problem
type ABAProblematic struct {
    head unsafe.Pointer
}

// ABANode represents a node that can cause ABA problems
type ABANode struct {
    value int
    next  unsafe.Pointer
}

// ProblematicPop demonstrates potential ABA problem
func (aba *ABAProblematic) ProblematicPop() interface{} {
    for {
        head := atomic.LoadPointer(&aba.head) // Read A
        if head == nil {
            return nil
        }
        
        node := (*ABANode)(head)
        next := atomic.LoadPointer(&node.next)
        
        // Another thread could:
        // 1. Pop A
        // 2. Pop B  
        // 3. Push A back (reusing same memory)
        // Now head points to A again, but it's different A
        
        if atomic.CompareAndSwapPointer(&aba.head, head, next) { // Still reads A, but wrong A!
            return node.value
        }
    }
}
```

### Hazard Pointer Solution

```go
// HazardPointerSolution solves ABA using hazard pointers
type HazardPointerSolution struct {
    head           unsafe.Pointer
    hazardPointers []unsafe.Pointer
    retired        []unsafe.Pointer
    threadID       int64
}

// NewHazardPointerSolution creates a new hazard pointer solution
func NewHazardPointerSolution(maxThreads int) *HazardPointerSolution {
    return &HazardPointerSolution{
        hazardPointers: make([]unsafe.Pointer, maxThreads),
        retired:        make([]unsafe.Pointer, 0),
    }
}

// AcquireHazardPointer acquires a hazard pointer for a thread
func (hps *HazardPointerSolution) AcquireHazardPointer(ptr unsafe.Pointer) {
    threadID := hps.getThreadID()
    atomic.StorePointer(&hps.hazardPointers[threadID], ptr)
}

// ReleaseHazardPointer releases a hazard pointer for a thread
func (hps *HazardPointerSolution) ReleaseHazardPointer() {
    threadID := hps.getThreadID()
    atomic.StorePointer(&hps.hazardPointers[threadID], nil)
}

// getThreadID gets a thread ID (simplified)
func (hps *HazardPointerSolution) getThreadID() int {
    // In a real implementation, this would use thread-local storage
    // or a more sophisticated thread ID assignment
    return int(atomic.AddInt64(&hps.threadID, 1) % int64(len(hps.hazardPointers)))
}

// SafePop safely pops an element using hazard pointers
func (hps *HazardPointerSolution) SafePop() interface{} {
    for {
        head := atomic.LoadPointer(&hps.head)
        if head == nil {
            return nil
        }
        
        // Acquire hazard pointer
        hps.AcquireHazardPointer(head)
        
        // Re-read head to ensure it hasn't changed
        if head != atomic.LoadPointer(&hps.head) {
            hps.ReleaseHazardPointer()
            continue // Retry
        }
        
        node := (*ABANode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&hps.head, head, next) {
            hps.ReleaseHazardPointer()
            
            // Retire the node instead of freeing immediately
            hps.retireNode(head)
            
            return node.value
        }
        
        hps.ReleaseHazardPointer()
    }
}

// retireNode retires a node for later reclamation
func (hps *HazardPointerSolution) retireNode(node unsafe.Pointer) {
    hps.retired = append(hps.retired, node)
    
    // Periodically scan and reclaim safe nodes
    if len(hps.retired) > 100 {
        hps.reclaimRetiredNodes()
    }
}

// reclaimRetiredNodes reclaims nodes that are not hazardous
func (hps *HazardPointerSolution) reclaimRetiredNodes() {
    // Collect all current hazard pointers
    hazards := make(map[unsafe.Pointer]bool)
    for _, hp := range hps.hazardPointers {
        if hp != nil {
            hazards[hp] = true
        }
    }
    
    // Reclaim non-hazardous nodes
    newRetired := hps.retired[:0]
    for _, node := range hps.retired {
        if !hazards[node] {
            // Safe to reclaim
            // In a real implementation, would free the memory
        } else {
            // Still hazardous, keep in retired list
            newRetired = append(newRetired, node)
        }
    }
    hps.retired = newRetired
}
```

### Epoch-Based Solution

```go
// EpochBasedSolution solves ABA using epoch-based reclamation
type EpochBasedSolution struct {
    head         unsafe.Pointer
    globalEpoch  int64
    threadEpochs []int64
    retired      [][]unsafe.Pointer // retired[epoch % 3]
}

// NewEpochBasedSolution creates a new epoch-based solution
func NewEpochBasedSolution(maxThreads int) *EpochBasedSolution {
    return &EpochBasedSolution{
        threadEpochs: make([]int64, maxThreads),
        retired:      make([][]unsafe.Pointer, 3),
    }
}

// EnterCriticalSection enters a critical section
func (ebs *EpochBasedSolution) EnterCriticalSection(threadID int) {
    epoch := atomic.LoadInt64(&ebs.globalEpoch)
    atomic.StoreInt64(&ebs.threadEpochs[threadID], epoch)
}

// ExitCriticalSection exits a critical section
func (ebs *EpochBasedSolution) ExitCriticalSection(threadID int) {
    atomic.StoreInt64(&ebs.threadEpochs[threadID], -1)
    
    // Try to advance global epoch
    ebs.tryAdvanceEpoch()
}

// tryAdvanceEpoch attempts to advance the global epoch
func (ebs *EpochBasedSolution) tryAdvanceEpoch() {
    currentEpoch := atomic.LoadInt64(&ebs.globalEpoch)
    
    // Check if all threads have moved to current epoch
    allAdvanced := true
    for _, threadEpoch := range ebs.threadEpochs {
        epoch := atomic.LoadInt64(&threadEpoch)
        if epoch != -1 && epoch < currentEpoch {
            allAdvanced = false
            break
        }
    }
    
    if allAdvanced {
        // All threads have advanced, try to move to next epoch
        if atomic.CompareAndSwapInt64(&ebs.globalEpoch, currentEpoch, currentEpoch+1) {
            // Reclaim nodes from 3 epochs ago
            oldEpochIndex := (currentEpoch + 1) % 3
            for _, node := range ebs.retired[oldEpochIndex] {
                // Safe to reclaim
                _ = node // In real implementation, would free memory
            }
            ebs.retired[oldEpochIndex] = ebs.retired[oldEpochIndex][:0]
        }
    }
}

// SafePopWithEpoch safely pops using epoch-based reclamation
func (ebs *EpochBasedSolution) SafePopWithEpoch(threadID int) interface{} {
    ebs.EnterCriticalSection(threadID)
    defer ebs.ExitCriticalSection(threadID)
    
    for {
        head := atomic.LoadPointer(&ebs.head)
        if head == nil {
            return nil
        }
        
        node := (*ABANode)(head)
        next := atomic.LoadPointer(&node.next)
        
        if atomic.CompareAndSwapPointer(&ebs.head, head, next) {
            // Retire node to current epoch
            currentEpoch := atomic.LoadInt64(&ebs.globalEpoch)
            epochIndex := currentEpoch % 3
            ebs.retired[epochIndex] = append(ebs.retired[epochIndex], head)
            
            return node.value
        }
    }
}
```

## Performance Benchmarking

Comprehensive benchmarking of lock-free vs. lock-based solutions.

### Benchmark Framework

```go
// BenchmarkFramework provides comprehensive benchmarking for lock-free structures
type BenchmarkFramework struct {
    lockFreeStack  *LockFreeStack
    lockedStack    *LockedStack
    metrics       *BenchmarkMetrics
}

// BenchmarkMetrics tracks benchmark performance metrics
type BenchmarkMetrics struct {
    OperationsPerSecond map[string]float64
    AverageLatency     map[string]time.Duration
    MemoryUsage        map[string]int64
    CPUUsage           map[string]float64
    Scalability        map[string][]float64
}

// LockedStack implements a traditional locked stack for comparison
type LockedStack struct {
    items []interface{}
    mutex sync.Mutex
}

// NewLockedStack creates a new locked stack
func NewLockedStack() *LockedStack {
    return &LockedStack{
        items: make([]interface{}, 0),
    }
}

// Push adds an element to the locked stack
func (ls *LockedStack) Push(value interface{}) {
    ls.mutex.Lock()
    ls.items = append(ls.items, value)
    ls.mutex.Unlock()
}

// Pop removes an element from the locked stack
func (ls *LockedStack) Pop() (interface{}, bool) {
    ls.mutex.Lock()
    defer ls.mutex.Unlock()
    
    if len(ls.items) == 0 {
        return nil, false
    }
    
    value := ls.items[len(ls.items)-1]
    ls.items = ls.items[:len(ls.items)-1]
    return value, true
}

// NewBenchmarkFramework creates a new benchmark framework
func NewBenchmarkFramework() *BenchmarkFramework {
    return &BenchmarkFramework{
        lockFreeStack: NewLockFreeStack(),
        lockedStack:   NewLockedStack(),
        metrics: &BenchmarkMetrics{
            OperationsPerSecond: make(map[string]float64),
            AverageLatency:     make(map[string]time.Duration),
            MemoryUsage:        make(map[string]int64),
            CPUUsage:           make(map[string]float64),
            Scalability:        make(map[string][]float64),
        },
    }
}

// BenchmarkStackOperations benchmarks stack operations
func (bf *BenchmarkFramework) BenchmarkStackOperations(numGoroutines, numOperations int) {
    // Benchmark lock-free stack
    bf.benchmarkLockFreeStack(numGoroutines, numOperations)
    
    // Benchmark locked stack
    bf.benchmarkLockedStack(numGoroutines, numOperations)
}

// benchmarkLockFreeStack benchmarks lock-free stack performance
func (bf *BenchmarkFramework) benchmarkLockFreeStack(numGoroutines, numOperations int) {
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                bf.lockFreeStack.Push(j)
                bf.lockFreeStack.Pop()
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := float64(numGoroutines * numOperations * 2) // Push + Pop
    opsPerSecond := totalOps / duration.Seconds()
    avgLatency := duration / time.Duration(totalOps)
    
    bf.metrics.OperationsPerSecond["lockfree_stack"] = opsPerSecond
    bf.metrics.AverageLatency["lockfree_stack"] = avgLatency
}

// benchmarkLockedStack benchmarks locked stack performance
func (bf *BenchmarkFramework) benchmarkLockedStack(numGoroutines, numOperations int) {
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                bf.lockedStack.Push(j)
                bf.lockedStack.Pop()
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := float64(numGoroutines * numOperations * 2)
    opsPerSecond := totalOps / duration.Seconds()
    avgLatency := duration / time.Duration(totalOps)
    
    bf.metrics.OperationsPerSecond["locked_stack"] = opsPerSecond
    bf.metrics.AverageLatency["locked_stack"] = avgLatency
}

// BenchmarkScalability benchmarks scalability with varying goroutine counts
func (bf *BenchmarkFramework) BenchmarkScalability() {
    goroutineCounts := []int{1, 2, 4, 8, 16, 32, 64}
    operationsPerGoroutine := 10000
    
    lockFreeResults := make([]float64, len(goroutineCounts))
    lockedResults := make([]float64, len(goroutineCounts))
    
    for i, count := range goroutineCounts {
        // Reset stacks
        bf.lockFreeStack = NewLockFreeStack()
        bf.lockedStack = NewLockedStack()
        
        // Benchmark lock-free
        start := time.Now()
        bf.benchmarkLockFreeStack(count, operationsPerGoroutine)
        lockFreeResults[i] = bf.metrics.OperationsPerSecond["lockfree_stack"]
        
        // Benchmark locked
        start = time.Now()
        bf.benchmarkLockedStack(count, operationsPerGoroutine)
        lockedResults[i] = bf.metrics.OperationsPerSecond["locked_stack"]
    }
    
    bf.metrics.Scalability["lockfree_stack"] = lockFreeResults
    bf.metrics.Scalability["locked_stack"] = lockedResults
}

// GetResults returns benchmark results
func (bf *BenchmarkFramework) GetResults() *BenchmarkMetrics {
    return bf.metrics
}

// PrintResults prints benchmark results
func (bf *BenchmarkFramework) PrintResults() {
    fmt.Println("=== Lock-Free vs Locked Performance Comparison ===")
    
    for name, ops := range bf.metrics.OperationsPerSecond {
        latency := bf.metrics.AverageLatency[name]
        fmt.Printf("%s: %.2f ops/sec, %.2fμs avg latency\n", 
            name, ops, float64(latency.Nanoseconds())/1000.0)
    }
    
    fmt.Println("\n=== Scalability Results ===")
    if lockFree, exists := bf.metrics.Scalability["lockfree_stack"]; exists {
        fmt.Println("Goroutines\tLock-Free\tLocked\t\tSpeedup")
        locked := bf.metrics.Scalability["locked_stack"]
        goroutineCounts := []int{1, 2, 4, 8, 16, 32, 64}
        
        for i, count := range goroutineCounts {
            if i < len(lockFree) && i < len(locked) {
                speedup := lockFree[i] / locked[i]
                fmt.Printf("%d\t\t%.2f\t\t%.2f\t\t%.2fx\n", 
                    count, lockFree[i], locked[i], speedup)
            }
        }
    }
}
```

## Best Practices

Guidelines for effective lock-free programming in Go.

### Design Principles

```go
// LockFreeBestPractices demonstrates best practices for lock-free programming
type LockFreeBestPractices struct {
    examples map[string]interface{}
}

// NewLockFreeBestPractices creates a new best practices guide
func NewLockFreeBestPractices() *LockFreeBestPractices {
    return &LockFreeBestPractices{
        examples: make(map[string]interface{}),
    }
}

// MinimizeABAExposure demonstrates how to minimize ABA problem exposure
func (bp *LockFreeBestPractices) MinimizeABAExposure() {
    // Principle 1: Use generation counters or tags
    type TaggedPointer struct {
        ptr unsafe.Pointer
        tag uint64
    }
    
    // Principle 2: Avoid pointer reuse
    // Use memory pools instead of immediate reallocation
    
    // Principle 3: Use hazard pointers for safe memory reclamation
}

// OptimizeForCommonCase demonstrates optimizing for the common case
func (bp *LockFreeBestPractices) OptimizeForCommonCase() {
    // Fast path for common operations
    fastPathExample := func(value int64) bool {
        // Try optimistic approach first
        if atomic.CompareAndSwapInt64(&value, 0, 1) {
            return true // Fast path succeeded
        }
        
        // Fall back to slower but more general approach
        return bp.slowPathOperation(&value)
    }
    
    _ = fastPathExample
}

// slowPathOperation provides fallback for complex cases
func (bp *LockFreeBestPractices) slowPathOperation(value *int64) bool {
    // More complex logic for edge cases
    for {
        old := atomic.LoadInt64(value)
        if old != 0 {
            return false
        }
        
        if atomic.CompareAndSwapInt64(value, old, 1) {
            return true
        }
        // Retry
    }
}

// UseProgressGuarantees demonstrates different progress guarantees
func (bp *LockFreeBestPractices) UseProgressGuarantees() {
    // Wait-free: All operations complete in bounded steps
    waitFreeCounter := func(counter *int64) int64 {
        return atomic.AddInt64(counter, 1) // Always completes
    }
    
    // Lock-free: System-wide progress guaranteed
    lockFreeOperation := func(head *unsafe.Pointer) bool {
        for {
            old := atomic.LoadPointer(head)
            if atomic.CompareAndSwapPointer(head, old, nil) {
                return true
            }
            // Some thread will make progress
        }
    }
    
    // Obstruction-free: Progress when running alone
    obstructionFreeOperation := func(value *int64) bool {
        old := atomic.LoadInt64(value)
        return atomic.CompareAndSwapInt64(value, old, old+1)
    }
    
    _ = waitFreeCounter
    _ = lockFreeOperation
    _ = obstructionFreeOperation
}

// MemoryManagementBestPractices demonstrates safe memory management
func (bp *LockFreeBestPractices) MemoryManagementBestPractices() {
    // Use object pools to avoid allocation overhead
    type ObjectPool struct {
        pool sync.Pool
    }
    
    pool := &ObjectPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &ABANode{}
            },
        },
    }
    
    // Get object from pool
    getObject := func() *ABANode {
        return pool.pool.Get().(*ABANode)
    }
    
    // Return object to pool
    putObject := func(obj *ABANode) {
        // Reset object state
        obj.value = 0
        obj.next = nil
        pool.pool.Put(obj)
    }
    
    _ = getObject
    _ = putObject
}

// ErrorHandlingPatterns demonstrates error handling in lock-free code
func (bp *LockFreeBestPractices) ErrorHandlingPatterns() {
    // Pattern 1: Return success/failure status
    tryOperation := func(value *int64) bool {
        old := atomic.LoadInt64(value)
        if old > 100 {
            return false // Precondition failed
        }
        return atomic.CompareAndSwapInt64(value, old, old+1)
    }
    
    // Pattern 2: Use sentinel values for errors
    operationWithSentinel := func(value *int64) int64 {
        const ErrorValue = -1
        
        old := atomic.LoadInt64(value)
        if old < 0 {
            return ErrorValue
        }
        
        if atomic.CompareAndSwapInt64(value, old, old*2) {
            return old * 2
        }
        
        return ErrorValue // CAS failed
    }
    
    // Pattern 3: Retry with exponential backoff
    operationWithBackoff := func(value *int64) bool {
        backoff := 1
        maxRetries := 100
        
        for retry := 0; retry < maxRetries; retry++ {
            old := atomic.LoadInt64(value)
            if atomic.CompareAndSwapInt64(value, old, old+1) {
                return true
            }
            
            // Exponential backoff
            for i := 0; i < backoff; i++ {
                runtime.Gosched()
            }
            
            backoff = min(backoff*2, 64)
        }
        
        return false // Failed after max retries
    }
    
    _ = tryOperation
    _ = operationWithSentinel
    _ = operationWithBackoff
}

// TestingStrategies demonstrates testing strategies for lock-free code
func (bp *LockFreeBestPractices) TestingStrategies() {
    // Strategy 1: Stress testing with many goroutines
    stressTest := func(operation func()) {
        numGoroutines := runtime.NumCPU() * 4
        iterations := 10000
        
        var wg sync.WaitGroup
        for i := 0; i < numGoroutines; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < iterations; j++ {
                    operation()
                }
            }()
        }
        wg.Wait()
    }
    
    // Strategy 2: Property-based testing
    propertyTest := func(property func() bool) bool {
        // Run property check many times with different schedules
        for i := 0; i < 1000; i++ {
            if !property() {
                return false
            }
            runtime.Gosched() // Yield to encourage different interleavings
        }
        return true
    }
    
    // Strategy 3: Race detection
    // Always run tests with -race flag
    
    _ = stressTest
    _ = propertyTest
}

// Helper function
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Summary

Lock-free programming in Go provides powerful tools for building high-performance concurrent systems. Key takeaways:

1. **Use Atomic Operations**: Leverage Go's atomic package for thread-safe operations
2. **Understand Memory Ordering**: Go provides sequential consistency for atomic operations
3. **Handle ABA Problem**: Use hazard pointers or epoch-based reclamation
4. **Optimize for Common Cases**: Design fast paths for typical scenarios
5. **Test Thoroughly**: Use stress testing and race detection
6. **Choose Right Guarantees**: Select appropriate progress guarantees for your use case

Lock-free programming requires careful design and thorough testing, but can provide significant performance benefits in high-contention scenarios.
