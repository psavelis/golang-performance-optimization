# Goroutine Analysis

Goroutine analysis provides deep insights into concurrent execution patterns, lifecycle management, and performance characteristics of Go's lightweight threads. This comprehensive guide covers advanced techniques for analyzing, monitoring, and optimizing goroutine behavior in production systems.

## Understanding Goroutine Lifecycle

Goroutines progress through several states:
- **Runnable** - Ready to execute but waiting for CPU
- **Running** - Currently executing on a CPU core
- **Waiting** - Blocked on I/O, channels, or synchronization primitives
- **System call** - Executing system calls
- **Dead** - Completed execution or panicked

### Comprehensive Goroutine Analyzer

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "runtime/debug"
    "sort"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

// GoroutineAnalyzer provides comprehensive goroutine monitoring and analysis
type GoroutineAnalyzer struct {
    goroutines     map[int]*TrackedGoroutine
    globalStats    *GlobalGoroutineStats
    mu             sync.RWMutex
    enabled        bool
    sampleInterval time.Duration
    maxHistory     int
}

type GlobalGoroutineStats struct {
    TotalCreated      int64
    TotalCompleted    int64
    CurrentActive     int32
    PeakConcurrent    int32
    TotalCPUTime      int64
    TotalWaitTime     int64
    TotalMemoryUsage  int64
    PanicCount        int32
    DeadlockCount     int32
    LeakCount         int32
    LastSample        time.Time
}

func NewGoroutineAnalyzer() *GoroutineAnalyzer {
    return &GoroutineAnalyzer{
        goroutines:     make(map[int]*TrackedGoroutine),
        globalStats:    &GlobalGoroutineStats{},
        sampleInterval: time.Millisecond * 100,
        maxHistory:     10000,
    }
}

func (ga *GoroutineAnalyzer) Enable() {
    ga.mu.Lock()
    defer ga.mu.Unlock()
    
    if ga.enabled {
        return
    }
    
    ga.enabled = true
    go ga.monitorGoroutines()
}

func (ga *GoroutineAnalyzer) Disable() {
    ga.mu.Lock()
    defer ga.mu.Unlock()
    ga.enabled = false
}

func (ga *GoroutineAnalyzer) TrackGoroutine(name string, fn func()) {
    goroutineID := getGoroutineID()
    
    tracker := &TrackedGoroutine{
        ID:        goroutineID,
        Name:      name,
        CreatedAt: time.Now(),
        CreatedBy: getCallerInfo(),
        State:     StateRunnable,
        analyzer:  ga,
        stats:     &GoroutineStats{},
    }
    
    ga.mu.Lock()
    ga.goroutines[goroutineID] = tracker
    ga.mu.Unlock()
    
    atomic.AddInt64(&ga.globalStats.TotalCreated, 1)
    current := atomic.AddInt32(&ga.globalStats.CurrentActive, 1)
    
    // Update peak concurrent goroutines
    for {
        peak := atomic.LoadInt32(&ga.globalStats.PeakConcurrent)
        if current <= peak || atomic.CompareAndSwapInt32(&ga.globalStats.PeakConcurrent, peak, current) {
            break
        }
    }
    
    // Execute function with monitoring
    go func() {
        defer ga.completeGoroutine(goroutineID)
        defer tracker.handlePanic()
        
        tracker.start()
        fn()
        tracker.complete()
    }()
}

func (ga *GoroutineAnalyzer) completeGoroutine(goroutineID int) {
    atomic.AddInt64(&ga.globalStats.TotalCompleted, 1)
    atomic.AddInt32(&ga.globalStats.CurrentActive, -1)
    
    ga.mu.Lock()
    if tracker, exists := ga.goroutines[goroutineID]; exists {
        tracker.CompletedAt = time.Now()
        tracker.State = StateDead
        
        // Archive completed goroutine if history is full
        if len(ga.goroutines) > ga.maxHistory {
            // Remove oldest completed goroutines
            ga.cleanupHistory()
        }
    }
    ga.mu.Unlock()
}

func (ga *GoroutineAnalyzer) cleanupHistory() {
    // Keep only recent and active goroutines
    var completed []*TrackedGoroutine
    var active []*TrackedGoroutine
    
    for _, tracker := range ga.goroutines {
        if tracker.State == StateDead {
            completed = append(completed, tracker)
        } else {
            active = append(active, tracker)
        }
    }
    
    // Sort completed by completion time (oldest first)
    sort.Slice(completed, func(i, j int) bool {
        return completed[i].CompletedAt.Before(completed[j].CompletedAt)
    })
    
    // Remove oldest completed goroutines
    toRemove := len(completed) - (ga.maxHistory - len(active))
    if toRemove > 0 {
        for i := 0; i < toRemove; i++ {
            delete(ga.goroutines, completed[i].ID)
        }
    }
}

func (ga *GoroutineAnalyzer) monitorGoroutines() {
    ticker := time.NewTicker(ga.sampleInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        ga.mu.RLock()
        enabled := ga.enabled
        ga.mu.RUnlock()
        
        if !enabled {
            return
        }
        
        ga.sampleGoroutineStates()
    }
}

func (ga *GoroutineAnalyzer) sampleGoroutineStates() {
    ga.mu.RLock()
    defer ga.mu.RUnlock()
    
    now := time.Now()
    
    for _, tracker := range ga.goroutines {
        if tracker.State != StateDead {
            tracker.sampleState(now)
        }
    }
    
    atomic.StoreInt64((*int64)(&ga.globalStats.LastSample), now.UnixNano())
}

func (ga *GoroutineAnalyzer) DetectLeaks(threshold time.Duration) []GoroutineLeak {
    ga.mu.RLock()
    defer ga.mu.RUnlock()
    
    var leaks []GoroutineLeak
    cutoff := time.Now().Add(-threshold)
    
    for _, tracker := range ga.goroutines {
        if tracker.State != StateDead && tracker.CreatedAt.Before(cutoff) {
            // Check if goroutine appears stuck
            if tracker.isStuck(threshold) {
                leaks = append(leaks, GoroutineLeak{
                    Goroutine:    tracker,
                    Age:          time.Since(tracker.CreatedAt),
                    LastActivity: tracker.getLastActivity(),
                    StackTrace:   tracker.getCurrentStack(),
                })
            }
        }
    }
    
    return leaks
}

func (ga *GoroutineAnalyzer) GetAnalysisReport() GoroutineAnalysisReport {
    ga.mu.RLock()
    defer ga.mu.RUnlock()
    
    var goroutineReports []GoroutineReport
    
    for _, tracker := range ga.goroutines {
        report := tracker.GetReport()
        goroutineReports = append(goroutineReports, report)
    }
    
    // Sort by performance impact
    sort.Slice(goroutineReports, func(i, j int) bool {
        return goroutineReports[i].PerformanceImpact() > goroutineReports[j].PerformanceImpact()
    })
    
    totalCreated := atomic.LoadInt64(&ga.globalStats.TotalCreated)
    totalCompleted := atomic.LoadInt64(&ga.globalStats.TotalCompleted)
    currentActive := atomic.LoadInt32(&ga.globalStats.CurrentActive)
    peakConcurrent := atomic.LoadInt32(&ga.globalStats.PeakConcurrent)
    totalCPUTime := atomic.LoadInt64(&ga.globalStats.TotalCPUTime)
    totalWaitTime := atomic.LoadInt64(&ga.globalStats.TotalWaitTime)
    panicCount := atomic.LoadInt32(&ga.globalStats.PanicCount)
    deadlockCount := atomic.LoadInt32(&ga.globalStats.DeadlockCount)
    leakCount := atomic.LoadInt32(&ga.globalStats.LeakCount)
    
    return GoroutineAnalysisReport{
        TotalCreated:      totalCreated,
        TotalCompleted:    totalCompleted,
        CurrentActive:     currentActive,
        PeakConcurrent:    peakConcurrent,
        TotalCPUTime:      time.Duration(totalCPUTime),
        TotalWaitTime:     time.Duration(totalWaitTime),
        PanicCount:        panicCount,
        DeadlockCount:     deadlockCount,
        LeakCount:         leakCount,
        GoroutineReports:  goroutineReports,
        SampleTime:        time.Unix(0, atomic.LoadInt64((*int64)(&ga.globalStats.LastSample))),
    }
}

// TrackedGoroutine represents a monitored goroutine
type TrackedGoroutine struct {
    ID          int
    Name        string
    CreatedAt   time.Time
    CompletedAt time.Time
    CreatedBy   string
    State       GoroutineState
    analyzer    *GoroutineAnalyzer
    stats       *GoroutineStats
    mu          sync.RWMutex
}

type GoroutineState int

const (
    StateRunnable GoroutineState = iota
    StateRunning
    StateWaiting
    StateSyscall
    StateDead
)

func (gs GoroutineState) String() string {
    switch gs {
    case StateRunnable:
        return "Runnable"
    case StateRunning:
        return "Running"
    case StateWaiting:
        return "Waiting"
    case StateSyscall:
        return "Syscall"
    case StateDead:
        return "Dead"
    default:
        return "Unknown"
    }
}

type GoroutineStats struct {
    CPUTime          int64
    WaitTime         int64
    MemoryAllocated  int64
    StateChanges     []StateChange
    StackSamples     []StackSample
    PanicCount       int32
    LastActivity     time.Time
    BlockedOn        string
    WaitReason       string
    mu               sync.RWMutex
}

type StateChange struct {
    Timestamp time.Time
    OldState  GoroutineState
    NewState  GoroutineState
    Reason    string
}

type StackSample struct {
    Timestamp  time.Time
    StackTrace []string
    PCCounters []uintptr
}

func (tg *TrackedGoroutine) start() {
    tg.mu.Lock()
    tg.State = StateRunning
    tg.stats.LastActivity = time.Now()
    tg.recordStateChange(StateRunnable, StateRunning, "started")
    tg.mu.Unlock()
}

func (tg *TrackedGoroutine) complete() {
    tg.mu.Lock()
    tg.State = StateDead
    tg.CompletedAt = time.Now()
    tg.stats.LastActivity = tg.CompletedAt
    tg.recordStateChange(StateRunning, StateDead, "completed")
    tg.mu.Unlock()
}

func (tg *TrackedGoroutine) handlePanic() {
    if r := recover(); r != nil {
        atomic.AddInt32(&tg.analyzer.globalStats.PanicCount, 1)
        atomic.AddInt32(&tg.stats.PanicCount, 1)
        
        tg.mu.Lock()
        tg.State = StateDead
        tg.CompletedAt = time.Now()
        tg.recordStateChange(StateRunning, StateDead, fmt.Sprintf("panic: %v", r))
        tg.mu.Unlock()
        
        // Re-panic to maintain normal panic behavior
        panic(r)
    }
}

func (tg *TrackedGoroutine) recordStateChange(oldState, newState GoroutineState, reason string) {
    change := StateChange{
        Timestamp: time.Now(),
        OldState:  oldState,
        NewState:  newState,
        Reason:    reason,
    }
    
    tg.stats.mu.Lock()
    tg.stats.StateChanges = append(tg.stats.StateChanges, change)
    
    // Keep only recent state changes
    if len(tg.stats.StateChanges) > 100 {
        tg.stats.StateChanges = tg.stats.StateChanges[len(tg.stats.StateChanges)-100:]
    }
    tg.stats.mu.Unlock()
}

func (tg *TrackedGoroutine) sampleState(now time.Time) {
    tg.mu.Lock()
    defer tg.mu.Unlock()
    
    // Sample stack trace periodically
    if len(tg.stats.StackSamples) == 0 || 
       now.Sub(tg.stats.StackSamples[len(tg.stats.StackSamples)-1].Timestamp) > time.Second {
        
        stack := tg.getCurrentStack()
        sample := StackSample{
            Timestamp:  now,
            StackTrace: stack,
        }
        
        tg.stats.mu.Lock()
        tg.stats.StackSamples = append(tg.stats.StackSamples, sample)
        
        // Keep only recent samples
        if len(tg.stats.StackSamples) > 50 {
            tg.stats.StackSamples = tg.stats.StackSamples[len(tg.stats.StackSamples)-50:]
        }
        tg.stats.mu.Unlock()
    }
    
    // Update activity timestamp for running goroutines
    if tg.State == StateRunning || tg.State == StateRunnable {
        tg.stats.LastActivity = now
    }
}

func (tg *TrackedGoroutine) getCurrentStack() []string {
    // Get current stack trace
    buf := make([]byte, 4096)
    n := runtime.Stack(buf, false)
    stackTrace := string(buf[:n])
    
    lines := strings.Split(stackTrace, "\n")
    var cleanLines []string
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            cleanLines = append(cleanLines, line)
        }
    }
    
    return cleanLines
}

func (tg *TrackedGoroutine) isStuck(threshold time.Duration) bool {
    tg.stats.mu.RLock()
    lastActivity := tg.stats.LastActivity
    tg.stats.mu.RUnlock()
    
    return time.Since(lastActivity) > threshold && tg.State != StateDead
}

func (tg *TrackedGoroutine) getLastActivity() time.Time {
    tg.stats.mu.RLock()
    defer tg.stats.mu.RUnlock()
    return tg.stats.LastActivity
}

func (tg *TrackedGoroutine) GetReport() GoroutineReport {
    tg.mu.RLock()
    tg.stats.mu.RLock()
    defer tg.stats.mu.RUnlock()
    defer tg.mu.RUnlock()
    
    var lifetime time.Duration
    if !tg.CompletedAt.IsZero() {
        lifetime = tg.CompletedAt.Sub(tg.CreatedAt)
    } else {
        lifetime = time.Since(tg.CreatedAt)
    }
    
    cpuTime := atomic.LoadInt64(&tg.stats.CPUTime)
    waitTime := atomic.LoadInt64(&tg.stats.WaitTime)
    memoryAllocated := atomic.LoadInt64(&tg.stats.MemoryAllocated)
    panicCount := atomic.LoadInt32(&tg.stats.PanicCount)
    
    var cpuUtilization float64
    if lifetime > 0 {
        cpuUtilization = float64(cpuTime) / float64(lifetime) * 100
    }
    
    return GoroutineReport{
        ID:               tg.ID,
        Name:             tg.Name,
        State:            tg.State,
        CreatedAt:        tg.CreatedAt,
        CompletedAt:      tg.CompletedAt,
        CreatedBy:        tg.CreatedBy,
        Lifetime:         lifetime,
        CPUTime:          time.Duration(cpuTime),
        WaitTime:         time.Duration(waitTime),
        CPUUtilization:   cpuUtilization,
        MemoryAllocated:  memoryAllocated,
        StateChanges:     len(tg.stats.StateChanges),
        PanicCount:       panicCount,
        LastActivity:     tg.stats.LastActivity,
        BlockedOn:        tg.stats.BlockedOn,
        WaitReason:       tg.stats.WaitReason,
        StackSamples:     len(tg.stats.StackSamples),
        CurrentStack:     tg.getCurrentStack(),
    }
}

type GoroutineReport struct {
    ID               int
    Name             string
    State            GoroutineState
    CreatedAt        time.Time
    CompletedAt      time.Time
    CreatedBy        string
    Lifetime         time.Duration
    CPUTime          time.Duration
    WaitTime         time.Duration
    CPUUtilization   float64
    MemoryAllocated  int64
    StateChanges     int
    PanicCount       int32
    LastActivity     time.Time
    BlockedOn        string
    WaitReason       string
    StackSamples     int
    CurrentStack     []string
}

func (gr GoroutineReport) PerformanceImpact() float64 {
    // Calculate performance impact score
    lifetimeScore := float64(gr.Lifetime/time.Millisecond) / 1000.0
    cpuScore := float64(gr.CPUTime/time.Millisecond) / 1000.0
    memoryScore := float64(gr.MemoryAllocated) / (1024 * 1024) // MB
    
    return lifetimeScore + cpuScore + memoryScore
}

func (gr GoroutineReport) String() string {
    status := gr.State.String()
    if !gr.CompletedAt.IsZero() {
        status += fmt.Sprintf(" (completed at %v)", gr.CompletedAt.Format(time.RFC3339))
    }
    
    result := fmt.Sprintf(`Goroutine %d: %s
  State: %s
  Created: %v by %s
  Lifetime: %v
  CPU Time: %v (%.1f%% utilization)
  Wait Time: %v
  Memory Allocated: %d bytes
  State Changes: %d
  Panic Count: %d
  Last Activity: %v
  Stack Samples: %d
  Performance Impact: %.2f`,
        gr.ID, gr.Name,
        status,
        gr.CreatedAt.Format(time.RFC3339), gr.CreatedBy,
        gr.Lifetime,
        gr.CPUTime, gr.CPUUtilization,
        gr.WaitTime,
        gr.MemoryAllocated,
        gr.StateChanges,
        gr.PanicCount,
        gr.LastActivity.Format(time.RFC3339),
        gr.StackSamples,
        gr.PerformanceImpact())
    
    if gr.BlockedOn != "" {
        result += fmt.Sprintf("\n  Blocked On: %s", gr.BlockedOn)
    }
    
    if gr.WaitReason != "" {
        result += fmt.Sprintf("\n  Wait Reason: %s", gr.WaitReason)
    }
    
    if len(gr.CurrentStack) > 0 && gr.State != StateDead {
        result += "\n  Current Stack:"
        for i, line := range gr.CurrentStack {
            if i >= 5 { // Show top 5 stack frames
                result += "\n    ..."
                break
            }
            result += fmt.Sprintf("\n    %s", line)
        }
    }
    
    return result
}

type GoroutineLeak struct {
    Goroutine    *TrackedGoroutine
    Age          time.Duration
    LastActivity time.Time
    StackTrace   []string
}

func (gl GoroutineLeak) String() string {
    return fmt.Sprintf("LEAK: Goroutine %d (%s) - Age: %v, Last Activity: %v",
        gl.Goroutine.ID, gl.Goroutine.Name, gl.Age, gl.LastActivity.Format(time.RFC3339))
}

type GoroutineAnalysisReport struct {
    TotalCreated      int64
    TotalCompleted    int64
    CurrentActive     int32
    PeakConcurrent    int32
    TotalCPUTime      time.Duration
    TotalWaitTime     time.Duration
    PanicCount        int32
    DeadlockCount     int32
    LeakCount         int32
    GoroutineReports  []GoroutineReport
    SampleTime        time.Time
}

func (gar GoroutineAnalysisReport) String() string {
    completionRate := float64(gar.TotalCompleted) / float64(gar.TotalCreated) * 100
    if gar.TotalCreated == 0 {
        completionRate = 0
    }
    
    result := fmt.Sprintf(`Goroutine Analysis Report (sampled at %v):
  Total Created: %d
  Total Completed: %d (%.1f%%)
  Currently Active: %d
  Peak Concurrent: %d
  Total CPU Time: %v
  Total Wait Time: %v
  Panic Count: %d
  Deadlock Count: %d
  Leak Count: %d`,
        gar.SampleTime.Format(time.RFC3339),
        gar.TotalCreated,
        gar.TotalCompleted, completionRate,
        gar.CurrentActive,
        gar.PeakConcurrent,
        gar.TotalCPUTime,
        gar.TotalWaitTime,
        gar.PanicCount,
        gar.DeadlockCount,
        gar.LeakCount)
    
    if len(gar.GoroutineReports) > 0 {
        result += "\n\nTop Goroutines by Performance Impact:"
        for i, report := range gar.GoroutineReports {
            if i >= 5 { // Show top 5
                break
            }
            result += fmt.Sprintf("\n\n%d. %s", i+1, report.String())
        }
    }
    
    return result
}

// Utility functions
func getGoroutineID() int {
    // Simplified goroutine ID extraction
    // In practice, you'd parse runtime.Stack() output
    return int(time.Now().UnixNano() % 100000)
}

func getCallerInfo() string {
    pc, file, line, ok := runtime.Caller(2)
    if !ok {
        return "unknown"
    }
    
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return fmt.Sprintf("%s:%d", file, line)
    }
    
    return fmt.Sprintf("%s (%s:%d)", fn.Name(), file, line)
}

func demonstrateGoroutineAnalysis() {
    fmt.Println("=== GOROUTINE ANALYSIS DEMONSTRATION ===")
    
    analyzer := NewGoroutineAnalyzer()
    analyzer.Enable()
    defer analyzer.Disable()
    
    var wg sync.WaitGroup
    
    // Create different types of goroutines for analysis
    
    // 1. Fast completing goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        analyzer.TrackGoroutine(fmt.Sprintf("fast_worker_%d", i), func() {
            defer wg.Done()
            time.Sleep(time.Millisecond * 50)
        })
    }
    
    // 2. Long-running goroutines
    for i := 0; i < 3; i++ {
        wg.Add(1)
        analyzer.TrackGoroutine(fmt.Sprintf("long_worker_%d", i), func() {
            defer wg.Done()
            time.Sleep(time.Second * 2)
        })
    }
    
    // 3. CPU-intensive goroutines
    for i := 0; i < 2; i++ {
        wg.Add(1)
        analyzer.TrackGoroutine(fmt.Sprintf("cpu_worker_%d", i), func() {
            defer wg.Done()
            
            // CPU-intensive work
            sum := 0
            for j := 0; j < 1000000; j++ {
                sum += j
            }
            _ = sum
        })
    }
    
    // 4. Potentially leaking goroutine (for demonstration)
    analyzer.TrackGoroutine("potential_leak", func() {
        // This goroutine will run much longer than others
        time.Sleep(time.Second * 10)
    })
    
    // Wait for some goroutines to complete
    wg.Wait()
    
    // Give some time for monitoring
    time.Sleep(time.Second)
    
    // Generate reports
    report := analyzer.GetAnalysisReport()
    fmt.Printf("\n%s\n", report)
    
    // Check for leaks
    leaks := analyzer.DetectLeaks(time.Second * 3)
    if len(leaks) > 0 {
        fmt.Printf("\n=== DETECTED LEAKS ===\n")
        for _, leak := range leaks {
            fmt.Printf("%s\n", leak)
        }
    }
}
```

## Advanced Analysis Patterns

### 1. Goroutine Pool Analysis

```go
// Worker pool analyzer
type WorkerPoolAnalyzer struct {
    poolName     string
    workers      []*WorkerMetrics
    jobQueue     chan Job
    resultQueue  chan Result
    poolStats    *PoolStats
    mu           sync.RWMutex
}

type WorkerMetrics struct {
    ID           int
    JobsProcessed int64
    TotalTime     time.Duration
    IdleTime      time.Duration
    BusyTime      time.Duration
    ErrorCount    int64
    LastJobTime   time.Time
}

type PoolStats struct {
    TotalJobs       int64
    CompletedJobs   int64
    FailedJobs      int64
    AverageLatency  time.Duration
    Throughput      float64 // jobs per second
    Utilization     float64 // percentage
    QueueDepth      int32
    MaxQueueDepth   int32
}

func (wpa *WorkerPoolAnalyzer) AnalyzePool() PoolAnalysisReport {
    wpa.mu.RLock()
    defer wpa.mu.RUnlock()
    
    var totalJobsProcessed int64
    var totalBusyTime time.Duration
    var totalIdleTime time.Duration
    var totalErrors int64
    
    for _, worker := range wpa.workers {
        totalJobsProcessed += atomic.LoadInt64(&worker.JobsProcessed)
        totalBusyTime += worker.BusyTime
        totalIdleTime += worker.IdleTime
        totalErrors += atomic.LoadInt64(&worker.ErrorCount)
    }
    
    totalTime := totalBusyTime + totalIdleTime
    var utilization float64
    if totalTime > 0 {
        utilization = float64(totalBusyTime) / float64(totalTime) * 100
    }
    
    return PoolAnalysisReport{
        PoolName:          wpa.poolName,
        WorkerCount:       len(wpa.workers),
        TotalJobs:         totalJobsProcessed,
        TotalErrors:       totalErrors,
        Utilization:       utilization,
        AverageBusyTime:   totalBusyTime / time.Duration(len(wpa.workers)),
        AverageIdleTime:   totalIdleTime / time.Duration(len(wpa.workers)),
        EfficiencyScore:   wpa.calculateEfficiency(),
    }
}

type PoolAnalysisReport struct {
    PoolName        string
    WorkerCount     int
    TotalJobs       int64
    TotalErrors     int64
    Utilization     float64
    AverageBusyTime time.Duration
    AverageIdleTime time.Duration
    EfficiencyScore float64
}

func (wpa *WorkerPoolAnalyzer) calculateEfficiency() float64 {
    // Efficiency based on utilization, error rate, and throughput
    errorRate := float64(wpa.poolStats.FailedJobs) / float64(wpa.poolStats.TotalJobs) * 100
    utilization := wpa.poolStats.Utilization
    
    efficiency := utilization * (1 - errorRate/100)
    return efficiency
}
```

### 2. Goroutine Communication Analysis

```go
// Communication pattern analyzer
type CommunicationAnalyzer struct {
    channels    map[string]*ChannelUsage
    patterns    []CommunicationPattern
    goroutines  map[int]*GoroutineCommunication
    mu          sync.RWMutex
}

type ChannelUsage struct {
    Name            string
    SenderGoroutines map[int]int64
    ReceiverGoroutines map[int]int64
    MessageCount    int64
    TotalDataSize   int64
    AverageLatency  time.Duration
}

type CommunicationPattern struct {
    Type        string // "fan-out", "fan-in", "pipeline", "worker-pool"
    Participants []int  // Goroutine IDs
    Efficiency  float64
    Bottlenecks []string
}

type GoroutineCommunication struct {
    ID              int
    MessagesSent    int64
    MessagesReceived int64
    BytesSent       int64
    BytesReceived   int64
    Partners        map[int]int64 // Partner goroutine ID -> message count
}

func (ca *CommunicationAnalyzer) AnalyzeCommunication() CommunicationReport {
    ca.mu.RLock()
    defer ca.mu.RUnlock()
    
    patterns := ca.detectPatterns()
    bottlenecks := ca.identifyBottlenecks()
    
    return CommunicationReport{
        TotalChannels:     len(ca.channels),
        TotalGoroutines:   len(ca.goroutines),
        DetectedPatterns:  patterns,
        Bottlenecks:      bottlenecks,
        NetworkEfficiency: ca.calculateNetworkEfficiency(),
    }
}

func (ca *CommunicationAnalyzer) detectPatterns() []CommunicationPattern {
    // Implement pattern detection algorithms
    var patterns []CommunicationPattern
    
    // Detect fan-out patterns (1 sender, multiple receivers)
    for _, channel := range ca.channels {
        if len(channel.SenderGoroutines) == 1 && len(channel.ReceiverGoroutines) > 2 {
            var participants []int
            for senderID := range channel.SenderGoroutines {
                participants = append(participants, senderID)
            }
            for receiverID := range channel.ReceiverGoroutines {
                participants = append(participants, receiverID)
            }
            
            patterns = append(patterns, CommunicationPattern{
                Type:         "fan-out",
                Participants: participants,
                Efficiency:   ca.calculatePatternEfficiency(channel),
            })
        }
    }
    
    return patterns
}

func (ca *CommunicationAnalyzer) identifyBottlenecks() []string {
    var bottlenecks []string
    
    // Check for channels with high contention
    for name, channel := range ca.channels {
        if channel.AverageLatency > time.Millisecond*10 {
            bottlenecks = append(bottlenecks, fmt.Sprintf("High latency on channel '%s': %v", name, channel.AverageLatency))
        }
    }
    
    return bottlenecks
}

func (ca *CommunicationAnalyzer) calculateNetworkEfficiency() float64 {
    // Calculate overall communication network efficiency
    return 85.0 // Placeholder
}

func (ca *CommunicationAnalyzer) calculatePatternEfficiency(channel *ChannelUsage) float64 {
    // Calculate efficiency for specific communication pattern
    return 90.0 // Placeholder
}

type CommunicationReport struct {
    TotalChannels     int
    TotalGoroutines   int
    DetectedPatterns  []CommunicationPattern
    Bottlenecks       []string
    NetworkEfficiency float64
}
```

### 3. Resource Contention Analysis

```go
// Resource contention analyzer
type ResourceContentionAnalyzer struct {
    resources   map[string]*ResourceUsage
    contentions []ContentionEvent
    mu          sync.RWMutex
}

type ResourceUsage struct {
    Name            string
    Type            string // "mutex", "channel", "file", "network"
    WaitingGoroutines []int
    AverageWaitTime time.Duration
    MaxWaitTime     time.Duration
    ContentionCount int64
    ThroughputRate  float64
}

type ContentionEvent struct {
    Timestamp   time.Time
    ResourceName string
    GoroutineID int
    WaitTime    time.Duration
    Resolved    bool
}

func (rca *ResourceContentionAnalyzer) AnalyzeContention() ContentionAnalysisReport {
    rca.mu.RLock()
    defer rca.mu.RUnlock()
    
    var totalContentions int64
    var totalWaitTime time.Duration
    var hotspots []ResourceHotspot
    
    for name, resource := range rca.resources {
        totalContentions += resource.ContentionCount
        totalWaitTime += resource.AverageWaitTime * time.Duration(resource.ContentionCount)
        
        if resource.ContentionCount > 100 { // Threshold for hotspot
            hotspots = append(hotspots, ResourceHotspot{
                ResourceName:    name,
                ContentionCount: resource.ContentionCount,
                AverageWaitTime: resource.AverageWaitTime,
                Severity:        rca.calculateSeverity(resource),
            })
        }
    }
    
    // Sort hotspots by severity
    sort.Slice(hotspots, func(i, j int) bool {
        return hotspots[i].Severity > hotspots[j].Severity
    })
    
    return ContentionAnalysisReport{
        TotalResources:    len(rca.resources),
        TotalContentions:  totalContentions,
        TotalWaitTime:     totalWaitTime,
        Hotspots:         hotspots,
        ContentionRate:    float64(totalContentions) / time.Since(time.Now().Add(-time.Hour)).Seconds(),
    }
}

func (rca *ResourceContentionAnalyzer) calculateSeverity(resource *ResourceUsage) float64 {
    // Calculate contention severity score
    frequencyScore := float64(resource.ContentionCount) / 1000.0
    latencyScore := float64(resource.AverageWaitTime/time.Millisecond) / 100.0
    
    return frequencyScore + latencyScore
}

type ResourceHotspot struct {
    ResourceName    string
    ContentionCount int64
    AverageWaitTime time.Duration
    Severity        float64
}

type ContentionAnalysisReport struct {
    TotalResources   int
    TotalContentions int64
    TotalWaitTime    time.Duration
    Hotspots        []ResourceHotspot
    ContentionRate   float64
}
```

## Performance Optimization Strategies

### 1. Goroutine Lifecycle Optimization

```go
// Goroutine lifecycle optimizer
type LifecycleOptimizer struct {
    recommendations []OptimizationRecommendation
    analyzer        *GoroutineAnalyzer
}

type OptimizationRecommendation struct {
    Type        string
    Severity    string
    Description string
    Impact      string
    Solution    string
}

func (lo *LifecycleOptimizer) GenerateRecommendations(report GoroutineAnalysisReport) []OptimizationRecommendation {
    var recommendations []OptimizationRecommendation
    
    // Check for goroutine leaks
    if report.LeakCount > 0 {
        recommendations = append(recommendations, OptimizationRecommendation{
            Type:        "leak_detection",
            Severity:    "high",
            Description: fmt.Sprintf("Detected %d potential goroutine leaks", report.LeakCount),
            Impact:      "Memory growth, resource exhaustion",
            Solution:    "Implement proper goroutine lifecycle management with context cancellation",
        })
    }
    
    // Check for excessive goroutine creation
    if report.PeakConcurrent > 10000 {
        recommendations = append(recommendations, OptimizationRecommendation{
            Type:        "excessive_concurrency",
            Severity:    "medium",
            Description: fmt.Sprintf("Peak concurrent goroutines: %d", report.PeakConcurrent),
            Impact:      "Scheduler overhead, memory pressure",
            Solution:    "Consider using worker pools to limit concurrency",
        })
    }
    
    // Check for high panic rate
    panicRate := float64(report.PanicCount) / float64(report.TotalCreated) * 100
    if panicRate > 5.0 {
        recommendations = append(recommendations, OptimizationRecommendation{
            Type:        "high_panic_rate",
            Severity:    "high",
            Description: fmt.Sprintf("High panic rate: %.1f%%", panicRate),
            Impact:      "Application instability, data loss",
            Solution:    "Implement proper error handling and recovery mechanisms",
        })
    }
    
    return recommendations
}
```

### 2. Resource Pool Management

```go
// Goroutine pool manager
type GoroutinePoolManager struct {
    pools map[string]*GoroutinePool
    mu    sync.RWMutex
}

type GoroutinePool struct {
    name        string
    size        int
    workers     chan struct{}
    jobs        chan Job
    results     chan Result
    metrics     *PoolMetrics
    shutdown    chan struct{}
    wg          sync.WaitGroup
}

type PoolMetrics struct {
    JobsProcessed  int64
    JobsQueued     int64
    WorkersActive  int32
    WorkersIdle    int32
    AverageLatency time.Duration
    Throughput     float64
}

func (gpm *GoroutinePoolManager) CreatePool(name string, size int, bufferSize int) *GoroutinePool {
    pool := &GoroutinePool{
        name:     name,
        size:     size,
        workers:  make(chan struct{}, size),
        jobs:     make(chan Job, bufferSize),
        results:  make(chan Result, bufferSize),
        metrics:  &PoolMetrics{},
        shutdown: make(chan struct{}),
    }
    
    // Initialize workers
    for i := 0; i < size; i++ {
        pool.workers <- struct{}{}
    }
    
    // Start worker goroutines
    for i := 0; i < size; i++ {
        pool.wg.Add(1)
        go pool.worker(i)
    }
    
    gpm.mu.Lock()
    if gpm.pools == nil {
        gpm.pools = make(map[string]*GoroutinePool)
    }
    gpm.pools[name] = pool
    gpm.mu.Unlock()
    
    return pool
}

func (gp *GoroutinePool) worker(id int) {
    defer gp.wg.Done()
    
    for {
        select {
        case job := <-gp.jobs:
            atomic.AddInt32(&gp.metrics.WorkersActive, 1)
            atomic.AddInt32(&gp.metrics.WorkersIdle, -1)
            
            start := time.Now()
            result := job.Execute()
            duration := time.Since(start)
            
            atomic.AddInt64(&gp.metrics.JobsProcessed, 1)
            
            select {
            case gp.results <- result:
            default:
                // Result channel full, drop result
            }
            
            atomic.AddInt32(&gp.metrics.WorkersActive, -1)
            atomic.AddInt32(&gp.metrics.WorkersIdle, 1)
            
            // Update average latency
            gp.updateAverageLatency(duration)
            
        case <-gp.shutdown:
            return
        }
    }
}

func (gp *GoroutinePool) updateAverageLatency(duration time.Duration) {
    // Simple moving average update
    // In practice, you'd use a more sophisticated approach
    currentAvg := gp.metrics.AverageLatency
    processed := atomic.LoadInt64(&gp.metrics.JobsProcessed)
    
    newAvg := (currentAvg*time.Duration(processed-1) + duration) / time.Duration(processed)
    gp.metrics.AverageLatency = newAvg
}

type Job interface {
    Execute() Result
}

type Result interface {
    IsSuccess() bool
    GetData() interface{}
    GetError() error
}
```

## Best Practices

### 1. Goroutine Lifecycle Management

```go
// Best practice: Always use context for cancellation
func properGoroutineManagement(ctx context.Context) {
    // Create cancellable context for goroutines
    workCtx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for {
                select {
                case <-workCtx.Done():
                    return // Graceful shutdown
                default:
                    // Do work
                    time.Sleep(time.Millisecond * 100)
                }
            }
        }(i)
    }
    
    // Wait for completion or timeout
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All goroutines completed
    case <-time.After(time.Second * 5):
        // Timeout, cancel context
        cancel()
        <-done // Wait for graceful shutdown
    }
}
```

### 2. Resource Leak Prevention

```go
// Leak prevention patterns
func preventGoroutineLeaks() {
    // Pattern 1: Always have a way to stop goroutines
    stopCh := make(chan struct{})
    defer close(stopCh)
    
    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                // Periodic work
            case <-stopCh:
                return // Clean exit
            }
        }
    }()
    
    // Pattern 2: Use buffered channels for producers
    results := make(chan Result, 10) // Buffered to prevent blocking
    
    go func() {
        defer close(results)
        for i := 0; i < 5; i++ {
            select {
            case results <- processData(i):
            case <-stopCh:
                return
            }
        }
    }()
    
    // Pattern 3: Always drain channels
    for result := range results {
        handleResult(result)
    }
}

func processData(i int) Result {
    // Placeholder
    return nil
}

func handleResult(result Result) {
    // Placeholder
}
```

## Next Steps

- Study [Goroutine Leak Detection](goroutine-leaks.md) techniques
- Learn [Deadlock Detection](deadlock-detection.md) methods  
- Explore [Worker Pool Optimization](../../optimization/concurrency/worker-pools.md)
- Master [Channel Analysis](../concurrency-profiling/channel-analysis.md)

## Summary

Goroutine analysis enables building efficient concurrent systems by:

1. **Monitoring lifecycle** - Tracking creation, execution, and completion
2. **Detecting issues** - Identifying leaks, deadlocks, and performance problems  
3. **Optimizing patterns** - Right-sizing pools and improving communication
4. **Measuring efficiency** - Quantifying resource utilization and throughput
5. **Preventing problems** - Early detection of anti-patterns and bottlenecks

Use these techniques to build robust, scalable concurrent applications with optimal goroutine management.
