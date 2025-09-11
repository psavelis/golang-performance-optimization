# Execution Tracing

Execution tracing provides the most comprehensive view of Go program behavior, capturing detailed information about goroutine scheduling, garbage collection, and system calls. This chapter covers advanced tracing techniques and analysis methods.

## Execution Tracing Fundamentals

### What Execution Tracing Captures

Go's execution tracer records detailed timeline information:

```
┌─────────────────────────────────────────────────────────────┐
│                   Execution Trace Data                     │
├─────────────────────────────────────────────────────────────┤
│  Event Type     │  Information Captured   │  Use Case      │
│                 │                         │                │
│  Goroutine      │  Creation, scheduling   │  Concurrency   │
│  Processor      │  P state changes        │  Load balance  │
│  GC             │  Collection phases      │  Memory tuning │
│  System Call    │  Syscall entry/exit     │  I/O analysis  │
│  Network        │  Network operations     │  Latency debug │
│  User Events    │  Custom annotations     │  App profiling │
└─────────────────────────────────────────────────────────────┘
```

### Basic Execution Tracing

```go
package main

import (
    "context"
    "fmt"
    "os"
    "runtime/trace"
    "sync"
    "time"
)

// Basic execution tracing setup
func basicExecutionTracing() {
    // Create trace file
    f, err := os.Create("trace.out")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    // Start tracing
    if err := trace.Start(f); err != nil {
        panic(err)
    }
    defer trace.Stop()
    
    fmt.Println("Execution tracing started")
    
    // Run application with various patterns
    runTracedApplication()
    
    fmt.Println("Execution tracing completed")
    fmt.Println("Analyze with: go tool trace trace.out")
}

func runTracedApplication() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    var wg sync.WaitGroup
    
    // CPU-bound work
    wg.Add(1)
    go func() {
        defer wg.Done()
        cpuBoundWork(ctx)
    }()
    
    // I/O-bound work
    wg.Add(1)
    go func() {
        defer wg.Done()
        ioBoundWork(ctx)
    }()
    
    // Channel-based communication
    wg.Add(1)
    go func() {
        defer wg.Done()
        channelWork(ctx)
    }()
    
    // Memory allocation patterns
    wg.Add(1)
    go func() {
        defer wg.Done()
        memoryWork(ctx)
    }()
    
    wg.Wait()
}

func cpuBoundWork(ctx context.Context) {
    for i := 0; i < 1000000; i++ {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // CPU-intensive computation
        result := 0
        for j := 0; j < 1000; j++ {
            result += j * j
        }
        
        // Yield occasionally
        if i%10000 == 0 {
            fmt.Printf("CPU work progress: %d\n", i)
        }
    }
}

func ioBoundWork(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Millisecond)
    defer ticker.Stop()
    
    count := 0
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Simulate I/O work
            time.Sleep(time.Millisecond)
            count++
            
            if count%100 == 0 {
                fmt.Printf("I/O work progress: %d\n", count)
            }
        }
    }
}

func channelWork(ctx context.Context) {
    ch := make(chan int, 10)
    
    // Producer
    go func() {
        defer close(ch)
        for i := 0; i < 1000; i++ {
            select {
            case <-ctx.Done():
                return
            case ch <- i:
                time.Sleep(time.Millisecond)
            }
        }
    }()
    
    // Consumer
    for {
        select {
        case <-ctx.Done():
            return
        case val, ok := <-ch:
            if !ok {
                return
            }
            // Process value
            _ = val * 2
        }
    }
}

func memoryWork(ctx context.Context) {
    var data [][]byte
    
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Allocate memory
            chunk := make([]byte, 64*1024) // 64KB chunks
            data = append(data, chunk)
            
            // Occasionally clean up to trigger GC
            if len(data) > 100 {
                data = data[:0] // Clear references
            }
        }
    }
}
```

### HTTP Endpoint Tracing

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "runtime/trace"
    "time"
)

func tracingHTTPServer() {
    // Handler for starting/stopping traces
    http.HandleFunc("/trace", traceHandler)
    
    // Application endpoints
    http.HandleFunc("/cpu", cpuHandler)
    http.HandleFunc("/memory", memoryHandler)
    http.HandleFunc("/goroutines", goroutineHandler)
    http.HandleFunc("/io", ioHandler)
    
    fmt.Println("Tracing HTTP server running on :8080")
    fmt.Println("Available endpoints:")
    fmt.Println("  /trace         - Start/stop execution tracing")
    fmt.Println("  /cpu           - CPU-intensive endpoint")
    fmt.Println("  /memory        - Memory-intensive endpoint")
    fmt.Println("  /goroutines    - Goroutine-heavy endpoint")
    fmt.Println("  /io            - I/O-intensive endpoint")
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}

var (
    traceFile *os.File
    tracing   bool
)

func traceHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        if tracing {
            http.Error(w, "Tracing already active", http.StatusConflict)
            return
        }
        
        // Start tracing
        timestamp := time.Now().Format("20060102_150405")
        filename := fmt.Sprintf("trace_%s.out", timestamp)
        
        f, err := os.Create(filename)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to create trace file: %v", err), 
                http.StatusInternalServerError)
            return
        }
        
        if err := trace.Start(f); err != nil {
            f.Close()
            http.Error(w, fmt.Sprintf("Failed to start tracing: %v", err), 
                http.StatusInternalServerError)
            return
        }
        
        traceFile = f
        tracing = true
        
        fmt.Fprintf(w, "Tracing started: %s\n", filename)
        
    case "DELETE":
        if !tracing {
            http.Error(w, "Tracing not active", http.StatusConflict)
            return
        }
        
        // Stop tracing
        trace.Stop()
        traceFile.Close()
        tracing = false
        
        fmt.Fprintf(w, "Tracing stopped\n")
        
    default:
        if tracing {
            fmt.Fprintf(w, "Tracing active\n")
        } else {
            fmt.Fprintf(w, "Tracing inactive\n")
        }
    }
}

func cpuHandler(w http.ResponseWriter, r *http.Request) {
    // CPU-intensive work
    result := 0
    for i := 0; i < 1000000; i++ {
        result += i * i
    }
    
    fmt.Fprintf(w, "CPU work completed, result: %d\n", result)
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
    // Memory-intensive work
    data := make([][]byte, 1000)
    for i := range data {
        data[i] = make([]byte, 1024*10) // 10KB per chunk
        
        // Fill with data
        for j := range data[i] {
            data[i][j] = byte(i % 256)
        }
    }
    
    total := len(data) * len(data[0])
    fmt.Fprintf(w, "Memory work completed, allocated: %d bytes\n", total)
}

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
    // Create many short-lived goroutines
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Short work
            time.Sleep(time.Duration(id%10) * time.Millisecond)
        }(i)
    }
    
    wg.Wait()
    fmt.Fprintf(w, "Goroutine work completed\n")
}

func ioHandler(w http.ResponseWriter, r *http.Request) {
    // I/O simulation
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Simulate I/O
            time.Sleep(time.Duration(10+id) * time.Millisecond)
        }(i)
    }
    
    wg.Wait()
    fmt.Fprintf(w, "I/O work completed\n")
}
```

## Advanced Tracing Techniques

### Custom Trace Events

```go
package main

import (
    "context"
    "runtime/trace"
    "time"
)

// Custom trace regions for application profiling
func customTraceEvents() {
    f, err := os.Create("custom_trace.out")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    if err := trace.Start(f); err != nil {
        panic(err)
    }
    defer trace.Stop()
    
    // Application with custom trace regions
    runApplicationWithTracing()
}

func runApplicationWithTracing() {
    ctx := context.Background()
    
    // Database operations
    traceDatabaseOperations(ctx)
    
    // Business logic
    traceBusinessLogic(ctx)
    
    // External API calls
    traceExternalAPIs(ctx)
    
    // Background processing
    traceBackgroundTasks(ctx)
}

func traceDatabaseOperations(ctx context.Context) {
    // Create trace region for database operations
    ctx, task := trace.NewTask(ctx, "database_operations")
    defer task.End()
    
    // Individual database queries
    trace.WithRegion(ctx, "user_query", func() {
        queryUsers()
    })
    
    trace.WithRegion(ctx, "order_query", func() {
        queryOrders()
    })
    
    trace.WithRegion(ctx, "analytics_query", func() {
        queryAnalytics()
    })
}

func traceBusinessLogic(ctx context.Context) {
    ctx, task := trace.NewTask(ctx, "business_logic")
    defer task.End()
    
    // Order processing pipeline
    trace.WithRegion(ctx, "validate_order", func() {
        validateOrder()
    })
    
    trace.WithRegion(ctx, "calculate_pricing", func() {
        calculatePricing()
    })
    
    trace.WithRegion(ctx, "inventory_check", func() {
        checkInventory()
    })
    
    trace.WithRegion(ctx, "process_payment", func() {
        processPayment()
    })
}

func traceExternalAPIs(ctx context.Context) {
    ctx, task := trace.NewTask(ctx, "external_apis")
    defer task.End()
    
    var wg sync.WaitGroup
    
    // Parallel API calls
    apis := []string{"payment_service", "inventory_service", "notification_service"}
    
    for _, api := range apis {
        wg.Add(1)
        go func(apiName string) {
            defer wg.Done()
            
            trace.WithRegion(ctx, apiName, func() {
                callExternalAPI(apiName)
            })
        }(api)
    }
    
    wg.Wait()
}

func traceBackgroundTasks(ctx context.Context) {
    ctx, task := trace.NewTask(ctx, "background_tasks")
    defer task.End()
    
    // Multiple background workers
    for i := 0; i < 5; i++ {
        go func(workerID int) {
            workerCtx, workerTask := trace.NewTask(ctx, fmt.Sprintf("worker_%d", workerID))
            defer workerTask.End()
            
            for j := 0; j < 10; j++ {
                trace.WithRegion(workerCtx, "process_item", func() {
                    processBackgroundItem(workerID, j)
                })
                
                time.Sleep(10 * time.Millisecond)
            }
        }(i)
    }
    
    time.Sleep(200 * time.Millisecond) // Wait for background work
}

// Simulated functions
func queryUsers()             { time.Sleep(5 * time.Millisecond) }
func queryOrders()            { time.Sleep(8 * time.Millisecond) }
func queryAnalytics()         { time.Sleep(15 * time.Millisecond) }
func validateOrder()          { time.Sleep(2 * time.Millisecond) }
func calculatePricing()       { time.Sleep(3 * time.Millisecond) }
func checkInventory()         { time.Sleep(10 * time.Millisecond) }
func processPayment()         { time.Sleep(20 * time.Millisecond) }

func callExternalAPI(api string) {
    // Simulate different API latencies
    latencies := map[string]time.Duration{
        "payment_service":      30 * time.Millisecond,
        "inventory_service":    15 * time.Millisecond,
        "notification_service": 5 * time.Millisecond,
    }
    
    if latency, exists := latencies[api]; exists {
        time.Sleep(latency)
    } else {
        time.Sleep(10 * time.Millisecond)
    }
}

func processBackgroundItem(workerID, itemID int) {
    // Simulate varying processing times
    processingTime := time.Duration((workerID*itemID)%10+1) * time.Millisecond
    time.Sleep(processingTime)
}
```

### Trace Analysis and Parsing

```go
package main

import (
    "fmt"
    "internal/trace"
    "log"
    "os"
    "sort"
    "time"
)

// Programmatic trace analysis
type TraceAnalysis struct {
    Duration       time.Duration
    Goroutines     int
    Events         map[string]int
    GCEvents       []GCEvent
    UserRegions    map[string]RegionStats
    SystemCalls    map[string]int
    NetworkEvents  int
}

type GCEvent struct {
    Timestamp time.Duration
    Duration  time.Duration
    Type      string
}

type RegionStats struct {
    Count       int
    TotalTime   time.Duration
    AverageTime time.Duration
    MinTime     time.Duration
    MaxTime     time.Duration
}

func analyzeTraceFile(filename string) (*TraceAnalysis, error) {
    // Open trace file
    f, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to open trace file: %v", err)
    }
    defer f.Close()
    
    // Parse trace
    res, err := trace.Parse(f, "")
    if err != nil {
        return nil, fmt.Errorf("failed to parse trace: %v", err)
    }
    
    // Analyze events
    analysis := &TraceAnalysis{
        Events:      make(map[string]int),
        UserRegions: make(map[string]RegionStats),
        SystemCalls: make(map[string]int),
    }
    
    return analyzeEvents(res.Events, analysis)
}

func analyzeEvents(events []*trace.Event, analysis *TraceAnalysis) (*TraceAnalysis, error) {
    if len(events) == 0 {
        return analysis, nil
    }
    
    // Calculate trace duration
    firstEvent := events[0]
    lastEvent := events[len(events)-1]
    analysis.Duration = time.Duration(lastEvent.Ts - firstEvent.Ts)
    
    // Track goroutines
    goroutines := make(map[uint64]bool)
    
    // Track user regions
    regionStarts := make(map[uint64]time.Duration) // taskID -> start time
    regionData := make(map[string][]time.Duration) // region name -> durations
    
    for _, ev := range events {
        eventType := ev.Type.String()
        analysis.Events[eventType]++
        
        // Track goroutines
        if ev.G != 0 {
            goroutines[ev.G] = true
        }
        
        switch ev.Type {
        case trace.EvGCStart, trace.EvGCDone, trace.EvGCSTWStart, trace.EvGCSTWDone:
            analysis.GCEvents = append(analysis.GCEvents, GCEvent{
                Timestamp: time.Duration(ev.Ts),
                Type:      eventType,
            })
            
        case trace.EvGoSysCall:
            if len(ev.SArgs) > 0 {
                syscallName := ev.SArgs[0]
                analysis.SystemCalls[syscallName]++
            }
            
        case trace.EvUserTaskCreate:
            if len(ev.SArgs) > 0 {
                taskName := ev.SArgs[0]
                regionStarts[ev.Args[0]] = time.Duration(ev.Ts)
                if _, exists := regionData[taskName]; !exists {
                    regionData[taskName] = make([]time.Duration, 0)
                }
            }
            
        case trace.EvUserTaskEnd:
            taskID := ev.Args[0]
            if startTime, exists := regionStarts[taskID]; exists {
                duration := time.Duration(ev.Ts) - startTime
                
                // Find task name from creation event
                for _, searchEv := range events {
                    if searchEv.Type == trace.EvUserTaskCreate && searchEv.Args[0] == taskID {
                        if len(searchEv.SArgs) > 0 {
                            taskName := searchEv.SArgs[0]
                            regionData[taskName] = append(regionData[taskName], duration)
                        }
                        break
                    }
                }
                
                delete(regionStarts, taskID)
            }
            
        case trace.EvUserRegion:
            // Handle user regions
            if len(ev.SArgs) > 0 {
                regionName := ev.SArgs[0]
                if _, exists := regionData[regionName]; !exists {
                    regionData[regionName] = make([]time.Duration, 0)
                }
            }
        }
    }
    
    analysis.Goroutines = len(goroutines)
    
    // Calculate region statistics
    for regionName, durations := range regionData {
        if len(durations) == 0 {
            continue
        }
        
        stats := RegionStats{
            Count: len(durations),
        }
        
        var total time.Duration
        stats.MinTime = durations[0]
        stats.MaxTime = durations[0]
        
        for _, duration := range durations {
            total += duration
            if duration < stats.MinTime {
                stats.MinTime = duration
            }
            if duration > stats.MaxTime {
                stats.MaxTime = duration
            }
        }
        
        stats.TotalTime = total
        stats.AverageTime = total / time.Duration(len(durations))
        
        analysis.UserRegions[regionName] = stats
    }
    
    return analysis, nil
}

func printTraceAnalysis(analysis *TraceAnalysis) {
    fmt.Printf("=== Trace Analysis ===\n")
    fmt.Printf("Duration: %v\n", analysis.Duration)
    fmt.Printf("Goroutines: %d\n", analysis.Goroutines)
    fmt.Printf("GC Events: %d\n", len(analysis.GCEvents))
    fmt.Printf("Network Events: %d\n", analysis.NetworkEvents)
    fmt.Printf("\n")
    
    // Top events by count
    fmt.Printf("Top Event Types:\n")
    
    type eventCount struct {
        Type  string
        Count int
    }
    
    var events []eventCount
    for eventType, count := range analysis.Events {
        events = append(events, eventCount{eventType, count})
    }
    
    sort.Slice(events, func(i, j int) bool {
        return events[i].Count > events[j].Count
    })
    
    for i, ev := range events {
        if i >= 10 { // Top 10
            break
        }
        fmt.Printf("  %-30s: %8d\n", ev.Type, ev.Count)
    }
    fmt.Printf("\n")
    
    // GC analysis
    if len(analysis.GCEvents) > 0 {
        fmt.Printf("GC Analysis:\n")
        
        var gcDuration time.Duration
        gcStarts := make(map[time.Duration]bool)
        
        for _, gcEvent := range analysis.GCEvents {
            if gcEvent.Type == "EvGCStart" {
                gcStarts[gcEvent.Timestamp] = true
            } else if gcEvent.Type == "EvGCDone" {
                // Find corresponding start
                for startTime := range gcStarts {
                    if startTime <= gcEvent.Timestamp {
                        gcDuration += gcEvent.Timestamp - startTime
                        delete(gcStarts, startTime)
                        break
                    }
                }
            }
        }
        
        gcPercent := float64(gcDuration) / float64(analysis.Duration) * 100
        fmt.Printf("  Total GC time: %v (%.2f%%)\n", gcDuration, gcPercent)
        fmt.Printf("  GC frequency: %.2f cycles/second\n", 
            float64(len(analysis.GCEvents)/2)/analysis.Duration.Seconds())
        fmt.Printf("\n")
    }
    
    // User regions analysis
    if len(analysis.UserRegions) > 0 {
        fmt.Printf("User Regions:\n")
        fmt.Printf("%-30s %8s %12s %12s %12s %12s\n", 
            "Region", "Count", "Total", "Average", "Min", "Max")
        fmt.Printf("%s\n", strings.Repeat("-", 90))
        
        for regionName, stats := range analysis.UserRegions {
            fmt.Printf("%-30s %8d %12s %12s %12s %12s\n",
                regionName,
                stats.Count,
                stats.TotalTime.Truncate(time.Microsecond),
                stats.AverageTime.Truncate(time.Microsecond),
                stats.MinTime.Truncate(time.Microsecond),
                stats.MaxTime.Truncate(time.Microsecond))
        }
        fmt.Printf("\n")
    }
    
    // System calls analysis
    if len(analysis.SystemCalls) > 0 {
        fmt.Printf("Top System Calls:\n")
        
        type syscallCount struct {
            Name  string
            Count int
        }
        
        var syscalls []syscallCount
        for name, count := range analysis.SystemCalls {
            syscalls = append(syscalls, syscallCount{name, count})
        }
        
        sort.Slice(syscalls, func(i, j int) bool {
            return syscalls[i].Count > syscalls[j].Count
        })
        
        for i, sc := range syscalls {
            if i >= 10 { // Top 10
                break
            }
            fmt.Printf("  %-20s: %8d\n", sc.Name, sc.Count)
        }
    }
}
```

### Performance Bottleneck Detection

```go
// Trace-based bottleneck detection
type BottleneckAnalyzer struct {
    traceFile string
    analysis  *TraceAnalysis
}

func NewBottleneckAnalyzer(traceFile string) *BottleneckAnalyzer {
    return &BottleneckAnalyzer{traceFile: traceFile}
}

func (ba *BottleneckAnalyzer) Analyze() error {
    analysis, err := analyzeTraceFile(ba.traceFile)
    if err != nil {
        return err
    }
    
    ba.analysis = analysis
    return nil
}

func (ba *BottleneckAnalyzer) DetectBottlenecks() []Bottleneck {
    if ba.analysis == nil {
        return nil
    }
    
    var bottlenecks []Bottleneck
    
    // Detect GC pressure
    if gcBottleneck := ba.detectGCBottleneck(); gcBottleneck != nil {
        bottlenecks = append(bottlenecks, *gcBottleneck)
    }
    
    // Detect scheduler contention
    if schedBottleneck := ba.detectSchedulerBottleneck(); schedBottleneck != nil {
        bottlenecks = append(bottlenecks, *schedBottleneck)
    }
    
    // Detect I/O bottlenecks
    if ioBottleneck := ba.detectIOBottleneck(); ioBottleneck != nil {
        bottlenecks = append(bottlenecks, *ioBottleneck)
    }
    
    // Detect slow regions
    slowRegions := ba.detectSlowRegions()
    bottlenecks = append(bottlenecks, slowRegions...)
    
    return bottlenecks
}

type Bottleneck struct {
    Type        string
    Severity    string
    Description string
    Impact      float64 // Percentage of total time
    Suggestion  string
}

func (ba *BottleneckAnalyzer) detectGCBottleneck() *Bottleneck {
    if len(ba.analysis.GCEvents) == 0 {
        return nil
    }
    
    gcTime := ba.calculateGCTime()
    gcPercent := float64(gcTime) / float64(ba.analysis.Duration) * 100
    
    if gcPercent > 10 { // More than 10% in GC
        severity := "medium"
        if gcPercent > 20 {
            severity = "high"
        }
        
        return &Bottleneck{
            Type:        "garbage_collection",
            Severity:    severity,
            Description: fmt.Sprintf("High GC overhead: %.1f%% of execution time", gcPercent),
            Impact:      gcPercent,
            Suggestion:  "Reduce allocation rate, increase GOGC, or optimize data structures",
        }
    }
    
    return nil
}

func (ba *BottleneckAnalyzer) detectSchedulerBottleneck() *Bottleneck {
    // Look for scheduler-related events
    schedulerEvents := 0
    for eventType, count := range ba.analysis.Events {
        if strings.Contains(eventType, "Sched") || strings.Contains(eventType, "Proc") {
            schedulerEvents += count
        }
    }
    
    totalEvents := 0
    for _, count := range ba.analysis.Events {
        totalEvents += count
    }
    
    if totalEvents > 0 {
        schedulerPercent := float64(schedulerEvents) / float64(totalEvents) * 100
        
        if schedulerPercent > 15 { // High scheduler activity
            return &Bottleneck{
                Type:        "scheduler_contention",
                Severity:    "medium",
                Description: fmt.Sprintf("High scheduler activity: %.1f%% of events", schedulerPercent),
                Impact:      schedulerPercent,
                Suggestion:  "Optimize goroutine creation/destruction, reduce lock contention",
            }
        }
    }
    
    return nil
}

func (ba *BottleneckAnalyzer) detectIOBottleneck() *Bottleneck {
    // Count I/O related system calls
    ioSyscalls := 0
    totalSyscalls := 0
    
    for syscall, count := range ba.analysis.SystemCalls {
        totalSyscalls += count
        if isIOSyscall(syscall) {
            ioSyscalls += count
        }
    }
    
    if totalSyscalls > 0 {
        ioPercent := float64(ioSyscalls) / float64(totalSyscalls) * 100
        
        if ioPercent > 50 { // More than 50% I/O syscalls
            return &Bottleneck{
                Type:        "io_bottleneck",
                Severity:    "medium",
                Description: fmt.Sprintf("High I/O activity: %.1f%% of syscalls", ioPercent),
                Impact:      ioPercent,
                Suggestion:  "Use buffering, connection pooling, or async I/O",
            }
        }
    }
    
    return nil
}

func (ba *BottleneckAnalyzer) detectSlowRegions() []Bottleneck {
    var bottlenecks []Bottleneck
    
    for regionName, stats := range ba.analysis.UserRegions {
        regionPercent := float64(stats.TotalTime) / float64(ba.analysis.Duration) * 100
        
        if regionPercent > 5 { // Region takes more than 5% of total time
            severity := "low"
            if regionPercent > 15 {
                severity = "medium"
            }
            if regionPercent > 30 {
                severity = "high"
            }
            
            bottlenecks = append(bottlenecks, Bottleneck{
                Type:        "slow_region",
                Severity:    severity,
                Description: fmt.Sprintf("Slow region '%s': %.1f%% of execution time", regionName, regionPercent),
                Impact:      regionPercent,
                Suggestion:  "Profile and optimize the slow region",
            })
        }
    }
    
    return bottlenecks
}

func (ba *BottleneckAnalyzer) calculateGCTime() time.Duration {
    var gcTime time.Duration
    gcStarts := make(map[time.Duration]bool)
    
    for _, gcEvent := range ba.analysis.GCEvents {
        if gcEvent.Type == "EvGCStart" {
            gcStarts[gcEvent.Timestamp] = true
        } else if gcEvent.Type == "EvGCDone" {
            for startTime := range gcStarts {
                if startTime <= gcEvent.Timestamp {
                    gcTime += gcEvent.Timestamp - startTime
                    delete(gcStarts, startTime)
                    break
                }
            }
        }
    }
    
    return gcTime
}

func isIOSyscall(syscall string) bool {
    ioSyscalls := map[string]bool{
        "read":   true,
        "write":  true,
        "open":   true,
        "close":  true,
        "accept": true,
        "recv":   true,
        "send":   true,
        "poll":   true,
        "select": true,
    }
    
    return ioSyscalls[syscall]
}

func (ba *BottleneckAnalyzer) PrintBottlenecks() {
    bottlenecks := ba.DetectBottlenecks()
    
    if len(bottlenecks) == 0 {
        fmt.Println("No significant bottlenecks detected")
        return
    }
    
    fmt.Printf("=== Detected Bottlenecks ===\n")
    
    // Sort by impact
    sort.Slice(bottlenecks, func(i, j int) bool {
        return bottlenecks[i].Impact > bottlenecks[j].Impact
    })
    
    for _, bottleneck := range bottlenecks {
        fmt.Printf("\n%s [%s]\n", strings.ToUpper(bottleneck.Severity), bottleneck.Type)
        fmt.Printf("  Impact: %.1f%%\n", bottleneck.Impact)
        fmt.Printf("  Description: %s\n", bottleneck.Description)
        fmt.Printf("  Suggestion: %s\n", bottleneck.Suggestion)
    }
}
```

## Trace Visualization and Reporting

### Custom Trace Viewer

```go
// Simple trace event viewer
type TraceViewer struct {
    events    []*trace.Event
    timeRange struct {
        start, end time.Duration
    }
    filters map[string]bool
}

func NewTraceViewer(filename string) (*TraceViewer, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    res, err := trace.Parse(f, "")
    if err != nil {
        return nil, err
    }
    
    viewer := &TraceViewer{
        events:  res.Events,
        filters: make(map[string]bool),
    }
    
    if len(viewer.events) > 0 {
        viewer.timeRange.start = time.Duration(viewer.events[0].Ts)
        viewer.timeRange.end = time.Duration(viewer.events[len(viewer.events)-1].Ts)
    }
    
    return viewer, nil
}

func (tv *TraceViewer) SetTimeRange(start, end time.Duration) {
    tv.timeRange.start = start
    tv.timeRange.end = end
}

func (tv *TraceViewer) SetFilter(eventType string, enabled bool) {
    tv.filters[eventType] = enabled
}

func (tv *TraceViewer) GetFilteredEvents() []*trace.Event {
    var filtered []*trace.Event
    
    for _, event := range tv.events {
        eventTime := time.Duration(event.Ts)
        
        // Time range filter
        if eventTime < tv.timeRange.start || eventTime > tv.timeRange.end {
            continue
        }
        
        // Event type filter
        eventType := event.Type.String()
        if len(tv.filters) > 0 {
            if enabled, exists := tv.filters[eventType]; exists && !enabled {
                continue
            }
        }
        
        filtered = append(filtered, event)
    }
    
    return filtered
}

func (tv *TraceViewer) PrintTimeline(maxEvents int) {
    events := tv.GetFilteredEvents()
    
    if len(events) > maxEvents {
        events = events[:maxEvents]
    }
    
    fmt.Printf("=== Trace Timeline ===\n")
    fmt.Printf("Time Range: %v - %v\n", tv.timeRange.start, tv.timeRange.end)
    fmt.Printf("Events: %d\n\n", len(events))
    
    fmt.Printf("%-12s %-20s %-10s %s\n", "Time", "Event", "Goroutine", "Details")
    fmt.Printf("%s\n", strings.Repeat("-", 80))
    
    baseTime := tv.timeRange.start
    
    for _, event := range events {
        relativeTime := time.Duration(event.Ts) - baseTime
        eventType := event.Type.String()
        
        details := ""
        if len(event.SArgs) > 0 {
            details = strings.Join(event.SArgs, ", ")
        }
        
        fmt.Printf("%-12s %-20s %-10d %s\n",
            relativeTime.Truncate(time.Microsecond),
            eventType,
            event.G,
            details)
    }
}

func (tv *TraceViewer) GenerateReport() TraceReport {
    events := tv.GetFilteredEvents()
    
    report := TraceReport{
        TimeRange:    tv.timeRange.end - tv.timeRange.start,
        EventCount:   len(events),
        EventTypes:   make(map[string]int),
        Goroutines:   make(map[uint64]int),
        UserRegions:  make(map[string]RegionStats),
    }
    
    for _, event := range events {
        eventType := event.Type.String()
        report.EventTypes[eventType]++
        
        if event.G != 0 {
            report.Goroutines[event.G]++
        }
    }
    
    report.UniqueGoroutines = len(report.Goroutines)
    
    return report
}

type TraceReport struct {
    TimeRange        time.Duration
    EventCount       int
    EventTypes       map[string]int
    Goroutines       map[uint64]int
    UniqueGoroutines int
    UserRegions      map[string]RegionStats
}

func (tr *TraceReport) Print() {
    fmt.Printf("=== Trace Report ===\n")
    fmt.Printf("Duration: %v\n", tr.TimeRange)
    fmt.Printf("Total Events: %d\n", tr.EventCount)
    fmt.Printf("Unique Goroutines: %d\n", tr.UniqueGoroutines)
    fmt.Printf("Event Rate: %.2f events/ms\n", 
        float64(tr.EventCount)/float64(tr.TimeRange.Nanoseconds()/1e6))
    fmt.Printf("\n")
    
    // Top event types
    fmt.Printf("Top Event Types:\n")
    
    type eventCount struct {
        Type  string
        Count int
    }
    
    var events []eventCount
    for eventType, count := range tr.EventTypes {
        events = append(events, eventCount{eventType, count})
    }
    
    sort.Slice(events, func(i, j int) bool {
        return events[i].Count > events[j].Count
    })
    
    for i, ev := range events {
        if i >= 10 {
            break
        }
        percentage := float64(ev.Count) / float64(tr.EventCount) * 100
        fmt.Printf("  %-30s: %8d (%.1f%%)\n", ev.Type, ev.Count, percentage)
    }
}
```

Execution tracing provides the most detailed view of Go program behavior. Master these techniques to understand complex concurrency issues and optimize system-wide performance.

---

**Next**: [Block and Mutex Profiling](block-mutex-profiling.md) - Analyze synchronization bottlenecks
