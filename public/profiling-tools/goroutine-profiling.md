# Goroutine Profiling

Goroutine profiling is essential for analyzing concurrent Go applications, identifying goroutine leaks, deadlocks, and optimizing concurrent performance. This chapter covers comprehensive goroutine analysis techniques.

## Goroutine Profiling Fundamentals

### Understanding Goroutine States

Goroutines can exist in several states that affect performance:

```
┌─────────────────────────────────────────────────────────────┐
│                   Goroutine States                         │
├─────────────────────────────────────────────────────────────┤
│  State      │  Description        │  Profiling Impact      │
│             │                     │                        │
│  running    │  Executing on CPU   │  Active computation    │
│  runnable   │  Ready to run       │  Scheduler queue       │
│  waiting    │  Blocked on I/O     │  Synchronization wait  │
│  syscall    │  In system call     │  OS interaction        │
│  dead       │  Finished execution │  Cleanup pending       │
│  copystack  │  Stack growing      │  Memory management     │
└─────────────────────────────────────────────────────────────┘
```

### Basic Goroutine Profiling

```go
package main

import (
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "os"
    "runtime"
    "runtime/pprof"
    "sync"
    "time"
)

// Basic goroutine profiling setup
func basicGoroutineProfiling() {
    // Create various goroutine patterns
    createTestGoroutines()
    
    // Collect goroutine profile
    f, err := os.Create("goroutine.prof")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    // Write goroutine profile
    // Parameter 1: include all goroutines and their stacks
    if err := pprof.Lookup("goroutine").WriteTo(f, 1); err != nil {
        panic(err)
    }
    
    fmt.Printf("Goroutine profile written to goroutine.prof\n")
    fmt.Printf("Active goroutines: %d\n", runtime.NumGoroutine())
    fmt.Printf("Analyze with: go tool pprof goroutine.prof\n")
}

func createTestGoroutines() {
    var wg sync.WaitGroup
    
    // CPU-bound goroutines
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cpuBoundWork(id)
        }(i)
    }
    
    // I/O-bound goroutines
    for i := 0; i < 10; i++ {
        go func(id int) {
            ioBoundWork(id)
        }(i)
    }
    
    // Channel-based goroutines
    ch := make(chan int, 5)
    for i := 0; i < 3; i++ {
        go producer(ch, i)
        go consumer(ch, i)
    }
    
    // Wait for CPU-bound work to complete
    wg.Wait()
    
    // Give other goroutines time to establish patterns
    time.Sleep(2 * time.Second)
}

func cpuBoundWork(id int) {
    for i := 0; i < 1000000; i++ {
        _ = i * i
    }
    fmt.Printf("CPU worker %d completed\n", id)
}

func ioBoundWork(id int) {
    for {
        time.Sleep(100 * time.Millisecond)
        // Simulate I/O work
        if id%100 == 0 {
            fmt.Printf("I/O worker %d tick\n", id)
        }
    }
}

func producer(ch chan<- int, id int) {
    for i := 0; ; i++ {
        select {
        case ch <- i:
            time.Sleep(50 * time.Millisecond)
        case <-time.After(time.Second):
            // Timeout to prevent deadlock in demo
        }
    }
}

func consumer(ch <-chan int, id int) {
    for {
        select {
        case val := <-ch:
            // Process value
            time.Sleep(time.Duration(val%10) * time.Millisecond)
        case <-time.After(time.Second):
            // Timeout to prevent deadlock in demo
        }
    }
}
```

### HTTP Endpoint Goroutine Profiling

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    _ "net/http/pprof"
    "sync"
    "time"
)

func goroutineProfilingServer() {
    // Start application with various goroutine patterns
    go runConcurrentApplication()
    
    fmt.Println("Goroutine profiling server running on :6060")
    fmt.Println("Available endpoints:")
    fmt.Println("  /debug/pprof/goroutine     - Current goroutine profile")
    fmt.Println("  /debug/pprof/goroutine?debug=1 - Human-readable goroutine dump")
    fmt.Println("  /debug/pprof/goroutine?debug=2 - Full goroutine stacks")
    fmt.Println()
    fmt.Println("Usage examples:")
    fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
    fmt.Println("  curl http://localhost:6060/debug/pprof/goroutine?debug=2")
    
    log.Fatal(http.ListenAndServe("localhost:6060", nil))
}

func runConcurrentApplication() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Worker pool pattern
    startWorkerPool(ctx, 10)
    
    // Pipeline pattern
    startPipeline(ctx)
    
    // Fan-out/fan-in pattern
    startFanOutFanIn(ctx)
    
    // Timer-based workers
    startTimerWorkers(ctx)
    
    // Keep application running
    select {}
}

func startWorkerPool(ctx context.Context, workers int) {
    jobs := make(chan Job, 100)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            workerPoolWorker(ctx, id, jobs)
        }(i)
    }
    
    // Job generator
    go func() {
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()
        
        jobID := 0
        for {
            select {
            case <-ctx.Done():
                close(jobs)
                return
            case <-ticker.C:
                select {
                case jobs <- Job{ID: jobID, Data: fmt.Sprintf("job-%d", jobID)}:
                    jobID++
                default:
                    // Channel full, skip
                }
            }
        }
    }()
}

type Job struct {
    ID   int
    Data string
}

func workerPoolWorker(ctx context.Context, id int, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            
            // Simulate work
            processJob(job)
            
            if id == 0 && job.ID%100 == 0 {
                fmt.Printf("Worker pool processed job %d\n", job.ID)
            }
        }
    }
}

func processJob(job Job) {
    // Simulate varying work time
    workTime := time.Duration(job.ID%10+1) * time.Millisecond
    time.Sleep(workTime)
}

func startPipeline(ctx context.Context) {
    // Stage 1: Data generation
    stage1 := make(chan Data, 10)
    go func() {
        defer close(stage1)
        dataID := 0
        ticker := time.NewTicker(50 * time.Millisecond)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                select {
                case stage1 <- Data{ID: dataID, Value: dataID * 2}:
                    dataID++
                default:
                }
            }
        }
    }()
    
    // Stage 2: Data processing
    stage2 := make(chan Data, 10)
    go func() {
        defer close(stage2)
        for data := range stage1 {
            // Process data
            data.Value = data.Value * 3
            time.Sleep(2 * time.Millisecond)
            
            select {
            case stage2 <- data:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // Stage 3: Data output
    go func() {
        count := 0
        for data := range stage2 {
            // Output data
            count++
            if count%100 == 0 {
                fmt.Printf("Pipeline processed %d items, latest: %+v\n", count, data)
            }
            time.Sleep(time.Millisecond)
        }
    }()
}

type Data struct {
    ID    int
    Value int
}

func startFanOutFanIn(ctx context.Context) {
    input := make(chan int, 5)
    
    // Input generator
    go func() {
        defer close(input)
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return
            case input <- i:
                time.Sleep(20 * time.Millisecond)
            }
        }
    }()
    
    // Fan-out: distribute work to multiple processors
    const processors = 5
    outputs := make([]chan int, processors)
    
    for i := 0; i < processors; i++ {
        outputs[i] = make(chan int, 5)
        go func(id int, output chan<- int) {
            defer close(output)
            for value := range input {
                // Process value
                result := value * value
                time.Sleep(5 * time.Millisecond)
                
                select {
                case output <- result:
                case <-ctx.Done():
                    return
                }
            }
        }(i, outputs[i])
    }
    
    // Fan-in: collect results
    result := make(chan int, 10)
    var wg sync.WaitGroup
    
    for i, output := range outputs {
        wg.Add(1)
        go func(id int, output <-chan int) {
            defer wg.Done()
            for value := range output {
                select {
                case result <- value:
                case <-ctx.Done():
                    return
                }
            }
        }(i, output)
    }
    
    go func() {
        wg.Wait()
        close(result)
    }()
    
    // Result consumer
    go func() {
        count := 0
        for value := range result {
            count++
            if count%50 == 0 {
                fmt.Printf("Fan-out/fan-in processed %d results, latest: %d\n", count, value)
            }
        }
    }()
}

func startTimerWorkers(ctx context.Context) {
    // Different timer patterns
    timers := []time.Duration{
        100 * time.Millisecond,
        250 * time.Millisecond,
        500 * time.Millisecond,
        time.Second,
    }
    
    for i, interval := range timers {
        go func(id int, interval time.Duration) {
            ticker := time.NewTicker(interval)
            defer ticker.Stop()
            
            count := 0
            for {
                select {
                case <-ctx.Done():
                    return
                case <-ticker.C:
                    count++
                    if count%10 == 0 {
                        fmt.Printf("Timer worker %d (every %v) tick %d\n", id, interval, count)
                    }
                }
            }
        }(i, interval)
    }
}
```

## Goroutine Analysis Techniques

### Programmatic Goroutine Analysis

```go
package main

import (
    "bytes"
    "fmt"
    "runtime"
    "runtime/pprof"
    "sort"
    "strings"
)

type GoroutineAnalysis struct {
    TotalGoroutines int
    States          map[string]int
    Functions       map[string]int
    Stacks          []StackInfo
    LeakedGoroutines []LeakInfo
}

type StackInfo struct {
    Count    int
    State    string
    Function string
    Stack    []string
}

type LeakInfo struct {
    Function string
    Count    int
    State    string
    Duration string
}

func analyzeGoroutines() GoroutineAnalysis {
    // Collect goroutine profile
    var buf bytes.Buffer
    if err := pprof.Lookup("goroutine").WriteTo(&buf, 1); err != nil {
        panic(err)
    }
    
    // Parse the profile data
    profile := buf.String()
    return parseGoroutineProfile(profile)
}

func parseGoroutineProfile(profile string) GoroutineAnalysis {
    lines := strings.Split(profile, "\n")
    
    analysis := GoroutineAnalysis{
        States:    make(map[string]int),
        Functions: make(map[string]int),
        Stacks:    make([]StackInfo, 0),
    }
    
    var currentStack StackInfo
    var stackLines []string
    inStack := false
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        if line == "" {
            if inStack && len(stackLines) > 0 {
                currentStack.Stack = make([]string, len(stackLines))
                copy(currentStack.Stack, stackLines)
                analysis.Stacks = append(analysis.Stacks, currentStack)
                
                // Update function count
                if currentStack.Function != "" {
                    analysis.Functions[currentStack.Function]++
                }
                
                // Reset for next stack
                currentStack = StackInfo{}
                stackLines = stackLines[:0]
                inStack = false
            }
            continue
        }
        
        // Parse goroutine header: "goroutine 123 [running]:"
        if strings.HasPrefix(line, "goroutine ") && strings.HasSuffix(line, ":") {
            parts := strings.Fields(line)
            if len(parts) >= 3 {
                state := strings.Trim(parts[2], "[]:")
                currentStack.State = state
                currentStack.Count = 1
                analysis.States[state]++
                analysis.TotalGoroutines++
                inStack = true
            }
            continue
        }
        
        if inStack {
            // Parse function calls in stack
            if strings.Contains(line, "(") && !strings.HasPrefix(line, "\t") {
                // Function name line
                funcName := extractFunctionName(line)
                if currentStack.Function == "" {
                    currentStack.Function = funcName
                }
                stackLines = append(stackLines, line)
            } else if strings.HasPrefix(line, "\t") {
                // Source file line
                stackLines = append(stackLines, line)
            }
        }
    }
    
    // Process final stack if exists
    if inStack && len(stackLines) > 0 {
        currentStack.Stack = stackLines
        analysis.Stacks = append(analysis.Stacks, currentStack)
        if currentStack.Function != "" {
            analysis.Functions[currentStack.Function]++
        }
    }
    
    // Detect potential leaks
    analysis.LeakedGoroutines = detectGoroutineLeaks(analysis)
    
    return analysis
}

func extractFunctionName(line string) string {
    // Extract function name from line like "main.worker(0xc000010000)"
    if idx := strings.Index(line, "("); idx > 0 {
        return line[:idx]
    }
    return line
}

func detectGoroutineLeaks(analysis GoroutineAnalysis) []LeakInfo {
    var leaks []LeakInfo
    
    // Heuristics for detecting leaks:
    // 1. Many goroutines in same function
    // 2. Goroutines waiting on channels
    // 3. Goroutines in select statements
    
    for function, count := range analysis.Functions {
        if count > 10 { // Threshold for potential leak
            // Find the state of these goroutines
            state := findMostCommonState(analysis.Stacks, function)
            
            leaks = append(leaks, LeakInfo{
                Function: function,
                Count:    count,
                State:    state,
                Duration: "unknown", // Would need timestamp tracking for duration
            })
        }
    }
    
    return leaks
}

func findMostCommonState(stacks []StackInfo, function string) string {
    stateCounts := make(map[string]int)
    
    for _, stack := range stacks {
        if stack.Function == function {
            stateCounts[stack.State]++
        }
    }
    
    maxCount := 0
    mostCommonState := "unknown"
    for state, count := range stateCounts {
        if count > maxCount {
            maxCount = count
            mostCommonState = state
        }
    }
    
    return mostCommonState
}

func printGoroutineAnalysis(analysis GoroutineAnalysis) {
    fmt.Printf("=== Goroutine Analysis ===\n")
    fmt.Printf("Total Goroutines: %d\n", analysis.TotalGoroutines)
    fmt.Printf("\n")
    
    // Print state distribution
    fmt.Printf("Goroutine States:\n")
    for state, count := range analysis.States {
        percentage := float64(count) / float64(analysis.TotalGoroutines) * 100
        fmt.Printf("  %-15s: %4d (%.1f%%)\n", state, count, percentage)
    }
    fmt.Printf("\n")
    
    // Print top functions by goroutine count
    fmt.Printf("Top Functions by Goroutine Count:\n")
    
    type funcCount struct {
        Function string
        Count    int
    }
    
    var funcs []funcCount
    for function, count := range analysis.Functions {
        funcs = append(funcs, funcCount{function, count})
    }
    
    sort.Slice(funcs, func(i, j int) bool {
        return funcs[i].Count > funcs[j].Count
    })
    
    for i, fc := range funcs {
        if i >= 10 { // Show top 10
            break
        }
        percentage := float64(fc.Count) / float64(analysis.TotalGoroutines) * 100
        fmt.Printf("  %-40s: %4d (%.1f%%)\n", fc.Function, fc.Count, percentage)
    }
    fmt.Printf("\n")
    
    // Print potential leaks
    if len(analysis.LeakedGoroutines) > 0 {
        fmt.Printf("Potential Goroutine Leaks:\n")
        for _, leak := range analysis.LeakedGoroutines {
            fmt.Printf("  %-40s: %4d goroutines [%s]\n", 
                leak.Function, leak.Count, leak.State)
        }
        fmt.Printf("\n")
    }
    
    // Print sample stacks for analysis
    fmt.Printf("Sample Goroutine Stacks:\n")
    printed := make(map[string]bool)
    count := 0
    
    for _, stack := range analysis.Stacks {
        if count >= 5 { // Show 5 samples
            break
        }
        
        if !printed[stack.Function] {
            fmt.Printf("\n  Function: %s [%s]\n", stack.Function, stack.State)
            for i, line := range stack.Stack {
                if i >= 8 { // Show first 8 lines of stack
                    fmt.Printf("    ...\n")
                    break
                }
                fmt.Printf("    %s\n", line)
            }
            printed[stack.Function] = true
            count++
        }
    }
}
```

### Goroutine Leak Detection

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
)

// Goroutine leak detector
type GoroutineLeakDetector struct {
    mu               sync.Mutex
    baseline         int
    threshold        int
    checkInterval    time.Duration
    alertCallback    func(int, int)
    running          bool
    stopCh           chan struct{}
}

func NewGoroutineLeakDetector(threshold int) *GoroutineLeakDetector {
    return &GoroutineLeakDetector{
        threshold:     threshold,
        checkInterval: 30 * time.Second,
        stopCh:        make(chan struct{}),
        alertCallback: defaultGoroutineAlertCallback,
    }
}

func (gld *GoroutineLeakDetector) SetAlertCallback(callback func(int, int)) {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    gld.alertCallback = callback
}

func (gld *GoroutineLeakDetector) SetBaseline() {
    gld.mu.Lock()
    defer gld.mu.Unlock()
    
    gld.baseline = runtime.NumGoroutine()
    fmt.Printf("Goroutine leak detector baseline set: %d goroutines\n", gld.baseline)
}

func (gld *GoroutineLeakDetector) Start() {
    gld.mu.Lock()
    if gld.running {
        gld.mu.Unlock()
        return
    }
    gld.running = true
    gld.mu.Unlock()
    
    go gld.monitorLoop()
}

func (gld *GoroutineLeakDetector) Stop() {
    gld.mu.Lock()
    if !gld.running {
        gld.mu.Unlock()
        return
    }
    gld.running = false
    gld.mu.Unlock()
    
    close(gld.stopCh)
}

func (gld *GoroutineLeakDetector) monitorLoop() {
    ticker := time.NewTicker(gld.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            gld.checkForLeaks()
        case <-gld.stopCh:
            return
        }
    }
}

func (gld *GoroutineLeakDetector) checkForLeaks() {
    gld.mu.Lock()
    baseline := gld.baseline
    threshold := gld.threshold
    callback := gld.alertCallback
    gld.mu.Unlock()
    
    current := runtime.NumGoroutine()
    
    if current-baseline > threshold {
        if callback != nil {
            callback(baseline, current)
        }
    }
}

func defaultGoroutineAlertCallback(baseline, current int) {
    growth := current - baseline
    
    fmt.Printf("=== GOROUTINE LEAK ALERT ===\n")
    fmt.Printf("Baseline: %d goroutines\n", baseline)
    fmt.Printf("Current: %d goroutines\n", current)
    fmt.Printf("Growth: +%d goroutines\n", growth)
    fmt.Printf("Timestamp: %v\n", time.Now().Format(time.RFC3339))
    
    // Dump current goroutine state for analysis
    analysis := analyzeGoroutines()
    fmt.Printf("Top functions:\n")
    
    type funcCount struct {
        Function string
        Count    int
    }
    
    var funcs []funcCount
    for function, count := range analysis.Functions {
        funcs = append(funcs, funcCount{function, count})
    }
    
    sort.Slice(funcs, func(i, j int) bool {
        return funcs[i].Count > funcs[j].Count
    })
    
    for i, fc := range funcs {
        if i >= 5 { // Show top 5
            break
        }
        fmt.Printf("  %s: %d goroutines\n", fc.Function, fc.Count)
    }
    
    fmt.Printf("===========================\n")
}

// Example: Detecting different types of goroutine leaks
func demonstrateGoroutineLeaks() {
    detector := NewGoroutineLeakDetector(5)
    detector.Start()
    defer detector.Stop()
    
    // Set baseline after initialization
    time.Sleep(time.Second)
    detector.SetBaseline()
    
    fmt.Println("Demonstrating goroutine leaks...")
    
    // Leak 1: Goroutines waiting on channels that never receive data
    fmt.Println("Creating channel leak...")
    createChannelLeak(10)
    time.Sleep(2 * time.Second)
    
    // Leak 2: Goroutines in infinite loops without proper cancellation
    fmt.Println("Creating infinite loop leak...")
    createInfiniteLoopLeak(5)
    time.Sleep(2 * time.Second)
    
    // Leak 3: Worker goroutines that don't exit when work is done
    fmt.Println("Creating worker leak...")
    createWorkerLeak(8)
    time.Sleep(2 * time.Second)
    
    fmt.Printf("Final goroutine count: %d\n", runtime.NumGoroutine())
}

func createChannelLeak(count int) {
    ch := make(chan int)
    
    for i := 0; i < count; i++ {
        go func(id int) {
            // This will block forever - classic channel leak
            val := <-ch
            fmt.Printf("Goroutine %d received: %d\n", id, val)
        }(i)
    }
}

func createInfiniteLoopLeak(count int) {
    for i := 0; i < count; i++ {
        go func(id int) {
            // Infinite loop without cancellation - CPU leak
            for {
                time.Sleep(100 * time.Millisecond)
                // No way to exit this loop
            }
        }(i)
    }
}

func createWorkerLeak(count int) {
    jobs := make(chan Job, 100)
    
    // Start workers but don't provide exit mechanism
    for i := 0; i < count; i++ {
        go func(id int) {
            for job := range jobs {
                processJob(job)
            }
            // Goroutine exits when channel closes, but we never close it
        }(i)
    }
    
    // Send some jobs then stop, leaving workers hanging
    for i := 0; i < 5; i++ {
        jobs <- Job{ID: i, Data: fmt.Sprintf("job-%d", i)}
    }
}
```

### Advanced Goroutine Monitoring

```go
// Comprehensive goroutine monitoring system
type GoroutineMonitor struct {
    mu          sync.RWMutex
    samples     []GoroutineSample
    maxSamples  int
    interval    time.Duration
    running     bool
    stopCh      chan struct{}
    alerts      []AlertRule
}

type GoroutineSample struct {
    Timestamp   time.Time
    Count       int
    States      map[string]int
    TopFunctions map[string]int
}

type AlertRule struct {
    Name      string
    Condition func(GoroutineSample) bool
    Action    func(GoroutineSample)
}

func NewGoroutineMonitor() *GoroutineMonitor {
    return &GoroutineMonitor{
        samples:    make([]GoroutineSample, 0),
        maxSamples: 1000, // Keep last 1000 samples
        interval:   10 * time.Second,
        stopCh:     make(chan struct{}),
        alerts:     make([]AlertRule, 0),
    }
}

func (gm *GoroutineMonitor) AddAlert(rule AlertRule) {
    gm.mu.Lock()
    defer gm.mu.Unlock()
    gm.alerts = append(gm.alerts, rule)
}

func (gm *GoroutineMonitor) Start() {
    gm.mu.Lock()
    if gm.running {
        gm.mu.Unlock()
        return
    }
    gm.running = true
    gm.mu.Unlock()
    
    go gm.monitorLoop()
}

func (gm *GoroutineMonitor) Stop() {
    gm.mu.Lock()
    if !gm.running {
        gm.mu.Unlock()
        return
    }
    gm.running = false
    gm.mu.Unlock()
    
    close(gm.stopCh)
}

func (gm *GoroutineMonitor) monitorLoop() {
    ticker := time.NewTicker(gm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            sample := gm.collectSample()
            gm.addSample(sample)
            gm.checkAlerts(sample)
        case <-gm.stopCh:
            return
        }
    }
}

func (gm *GoroutineMonitor) collectSample() GoroutineSample {
    analysis := analyzeGoroutines()
    
    // Extract top 10 functions
    topFunctions := make(map[string]int)
    
    type funcCount struct {
        Function string
        Count    int
    }
    
    var funcs []funcCount
    for function, count := range analysis.Functions {
        funcs = append(funcs, funcCount{function, count})
    }
    
    sort.Slice(funcs, func(i, j int) bool {
        return funcs[i].Count > funcs[j].Count
    })
    
    for i, fc := range funcs {
        if i >= 10 { // Top 10
            break
        }
        topFunctions[fc.Function] = fc.Count
    }
    
    return GoroutineSample{
        Timestamp:    time.Now(),
        Count:        analysis.TotalGoroutines,
        States:       analysis.States,
        TopFunctions: topFunctions,
    }
}

func (gm *GoroutineMonitor) addSample(sample GoroutineSample) {
    gm.mu.Lock()
    defer gm.mu.Unlock()
    
    gm.samples = append(gm.samples, sample)
    
    // Keep only recent samples
    if len(gm.samples) > gm.maxSamples {
        gm.samples = gm.samples[1:]
    }
}

func (gm *GoroutineMonitor) checkAlerts(sample GoroutineSample) {
    gm.mu.RLock()
    alerts := gm.alerts
    gm.mu.RUnlock()
    
    for _, alert := range alerts {
        if alert.Condition(sample) {
            go alert.Action(sample) // Run alert action in separate goroutine
        }
    }
}

func (gm *GoroutineMonitor) GetTrend(duration time.Duration) []GoroutineSample {
    gm.mu.RLock()
    defer gm.mu.RUnlock()
    
    cutoff := time.Now().Add(-duration)
    var result []GoroutineSample
    
    for _, sample := range gm.samples {
        if sample.Timestamp.After(cutoff) {
            result = append(result, sample)
        }
    }
    
    return result
}

func (gm *GoroutineMonitor) PrintReport() {
    gm.mu.RLock()
    defer gm.mu.RUnlock()
    
    if len(gm.samples) == 0 {
        fmt.Println("No goroutine samples collected")
        return
    }
    
    latest := gm.samples[len(gm.samples)-1]
    
    fmt.Printf("=== Goroutine Monitor Report ===\n")
    fmt.Printf("Latest Sample: %v\n", latest.Timestamp.Format(time.RFC3339))
    fmt.Printf("Goroutine Count: %d\n", latest.Count)
    fmt.Printf("Sample History: %d samples\n", len(gm.samples))
    fmt.Printf("\n")
    
    // Show trend over last hour
    trend := gm.GetTrend(time.Hour)
    if len(trend) > 1 {
        first := trend[0]
        last := trend[len(trend)-1]
        growth := last.Count - first.Count
        
        fmt.Printf("Trend (last hour):\n")
        fmt.Printf("  Start: %d goroutines\n", first.Count)
        fmt.Printf("  End: %d goroutines\n", last.Count)
        fmt.Printf("  Growth: %+d goroutines\n", growth)
        fmt.Printf("\n")
    }
    
    // Show current state distribution
    fmt.Printf("Current States:\n")
    for state, count := range latest.States {
        percentage := float64(count) / float64(latest.Count) * 100
        fmt.Printf("  %-15s: %4d (%.1f%%)\n", state, count, percentage)
    }
    fmt.Printf("\n")
    
    // Show top functions
    fmt.Printf("Top Functions:\n")
    for function, count := range latest.TopFunctions {
        percentage := float64(count) / float64(latest.Count) * 100
        fmt.Printf("  %-40s: %4d (%.1f%%)\n", function, count, percentage)
    }
}

// Example usage with custom alerts
func setupGoroutineMonitoring() {
    monitor := NewGoroutineMonitor()
    
    // Alert 1: High goroutine count
    monitor.AddAlert(AlertRule{
        Name: "high_goroutine_count",
        Condition: func(sample GoroutineSample) bool {
            return sample.Count > 1000
        },
        Action: func(sample GoroutineSample) {
            fmt.Printf("ALERT: High goroutine count: %d at %v\n", 
                sample.Count, sample.Timestamp.Format(time.RFC3339))
        },
    })
    
    // Alert 2: Rapid goroutine growth
    monitor.AddAlert(AlertRule{
        Name: "rapid_growth",
        Condition: func(sample GoroutineSample) bool {
            trend := monitor.GetTrend(5 * time.Minute)
            if len(trend) < 2 {
                return false
            }
            
            growth := sample.Count - trend[0].Count
            return growth > 100 // More than 100 goroutines in 5 minutes
        },
        Action: func(sample GoroutineSample) {
            fmt.Printf("ALERT: Rapid goroutine growth detected at %v\n", 
                sample.Timestamp.Format(time.RFC3339))
        },
    })
    
    // Alert 3: Too many waiting goroutines
    monitor.AddAlert(AlertRule{
        Name: "too_many_waiting",
        Condition: func(sample GoroutineSample) bool {
            waiting := sample.States["chan receive"] + sample.States["chan send"] + sample.States["select"]
            return waiting > sample.Count/2 // More than 50% waiting
        },
        Action: func(sample GoroutineSample) {
            fmt.Printf("ALERT: Too many waiting goroutines at %v\n", 
                sample.Timestamp.Format(time.RFC3339))
        },
    })
    
    monitor.Start()
    
    // Print periodic reports
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            monitor.PrintReport()
        }
    }()
    
    return monitor
}
```

## Goroutine Performance Optimization

### Efficient Concurrency Patterns

```go
// Optimized worker pool with controlled lifecycle
type OptimizedWorkerPool struct {
    workers     int
    jobQueue    chan Job
    workerQueue chan chan Job
    quit        chan bool
    wg          sync.WaitGroup
}

func NewOptimizedWorkerPool(workers, queueSize int) *OptimizedWorkerPool {
    pool := &OptimizedWorkerPool{
        workers:     workers,
        jobQueue:    make(chan Job, queueSize),
        workerQueue: make(chan chan Job, workers),
        quit:        make(chan bool),
    }
    
    pool.start()
    return pool
}

func (p *OptimizedWorkerPool) start() {
    // Start workers
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }
    
    // Start dispatcher
    go p.dispatch()
}

func (p *OptimizedWorkerPool) worker(id int) {
    defer p.wg.Done()
    
    // Create worker's job queue
    jobQueue := make(chan Job)
    
    for {
        // Register worker as available
        select {
        case p.workerQueue <- jobQueue:
            // Worker is now available
        case <-p.quit:
            return
        }
        
        // Wait for job
        select {
        case job := <-jobQueue:
            // Process job
            processJob(job)
            
        case <-p.quit:
            return
        }
    }
}

func (p *OptimizedWorkerPool) dispatch() {
    for {
        select {
        case job := <-p.jobQueue:
            // Get available worker
            select {
            case workerQueue := <-p.workerQueue:
                // Send job to worker
                select {
                case workerQueue <- job:
                case <-p.quit:
                    return
                }
            case <-p.quit:
                return
            }
            
        case <-p.quit:
            return
        }
    }
}

func (p *OptimizedWorkerPool) Submit(job Job) {
    select {
    case p.jobQueue <- job:
    default:
        // Job queue full, handle as appropriate
        fmt.Printf("Job queue full, dropping job %d\n", job.ID)
    }
}

func (p *OptimizedWorkerPool) Stop() {
    close(p.quit)
    p.wg.Wait()
}

// Context-aware goroutine management
func contextAwareGoroutines(ctx context.Context) {
    var wg sync.WaitGroup
    
    // Start multiple workers with context cancellation
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            ticker := time.NewTicker(100 * time.Millisecond)
            defer ticker.Stop()
            
            for {
                select {
                case <-ctx.Done():
                    fmt.Printf("Worker %d shutting down: %v\n", id, ctx.Err())
                    return
                case <-ticker.C:
                    // Do work
                    doWork()
                }
            }
        }(i)
    }
    
    // Wait for all workers to complete
    wg.Wait()
}

// Batched processing to reduce goroutine overhead
func batchedProcessing(items <-chan Item) {
    const batchSize = 100
    const maxWait = 100 * time.Millisecond
    
    batch := make([]Item, 0, batchSize)
    timer := time.NewTimer(maxWait)
    
    for {
        select {
        case item, ok := <-items:
            if !ok {
                // Channel closed, process remaining batch
                if len(batch) > 0 {
                    processBatch(batch)
                }
                return
            }
            
            batch = append(batch, item)
            
            if len(batch) >= batchSize {
                processBatch(batch)
                batch = batch[:0] // Reset, keep capacity
                timer.Reset(maxWait)
            }
            
        case <-timer.C:
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0]
            }
            timer.Reset(maxWait)
        }
    }
}
```

Goroutine profiling is essential for understanding and optimizing concurrent Go applications. Master these techniques to build efficient, leak-free concurrent systems.

---

**Next**: [Execution Tracing](execution-tracing.md) - Comprehensive execution flow analysis
