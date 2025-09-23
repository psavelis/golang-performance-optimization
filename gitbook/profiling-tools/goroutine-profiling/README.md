# Goroutine Profiling

Goroutine profiling provides deep insights into concurrent execution patterns, identifies deadlocks, reveals goroutine leaks, and guides concurrency optimization. This comprehensive guide covers all aspects of goroutine profiling in Go applications.

## Introduction to Goroutine Profiling

Goroutine profiling captures the state and behavior of all goroutines in your application:
- **Goroutine States**: Running, waiting, blocked, idle
- **Call Stacks**: Where each goroutine is currently executing
- **Wait Reasons**: Why goroutines are blocked or waiting
- **Goroutine Lifecycle**: Creation, execution, and termination patterns

### Key Concepts

- **Goroutine Stack**: Call chain showing execution context
- **Wait Channels**: Goroutines blocked on channel operations
- **Mutex Contention**: Goroutines waiting for locks
- **System Calls**: Goroutines blocked in system operations

## Collecting Goroutine Profiles

### Using net/http/pprof

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    _ "net/http/pprof"
    "sync"
    "time"
)

func main() {
    // Enable pprof endpoints
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Start various goroutine patterns
    startWorkerPool()
    startChannelProcessors()
    startMutexContentionExample()
    
    // Keep main running
    select {}
}

func startWorkerPool() {
    jobs := make(chan int, 100)
    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                // Simulate work
                time.Sleep(time.Duration(job%10) * time.Millisecond)
                fmt.Printf("Worker %d processed job %d\n", workerID, job)
            }
        }(i)
    }

    // Send jobs
    go func() {
        for i := 0; i < 1000; i++ {
            jobs <- i
            time.Sleep(10 * time.Millisecond)
        }
        close(jobs)
    }()

    go func() {
        wg.Wait()
        fmt.Println("All workers completed")
    }()
}

func startChannelProcessors() {
    // Pipeline pattern
    numbers := make(chan int)
    squares := make(chan int)
    cubes := make(chan int)

    // Stage 1: Generate numbers
    go func() {
        defer close(numbers)
        for i := 1; i <= 100; i++ {
            numbers <- i
            time.Sleep(50 * time.Millisecond)
        }
    }()

    // Stage 2: Calculate squares
    go func() {
        defer close(squares)
        for n := range numbers {
            squares <- n * n
            time.Sleep(20 * time.Millisecond)
        }
    }()

    // Stage 3: Calculate cubes
    go func() {
        defer close(cubes)
        for s := range squares {
            cubes <- s * s * s
            time.Sleep(30 * time.Millisecond)
        }
    }()

    // Consumer
    go func() {
        for c := range cubes {
            fmt.Printf("Result: %d\n", c)
        }
    }()
}

func startMutexContentionExample() {
    var mu sync.Mutex
    var counter int

    // Multiple goroutines competing for the same mutex
    for i := 0; i < 20; i++ {
        go func(id int) {
            for j := 0; j < 100; j++ {
                mu.Lock()
                counter++
                // Simulate work under lock
                time.Sleep(time.Millisecond)
                mu.Unlock()
                
                // Brief pause between operations
                time.Sleep(5 * time.Millisecond)
            }
        }(i)
    }
}
```

### HTTP Endpoints for Goroutine Data

```bash
# Get current goroutine stack traces
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Get goroutine data in text format
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Get detailed goroutine information
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Get full goroutine stacks with all frames
curl "http://localhost:6060/debug/pprof/goroutine?debug=2&full=1"
```

### Programmatic Goroutine Profiling

```go
package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "strconv"
    "strings"
    "time"
)

func captureGoroutineProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("could not create goroutine profile: %v", err)
    }
    defer f.Close()

    if err := pprof.Lookup("goroutine").WriteTo(f, 2); err != nil {
        return fmt.Errorf("could not write goroutine profile: %v", err)
    }

    return nil
}

func analyzeGoroutines() {
    // Capture initial state
    captureGoroutineProfile("goroutines_initial.prof")
    
    // Start some goroutines
    startTestGoroutines()
    
    // Wait a bit for goroutines to establish
    time.Sleep(2 * time.Second)
    
    // Capture after startup
    captureGoroutineProfile("goroutines_running.prof")
    
    // Get runtime statistics
    stats := getGoroutineStats()
    fmt.Printf("Goroutine Statistics:\n%s\n", stats)
}

func startTestGoroutines() {
    // CPU-intensive goroutines
    for i := 0; i < 5; i++ {
        go func(id int) {
            for {
                // CPU work
                sum := 0
                for j := 0; j < 100000; j++ {
                    sum += j
                }
                time.Sleep(10 * time.Millisecond)
            }
        }(i)
    }

    // I/O-waiting goroutines
    for i := 0; i < 3; i++ {
        go func(id int) {
            for {
                time.Sleep(1 * time.Second)
                fmt.Printf("I/O goroutine %d tick\n", id)
            }
        }(i)
    }

    // Channel-waiting goroutines
    ch := make(chan int)
    for i := 0; i < 4; i++ {
        go func(id int) {
            for val := range ch {
                fmt.Printf("Goroutine %d received: %d\n", id, val)
            }
        }(i)
    }

    // Channel sender
    go func() {
        for i := 0; ; i++ {
            ch <- i
            time.Sleep(500 * time.Millisecond)
        }
    }()
}

func getGoroutineStats() string {
    var stats strings.Builder
    
    // Get total goroutine count
    count := runtime.NumGoroutine()
    stats.WriteString(fmt.Sprintf("Total Goroutines: %d\n", count))
    
    // Get detailed breakdown
    profile := pprof.Lookup("goroutine")
    if profile != nil {
        var buf strings.Builder
        profile.WriteTo(&buf, 1)
        
        // Parse profile output for statistics
        lines := strings.Split(buf.String(), "\n")
        stateCount := make(map[string]int)
        
        for _, line := range lines {
            if strings.Contains(line, "goroutine") && strings.Contains(line, "[") {
                // Extract state from line like "goroutine 123 [running]:"
                if start := strings.Index(line, "["); start != -1 {
                    if end := strings.Index(line[start:], "]"); end != -1 {
                        state := line[start+1 : start+end]
                        stateCount[state]++
                    }
                }
            }
        }
        
        stats.WriteString("Goroutine States:\n")
        for state, count := range stateCount {
            stats.WriteString(fmt.Sprintf("  %s: %d\n", state, count))
        }
    }
    
    return stats.String()
}
```

## Advanced Goroutine Analysis

### Goroutine State Monitoring

```go
package main

import (
    "context"
    "fmt"
    "log"
    "runtime"
    "runtime/pprof"
    "sort"
    "strings"
    "sync"
    "time"
)

type GoroutineMonitor struct {
    mu              sync.RWMutex
    enabled         bool
    interval        time.Duration
    history         []GoroutineSnapshot
    maxHistory      int
    alertThreshold  int
}

type GoroutineSnapshot struct {
    Timestamp    time.Time
    Total        int
    StateBreakdown map[string]int
    TopStacks    []StackInfo
}

type StackInfo struct {
    Count    int
    State    string
    Function string
    Stack    string
}

func NewGoroutineMonitor(interval time.Duration, alertThreshold int) *GoroutineMonitor {
    return &GoroutineMonitor{
        interval:       interval,
        alertThreshold: alertThreshold,
        maxHistory:     100,
        history:        make([]GoroutineSnapshot, 0, 100),
    }
}

func (gm *GoroutineMonitor) Start(ctx context.Context) error {
    gm.mu.Lock()
    defer gm.mu.Unlock()

    if gm.enabled {
        return fmt.Errorf("goroutine monitor already running")
    }

    gm.enabled = true
    go gm.monitorLoop(ctx)

    log.Printf("Goroutine monitor started: interval=%v, threshold=%d", gm.interval, gm.alertThreshold)
    return nil
}

func (gm *GoroutineMonitor) Stop() {
    gm.mu.Lock()
    defer gm.mu.Unlock()
    gm.enabled = false
    log.Println("Goroutine monitor stopped")
}

func (gm *GoroutineMonitor) monitorLoop(ctx context.Context) {
    ticker := time.NewTicker(gm.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            gm.captureSnapshot()
        }
    }
}

func (gm *GoroutineMonitor) captureSnapshot() {
    snapshot := gm.analyzeCurrentGoroutines()
    
    gm.mu.Lock()
    gm.history = append(gm.history, snapshot)
    if len(gm.history) > gm.maxHistory {
        gm.history = gm.history[len(gm.history)-gm.maxHistory:]
    }
    gm.mu.Unlock()

    // Check for alerts
    if snapshot.Total > gm.alertThreshold {
        log.Printf("ALERT: High goroutine count: %d (threshold: %d)", snapshot.Total, gm.alertThreshold)
        gm.logDetailedBreakdown(snapshot)
    }

    // Check for potential leaks
    if gm.detectGoroutineLeak() {
        log.Printf("ALERT: Potential goroutine leak detected")
        gm.captureDetailedProfile()
    }

    log.Printf("Goroutines: %d total, states: %v", snapshot.Total, snapshot.StateBreakdown)
}

func (gm *GoroutineMonitor) analyzeCurrentGoroutines() GoroutineSnapshot {
    snapshot := GoroutineSnapshot{
        Timestamp:      time.Now(),
        Total:          runtime.NumGoroutine(),
        StateBreakdown: make(map[string]int),
        TopStacks:      []StackInfo{},
    }

    // Get detailed goroutine information
    profile := pprof.Lookup("goroutine")
    if profile == nil {
        return snapshot
    }

    var buf strings.Builder
    profile.WriteTo(&buf, 1)
    
    // Parse profile for state breakdown
    stackCounts := make(map[string]*StackInfo)
    lines := strings.Split(buf.String(), "\n")
    
    var currentStack strings.Builder
    var currentState string
    var currentFunction string
    
    for i, line := range lines {
        if strings.HasPrefix(line, "goroutine") {
            // Save previous stack if exists
            if currentStack.Len() > 0 {
                stackKey := fmt.Sprintf("%s:%s", currentState, currentFunction)
                if info, exists := stackCounts[stackKey]; exists {
                    info.Count++
                } else {
                    stackCounts[stackKey] = &StackInfo{
                        Count:    1,
                        State:    currentState,
                        Function: currentFunction,
                        Stack:    currentStack.String(),
                    }
                }
            }
            
            // Parse new goroutine header
            if start := strings.Index(line, "["); start != -1 {
                if end := strings.Index(line[start:], "]"); end != -1 {
                    currentState = line[start+1 : start+end]
                    snapshot.StateBreakdown[currentState]++
                }
            }
            
            // Get the function name from next line
            if i+1 < len(lines) {
                currentFunction = strings.TrimSpace(lines[i+1])
            }
            
            currentStack.Reset()
            currentStack.WriteString(line + "\n")
        } else if strings.TrimSpace(line) != "" {
            currentStack.WriteString(line + "\n")
        }
    }

    // Convert to sorted slice
    for _, info := range stackCounts {
        snapshot.TopStacks = append(snapshot.TopStacks, *info)
    }
    
    sort.Slice(snapshot.TopStacks, func(i, j int) bool {
        return snapshot.TopStacks[i].Count > snapshot.TopStacks[j].Count
    })

    // Keep only top 10
    if len(snapshot.TopStacks) > 10 {
        snapshot.TopStacks = snapshot.TopStacks[:10]
    }

    return snapshot
}

func (gm *GoroutineMonitor) detectGoroutineLeak() bool {
    gm.mu.RLock()
    defer gm.mu.RUnlock()

    if len(gm.history) < 10 {
        return false
    }

    // Check if goroutine count is consistently growing
    recentSnapshots := gm.history[len(gm.history)-10:]
    growthCount := 0

    for i := 1; i < len(recentSnapshots); i++ {
        if recentSnapshots[i].Total > recentSnapshots[i-1].Total {
            growthCount++
        }
    }

    // If goroutines grew in 80% of recent samples, consider it a leak
    return float64(growthCount)/float64(len(recentSnapshots)-1) > 0.8
}

func (gm *GoroutineMonitor) logDetailedBreakdown(snapshot GoroutineSnapshot) {
    log.Printf("Detailed Goroutine Breakdown at %s:", snapshot.Timestamp.Format("15:04:05"))
    log.Printf("  Total: %d", snapshot.Total)
    
    for state, count := range snapshot.StateBreakdown {
        log.Printf("  %s: %d", state, count)
    }
    
    log.Printf("Top Stack Traces:")
    for i, stack := range snapshot.TopStacks {
        log.Printf("  %d. %s (%s) - Count: %d", i+1, stack.Function, stack.State, stack.Count)
    }
}

func (gm *GoroutineMonitor) captureDetailedProfile() {
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("goroutine_leak_%s.prof", timestamp)
    
    if err := captureGoroutineProfile(filename); err != nil {
        log.Printf("Failed to capture detailed goroutine profile: %v", err)
        return
    }
    
    log.Printf("Detailed goroutine profile captured: %s", filename)
}

func (gm *GoroutineMonitor) GetReport() string {
    gm.mu.RLock()
    defer gm.mu.RUnlock()

    if len(gm.history) == 0 {
        return "No monitoring data available"
    }

    var report strings.Builder
    report.WriteString("=== GOROUTINE MONITORING REPORT ===\n\n")

    latest := gm.history[len(gm.history)-1]
    report.WriteString(fmt.Sprintf("Current Status (as of %s):\n", latest.Timestamp.Format("2006-01-02 15:04:05")))
    report.WriteString(fmt.Sprintf("  Total Goroutines: %d\n", latest.Total))
    
    report.WriteString("  State Breakdown:\n")
    for state, count := range latest.StateBreakdown {
        percentage := float64(count) / float64(latest.Total) * 100
        report.WriteString(fmt.Sprintf("    %s: %d (%.1f%%)\n", state, count, percentage))
    }

    // Trend analysis
    if len(gm.history) >= 2 {
        first := gm.history[0]
        growth := float64(latest.Total) / float64(first.Total)
        duration := latest.Timestamp.Sub(first.Timestamp)
        
        report.WriteString(fmt.Sprintf("\nTrend Analysis (over %v):\n", duration))
        report.WriteString(fmt.Sprintf("  Growth: %.2fx (%d → %d)\n", growth, first.Total, latest.Total))
        
        if growth > 1.5 {
            report.WriteString("  ⚠️  SIGNIFICANT GROWTH DETECTED\n")
        }
    }

    // Top stack traces
    report.WriteString("\nTop Stack Traces:\n")
    for i, stack := range latest.TopStacks {
        report.WriteString(fmt.Sprintf("  %d. %s (%s): %d goroutines\n", 
            i+1, stack.Function, stack.State, stack.Count))
    }

    return report.String()
}
```

### Deadlock Detection

```go
package main

import (
    "context"
    "fmt"
    "log"
    "runtime/pprof"
    "sort"
    "strings"
    "time"
)

type DeadlockDetector struct {
    checkInterval  time.Duration
    stuckThreshold time.Duration
    lastStates     map[string]GoroutineState
}

type GoroutineState struct {
    ID       string
    State    string
    Stack    string
    LastSeen time.Time
}

func NewDeadlockDetector(checkInterval, stuckThreshold time.Duration) *DeadlockDetector {
    return &DeadlockDetector{
        checkInterval:  checkInterval,
        stuckThreshold: stuckThreshold,
        lastStates:     make(map[string]GoroutineState),
    }
}

func (dd *DeadlockDetector) Start(ctx context.Context) {
    ticker := time.NewTicker(dd.checkInterval)
    defer ticker.Stop()

    log.Printf("Deadlock detector started: check interval=%v, stuck threshold=%v", 
        dd.checkInterval, dd.stuckThreshold)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            dd.checkForDeadlocks()
        }
    }
}

func (dd *DeadlockDetector) checkForDeadlocks() {
    currentStates := dd.getCurrentGoroutineStates()
    now := time.Now()
    
    stuckGoroutines := []GoroutineState{}
    suspiciousPatterns := []string{}

    for id, currentState := range currentStates {
        if lastState, exists := dd.lastStates[id]; exists {
            // Check if goroutine hasn't moved
            if lastState.Stack == currentState.Stack && 
               now.Sub(lastState.LastSeen) > dd.stuckThreshold {
                
                // This goroutine appears stuck
                stuckGoroutines = append(stuckGoroutines, currentState)
            }
        }
        
        // Update last seen time
        currentState.LastSeen = now
        dd.lastStates[id] = currentState
    }

    // Analyze patterns for potential deadlocks
    suspiciousPatterns = dd.analyzeDeadlockPatterns(currentStates)

    if len(stuckGoroutines) > 0 || len(suspiciousPatterns) > 0 {
        dd.reportPotentialDeadlock(stuckGoroutines, suspiciousPatterns)
    }

    // Clean up old states
    dd.cleanupOldStates(now)
}

func (dd *DeadlockDetector) getCurrentGoroutineStates() map[string]GoroutineState {
    states := make(map[string]GoroutineState)
    
    profile := pprof.Lookup("goroutine")
    if profile == nil {
        return states
    }

    var buf strings.Builder
    profile.WriteTo(&buf, 2) // Get full stack traces
    
    lines := strings.Split(buf.String(), "\n")
    var currentID, currentState string
    var currentStack strings.Builder
    
    for _, line := range lines {
        if strings.HasPrefix(line, "goroutine") {
            // Save previous goroutine if exists
            if currentID != "" {
                states[currentID] = GoroutineState{
                    ID:    currentID,
                    State: currentState,
                    Stack: currentStack.String(),
                }
            }
            
            // Parse new goroutine
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                currentID = parts[1]
            }
            
            if start := strings.Index(line, "["); start != -1 {
                if end := strings.Index(line[start:], "]"); end != -1 {
                    currentState = line[start+1 : start+end]
                }
            }
            
            currentStack.Reset()
            currentStack.WriteString(line + "\n")
        } else if strings.TrimSpace(line) != "" {
            currentStack.WriteString(line + "\n")
        }
    }
    
    // Don't forget the last goroutine
    if currentID != "" {
        states[currentID] = GoroutineState{
            ID:    currentID,
            State: currentState,
            Stack: currentStack.String(),
        }
    }
    
    return states
}

func (dd *DeadlockDetector) analyzeDeadlockPatterns(states map[string]GoroutineState) []string {
    patterns := []string{}
    
    // Count goroutines by state and location
    stateLocations := make(map[string][]string)
    
    for _, state := range states {
        key := fmt.Sprintf("%s:%s", state.State, dd.extractLocation(state.Stack))
        stateLocations[key] = append(stateLocations[key], state.ID)
    }
    
    // Look for suspicious patterns
    for pattern, goroutines := range stateLocations {
        if len(goroutines) >= 2 && strings.Contains(pattern, "chan") {
            patterns = append(patterns, fmt.Sprintf(
                "Multiple goroutines (%d) blocked on same location: %s", 
                len(goroutines), pattern))
        }
        
        if len(goroutines) >= 3 && strings.Contains(pattern, "sync") {
            patterns = append(patterns, fmt.Sprintf(
                "Multiple goroutines (%d) waiting on sync primitive: %s", 
                len(goroutines), pattern))
        }
    }
    
    return patterns
}

func (dd *DeadlockDetector) extractLocation(stack string) string {
    lines := strings.Split(stack, "\n")
    for i, line := range lines {
        if strings.Contains(line, ".go:") && i > 0 {
            // Return the function name from previous line
            return strings.TrimSpace(lines[i-1])
        }
    }
    return "unknown"
}

func (dd *DeadlockDetector) reportPotentialDeadlock(stuckGoroutines []GoroutineState, patterns []string) {
    log.Printf("🚨 POTENTIAL DEADLOCK DETECTED 🚨")
    
    if len(stuckGoroutines) > 0 {
        log.Printf("Stuck Goroutines (%d):", len(stuckGoroutines))
        for i, g := range stuckGoroutines {
            log.Printf("  %d. Goroutine %s [%s] - stuck for %v", 
                i+1, g.ID, g.State, time.Since(g.LastSeen))
            
            // Log first few lines of stack
            stackLines := strings.Split(g.Stack, "\n")
            for j, line := range stackLines {
                if j >= 5 { // Limit output
                    break
                }
                if strings.TrimSpace(line) != "" {
                    log.Printf("    %s", line)
                }
            }
        }
    }
    
    if len(patterns) > 0 {
        log.Printf("Suspicious Patterns:")
        for i, pattern := range patterns {
            log.Printf("  %d. %s", i+1, pattern)
        }
    }
    
    // Capture detailed profile for analysis
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("deadlock_detected_%s.prof", timestamp)
    captureGoroutineProfile(filename)
    log.Printf("Detailed goroutine profile saved: %s", filename)
}

func (dd *DeadlockDetector) cleanupOldStates(now time.Time) {
    cutoff := now.Add(-5 * dd.stuckThreshold)
    for id, state := range dd.lastStates {
        if state.LastSeen.Before(cutoff) {
            delete(dd.lastStates, id)
        }
    }
}
```

## Analyzing Goroutine Profiles

### Using go tool pprof

```bash
# Interactive analysis
go tool pprof goroutine.prof

# Common pprof commands for goroutine analysis
(pprof) top           # Top functions by goroutine count
(pprof) top -cum      # Top cumulative goroutine consumers
(pprof) web           # Visual call graph
(pprof) list main     # Source code view
(pprof) traces        # Show all stack traces
(pprof) peek          # Quick overview
```

### Sample Analysis Session

```bash
$ go tool pprof goroutine.prof
File: myapp
Type: goroutine
Time: Jan 2, 2023 at 3:04pm (UTC)
Entering interactive mode (type "help" for commands, "o" for options)

(pprof) top
Showing nodes accounting for 156 goroutines, 100% of 156 total
      flat  flat%   sum%        cum   cum%
        45  28.85% 28.85%         45  28.85%  runtime.gopark
        34  21.79% 50.64%         34  21.79%  runtime.chanrecv
        28  17.95% 68.59%         28  17.95%  sync.(*Mutex).Lock
        25  16.03% 84.62%         25  16.03%  time.Sleep
        12   7.69% 92.31%         12   7.69%  net.(*conn).Read
        12   7.69%   100%         12   7.69%  runtime.selectgo

(pprof) traces
-----------+-------------------------------------------------------
      45   runtime.gopark
             runtime.chanrecv
             runtime.selectgo
             main.worker
             
      34   runtime.chanrecv
             main.processChannel
             main.pipeline
             
      28   sync.(*Mutex).Lock
             main.criticalSection
             main.worker
```

### Automated Goroutine Analysis

```go
package main

import (
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

type GoroutineAnalysis struct {
    TotalGoroutines    int
    StateBreakdown     map[string]int
    TopFunctions       []FunctionCount
    ChannelOperations  int
    MutexOperations    int
    NetworkOperations  int
    TimerOperations    int
}

type FunctionCount struct {
    Function string
    Count    int
    Percent  float64
}

func analyzeGoroutineProfile(profileFile string) (*GoroutineAnalysis, error) {
    // Get top functions
    cmd := exec.Command("go", "tool", "pprof", "-top", profileFile)
    topOutput, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get top functions: %v", err)
    }

    // Get traces for detailed analysis
    tracesCmd := exec.Command("go", "tool", "pprof", "-traces", profileFile)
    tracesOutput, err := tracesCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get traces: %v", err)
    }

    analysis := &GoroutineAnalysis{
        StateBreakdown: make(map[string]int),
        TopFunctions:   []FunctionCount{},
    }

    // Parse top output
    analysis.parseTopOutput(string(topOutput))
    
    // Parse traces output
    analysis.parseTracesOutput(string(tracesOutput))

    return analysis, nil
}

func (ga *GoroutineAnalysis) parseTopOutput(output string) {
    lines := strings.Split(output, "\n")
    
    // Extract total from first line
    for _, line := range lines {
        if strings.Contains(line, "total") {
            re := regexp.MustCompile(`(\d+) total`)
            matches := re.FindStringSubmatch(line)
            if len(matches) >= 2 {
                if total, err := strconv.Atoi(matches[1]); err == nil {
                    ga.TotalGoroutines = total
                }
            }
            break
        }
    }

    // Parse function counts
    re := regexp.MustCompile(`^\s*(\d+)\s+(\d+\.\d+)%.*?(\S+)$`)
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if len(matches) >= 4 {
            count, _ := strconv.Atoi(matches[1])
            percent, _ := strconv.ParseFloat(matches[2], 64)
            function := matches[3]
            
            ga.TopFunctions = append(ga.TopFunctions, FunctionCount{
                Function: function,
                Count:    count,
                Percent:  percent,
            })
        }
    }
}

func (ga *GoroutineAnalysis) parseTracesOutput(output string) {
    traces := strings.Split(output, "\n\n")
    
    for _, trace := range traces {
        lines := strings.Split(trace, "\n")
        if len(lines) == 0 {
            continue
        }
        
        // First line contains count
        countLine := lines[0]
        re := regexp.MustCompile(`^\s*(\d+)`)
        matches := re.FindStringSubmatch(countLine)
        if len(matches) < 2 {
            continue
        }
        
        count, _ := strconv.Atoi(matches[1])
        
        // Analyze stack trace for patterns
        stackTrace := strings.Join(lines[1:], "\n")
        ga.categorizeTrace(stackTrace, count)
    }
}

func (ga *GoroutineAnalysis) categorizeTrace(trace string, count int) {
    trace = strings.ToLower(trace)
    
    switch {
    case strings.Contains(trace, "chanrecv") || strings.Contains(trace, "chansend"):
        ga.ChannelOperations += count
        ga.StateBreakdown["channel"] += count
    case strings.Contains(trace, "mutex") || strings.Contains(trace, "rwmutex"):
        ga.MutexOperations += count
        ga.StateBreakdown["mutex"] += count
    case strings.Contains(trace, "net.") || strings.Contains(trace, "syscall"):
        ga.NetworkOperations += count
        ga.StateBreakdown["network"] += count
    case strings.Contains(trace, "time.sleep") || strings.Contains(trace, "timer"):
        ga.TimerOperations += count
        ga.StateBreakdown["timer"] += count
    case strings.Contains(trace, "runtime.gopark"):
        ga.StateBreakdown["parked"] += count
    case strings.Contains(trace, "runtime.gosched"):
        ga.StateBreakdown["yielded"] += count
    default:
        ga.StateBreakdown["other"] += count
    }
}

func (ga *GoroutineAnalysis) GenerateReport() string {
    var report strings.Builder
    
    report.WriteString("=== GOROUTINE ANALYSIS REPORT ===\n\n")
    
    report.WriteString(fmt.Sprintf("Total Goroutines: %d\n\n", ga.TotalGoroutines))
    
    // State breakdown
    report.WriteString("State Breakdown:\n")
    for state, count := range ga.StateBreakdown {
        percent := float64(count) / float64(ga.TotalGoroutines) * 100
        report.WriteString(fmt.Sprintf("  %s: %d (%.1f%%)\n", state, count, percent))
    }
    
    // Operation categories
    report.WriteString("\nOperation Categories:\n")
    report.WriteString(fmt.Sprintf("  Channel Operations: %d\n", ga.ChannelOperations))
    report.WriteString(fmt.Sprintf("  Mutex Operations: %d\n", ga.MutexOperations))
    report.WriteString(fmt.Sprintf("  Network Operations: %d\n", ga.NetworkOperations))
    report.WriteString(fmt.Sprintf("  Timer Operations: %d\n", ga.TimerOperations))
    
    // Top functions
    report.WriteString("\nTop Functions by Goroutine Count:\n")
    for i, fn := range ga.TopFunctions {
        if i >= 10 { // Top 10
            break
        }
        report.WriteString(fmt.Sprintf("  %2d. %s: %d goroutines (%.1f%%)\n", 
            i+1, fn.Function, fn.Count, fn.Percent))
    }
    
    // Health assessment
    report.WriteString("\nHealth Assessment:\n")
    report.WriteString(ga.assessHealth())
    
    return report.String()
}

func (ga *GoroutineAnalysis) assessHealth() string {
    var assessment strings.Builder
    
    // Check for excessive goroutines
    if ga.TotalGoroutines > 10000 {
        assessment.WriteString("  ⚠️  Very high goroutine count (>10k)\n")
    } else if ga.TotalGoroutines > 1000 {
        assessment.WriteString("  ⚠️  High goroutine count (>1k)\n")
    } else {
        assessment.WriteString("  ✅ Normal goroutine count\n")
    }
    
    // Check for blocking patterns
    totalBlocked := ga.ChannelOperations + ga.MutexOperations + ga.NetworkOperations
    if float64(totalBlocked)/float64(ga.TotalGoroutines) > 0.8 {
        assessment.WriteString("  ⚠️  High percentage of blocked goroutines\n")
    } else {
        assessment.WriteString("  ✅ Reasonable blocking patterns\n")
    }
    
    // Check for mutex contention
    if float64(ga.MutexOperations)/float64(ga.TotalGoroutines) > 0.3 {
        assessment.WriteString("  ⚠️  Potential mutex contention detected\n")
    }
    
    // Check for channel bottlenecks
    if float64(ga.ChannelOperations)/float64(ga.TotalGoroutines) > 0.5 {
        assessment.WriteString("  ⚠️  Many goroutines blocked on channels\n")
    }
    
    return assessment.String()
}
```

## Best Practices for Goroutine Profiling

### 1. Regular Monitoring

```go
func setupGoroutineMonitoring() {
    ctx := context.Background()
    
    // Monitor goroutine count
    monitor := NewGoroutineMonitor(30*time.Second, 1000)
    go monitor.Start(ctx)
    
    // Detect deadlocks
    detector := NewDeadlockDetector(1*time.Minute, 5*time.Minute)
    go detector.Start(ctx)
    
    // Periodic detailed analysis
    go func() {
        ticker := time.NewTicker(15 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            captureGoroutineProfile(fmt.Sprintf("periodic_%d.prof", time.Now().Unix()))
        }
    }()
}
```

### 2. Production Safety

```bash
#!/bin/bash
# safe_goroutine_profile.sh

# Check if profiling is already running
if pgrep -f "pprof.*goroutine" > /dev/null; then
    echo "Profiling already in progress, skipping"
    exit 0
fi

# Check system load before profiling
LOAD=$(uptime | awk '{print $10}' | sed 's/,//')
if (( $(echo "$LOAD > 2.0" | bc -l) )); then
    echo "System load too high ($LOAD), skipping profile"
    exit 0
fi

# Capture profile with timeout
timeout 30 curl -s http://localhost:6060/debug/pprof/goroutine > goroutine_$(date +%s).prof
echo "Goroutine profile captured safely"
```

### 3. Automated Analysis Pipeline

```bash
#!/bin/bash
# analyze_goroutines.sh

PROFILE_FILE=$1
OUTPUT_DIR="analysis_$(date +%Y%m%d_%H%M%S)"

mkdir -p $OUTPUT_DIR

# Generate various reports
go tool pprof -top $PROFILE_FILE > $OUTPUT_DIR/top_functions.txt
go tool pprof -traces $PROFILE_FILE > $OUTPUT_DIR/stack_traces.txt
go tool pprof -svg $PROFILE_FILE > $OUTPUT_DIR/goroutine_graph.svg

# Custom analysis
go run goroutine_analyzer.go $PROFILE_FILE > $OUTPUT_DIR/custom_analysis.txt

echo "Goroutine analysis completed in $OUTPUT_DIR"
```

## Next Steps

- Learn [Goroutine Analysis](goroutine-analysis.md) techniques
- Study [Deadlock Detection](deadlock-detection.md) methods
- Explore [Goroutine Leak](goroutine-leaks.md) prevention

## Summary

Goroutine profiling is essential for concurrent Go applications:

1. **Monitor goroutine counts** to detect leaks early
2. **Analyze blocking patterns** to identify bottlenecks
3. **Detect deadlocks** before they impact production
4. **Profile regularly** in development and production
5. **Automate analysis** for continuous monitoring

Use goroutine profiling to maintain healthy concurrency patterns and prevent common pitfalls in concurrent Go applications.
