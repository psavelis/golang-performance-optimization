# Block Profiling

Block profiling in Go helps identify synchronization bottlenecks by measuring how long goroutines spend blocked on synchronization primitives like mutexes, channels, and select statements. This comprehensive guide covers advanced block profiling techniques for optimizing concurrent Go applications.

## Introduction to Block Profiling

Block profiling tracks:
- **Mutex contention** - Time spent waiting for mutex locks
- **Channel operations** - Blocking on channel send/receive
- **Select statements** - Waiting in select cases
- **Semaphore operations** - Blocking on semaphores
- **Condition variables** - Waiting on sync.Cond

### Enabling Block Profiling

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "sync"
    "time"
)

func init() {
    // Enable block profiling
    runtime.SetBlockProfileRate(1)
}

// BlockProfiler provides comprehensive block profiling capabilities
type BlockProfiler struct {
    samplingRate    int
    profileDuration time.Duration
    outputPath      string
    serverAddr      string
    collecting      bool
    mu              sync.RWMutex
}

func NewBlockProfiler() *BlockProfiler {
    return &BlockProfiler{
        samplingRate:    1,     // Sample every nanosecond of blocking
        profileDuration: 30 * time.Second,
        outputPath:      "block.prof",
        serverAddr:      ":6060",
    }
}

func (bp *BlockProfiler) SetSamplingRate(rate int) {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    bp.samplingRate = rate
    runtime.SetBlockProfileRate(rate)
}

func (bp *BlockProfiler) StartProfiling() error {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    if bp.collecting {
        return fmt.Errorf("profiling already in progress")
    }
    
    // Set block profile rate
    runtime.SetBlockProfileRate(bp.samplingRate)
    bp.collecting = true
    
    fmt.Printf("Block profiling started (rate: %d, duration: %v)\n", 
        bp.samplingRate, bp.profileDuration)
    
    return nil
}

func (bp *BlockProfiler) StopProfiling() error {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    if !bp.collecting {
        return fmt.Errorf("no profiling in progress")
    }
    
    // Disable block profiling
    runtime.SetBlockProfileRate(0)
    bp.collecting = false
    
    fmt.Println("Block profiling stopped")
    return nil
}

func (bp *BlockProfiler) StartServer() {
    fmt.Printf("Starting pprof server on %s\n", bp.serverAddr)
    fmt.Printf("Block profile: http://localhost%s/debug/pprof/block\n", bp.serverAddr)
    
    log.Fatal(http.ListenAndServe(bp.serverAddr, nil))
}

func (bp *BlockProfiler) CollectProfile(duration time.Duration) error {
    if err := bp.StartProfiling(); err != nil {
        return err
    }
    
    time.Sleep(duration)
    
    return bp.StopProfiling()
}

func demonstrateBasicBlockProfiling() {
    fmt.Println("=== BASIC BLOCK PROFILING ===")
    
    profiler := NewBlockProfiler()
    profiler.SetSamplingRate(1) // Sample every blocking event
    
    // Start profiling server in background
    go profiler.StartServer()
    
    // Start profiling
    profiler.StartProfiling()
    
    // Create some blocking scenarios
    var wg sync.WaitGroup
    
    // Mutex contention
    var mu sync.Mutex
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 10; j++ {
                mu.Lock()
                time.Sleep(10 * time.Millisecond) // Hold lock
                mu.Unlock()
                time.Sleep(5 * time.Millisecond)
            }
        }(i)
    }
    
    wg.Wait()
    
    // Channel blocking
    ch := make(chan int)
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 5; j++ {
                ch <- id*10 + j
            }
        }(i)
    }
    
    go func() {
        for i := 0; i < 15; i++ {
            <-ch
            time.Sleep(20 * time.Millisecond)
        }
    }()
    
    wg.Wait()
    
    time.Sleep(2 * time.Second)
    profiler.StopProfiling()
    
    fmt.Println("Block profiling completed. Check pprof server for results.")
}
```

## Advanced Block Analysis Techniques

### Mutex Contention Analysis

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
)

// MutexContentionAnalyzer tracks mutex contention patterns
type MutexContentionAnalyzer struct {
    mutexes      map[string]*MutexStats
    mu           sync.RWMutex
    totalBlocked int64
    totalTime    int64
}

type MutexStats struct {
    Name         string
    LockCount    int64
    BlockCount   int64
    BlockTime    int64
    ContentionRate float64
}

func NewMutexContentionAnalyzer() *MutexContentionAnalyzer {
    return &MutexContentionAnalyzer{
        mutexes: make(map[string]*MutexStats),
    }
}

// TrackedMutex wraps sync.Mutex with contention tracking
type TrackedMutex struct {
    sync.Mutex
    name     string
    analyzer *MutexContentionAnalyzer
    locks    int64
    blocks   int64
    blockTime int64
}

func (mca *MutexContentionAnalyzer) NewTrackedMutex(name string) *TrackedMutex {
    tm := &TrackedMutex{
        name:     name,
        analyzer: mca,
    }
    
    mca.mu.Lock()
    mca.mutexes[name] = &MutexStats{Name: name}
    mca.mu.Unlock()
    
    return tm
}

func (tm *TrackedMutex) Lock() {
    start := time.Now()
    
    // Try to lock without blocking first
    if tm.Mutex.TryLock() {
        atomic.AddInt64(&tm.locks, 1)
        return
    }
    
    // If we can't get it immediately, it's contended
    atomic.AddInt64(&tm.blocks, 1)
    tm.Mutex.Lock()
    
    blockDuration := time.Since(start)
    atomic.AddInt64(&tm.blockTime, int64(blockDuration))
    atomic.AddInt64(&tm.locks, 1)
    
    // Update global stats
    atomic.AddInt64(&tm.analyzer.totalBlocked, 1)
    atomic.AddInt64(&tm.analyzer.totalTime, int64(blockDuration))
}

func (tm *TrackedMutex) TryLock() bool {
    if tm.Mutex.TryLock() {
        atomic.AddInt64(&tm.locks, 1)
        return true
    }
    return false
}

func (tm *TrackedMutex) GetStats() MutexStats {
    locks := atomic.LoadInt64(&tm.locks)
    blocks := atomic.LoadInt64(&tm.blocks)
    blockTime := atomic.LoadInt64(&tm.blockTime)
    
    var contentionRate float64
    if locks > 0 {
        contentionRate = float64(blocks) / float64(locks) * 100
    }
    
    return MutexStats{
        Name:           tm.name,
        LockCount:      locks,
        BlockCount:     blocks,
        BlockTime:      blockTime,
        ContentionRate: contentionRate,
    }
}

func (mca *MutexContentionAnalyzer) GetReport() ContentionReport {
    mca.mu.RLock()
    defer mca.mu.RUnlock()
    
    var mutexStats []MutexStats
    totalContentions := int64(0)
    totalTime := time.Duration(0)
    
    for _, tm := range mca.mutexes {
        // Get current stats from TrackedMutex
        // This is a simplified version - in practice you'd maintain references
        stats := MutexStats{
            Name:           tm.Name,
            LockCount:      0, // Would be populated from actual TrackedMutex
            BlockCount:     0,
            BlockTime:      0,
            ContentionRate: 0,
        }
        mutexStats = append(mutexStats, stats)
        totalContentions += stats.BlockCount
        totalTime += time.Duration(stats.BlockTime)
    }
    
    return ContentionReport{
        TotalMutexes:     len(mutexStats),
        TotalContentions: totalContentions,
        TotalBlockTime:   totalTime,
        MutexStats:       mutexStats,
    }
}

type ContentionReport struct {
    TotalMutexes     int
    TotalContentions int64
    TotalBlockTime   time.Duration
    MutexStats       []MutexStats
}

func (cr ContentionReport) String() string {
    result := fmt.Sprintf(`Mutex Contention Report:
  Total Mutexes: %d
  Total Contentions: %d
  Total Block Time: %v`,
        cr.TotalMutexes,
        cr.TotalContentions,
        cr.TotalBlockTime)
    
    if len(cr.MutexStats) > 0 {
        result += "\n\nPer-Mutex Statistics:"
        for _, stats := range cr.MutexStats {
            result += fmt.Sprintf("\n  %s:", stats.Name)
            result += fmt.Sprintf("\n    Locks: %d", stats.LockCount)
            result += fmt.Sprintf("\n    Blocks: %d", stats.BlockCount)
            result += fmt.Sprintf("\n    Block Time: %v", time.Duration(stats.BlockTime))
            result += fmt.Sprintf("\n    Contention Rate: %.2f%%", stats.ContentionRate)
        }
    }
    
    return result
}

// Demonstration of different contention patterns
func demonstrateContentionPatterns() {
    fmt.Println("\n=== MUTEX CONTENTION PATTERNS ===")
    
    analyzer := NewMutexContentionAnalyzer()
    
    // High contention scenario
    highContentionMutex := analyzer.NewTrackedMutex("high_contention")
    
    var wg sync.WaitGroup
    numGoroutines := 10
    operationsPerGoroutine := 100
    
    fmt.Println("Testing high contention scenario...")
    
    start := time.Now()
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < operationsPerGoroutine; j++ {
                highContentionMutex.Lock()
                // Simulate work while holding lock
                time.Sleep(time.Microsecond * 100)
                highContentionMutex.Unlock()
                
                // Small gap between operations
                time.Sleep(time.Microsecond * 10)
            }
        }(i)
    }
    wg.Wait()
    highContentionTime := time.Since(start)
    
    // Low contention scenario
    lowContentionMutex := analyzer.NewTrackedMutex("low_contention")
    
    fmt.Println("Testing low contention scenario...")
    
    start = time.Now()
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < operationsPerGoroutine; j++ {
                lowContentionMutex.Lock()
                // Very brief work
                time.Sleep(time.Microsecond * 1)
                lowContentionMutex.Unlock()
                
                // Longer gap between operations
                time.Sleep(time.Microsecond * 500)
            }
        }(i)
    }
    wg.Wait()
    lowContentionTime := time.Since(start)
    
    // Show results
    fmt.Printf("\nHigh contention stats:\n%s\n", highContentionMutex.GetStats())
    fmt.Printf("Low contention stats:\n%s\n", lowContentionMutex.GetStats())
    
    fmt.Printf("\nTiming comparison:\n")
    fmt.Printf("  High contention: %v\n", highContentionTime)
    fmt.Printf("  Low contention: %v\n", lowContentionTime)
    fmt.Printf("  Ratio: %.2fx slower\n", float64(highContentionTime)/float64(lowContentionTime))
}

func (ms MutexStats) String() string {
    return fmt.Sprintf(`  Name: %s
  Locks: %d
  Blocks: %d
  Block Time: %v
  Contention Rate: %.2f%%`,
        ms.Name,
        ms.LockCount,
        ms.BlockCount,
        time.Duration(ms.BlockTime),
        ms.ContentionRate)
}
```

### Channel Blocking Analysis

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// ChannelBlockingAnalyzer tracks channel operation blocking
type ChannelBlockingAnalyzer struct {
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
}

func NewChannelBlockingAnalyzer() *ChannelBlockingAnalyzer {
    return &ChannelBlockingAnalyzer{
        channels: make(map[string]*ChannelStats),
    }
}

// TrackedChannel wraps a channel with blocking analysis
type TrackedChannel struct {
    ch       chan interface{}
    name     string
    analyzer *ChannelBlockingAnalyzer
    stats    *ChannelStats
}

func (cba *ChannelBlockingAnalyzer) NewTrackedChannel(name string, buffer int) *TrackedChannel {
    stats := &ChannelStats{Name: name}
    
    cba.mu.Lock()
    cba.channels[name] = stats
    cba.mu.Unlock()
    
    return &TrackedChannel{
        ch:       make(chan interface{}, buffer),
        name:     name,
        analyzer: cba,
        stats:    stats,
    }
}

func (tc *TrackedChannel) Send(value interface{}) {
    start := time.Now()
    
    select {
    case tc.ch <- value:
        // Non-blocking send
        atomic.AddInt64(&tc.stats.SendOps, 1)
        return
    default:
        // Would block
        atomic.AddInt64(&tc.stats.SendBlocks, 1)
    }
    
    // Perform blocking send
    tc.ch <- value
    
    duration := time.Since(start)
    atomic.AddInt64(&tc.stats.SendTime, int64(duration))
    atomic.AddInt64(&tc.stats.SendOps, 1)
}

func (tc *TrackedChannel) Receive() interface{} {
    start := time.Now()
    
    select {
    case value := <-tc.ch:
        // Non-blocking receive
        atomic.AddInt64(&tc.stats.RecvOps, 1)
        return value
    default:
        // Would block
        atomic.AddInt64(&tc.stats.RecvBlocks, 1)
    }
    
    // Perform blocking receive
    value := <-tc.ch
    
    duration := time.Since(start)
    atomic.AddInt64(&tc.stats.RecvTime, int64(duration))
    atomic.AddInt64(&tc.stats.RecvOps, 1)
    
    return value
}

func (tc *TrackedChannel) TryReceive() (interface{}, bool) {
    select {
    case value := <-tc.ch:
        atomic.AddInt64(&tc.stats.RecvOps, 1)
        return value, true
    default:
        return nil, false
    }
}

func (tc *TrackedChannel) Close() {
    close(tc.ch)
}

func (tc *TrackedChannel) GetStats() ChannelStats {
    return ChannelStats{
        Name:       tc.stats.Name,
        SendOps:    atomic.LoadInt64(&tc.stats.SendOps),
        RecvOps:    atomic.LoadInt64(&tc.stats.RecvOps),
        SendBlocks: atomic.LoadInt64(&tc.stats.SendBlocks),
        RecvBlocks: atomic.LoadInt64(&tc.stats.RecvBlocks),
        SendTime:   atomic.LoadInt64(&tc.stats.SendTime),
        RecvTime:   atomic.LoadInt64(&tc.stats.RecvTime),
    }
}

func (cs ChannelStats) String() string {
    sendBlockRate := float64(0)
    if cs.SendOps > 0 {
        sendBlockRate = float64(cs.SendBlocks) / float64(cs.SendOps) * 100
    }
    
    recvBlockRate := float64(0)
    if cs.RecvOps > 0 {
        recvBlockRate = float64(cs.RecvBlocks) / float64(cs.RecvOps) * 100
    }
    
    return fmt.Sprintf(`Channel: %s
  Send Operations: %d (%.1f%% blocked)
  Recv Operations: %d (%.1f%% blocked)
  Send Block Time: %v
  Recv Block Time: %v`,
        cs.Name,
        cs.SendOps, sendBlockRate,
        cs.RecvOps, recvBlockRate,
        time.Duration(cs.SendTime),
        time.Duration(cs.RecvTime))
}

func (cba *ChannelBlockingAnalyzer) GetReport() ChannelReport {
    cba.mu.RLock()
    defer cba.mu.RUnlock()
    
    var channelStats []ChannelStats
    totalSendBlocks := int64(0)
    totalRecvBlocks := int64(0)
    totalSendTime := time.Duration(0)
    totalRecvTime := time.Duration(0)
    
    for _, stats := range cba.channels {
        channelStats = append(channelStats, *stats)
        totalSendBlocks += stats.SendBlocks
        totalRecvBlocks += stats.RecvBlocks
        totalSendTime += time.Duration(stats.SendTime)
        totalRecvTime += time.Duration(stats.RecvTime)
    }
    
    return ChannelReport{
        TotalChannels:   len(channelStats),
        TotalSendBlocks: totalSendBlocks,
        TotalRecvBlocks: totalRecvBlocks,
        TotalSendTime:   totalSendTime,
        TotalRecvTime:   totalRecvTime,
        ChannelStats:    channelStats,
    }
}

type ChannelReport struct {
    TotalChannels   int
    TotalSendBlocks int64
    TotalRecvBlocks int64
    TotalSendTime   time.Duration
    TotalRecvTime   time.Duration
    ChannelStats    []ChannelStats
}

func (cr ChannelReport) String() string {
    result := fmt.Sprintf(`Channel Blocking Report:
  Total Channels: %d
  Total Send Blocks: %d
  Total Recv Blocks: %d
  Total Send Block Time: %v
  Total Recv Block Time: %v`,
        cr.TotalChannels,
        cr.TotalSendBlocks,
        cr.TotalRecvBlocks,
        cr.TotalSendTime,
        cr.TotalRecvTime)
    
    if len(cr.ChannelStats) > 0 {
        result += "\n\nPer-Channel Statistics:"
        for _, stats := range cr.ChannelStats {
            result += "\n" + stats.String()
        }
    }
    
    return result
}

func demonstrateChannelBlocking() {
    fmt.Println("\n=== CHANNEL BLOCKING ANALYSIS ===")
    
    analyzer := NewChannelBlockingAnalyzer()
    
    // Test unbuffered channel (high blocking)
    unbuffered := analyzer.NewTrackedChannel("unbuffered", 0)
    
    // Test buffered channel (lower blocking)
    buffered := analyzer.NewTrackedChannel("buffered", 10)
    
    var wg sync.WaitGroup
    
    // Producer-consumer pattern with unbuffered channel
    fmt.Println("Testing unbuffered channel...")
    
    // Slow consumer
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 20; i++ {
            unbuffered.Receive()
            time.Sleep(10 * time.Millisecond) // Slow consumption
        }
    }()
    
    // Fast producer
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 20; i++ {
            unbuffered.Send(i)
            time.Sleep(2 * time.Millisecond) // Fast production
        }
    }()
    
    wg.Wait()
    
    // Producer-consumer pattern with buffered channel
    fmt.Println("Testing buffered channel...")
    
    // Same slow consumer
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 20; i++ {
            buffered.Receive()
            time.Sleep(10 * time.Millisecond)
        }
    }()
    
    // Same fast producer
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 20; i++ {
            buffered.Send(i)
            time.Sleep(2 * time.Millisecond)
        }
    }()
    
    wg.Wait()
    
    // Show results
    fmt.Printf("\nUnbuffered channel stats:\n%s\n", unbuffered.GetStats())
    fmt.Printf("Buffered channel stats:\n%s\n", buffered.GetStats())
    
    report := analyzer.GetReport()
    fmt.Printf("\n%s\n", report)
}
```

### Select Statement Analysis

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)

// SelectAnalyzer tracks select statement blocking patterns
type SelectAnalyzer struct {
    selects map[string]*SelectStats
    mu      sync.RWMutex
}

type SelectStats struct {
    Name         string
    Executions   int64
    CaseHits     map[int]int64
    Timeouts     int64
    Defaults     int64
    BlockTime    int64
    AvgBlockTime time.Duration
}

func NewSelectAnalyzer() *SelectAnalyzer {
    return &SelectAnalyzer{
        selects: make(map[string]*SelectStats),
    }
}

// TrackedSelect provides instrumented select operations
type TrackedSelect struct {
    name     string
    analyzer *SelectAnalyzer
    stats    *SelectStats
}

func (sa *SelectAnalyzer) NewTrackedSelect(name string, numCases int) *TrackedSelect {
    stats := &SelectStats{
        Name:     name,
        CaseHits: make(map[int]int64),
    }
    
    for i := 0; i < numCases; i++ {
        stats.CaseHits[i] = 0
    }
    
    sa.mu.Lock()
    sa.selects[name] = stats
    sa.mu.Unlock()
    
    return &TrackedSelect{
        name:     name,
        analyzer: sa,
        stats:    stats,
    }
}

func (ts *TrackedSelect) ExecuteSelect(selectFunc func() (int, bool)) {
    start := time.Now()
    
    caseIndex, blocked := selectFunc()
    
    duration := time.Since(start)
    atomic.AddInt64(&ts.stats.Executions, 1)
    
    if blocked {
        atomic.AddInt64(&ts.stats.BlockTime, int64(duration))
    }
    
    if caseIndex >= 0 {
        ts.stats.CaseHits[caseIndex]++
    } else if caseIndex == -1 {
        atomic.AddInt64(&ts.stats.Defaults, 1)
    } else if caseIndex == -2 {
        atomic.AddInt64(&ts.stats.Timeouts, 1)
    }
}

func (ts *TrackedSelect) GetStats() SelectStats {
    executions := atomic.LoadInt64(&ts.stats.Executions)
    blockTime := atomic.LoadInt64(&ts.stats.BlockTime)
    
    var avgBlockTime time.Duration
    if executions > 0 {
        avgBlockTime = time.Duration(blockTime / executions)
    }
    
    return SelectStats{
        Name:         ts.stats.Name,
        Executions:   executions,
        CaseHits:     ts.stats.CaseHits,
        Timeouts:     atomic.LoadInt64(&ts.stats.Timeouts),
        Defaults:     atomic.LoadInt64(&ts.stats.Defaults),
        BlockTime:    blockTime,
        AvgBlockTime: avgBlockTime,
    }
}

func (ss SelectStats) String() string {
    result := fmt.Sprintf(`Select: %s
  Executions: %d
  Average Block Time: %v
  Total Block Time: %v
  Timeouts: %d
  Defaults: %d`,
        ss.Name,
        ss.Executions,
        ss.AvgBlockTime,
        time.Duration(ss.BlockTime),
        ss.Timeouts,
        ss.Defaults)
    
    if len(ss.CaseHits) > 0 {
        result += "\n  Case Distribution:"
        for i, hits := range ss.CaseHits {
            percentage := float64(0)
            if ss.Executions > 0 {
                percentage = float64(hits) / float64(ss.Executions) * 100
            }
            result += fmt.Sprintf("\n    Case %d: %d (%.1f%%)", i, hits, percentage)
        }
    }
    
    return result
}

// Complex select scenarios for testing
func demonstrateSelectAnalysis() {
    fmt.Println("\n=== SELECT STATEMENT ANALYSIS ===")
    
    analyzer := NewSelectAnalyzer()
    
    // Test multi-channel select
    multiChannelSelect := analyzer.NewTrackedSelect("multi_channel", 3)
    
    ch1 := make(chan int, 5)
    ch2 := make(chan string, 5)
    ch3 := make(chan bool, 5)
    
    var wg sync.WaitGroup
    
    // Producer for different channels at different rates
    wg.Add(3)
    go func() {
        defer wg.Done()
        for i := 0; i < 50; i++ {
            ch1 <- i
            time.Sleep(10 * time.Millisecond)
        }
        close(ch1)
    }()
    
    go func() {
        defer wg.Done()
        for i := 0; i < 30; i++ {
            ch2 <- fmt.Sprintf("msg_%d", i)
            time.Sleep(15 * time.Millisecond)
        }
        close(ch2)
    }()
    
    go func() {
        defer wg.Done()
        for i := 0; i < 20; i++ {
            ch3 <- i%2 == 0
            time.Sleep(25 * time.Millisecond)
        }
        close(ch3)
    }()
    
    // Consumer using tracked select
    wg.Add(1)
    go func() {
        defer wg.Done()
        
        for {
            multiChannelSelect.ExecuteSelect(func() (int, bool) {
                start := time.Now()
                
                select {
                case val, ok := <-ch1:
                    if !ok {
                        return -1, time.Since(start) > time.Microsecond
                    }
                    _ = val
                    return 0, time.Since(start) > time.Microsecond
                    
                case val, ok := <-ch2:
                    if !ok {
                        return -1, time.Since(start) > time.Microsecond
                    }
                    _ = val
                    return 1, time.Since(start) > time.Microsecond
                    
                case val, ok := <-ch3:
                    if !ok {
                        return -1, time.Since(start) > time.Microsecond
                    }
                    _ = val
                    return 2, time.Since(start) > time.Microsecond
                    
                default:
                    return -1, false // Default case
                }
            })
            
            // Check if all channels are closed
            if isClosed(ch1) && isClosed(ch2) && isClosed(ch3) {
                break
            }
            
            time.Sleep(1 * time.Millisecond)
        }
    }()
    
    wg.Wait()
    
    // Test timeout select
    timeoutSelect := analyzer.NewTrackedSelect("timeout", 2)
    slowCh := make(chan int)
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 10; i++ {
            timeoutSelect.ExecuteSelect(func() (int, bool) {
                start := time.Now()
                
                select {
                case val := <-slowCh:
                    _ = val
                    return 0, time.Since(start) > time.Microsecond
                    
                case <-time.After(50 * time.Millisecond):
                    return -2, true // Timeout case
                }
            })
        }
    }()
    
    // Occasionally send to slow channel
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 3; i++ {
            time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
            select {
            case slowCh <- i:
            default:
            }
        }
        close(slowCh)
    }()
    
    wg.Wait()
    
    // Show results
    fmt.Printf("Multi-channel select stats:\n%s\n", multiChannelSelect.GetStats())
    fmt.Printf("Timeout select stats:\n%s\n", timeoutSelect.GetStats())
}

func isClosed(ch chan int) bool {
    select {
    case <-ch:
        return true
    default:
        return false
    }
}
```

## Optimization Strategies

### Lock-Free Alternatives

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// Compare lock-based vs lock-free implementations
func compareLockVsLockFree() {
    fmt.Println("\n=== LOCK VS LOCK-FREE COMPARISON ===")
    
    iterations := 1000000
    numGoroutines := 10
    
    // Lock-based counter
    var lockCounter int64
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    fmt.Println("Testing lock-based counter...")
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < iterations/numGoroutines; j++ {
                mu.Lock()
                lockCounter++
                mu.Unlock()
            }
        }()
    }
    
    wg.Wait()
    lockTime := time.Since(start)
    
    // Lock-free counter
    var atomicCounter int64
    
    fmt.Println("Testing lock-free counter...")
    start = time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < iterations/numGoroutines; j++ {
                atomic.AddInt64(&atomicCounter, 1)
            }
        }()
    }
    
    wg.Wait()
    atomicTime := time.Since(start)
    
    fmt.Printf("Results:\n")
    fmt.Printf("  Lock-based: %d operations in %v\n", lockCounter, lockTime)
    fmt.Printf("  Lock-free: %d operations in %v\n", atomicCounter, atomicTime)
    fmt.Printf("  Speedup: %.2fx\n", float64(lockTime)/float64(atomicTime))
}
```

### Channel Buffer Optimization

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func optimizeChannelBuffers() {
    fmt.Println("\n=== CHANNEL BUFFER OPTIMIZATION ===")
    
    testSizes := []int{0, 1, 10, 100, 1000}
    numMessages := 10000
    
    for _, bufferSize := range testSizes {
        fmt.Printf("Testing buffer size: %d\n", bufferSize)
        
        ch := make(chan int, bufferSize)
        var wg sync.WaitGroup
        
        start := time.Now()
        
        // Producer
        wg.Add(1)
        go func() {
            defer wg.Done()
            defer close(ch)
            for i := 0; i < numMessages; i++ {
                ch <- i
            }
        }()
        
        // Consumer
        wg.Add(1)
        go func() {
            defer wg.Done()
            for range ch {
                // Simulate processing
                time.Sleep(time.Microsecond)
            }
        }()
        
        wg.Wait()
        duration := time.Since(start)
        
        fmt.Printf("  Duration: %v\n", duration)
    }
}
```

## Best Practices for Block Profiling

### 1. Enable Profiling Strategically

```go
func enableBlockProfiling() {
    // Set appropriate sampling rate
    runtime.SetBlockProfileRate(1) // Sample every blocking event
    
    // Or sample less frequently for production
    // runtime.SetBlockProfileRate(1000) // Sample every 1µs of blocking
}
```

### 2. Analyze Profile Data

```bash
# Collect block profile
go tool pprof http://localhost:6060/debug/pprof/block

# Analyze in web interface
go tool pprof -http=:8080 block.prof

# Find top blocking operations
(pprof) top10

# Show blocking call tree
(pprof) tree

# Focus on specific functions
(pprof) focus sync.Mutex
```

### 3. Optimize Based on Results

```go
// Replace highly contended mutexes with:
// 1. Lock-free operations
var counter int64
atomic.AddInt64(&counter, 1)

// 2. Sharded locks
type ShardedCounter struct {
    shards []struct {
        mu    sync.Mutex
        count int64
    }
}

// 3. Channels for coordination
func coordinateWithChannels() {
    requests := make(chan Request, 100)
    responses := make(chan Response, 100)
    
    // Process requests without locks
    go func() {
        for req := range requests {
            resp := process(req)
            responses <- resp
        }
    }()
}
```

## Next Steps

- Learn [Trace Analysis](trace-analysis.md) techniques
- Study [Concurrency Optimization](../optimization/concurrency/README.md)
- Explore [Performance Benchmarking](../benchmarking/README.md)
- Master [Production Monitoring](../monitoring/README.md)

## Summary

Block profiling in Go helps identify:

1. **Mutex contention** - High lock contention points
2. **Channel bottlenecks** - Blocking channel operations
3. **Select inefficiencies** - Suboptimal select patterns
4. **Synchronization overhead** - Cost of coordination
5. **Optimization opportunities** - Lock-free alternatives

Use block profiling to build more efficient concurrent applications with reduced synchronization overhead.
