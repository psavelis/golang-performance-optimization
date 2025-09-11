# Block and Mutex Profiling

Block and mutex profiling is essential for identifying synchronization bottlenecks in Go applications. These profiling techniques help detect contention points where goroutines are blocked waiting for shared resources, which can severely impact application performance and scalability.

## Understanding Block and Mutex Profiling

### Block Profiling
Block profiling tracks time spent by goroutines blocked on synchronization primitives. It measures:
- Time waiting on channel operations
- Mutex lock acquisition delays
- Wait group synchronization
- Select statement blocking
- Context cancellation waits

### Mutex Profiling
Mutex profiling specifically tracks contention on mutex operations:
- Lock acquisition time
- Lock hold duration
- Contention frequency
- Blocking call stacks

## Enabling Block and Mutex Profiling

### Runtime Configuration

```go
package main

import (
    "runtime"
    "time"
)

func init() {
    // Enable block profiling with 1% sampling rate
    runtime.SetBlockProfileRate(1)
    
    // Enable mutex profiling with detailed fraction
    runtime.SetMutexProfileFraction(1)
}

// Production configuration with environment-based tuning
func configureProfilingForProduction() {
    // Adjust based on application load
    blockRate := getBlockProfileRate() // Typically 1-100
    mutexFraction := getMutexProfileFraction() // Typically 1-1000
    
    runtime.SetBlockProfileRate(blockRate)
    runtime.SetMutexProfileFraction(mutexFraction)
}

func getBlockProfileRate() int {
    // Higher rates for development, lower for production
    if isProduction() {
        return 10 // Sample 10% of blocking events
    }
    return 1 // Sample all blocking events in development
}

func getMutexProfileFraction() int {
    // Fraction of mutex contention events to report
    if isProduction() {
        return 100 // Report 1% of mutex events
    }
    return 1 // Report all mutex events in development
}

func isProduction() bool {
    return os.Getenv("ENV") == "production"
}
```

### HTTP Profiling Endpoints

```go
package main

import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
    "log"
)

func setupProfilingServer() {
    // Configure profiling rates
    runtime.SetBlockProfileRate(1)
    runtime.SetMutexProfileFraction(1)
    
    // Start profiling server
    go func() {
        log.Println("Profiling server starting on :6060")
        if err := http.ListenAndServe(":6060", nil); err != nil {
            log.Printf("Profiling server error: %v", err)
        }
    }()
    
    // Allow profiling server to start
    time.Sleep(100 * time.Millisecond)
}

// Custom profiling endpoints with enhanced metadata
func setupCustomProfilingEndpoints() {
    http.HandleFunc("/debug/pprof/block-detailed", blockProfileHandler)
    http.HandleFunc("/debug/pprof/mutex-detailed", mutexProfileHandler)
    http.HandleFunc("/debug/pprof/contention-summary", contentionSummaryHandler)
}

func blockProfileHandler(w http.ResponseWriter, r *http.Request) {
    profile := pprof.Lookup("block")
    if profile == nil {
        http.Error(w, "Block profile not available", http.StatusNotFound)
        return
    }
    
    // Add custom headers with profiling metadata
    w.Header().Set("X-Profile-Type", "block")
    w.Header().Set("X-Profile-Duration", "30s")
    w.Header().Set("X-Sample-Rate", fmt.Sprintf("%d", runtime.BlockProfileRate()))
    
    profile.WriteTo(w, 0)
}

func mutexProfileHandler(w http.ResponseWriter, r *http.Request) {
    profile := pprof.Lookup("mutex")
    if profile == nil {
        http.Error(w, "Mutex profile not available", http.StatusNotFound)
        return
    }
    
    w.Header().Set("X-Profile-Type", "mutex")
    w.Header().Set("X-Mutex-Fraction", fmt.Sprintf("%d", runtime.MutexProfileFraction()))
    
    profile.WriteTo(w, 0)
}
```

## Collecting Block and Mutex Profiles

### Command Line Collection

```bash
# Collect block profile for 30 seconds
go tool pprof http://localhost:6060/debug/pprof/block?seconds=30

# Collect mutex profile
go tool pprof http://localhost:6060/debug/pprof/mutex

# Collect both profiles with file output
go tool pprof -output=block.prof http://localhost:6060/debug/pprof/block?seconds=30
go tool pprof -output=mutex.prof http://localhost:6060/debug/pprof/mutex

# Generate comparative reports
go tool pprof -top -output=block_top.txt block.prof
go tool pprof -top -output=mutex_top.txt mutex.prof
```

### Programmatic Collection

```go
package profiling

import (
    "fmt"
    "os"
    "runtime/pprof"
    "time"
    "context"
)

type ContentionProfiler struct {
    blockProfile  *pprof.Profile
    mutexProfile  *pprof.Profile
    sampleDuration time.Duration
}

func NewContentionProfiler(duration time.Duration) *ContentionProfiler {
    return &ContentionProfiler{
        sampleDuration: duration,
    }
}

func (cp *ContentionProfiler) CollectProfiles(ctx context.Context) error {
    // Start collection
    startTime := time.Now()
    
    // Wait for sample duration or context cancellation
    select {
    case <-time.After(cp.sampleDuration):
        // Normal completion
    case <-ctx.Done():
        return ctx.Err()
    }
    
    // Collect block profile
    cp.blockProfile = pprof.Lookup("block")
    if cp.blockProfile == nil {
        return fmt.Errorf("block profile not available")
    }
    
    // Collect mutex profile
    cp.mutexProfile = pprof.Lookup("mutex")
    if cp.mutexProfile == nil {
        return fmt.Errorf("mutex profile not available")
    }
    
    duration := time.Since(startTime)
    fmt.Printf("Profiles collected in %v\n", duration)
    
    return nil
}

func (cp *ContentionProfiler) SaveProfiles(prefix string) error {
    timestamp := time.Now().Format("20060102-150405")
    
    // Save block profile
    blockFile, err := os.Create(fmt.Sprintf("%s-block-%s.prof", prefix, timestamp))
    if err != nil {
        return fmt.Errorf("creating block profile file: %w", err)
    }
    defer blockFile.Close()
    
    if err := cp.blockProfile.WriteTo(blockFile, 0); err != nil {
        return fmt.Errorf("writing block profile: %w", err)
    }
    
    // Save mutex profile
    mutexFile, err := os.Create(fmt.Sprintf("%s-mutex-%s.prof", prefix, timestamp))
    if err != nil {
        return fmt.Errorf("creating mutex profile file: %w", err)
    }
    defer mutexFile.Close()
    
    if err := cp.mutexProfile.WriteTo(mutexFile, 0); err != nil {
        return fmt.Errorf("writing mutex profile: %w", err)
    }
    
    fmt.Printf("Profiles saved: %s-{block,mutex}-%s.prof\n", prefix, timestamp)
    return nil
}
```

## Analyzing Block Profiles

### Common Blocking Patterns

```go
package analysis

import (
    "fmt"
    "sync"
    "time"
    "context"
)

// Example: Channel blocking analysis
func demonstrateChannelBlocking() {
    // Problematic: Unbuffered channel with slow consumer
    ch := make(chan int)
    
    // Fast producers
    for i := 0; i < 10; i++ {
        go func(id int) {
            for j := 0; j < 1000; j++ {
                ch <- id*1000 + j // This will block frequently
            }
        }(i)
    }
    
    // Slow consumer
    go func() {
        for val := range ch {
            time.Sleep(10 * time.Millisecond) // Simulate slow processing
            _ = val
        }
    }()
    
    time.Sleep(5 * time.Second)
}

// Optimized version with buffered channels
func optimizedChannelPattern() {
    // Solution: Buffered channel with appropriate size
    bufferSize := 1000
    ch := make(chan int, bufferSize)
    
    var wg sync.WaitGroup
    
    // Producers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                select {
                case ch <- id*1000 + j:
                    // Successfully sent
                case <-time.After(100 * time.Millisecond):
                    // Timeout handling
                    fmt.Printf("Producer %d: send timeout\n", id)
                    return
                }
            }
        }(i)
    }
    
    // Consumer with batch processing
    go func() {
        batch := make([]int, 0, 100)
        for val := range ch {
            batch = append(batch, val)
            if len(batch) >= 100 {
                processBatch(batch)
                batch = batch[:0] // Reset slice
            }
        }
        if len(batch) > 0 {
            processBatch(batch)
        }
    }()
    
    wg.Wait()
    close(ch)
}

func processBatch(batch []int) {
    // Batch processing is more efficient
    time.Sleep(time.Duration(len(batch)) * time.Microsecond)
}
```

### Block Profile Analysis Tools

```go
package analysis

import (
    "bufio"
    "fmt"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

type BlockEvent struct {
    Count     int64
    Duration  time.Duration
    Function  string
    Location  string
    StackId   string
}

type BlockAnalyzer struct {
    events    []BlockEvent
    threshold time.Duration
}

func NewBlockAnalyzer(threshold time.Duration) *BlockAnalyzer {
    return &BlockAnalyzer{
        threshold: threshold,
    }
}

func (ba *BlockAnalyzer) ParseProfile(profileData string) error {
    scanner := bufio.NewScanner(strings.NewReader(profileData))
    
    // Parse pprof text format
    var currentEvent BlockEvent
    inSample := false
    
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        if strings.HasPrefix(line, "--- ") {
            // New sample
            if inSample && currentEvent.Duration >= ba.threshold {
                ba.events = append(ba.events, currentEvent)
            }
            currentEvent = BlockEvent{}
            inSample = true
            
            // Parse sample header: "--- contention at 0x... ---"
            continue
        }
        
        if strings.Contains(line, "contentions") && strings.Contains(line, "delay") {
            // Parse: "123 contentions, 456789 delay"
            if err := ba.parseSampleLine(line, &currentEvent); err != nil {
                continue // Skip malformed lines
            }
        }
        
        if strings.Contains(line, "(") && strings.Contains(line, ")") {
            // Stack trace line
            currentEvent.Function = ba.extractFunction(line)
            currentEvent.Location = ba.extractLocation(line)
        }
    }
    
    // Add final event
    if inSample && currentEvent.Duration >= ba.threshold {
        ba.events = append(ba.events, currentEvent)
    }
    
    return scanner.Err()
}

func (ba *BlockAnalyzer) parseSampleLine(line string, event *BlockEvent) error {
    re := regexp.MustCompile(`(\d+)\s+contentions,\s+(\d+)\s+delay`)
    matches := re.FindStringSubmatch(line)
    
    if len(matches) != 3 {
        return fmt.Errorf("invalid sample line format")
    }
    
    count, err := strconv.ParseInt(matches[1], 10, 64)
    if err != nil {
        return err
    }
    
    delayNs, err := strconv.ParseInt(matches[2], 10, 64)
    if err != nil {
        return err
    }
    
    event.Count = count
    event.Duration = time.Duration(delayNs) * time.Nanosecond
    
    return nil
}

func (ba *BlockAnalyzer) extractFunction(line string) string {
    // Extract function name from stack trace line
    parts := strings.Fields(line)
    if len(parts) > 0 {
        return parts[0]
    }
    return "unknown"
}

func (ba *BlockAnalyzer) extractLocation(line string) string {
    // Extract file:line from stack trace
    if idx := strings.LastIndex(line, "("); idx != -1 {
        if end := strings.LastIndex(line, ")"); end > idx {
            return line[idx+1 : end]
        }
    }
    return "unknown"
}

func (ba *BlockAnalyzer) GenerateReport() string {
    // Sort by total duration (count * duration)
    sort.Slice(ba.events, func(i, j int) bool {
        totalI := time.Duration(ba.events[i].Count) * ba.events[i].Duration
        totalJ := time.Duration(ba.events[j].Count) * ba.events[j].Duration
        return totalI > totalJ
    })
    
    var report strings.Builder
    report.WriteString("Block Profile Analysis Report\n")
    report.WriteString("=============================\n\n")
    
    totalEvents := len(ba.events)
    if totalEvents == 0 {
        report.WriteString("No blocking events detected above threshold.\n")
        return report.String()
    }
    
    var totalDuration time.Duration
    var totalCount int64
    
    for _, event := range ba.events {
        totalDuration += time.Duration(event.Count) * event.Duration
        totalCount += event.Count
    }
    
    report.WriteString(fmt.Sprintf("Total Events: %d\n", totalEvents))
    report.WriteString(fmt.Sprintf("Total Contentions: %d\n", totalCount))
    report.WriteString(fmt.Sprintf("Total Duration: %v\n", totalDuration))
    report.WriteString(fmt.Sprintf("Average per Event: %v\n", totalDuration/time.Duration(totalEvents)))
    report.WriteString("\nTop Blocking Events:\n")
    report.WriteString("-------------------\n")
    
    for i, event := range ba.events {
        if i >= 10 { // Show top 10
            break
        }
        
        totalEventDuration := time.Duration(event.Count) * event.Duration
        percentage := float64(totalEventDuration) / float64(totalDuration) * 100
        
        report.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, event.Function))
        report.WriteString(fmt.Sprintf("   Location: %s\n", event.Location))
        report.WriteString(fmt.Sprintf("   Count: %d contentions\n", event.Count))
        report.WriteString(fmt.Sprintf("   Duration: %v per contention\n", event.Duration))
        report.WriteString(fmt.Sprintf("   Total: %v (%.2f%%)\n", totalEventDuration, percentage))
    }
    
    return report.String()
}
```

## Analyzing Mutex Profiles

### Mutex Contention Detection

```go
package mutex

import (
    "fmt"
    "sync"
    "time"
    "runtime"
    "context"
)

// Problematic mutex pattern
type ContentionDemo struct {
    mu    sync.Mutex
    data  map[string]int
    stats map[string]int64
}

func NewContentionDemo() *ContentionDemo {
    return &ContentionDemo{
        data:  make(map[string]int),
        stats: make(map[string]int64),
    }
}

// This will show high contention in mutex profile
func (cd *ContentionDemo) HighContentionPattern() {
    var wg sync.WaitGroup
    
    // Many goroutines competing for same mutex
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < 1000; j++ {
                key := fmt.Sprintf("key-%d", j%10) // Limited key space
                
                cd.mu.Lock()
                cd.data[key]++
                cd.stats["updates"]++
                
                // Simulate work while holding lock (BAD PRACTICE)
                time.Sleep(time.Microsecond * 10)
                
                cd.mu.Unlock()
            }
        }(i)
    }
    
    wg.Wait()
}

// Optimized pattern with reduced contention
type OptimizedContentionDemo struct {
    shards    []*shard
    numShards int
}

type shard struct {
    mu   sync.RWMutex
    data map[string]int
}

func NewOptimizedContentionDemo(numShards int) *OptimizedContentionDemo {
    shards := make([]*shard, numShards)
    for i := range shards {
        shards[i] = &shard{
            data: make(map[string]int),
        }
    }
    
    return &OptimizedContentionDemo{
        shards:    shards,
        numShards: numShards,
    }
}

func (ocd *OptimizedContentionDemo) getShard(key string) *shard {
    hash := fnv32(key)
    return ocd.shards[hash%uint32(ocd.numShards)]
}

func fnv32(key string) uint32 {
    hash := uint32(2166136261)
    for _, b := range []byte(key) {
        hash ^= uint32(b)
        hash *= 16777619
    }
    return hash
}

func (ocd *OptimizedContentionDemo) LowContentionPattern() {
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < 1000; j++ {
                key := fmt.Sprintf("key-%d", j%100) // Larger key space
                shard := ocd.getShard(key)
                
                shard.mu.Lock()
                shard.data[key]++
                shard.mu.Unlock()
                
                // Work done outside critical section
                time.Sleep(time.Microsecond * 10)
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Mutex Profile Monitoring

```go
package monitoring

import (
    "context"
    "fmt"
    "runtime/pprof"
    "time"
    "log"
    "sync/atomic"
)

type MutexMonitor struct {
    alertThreshold  time.Duration
    checkInterval   time.Duration
    lastProfile     *pprof.Profile
    contentionCount int64
    alertCallback   func(string)
}

func NewMutexMonitor(threshold time.Duration, interval time.Duration) *MutexMonitor {
    return &MutexMonitor{
        alertThreshold: threshold,
        checkInterval:  interval,
        alertCallback:  defaultAlertCallback,
    }
}

func defaultAlertCallback(message string) {
    log.Printf("MUTEX CONTENTION ALERT: %s", message)
}

func (mm *MutexMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(mm.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := mm.checkContention(); err != nil {
                log.Printf("Mutex monitoring error: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (mm *MutexMonitor) checkContention() error {
    profile := pprof.Lookup("mutex")
    if profile == nil {
        return fmt.Errorf("mutex profile not available")
    }
    
    // Simple contention detection based on profile sample count
    sampleCount := profile.Count()
    
    if mm.lastProfile != nil {
        delta := sampleCount - mm.lastProfile.Count()
        if delta > 0 {
            atomic.AddInt64(&mm.contentionCount, int64(delta))
            
            // Check if contention rate exceeds threshold
            rate := time.Duration(delta) * mm.checkInterval
            if rate > mm.alertThreshold {
                mm.alertCallback(fmt.Sprintf(
                    "High mutex contention detected: %d events in %v (rate: %v)",
                    delta, mm.checkInterval, rate))
            }
        }
    }
    
    mm.lastProfile = profile
    return nil
}

func (mm *MutexMonitor) GetContentionCount() int64 {
    return atomic.LoadInt64(&mm.contentionCount)
}

func (mm *MutexMonitor) SetAlertCallback(callback func(string)) {
    mm.alertCallback = callback
}
```

## Advanced Analysis Techniques

### Comparative Analysis

```go
package analysis

import (
    "fmt"
    "time"
)

type ProfileComparison struct {
    Baseline *ContentionProfile
    Current  *ContentionProfile
}

type ContentionProfile struct {
    Timestamp      time.Time
    BlockEvents    []BlockEvent
    MutexEvents    []MutexEvent
    TotalDuration  time.Duration
    TotalCount     int64
}

type MutexEvent struct {
    Count    int64
    Duration time.Duration
    Function string
    Location string
}

func CompareProfiles(baseline, current *ContentionProfile) *ProfileComparison {
    return &ProfileComparison{
        Baseline: baseline,
        Current:  current,
    }
}

func (pc *ProfileComparison) GenerateReport() string {
    var report strings.Builder
    
    report.WriteString("Contention Profile Comparison\n")
    report.WriteString("============================\n\n")
    
    // Overall metrics comparison
    durationChange := pc.Current.TotalDuration - pc.Baseline.TotalDuration
    countChange := pc.Current.TotalCount - pc.Baseline.TotalCount
    
    durationPercent := float64(durationChange) / float64(pc.Baseline.TotalDuration) * 100
    countPercent := float64(countChange) / float64(pc.Baseline.TotalCount) * 100
    
    report.WriteString(fmt.Sprintf("Duration Change: %v (%.2f%%)\n", durationChange, durationPercent))
    report.WriteString(fmt.Sprintf("Count Change: %d (%.2f%%)\n", countChange, countPercent))
    
    // Performance verdict
    if durationChange < 0 {
        report.WriteString("✅ IMPROVEMENT: Contention duration decreased\n")
    } else if durationChange > time.Millisecond*100 {
        report.WriteString("❌ REGRESSION: Significant contention increase\n")
    } else {
        report.WriteString("➡️  STABLE: Minor changes in contention\n")
    }
    
    report.WriteString("\nDetailed Analysis:\n")
    report.WriteString("-----------------\n")
    
    // Function-level comparison
    pc.compareFunctions(&report)
    
    return report.String()
}

func (pc *ProfileComparison) compareFunctions(report *strings.Builder) {
    // Create maps for easier comparison
    baselineFuncs := make(map[string]time.Duration)
    currentFuncs := make(map[string]time.Duration)
    
    for _, event := range pc.Baseline.BlockEvents {
        baselineFuncs[event.Function] += time.Duration(event.Count) * event.Duration
    }
    
    for _, event := range pc.Current.BlockEvents {
        currentFuncs[event.Function] += time.Duration(event.Count) * event.Duration
    }
    
    // Find regressions and improvements
    var regressions []string
    var improvements []string
    
    for fn, currentDur := range currentFuncs {
        baselineDur, exists := baselineFuncs[fn]
        if !exists {
            regressions = append(regressions, fmt.Sprintf("  NEW: %s - %v", fn, currentDur))
            continue
        }
        
        change := currentDur - baselineDur
        if change > time.Millisecond*10 {
            percent := float64(change) / float64(baselineDur) * 100
            regressions = append(regressions, fmt.Sprintf("  %s: +%v (%.1f%%)", fn, change, percent))
        } else if change < -time.Millisecond*10 {
            percent := float64(-change) / float64(baselineDur) * 100
            improvements = append(improvements, fmt.Sprintf("  %s: %v (%.1f%%)", fn, change, percent))
        }
    }
    
    if len(regressions) > 0 {
        report.WriteString("\nRegressions:\n")
        for _, reg := range regressions {
            report.WriteString(reg + "\n")
        }
    }
    
    if len(improvements) > 0 {
        report.WriteString("\nImprovements:\n")
        for _, imp := range improvements {
            report.WriteString(imp + "\n")
        }
    }
}
```

## Production Deployment Strategies

### Automated Profiling Pipeline

```go
package production

import (
    "context"
    "fmt"
    "time"
    "path/filepath"
    "os"
)

type ProductionProfiler struct {
    config        *ProfilerConfig
    scheduler     *ProfileScheduler
    alertManager  *AlertManager
    storage       *ProfileStorage
}

type ProfilerConfig struct {
    BlockProfileRate      int           `json:"block_profile_rate"`
    MutexProfileFraction  int           `json:"mutex_profile_fraction"`
    ProfileInterval       time.Duration `json:"profile_interval"`
    RetentionPeriod       time.Duration `json:"retention_period"`
    AlertThresholds       AlertThresholds `json:"alert_thresholds"`
}

type AlertThresholds struct {
    MaxBlockDuration    time.Duration `json:"max_block_duration"`
    MaxMutexContentions int64         `json:"max_mutex_contentions"`
    ComparisonThreshold float64       `json:"comparison_threshold"`
}

func NewProductionProfiler(config *ProfilerConfig) *ProductionProfiler {
    return &ProductionProfiler{
        config:       config,
        scheduler:    NewProfileScheduler(config.ProfileInterval),
        alertManager: NewAlertManager(config.AlertThresholds),
        storage:      NewProfileStorage(config.RetentionPeriod),
    }
}

func (pp *ProductionProfiler) Start(ctx context.Context) error {
    // Configure runtime profiling
    runtime.SetBlockProfileRate(pp.config.BlockProfileRate)
    runtime.SetMutexProfileFraction(pp.config.MutexProfileFraction)
    
    // Start components
    go pp.scheduler.Start(ctx, pp.collectProfile)
    go pp.alertManager.Start(ctx)
    go pp.storage.Start(ctx)
    
    fmt.Printf("Production profiler started with config: %+v\n", pp.config)
    return nil
}

func (pp *ProductionProfiler) collectProfile(ctx context.Context) error {
    timestamp := time.Now()
    
    // Collect profiles
    profiler := NewContentionProfiler(30 * time.Second)
    if err := profiler.CollectProfiles(ctx); err != nil {
        return fmt.Errorf("collecting profiles: %w", err)
    }
    
    // Save to storage
    profilePath := filepath.Join(
        pp.storage.GetStoragePath(),
        fmt.Sprintf("contention-%s", timestamp.Format("20060102-150405")))
    
    if err := profiler.SaveProfiles(profilePath); err != nil {
        return fmt.Errorf("saving profiles: %w", err)
    }
    
    // Analyze and alert
    analyzer := NewBlockAnalyzer(pp.config.AlertThresholds.MaxBlockDuration)
    
    // Trigger analysis in background
    go func() {
        if err := pp.analyzeAndAlert(profilePath, analyzer); err != nil {
            fmt.Printf("Analysis error: %v\n", err)
        }
    }()
    
    return nil
}

func (pp *ProductionProfiler) analyzeAndAlert(profilePath string, analyzer *BlockAnalyzer) error {
    // Load and analyze profile
    profileData, err := os.ReadFile(profilePath + "-block-*.prof")
    if err != nil {
        return fmt.Errorf("reading profile: %w", err)
    }
    
    if err := analyzer.ParseProfile(string(profileData)); err != nil {
        return fmt.Errorf("parsing profile: %w", err)
    }
    
    report := analyzer.GenerateReport()
    
    // Check thresholds and send alerts
    return pp.alertManager.ProcessReport(report, profilePath)
}

type ProfileScheduler struct {
    interval time.Duration
}

func NewProfileScheduler(interval time.Duration) *ProfileScheduler {
    return &ProfileScheduler{interval: interval}
}

func (ps *ProfileScheduler) Start(ctx context.Context, collectFunc func(context.Context) error) {
    ticker := time.NewTicker(ps.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := collectFunc(ctx); err != nil {
                fmt.Printf("Profile collection error: %v\n", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

type AlertManager struct {
    thresholds AlertThresholds
    lastAlert  time.Time
    cooldown   time.Duration
}

func NewAlertManager(thresholds AlertThresholds) *AlertManager {
    return &AlertManager{
        thresholds: thresholds,
        cooldown:   5 * time.Minute, // Prevent alert spam
    }
}

func (am *AlertManager) Start(ctx context.Context) {
    // Alert manager background processing
    // Implementation depends on alerting system (Slack, PagerDuty, etc.)
}

func (am *AlertManager) ProcessReport(report, profilePath string) error {
    if time.Since(am.lastAlert) < am.cooldown {
        return nil // Cooldown period
    }
    
    // Check if report indicates critical contention
    if am.shouldAlert(report) {
        alert := fmt.Sprintf(
            "Critical contention detected at %s\nProfile: %s\nReport:\n%s",
            time.Now().Format(time.RFC3339),
            profilePath,
            report)
        
        if err := am.sendAlert(alert); err != nil {
            return fmt.Errorf("sending alert: %w", err)
        }
        
        am.lastAlert = time.Now()
    }
    
    return nil
}

func (am *AlertManager) shouldAlert(report string) bool {
    // Simple heuristic - check for keywords indicating high contention
    criticalIndicators := []string{
        "High mutex contention",
        "Critical blocking",
        "Deadlock detected",
    }
    
    for _, indicator := range criticalIndicators {
        if strings.Contains(report, indicator) {
            return true
        }
    }
    
    return false
}

func (am *AlertManager) sendAlert(message string) error {
    // Implementation depends on alerting system
    fmt.Printf("🚨 CONTENTION ALERT: %s\n", message)
    return nil
}

type ProfileStorage struct {
    retention time.Duration
    basePath  string
}

func NewProfileStorage(retention time.Duration) *ProfileStorage {
    return &ProfileStorage{
        retention: retention,
        basePath:  "/var/lib/profiling/contention",
    }
}

func (ps *ProfileStorage) Start(ctx context.Context) {
    // Cleanup old profiles periodically
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ps.cleanup()
        case <-ctx.Done():
            return
        }
    }
}

func (ps *ProfileStorage) GetStoragePath() string {
    return ps.basePath
}

func (ps *ProfileStorage) cleanup() {
    // Remove profiles older than retention period
    cutoff := time.Now().Add(-ps.retention)
    
    err := filepath.Walk(ps.basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Continue walking
        }
        
        if info.ModTime().Before(cutoff) && strings.HasSuffix(path, ".prof") {
            if err := os.Remove(path); err != nil {
                fmt.Printf("Failed to remove old profile %s: %v\n", path, err)
            } else {
                fmt.Printf("Cleaned up old profile: %s\n", path)
            }
        }
        
        return nil
    })
    
    if err != nil {
        fmt.Printf("Profile cleanup error: %v\n", err)
    }
}
```

Block and mutex profiling are essential tools for identifying and resolving synchronization bottlenecks in Go applications. By systematically collecting, analyzing, and monitoring contention patterns, you can significantly improve application performance and scalability.

## Key Takeaways

1. **Enable profiling early** in development to catch contention issues
2. **Use appropriate sampling rates** for production environments
3. **Analyze patterns** to identify hotspots and optimization opportunities
4. **Implement monitoring** to detect regressions in production
5. **Compare profiles** over time to validate optimization efforts
6. **Automate collection** and alerting for production systems

The combination of block and mutex profiling provides comprehensive visibility into synchronization performance, enabling data-driven optimization decisions.
