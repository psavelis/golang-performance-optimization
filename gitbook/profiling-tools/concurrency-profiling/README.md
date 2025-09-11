# Mutex & Block Profiling

Concurrency profiling in Go focuses on analyzing synchronization primitives, identifying contention points, and optimizing parallel performance. This comprehensive guide covers mutex profiling, block profiling, and channel analysis for building efficient concurrent Go applications.

## Introduction to Concurrency Profiling

Concurrency profiling helps identify:
- **Mutex contention** - Where threads compete for locks
- **Blocking operations** - Time spent waiting on synchronization
- **Channel bottlenecks** - Inefficient channel usage patterns
- **Goroutine coordination** - Overhead in concurrent operations
- **Lock-free opportunities** - Places to eliminate synchronization

### Understanding Go Concurrency Primitives

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

// ConcurrencyProfiler provides comprehensive concurrency analysis
type ConcurrencyProfiler struct {
    mutexProfile    *MutexProfiler
    blockProfile    *BlockProfiler
    channelProfile  *ChannelProfiler
    enabled         bool
    mu              sync.RWMutex
}

func NewConcurrencyProfiler() *ConcurrencyProfiler {
    return &ConcurrencyProfiler{
        mutexProfile:   NewMutexProfiler(),
        blockProfile:   NewBlockProfiler(),
        channelProfile: NewChannelProfiler(),
    }
}

func (cp *ConcurrencyProfiler) Enable() {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if cp.enabled {
        return
    }
    
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
    
    // Enable block profiling
    runtime.SetBlockProfileRate(1)
    
    cp.enabled = true
    fmt.Println("Concurrency profiling enabled")
}

func (cp *ConcurrencyProfiler) Disable() {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if !cp.enabled {
        return
    }
    
    // Disable profiling
    runtime.SetMutexProfileFraction(0)
    runtime.SetBlockProfileRate(0)
    
    cp.enabled = false
    fmt.Println("Concurrency profiling disabled")
}

func (cp *ConcurrencyProfiler) GetComprehensiveReport() ConcurrencyReport {
    return ConcurrencyReport{
        MutexReport:   cp.mutexProfile.GetReport(),
        BlockReport:   cp.blockProfile.GetReport(),
        ChannelReport: cp.channelProfile.GetReport(),
        Enabled:       cp.enabled,
    }
}

type ConcurrencyReport struct {
    MutexReport   MutexReport
    BlockReport   BlockReport
    ChannelReport ChannelReport
    Enabled       bool
}

func (cr ConcurrencyReport) String() string {
    result := fmt.Sprintf("Concurrency Profiling Report (Enabled: %v)\n", cr.Enabled)
    result += "=" + strings.Repeat("=", 50) + "\n\n"
    result += cr.MutexReport.String() + "\n\n"
    result += cr.BlockReport.String() + "\n\n"
    result += cr.ChannelReport.String()
    return result
}

// MutexProfiler tracks mutex contention
type MutexProfiler struct {
    mutexes    map[string]*MutexStats
    mu         sync.RWMutex
    totalLocks int64
    totalTime  int64
}

type MutexStats struct {
    Name          string
    Contentions   int64
    ContentionTime int64
    AvgWaitTime   time.Duration
    MaxWaitTime   time.Duration
    FirstSeen     time.Time
    LastSeen      time.Time
}

func NewMutexProfiler() *MutexProfiler {
    return &MutexProfiler{
        mutexes: make(map[string]*MutexStats),
    }
}

func (mp *MutexProfiler) RecordContention(mutexName string, waitTime time.Duration) {
    mp.mu.Lock()
    defer mp.mu.Unlock()
    
    stats, exists := mp.mutexes[mutexName]
    if !exists {
        stats = &MutexStats{
            Name:      mutexName,
            FirstSeen: time.Now(),
        }
        mp.mutexes[mutexName] = stats
    }
    
    stats.Contentions++
    stats.ContentionTime += int64(waitTime)
    stats.LastSeen = time.Now()
    stats.AvgWaitTime = time.Duration(stats.ContentionTime / stats.Contentions)
    
    if waitTime > stats.MaxWaitTime {
        stats.MaxWaitTime = waitTime
    }
    
    atomic.AddInt64(&mp.totalLocks, 1)
    atomic.AddInt64(&mp.totalTime, int64(waitTime))
}

func (mp *MutexProfiler) GetReport() MutexReport {
    mp.mu.RLock()
    defer mp.mu.RUnlock()
    
    var mutexStats []MutexStats
    for _, stats := range mp.mutexes {
        mutexStats = append(mutexStats, *stats)
    }
    
    // Sort by contention time
    sort.Slice(mutexStats, func(i, j int) bool {
        return mutexStats[i].ContentionTime > mutexStats[j].ContentionTime
    })
    
    return MutexReport{
        TotalMutexes:   len(mutexStats),
        TotalLocks:     atomic.LoadInt64(&mp.totalLocks),
        TotalWaitTime:  time.Duration(atomic.LoadInt64(&mp.totalTime)),
        MutexStats:     mutexStats,
    }
}

type MutexReport struct {
    TotalMutexes  int
    TotalLocks    int64
    TotalWaitTime time.Duration
    MutexStats    []MutexStats
}

func (mr MutexReport) String() string {
    result := fmt.Sprintf(`Mutex Contention Report:
  Total Mutexes: %d
  Total Locks: %d
  Total Wait Time: %v`,
        mr.TotalMutexes,
        mr.TotalLocks,
        mr.TotalWaitTime)
    
    if len(mr.MutexStats) > 0 {
        result += "\n\nTop Contended Mutexes:"
        for i, stats := range mr.MutexStats {
            if i >= 5 { // Top 5
                break
            }
            result += fmt.Sprintf("\n  %d. %s", i+1, stats.Name)
            result += fmt.Sprintf("\n     Contentions: %d", stats.Contentions)
            result += fmt.Sprintf("\n     Total Wait: %v", time.Duration(stats.ContentionTime))
            result += fmt.Sprintf("\n     Avg Wait: %v", stats.AvgWaitTime)
            result += fmt.Sprintf("\n     Max Wait: %v", stats.MaxWaitTime)
        }
    }
    
    return result
}

// BlockProfiler tracks blocking operations
type BlockProfiler struct {
    blockEvents map[string]*BlockStats
    mu          sync.RWMutex
    totalBlocks int64
    totalTime   int64
}

type BlockStats struct {
    Operation   string
    Count       int64
    TotalTime   int64
    AvgTime     time.Duration
    MaxTime     time.Duration
    FirstSeen   time.Time
    LastSeen    time.Time
}

func NewBlockProfiler() *BlockProfiler {
    return &BlockProfiler{
        blockEvents: make(map[string]*BlockStats),
    }
}

func (bp *BlockProfiler) RecordBlock(operation string, blockTime time.Duration) {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    stats, exists := bp.blockEvents[operation]
    if !exists {
        stats = &BlockStats{
            Operation: operation,
            FirstSeen: time.Now(),
        }
        bp.blockEvents[operation] = stats
    }
    
    stats.Count++
    stats.TotalTime += int64(blockTime)
    stats.LastSeen = time.Now()
    stats.AvgTime = time.Duration(stats.TotalTime / stats.Count)
    
    if blockTime > stats.MaxTime {
        stats.MaxTime = blockTime
    }
    
    atomic.AddInt64(&bp.totalBlocks, 1)
    atomic.AddInt64(&bp.totalTime, int64(blockTime))
}

func (bp *BlockProfiler) GetReport() BlockReport {
    bp.mu.RLock()
    defer bp.mu.RUnlock()
    
    var blockStats []BlockStats
    for _, stats := range bp.blockEvents {
        blockStats = append(blockStats, *stats)
    }
    
    // Sort by total time
    sort.Slice(blockStats, func(i, j int) bool {
        return blockStats[i].TotalTime > blockStats[j].TotalTime
    })
    
    return BlockReport{
        TotalOperations: len(blockStats),
        TotalBlocks:     atomic.LoadInt64(&bp.totalBlocks),
        TotalBlockTime:  time.Duration(atomic.LoadInt64(&bp.totalTime)),
        BlockStats:      blockStats,
    }
}

type BlockReport struct {
    TotalOperations int
    TotalBlocks     int64
    TotalBlockTime  time.Duration
    BlockStats      []BlockStats
}

func (br BlockReport) String() string {
    result := fmt.Sprintf(`Block Profiling Report:
  Total Operations: %d
  Total Blocks: %d
  Total Block Time: %v`,
        br.TotalOperations,
        br.TotalBlocks,
        br.TotalBlockTime)
    
    if len(br.BlockStats) > 0 {
        result += "\n\nTop Blocking Operations:"
        for i, stats := range br.BlockStats {
            if i >= 5 { // Top 5
                break
            }
            result += fmt.Sprintf("\n  %d. %s", i+1, stats.Operation)
            result += fmt.Sprintf("\n     Blocks: %d", stats.Count)
            result += fmt.Sprintf("\n     Total Time: %v", time.Duration(stats.TotalTime))
            result += fmt.Sprintf("\n     Avg Time: %v", stats.AvgTime)
            result += fmt.Sprintf("\n     Max Time: %v", stats.MaxTime)
        }
    }
    
    return result
}

// ChannelProfiler tracks channel operations
type ChannelProfiler struct {
    channels map[string]*ChannelStats
    mu       sync.RWMutex
}

type ChannelStats struct {
    Name        string
    SendOps     int64
    RecvOps     int64
    SendBlocks  int64
    RecvBlocks  int64
    SendTime    int64
    RecvTime    int64
    BufferSize  int
    FirstSeen   time.Time
    LastSeen    time.Time
}

func NewChannelProfiler() *ChannelProfiler {
    return &ChannelProfiler{
        channels: make(map[string]*ChannelStats),
    }
}

func (cp *ChannelProfiler) RecordChannelOp(channelName string, operation string, blocked bool, duration time.Duration, bufferSize int) {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    stats, exists := cp.channels[channelName]
    if !exists {
        stats = &ChannelStats{
            Name:       channelName,
            BufferSize: bufferSize,
            FirstSeen:  time.Now(),
        }
        cp.channels[channelName] = stats
    }
    
    stats.LastSeen = time.Now()
    
    switch operation {
    case "send":
        stats.SendOps++
        if blocked {
            stats.SendBlocks++
            stats.SendTime += int64(duration)
        }
    case "recv":
        stats.RecvOps++
        if blocked {
            stats.RecvBlocks++
            stats.RecvTime += int64(duration)
        }
    }
}

func (cp *ChannelProfiler) GetReport() ChannelReport {
    cp.mu.RLock()
    defer cp.mu.RUnlock()
    
    var channelStats []ChannelStats
    for _, stats := range cp.channels {
        channelStats = append(channelStats, *stats)
    }
    
    // Sort by total operations
    sort.Slice(channelStats, func(i, j int) bool {
        return (channelStats[i].SendOps + channelStats[i].RecvOps) > 
               (channelStats[j].SendOps + channelStats[j].RecvOps)
    })
    
    return ChannelReport{
        TotalChannels: len(channelStats),
        ChannelStats:  channelStats,
    }
}

type ChannelReport struct {
    TotalChannels int
    ChannelStats  []ChannelStats
}

func (cr ChannelReport) String() string {
    result := fmt.Sprintf("Channel Analysis Report:\n  Total Channels: %d", cr.TotalChannels)
    
    if len(cr.ChannelStats) > 0 {
        result += "\n\nChannel Usage:"
        for i, stats := range cr.ChannelStats {
            if i >= 5 { // Top 5
                break
            }
            
            sendBlockRate := float64(0)
            if stats.SendOps > 0 {
                sendBlockRate = float64(stats.SendBlocks) / float64(stats.SendOps) * 100
            }
            
            recvBlockRate := float64(0)
            if stats.RecvOps > 0 {
                recvBlockRate = float64(stats.RecvBlocks) / float64(stats.RecvOps) * 100
            }
            
            result += fmt.Sprintf("\n  %d. %s (buffer: %d)", i+1, stats.Name, stats.BufferSize)
            result += fmt.Sprintf("\n     Send Ops: %d (%.1f%% blocked)", stats.SendOps, sendBlockRate)
            result += fmt.Sprintf("\n     Recv Ops: %d (%.1f%% blocked)", stats.RecvOps, recvBlockRate)
            result += fmt.Sprintf("\n     Send Time: %v", time.Duration(stats.SendTime))
            result += fmt.Sprintf("\n     Recv Time: %v", time.Duration(stats.RecvTime))
        }
    }
    
    return result
}

func demonstrateConcurrencyProfiling() {
    fmt.Println("=== COMPREHENSIVE CONCURRENCY PROFILING ===")
    
    profiler := NewConcurrencyProfiler()
    profiler.Enable()
    defer profiler.Disable()
    
    // Simulate various concurrency scenarios
    simulateMutexContention(profiler)
    simulateChannelOperations(profiler)
    simulateBlockingOperations(profiler)
    
    // Get comprehensive report
    report := profiler.GetComprehensiveReport()
    fmt.Printf("\n%s\n", report)
}

func simulateMutexContention(profiler *ConcurrencyProfiler) {
    fmt.Println("Simulating mutex contention...")
    
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                start := time.Now()
                mu.Lock()
                
                // Record contention if we waited
                waitTime := time.Since(start)
                if waitTime > time.Microsecond {
                    profiler.mutexProfile.RecordContention("main_mutex", waitTime)
                }
                
                // Hold lock for some time
                time.Sleep(time.Microsecond * 100)
                mu.Unlock()
                
                time.Sleep(time.Microsecond * 50)
            }
        }(i)
    }
    
    wg.Wait()
}

func simulateChannelOperations(profiler *ConcurrencyProfiler) {
    fmt.Println("Simulating channel operations...")
    
    ch := make(chan int, 5)
    var wg sync.WaitGroup
    
    // Producer
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(ch)
        
        for i := 0; i < 100; i++ {
            start := time.Now()
            
            select {
            case ch <- i:
                // Non-blocking
                profiler.channelProfile.RecordChannelOp("test_channel", "send", false, 0, 5)
            default:
                // Would block
                ch <- i // Actually send (will block)
                duration := time.Since(start)
                profiler.channelProfile.RecordChannelOp("test_channel", "send", true, duration, 5)
            }
            
            time.Sleep(time.Millisecond * 2)
        }
    }()
    
    // Consumer
    wg.Add(1)
    go func() {
        defer wg.Done()
        
        for val := range ch {
            start := time.Now()
            _ = val
            
            // Simulate processing time
            time.Sleep(time.Millisecond * 5)
            
            duration := time.Since(start)
            profiler.channelProfile.RecordChannelOp("test_channel", "recv", true, duration, 5)
        }
    }()
    
    wg.Wait()
}

func simulateBlockingOperations(profiler *ConcurrencyProfiler) {
    fmt.Println("Simulating blocking operations...")
    
    var wg sync.WaitGroup
    
    // Simulate different blocking scenarios
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < 20; j++ {
                // Simulate I/O operation
                start := time.Now()
                time.Sleep(time.Millisecond * time.Duration(10+id*2))
                duration := time.Since(start)
                
                profiler.blockProfile.RecordBlock("io_operation", duration)
                
                // Simulate network call
                start = time.Now()
                time.Sleep(time.Millisecond * time.Duration(5+id))
                duration = time.Since(start)
                
                profiler.blockProfile.RecordBlock("network_call", duration)
            }
        }(i)
    }
    
    wg.Wait()
}
```

## Key Profiling Techniques

### 1. Mutex Profiling
- **Purpose**: Identify lock contention hotspots
- **Metrics**: Contention time, frequency, wait times
- **Tools**: `runtime.SetMutexProfileFraction()`

### 2. Block Profiling  
- **Purpose**: Find blocking synchronization operations
- **Metrics**: Block time, frequency, operation types
- **Tools**: `runtime.SetBlockProfileRate()`

### 3. Channel Analysis
- **Purpose**: Optimize channel usage patterns
- **Metrics**: Send/receive rates, blocking frequency
- **Tools**: Custom instrumentation

## Optimization Strategies

### Lock-Free Alternatives
```go
// Replace mutex with atomic operations where possible
var counter int64
atomic.AddInt64(&counter, 1) // Instead of mutex-protected increment
```

### Channel Buffer Optimization
```go
// Size buffers appropriately
ch := make(chan Work, runtime.NumCPU()) // CPU-bound work
ch := make(chan Request, 1000)          // High-throughput scenarios
```

### Reduced Lock Scope
```go
// Minimize time holding locks
func optimizedFunction() {
    // Prepare data outside lock
    data := prepareData()
    
    mu.Lock()
    // Minimal work under lock
    updateSharedState(data)
    mu.Unlock()
}
```

## Analysis Tools

### Built-in Profiling
```bash
# Enable mutex profiling
go tool pprof http://localhost:6060/debug/pprof/mutex

# Enable block profiling  
go tool pprof http://localhost:6060/debug/pprof/block
```

### Custom Instrumentation
```go
// Wrap mutexes with timing
type TimedMutex struct {
    sync.Mutex
    name string
}

func (tm *TimedMutex) Lock() {
    start := time.Now()
    tm.Mutex.Lock()
    recordContentionTime(tm.name, time.Since(start))
}
```

## Best Practices

1. **Profile Representative Workloads** - Use realistic concurrency patterns
2. **Focus on Hotspots** - Optimize the highest contention points first  
3. **Measure Before/After** - Validate optimization effectiveness
4. **Consider Lock-Free** - Atomic operations for simple cases
5. **Right-Size Buffers** - Match channel buffers to usage patterns

## Next Steps

- Study [Mutex Contention](mutex-contention.md) analysis
- Learn [Block Profiling](block-profiling.md) techniques  
- Explore [Channel Analysis](channel-analysis.md) optimization
- Master [Concurrency Optimization](../../optimization/concurrency/README.md)

## Summary

Concurrency profiling enables building efficient parallel Go applications by:

1. **Identifying contention** - Finding synchronization bottlenecks
2. **Measuring overhead** - Quantifying coordination costs
3. **Guiding optimization** - Focusing on impactful improvements
4. **Validating changes** - Confirming performance benefits
5. **Scaling effectively** - Building for concurrent workloads

Use these techniques to optimize concurrent applications for maximum throughput and minimal contention.
