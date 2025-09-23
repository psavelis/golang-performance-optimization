# Memory Leak Detection

Memory leak detection is crucial for building robust Go applications that maintain stable memory usage over time. This comprehensive guide covers advanced techniques for detecting, analyzing, and preventing memory leaks in Go applications.

## Introduction to Memory Leak Detection

Memory leaks in Go can occur despite having a garbage collector. Common causes include:
- **Goroutine leaks** - Goroutines that never terminate
- **Reference cycles** - Objects that reference each other
- **Global variable accumulation** - Growing global data structures
- **Resource leaks** - Unclosed files, connections, or channels
- **Finalizer issues** - Objects with finalizers not being collected

### Understanding Go Memory Management

```go
package main

import (
    "fmt"
    "runtime"
    "runtime/debug"
    "time"
)

// MemoryStats provides detailed memory statistics
type MemoryStats struct {
    Timestamp    time.Time
    Alloc        uint64 // Current allocated memory
    TotalAlloc   uint64 // Total allocated memory
    Sys          uint64 // System memory
    Lookups      uint64 // Number of pointer lookups
    Mallocs      uint64 // Number of mallocs
    Frees        uint64 // Number of frees
    HeapAlloc    uint64 // Heap allocated memory
    HeapSys      uint64 // Heap system memory
    HeapIdle     uint64 // Idle heap memory
    HeapInuse    uint64 // In-use heap memory
    HeapReleased uint64 // Released heap memory
    HeapObjects  uint64 // Number of heap objects
    StackInuse   uint64 // Stack memory in use
    StackSys     uint64 // Stack system memory
    MSpanInuse   uint64 // MSpan structures in use
    MSpanSys     uint64 // MSpan system memory
    MCacheInuse  uint64 // MCache structures in use
    MCacheSys    uint64 // MCache system memory
    BuckHashSys  uint64 // Profiling bucket hash table memory
    GCSys        uint64 // GC metadata memory
    OtherSys     uint64 // Other system reservations
    NextGC       uint64 // Next GC cycle target
    LastGC       uint64 // Last GC time
    PauseTotalNs uint64 // Total GC pause time
    PauseNs      [256]uint64 // Recent GC pause times
    PauseEnd     [256]uint64 // Recent GC pause end times
    NumGC        uint32 // Number of GC cycles
    NumForcedGC  uint32 // Number of forced GC cycles
    GCCPUFraction float64 // Fraction of CPU time used by GC
    EnableGC     bool    // GC enabled flag
    DebugGC      bool    // GC debug flag
}

func GetMemoryStats() MemoryStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return MemoryStats{
        Timestamp:     time.Now(),
        Alloc:         m.Alloc,
        TotalAlloc:    m.TotalAlloc,
        Sys:           m.Sys,
        Lookups:       m.Lookups,
        Mallocs:       m.Mallocs,
        Frees:         m.Frees,
        HeapAlloc:     m.HeapAlloc,
        HeapSys:       m.HeapSys,
        HeapIdle:      m.HeapIdle,
        HeapInuse:     m.HeapInuse,
        HeapReleased:  m.HeapReleased,
        HeapObjects:   m.HeapObjects,
        StackInuse:    m.StackInuse,
        StackSys:      m.StackSys,
        MSpanInuse:    m.MSpanInuse,
        MSpanSys:      m.MSpanSys,
        MCacheInuse:   m.MCacheInuse,
        MCacheSys:     m.MCacheSys,
        BuckHashSys:   m.BuckHashSys,
        GCSys:         m.GCSys,
        OtherSys:      m.OtherSys,
        NextGC:        m.NextGC,
        LastGC:        m.LastGC,
        PauseTotalNs:  m.PauseTotalNs,
        PauseNs:       m.PauseNs,
        PauseEnd:      m.PauseEnd,
        NumGC:         m.NumGC,
        NumForcedGC:   m.NumForcedGC,
        GCCPUFraction: m.GCCPUFraction,
        EnableGC:      m.EnableGC,
        DebugGC:       m.DebugGC,
    }
}

func (ms MemoryStats) String() string {
    return fmt.Sprintf(`Memory Statistics (%s):
  Allocated: %s
  Total Allocated: %s
  System: %s
  Heap Allocated: %s
  Heap Objects: %d
  GC Cycles: %d
  GC CPU Fraction: %.4f
  Live Objects: %d (Mallocs: %d - Frees: %d)`,
        ms.Timestamp.Format("15:04:05"),
        formatBytes(ms.Alloc),
        formatBytes(ms.TotalAlloc),
        formatBytes(ms.Sys),
        formatBytes(ms.HeapAlloc),
        ms.HeapObjects,
        ms.NumGC,
        ms.GCCPUFraction,
        ms.Mallocs-ms.Frees,
        ms.Mallocs,
        ms.Frees)
}

func formatBytes(bytes uint64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )
    
    switch {
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}

// Memory trend analysis
type MemoryTrendAnalyzer struct {
    samples        []MemoryStats
    maxSamples     int
    alertThresholds AlertThresholds
}

type AlertThresholds struct {
    HeapGrowthRate    float64 // MB/minute
    ObjectGrowthRate  float64 // objects/minute
    GCFrequencyMax    float64 // GCs/minute
    HeapUtilization   float64 // percentage
}

func NewMemoryTrendAnalyzer(maxSamples int) *MemoryTrendAnalyzer {
    return &MemoryTrendAnalyzer{
        maxSamples: maxSamples,
        alertThresholds: AlertThresholds{
            HeapGrowthRate:   10.0,  // 10 MB/minute
            ObjectGrowthRate: 10000, // 10k objects/minute
            GCFrequencyMax:   60,    // 1 GC/second
            HeapUtilization:  0.9,   // 90%
        },
    }
}

func (mta *MemoryTrendAnalyzer) AddSample(stats MemoryStats) {
    mta.samples = append(mta.samples, stats)
    
    // Keep only recent samples
    if len(mta.samples) > mta.maxSamples {
        mta.samples = mta.samples[1:]
    }
}

func (mta *MemoryTrendAnalyzer) AnalyzeTrends() TrendAnalysis {
    if len(mta.samples) < 2 {
        return TrendAnalysis{Insufficient: true}
    }
    
    first := mta.samples[0]
    last := mta.samples[len(mta.samples)-1]
    duration := last.Timestamp.Sub(first.Timestamp)
    
    if duration == 0 {
        return TrendAnalysis{Insufficient: true}
    }
    
    minutes := duration.Minutes()
    
    // Calculate growth rates
    heapGrowth := float64(last.HeapAlloc-first.HeapAlloc) / (1024 * 1024) // MB
    objectGrowth := float64(int64(last.HeapObjects) - int64(first.HeapObjects))
    gcGrowth := float64(last.NumGC - first.NumGC)
    
    heapGrowthRate := heapGrowth / minutes
    objectGrowthRate := objectGrowth / minutes
    gcFrequency := gcGrowth / minutes
    
    // Calculate heap utilization
    heapUtilization := float64(last.HeapInuse) / float64(last.HeapSys)
    
    // Detect anomalies
    alerts := []string{}
    
    if heapGrowthRate > mta.alertThresholds.HeapGrowthRate {
        alerts = append(alerts, fmt.Sprintf("High heap growth rate: %.2f MB/min", heapGrowthRate))
    }
    
    if objectGrowthRate > mta.alertThresholds.ObjectGrowthRate {
        alerts = append(alerts, fmt.Sprintf("High object growth rate: %.0f objects/min", objectGrowthRate))
    }
    
    if gcFrequency > mta.alertThresholds.GCFrequencyMax {
        alerts = append(alerts, fmt.Sprintf("High GC frequency: %.1f GCs/min", gcFrequency))
    }
    
    if heapUtilization > mta.alertThresholds.HeapUtilization {
        alerts = append(alerts, fmt.Sprintf("High heap utilization: %.1f%%", heapUtilization*100))
    }
    
    return TrendAnalysis{
        Duration:         duration,
        HeapGrowthRate:   heapGrowthRate,
        ObjectGrowthRate: objectGrowthRate,
        GCFrequency:      gcFrequency,
        HeapUtilization:  heapUtilization,
        Alerts:           alerts,
        SampleCount:      len(mta.samples),
    }
}

type TrendAnalysis struct {
    Insufficient     bool
    Duration         time.Duration
    HeapGrowthRate   float64 // MB/minute
    ObjectGrowthRate float64 // objects/minute
    GCFrequency      float64 // GCs/minute
    HeapUtilization  float64 // percentage
    Alerts           []string
    SampleCount      int
}

func (ta TrendAnalysis) String() string {
    if ta.Insufficient {
        return "Insufficient data for trend analysis"
    }
    
    result := fmt.Sprintf(`Memory Trend Analysis (%d samples over %v):
  Heap Growth Rate: %.2f MB/min
  Object Growth Rate: %.0f objects/min
  GC Frequency: %.1f GCs/min
  Heap Utilization: %.1f%%`,
        ta.SampleCount,
        ta.Duration,
        ta.HeapGrowthRate,
        ta.ObjectGrowthRate,
        ta.GCFrequency,
        ta.HeapUtilization*100)
    
    if len(ta.Alerts) > 0 {
        result += "\n\nALERTS:"
        for _, alert := range ta.Alerts {
            result += "\n  ⚠️  " + alert
        }
    }
    
    return result
}

func demonstrateBasicMemoryMonitoring() {
    fmt.Println("=== BASIC MEMORY MONITORING ===")
    
    analyzer := NewMemoryTrendAnalyzer(10)
    
    // Simulate memory usage
    var data [][]byte
    
    for i := 0; i < 10; i++ {
        // Allocate some memory
        chunk := make([]byte, 1024*1024) // 1MB
        data = append(data, chunk)
        
        // Record memory stats
        stats := GetMemoryStats()
        analyzer.AddSample(stats)
        
        fmt.Printf("Step %d: %s\n", i+1, formatBytes(stats.HeapAlloc))
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Analyze trends
    analysis := analyzer.AnalyzeTrends()
    fmt.Println("\n" + analysis.String())
    
    // Keep data alive to prevent GC
    runtime.KeepAlive(data)
}
```

## Advanced Leak Detection Techniques

### Goroutine Leak Detection

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "strings"
    "sync"
    "time"
)

// GoroutineTracker monitors goroutine creation and lifecycle
type GoroutineTracker struct {
    mu               sync.RWMutex
    baseCount        int
    samples          []GoroutineSample
    alertThreshold   int
    alertCallback    func(GoroutineAlert)
    trackingEnabled  bool
}

type GoroutineSample struct {
    Timestamp time.Time
    Count     int
    Stacks    []string
}

type GoroutineAlert struct {
    Timestamp     time.Time
    Count         int
    BaseCount     int
    Increase      int
    LeakedStacks  []string
}

func NewGoroutineTracker() *GoroutineTracker {
    return &GoroutineTracker{
        baseCount:      runtime.NumGoroutine(),
        alertThreshold: 10, // Alert if 10+ goroutines increase
        trackingEnabled: true,
    }
}

func (gt *GoroutineTracker) SetAlertCallback(callback func(GoroutineAlert)) {
    gt.mu.Lock()
    defer gt.mu.Unlock()
    gt.alertCallback = callback
}

func (gt *GoroutineTracker) SetAlertThreshold(threshold int) {
    gt.mu.Lock()
    defer gt.mu.Unlock()
    gt.alertThreshold = threshold
}

func (gt *GoroutineTracker) StartTracking(interval time.Duration) context.CancelFunc {
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                gt.checkGoroutines()
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return cancel
}

func (gt *GoroutineTracker) checkGoroutines() {
    if !gt.trackingEnabled {
        return
    }
    
    gt.mu.Lock()
    defer gt.mu.Unlock()
    
    currentCount := runtime.NumGoroutine()
    stacks := gt.getGoroutineStacks()
    
    sample := GoroutineSample{
        Timestamp: time.Now(),
        Count:     currentCount,
        Stacks:    stacks,
    }
    
    gt.samples = append(gt.samples, sample)
    
    // Keep only recent samples
    if len(gt.samples) > 100 {
        gt.samples = gt.samples[1:]
    }
    
    // Check for potential leaks
    increase := currentCount - gt.baseCount
    if increase >= gt.alertThreshold && gt.alertCallback != nil {
        leakedStacks := gt.identifyLeakedStacks()
        
        alert := GoroutineAlert{
            Timestamp:    time.Now(),
            Count:        currentCount,
            BaseCount:    gt.baseCount,
            Increase:     increase,
            LeakedStacks: leakedStacks,
        }
        
        gt.alertCallback(alert)
    }
}

func (gt *GoroutineTracker) getGoroutineStacks() []string {
    buf := make([]byte, 1024*1024) // 1MB buffer
    stackSize := runtime.Stack(buf, true)
    stackData := string(buf[:stackSize])
    
    // Split by goroutine boundaries
    stacks := strings.Split(stackData, "\n\ngoroutine")
    
    // Clean up stacks
    var cleanStacks []string
    for i, stack := range stacks {
        if i > 0 {
            stack = "goroutine" + stack
        }
        if strings.TrimSpace(stack) != "" {
            cleanStacks = append(cleanStacks, strings.TrimSpace(stack))
        }
    }
    
    return cleanStacks
}

func (gt *GoroutineTracker) identifyLeakedStacks() []string {
    if len(gt.samples) < 2 {
        return nil
    }
    
    current := gt.samples[len(gt.samples)-1]
    baseline := gt.samples[0]
    
    // Find stacks that appear frequently in recent samples
    stackCounts := make(map[string]int)
    
    for _, stack := range current.Stacks {
        // Extract the function signature
        signature := gt.extractStackSignature(stack)
        if signature != "" {
            stackCounts[signature]++
        }
    }
    
    // Identify potentially leaked stack signatures
    var leakedStacks []string
    for signature, count := range stackCounts {
        if count >= 3 && gt.isLikelyLeak(signature) {
            leakedStacks = append(leakedStacks, signature)
        }
    }
    
    return leakedStacks
}

func (gt *GoroutineTracker) extractStackSignature(stack string) string {
    lines := strings.Split(stack, "\n")
    if len(lines) < 3 {
        return ""
    }
    
    // Look for the first non-runtime function call
    for i, line := range lines {
        if strings.Contains(line, "(") && !strings.Contains(line, "runtime.") {
            // Extract function name and file location
            parts := strings.Fields(line)
            if len(parts) > 0 {
                return parts[0]
            }
        }
    }
    
    return ""
}

func (gt *GoroutineTracker) isLikelyLeak(signature string) bool {
    // Common patterns that indicate potential leaks
    leakPatterns := []string{
        "time.Sleep",
        "chan.recv",
        "chan.send",
        "sync.WaitGroup.Wait",
        "sync.Mutex.Lock",
        "net.Conn.Read",
        "net.Conn.Write",
        "http.Client.Do",
    }
    
    for _, pattern := range leakPatterns {
        if strings.Contains(signature, pattern) {
            return true
        }
    }
    
    return false
}

func (gt *GoroutineTracker) GetReport() GoroutineReport {
    gt.mu.RLock()
    defer gt.mu.RUnlock()
    
    if len(gt.samples) == 0 {
        return GoroutineReport{NoData: true}
    }
    
    current := gt.samples[len(gt.samples)-1]
    
    return GoroutineReport{
        Timestamp:   current.Timestamp,
        CurrentCount: current.Count,
        BaseCount:   gt.baseCount,
        Increase:    current.Count - gt.baseCount,
        SampleCount: len(gt.samples),
        TopStacks:   gt.getTopStackSignatures(5),
    }
}

func (gt *GoroutineTracker) getTopStackSignatures(limit int) []StackSignature {
    if len(gt.samples) == 0 {
        return nil
    }
    
    current := gt.samples[len(gt.samples)-1]
    signatureCounts := make(map[string]int)
    
    for _, stack := range current.Stacks {
        signature := gt.extractStackSignature(stack)
        if signature != "" {
            signatureCounts[signature]++
        }
    }
    
    // Sort by count
    type sigCount struct {
        signature string
        count     int
    }
    
    var sigs []sigCount
    for sig, count := range signatureCounts {
        sigs = append(sigs, sigCount{sig, count})
    }
    
    // Simple bubble sort for small data
    for i := 0; i < len(sigs)-1; i++ {
        for j := 0; j < len(sigs)-i-1; j++ {
            if sigs[j].count < sigs[j+1].count {
                sigs[j], sigs[j+1] = sigs[j+1], sigs[j]
            }
        }
    }
    
    var result []StackSignature
    for i := 0; i < len(sigs) && i < limit; i++ {
        result = append(result, StackSignature{
            Signature: sigs[i].signature,
            Count:     sigs[i].count,
        })
    }
    
    return result
}

type GoroutineReport struct {
    NoData       bool
    Timestamp    time.Time
    CurrentCount int
    BaseCount    int
    Increase     int
    SampleCount  int
    TopStacks    []StackSignature
}

type StackSignature struct {
    Signature string
    Count     int
}

func (gr GoroutineReport) String() string {
    if gr.NoData {
        return "No goroutine tracking data available"
    }
    
    result := fmt.Sprintf(`Goroutine Report (%s):
  Current Count: %d
  Baseline Count: %d
  Increase: %d
  Samples Collected: %d`,
        gr.Timestamp.Format("15:04:05"),
        gr.CurrentCount,
        gr.BaseCount,
        gr.Increase,
        gr.SampleCount)
    
    if len(gr.TopStacks) > 0 {
        result += "\n\nTop Stack Signatures:"
        for _, stack := range gr.TopStacks {
            result += fmt.Sprintf("\n  %s: %d goroutines", stack.Signature, stack.Count)
        }
    }
    
    return result
}

// Leak simulation for testing
func simulateGoroutineLeak(count int) {
    for i := 0; i < count; i++ {
        go func(id int) {
            // Simulate a goroutine that blocks indefinitely
            ch := make(chan bool)
            <-ch // This will block forever
        }(i)
    }
}

func simulateChannelLeak(count int) {
    for i := 0; i < count; i++ {
        go func(id int) {
            ch := make(chan int, 1)
            for {
                select {
                case ch <- id:
                    time.Sleep(100 * time.Millisecond)
                default:
                    time.Sleep(50 * time.Millisecond)
                }
            }
        }(i)
    }
}

func demonstrateGoroutineLeakDetection() {
    fmt.Println("\n=== GOROUTINE LEAK DETECTION ===")
    
    tracker := NewGoroutineTracker()
    tracker.SetAlertThreshold(5)
    
    // Set up alert callback
    tracker.SetAlertCallback(func(alert GoroutineAlert) {
        fmt.Printf("\n🚨 GOROUTINE LEAK ALERT 🚨\n")
        fmt.Printf("Count: %d (baseline: %d, increase: %d)\n", 
            alert.Count, alert.BaseCount, alert.Increase)
        
        if len(alert.LeakedStacks) > 0 {
            fmt.Printf("Potential leak sources:\n")
            for _, stack := range alert.LeakedStacks {
                fmt.Printf("  - %s\n", stack)
            }
        }
        fmt.Println()
    })
    
    // Start tracking
    cancel := tracker.StartTracking(500 * time.Millisecond)
    defer cancel()
    
    // Show initial state
    time.Sleep(1 * time.Second)
    fmt.Println(tracker.GetReport().String())
    
    // Simulate normal goroutine creation
    fmt.Println("\nCreating normal goroutines...")
    for i := 0; i < 3; i++ {
        go func(id int) {
            time.Sleep(2 * time.Second)
        }(i)
    }
    
    time.Sleep(1 * time.Second)
    fmt.Println(tracker.GetReport().String())
    
    // Simulate goroutine leak
    fmt.Println("\nSimulating goroutine leak...")
    simulateGoroutineLeak(8)
    
    time.Sleep(2 * time.Second)
    fmt.Println(tracker.GetReport().String())
    
    // Simulate channel leak
    fmt.Println("\nSimulating channel-based leak...")
    simulateChannelLeak(3)
    
    time.Sleep(2 * time.Second)
    fmt.Println(tracker.GetReport().String())
}
```

### Memory Reference Analysis

```go
package main

import (
    "fmt"
    "reflect"
    "runtime"
    "unsafe"
)

// ReferenceTracker analyzes object references and potential cycles
type ReferenceTracker struct {
    tracked map[uintptr]*ObjectInfo
    roots   []uintptr
}

type ObjectInfo struct {
    Address    uintptr
    Type       reflect.Type
    Size       uintptr
    References []uintptr
    RefCount   int
    Reachable  bool
}

func NewReferenceTracker() *ReferenceTracker {
    return &ReferenceTracker{
        tracked: make(map[uintptr]*ObjectInfo),
    }
}

func (rt *ReferenceTracker) TrackObject(obj interface{}) {
    if obj == nil {
        return
    }
    
    v := reflect.ValueOf(obj)
    if v.Kind() == reflect.Ptr && !v.IsNil() {
        addr := v.Pointer()
        
        info := &ObjectInfo{
            Address: addr,
            Type:    v.Elem().Type(),
            Size:    v.Elem().Type().Size(),
        }
        
        rt.tracked[addr] = info
        rt.analyzeReferences(v.Elem(), info)
    }
}

func (rt *ReferenceTracker) analyzeReferences(v reflect.Value, info *ObjectInfo) {
    switch v.Kind() {
    case reflect.Ptr:
        if !v.IsNil() {
            refAddr := v.Pointer()
            info.References = append(info.References, refAddr)
        }
        
    case reflect.Slice:
        for i := 0; i < v.Len(); i++ {
            rt.analyzeReferences(v.Index(i), info)
        }
        
    case reflect.Array:
        for i := 0; i < v.Len(); i++ {
            rt.analyzeReferences(v.Index(i), info)
        }
        
    case reflect.Struct:
        for i := 0; i < v.NumField(); i++ {
            field := v.Field(i)
            if field.CanInterface() {
                rt.analyzeReferences(field, info)
            }
        }
        
    case reflect.Map:
        for _, key := range v.MapKeys() {
            rt.analyzeReferences(key, info)
            rt.analyzeReferences(v.MapIndex(key), info)
        }
        
    case reflect.Interface:
        if !v.IsNil() {
            rt.analyzeReferences(v.Elem(), info)
        }
    }
}

func (rt *ReferenceTracker) FindCycles() [][]uintptr {
    var cycles [][]uintptr
    visited := make(map[uintptr]bool)
    recStack := make(map[uintptr]bool)
    
    for addr := range rt.tracked {
        if !visited[addr] {
            if cycle := rt.dfsForCycle(addr, visited, recStack, []uintptr{}); cycle != nil {
                cycles = append(cycles, cycle)
            }
        }
    }
    
    return cycles
}

func (rt *ReferenceTracker) dfsForCycle(addr uintptr, visited, recStack map[uintptr]bool, path []uintptr) []uintptr {
    visited[addr] = true
    recStack[addr] = true
    path = append(path, addr)
    
    info, exists := rt.tracked[addr]
    if !exists {
        recStack[addr] = false
        return nil
    }
    
    for _, refAddr := range info.References {
        if !visited[refAddr] {
            if cycle := rt.dfsForCycle(refAddr, visited, recStack, path); cycle != nil {
                return cycle
            }
        } else if recStack[refAddr] {
            // Found cycle - return the cycle portion
            cycleStart := -1
            for i, p := range path {
                if p == refAddr {
                    cycleStart = i
                    break
                }
            }
            if cycleStart >= 0 {
                return append(path[cycleStart:], refAddr)
            }
        }
    }
    
    recStack[addr] = false
    return nil
}

func (rt *ReferenceTracker) GetReport() ReferenceReport {
    totalObjects := len(rt.tracked)
    totalSize := uintptr(0)
    cycles := rt.FindCycles()
    
    for _, info := range rt.tracked {
        totalSize += info.Size
    }
    
    return ReferenceReport{
        TotalObjects: totalObjects,
        TotalSize:    totalSize,
        Cycles:       cycles,
        Objects:      rt.tracked,
    }
}

type ReferenceReport struct {
    TotalObjects int
    TotalSize    uintptr
    Cycles       [][]uintptr
    Objects      map[uintptr]*ObjectInfo
}

func (rr ReferenceReport) String() string {
    result := fmt.Sprintf(`Reference Analysis Report:
  Total Objects: %d
  Total Size: %s
  Reference Cycles: %d`,
        rr.TotalObjects,
        formatBytes(uint64(rr.TotalSize)),
        len(rr.Cycles))
    
    if len(rr.Cycles) > 0 {
        result += "\n\nDetected Cycles:"
        for i, cycle := range rr.Cycles {
            result += fmt.Sprintf("\n  Cycle %d: %d objects", i+1, len(cycle))
            for _, addr := range cycle {
                if info, exists := rr.Objects[addr]; exists {
                    result += fmt.Sprintf("\n    %s (0x%x)", info.Type.String(), addr)
                }
            }
        }
    }
    
    return result
}

// Example objects for testing reference cycles
type Node struct {
    Value int
    Next  *Node
    Prev  *Node
    Data  []byte
}

type Container struct {
    Items []*Item
    Owner *Owner
}

type Item struct {
    ID        int
    Container *Container
    Related   []*Item
}

type Owner struct {
    Name       string
    Containers []*Container
}

func demonstrateReferenceAnalysis() {
    fmt.Println("\n=== REFERENCE CYCLE ANALYSIS ===")
    
    tracker := NewReferenceTracker()
    
    // Create objects with potential cycles
    
    // Simple circular linked list
    node1 := &Node{Value: 1, Data: make([]byte, 1024)}
    node2 := &Node{Value: 2, Data: make([]byte, 1024)}
    node3 := &Node{Value: 3, Data: make([]byte, 1024)}
    
    node1.Next = node2
    node2.Next = node3
    node3.Next = node1 // Creates cycle
    
    node1.Prev = node3
    node2.Prev = node1
    node3.Prev = node2
    
    tracker.TrackObject(node1)
    tracker.TrackObject(node2)
    tracker.TrackObject(node3)
    
    // Complex object hierarchy with cycles
    owner := &Owner{Name: "Owner1"}
    container := &Container{Owner: owner}
    owner.Containers = []*Container{container}
    
    item1 := &Item{ID: 1, Container: container}
    item2 := &Item{ID: 2, Container: container}
    item1.Related = []*Item{item2}
    item2.Related = []*Item{item1}
    
    container.Items = []*Item{item1, item2}
    
    tracker.TrackObject(owner)
    tracker.TrackObject(container)
    tracker.TrackObject(item1)
    tracker.TrackObject(item2)
    
    // Generate report
    report := tracker.GetReport()
    fmt.Println(report.String())
    
    // Keep objects alive
    runtime.KeepAlive(node1)
    runtime.KeepAlive(owner)
}
```

## Automated Leak Detection Tools

### Continuous Memory Monitor

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// MemoryMonitor provides continuous memory leak detection
type MemoryMonitor struct {
    config           MonitorConfig
    trendAnalyzer    *MemoryTrendAnalyzer
    goroutineTracker *GoroutineTracker
    alertManager     *AlertManager
    running          bool
    mu               sync.RWMutex
    cancel           context.CancelFunc
}

type MonitorConfig struct {
    SampleInterval    time.Duration
    AlertThresholds   AlertThresholds
    LogFile          string
    WebServerPort    int
    AlertWebhook     string
    MaxLogSize       int64
}

type AlertManager struct {
    webhookURL    string
    alertHistory  []Alert
    mu            sync.RWMutex
    maxHistory    int
}

type Alert struct {
    Timestamp   time.Time    `json:"timestamp"`
    Type        string       `json:"type"`
    Severity    string       `json:"severity"`
    Message     string       `json:"message"`
    Details     interface{}  `json:"details"`
    Resolved    bool         `json:"resolved"`
}

func NewMemoryMonitor(config MonitorConfig) *MemoryMonitor {
    return &MemoryMonitor{
        config:           config,
        trendAnalyzer:    NewMemoryTrendAnalyzer(100),
        goroutineTracker: NewGoroutineTracker(),
        alertManager:     NewAlertManager(config.AlertWebhook, 1000),
    }
}

func NewAlertManager(webhookURL string, maxHistory int) *AlertManager {
    return &AlertManager{
        webhookURL: webhookURL,
        maxHistory: maxHistory,
    }
}

func (am *AlertManager) SendAlert(alert Alert) {
    am.mu.Lock()
    am.alertHistory = append(am.alertHistory, alert)
    
    // Keep only recent alerts
    if len(am.alertHistory) > am.maxHistory {
        am.alertHistory = am.alertHistory[1:]
    }
    am.mu.Unlock()
    
    // Log alert
    log.Printf("ALERT [%s]: %s - %s", alert.Severity, alert.Type, alert.Message)
    
    // Send webhook if configured
    if am.webhookURL != "" {
        go am.sendWebhook(alert)
    }
}

func (am *AlertManager) sendWebhook(alert Alert) {
    data, err := json.Marshal(alert)
    if err != nil {
        log.Printf("Failed to marshal alert: %v", err)
        return
    }
    
    resp, err := http.Post(am.webhookURL, "application/json", strings.NewReader(string(data)))
    if err != nil {
        log.Printf("Failed to send webhook: %v", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        log.Printf("Webhook returned error status: %d", resp.StatusCode)
    }
}

func (am *AlertManager) GetAlerts() []Alert {
    am.mu.RLock()
    defer am.mu.RUnlock()
    
    alerts := make([]Alert, len(am.alertHistory))
    copy(alerts, am.alertHistory)
    return alerts
}

func (mm *MemoryMonitor) Start() error {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    if mm.running {
        return fmt.Errorf("monitor already running")
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    mm.cancel = cancel
    mm.running = true
    
    // Set up trend analyzer alerts
    mm.trendAnalyzer.alertThresholds = mm.config.AlertThresholds
    
    // Set up goroutine tracker alerts
    mm.goroutineTracker.SetAlertCallback(func(alert GoroutineAlert) {
        mm.alertManager.SendAlert(Alert{
            Timestamp: alert.Timestamp,
            Type:      "goroutine_leak",
            Severity:  "warning",
            Message:   fmt.Sprintf("Potential goroutine leak: %d goroutines (+%d from baseline)", alert.Count, alert.Increase),
            Details:   alert,
        })
    })
    
    // Start monitoring goroutines
    mm.goroutineTracker.StartTracking(mm.config.SampleInterval)
    
    // Start memory monitoring
    go mm.monitorLoop(ctx)
    
    // Start web server if configured
    if mm.config.WebServerPort > 0 {
        go mm.startWebServer(ctx)
    }
    
    return nil
}

func (mm *MemoryMonitor) Stop() {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    if !mm.running {
        return
    }
    
    mm.running = false
    if mm.cancel != nil {
        mm.cancel()
    }
}

func (mm *MemoryMonitor) monitorLoop(ctx context.Context) {
    ticker := time.NewTicker(mm.config.SampleInterval)
    defer ticker.Stop()
    
    logFile := mm.openLogFile()
    defer logFile.Close()
    
    for {
        select {
        case <-ticker.C:
            mm.collectSample(logFile)
        case <-ctx.Done():
            return
        }
    }
}

func (mm *MemoryMonitor) collectSample(logFile *os.File) {
    stats := GetMemoryStats()
    mm.trendAnalyzer.AddSample(stats)
    
    // Analyze trends
    analysis := mm.trendAnalyzer.AnalyzeTrends()
    
    // Log sample
    if logFile != nil {
        logEntry := map[string]interface{}{
            "timestamp": stats.Timestamp,
            "memory":    stats,
            "analysis":  analysis,
        }
        
        data, _ := json.Marshal(logEntry)
        logFile.WriteString(string(data) + "\n")
    }
    
    // Check for alerts
    for _, alertMsg := range analysis.Alerts {
        severity := "warning"
        if analysis.HeapGrowthRate > mm.config.AlertThresholds.HeapGrowthRate*2 {
            severity = "critical"
        }
        
        mm.alertManager.SendAlert(Alert{
            Timestamp: time.Now(),
            Type:      "memory_leak",
            Severity:  severity,
            Message:   alertMsg,
            Details:   analysis,
        })
    }
}

func (mm *MemoryMonitor) openLogFile() *os.File {
    if mm.config.LogFile == "" {
        return nil
    }
    
    // Create directory if needed
    dir := filepath.Dir(mm.config.LogFile)
    os.MkdirAll(dir, 0755)
    
    // Open log file
    file, err := os.OpenFile(mm.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("Failed to open log file: %v", err)
        return nil
    }
    
    // Check file size
    if info, err := file.Stat(); err == nil && info.Size() > mm.config.MaxLogSize {
        // Rotate log file
        file.Close()
        os.Rename(mm.config.LogFile, mm.config.LogFile+".old")
        file, _ = os.OpenFile(mm.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    }
    
    return file
}

func (mm *MemoryMonitor) startWebServer(ctx context.Context) {
    mux := http.NewServeMux()
    
    // Status endpoint
    mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        mm.mu.RLock()
        running := mm.running
        mm.mu.RUnlock()
        
        status := map[string]interface{}{
            "running":          running,
            "memory_stats":     GetMemoryStats(),
            "goroutine_report": mm.goroutineTracker.GetReport(),
            "trend_analysis":   mm.trendAnalyzer.AnalyzeTrends(),
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(status)
    })
    
    // Alerts endpoint
    mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
        alerts := mm.alertManager.GetAlerts()
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "alerts": alerts,
            "count":  len(alerts),
        })
    })
    
    // Force GC endpoint
    mux.HandleFunc("/gc", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        runtime.GC()
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "GC triggered",
        })
    })
    
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", mm.config.WebServerPort),
        Handler: mux,
    }
    
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()
    
    log.Printf("Memory monitor web server starting on port %d", mm.config.WebServerPort)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Printf("Web server error: %v", err)
    }
}

func (mm *MemoryMonitor) GetDashboardData() map[string]interface{} {
    return map[string]interface{}{
        "memory_stats":     GetMemoryStats(),
        "goroutine_report": mm.goroutineTracker.GetReport(),
        "trend_analysis":   mm.trendAnalyzer.AnalyzeTrends(),
        "recent_alerts":    mm.alertManager.GetAlerts(),
    }
}

func demonstrateAutomatedMonitoring() {
    fmt.Println("\n=== AUTOMATED MEMORY MONITORING ===")
    
    config := MonitorConfig{
        SampleInterval: 1 * time.Second,
        AlertThresholds: AlertThresholds{
            HeapGrowthRate:   5.0,  // 5 MB/minute
            ObjectGrowthRate: 5000, // 5k objects/minute
            GCFrequencyMax:   30,   // 30 GCs/minute
            HeapUtilization:  0.85, // 85%
        },
        LogFile:       "/tmp/memory_monitor.log",
        WebServerPort: 8080,
        MaxLogSize:    10 * 1024 * 1024, // 10MB
    }
    
    monitor := NewMemoryMonitor(config)
    
    if err := monitor.Start(); err != nil {
        log.Fatalf("Failed to start monitor: %v", err)
    }
    defer monitor.Stop()
    
    fmt.Printf("Memory monitor started. Web interface: http://localhost:%d/status\n", config.WebServerPort)
    
    // Simulate memory usage patterns
    fmt.Println("Simulating memory usage...")
    
    var data [][]byte
    
    // Gradual memory growth
    for i := 0; i < 10; i++ {
        chunk := make([]byte, 1024*1024) // 1MB chunks
        data = append(data, chunk)
        
        fmt.Printf("Allocated %d MB\n", i+1)
        time.Sleep(2 * time.Second)
    }
    
    // Simulate goroutine leak
    fmt.Println("Simulating goroutine leak...")
    simulateGoroutineLeak(10)
    
    // Let monitor collect data
    time.Sleep(10 * time.Second)
    
    // Show dashboard data
    dashboard := monitor.GetDashboardData()
    dashboardJSON, _ := json.MarshalIndent(dashboard, "", "  ")
    fmt.Printf("\nDashboard Data:\n%s\n", dashboardJSON)
    
    // Keep data alive
    runtime.KeepAlive(data)
}
```

## Next Steps

- Learn [Profiling Integration](../profiling-tools/integration.md) techniques
- Study [Performance Benchmarking](../benchmarking/README.md) methods
- Explore [Production Monitoring](../monitoring/README.md) strategies
- Master [Automated Testing](../testing/README.md) for memory leaks

## Summary

Memory leak detection in Go requires:

1. **Understanding leak patterns** - Goroutine, reference, and resource leaks
2. **Implementing monitoring** - Continuous tracking of memory trends
3. **Analyzing references** - Detecting cycles and unreachable objects
4. **Automated detection** - Building tools for production monitoring
5. **Proactive prevention** - Designing leak-resistant code patterns

Use these techniques to build robust applications that maintain stable memory usage over time.
