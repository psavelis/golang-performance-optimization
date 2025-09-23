# Advanced CPU Profiling Techniques

Beyond basic CPU profiling lies a sophisticated ecosystem of techniques for deep performance analysis. This guide explores advanced profiling methods, custom instrumentation, and expert-level analysis techniques.

## Advanced Profiling Strategies

### Multi-Profile Comparison

```go
package main

import (
    "fmt"
    "os"
    "runtime/pprof"
    "time"
)

type ProfileSession struct {
    Name     string
    Duration time.Duration
    Warmup   time.Duration
}

func compareProfileSessions() {
    sessions := []ProfileSession{
        {"baseline", 30 * time.Second, 5 * time.Second},
        {"optimized", 30 * time.Second, 5 * time.Second},
        {"alternative", 30 * time.Second, 5 * time.Second},
    }

    for _, session := range sessions {
        runProfileSession(session)
    }
}

func runProfileSession(session ProfileSession) {
    filename := fmt.Sprintf("cpu_%s.prof", session.Name)
    f, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    // Warmup period
    fmt.Printf("Warming up %s...\n", session.Name)
    time.Sleep(session.Warmup)

    // Start profiling
    fmt.Printf("Profiling %s for %v...\n", session.Name, session.Duration)
    pprof.StartCPUProfile(f)
    
    // Run workload
    runWorkload(session.Duration)
    
    pprof.StopCPUProfile()
    fmt.Printf("Profile saved to %s\n", filename)
}

func runWorkload(duration time.Duration) {
    start := time.Now()
    for time.Since(start) < duration {
        // Your application workload
        performComplexCalculation()
    }
}
```

### Conditional Profiling

```go
package main

import (
    "context"
    "log"
    "os"
    "runtime/pprof"
    "sync"
    "time"
)

type ConditionalProfiler struct {
    mu          sync.RWMutex
    enabled     bool
    activeFile  *os.File
    conditions  []ProfilingCondition
}

type ProfilingCondition struct {
    Name      string
    Condition func() bool
    Action    func()
}

func NewConditionalProfiler() *ConditionalProfiler {
    return &ConditionalProfiler{
        conditions: make([]ProfilingCondition, 0),
    }
}

func (cp *ConditionalProfiler) AddCondition(name string, condition func() bool, action func()) {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    cp.conditions = append(cp.conditions, ProfilingCondition{
        Name:      name,
        Condition: condition,
        Action:    action,
    })
}

func (cp *ConditionalProfiler) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            cp.Stop()
            return
        case <-ticker.C:
            cp.checkConditions()
        }
    }
}

func (cp *ConditionalProfiler) checkConditions() {
    cp.mu.RLock()
    conditions := cp.conditions
    cp.mu.RUnlock()

    for _, condition := range conditions {
        if condition.Condition() {
            log.Printf("Condition '%s' triggered, starting profiling", condition.Name)
            cp.startProfiling(condition.Name)
            condition.Action()
            
            // Profile for 30 seconds
            time.Sleep(30 * time.Second)
            cp.Stop()
            break
        }
    }
}

func (cp *ConditionalProfiler) startProfiling(reason string) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if cp.enabled {
        return // Already profiling
    }

    filename := fmt.Sprintf("cpu_triggered_%s_%d.prof", reason, time.Now().Unix())
    f, err := os.Create(filename)
    if err != nil {
        log.Printf("Failed to create profile file: %v", err)
        return
    }

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Printf("Failed to start CPU profile: %v", err)
        f.Close()
        return
    }

    cp.activeFile = f
    cp.enabled = true
    log.Printf("CPU profiling started: %s", filename)
}

func (cp *ConditionalProfiler) Stop() {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if !cp.enabled {
        return
    }

    pprof.StopCPUProfile()
    if cp.activeFile != nil {
        cp.activeFile.Close()
        cp.activeFile = nil
    }
    cp.enabled = false
    log.Println("CPU profiling stopped")
}

// Usage example
func useConditionalProfiler() {
    profiler := NewConditionalProfiler()

    // High CPU usage condition
    profiler.AddCondition("high_cpu", func() bool {
        return getCurrentCPUUsage() > 80.0
    }, func() {
        log.Println("High CPU usage detected")
    })

    // High memory condition
    profiler.AddCondition("high_memory", func() bool {
        return getCurrentMemoryUsage() > 1024*1024*1024 // 1GB
    }, func() {
        log.Println("High memory usage detected")
    })

    // Error rate condition
    profiler.AddCondition("high_errors", func() bool {
        return getErrorRate() > 0.05 // 5% error rate
    }, func() {
        log.Println("High error rate detected")
    })

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go profiler.Start(ctx)
    
    // Run your application
    runApplication()
}
```

### Sampling Rate Optimization

```go
package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

func optimizeSamplingRate() {
    rates := []int{50, 100, 200, 500, 1000}
    
    for _, rate := range rates {
        fmt.Printf("Testing sampling rate: %d Hz\n", rate)
        profileWithRate(rate)
    }
}

func profileWithRate(hz int) {
    // Set CPU profile rate
    runtime.SetCPUProfileRate(hz)
    
    filename := fmt.Sprintf("cpu_rate_%dhz.prof", hz)
    f, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    pprof.StartCPUProfile(f)
    
    // Run consistent workload
    start := time.Now()
    workload := 0
    for time.Since(start) < 10*time.Second {
        workload += performWork()
    }
    
    pprof.StopCPUProfile()
    
    // Reset to default
    runtime.SetCPUProfileRate(100)
    
    fmt.Printf("Completed workload: %d operations\n", workload)
}

func performWork() int {
    sum := 0
    for i := 0; i < 10000; i++ {
        sum += i * i
    }
    return sum
}
```

## Custom Profile Labels

```go
package main

import (
    "context"
    "runtime/pprof"
    "strconv"
)

func advancedLabelingExample() {
    // User ID labeling
    for userID := 1; userID <= 100; userID++ {
        go func(id int) {
            ctx := pprof.WithLabels(context.Background(), pprof.Labels(
                "user_id", strconv.Itoa(id),
                "user_type", getUserType(id),
                "region", getUserRegion(id),
            ))
            pprof.Do(ctx, pprof.Labels("operation", "user_processing"), func(ctx context.Context) {
                processUser(id)
            })
        }(userID)
    }

    // Request type labeling
    requestTypes := []string{"read", "write", "delete", "update"}
    for _, reqType := range requestTypes {
        go func(rt string) {
            ctx := pprof.WithLabels(context.Background(), pprof.Labels("request_type", rt))
            pprof.Do(ctx, pprof.Labels("operation", "request_handling"), func(ctx context.Context) {
                handleRequest(rt)
            })
        }(reqType)
    }

    // Nested operation labeling
    ctx := pprof.WithLabels(context.Background(), pprof.Labels("component", "data_processor"))
    pprof.Do(ctx, pprof.Labels("phase", "initialization"), func(ctx context.Context) {
        initialize()
        
        pprof.Do(ctx, pprof.Labels("phase", "processing"), func(ctx context.Context) {
            process()
            
            pprof.Do(ctx, pprof.Labels("phase", "cleanup"), func(ctx context.Context) {
                cleanup()
            })
        })
    })
}

func getUserType(userID int) string {
    if userID%10 == 0 {
        return "premium"
    }
    return "standard"
}

func getUserRegion(userID int) string {
    regions := []string{"us-east", "us-west", "eu-west", "ap-south"}
    return regions[userID%len(regions)]
}

func processUser(userID int) {
    // Simulate user processing work
    for i := 0; i < 100000; i++ {
        _ = userID * i
    }
}

func handleRequest(requestType string) {
    // Simulate request handling
    switch requestType {
    case "read":
        performRead()
    case "write":
        performWrite()
    case "delete":
        performDelete()
    case "update":
        performUpdate()
    }
}

func performRead() {
    for i := 0; i < 50000; i++ {
        _ = i * 2
    }
}

func performWrite() {
    for i := 0; i < 75000; i++ {
        _ = i * 3
    }
}

func performDelete() {
    for i := 0; i < 25000; i++ {
        _ = i * 4
    }
}

func performUpdate() {
    for i := 0; i < 100000; i++ {
        _ = i * 5
    }
}

func initialize() {
    for i := 0; i < 10000; i++ {
        _ = i * i
    }
}

func process() {
    for i := 0; i < 100000; i++ {
        _ = i * i * i
    }
}

func cleanup() {
    for i := 0; i < 5000; i++ {
        _ = i + i
    }
}
```

## Profile Analysis with Labels

```bash
# Analyze profiles with labels
go tool pprof -tagfocus=user_type:premium cpu.prof
go tool pprof -tagfocus=request_type:write cpu.prof
go tool pprof -taghide=user_type:standard cpu.prof

# Show label information
go tool pprof -tags cpu.prof

# Focus on specific operations
go tool pprof -tagfocus=operation:user_processing cpu.prof
```

## Advanced pprof Commands

### Statistical Analysis

```bash
# Sample-based analysis
go tool pprof -sample_index=cpu cpu.prof

# Show all sample locations
go tool pprof -traces cpu.prof

# Generate detailed call graph
go tool pprof -nodecount=1000 -call_tree cpu.prof

# Export to different formats
go tool pprof -proto cpu.prof > profile.pb.gz
go tool pprof -text cpu.prof > profile.txt
go tool pprof -dot cpu.prof > profile.dot
```

### Filtering and Focus

```bash
# Focus on specific functions
go tool pprof -focus="main\..*" cpu.prof

# Hide runtime functions
go tool pprof -hide="runtime\..*" cpu.prof

# Show only functions above threshold
go tool pprof -nodefraction=0.01 cpu.prof

# Ignore functions below threshold
go tool pprof -edgefraction=0.001 cpu.prof
```

## Continuous Profiling Integration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "runtime/pprof"
    "sync"
    "time"
)

type ContinuousProfiler struct {
    mu           sync.RWMutex
    enabled      bool
    interval     time.Duration
    profileDir   string
    retention    time.Duration
    activeFile   *os.File
    stopCh       chan struct{}
}

func NewContinuousProfiler(interval time.Duration, profileDir string, retention time.Duration) *ContinuousProfiler {
    return &ContinuousProfiler{
        interval:   interval,
        profileDir: profileDir,
        retention:  retention,
        stopCh:     make(chan struct{}),
    }
}

func (cp *ContinuousProfiler) Start() error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if cp.enabled {
        return fmt.Errorf("profiler already running")
    }

    // Ensure profile directory exists
    if err := os.MkdirAll(cp.profileDir, 0755); err != nil {
        return fmt.Errorf("failed to create profile directory: %v", err)
    }

    cp.enabled = true
    go cp.profileLoop()
    go cp.cleanupLoop()

    log.Println("Continuous profiler started")
    return nil
}

func (cp *ContinuousProfiler) Stop() {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if !cp.enabled {
        return
    }

    cp.enabled = false
    close(cp.stopCh)

    // Stop current profiling session
    if cp.activeFile != nil {
        pprof.StopCPUProfile()
        cp.activeFile.Close()
        cp.activeFile = nil
    }

    log.Println("Continuous profiler stopped")
}

func (cp *ContinuousProfiler) profileLoop() {
    ticker := time.NewTicker(cp.interval)
    defer ticker.Stop()

    for {
        select {
        case <-cp.stopCh:
            return
        case <-ticker.C:
            cp.createProfile()
        }
    }
}

func (cp *ContinuousProfiler) createProfile() {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if !cp.enabled {
        return
    }

    // Stop previous profile if running
    if cp.activeFile != nil {
        pprof.StopCPUProfile()
        cp.activeFile.Close()
    }

    // Create new profile file
    timestamp := time.Now().Format("20060102_150405")
    filename := filepath.Join(cp.profileDir, fmt.Sprintf("cpu_%s.prof", timestamp))
    
    f, err := os.Create(filename)
    if err != nil {
        log.Printf("Failed to create profile file %s: %v", filename, err)
        return
    }

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Printf("Failed to start CPU profile: %v", err)
        f.Close()
        return
    }

    cp.activeFile = f
    log.Printf("Started CPU profile: %s", filename)
}

func (cp *ContinuousProfiler) cleanupLoop() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-cp.stopCh:
            return
        case <-ticker.C:
            cp.cleanupOldProfiles()
        }
    }
}

func (cp *ContinuousProfiler) cleanupOldProfiles() {
    files, err := filepath.Glob(filepath.Join(cp.profileDir, "cpu_*.prof"))
    if err != nil {
        log.Printf("Failed to list profile files: %v", err)
        return
    }

    cutoff := time.Now().Add(-cp.retention)
    removed := 0

    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            continue
        }

        if info.ModTime().Before(cutoff) {
            if err := os.Remove(file); err != nil {
                log.Printf("Failed to remove old profile %s: %v", file, err)
            } else {
                removed++
            }
        }
    }

    if removed > 0 {
        log.Printf("Cleaned up %d old profile files", removed)
    }
}

// HTTP endpoint for profile management
func (cp *ContinuousProfiler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/profiles/start":
        if err := cp.Start(); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Profiler started"))

    case "/profiles/stop":
        cp.Stop()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Profiler stopped"))

    case "/profiles/list":
        cp.listProfiles(w, r)

    default:
        http.NotFound(w, r)
    }
}

func (cp *ContinuousProfiler) listProfiles(w http.ResponseWriter, r *http.Request) {
    files, err := filepath.Glob(filepath.Join(cp.profileDir, "cpu_*.prof"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    profiles := make([]map[string]interface{}, 0, len(files))
    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            continue
        }

        profiles = append(profiles, map[string]interface{}{
            "name":    filepath.Base(file),
            "path":    file,
            "size":    info.Size(),
            "created": info.ModTime().Format(time.RFC3339),
        })
    }

    fmt.Fprintf(w, `{"profiles": %+v}`, profiles)
}
```

## Performance Impact Analysis

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func measureProfilingOverhead() {
    // Baseline measurement
    baselineDuration := measureWorkload(false)
    
    // Profiled measurement
    profiledDuration := measureWorkload(true)
    
    overhead := float64(profiledDuration-baselineDuration) / float64(baselineDuration) * 100
    
    fmt.Printf("Baseline duration: %v\n", baselineDuration)
    fmt.Printf("Profiled duration: %v\n", profiledDuration)
    fmt.Printf("Overhead: %.2f%%\n", overhead)
}

func measureWorkload(enableProfiling bool) time.Duration {
    if enableProfiling {
        // Simulate profiling overhead
        runtime.SetCPUProfileRate(100)
    } else {
        runtime.SetCPUProfileRate(0)
    }

    start := time.Now()
    
    // Consistent workload
    for i := 0; i < 1000000; i++ {
        _ = fibonacci(20)
    }
    
    return time.Since(start)
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
```

## Best Practices for Advanced Profiling

### 1. Production Profiling Safety

```go
type SafeProfiler struct {
    maxConcurrent int
    current       int32
    mu            sync.RWMutex
}

func (sp *SafeProfiler) SafeProfile(fn func()) bool {
    sp.mu.Lock()
    defer sp.mu.Unlock()

    if atomic.LoadInt32(&sp.current) >= int32(sp.maxConcurrent) {
        return false // Too many concurrent profiles
    }

    atomic.AddInt32(&sp.current, 1)
    defer atomic.AddInt32(&sp.current, -1)

    fn()
    return true
}
```

### 2. Automated Analysis

```bash
#!/bin/bash
# analyze_profile.sh

PROFILE_FILE=$1
OUTPUT_DIR=$2

# Generate multiple views
go tool pprof -text $PROFILE_FILE > $OUTPUT_DIR/profile.txt
go tool pprof -svg $PROFILE_FILE > $OUTPUT_DIR/profile.svg
go tool pprof -top $PROFILE_FILE > $OUTPUT_DIR/top.txt
go tool pprof -traces $PROFILE_FILE > $OUTPUT_DIR/traces.txt

# Generate focused analysis
go tool pprof -focus="main\." -text $PROFILE_FILE > $OUTPUT_DIR/main_functions.txt
go tool pprof -hide="runtime\." -text $PROFILE_FILE > $OUTPUT_DIR/user_code.txt
```

### 3. Profile Comparison Automation

```go
func compareProfiles(before, after string) error {
    cmd := exec.Command("go", "tool", "pprof", "-base", before, "-text", after)
    output, err := cmd.Output()
    if err != nil {
        return err
    }

    // Parse comparison output
    lines := strings.Split(string(output), "\n")
    improvements := []string{}
    regressions := []string{}

    for _, line := range lines {
        if strings.Contains(line, "+") {
            regressions = append(regressions, line)
        } else if strings.Contains(line, "-") {
            improvements = append(improvements, line)
        }
    }

    fmt.Printf("Improvements: %d\n", len(improvements))
    fmt.Printf("Regressions: %d\n", len(regressions))

    return nil
}
```

## Next Steps

- Explore [Flamegraph Analysis](flamegraph-analysis.md) for visual profiling
- Learn [Memory Profiling](../memory-profiling/README.md) techniques
- Study [Goroutine Profiling](../goroutine-profiling/README.md) patterns

## Summary

Advanced CPU profiling techniques enable sophisticated performance analysis:

1. **Multi-profile comparison** for optimization validation
2. **Conditional profiling** for automatic performance monitoring
3. **Custom labels** for fine-grained analysis
4. **Continuous profiling** for production monitoring
5. **Sampling rate optimization** for accuracy vs. overhead balance

These techniques form the foundation for production-grade performance engineering and continuous optimization workflows.
