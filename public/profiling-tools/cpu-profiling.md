# CPU Profiling

CPU profiling is the most commonly used profiling technique for identifying performance bottlenecks. This chapter covers comprehensive CPU profiling techniques, analysis methods, and optimization strategies.

## CPU Profiling Fundamentals

### How CPU Profiling Works

CPU profiling uses statistical sampling to measure where your program spends time:

```
┌─────────────────────────────────────────────────────────────┐
│                   CPU Profiling Process                    │
├─────────────────────────────────────────────────────────────┤
│  Sample Timer   │  Stack Walking  │  Aggregation           │
│                 │                 │                        │
│  • 100Hz (10ms) │  • Capture PC   │  • Count samples       │
│  • Interrupts   │  • Unwind stack │  • Build call graph    │
│  • Low overhead │  • Symbol lookup│  • Calculate %time     │
└─────────────────────────────────────────────────────────────┘
```

### Sampling Methodology

```go
package main

import (
    "log"
    "os"
    "runtime/pprof"
    "time"
)

// Basic CPU profiling setup
func basicCPUProfiling() {
    // Create profile file
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal("could not create CPU profile: ", err)
    }
    defer f.Close()
    
    // Start CPU profiling
    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal("could not start CPU profile: ", err)
    }
    defer pprof.StopCPUProfile()
    
    // Run your application code
    runWorkload()
}

// Sample workload for profiling
func runWorkload() {
    // CPU-intensive operations
    for i := 0; i < 1000000; i++ {
        computeHeavy(i)
        if i%10000 == 0 {
            performIO()
        }
    }
}

func computeHeavy(n int) int {
    result := 0
    for i := 0; i < n%1000; i++ {
        result += i * i
    }
    return result
}

func performIO() {
    time.Sleep(time.Microsecond)
}
```

## Collection Methods

### 1. HTTP Endpoint Collection

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    _ "net/http/pprof"
    "time"
)

func httpProfilingServer() {
    // Start application
    go runApplication()
    
    // Start profiling server
    fmt.Println("Profiling server starting on :6060")
    fmt.Println("Collect CPU profile with:")
    fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30")
    fmt.Println("  curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof")
    
    log.Fatal(http.ListenAndServe("localhost:6060", nil))
}

func runApplication() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Simulate work
            doWork()
        }
    }
}

func doWork() {
    // Simulate different workload patterns
    switch time.Now().Second() % 3 {
    case 0:
        cpuIntensiveWork()
    case 1:
        memoryIntensiveWork()
    case 2:
        ioIntensiveWork()
    }
}
```

### 2. Programmatic Collection

```go
package main

import (
    "context"
    "fmt"
    "os"
    "runtime/pprof"
    "sync"
    "time"
)

// Advanced CPU profiling with control
type CPUProfiler struct {
    mu       sync.Mutex
    active   bool
    file     *os.File
    duration time.Duration
}

func NewCPUProfiler() *CPUProfiler {
    return &CPUProfiler{
        duration: 30 * time.Second,
    }
}

func (cp *CPUProfiler) Start(filename string) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if cp.active {
        return fmt.Errorf("profiling already active")
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("could not create profile file: %v", err)
    }
    
    if err := pprof.StartCPUProfile(f); err != nil {
        f.Close()
        return fmt.Errorf("could not start CPU profile: %v", err)
    }
    
    cp.file = f
    cp.active = true
    
    // Auto-stop after duration
    go func() {
        time.Sleep(cp.duration)
        cp.Stop()
    }()
    
    return nil
}

func (cp *CPUProfiler) Stop() error {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if !cp.active {
        return nil
    }
    
    pprof.StopCPUProfile()
    err := cp.file.Close()
    cp.active = false
    cp.file = nil
    
    return err
}

func (cp *CPUProfiler) IsActive() bool {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    return cp.active
}

// Context-based profiling
func contextBasedProfiling(ctx context.Context) {
    profiler := NewCPUProfiler()
    
    // Start profiling
    if err := profiler.Start("context_cpu.prof"); err != nil {
        log.Printf("Failed to start profiling: %v", err)
        return
    }
    
    // Ensure cleanup
    defer profiler.Stop()
    
    // Run workload with context
    select {
    case <-ctx.Done():
        log.Println("Context cancelled, stopping profiling")
        return
    case <-time.After(time.Minute):
        log.Println("Workload completed")
    }
}
```

### 3. Conditional Profiling

```go
package main

import (
    "os"
    "os/signal"
    "runtime/pprof"
    "syscall"
    "sync"
)

// Signal-triggered profiling
func signalBasedProfiling() {
    var (
        profiling   bool
        profilingMu sync.Mutex
        cpuFile     *os.File
    )
    
    // Handle SIGUSR1 for starting/stopping profiling
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGUSR1)
    
    go func() {
        for range c {
            profilingMu.Lock()
            
            if profiling {
                // Stop profiling
                pprof.StopCPUProfile()
                if cpuFile != nil {
                    cpuFile.Close()
                }
                profiling = false
                log.Println("CPU profiling stopped")
            } else {
                // Start profiling
                timestamp := time.Now().Format("20060102_150405")
                filename := fmt.Sprintf("cpu_profile_%s.prof", timestamp)
                
                f, err := os.Create(filename)
                if err != nil {
                    log.Printf("Could not create profile: %v", err)
                    profilingMu.Unlock()
                    continue
                }
                
                if err := pprof.StartCPUProfile(f); err != nil {
                    log.Printf("Could not start profiling: %v", err)
                    f.Close()
                    profilingMu.Unlock()
                    continue
                }
                
                cpuFile = f
                profiling = true
                log.Printf("CPU profiling started: %s", filename)
            }
            
            profilingMu.Unlock()
        }
    }()
    
    // Run application
    runApplication()
}

// Performance threshold-based profiling
func thresholdBasedProfiling() {
    const latencyThreshold = 100 * time.Millisecond
    
    profiler := NewCPUProfiler()
    var consecutiveSlowRequests int
    
    for {
        start := time.Now()
        
        // Process request
        processRequest()
        
        latency := time.Since(start)
        
        if latency > latencyThreshold {
            consecutiveSlowRequests++
            
            // Start profiling after 3 consecutive slow requests
            if consecutiveSlowRequests >= 3 && !profiler.IsActive() {
                timestamp := time.Now().Format("20060102_150405")
                filename := fmt.Sprintf("slow_requests_%s.prof", timestamp)
                
                if err := profiler.Start(filename); err == nil {
                    log.Printf("Started profiling due to slow requests: %s", filename)
                }
            }
        } else {
            consecutiveSlowRequests = 0
            
            // Stop profiling when performance is good
            if profiler.IsActive() {
                profiler.Stop()
                log.Println("Stopped profiling - performance normalized")
            }
        }
        
        time.Sleep(time.Second)
    }
}
```

## Profile Analysis

### Command Line Analysis

```bash
# Basic profile analysis
go tool pprof cpu.prof

# Interactive commands
(pprof) top10          # Show top 10 functions by CPU time
(pprof) top10 -cum     # Show top 10 by cumulative time
(pprof) list main.main # Show annotated source for function
(pprof) web            # Generate web visualization
(pprof) png            # Generate PNG call graph
(pprof) help           # Show all commands

# Direct analysis from HTTP endpoint
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Compare two profiles
go tool pprof -base=baseline.prof current.prof

# Focus on specific functions
go tool pprof -focus="main.*" cpu.prof

# Ignore specific functions
go tool pprof -ignore="runtime.*" cpu.prof
```

### Programmatic Analysis

```go
package main

import (
    "bytes"
    "fmt"
    "log"
    "sort"
    
    "github.com/google/pprof/profile"
)

// Analyze CPU profile programmatically
func analyzeCPUProfile(filename string) {
    // Read profile
    f, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    p, err := profile.Parse(f)
    if err != nil {
        log.Fatal(err)
    }
    
    // Analyze the profile
    analysis := analyzeSamples(p)
    printAnalysis(analysis)
}

type FunctionAnalysis struct {
    Name           string
    SelfSamples    int64
    CumSamples     int64
    SelfPercent    float64
    CumPercent     float64
    SelfTime       time.Duration
    CumTime        time.Duration
}

type ProfileAnalysis struct {
    TotalSamples   int64
    SampleRate     time.Duration
    TotalTime      time.Duration
    Functions      []FunctionAnalysis
}

func analyzeSamples(p *profile.Profile) ProfileAnalysis {
    if len(p.SampleType) == 0 {
        return ProfileAnalysis{}
    }
    
    // Calculate sample rate
    var sampleRate time.Duration = time.Second / 100 // Default 100Hz
    if p.Period > 0 {
        sampleRate = time.Duration(p.Period)
    }
    
    // Aggregate samples by function
    functionSamples := make(map[string]*FunctionAnalysis)
    var totalSamples int64
    
    for _, sample := range p.Sample {
        if len(sample.Value) == 0 {
            continue
        }
        
        sampleValue := sample.Value[0]
        totalSamples += sampleValue
        
        // Walk the call stack
        for i, location := range sample.Location {
            for _, line := range location.Line {
                funcName := line.Function.Name
                
                analysis, exists := functionSamples[funcName]
                if !exists {
                    analysis = &FunctionAnalysis{Name: funcName}
                    functionSamples[funcName] = analysis
                }
                
                // Self samples (only for leaf function)
                if i == 0 {
                    analysis.SelfSamples += sampleValue
                }
                
                // Cumulative samples (for all functions in stack)
                analysis.CumSamples += sampleValue
            }
        }
    }
    
    // Convert to slice and calculate percentages
    functions := make([]FunctionAnalysis, 0, len(functionSamples))
    totalTime := time.Duration(totalSamples) * sampleRate
    
    for _, analysis := range functionSamples {
        analysis.SelfPercent = float64(analysis.SelfSamples) / float64(totalSamples) * 100
        analysis.CumPercent = float64(analysis.CumSamples) / float64(totalSamples) * 100
        analysis.SelfTime = time.Duration(analysis.SelfSamples) * sampleRate
        analysis.CumTime = time.Duration(analysis.CumSamples) * sampleRate
        
        functions = append(functions, *analysis)
    }
    
    // Sort by cumulative samples
    sort.Slice(functions, func(i, j int) bool {
        return functions[i].CumSamples > functions[j].CumSamples
    })
    
    return ProfileAnalysis{
        TotalSamples: totalSamples,
        SampleRate:   sampleRate,
        TotalTime:    totalTime,
        Functions:    functions,
    }
}

func printAnalysis(analysis ProfileAnalysis) {
    fmt.Printf("=== CPU Profile Analysis ===\n")
    fmt.Printf("Total samples: %d\n", analysis.TotalSamples)
    fmt.Printf("Sample rate: %v\n", analysis.SampleRate)
    fmt.Printf("Total time: %v\n", analysis.TotalTime)
    fmt.Printf("\n")
    
    fmt.Printf("%-50s %8s %8s %8s %8s\n", "Function", "Self%", "Cum%", "Self", "Cum")
    fmt.Printf("%s\n", strings.Repeat("-", 90))
    
    for i, fn := range analysis.Functions {
        if i >= 20 { // Show top 20
            break
        }
        
        fmt.Printf("%-50s %7.2f%% %7.2f%% %8s %8s\n",
            truncateString(fn.Name, 50),
            fn.SelfPercent,
            fn.CumPercent,
            fn.SelfTime.Truncate(time.Millisecond),
            fn.CumTime.Truncate(time.Millisecond))
    }
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

### Comparative Analysis

```go
// Compare two CPU profiles
func compareProfiles(baseline, current string) {
    baselineProfile := loadProfile(baseline)
    currentProfile := loadProfile(current)
    
    comparison := compareProfileAnalysis(
        analyzeSamples(baselineProfile),
        analyzeSamples(currentProfile),
    )
    
    printComparison(comparison)
}

type ProfileComparison struct {
    Baseline ProfileAnalysis
    Current  ProfileAnalysis
    Changes  []FunctionChange
}

type FunctionChange struct {
    Name            string
    BaselineSelf    time.Duration
    CurrentSelf     time.Duration
    ChangePercent   float64
    ChangeAbsolute  time.Duration
}

func compareProfileAnalysis(baseline, current ProfileAnalysis) ProfileComparison {
    // Create function maps for easier lookup
    baselineMap := make(map[string]FunctionAnalysis)
    for _, fn := range baseline.Functions {
        baselineMap[fn.Name] = fn
    }
    
    currentMap := make(map[string]FunctionAnalysis)
    for _, fn := range current.Functions {
        currentMap[fn.Name] = fn
    }
    
    // Calculate changes
    var changes []FunctionChange
    
    // Check all functions in current profile
    for name, currentFn := range currentMap {
        baselineFn, exists := baselineMap[name]
        if !exists {
            // New function
            changes = append(changes, FunctionChange{
                Name:           name,
                BaselineSelf:   0,
                CurrentSelf:    currentFn.SelfTime,
                ChangePercent:  100, // New function = 100% increase
                ChangeAbsolute: currentFn.SelfTime,
            })
            continue
        }
        
        // Calculate change
        if baselineFn.SelfTime > 0 {
            changePercent := (float64(currentFn.SelfTime-baselineFn.SelfTime) / float64(baselineFn.SelfTime)) * 100
            changeAbsolute := currentFn.SelfTime - baselineFn.SelfTime
            
            changes = append(changes, FunctionChange{
                Name:           name,
                BaselineSelf:   baselineFn.SelfTime,
                CurrentSelf:    currentFn.SelfTime,
                ChangePercent:  changePercent,
                ChangeAbsolute: changeAbsolute,
            })
        }
    }
    
    // Sort by absolute change
    sort.Slice(changes, func(i, j int) bool {
        return abs(changes[i].ChangeAbsolute) > abs(changes[j].ChangeAbsolute)
    })
    
    return ProfileComparison{
        Baseline: baseline,
        Current:  current,
        Changes:  changes,
    }
}

func printComparison(comparison ProfileComparison) {
    fmt.Printf("=== Profile Comparison ===\n")
    fmt.Printf("Baseline total time: %v\n", comparison.Baseline.TotalTime)
    fmt.Printf("Current total time: %v\n", comparison.Current.TotalTime)
    
    totalChange := comparison.Current.TotalTime - comparison.Baseline.TotalTime
    totalChangePercent := float64(totalChange) / float64(comparison.Baseline.TotalTime) * 100
    
    fmt.Printf("Total change: %v (%.2f%%)\n", totalChange, totalChangePercent)
    fmt.Printf("\n")
    
    fmt.Printf("Top Changes:\n")
    fmt.Printf("%-50s %12s %12s %12s %8s\n", "Function", "Baseline", "Current", "Change", "Change%")
    fmt.Printf("%s\n", strings.Repeat("-", 100))
    
    for i, change := range comparison.Changes {
        if i >= 15 { // Show top 15 changes
            break
        }
        
        if abs(change.ChangeAbsolute) < time.Millisecond {
            continue // Skip small changes
        }
        
        fmt.Printf("%-50s %12s %12s %12s %7.1f%%\n",
            truncateString(change.Name, 50),
            change.BaselineSelf.Truncate(time.Millisecond),
            change.CurrentSelf.Truncate(time.Millisecond),
            change.ChangeAbsolute.Truncate(time.Millisecond),
            change.ChangePercent)
    }
}

func abs(d time.Duration) time.Duration {
    if d < 0 {
        return -d
    }
    return d
}
```

## Hot Path Identification

### Top-Down Analysis

```go
// Identify hot paths in the application
func identifyHotPaths(profileFile string) {
    p := loadProfile(profileFile)
    analysis := analyzeSamples(p)
    
    // Find functions consuming > 5% of total CPU time
    hotFunctions := make([]FunctionAnalysis, 0)
    for _, fn := range analysis.Functions {
        if fn.SelfPercent > 5.0 {
            hotFunctions = append(hotFunctions, fn)
        }
    }
    
    fmt.Printf("=== Hot Paths (>5%% CPU time) ===\n")
    for _, fn := range hotFunctions {
        fmt.Printf("%s: %.2f%% (%v)\n", fn.Name, fn.SelfPercent, fn.SelfTime)
        
        // Analyze call patterns for this function
        analyzeCallPatterns(p, fn.Name)
    }
}

func analyzeCallPatterns(p *profile.Profile, functionName string) {
    callers := make(map[string]int64)
    callees := make(map[string]int64)
    
    for _, sample := range p.Sample {
        if len(sample.Value) == 0 {
            continue
        }
        
        sampleValue := sample.Value[0]
        
        // Find the function in the call stack
        for i, location := range sample.Location {
            for _, line := range location.Line {
                if line.Function.Name == functionName {
                    // Record caller (previous function in stack)
                    if i+1 < len(sample.Location) {
                        for _, callerLine := range sample.Location[i+1].Line {
                            callers[callerLine.Function.Name] += sampleValue
                        }
                    }
                    
                    // Record callee (next function in stack)
                    if i > 0 {
                        for _, calleeLine := range sample.Location[i-1].Line {
                            callees[calleeLine.Function.Name] += sampleValue
                        }
                    }
                }
            }
        }
    }
    
    // Print top callers and callees
    fmt.Printf("  Top callers:\n")
    for caller, samples := range callers {
        fmt.Printf("    %s: %d samples\n", caller, samples)
    }
    
    fmt.Printf("  Top callees:\n")
    for callee, samples := range callees {
        fmt.Printf("    %s: %d samples\n", callee, samples)
    }
}
```

### Call Graph Analysis

```go
import "github.com/google/pprof/profile"

// Build and analyze call graph
type CallGraphNode struct {
    Function     string
    SelfSamples  int64
    TotalSamples int64
    Callers      map[string]int64
    Callees      map[string]int64
    Depth        int
}

type CallGraph struct {
    Nodes      map[string]*CallGraphNode
    TotalSamples int64
}

func buildCallGraph(p *profile.Profile) *CallGraph {
    graph := &CallGraph{
        Nodes: make(map[string]*CallGraphNode),
    }
    
    for _, sample := range p.Sample {
        if len(sample.Value) == 0 {
            continue
        }
        
        sampleValue := sample.Value[0]
        graph.TotalSamples += sampleValue
        
        // Process call stack
        for i, location := range sample.Location {
            for _, line := range location.Line {
                funcName := line.Function.Name
                
                // Get or create node
                node := graph.getOrCreateNode(funcName)
                
                // Add total samples (cumulative)
                node.TotalSamples += sampleValue
                
                // Add self samples (only for leaf)
                if i == 0 {
                    node.SelfSamples += sampleValue
                }
                
                // Record call relationships
                if i+1 < len(sample.Location) {
                    // Has caller
                    for _, callerLine := range sample.Location[i+1].Line {
                        callerName := callerLine.Function.Name
                        callerNode := graph.getOrCreateNode(callerName)
                        
                        // Record caller -> callee relationship
                        if callerNode.Callees == nil {
                            callerNode.Callees = make(map[string]int64)
                        }
                        callerNode.Callees[funcName] += sampleValue
                        
                        // Record callee -> caller relationship
                        if node.Callers == nil {
                            node.Callers = make(map[string]int64)
                        }
                        node.Callers[callerName] += sampleValue
                    }
                }
            }
        }
    }
    
    return graph
}

func (cg *CallGraph) getOrCreateNode(funcName string) *CallGraphNode {
    node, exists := cg.Nodes[funcName]
    if !exists {
        node = &CallGraphNode{Function: funcName}
        cg.Nodes[funcName] = node
    }
    return node
}

func (cg *CallGraph) FindBottlenecks() []CallGraphNode {
    var bottlenecks []CallGraphNode
    
    for _, node := range cg.Nodes {
        selfPercent := float64(node.SelfSamples) / float64(cg.TotalSamples) * 100
        
        // Identify functions with high self time
        if selfPercent > 1.0 { // More than 1% self time
            bottlenecks = append(bottlenecks, *node)
        }
    }
    
    // Sort by self samples
    sort.Slice(bottlenecks, func(i, j int) bool {
        return bottlenecks[i].SelfSamples > bottlenecks[j].SelfSamples
    })
    
    return bottlenecks
}

func (cg *CallGraph) PrintBottlenecks() {
    bottlenecks := cg.FindBottlenecks()
    
    fmt.Printf("=== Performance Bottlenecks ===\n")
    fmt.Printf("%-50s %8s %8s %8s\n", "Function", "Self%", "Total%", "Calls")
    fmt.Printf("%s\n", strings.Repeat("-", 80))
    
    for _, node := range bottlenecks {
        selfPercent := float64(node.SelfSamples) / float64(cg.TotalSamples) * 100
        totalPercent := float64(node.TotalSamples) / float64(cg.TotalSamples) * 100
        callCount := len(node.Callers)
        
        fmt.Printf("%-50s %7.2f%% %7.2f%% %8d\n",
            truncateString(node.Function, 50),
            selfPercent,
            totalPercent,
            callCount)
    }
}
```

## Optimization Strategies

### CPU-Bound Optimization

```go
// Before optimization: inefficient algorithm
func inefficientSearch(data []int, target int) int {
    for i := 0; i < len(data); i++ {
        if data[i] == target {
            return i
        }
    }
    return -1
}

// After optimization: binary search for sorted data
func efficientSearch(data []int, target int) int {
    left, right := 0, len(data)-1
    
    for left <= right {
        mid := (left + right) / 2
        if data[mid] == target {
            return mid
        } else if data[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    return -1
}

// Profile-guided optimization example
func optimizeBasedOnProfile() {
    // Original hot function identified by profiling
    original := func(data []string) map[string]int {
        result := make(map[string]int)
        for _, item := range data {
            result[item]++ // Hot path: map operations
        }
        return result
    }
    
    // Optimized version: pre-allocate map
    optimized := func(data []string) map[string]int {
        result := make(map[string]int, len(data)) // Pre-allocate
        for _, item := range data {
            result[item]++
        }
        return result
    }
    
    // Further optimization: avoid string operations in hot path
    furtherOptimized := func(data []string) map[string]int {
        result := make(map[string]int, len(data))
        for i := range data { // Avoid string copies
            result[data[i]]++
        }
        return result
    }
    
    // Benchmark to validate optimizations
    benchmarkFunctions(original, optimized, furtherOptimized)
}
```

### Memory Access Optimization

```go
// Cache-friendly data access patterns
func cacheOptimizedProcessing(matrix [][]int) {
    rows := len(matrix)
    cols := len(matrix[0])
    
    // Row-major access (cache-friendly)
    for i := 0; i < rows; i++ {
        for j := 0; j < cols; j++ {
            matrix[i][j] *= 2
        }
    }
    
    // Avoid column-major access (cache-unfriendly)
    // for j := 0; j < cols; j++ {
    //     for i := 0; i < rows; i++ {
    //         matrix[i][j] *= 2
    //     }
    // }
}

// Minimize allocation in hot paths
func minimizeAllocations() {
    // Use object pools
    var bufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 0, 1024)
        },
    }
    
    processData := func(input []byte) []byte {
        buffer := bufferPool.Get().([]byte)
        buffer = buffer[:0] // Reset length
        defer bufferPool.Put(buffer)
        
        // Process using pooled buffer
        buffer = append(buffer, input...)
        // ... processing logic
        
        // Return copy, not pooled buffer
        result := make([]byte, len(buffer))
        copy(result, buffer)
        return result
    }
    
    // Use the optimized function
    data := []byte("test data")
    result := processData(data)
    _ = result
}
```

## Advanced CPU Profiling

### Custom Profilers

```go
// Custom function-level profiler
type FunctionProfiler struct {
    mu      sync.Mutex
    samples map[string][]time.Duration
}

func NewFunctionProfiler() *FunctionProfiler {
    return &FunctionProfiler{
        samples: make(map[string][]time.Duration),
    }
}

func (fp *FunctionProfiler) Profile(name string, fn func()) {
    start := time.Now()
    fn()
    duration := time.Since(start)
    
    fp.mu.Lock()
    fp.samples[name] = append(fp.samples[name], duration)
    fp.mu.Unlock()
}

func (fp *FunctionProfiler) Report() {
    fp.mu.Lock()
    defer fp.mu.Unlock()
    
    fmt.Printf("=== Function Profile Report ===\n")
    fmt.Printf("%-30s %8s %8s %8s %8s\n", "Function", "Count", "Total", "Average", "Max")
    fmt.Printf("%s\n", strings.Repeat("-", 70))
    
    for name, samples := range fp.samples {
        if len(samples) == 0 {
            continue
        }
        
        var total time.Duration
        var max time.Duration
        
        for _, duration := range samples {
            total += duration
            if duration > max {
                max = duration
            }
        }
        
        avg := total / time.Duration(len(samples))
        
        fmt.Printf("%-30s %8d %8s %8s %8s\n",
            name, len(samples),
            total.Truncate(time.Millisecond),
            avg.Truncate(time.Millisecond),
            max.Truncate(time.Millisecond))
    }
}

// Usage example
func useFunctionProfiler() {
    profiler := NewFunctionProfiler()
    
    // Profile different functions
    profiler.Profile("database_query", func() {
        time.Sleep(10 * time.Millisecond) // Simulate DB query
    })
    
    profiler.Profile("computation", func() {
        for i := 0; i < 1000000; i++ {
            _ = i * i
        }
    })
    
    profiler.Report()
}
```

### Sampling Profilers

```go
// Custom sampling profiler
type SamplingProfiler struct {
    interval    time.Duration
    samples     []StackSample
    running     bool
    mu          sync.Mutex
}

type StackSample struct {
    Timestamp time.Time
    Stack     []uintptr
}

func NewSamplingProfiler(interval time.Duration) *SamplingProfiler {
    return &SamplingProfiler{
        interval: interval,
        samples:  make([]StackSample, 0),
    }
}

func (sp *SamplingProfiler) Start() {
    sp.mu.Lock()
    if sp.running {
        sp.mu.Unlock()
        return
    }
    sp.running = true
    sp.mu.Unlock()
    
    go sp.samplingLoop()
}

func (sp *SamplingProfiler) Stop() {
    sp.mu.Lock()
    sp.running = false
    sp.mu.Unlock()
}

func (sp *SamplingProfiler) samplingLoop() {
    ticker := time.NewTicker(sp.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            sp.mu.Lock()
            if !sp.running {
                sp.mu.Unlock()
                return
            }
            
            // Capture stack trace
            stack := make([]uintptr, 32)
            n := runtime.Callers(2, stack)
            stack = stack[:n]
            
            sp.samples = append(sp.samples, StackSample{
                Timestamp: time.Now(),
                Stack:     stack,
            })
            
            sp.mu.Unlock()
        }
    }
}

func (sp *SamplingProfiler) GetSamples() []StackSample {
    sp.mu.Lock()
    defer sp.mu.Unlock()
    
    result := make([]StackSample, len(sp.samples))
    copy(result, sp.samples)
    return result
}
```

## CPU Profiling Best Practices

### ✅ **Do's**

1. **Profile production-like workloads**
   ```go
   // Use realistic data sizes and access patterns
   ```

2. **Profile for sufficient duration**
   ```go
   // CPU profiles: 10-60 seconds minimum
   // HTTP: ?seconds=30
   ```

3. **Focus on self time first**
   ```go
   // Functions that do actual work, not just call others
   ```

4. **Validate optimizations with benchmarks**
   ```go
   func BenchmarkBefore(b *testing.B) { /* ... */ }
   func BenchmarkAfter(b *testing.B)  { /* ... */ }
   ```

5. **Use comparative analysis**
   ```go
   go tool pprof -base=before.prof after.prof
   ```

### ❌ **Don'ts**

1. **Don't profile debug builds**
   ```bash
   # Use optimized builds: go build -o app
   ```

2. **Don't optimize without profiling**
   ```go
   // Always measure first
   ```

3. **Don't ignore call context**
   ```go
   // Look at callers and callees
   ```

4. **Don't profile too briefly**
   ```go
   // <10 seconds may not be representative
   ```

5. **Don't assume uniform distribution**
   ```go
   // Different workloads may have different hot paths
   ```

CPU profiling is your primary tool for identifying and resolving performance bottlenecks. Master these techniques to systematically optimize your applications.

---

**Next**: [Memory Profiling](memory-profiling.md) - Analyze memory allocation and usage patterns
