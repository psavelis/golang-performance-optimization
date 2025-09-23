# Allocation Profiling

Allocation profiling in Go tracks memory allocation patterns to identify excessive allocations, allocation hot spots, and optimization opportunities. This comprehensive guide covers advanced allocation profiling techniques for memory-efficient Go applications.

## Introduction to Allocation Profiling

Allocation profiling helps identify:
- **Allocation hot spots** - Functions creating the most allocations
- **Allocation frequency** - Rate of memory allocations
- **Allocation patterns** - Types and sizes of allocated objects
- **Short-lived allocations** - Objects that are quickly garbage collected
- **Allocation call paths** - Source of allocations in your code

### Understanding Go Memory Allocations

```go
package main

import (
    "fmt"
    "runtime"
    "sort"
    "sync"
    "time"
)

// AllocationTracker monitors memory allocation patterns
type AllocationTracker struct {
    baseline    runtime.MemStats
    samples     []AllocationSample
    mu          sync.RWMutex
    isTracking  bool
    interval    time.Duration
}

type AllocationSample struct {
    Timestamp    time.Time
    TotalAlloc   uint64
    Mallocs      uint64
    Frees        uint64
    LiveObjects  uint64
    HeapAlloc    uint64
    HeapObjects  uint64
    GCCycles     uint32
    AllocRate    float64 // Bytes per second
    ObjectRate   float64 // Objects per second
}

func NewAllocationTracker(interval time.Duration) *AllocationTracker {
    tracker := &AllocationTracker{
        interval: interval,
    }
    
    // Capture baseline
    runtime.ReadMemStats(&tracker.baseline)
    
    return tracker
}

func (at *AllocationTracker) StartTracking() {
    at.mu.Lock()
    defer at.mu.Unlock()
    
    if at.isTracking {
        return
    }
    
    at.isTracking = true
    go at.trackingLoop()
}

func (at *AllocationTracker) StopTracking() {
    at.mu.Lock()
    defer at.mu.Unlock()
    at.isTracking = false
}

func (at *AllocationTracker) trackingLoop() {
    ticker := time.NewTicker(at.interval)
    defer ticker.Stop()
    
    var lastSample *AllocationSample
    
    for {
        at.mu.RLock()
        isTracking := at.isTracking
        at.mu.RUnlock()
        
        if !isTracking {
            break
        }
        
        sample := at.captureSample()
        
        // Calculate rates if we have a previous sample
        if lastSample != nil {
            duration := sample.Timestamp.Sub(lastSample.Timestamp).Seconds()
            if duration > 0 {
                sample.AllocRate = float64(sample.TotalAlloc-lastSample.TotalAlloc) / duration
                sample.ObjectRate = float64(sample.Mallocs-lastSample.Mallocs) / duration
            }
        }
        
        at.mu.Lock()
        at.samples = append(at.samples, sample)
        // Keep only recent samples
        if len(at.samples) > 1000 {
            at.samples = at.samples[1:]
        }
        at.mu.Unlock()
        
        lastSample = &sample
        
        select {
        case <-ticker.C:
        }
    }
}

func (at *AllocationTracker) captureSample() AllocationSample {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return AllocationSample{
        Timestamp:   time.Now(),
        TotalAlloc:  m.TotalAlloc,
        Mallocs:     m.Mallocs,
        Frees:       m.Frees,
        LiveObjects: m.Mallocs - m.Frees,
        HeapAlloc:   m.HeapAlloc,
        HeapObjects: m.HeapObjects,
        GCCycles:    m.NumGC,
    }
}

func (at *AllocationTracker) GetSamples() []AllocationSample {
    at.mu.RLock()
    defer at.mu.RUnlock()
    
    samples := make([]AllocationSample, len(at.samples))
    copy(samples, at.samples)
    return samples
}

func (at *AllocationTracker) GetAllocationReport() AllocationReport {
    samples := at.GetSamples()
    
    if len(samples) == 0 {
        return AllocationReport{NoData: true}
    }
    
    first := samples[0]
    last := samples[len(samples)-1]
    duration := last.Timestamp.Sub(first.Timestamp)
    
    // Calculate overall rates
    totalAllocated := last.TotalAlloc - first.TotalAlloc
    totalObjects := last.Mallocs - first.Mallocs
    
    var avgAllocRate, avgObjectRate float64
    if duration.Seconds() > 0 {
        avgAllocRate = float64(totalAllocated) / duration.Seconds()
        avgObjectRate = float64(totalObjects) / duration.Seconds()
    }
    
    // Find peak rates
    var peakAllocRate, peakObjectRate float64
    for _, sample := range samples {
        if sample.AllocRate > peakAllocRate {
            peakAllocRate = sample.AllocRate
        }
        if sample.ObjectRate > peakObjectRate {
            peakObjectRate = sample.ObjectRate
        }
    }
    
    return AllocationReport{
        Duration:        duration,
        SampleCount:     len(samples),
        TotalAllocated:  totalAllocated,
        TotalObjects:    totalObjects,
        AvgAllocRate:    avgAllocRate,
        AvgObjectRate:   avgObjectRate,
        PeakAllocRate:   peakAllocRate,
        PeakObjectRate:  peakObjectRate,
        CurrentAlloc:    last.HeapAlloc,
        CurrentObjects:  last.HeapObjects,
        GCCycles:        last.GCCycles - first.GCCycles,
    }
}

type AllocationReport struct {
    NoData          bool
    Duration        time.Duration
    SampleCount     int
    TotalAllocated  uint64
    TotalObjects    uint64
    AvgAllocRate    float64 // Bytes per second
    AvgObjectRate   float64 // Objects per second
    PeakAllocRate   float64
    PeakObjectRate  float64
    CurrentAlloc    uint64
    CurrentObjects  uint64
    GCCycles        uint32
}

func (ar AllocationReport) String() string {
    if ar.NoData {
        return "No allocation data available"
    }
    
    return fmt.Sprintf(`Allocation Report (%v):
  Total Allocated: %s (%d objects)
  Average Allocation Rate: %s/sec (%.0f objects/sec)
  Peak Allocation Rate: %s/sec (%.0f objects/sec)
  Current Heap: %s (%d objects)
  GC Cycles: %d
  Samples: %d`,
        ar.Duration,
        formatBytes(ar.TotalAllocated), ar.TotalObjects,
        formatBytes(uint64(ar.AvgAllocRate)), ar.AvgObjectRate,
        formatBytes(uint64(ar.PeakAllocRate)), ar.PeakObjectRate,
        formatBytes(ar.CurrentAlloc), ar.CurrentObjects,
        ar.GCCycles,
        ar.SampleCount)
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

func demonstrateAllocationTracking() {
    fmt.Println("=== ALLOCATION TRACKING DEMONSTRATION ===")
    
    tracker := NewAllocationTracker(100 * time.Millisecond)
    tracker.StartTracking()
    defer tracker.StopTracking()
    
    // Simulate different allocation patterns
    
    // Pattern 1: Small frequent allocations
    fmt.Println("Testing small frequent allocations...")
    for i := 0; i < 1000; i++ {
        data := make([]byte, 64) // Small allocations
        _ = data
        if i%100 == 0 {
            time.Sleep(10 * time.Millisecond)
        }
    }
    
    time.Sleep(500 * time.Millisecond) // Let tracker collect data
    
    // Pattern 2: Large infrequent allocations
    fmt.Println("Testing large infrequent allocations...")
    for i := 0; i < 10; i++ {
        data := make([]byte, 1024*1024) // 1MB allocations
        _ = data
        time.Sleep(100 * time.Millisecond)
    }
    
    time.Sleep(500 * time.Millisecond)
    
    // Pattern 3: Burst allocations
    fmt.Println("Testing burst allocations...")
    for burst := 0; burst < 5; burst++ {
        for i := 0; i < 100; i++ {
            data := make([]int, 1000)
            _ = data
        }
        time.Sleep(200 * time.Millisecond)
    }
    
    time.Sleep(1 * time.Second) // Final data collection
    
    // Show report
    report := tracker.GetAllocationReport()
    fmt.Printf("\n%s\n", report)
}
```

## Advanced Allocation Analysis

### Allocation Hot Spot Detection

```go
package main

import (
    "fmt"
    "runtime"
    "runtime/pprof"
    "sort"
    "strings"
    "sync"
    "time"
)

// AllocationHotSpotAnalyzer identifies functions with high allocation rates
type AllocationHotSpotAnalyzer struct {
    samples      map[string]*HotSpotData
    mu           sync.RWMutex
    enabled      bool
    sampleRate   int
}

type HotSpotData struct {
    FunctionName string
    Allocations  int64
    TotalBytes   int64
    Samples      int64
    AvgSize      float64
    FirstSeen    time.Time
    LastSeen     time.Time
}

func NewAllocationHotSpotAnalyzer() *AllocationHotSpotAnalyzer {
    return &AllocationHotSpotAnalyzer{
        samples:    make(map[string]*HotSpotData),
        sampleRate: 1, // Sample every allocation
    }
}

func (ahsa *AllocationHotSpotAnalyzer) Enable() {
    ahsa.mu.Lock()
    defer ahsa.mu.Unlock()
    
    if ahsa.enabled {
        return
    }
    
    // Set allocation profiling rate
    runtime.MemProfileRate = ahsa.sampleRate
    ahsa.enabled = true
}

func (ahsa *AllocationHotSpotAnalyzer) Disable() {
    ahsa.mu.Lock()
    defer ahsa.mu.Unlock()
    
    // Disable allocation profiling
    runtime.MemProfileRate = 0
    ahsa.enabled = false
}

func (ahsa *AllocationHotSpotAnalyzer) CaptureProfile() error {
    if !ahsa.enabled {
        return fmt.Errorf("analyzer not enabled")
    }
    
    // Force GC to get accurate numbers
    runtime.GC()
    
    // Capture allocation profile
    profile := pprof.Lookup("allocs")
    if profile == nil {
        return fmt.Errorf("allocation profile not available")
    }
    
    // In a real implementation, you would parse the profile data
    // This is a simplified version that demonstrates the concept
    ahsa.parseProfileData()
    
    return nil
}

func (ahsa *AllocationHotSpotAnalyzer) parseProfileData() {
    // Simulate parsing profile data
    // In practice, you would parse the actual pprof data
    
    ahsa.mu.Lock()
    defer ahsa.mu.Unlock()
    
    now := time.Now()
    
    // Simulate some hot spot data
    hotSpots := map[string]struct {
        allocs int64
        bytes  int64
    }{
        "main.allocateSlices":    {1000, 1024000},
        "main.createStructs":     {500, 50000},
        "main.stringOperations":  {2000, 100000},
        "runtime.makeslice":      {800, 800000},
        "runtime.newobject":      {1500, 75000},
    }
    
    for funcName, data := range hotSpots {
        if existing, exists := ahsa.samples[funcName]; exists {
            existing.Allocations += data.allocs
            existing.TotalBytes += data.bytes
            existing.Samples++
            existing.LastSeen = now
            existing.AvgSize = float64(existing.TotalBytes) / float64(existing.Allocations)
        } else {
            ahsa.samples[funcName] = &HotSpotData{
                FunctionName: funcName,
                Allocations:  data.allocs,
                TotalBytes:   data.bytes,
                Samples:      1,
                AvgSize:      float64(data.bytes) / float64(data.allocs),
                FirstSeen:    now,
                LastSeen:     now,
            }
        }
    }
}

func (ahsa *AllocationHotSpotAnalyzer) GetTopAllocators(limit int) []HotSpotData {
    ahsa.mu.RLock()
    defer ahsa.mu.RUnlock()
    
    var hotSpots []HotSpotData
    for _, data := range ahsa.samples {
        hotSpots = append(hotSpots, *data)
    }
    
    // Sort by total bytes allocated
    sort.Slice(hotSpots, func(i, j int) bool {
        return hotSpots[i].TotalBytes > hotSpots[j].TotalBytes
    })
    
    if limit > 0 && len(hotSpots) > limit {
        hotSpots = hotSpots[:limit]
    }
    
    return hotSpots
}

func (ahsa *AllocationHotSpotAnalyzer) GetReport() HotSpotReport {
    hotSpots := ahsa.GetTopAllocators(0)
    
    var totalAllocations int64
    var totalBytes int64
    
    for _, hs := range hotSpots {
        totalAllocations += hs.Allocations
        totalBytes += hs.TotalBytes
    }
    
    return HotSpotReport{
        TotalHotSpots:    len(hotSpots),
        TotalAllocations: totalAllocations,
        TotalBytes:       totalBytes,
        HotSpots:         hotSpots,
    }
}

type HotSpotReport struct {
    TotalHotSpots    int
    TotalAllocations int64
    TotalBytes       int64
    HotSpots         []HotSpotData
}

func (hsr HotSpotReport) String() string {
    result := fmt.Sprintf(`Allocation Hot Spot Report:
  Total Hot Spots: %d
  Total Allocations: %d
  Total Bytes: %s`,
        hsr.TotalHotSpots,
        hsr.TotalAllocations,
        formatBytes(uint64(hsr.TotalBytes)))
    
    if len(hsr.HotSpots) > 0 {
        result += "\n\nTop Allocators:"
        for i, hs := range hsr.HotSpots {
            if i >= 10 { // Show top 10
                break
            }
            
            percentage := float64(hs.TotalBytes) / float64(hsr.TotalBytes) * 100
            result += fmt.Sprintf("\n  %d. %s", i+1, hs.FunctionName)
            result += fmt.Sprintf("\n     Allocations: %d", hs.Allocations)
            result += fmt.Sprintf("\n     Total Bytes: %s (%.1f%%)", formatBytes(uint64(hs.TotalBytes)), percentage)
            result += fmt.Sprintf("\n     Avg Size: %.1f bytes", hs.AvgSize)
            result += fmt.Sprintf("\n     Duration: %v", hs.LastSeen.Sub(hs.FirstSeen))
        }
    }
    
    return result
}

func demonstrateHotSpotAnalysis() {
    fmt.Println("\n=== ALLOCATION HOT SPOT ANALYSIS ===")
    
    analyzer := NewAllocationHotSpotAnalyzer()
    analyzer.Enable()
    defer analyzer.Disable()
    
    // Simulate allocation-heavy operations
    allocateSlices()
    createStructs()
    stringOperations()
    
    // Capture and analyze
    if err := analyzer.CaptureProfile(); err != nil {
        fmt.Printf("Error capturing profile: %v\n", err)
        return
    }
    
    report := analyzer.GetReport()
    fmt.Printf("\n%s\n", report)
}

func allocateSlices() {
    for i := 0; i < 100; i++ {
        data := make([]int, 1000)
        _ = data
    }
}

func createStructs() {
    type TestStruct struct {
        Data [100]byte
    }
    
    for i := 0; i < 50; i++ {
        s := &TestStruct{}
        _ = s
    }
}

func stringOperations() {
    var builder strings.Builder
    for i := 0; i < 200; i++ {
        builder.WriteString(fmt.Sprintf("string %d ", i))
    }
    _ = builder.String()
}
```

### Allocation Pattern Analysis

```go
package main

import (
    "fmt"
    "sort"
    "time"
)

// AllocationPatternAnalyzer identifies common allocation patterns
type AllocationPatternAnalyzer struct {
    patterns map[PatternType]*PatternStats
}

type PatternType int

const (
    SmallFrequent PatternType = iota
    LargeInfrequent
    BurstPattern
    GrowingPattern
    ConstantPattern
)

func (pt PatternType) String() string {
    switch pt {
    case SmallFrequent:
        return "Small Frequent"
    case LargeInfrequent:
        return "Large Infrequent"
    case BurstPattern:
        return "Burst Pattern"
    case GrowingPattern:
        return "Growing Pattern"
    case ConstantPattern:
        return "Constant Pattern"
    default:
        return "Unknown"
    }
}

type PatternStats struct {
    Type        PatternType
    Count       int64
    TotalBytes  int64
    AvgSize     float64
    MinSize     uint64
    MaxSize     uint64
    Frequency   float64 // Allocations per second
    Examples    []AllocationExample
}

type AllocationExample struct {
    Timestamp time.Time
    Size      uint64
    Location  string
}

func NewAllocationPatternAnalyzer() *AllocationPatternAnalyzer {
    return &AllocationPatternAnalyzer{
        patterns: make(map[PatternType]*PatternStats),
    }
}

func (apa *AllocationPatternAnalyzer) AnalyzeAllocationSample(sample AllocationSample, previous *AllocationSample) {
    if previous == nil {
        return
    }
    
    duration := sample.Timestamp.Sub(previous.Timestamp)
    if duration <= 0 {
        return
    }
    
    allocDiff := sample.TotalAlloc - previous.TotalAlloc
    objectDiff := sample.Mallocs - previous.Mallocs
    
    if objectDiff == 0 {
        return
    }
    
    avgSize := float64(allocDiff) / float64(objectDiff)
    frequency := float64(objectDiff) / duration.Seconds()
    
    // Classify allocation pattern
    pattern := apa.classifyPattern(avgSize, frequency, allocDiff, objectDiff)
    
    // Update pattern statistics
    apa.updatePatternStats(pattern, allocDiff, objectDiff, avgSize, frequency, sample.Timestamp)
}

func (apa *AllocationPatternAnalyzer) classifyPattern(avgSize, frequency float64, totalBytes, objects uint64) PatternType {
    // Classification logic based on size and frequency
    
    if avgSize < 1024 && frequency > 100 { // Small objects, high frequency
        return SmallFrequent
    } else if avgSize > 1024*1024 && frequency < 10 { // Large objects, low frequency
        return LargeInfrequent
    } else if frequency > 500 { // Very high frequency regardless of size
        return BurstPattern
    } else if avgSize > 10*1024 && avgSize < 100*1024 { // Medium size, steady rate
        return GrowingPattern
    } else {
        return ConstantPattern
    }
}

func (apa *AllocationPatternAnalyzer) updatePatternStats(patternType PatternType, totalBytes, objects uint64, avgSize, frequency float64, timestamp time.Time) {
    stats, exists := apa.patterns[patternType]
    if !exists {
        stats = &PatternStats{
            Type:    patternType,
            MinSize: ^uint64(0), // Max uint64
            MaxSize: 0,
        }
        apa.patterns[patternType] = stats
    }
    
    stats.Count += int64(objects)
    stats.TotalBytes += int64(totalBytes)
    stats.AvgSize = float64(stats.TotalBytes) / float64(stats.Count)
    stats.Frequency = (stats.Frequency + frequency) / 2 // Running average
    
    objectSize := uint64(avgSize)
    if objectSize < stats.MinSize {
        stats.MinSize = objectSize
    }
    if objectSize > stats.MaxSize {
        stats.MaxSize = objectSize
    }
    
    // Add example (keep only recent ones)
    example := AllocationExample{
        Timestamp: timestamp,
        Size:      objectSize,
        Location:  "unknown", // Would be filled from stack trace
    }
    
    stats.Examples = append(stats.Examples, example)
    if len(stats.Examples) > 10 {
        stats.Examples = stats.Examples[1:] // Keep only recent examples
    }
}

func (apa *AllocationPatternAnalyzer) GetPatternReport() PatternReport {
    var patterns []PatternStats
    var totalAllocations int64
    var totalBytes int64
    
    for _, stats := range apa.patterns {
        patterns = append(patterns, *stats)
        totalAllocations += stats.Count
        totalBytes += stats.TotalBytes
    }
    
    // Sort by total bytes
    sort.Slice(patterns, func(i, j int) bool {
        return patterns[i].TotalBytes > patterns[j].TotalBytes
    })
    
    return PatternReport{
        TotalPatterns:    len(patterns),
        TotalAllocations: totalAllocations,
        TotalBytes:       totalBytes,
        Patterns:         patterns,
    }
}

type PatternReport struct {
    TotalPatterns    int
    TotalAllocations int64
    TotalBytes       int64
    Patterns         []PatternStats
}

func (pr PatternReport) String() string {
    result := fmt.Sprintf(`Allocation Pattern Report:
  Total Patterns: %d
  Total Allocations: %d
  Total Bytes: %s`,
        pr.TotalPatterns,
        pr.TotalAllocations,
        formatBytes(uint64(pr.TotalBytes)))
    
    if len(pr.Patterns) > 0 {
        result += "\n\nPattern Analysis:"
        for i, pattern := range pr.Patterns {
            percentage := float64(pattern.TotalBytes) / float64(pr.TotalBytes) * 100
            result += fmt.Sprintf("\n  %d. %s", i+1, pattern.Type)
            result += fmt.Sprintf("\n     Allocations: %d", pattern.Count)
            result += fmt.Sprintf("\n     Total Bytes: %s (%.1f%%)", formatBytes(uint64(pattern.TotalBytes)), percentage)
            result += fmt.Sprintf("\n     Average Size: %.1f bytes", pattern.AvgSize)
            result += fmt.Sprintf("\n     Size Range: %s - %s", formatBytes(pattern.MinSize), formatBytes(pattern.MaxSize))
            result += fmt.Sprintf("\n     Frequency: %.1f allocs/sec", pattern.Frequency)
            
            if len(pattern.Examples) > 0 {
                result += fmt.Sprintf("\n     Recent Examples: %d", len(pattern.Examples))
            }
        }
    }
    
    return result
}

func demonstratePatternAnalysis() {
    fmt.Println("\n=== ALLOCATION PATTERN ANALYSIS ===")
    
    tracker := NewAllocationTracker(50 * time.Millisecond)
    analyzer := NewAllocationPatternAnalyzer()
    
    tracker.StartTracking()
    defer tracker.StopTracking()
    
    // Generate different allocation patterns
    
    // Small frequent pattern
    fmt.Println("Generating small frequent allocations...")
    for i := 0; i < 500; i++ {
        data := make([]byte, 64)
        _ = data
        if i%50 == 0 {
            time.Sleep(10 * time.Millisecond)
        }
    }
    
    time.Sleep(200 * time.Millisecond)
    
    // Large infrequent pattern
    fmt.Println("Generating large infrequent allocations...")
    for i := 0; i < 5; i++ {
        data := make([]byte, 2*1024*1024) // 2MB
        _ = data
        time.Sleep(100 * time.Millisecond)
    }
    
    // Burst pattern
    fmt.Println("Generating burst allocations...")
    for burst := 0; burst < 3; burst++ {
        for i := 0; i < 200; i++ {
            data := make([]int, 500)
            _ = data
        }
        time.Sleep(200 * time.Millisecond)
    }
    
    time.Sleep(500 * time.Millisecond)
    
    // Analyze patterns
    samples := tracker.GetSamples()
    var previous *AllocationSample
    
    for i := range samples {
        analyzer.AnalyzeAllocationSample(samples[i], previous)
        previous = &samples[i]
    }
    
    report := analyzer.GetPatternReport()
    fmt.Printf("\n%s\n", report)
}
```

## Optimization Strategies

### Allocation Reduction Techniques

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Object pool for reducing allocations
type ObjectPool struct {
    pool sync.Pool
    new  func() interface{}
}

func NewObjectPool(newFunc func() interface{}) *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{New: newFunc},
        new:  newFunc,
    }
}

func (op *ObjectPool) Get() interface{} {
    return op.pool.Get()
}

func (op *ObjectPool) Put(obj interface{}) {
    op.pool.Put(obj)
}

// Buffer pool example
type BufferPool struct {
    pool sync.Pool
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 1024) // 1KB initial capacity
            },
        },
    }
}

func (bp *BufferPool) Get() []byte {
    return bp.pool.Get().([]byte)
}

func (bp *BufferPool) Put(buf []byte) {
    // Reset length but keep capacity
    buf = buf[:0]
    bp.pool.Put(buf)
}

// String builder pool
type StringBuilderPool struct {
    pool sync.Pool
}

func NewStringBuilderPool() *StringBuilderPool {
    return &StringBuilderPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &strings.Builder{}
            },
        },
    }
}

func (sbp *StringBuilderPool) Get() *strings.Builder {
    return sbp.pool.Get().(*strings.Builder)
}

func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
    sb.Reset()
    sbp.pool.Put(sb)
}

func demonstrateAllocationOptimization() {
    fmt.Println("\n=== ALLOCATION OPTIMIZATION TECHNIQUES ===")
    
    // Test without pools (high allocation)
    tracker1 := NewAllocationTracker(100 * time.Millisecond)
    tracker1.StartTracking()
    
    fmt.Println("Testing without object pooling...")
    start := time.Now()
    for i := 0; i < 10000; i++ {
        // Allocate new buffer each time
        buffer := make([]byte, 1024)
        // Simulate work
        for j := 0; j < len(buffer); j++ {
            buffer[j] = byte(j % 256)
        }
        _ = buffer // Use buffer
    }
    timeWithoutPool := time.Since(start)
    tracker1.StopTracking()
    
    time.Sleep(200 * time.Millisecond)
    reportWithoutPool := tracker1.GetAllocationReport()
    
    // Test with pools (lower allocation)
    tracker2 := NewAllocationTracker(100 * time.Millisecond)
    tracker2.StartTracking()
    
    bufferPool := NewBufferPool()
    
    fmt.Println("Testing with object pooling...")
    start = time.Now()
    for i := 0; i < 10000; i++ {
        // Reuse buffer from pool
        buffer := bufferPool.Get()
        if cap(buffer) < 1024 {
            buffer = make([]byte, 1024)
        }
        buffer = buffer[:1024]
        
        // Simulate work
        for j := 0; j < len(buffer); j++ {
            buffer[j] = byte(j % 256)
        }
        
        bufferPool.Put(buffer)
    }
    timeWithPool := time.Since(start)
    tracker2.StopTracking()
    
    time.Sleep(200 * time.Millisecond)
    reportWithPool := tracker2.GetAllocationReport()
    
    // Compare results
    fmt.Printf("\nResults Comparison:\n")
    fmt.Printf("Without pooling:\n  Time: %v\n  %s\n", timeWithoutPool, reportWithoutPool)
    fmt.Printf("\nWith pooling:\n  Time: %v\n  %s\n", timeWithPool, reportWithPool)
    
    fmt.Printf("\nImprovement:\n")
    fmt.Printf("  Time: %.2fx faster\n", float64(timeWithoutPool)/float64(timeWithPool))
    if reportWithoutPool.TotalAllocated > 0 && reportWithPool.TotalAllocated > 0 {
        fmt.Printf("  Allocations: %.2fx reduction\n", 
            float64(reportWithoutPool.TotalAllocated)/float64(reportWithPool.TotalAllocated))
    }
}
```

## Best Practices

### 1. Enable Allocation Profiling

```go
func init() {
    // Enable allocation profiling
    runtime.MemProfileRate = 1 // Sample every allocation
}
```

### 2. Use pprof for Analysis

```bash
# Capture allocation profile
go tool pprof http://localhost:6060/debug/pprof/allocs

# Analyze in web interface
go tool pprof -http=:8080 allocs.prof

# Show top allocating functions
(pprof) top10

# Show allocation call graph
(pprof) tree
```

### 3. Optimize Based on Findings

```go
// Reduce allocations with:
// 1. Object pooling
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

// 2. Pre-allocated slices
func processItems(items []Item) {
    results := make([]Result, 0, len(items)) // Pre-allocate capacity
    for _, item := range items {
        results = append(results, process(item))
    }
}

// 3. String builder for concatenation
func buildString(parts []string) string {
    var builder strings.Builder
    builder.Grow(estimateSize(parts)) // Pre-allocate
    for _, part := range parts {
        builder.WriteString(part)
    }
    return builder.String()
}
```

## Next Steps

- Learn [Memory Leak Detection](leak-detection.md) techniques
- Study [Heap Profiling](heap-profiling.md) analysis
- Explore [Memory Optimization](../../optimization/memory/README.md) strategies
- Master [Object Pooling](../../optimization/memory/memory-pools.md) patterns

## Summary

Allocation profiling helps optimize Go applications by:

1. **Identifying hot spots** - Functions with high allocation rates
2. **Understanding patterns** - Different allocation behaviors
3. **Measuring impact** - Quantifying allocation overhead
4. **Guiding optimization** - Focusing efforts on biggest wins
5. **Validating improvements** - Measuring optimization effectiveness

Use allocation profiling to build memory-efficient applications with minimal allocation overhead.
