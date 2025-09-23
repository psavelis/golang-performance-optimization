# Goroutine Leak Detection

Goroutine leaks are one of the most critical issues in Go applications, leading to memory exhaustion, resource starvation, and degraded performance. This comprehensive guide covers advanced techniques for detecting, analyzing, and preventing goroutine leaks in production systems.

## Understanding Goroutine Leaks

Goroutine leaks occur when:
- **Blocked operations** - Goroutines wait indefinitely on channels or locks
- **Missing termination** - No proper shutdown mechanism or context cancellation
- **Resource contention** - Deadlocks preventing goroutine completion
- **Infinite loops** - Runaway goroutines consuming CPU indefinitely
- **Lost references** - No way to signal goroutines to terminate

### Advanced Leak Detection System

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

// GoroutineLeakDetector provides comprehensive leak detection and analysis
type GoroutineLeakDetector struct {
    trackedGoroutines  map[int]*TrackedGoroutine
    leakThreshold      time.Duration
    sampleInterval     time.Duration
    maxStackDepth      int
    leakCallbacks      []LeakCallback
    statistics         *LeakStatistics
    enabled            bool
    mu                 sync.RWMutex
}

type TrackedGoroutine struct {
    ID              int
    Name            string
    CreatedAt       time.Time
    LastSeen        time.Time
    StackTrace      []string
    State           GoroutineState
    BlockedOn       string
    WaitReason      string
    CreatedBy       string
    Context         string
    ActivityHistory []ActivityRecord
    SuspicionLevel  SuspicionLevel
    LeakType        LeakType
}

type ActivityRecord struct {
    Timestamp   time.Time
    Action      string
    StackTrace  []string
    CPUTime     time.Duration
    MemoryUsage int64
}

type SuspicionLevel int

const (
    SuspicionNone SuspicionLevel = iota
    SuspicionLow
    SuspicionMedium
    SuspicionHigh
    SuspicionCritical
)

func (sl SuspicionLevel) String() string {
    switch sl {
    case SuspicionNone:
        return "None"
    case SuspicionLow:
        return "Low"
    case SuspicionMedium:
        return "Medium"
    case SuspicionHigh:
        return "High"
    case SuspicionCritical:
        return "Critical"
    default:
        return "Unknown"
    }
}

type LeakType int

const (
    LeakTypeUnknown LeakType = iota
    LeakTypeChannelBlocked
    LeakTypeMutexBlocked
    LeakTypeNetworkBlocked
    LeakTypeInfiniteLoop
    LeakTypeResourceLeak
    LeakTypeDeadlock
)

func (lt LeakType) String() string {
    switch lt {
    case LeakTypeChannelBlocked:
        return "Channel Blocked"
    case LeakTypeMutexBlocked:
        return "Mutex Blocked"
    case LeakTypeNetworkBlocked:
        return "Network Blocked"
    case LeakTypeInfiniteLoop:
        return "Infinite Loop"
    case LeakTypeResourceLeak:
        return "Resource Leak"
    case LeakTypeDeadlock:
        return "Deadlock"
    default:
        return "Unknown"
    }
}

type GoroutineState int

const (
    StateRunning GoroutineState = iota
    StateRunnable
    StateWaiting
    StateSyscall
    StateBlocked
    StateDead
)

type LeakCallback func(leak DetectedLeak)

type DetectedLeak struct {
    Goroutine    *TrackedGoroutine
    Age          time.Duration
    Confidence   float64
    Evidence     []Evidence
    Severity     Severity
    DetectedAt   time.Time
}

type Evidence struct {
    Type        string
    Description string
    Timestamp   time.Time
    Data        interface{}
}

type Severity int

const (
    SeverityLow Severity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)

func (s Severity) String() string {
    switch s {
    case SeverityLow:
        return "Low"
    case SeverityMedium:
        return "Medium"
    case SeverityHigh:
        return "High"
    case SeverityCritical:
        return "Critical"
    default:
        return "Unknown"
    }
}

type LeakStatistics struct {
    TotalGoroutines    int64
    SuspiciousCount    int64
    ConfirmedLeaks     int64
    FalsePositives     int64
    AverageAge         time.Duration
    OldestGoroutine    time.Duration
    LeakRate           float64
    DetectionAccuracy  float64
    LastScanTime       time.Time
}

func NewGoroutineLeakDetector() *GoroutineLeakDetector {
    return &GoroutineLeakDetector{
        trackedGoroutines: make(map[int]*TrackedGoroutine),
        leakThreshold:     time.Minute * 5,  // 5 minutes default
        sampleInterval:    time.Second * 10, // 10 seconds default
        maxStackDepth:     50,
        statistics:        &LeakStatistics{},
    }
}

func (gld *GoroutineLeakDetector) Enable() {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    
    if gld.enabled {
        return
    }
    
    gld.enabled = true
    go gld.monitorGoroutines()
    go gld.analyzeLeaks()
}

func (gld *GoroutineLeakDetector) Disable() {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    gld.enabled = false
}

func (gld *GoroutineLeakDetector) SetLeakThreshold(threshold time.Duration) {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    gld.leakThreshold = threshold
}

func (gld *GoroutineLeakDetector) RegisterLeakCallback(callback LeakCallback) {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    gld.leakCallbacks = append(gld.leakCallbacks, callback)
}

func (gld *GoroutineLeakDetector) monitorGoroutines() {
    ticker := time.NewTicker(gld.sampleInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        gld.mu.RLock()
        enabled := gld.enabled
        gld.mu.RUnlock()
        
        if !enabled {
            return
        }
        
        gld.scanGoroutines()
    }
}

func (gld *GoroutineLeakDetector) scanGoroutines() {
    // Get current goroutine stack traces
    buf := make([]byte, 64*1024*1024) // 64MB buffer for stack traces
    stackSize := runtime.Stack(buf, true)
    stacks := string(buf[:stackSize])
    
    // Parse stack traces to extract goroutine information
    goroutines := gld.parseStackTraces(stacks)
    
    gld.mu.Lock()
    defer gld.mu.Unlock()
    
    now := time.Now()
    seenIDs := make(map[int]bool)
    
    // Update tracked goroutines
    for _, goroutineInfo := range goroutines {
        seenIDs[goroutineInfo.ID] = true
        
        if existing, exists := gld.trackedGoroutines[goroutineInfo.ID]; exists {
            // Update existing goroutine
            existing.LastSeen = now
            existing.StackTrace = goroutineInfo.StackTrace
            existing.State = goroutineInfo.State
            existing.BlockedOn = goroutineInfo.BlockedOn
            existing.WaitReason = goroutineInfo.WaitReason
            
            // Record activity
            activity := ActivityRecord{
                Timestamp:  now,
                Action:     "stack_update",
                StackTrace: goroutineInfo.StackTrace,
            }
            existing.ActivityHistory = append(existing.ActivityHistory, activity)
            
            // Keep activity history manageable
            if len(existing.ActivityHistory) > 100 {
                existing.ActivityHistory = existing.ActivityHistory[len(existing.ActivityHistory)-100:]
            }
            
            // Update suspicion level
            existing.SuspicionLevel = gld.calculateSuspicionLevel(existing)
            existing.LeakType = gld.detectLeakType(existing)
            
        } else {
            // New goroutine
            tracked := &TrackedGoroutine{
                ID:              goroutineInfo.ID,
                Name:            goroutineInfo.Name,
                CreatedAt:       now,
                LastSeen:        now,
                StackTrace:      goroutineInfo.StackTrace,
                State:           goroutineInfo.State,
                BlockedOn:       goroutineInfo.BlockedOn,
                WaitReason:      goroutineInfo.WaitReason,
                CreatedBy:       gld.extractCreatedBy(goroutineInfo.StackTrace),
                Context:         gld.extractContext(goroutineInfo.StackTrace),
                SuspicionLevel:  SuspicionNone,
                LeakType:        LeakTypeUnknown,
                ActivityHistory: []ActivityRecord{{
                    Timestamp:  now,
                    Action:     "created",
                    StackTrace: goroutineInfo.StackTrace,
                }},
            }
            
            gld.trackedGoroutines[goroutineInfo.ID] = tracked
            atomic.AddInt64(&gld.statistics.TotalGoroutines, 1)
        }
    }
    
    // Mark missing goroutines as dead
    for id, tracked := range gld.trackedGoroutines {
        if !seenIDs[id] && tracked.State != StateDead {
            tracked.State = StateDead
            tracked.LastSeen = now
            
            activity := ActivityRecord{
                Timestamp: now,
                Action:    "terminated",
            }
            tracked.ActivityHistory = append(tracked.ActivityHistory, activity)
        }
    }
    
    gld.statistics.LastScanTime = now
}

type GoroutineInfo struct {
    ID         int
    Name       string
    State      GoroutineState
    StackTrace []string
    BlockedOn  string
    WaitReason string
}

func (gld *GoroutineLeakDetector) parseStackTraces(stacks string) []GoroutineInfo {
    var goroutines []GoroutineInfo
    
    // Split by goroutine boundaries
    goroutineBlocks := strings.Split(stacks, "\n\ngoroutine ")
    
    for i, block := range goroutineBlocks {
        if i == 0 {
            // First block has different format
            block = strings.TrimPrefix(block, "goroutine ")
        }
        
        if strings.TrimSpace(block) == "" {
            continue
        }
        
        goroutineInfo := gld.parseGoroutineBlock(block)
        if goroutineInfo.ID > 0 {
            goroutines = append(goroutines, goroutineInfo)
        }
    }
    
    return goroutines
}

func (gld *GoroutineLeakDetector) parseGoroutineBlock(block string) GoroutineInfo {
    lines := strings.Split(block, "\n")
    if len(lines) == 0 {
        return GoroutineInfo{}
    }
    
    // Parse first line: "123 [running]: main.main()"
    firstLine := lines[0]
    parts := strings.SplitN(firstLine, " ", 2)
    if len(parts) < 2 {
        return GoroutineInfo{}
    }
    
    // Extract goroutine ID
    var id int
    fmt.Sscanf(parts[0], "%d", &id)
    
    // Extract state
    statePart := parts[1]
    var state GoroutineState
    var blockedOn, waitReason string
    
    if strings.Contains(statePart, "[running]") {
        state = StateRunning
    } else if strings.Contains(statePart, "[runnable]") {
        state = StateRunnable
    } else if strings.Contains(statePart, "[syscall]") {
        state = StateSyscall
    } else if strings.Contains(statePart, "[chan") {
        state = StateBlocked
        blockedOn = "channel"
        if strings.Contains(statePart, "chan receive") {
            waitReason = "channel receive"
        } else if strings.Contains(statePart, "chan send") {
            waitReason = "channel send"
        }
    } else if strings.Contains(statePart, "[semacquire]") {
        state = StateBlocked
        blockedOn = "semaphore"
        waitReason = "semaphore acquire"
    } else if strings.Contains(statePart, "[select]") {
        state = StateBlocked
        blockedOn = "select"
        waitReason = "select statement"
    } else {
        state = StateWaiting
    }
    
    // Extract stack trace
    var stackTrace []string
    for i := 1; i < len(lines) && len(stackTrace) < gld.maxStackDepth; i++ {
        line := strings.TrimSpace(lines[i])
        if line != "" {
            stackTrace = append(stackTrace, line)
        }
    }
    
    // Extract function name for goroutine name
    var name string
    if len(stackTrace) > 0 {
        // First stack frame usually contains the function name
        if strings.Contains(stackTrace[0], "(") {
            name = strings.Split(stackTrace[0], "(")[0]
        } else {
            name = stackTrace[0]
        }
    }
    
    return GoroutineInfo{
        ID:         id,
        Name:       name,
        State:      state,
        StackTrace: stackTrace,
        BlockedOn:  blockedOn,
        WaitReason: waitReason,
    }
}

func (gld *GoroutineLeakDetector) extractCreatedBy(stackTrace []string) string {
    // Look for the function that created this goroutine
    for _, frame := range stackTrace {
        if strings.Contains(frame, "go ") || strings.Contains(frame, "created by") {
            return frame
        }
    }
    
    if len(stackTrace) > 0 {
        return stackTrace[len(stackTrace)-1] // Last frame
    }
    
    return "unknown"
}

func (gld *GoroutineLeakDetector) extractContext(stackTrace []string) string {
    // Extract context information from stack trace
    for _, frame := range stackTrace {
        if strings.Contains(frame, "context.") {
            return frame
        }
    }
    return ""
}

func (gld *GoroutineLeakDetector) calculateSuspicionLevel(goroutine *TrackedGoroutine) SuspicionLevel {
    age := time.Since(goroutine.CreatedAt)
    
    // Age-based suspicion
    if age > time.Hour {
        return SuspicionCritical
    } else if age > time.Minute*30 {
        return SuspicionHigh
    } else if age > time.Minute*10 {
        return SuspicionMedium
    } else if age > gld.leakThreshold {
        return SuspicionLow
    }
    
    // State-based suspicion
    if goroutine.State == StateBlocked {
        if age > time.Minute*5 {
            return SuspicionHigh
        } else if age > time.Minute*2 {
            return SuspicionMedium
        }
    }
    
    // Stack-based suspicion
    if gld.isInfiniteLoopSuspected(goroutine) {
        return SuspicionCritical
    }
    
    return SuspicionNone
}

func (gld *GoroutineLeakDetector) detectLeakType(goroutine *TrackedGoroutine) LeakType {
    // Analyze stack trace to determine leak type
    stackStr := strings.Join(goroutine.StackTrace, " ")
    
    if strings.Contains(stackStr, "chan ") || strings.Contains(stackStr, "select") {
        return LeakTypeChannelBlocked
    }
    
    if strings.Contains(stackStr, "sync.") && strings.Contains(stackStr, "Lock") {
        return LeakTypeMutexBlocked
    }
    
    if strings.Contains(stackStr, "net.") || strings.Contains(stackStr, "Read") || strings.Contains(stackStr, "Write") {
        return LeakTypeNetworkBlocked
    }
    
    if gld.isInfiniteLoopSuspected(goroutine) {
        return LeakTypeInfiniteLoop
    }
    
    if strings.Contains(stackStr, "deadlock") {
        return LeakTypeDeadlock
    }
    
    if goroutine.State == StateBlocked {
        return LeakTypeResourceLeak
    }
    
    return LeakTypeUnknown
}

func (gld *GoroutineLeakDetector) isInfiniteLoopSuspected(goroutine *TrackedGoroutine) bool {
    // Check if goroutine has been in same state for too long
    if len(goroutine.ActivityHistory) < 10 {
        return false
    }
    
    recent := goroutine.ActivityHistory[len(goroutine.ActivityHistory)-10:]
    
    // If all recent activities show same stack trace, might be infinite loop
    firstStack := strings.Join(recent[0].StackTrace, "")
    for _, activity := range recent[1:] {
        if strings.Join(activity.StackTrace, "") != firstStack {
            return false
        }
    }
    
    return true
}

func (gld *GoroutineLeakDetector) analyzeLeaks() {
    ticker := time.NewTicker(gld.leakThreshold / 2) // Check more frequently than threshold
    defer ticker.Stop()
    
    for range ticker.C {
        gld.mu.RLock()
        enabled := gld.enabled
        gld.mu.RUnlock()
        
        if !enabled {
            return
        }
        
        leaks := gld.detectLeaks()
        
        for _, leak := range leaks {
            gld.notifyLeakCallbacks(leak)
        }
    }
}

func (gld *GoroutineLeakDetector) detectLeaks() []DetectedLeak {
    gld.mu.RLock()
    defer gld.mu.RUnlock()
    
    var leaks []DetectedLeak
    now := time.Now()
    
    for _, goroutine := range gld.trackedGoroutines {
        if goroutine.State == StateDead {
            continue
        }
        
        age := now.Sub(goroutine.CreatedAt)
        
        // Check if goroutine meets leak criteria
        if goroutine.SuspicionLevel >= SuspicionMedium {
            confidence := gld.calculateLeakConfidence(goroutine)
            
            if confidence >= 0.6 { // 60% confidence threshold
                evidence := gld.gatherEvidence(goroutine)
                severity := gld.calculateSeverity(goroutine, confidence)
                
                leak := DetectedLeak{
                    Goroutine:  goroutine,
                    Age:        age,
                    Confidence: confidence,
                    Evidence:   evidence,
                    Severity:   severity,
                    DetectedAt: now,
                }
                
                leaks = append(leaks, leak)
                atomic.AddInt64(&gld.statistics.ConfirmedLeaks, 1)
            }
        }
    }
    
    return leaks
}

func (gld *GoroutineLeakDetector) calculateLeakConfidence(goroutine *TrackedGoroutine) float64 {
    confidence := 0.0
    
    // Age factor
    age := time.Since(goroutine.CreatedAt)
    if age > time.Hour {
        confidence += 0.4
    } else if age > time.Minute*30 {
        confidence += 0.3
    } else if age > time.Minute*10 {
        confidence += 0.2
    } else if age > gld.leakThreshold {
        confidence += 0.1
    }
    
    // State factor
    switch goroutine.State {
    case StateBlocked:
        confidence += 0.3
    case StateWaiting:
        confidence += 0.2
    case StateRunning:
        if gld.isInfiniteLoopSuspected(goroutine) {
            confidence += 0.4
        }
    }
    
    // Stack trace analysis
    stackStr := strings.Join(goroutine.StackTrace, " ")
    
    // Known problematic patterns
    if strings.Contains(stackStr, "chan ") && goroutine.State == StateBlocked {
        confidence += 0.2
    }
    
    if strings.Contains(stackStr, "sync.") && strings.Contains(stackStr, "Lock") {
        confidence += 0.2
    }
    
    if strings.Contains(stackStr, "for {") || strings.Contains(stackStr, "infinite") {
        confidence += 0.3
    }
    
    // Activity pattern analysis
    if len(goroutine.ActivityHistory) > 5 {
        recentActivity := goroutine.ActivityHistory[len(goroutine.ActivityHistory)-5:]
        if gld.isStuckPattern(recentActivity) {
            confidence += 0.2
        }
    }
    
    return confidence
}

func (gld *GoroutineLeakDetector) isStuckPattern(activities []ActivityRecord) bool {
    if len(activities) < 3 {
        return false
    }
    
    // Check if recent activities show no progress
    firstStack := strings.Join(activities[0].StackTrace, "")
    for _, activity := range activities[1:] {
        if strings.Join(activity.StackTrace, "") != firstStack {
            return false
        }
    }
    
    return true
}

func (gld *GoroutineLeakDetector) gatherEvidence(goroutine *TrackedGoroutine) []Evidence {
    var evidence []Evidence
    
    // Age evidence
    age := time.Since(goroutine.CreatedAt)
    evidence = append(evidence, Evidence{
        Type:        "age",
        Description: fmt.Sprintf("Goroutine has been running for %v", age),
        Timestamp:   time.Now(),
        Data:        age,
    })
    
    // State evidence
    evidence = append(evidence, Evidence{
        Type:        "state",
        Description: fmt.Sprintf("Goroutine is in %v state", goroutine.State),
        Timestamp:   time.Now(),
        Data:        goroutine.State,
    })
    
    // Stack trace evidence
    if len(goroutine.StackTrace) > 0 {
        evidence = append(evidence, Evidence{
            Type:        "stack_trace",
            Description: "Current stack trace shows blocking operation",
            Timestamp:   time.Now(),
            Data:        goroutine.StackTrace,
        })
    }
    
    // Blocking evidence
    if goroutine.BlockedOn != "" {
        evidence = append(evidence, Evidence{
            Type:        "blocked_on",
            Description: fmt.Sprintf("Blocked on %s: %s", goroutine.BlockedOn, goroutine.WaitReason),
            Timestamp:   time.Now(),
            Data:        map[string]string{"blocked_on": goroutine.BlockedOn, "reason": goroutine.WaitReason},
        })
    }
    
    // Activity pattern evidence
    if gld.isInfiniteLoopSuspected(goroutine) {
        evidence = append(evidence, Evidence{
            Type:        "infinite_loop",
            Description: "Goroutine shows repetitive stack trace pattern indicating infinite loop",
            Timestamp:   time.Now(),
            Data:        goroutine.ActivityHistory,
        })
    }
    
    return evidence
}

func (gld *GoroutineLeakDetector) calculateSeverity(goroutine *TrackedGoroutine, confidence float64) Severity {
    age := time.Since(goroutine.CreatedAt)
    
    if confidence >= 0.9 || age > time.Hour*2 {
        return SeverityCritical
    } else if confidence >= 0.8 || age > time.Hour {
        return SeverityHigh
    } else if confidence >= 0.7 || age > time.Minute*30 {
        return SeverityMedium
    } else {
        return SeverityLow
    }
}

func (gld *GoroutineLeakDetector) notifyLeakCallbacks(leak DetectedLeak) {
    gld.mu.RLock()
    callbacks := make([]LeakCallback, len(gld.leakCallbacks))
    copy(callbacks, gld.leakCallbacks)
    gld.mu.RUnlock()
    
    for _, callback := range callbacks {
        go func(cb LeakCallback) {
            defer func() {
                if r := recover(); r != nil {
                    // Prevent callback panics from affecting detector
                }
            }()
            cb(leak)
        }(callback)
    }
}

func (gld *GoroutineLeakDetector) GetStatistics() LeakStatistics {
    gld.mu.RLock()
    defer gld.mu.RUnlock()
    
    stats := *gld.statistics
    
    // Calculate current statistics
    var totalAge time.Duration
    var oldestAge time.Duration
    var suspiciousCount int64
    
    for _, goroutine := range gld.trackedGoroutines {
        if goroutine.State == StateDead {
            continue
        }
        
        age := time.Since(goroutine.CreatedAt)
        totalAge += age
        
        if age > oldestAge {
            oldestAge = age
        }
        
        if goroutine.SuspicionLevel >= SuspicionMedium {
            suspiciousCount++
        }
    }
    
    activeCount := int64(len(gld.trackedGoroutines))
    if activeCount > 0 {
        stats.AverageAge = totalAge / time.Duration(activeCount)
    }
    
    stats.OldestGoroutine = oldestAge
    stats.SuspiciousCount = suspiciousCount
    
    if stats.TotalGoroutines > 0 {
        stats.LeakRate = float64(stats.ConfirmedLeaks) / float64(stats.TotalGoroutines) * 100
    }
    
    return stats
}

func (gld *GoroutineLeakDetector) GetDetailedReport() DetailedLeakReport {
    gld.mu.RLock()
    defer gld.mu.RUnlock()
    
    var activeGoroutines []*TrackedGoroutine
    var suspiciousGoroutines []*TrackedGoroutine
    var deadGoroutines []*TrackedGoroutine
    
    for _, goroutine := range gld.trackedGoroutines {
        if goroutine.State == StateDead {
            deadGoroutines = append(deadGoroutines, goroutine)
        } else {
            activeGoroutines = append(activeGoroutines, goroutine)
            
            if goroutine.SuspicionLevel >= SuspicionMedium {
                suspiciousGoroutines = append(suspiciousGoroutines, goroutine)
            }
        }
    }
    
    // Sort by age (oldest first)
    sort.Slice(activeGoroutines, func(i, j int) bool {
        return activeGoroutines[i].CreatedAt.Before(activeGoroutines[j].CreatedAt)
    })
    
    sort.Slice(suspiciousGoroutines, func(i, j int) bool {
        return suspiciousGoroutines[i].SuspicionLevel > suspiciousGoroutines[j].SuspicionLevel
    })
    
    return DetailedLeakReport{
        Statistics:            gld.GetStatistics(),
        ActiveGoroutines:      activeGoroutines,
        SuspiciousGoroutines:  suspiciousGoroutines,
        DeadGoroutines:        deadGoroutines,
        GeneratedAt:           time.Now(),
    }
}

type DetailedLeakReport struct {
    Statistics           LeakStatistics
    ActiveGoroutines     []*TrackedGoroutine
    SuspiciousGoroutines []*TrackedGoroutine
    DeadGoroutines       []*TrackedGoroutine
    GeneratedAt          time.Time
}

func (dlr DetailedLeakReport) String() string {
    result := fmt.Sprintf(`Goroutine Leak Detection Report
Generated: %v

=== STATISTICS ===
Total Goroutines: %d
Suspicious Count: %d
Confirmed Leaks: %d
False Positives: %d
Average Age: %v
Oldest Goroutine: %v
Leak Rate: %.1f%%
Detection Accuracy: %.1f%%

=== ACTIVE GOROUTINES ===
Total: %d`,
        dlr.GeneratedAt.Format(time.RFC3339),
        dlr.Statistics.TotalGoroutines,
        dlr.Statistics.SuspiciousCount,
        dlr.Statistics.ConfirmedLeaks,
        dlr.Statistics.FalsePositives,
        dlr.Statistics.AverageAge,
        dlr.Statistics.OldestGoroutine,
        dlr.Statistics.LeakRate,
        dlr.Statistics.DetectionAccuracy,
        len(dlr.ActiveGoroutines))
    
    // Show top 5 oldest active goroutines
    for i, goroutine := range dlr.ActiveGoroutines {
        if i >= 5 {
            result += "\n  ..."
            break
        }
        
        age := time.Since(goroutine.CreatedAt)
        result += fmt.Sprintf("\n  %d. ID:%d Age:%v State:%v Suspicion:%v",
            i+1, goroutine.ID, age, goroutine.State, goroutine.SuspicionLevel)
    }
    
    if len(dlr.SuspiciousGoroutines) > 0 {
        result += fmt.Sprintf("\n\n=== SUSPICIOUS GOROUTINES ===\nTotal: %d", len(dlr.SuspiciousGoroutines))
        
        for i, goroutine := range dlr.SuspiciousGoroutines {
            if i >= 10 {
                result += "\n  ..."
                break
            }
            
            age := time.Since(goroutine.CreatedAt)
            result += fmt.Sprintf("\n  %d. ID:%d Age:%v Type:%v Suspicion:%v",
                i+1, goroutine.ID, age, goroutine.LeakType, goroutine.SuspicionLevel)
            
            if goroutine.BlockedOn != "" {
                result += fmt.Sprintf(" (blocked on %s)", goroutine.BlockedOn)
            }
            
            if len(goroutine.StackTrace) > 0 {
                result += fmt.Sprintf("\n     Stack: %s", goroutine.StackTrace[0])
            }
        }
    }
    
    return result
}

func (dl DetectedLeak) String() string {
    return fmt.Sprintf(`LEAK DETECTED - ID:%d Severity:%v Confidence:%.1f%%
Age: %v
Type: %v
State: %v
Created By: %s
Stack: %s
Evidence: %d items`,
        dl.Goroutine.ID,
        dl.Severity,
        dl.Confidence*100,
        dl.Age,
        dl.Goroutine.LeakType,
        dl.Goroutine.State,
        dl.Goroutine.CreatedBy,
        func() string {
            if len(dl.Goroutine.StackTrace) > 0 {
                return dl.Goroutine.StackTrace[0]
            }
            return "unknown"
        }(),
        len(dl.Evidence))
}

func demonstrateLeakDetection() {
    fmt.Println("=== GOROUTINE LEAK DETECTION DEMONSTRATION ===")
    
    detector := NewGoroutineLeakDetector()
    detector.SetLeakThreshold(time.Second * 5) // Short threshold for demo
    
    // Register leak callback
    detector.RegisterLeakCallback(func(leak DetectedLeak) {
        fmt.Printf("\n🚨 LEAK ALERT: %s\n", leak)
    })
    
    detector.Enable()
    defer detector.Disable()
    
    // Create various types of potentially leaking goroutines
    
    // 1. Channel blocked goroutine
    ch := make(chan int)
    go func() {
        fmt.Println("Goroutine waiting on channel...")
        <-ch // Will block forever
    }()
    
    // 2. Mutex blocked goroutine
    var mu sync.Mutex
    mu.Lock() // Lock it first
    go func() {
        fmt.Println("Goroutine waiting on mutex...")
        mu.Lock() // Will block forever
        defer mu.Unlock()
    }()
    
    // 3. Infinite loop goroutine
    go func() {
        fmt.Println("Goroutine in infinite loop...")
        for {
            // Busy loop
            time.Sleep(time.Nanosecond)
        }
    }()
    
    // 4. Normal goroutine that will complete
    go func() {
        fmt.Println("Normal goroutine working...")
        time.Sleep(time.Second * 2)
        fmt.Println("Normal goroutine completed")
    }()
    
    // 5. Context-aware goroutine (good practice)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
    defer cancel()
    
    go func() {
        fmt.Println("Context-aware goroutine...")
        select {
        case <-ctx.Done():
            fmt.Println("Context-aware goroutine cancelled")
        case <-time.After(time.Second * 10):
            fmt.Println("This shouldn't happen")
        }
    }()
    
    // Wait for leak detection to run
    fmt.Println("\nWaiting for leak detection...")
    time.Sleep(time.Second * 10)
    
    // Get detailed report
    report := detector.GetDetailedReport()
    fmt.Printf("\n%s\n", report)
    
    // Get statistics
    stats := detector.GetStatistics()
    fmt.Printf("\nLeak Detection Statistics:\n")
    fmt.Printf("Leak Rate: %.1f%%\n", stats.LeakRate)
    fmt.Printf("Average Goroutine Age: %v\n", stats.AverageAge)
    fmt.Printf("Oldest Goroutine: %v\n", stats.OldestGoroutine)
}
```

## Prevention Strategies

### 1. Context-Based Lifecycle Management

```go
// Always use context for goroutine lifecycle management
func properGoroutineManagement(ctx context.Context) error {
    // Create cancellable context
    workCtx, cancel := context.WithCancel(ctx)
    defer cancel() // Ensure cleanup
    
    // Channel for goroutine completion
    done := make(chan error, 1)
    
    go func() {
        defer close(done)
        
        // Simulate work with context checking
        for {
            select {
            case <-workCtx.Done():
                done <- workCtx.Err()
                return
            default:
                // Do actual work
                time.Sleep(time.Millisecond * 100)
            }
        }
    }()
    
    // Wait for completion or timeout
    select {
    case err := <-done:
        return err
    case <-time.After(time.Second * 30):
        cancel() // Cancel on timeout
        return <-done // Wait for graceful shutdown
    }
}
```

### 2. Channel Best Practices

```go
// Prevent channel-related leaks
func safeChannelUsage() {
    // Always use buffered channels for fire-and-forget
    results := make(chan Result, 10)
    
    // Producer with timeout
    go func() {
        defer close(results)
        
        for i := 0; i < 5; i++ {
            select {
            case results <- processData(i):
                // Success
            case <-time.After(time.Second * 5):
                // Timeout - prevent blocking forever
                return
            }
        }
    }()
    
    // Consumer with timeout
    timeout := time.After(time.Second * 30)
    for {
        select {
        case result, ok := <-results:
            if !ok {
                return // Channel closed
            }
            handleResult(result)
            
        case <-timeout:
            return // Prevent infinite waiting
        }
    }
}

func processData(i int) Result {
    // Placeholder
    return Result{}
}

func handleResult(result Result) {
    // Placeholder
}

type Result struct{}
```

### 3. Sync Primitive Safety

```go
// Safe mutex usage patterns
type SafeCounter struct {
    mu    sync.RWMutex
    value int64
    done  chan struct{}
}

func NewSafeCounter() *SafeCounter {
    return &SafeCounter{
        done: make(chan struct{}),
    }
}

func (sc *SafeCounter) Increment(ctx context.Context) error {
    // Try to acquire lock with context
    acquired := make(chan struct{})
    
    go func() {
        sc.mu.Lock()
        close(acquired)
    }()
    
    select {
    case <-acquired:
        defer sc.mu.Unlock()
        sc.value++
        return nil
        
    case <-ctx.Done():
        return ctx.Err()
        
    case <-sc.done:
        return fmt.Errorf("counter closed")
    }
}

func (sc *SafeCounter) Close() {
    close(sc.done)
}
```

## Monitoring Integration

### 1. Metrics Export

```go
// Export leak detection metrics
type LeakMetrics struct {
    detector *GoroutineLeakDetector
}

func (lm *LeakMetrics) ExportMetrics() map[string]interface{} {
    stats := lm.detector.GetStatistics()
    
    return map[string]interface{}{
        "goroutines_total":      stats.TotalGoroutines,
        "goroutines_suspicious": stats.SuspiciousCount,
        "leaks_confirmed":       stats.ConfirmedLeaks,
        "leaks_false_positive":  stats.FalsePositives,
        "leak_rate_percent":     stats.LeakRate,
        "detection_accuracy":    stats.DetectionAccuracy,
        "oldest_goroutine_age_seconds": stats.OldestGoroutine.Seconds(),
        "average_goroutine_age_seconds": stats.AverageAge.Seconds(),
    }
}
```

### 2. Alerting Integration

```go
// Alert manager for leak detection
type LeakAlertManager struct {
    detector      *GoroutineLeakDetector
    alertThreshold int
    webhookURL     string
    mu             sync.RWMutex
}

func (lam *LeakAlertManager) SetupAlerting() {
    lam.detector.RegisterLeakCallback(func(leak DetectedLeak) {
        if leak.Severity >= SeverityHigh {
            lam.sendAlert(leak)
        }
    })
}

func (lam *LeakAlertManager) sendAlert(leak DetectedLeak) {
    alert := map[string]interface{}{
        "title":       "Goroutine Leak Detected",
        "severity":    leak.Severity.String(),
        "confidence":  leak.Confidence,
        "goroutine_id": leak.Goroutine.ID,
        "age_seconds": leak.Age.Seconds(),
        "leak_type":   leak.Goroutine.LeakType.String(),
        "stack_trace": leak.Goroutine.StackTrace,
        "evidence":    leak.Evidence,
    }
    
    // Send to monitoring system
    go lam.sendWebhook(alert)
}

func (lam *LeakAlertManager) sendWebhook(alert map[string]interface{}) {
    // Implementation would send HTTP POST to webhook URL
    fmt.Printf("ALERT: %+v\n", alert)
}
```

## Next Steps

- Study [Deadlock Detection](deadlock-detection.md) techniques
- Learn [Goroutine Analysis](goroutine-analysis.md) patterns
- Explore [Channel Analysis](../concurrency-profiling/channel-analysis.md)
- Master [Worker Pool Optimization](../../optimization/concurrency/worker-pools.md)

## Summary

Goroutine leak detection enables building robust concurrent applications by:

1. **Early detection** - Identifying potential leaks before they cause problems
2. **Root cause analysis** - Understanding why goroutines are not terminating
3. **Prevention patterns** - Using proper lifecycle management and cancellation
4. **Monitoring integration** - Tracking leak metrics in production systems
5. **Automated remediation** - Taking action when leaks are detected

Use these techniques to build leak-free Go applications that maintain optimal resource usage over time.
