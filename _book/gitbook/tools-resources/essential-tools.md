# Essential Tools and Frameworks

This comprehensive guide covers the essential tools and frameworks for Go performance engineering, from basic profiling to advanced distributed system analysis.

## Core Go Performance Tools

### pprof - The Foundation Tool

The most essential tool for Go performance analysis:

```go
// Comprehensive pprof integration example
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    _ "net/http/pprof" // Import for side effects
    "os"
    "runtime"
    "runtime/pprof"
    "runtime/trace"
    "time"
)

// Advanced pprof wrapper for production use
type ProfilerManager struct {
    enabled     bool
    profiles    map[string]*ProfileConfig
    server      *http.Server
    outputDir   string
    maxProfiles int
}

type ProfileConfig struct {
    Name        string        `yaml:"name"`
    Type        string        `yaml:"type"` // "cpu", "heap", "goroutine", "allocs", "block", "mutex"
    Duration    time.Duration `yaml:"duration"`
    Rate        int           `yaml:"rate"`
    Enabled     bool          `yaml:"enabled"`
    AutoCapture bool          `yaml:"auto_capture"`
    Triggers    []string      `yaml:"triggers"`
}

func NewProfilerManager(config *ProfilerConfig) *ProfilerManager {
    pm := &ProfilerManager{
        enabled:     config.Enabled,
        profiles:    make(map[string]*ProfileConfig),
        outputDir:   config.OutputDir,
        maxProfiles: config.MaxProfiles,
    }
    
    // Configure default profiles
    pm.profiles["cpu"] = &ProfileConfig{
        Name:        "cpu",
        Type:        "cpu",
        Duration:    30 * time.Second,
        Enabled:     true,
        AutoCapture: false,
        Triggers:    []string{"high_cpu", "manual"},
    }
    
    pm.profiles["heap"] = &ProfileConfig{
        Name:        "heap",
        Type:        "heap", 
        Enabled:     true,
        AutoCapture: true,
        Triggers:    []string{"high_memory", "gc_pressure", "manual"},
    }
    
    pm.profiles["goroutine"] = &ProfileConfig{
        Name:        "goroutine",
        Type:        "goroutine",
        Enabled:     true,
        AutoCapture: true,
        Triggers:    []string{"goroutine_leak", "manual"},
    }
    
    pm.profiles["allocs"] = &ProfileConfig{
        Name:     "allocs",
        Type:     "allocs",
        Duration: 60 * time.Second,
        Enabled:  true,
        Triggers: []string{"memory_churn", "manual"},
    }
    
    pm.profiles["block"] = &ProfileConfig{
        Name:     "block",
        Type:     "block",
        Duration: 30 * time.Second,
        Rate:     1, // Block profiling rate
        Enabled:  true,
        Triggers: []string{"high_blocking", "manual"},
    }
    
    pm.profiles["mutex"] = &ProfileConfig{
        Name:     "mutex",
        Type:     "mutex",
        Duration: 30 * time.Second,
        Rate:     1, // Mutex profiling rate
        Enabled:  true,
        Triggers: []string{"lock_contention", "manual"},
    }
    
    return pm
}

func (pm *ProfilerManager) Start() error {
    if !pm.enabled {
        return nil
    }
    
    // Enable profiling rates
    if config, exists := pm.profiles["block"]; exists && config.Enabled {
        runtime.SetBlockProfileRate(config.Rate)
    }
    
    if config, exists := pm.profiles["mutex"]; exists && config.Enabled {
        runtime.SetMutexProfileFraction(config.Rate)
    }
    
    // Start pprof HTTP server
    mux := http.NewServeMux()
    
    // Add custom profiling endpoints
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    // Custom profiling endpoints
    mux.HandleFunc("/debug/pprof/capture", pm.handleCaptureProfile)
    mux.HandleFunc("/debug/pprof/status", pm.handleProfileStatus)
    mux.HandleFunc("/debug/pprof/config", pm.handleProfileConfig)
    
    pm.server = &http.Server{
        Addr:    ":6060",
        Handler: mux,
    }
    
    go func() {
        if err := pm.server.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("pprof server error: %v", err)
        }
    }()
    
    // Start automatic profiling if configured
    pm.startAutomaticProfiling()
    
    log.Println("ProfilerManager started on :6060")
    return nil
}

func (pm *ProfilerManager) CaptureProfile(profileType string, duration time.Duration) (string, error) {
    config, exists := pm.profiles[profileType]
    if !exists || !config.Enabled {
        return "", fmt.Errorf("profile type %s not available", profileType)
    }
    
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("%s/%s-profile-%s.pprof", pm.outputDir, profileType, timestamp)
    
    file, err := os.Create(filename)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    switch profileType {
    case "cpu":
        if err := pprof.StartCPUProfile(file); err != nil {
            return "", err
        }
        time.Sleep(duration)
        pprof.StopCPUProfile()
        
    case "heap":
        runtime.GC() // Force GC for accurate heap profile
        if err := pprof.WriteHeapProfile(file); err != nil {
            return "", err
        }
        
    case "goroutine":
        profile := pprof.Lookup("goroutine")
        if err := profile.WriteTo(file, 0); err != nil {
            return "", err
        }
        
    case "allocs":
        profile := pprof.Lookup("allocs")
        if err := profile.WriteTo(file, 0); err != nil {
            return "", err
        }
        
    case "block":
        profile := pprof.Lookup("block")
        if err := profile.WriteTo(file, 0); err != nil {
            return "", err
        }
        
    case "mutex":
        profile := pprof.Lookup("mutex")
        if err := profile.WriteTo(file, 0); err != nil {
            return "", err
        }
        
    default:
        return "", fmt.Errorf("unknown profile type: %s", profileType)
    }
    
    log.Printf("Captured %s profile: %s", profileType, filename)
    return filename, nil
}

// Enhanced trace integration
func (pm *ProfilerManager) CaptureTrace(duration time.Duration) (string, error) {
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("%s/trace-%s.out", pm.outputDir, timestamp)
    
    file, err := os.Create(filename)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    if err := trace.Start(file); err != nil {
        return "", err
    }
    
    time.Sleep(duration)
    trace.Stop()
    
    log.Printf("Captured trace: %s", filename)
    return filename, nil
}

// Advanced profiling analysis
func (pm *ProfilerManager) AnalyzeProfile(filename string) (*ProfileAnalysis, error) {
    analysis := &ProfileAnalysis{
        Filename:  filename,
        Timestamp: time.Now(),
    }
    
    // Parse the profile file
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    profile, err := pprof.Parse(file)
    if err != nil {
        return nil, err
    }
    
    // Extract key metrics
    analysis.SampleCount = int64(len(profile.Sample))
    analysis.Duration = profile.DurationNanos
    analysis.PeriodType = profile.PeriodType.Type
    analysis.Period = profile.Period
    
    // Analyze top functions
    analysis.TopFunctions = pm.extractTopFunctions(profile, 10)
    
    // Analyze allocation patterns if heap profile
    if strings.Contains(filename, "heap") || strings.Contains(filename, "allocs") {
        analysis.AllocationPatterns = pm.analyzeAllocations(profile)
    }
    
    // Analyze goroutine patterns if goroutine profile
    if strings.Contains(filename, "goroutine") {
        analysis.GoroutinePatterns = pm.analyzeGoroutines(profile)
    }
    
    return analysis, nil
}

type ProfileAnalysis struct {
    Filename            string               `json:"filename"`
    Timestamp          time.Time            `json:"timestamp"`
    SampleCount        int64                `json:"sample_count"`
    Duration           int64                `json:"duration_nanos"`
    PeriodType         string               `json:"period_type"`
    Period             int64                `json:"period"`
    TopFunctions       []FunctionSample     `json:"top_functions"`
    AllocationPatterns *AllocationAnalysis  `json:"allocation_patterns,omitempty"`
    GoroutinePatterns  *GoroutineAnalysis   `json:"goroutine_patterns,omitempty"`
}

type FunctionSample struct {
    FunctionName string  `json:"function_name"`
    SelfValue    int64   `json:"self_value"`
    CumValue     int64   `json:"cum_value"`
    SelfPercent  float64 `json:"self_percent"`
    CumPercent   float64 `json:"cum_percent"`
}

type AllocationAnalysis struct {
    TotalAllocations int64               `json:"total_allocations"`
    TotalBytes       int64               `json:"total_bytes"`
    LargeAllocations []AllocationSample  `json:"large_allocations"`
    AllocationHotSpots []FunctionSample  `json:"allocation_hot_spots"`
}

type GoroutineAnalysis struct {
    TotalGoroutines    int                    `json:"total_goroutines"`
    BlockedGoroutines  int                    `json:"blocked_goroutines"`
    GoroutineStates    map[string]int         `json:"goroutine_states"`
    CommonStackTraces  []StackTraceFrequency  `json:"common_stack_traces"`
}
```

### Benchmarking Framework

```go
// Advanced benchmarking framework
package benchmarks

import (
    "context"
    "fmt"
    "math"
    "runtime"
    "sort"
    "testing"
    "time"
)

// Enhanced benchmark runner with statistical analysis
type BenchmarkSuite struct {
    name       string
    benchmarks []BenchmarkFunc
    config     *BenchmarkConfig
    results    []BenchmarkResult
    baseline   *BenchmarkBaseline
}

type BenchmarkConfig struct {
    Iterations      int           `yaml:"iterations"`
    WarmupRounds    int           `yaml:"warmup_rounds"`
    MinDuration     time.Duration `yaml:"min_duration"`
    MaxDuration     time.Duration `yaml:"max_duration"`
    CPUCores        []int         `yaml:"cpu_cores"`
    GCEnabled       bool          `yaml:"gc_enabled"`
    MemProfileRate  int           `yaml:"mem_profile_rate"`
    Statistical     bool          `yaml:"statistical_analysis"`
    OutputFormat    string        `yaml:"output_format"` // "json", "csv", "html"
}

type BenchmarkFunc struct {
    Name        string
    Function    func(*testing.B)
    Setup       func() error
    Teardown    func() error
    Parallel    bool
    Category    string
    Tags        []string
}

type BenchmarkResult struct {
    Name              string                 `json:"name"`
    Iterations        int                    `json:"iterations"`
    NsPerOp           int64                  `json:"ns_per_op"`
    AllocsPerOp       int64                  `json:"allocs_per_op"`
    BytesPerOp        int64                  `json:"bytes_per_op"`
    MemoryUsage       MemoryStats            `json:"memory_usage"`
    Statistics        *BenchmarkStatistics   `json:"statistics,omitempty"`
    Timestamp         time.Time              `json:"timestamp"`
    GoVersion         string                 `json:"go_version"`
    CPUCount          int                    `json:"cpu_count"`
    Environment       map[string]interface{} `json:"environment"`
}

type BenchmarkStatistics struct {
    Mean         float64 `json:"mean_ns"`
    Median       float64 `json:"median_ns"`
    StdDev       float64 `json:"std_dev_ns"`
    Min          float64 `json:"min_ns"`
    Max          float64 `json:"max_ns"`
    P95          float64 `json:"p95_ns"`
    P99          float64 `json:"p99_ns"`
    CoefficientOfVariation float64 `json:"cv"`
    Samples      []float64 `json:"samples"`
}

func NewBenchmarkSuite(name string, config *BenchmarkConfig) *BenchmarkSuite {
    return &BenchmarkSuite{
        name:    name,
        config:  config,
        results: make([]BenchmarkResult, 0),
    }
}

func (bs *BenchmarkSuite) AddBenchmark(name string, fn func(*testing.B), options ...BenchmarkOption) {
    benchmark := BenchmarkFunc{
        Name:     name,
        Function: fn,
        Category: "default",
        Tags:     []string{},
    }
    
    for _, option := range options {
        option(&benchmark)
    }
    
    bs.benchmarks = append(bs.benchmarks, benchmark)
}

// Benchmark options
type BenchmarkOption func(*BenchmarkFunc)

func WithSetup(setup func() error) BenchmarkOption {
    return func(bf *BenchmarkFunc) {
        bf.Setup = setup
    }
}

func WithTeardown(teardown func() error) BenchmarkOption {
    return func(bf *BenchmarkFunc) {
        bf.Teardown = teardown
    }
}

func WithCategory(category string) BenchmarkOption {
    return func(bf *BenchmarkFunc) {
        bf.Category = category
    }
}

func WithTags(tags ...string) BenchmarkOption {
    return func(bf *BenchmarkFunc) {
        bf.Tags = append(bf.Tags, tags...)
    }
}

func Parallel() BenchmarkOption {
    return func(bf *BenchmarkFunc) {
        bf.Parallel = true
    }
}

// Run comprehensive benchmarks
func (bs *BenchmarkSuite) Run() error {
    fmt.Printf("Running benchmark suite: %s\n", bs.name)
    fmt.Printf("Configuration: %+v\n", bs.config)
    
    // Collect environment information
    env := bs.collectEnvironmentInfo()
    
    for _, benchmark := range bs.benchmarks {
        fmt.Printf("Running benchmark: %s\n", benchmark.Name)
        
        // Setup
        if benchmark.Setup != nil {
            if err := benchmark.Setup(); err != nil {
                return fmt.Errorf("setup failed for %s: %v", benchmark.Name, err)
            }
        }
        
        // Run benchmark
        result, err := bs.runSingleBenchmark(benchmark, env)
        if err != nil {
            return fmt.Errorf("benchmark %s failed: %v", benchmark.Name, err)
        }
        
        bs.results = append(bs.results, *result)
        
        // Teardown
        if benchmark.Teardown != nil {
            if err := benchmark.Teardown(); err != nil {
                fmt.Printf("Warning: teardown failed for %s: %v\n", benchmark.Name, err)
            }
        }
        
        // Compare with baseline if available
        if bs.baseline != nil {
            comparison := bs.compareWithBaseline(*result)
            bs.printComparison(comparison)
        }
    }
    
    return nil
}

func (bs *BenchmarkSuite) runSingleBenchmark(benchmark BenchmarkFunc, env map[string]interface{}) (*BenchmarkResult, error) {
    // Create a testing.B-like structure for collection
    var samples []float64
    var totalIterations int
    var totalNs int64
    var totalAllocs int64
    var totalBytes int64
    
    memBefore := bs.getMemoryStats()
    
    // Run multiple rounds for statistical analysis
    rounds := 1
    if bs.config.Statistical {
        rounds = max(5, bs.config.Iterations/10) // At least 5 rounds
    }
    
    for round := 0; round < rounds; round++ {
        // Warmup
        if bs.config.WarmupRounds > 0 {
            for i := 0; i < bs.config.WarmupRounds; i++ {
                testing.Benchmark(benchmark.Function)
            }
        }
        
        // Actual benchmark
        result := testing.Benchmark(benchmark.Function)
        
        samples = append(samples, float64(result.NsPerOp()))
        totalIterations += result.N
        totalNs += int64(result.N) * result.NsPerOp()
        totalAllocs += int64(result.N) * result.AllocsPerOp()
        totalBytes += int64(result.N) * result.BytesPerOp()
    }
    
    memAfter := bs.getMemoryStats()
    
    // Calculate averages
    avgIterations := totalIterations / rounds
    avgNsPerOp := totalNs / int64(totalIterations)
    avgAllocsPerOp := totalAllocs / int64(totalIterations)
    avgBytesPerOp := totalBytes / int64(totalIterations)
    
    result := &BenchmarkResult{
        Name:        benchmark.Name,
        Iterations:  avgIterations,
        NsPerOp:     avgNsPerOp,
        AllocsPerOp: avgAllocsPerOp,
        BytesPerOp:  avgBytesPerOp,
        MemoryUsage: MemoryStats{
            Before: memBefore,
            After:  memAfter,
            Delta:  memAfter.HeapInuse - memBefore.HeapInuse,
        },
        Timestamp:   time.Now(),
        GoVersion:   runtime.Version(),
        CPUCount:    runtime.NumCPU(),
        Environment: env,
    }
    
    // Calculate statistics if enabled
    if bs.config.Statistical && len(samples) > 1 {
        result.Statistics = bs.calculateStatistics(samples)
    }
    
    return result, nil
}

func (bs *BenchmarkSuite) calculateStatistics(samples []float64) *BenchmarkStatistics {
    if len(samples) == 0 {
        return nil
    }
    
    // Sort samples for percentile calculations
    sorted := make([]float64, len(samples))
    copy(sorted, samples)
    sort.Float64s(sorted)
    
    // Calculate mean
    var sum float64
    for _, sample := range samples {
        sum += sample
    }
    mean := sum / float64(len(samples))
    
    // Calculate variance and standard deviation
    var variance float64
    for _, sample := range samples {
        diff := sample - mean
        variance += diff * diff
    }
    variance /= float64(len(samples))
    stdDev := math.Sqrt(variance)
    
    // Calculate percentiles
    p95Index := int(0.95 * float64(len(sorted)-1))
    p99Index := int(0.99 * float64(len(sorted)-1))
    medianIndex := len(sorted) / 2
    
    stats := &BenchmarkStatistics{
        Mean:    mean,
        Median:  sorted[medianIndex],
        StdDev:  stdDev,
        Min:     sorted[0],
        Max:     sorted[len(sorted)-1],
        P95:     sorted[p95Index],
        P99:     sorted[p99Index],
        CoefficientOfVariation: stdDev / mean,
        Samples: samples,
    }
    
    return stats
}

// Baseline comparison system
type BenchmarkBaseline struct {
    Version    string                       `json:"version"`
    Timestamp  time.Time                    `json:"timestamp"`
    Results    map[string]BenchmarkResult   `json:"results"`
    Metadata   map[string]interface{}       `json:"metadata"`
}

type BenchmarkComparison struct {
    BenchmarkName string                `json:"benchmark_name"`
    Current       BenchmarkResult       `json:"current"`
    Baseline      BenchmarkResult       `json:"baseline"`
    Improvements  ComparisonMetrics     `json:"improvements"`
    Regressions   ComparisonMetrics     `json:"regressions"`
    Status        string                `json:"status"` // "improved", "regressed", "unchanged"
    Significance  string                `json:"significance"` // "minor", "moderate", "major"
}

type ComparisonMetrics struct {
    PerformanceChange float64 `json:"performance_change_percent"`
    MemoryChange      float64 `json:"memory_change_percent"`
    AllocationChange  float64 `json:"allocation_change_percent"`
}

func (bs *BenchmarkSuite) compareWithBaseline(current BenchmarkResult) *BenchmarkComparison {
    baseline, exists := bs.baseline.Results[current.Name]
    if !exists {
        return nil
    }
    
    comparison := &BenchmarkComparison{
        BenchmarkName: current.Name,
        Current:       current,
        Baseline:      baseline,
    }
    
    // Calculate performance change (lower is better for ns/op)
    perfChange := float64(current.NsPerOp-baseline.NsPerOp) / float64(baseline.NsPerOp) * 100
    memChange := float64(current.BytesPerOp-baseline.BytesPerOp) / float64(baseline.BytesPerOp) * 100
    allocChange := float64(current.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp) * 100
    
    if perfChange < 0 { // Improvement
        comparison.Improvements.PerformanceChange = -perfChange
        comparison.Status = "improved"
    } else if perfChange > 0 { // Regression
        comparison.Regressions.PerformanceChange = perfChange
        comparison.Status = "regressed"
    } else {
        comparison.Status = "unchanged"
    }
    
    comparison.Improvements.MemoryChange = max(0, -memChange)
    comparison.Regressions.MemoryChange = max(0, memChange)
    comparison.Improvements.AllocationChange = max(0, -allocChange)
    comparison.Regressions.AllocationChange = max(0, allocChange)
    
    // Determine significance
    maxChange := math.Max(math.Abs(perfChange), math.Max(math.Abs(memChange), math.Abs(allocChange)))
    if maxChange < 5 {
        comparison.Significance = "minor"
    } else if maxChange < 25 {
        comparison.Significance = "moderate"
    } else {
        comparison.Significance = "major"
    }
    
    return comparison
}
```

## External Monitoring Tools Integration

### Prometheus Integration

```go
// Comprehensive Prometheus metrics integration
package monitoring

import (
    "context"
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/common/model"
)

// Custom metrics registry for Go applications
type GoMetricsRegistry struct {
    registry *prometheus.Registry
    
    // Runtime metrics
    goRoutines    prometheus.Gauge
    gcDuration    prometheus.Histogram
    heapSize      prometheus.Gauge
    heapInuse     prometheus.Gauge
    stackSize     prometheus.Gauge
    
    // Application metrics
    httpRequests     *prometheus.CounterVec
    httpDuration     *prometheus.HistogramVec
    dbConnections    prometheus.Gauge
    dbQueryDuration  *prometheus.HistogramVec
    cacheHitRate     prometheus.Gauge
    
    // Performance metrics
    cpuUsage         prometheus.Gauge
    memoryUsage      prometheus.Gauge
    allocRate        prometheus.Gauge
    gcPressure       prometheus.Gauge
    
    // Custom business metrics
    customMetrics    map[string]prometheus.Collector
}

func NewGoMetricsRegistry() *GoMetricsRegistry {
    registry := prometheus.NewRegistry()
    
    gmr := &GoMetricsRegistry{
        registry:      registry,
        customMetrics: make(map[string]prometheus.Collector),
    }
    
    gmr.initializeRuntimeMetrics()
    gmr.initializeApplicationMetrics()
    gmr.initializePerformanceMetrics()
    
    return gmr
}

func (gmr *GoMetricsRegistry) initializeRuntimeMetrics() {
    // Goroutine count
    gmr.goRoutines = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "runtime",
        Name:      "goroutines_total",
        Help:      "Number of goroutines currently running",
    })
    
    // GC duration
    gmr.gcDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Namespace: "go",
        Subsystem: "runtime",
        Name:      "gc_duration_seconds",
        Help:      "Time spent in garbage collection",
        Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
    })
    
    // Memory metrics
    gmr.heapSize = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memory",
        Name:      "heap_bytes",
        Help:      "Total heap size in bytes",
    })
    
    gmr.heapInuse = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memory",
        Name:      "heap_inuse_bytes",
        Help:      "Heap memory currently in use",
    })
    
    gmr.stackSize = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memory",
        Name:      "stack_bytes",
        Help:      "Stack memory in use",
    })
    
    // Register metrics
    gmr.registry.MustRegister(
        gmr.goRoutines,
        gmr.gcDuration,
        gmr.heapSize,
        gmr.heapInuse,
        gmr.stackSize,
    )
}

func (gmr *GoMetricsRegistry) initializeApplicationMetrics() {
    // HTTP metrics
    gmr.httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "app",
            Subsystem: "http",
            Name:      "requests_total",
            Help:      "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    gmr.httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "app",
            Subsystem: "http",
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration",
            Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )
    
    // Database metrics
    gmr.dbConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "app",
        Subsystem: "database",
        Name:      "connections_active",
        Help:      "Number of active database connections",
    })
    
    gmr.dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "app",
            Subsystem: "database",
            Name:      "query_duration_seconds",
            Help:      "Database query duration",
            Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
        },
        []string{"query_type", "table"},
    )
    
    // Cache metrics
    gmr.cacheHitRate = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "app",
        Subsystem: "cache",
        Name:      "hit_rate",
        Help:      "Cache hit rate percentage",
    })
    
    // Register metrics
    gmr.registry.MustRegister(
        gmr.httpRequests,
        gmr.httpDuration,
        gmr.dbConnections,
        gmr.dbQueryDuration,
        gmr.cacheHitRate,
    )
}

func (gmr *GoMetricsRegistry) StartMetricsCollection(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            gmr.collectRuntimeMetrics()
            gmr.collectPerformanceMetrics()
        }
    }()
}

func (gmr *GoMetricsRegistry) collectRuntimeMetrics() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    gmr.goRoutines.Set(float64(runtime.NumGoroutine()))
    gmr.heapSize.Set(float64(memStats.HeapSys))
    gmr.heapInuse.Set(float64(memStats.HeapInuse))
    gmr.stackSize.Set(float64(memStats.StackInuse))
    
    // GC metrics
    gmr.gcPressure.Set(float64(memStats.NumGC))
}

// Prometheus query client for analysis
type PrometheusAnalyzer struct {
    client v1.API
    config *AnalyzerConfig
}

type AnalyzerConfig struct {
    BaseURL        string        `yaml:"base_url"`
    Timeout        time.Duration `yaml:"timeout"`
    QueryInterval  time.Duration `yaml:"query_interval"`
    RetentionDays  int          `yaml:"retention_days"`
}

func NewPrometheusAnalyzer(config *AnalyzerConfig) (*PrometheusAnalyzer, error) {
    client, err := api.NewClient(api.Config{
        Address: config.BaseURL,
    })
    if err != nil {
        return nil, err
    }
    
    return &PrometheusAnalyzer{
        client: v1.NewAPI(client),
        config: config,
    }, nil
}

func (pa *PrometheusAnalyzer) AnalyzePerformanceTrends(ctx context.Context, duration time.Duration) (*PerformanceTrendAnalysis, error) {
    endTime := time.Now()
    startTime := endTime.Add(-duration)
    
    analysis := &PerformanceTrendAnalysis{
        StartTime: startTime,
        EndTime:   endTime,
        Metrics:   make(map[string]*MetricTrend),
    }
    
    // Define queries for key performance metrics
    queries := map[string]string{
        "response_time_p95": `histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))`,
        "response_time_p99": `histogram_quantile(0.99, rate(app_http_request_duration_seconds_bucket[5m]))`,
        "throughput":        `rate(app_http_requests_total[5m])`,
        "error_rate":        `rate(app_http_requests_total{status=~"5.."}[5m]) / rate(app_http_requests_total[5m])`,
        "memory_usage":      `go_memory_heap_inuse_bytes`,
        "goroutines":        `go_runtime_goroutines_total`,
        "gc_duration":       `rate(go_runtime_gc_duration_seconds_sum[5m])`,
    }
    
    for metricName, query := range queries {
        trend, err := pa.analyzeMetricTrend(ctx, query, startTime, endTime)
        if err != nil {
            continue // Log error but continue with other metrics
        }
        analysis.Metrics[metricName] = trend
    }
    
    // Calculate overall health score
    analysis.HealthScore = pa.calculateHealthScore(analysis.Metrics)
    
    return analysis, nil
}

func (pa *PrometheusAnalyzer) analyzeMetricTrend(ctx context.Context, query string, start, end time.Time) (*MetricTrend, error) {
    // Query range data
    r := v1.Range{
        Start: start,
        End:   end,
        Step:  pa.config.QueryInterval,
    }
    
    result, warnings, err := pa.client.QueryRange(ctx, query, r)
    if err != nil {
        return nil, err
    }
    
    if len(warnings) > 0 {
        fmt.Printf("Query warnings: %v\n", warnings)
    }
    
    trend := &MetricTrend{
        Query:      query,
        DataPoints: []DataPoint{},
    }
    
    // Process result
    if matrix, ok := result.(model.Matrix); ok {
        for _, sampleStream := range matrix {
            for _, pair := range sampleStream.Values {
                trend.DataPoints = append(trend.DataPoints, DataPoint{
                    Timestamp: time.Unix(int64(pair.Timestamp), 0),
                    Value:     float64(pair.Value),
                })
            }
        }
    }
    
    // Calculate trend statistics
    trend.Statistics = pa.calculateTrendStatistics(trend.DataPoints)
    
    return trend, nil
}

type MetricTrend struct {
    Query      string           `json:"query"`
    DataPoints []DataPoint      `json:"data_points"`
    Statistics *TrendStatistics `json:"statistics"`
}

type TrendStatistics struct {
    Mean       float64 `json:"mean"`
    Min        float64 `json:"min"`
    Max        float64 `json:"max"`
    StdDev     float64 `json:"std_dev"`
    Trend      string  `json:"trend"` // "increasing", "decreasing", "stable"
    Slope      float64 `json:"slope"`
    R2         float64 `json:"r_squared"`
}
```

### Grafana Dashboard Configuration

```go
// Grafana dashboard provisioning
package grafana

import (
    "encoding/json"
    "fmt"
    "time"
)

// Go Performance Dashboard definition
type GrafanaDashboard struct {
    ID            int                    `json:"id,omitempty"`
    UID           string                 `json:"uid,omitempty"`
    Title         string                 `json:"title"`
    Description   string                 `json:"description"`
    Tags          []string               `json:"tags"`
    Timezone      string                 `json:"timezone"`
    Panels        []Panel                `json:"panels"`
    Templating    Templating             `json:"templating"`
    Time          TimeRange              `json:"time"`
    Refresh       string                 `json:"refresh"`
    SchemaVersion int                    `json:"schemaVersion"`
}

type Panel struct {
    ID          int               `json:"id"`
    Title       string            `json:"title"`
    Type        string            `json:"type"`
    GridPos     GridPos           `json:"gridPos"`
    Targets     []Target          `json:"targets"`
    Options     map[string]interface{} `json:"options,omitempty"`
    FieldConfig FieldConfig       `json:"fieldConfig"`
    Alert       *AlertConfig      `json:"alert,omitempty"`
}

func CreateGoPerformanceDashboard() *GrafanaDashboard {
    dashboard := &GrafanaDashboard{
        UID:         "go-performance-monitoring",
        Title:       "Go Application Performance Monitoring",
        Description: "Comprehensive performance monitoring for Go applications",
        Tags:        []string{"go", "performance", "monitoring"},
        Timezone:    "browser",
        SchemaVersion: 27,
        Time: TimeRange{
            From: "now-1h",
            To:   "now",
        },
        Refresh: "5s",
    }
    
    // Add panels
    dashboard.Panels = []Panel{
        createOverviewPanel(),
        createResponseTimePanel(),
        createThroughputPanel(),
        createErrorRatePanel(),
        createMemoryPanel(),
        createGoroutinesPanel(),
        createGCPanel(),
        createCPUPanel(),
        createDatabasePanel(),
        createCachePanel(),
    }
    
    // Add templating variables
    dashboard.Templating = createTemplatingConfig()
    
    return dashboard
}

func createOverviewPanel() Panel {
    return Panel{
        ID:    1,
        Title: "Application Overview",
        Type:  "stat",
        GridPos: GridPos{
            X: 0, Y: 0, W: 24, H: 4,
        },
        Targets: []Target{
            {
                Expr:         "up{job=\"go-app\"}",
                RefID:        "A",
                LegendFormat: "Service Status",
            },
            {
                Expr:         "rate(app_http_requests_total[5m])",
                RefID:        "B", 
                LegendFormat: "RPS",
            },
            {
                Expr:         "histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))",
                RefID:        "C",
                LegendFormat: "P95 Latency",
            },
            {
                Expr:         "rate(app_http_requests_total{status=~\"5..\"}[5m]) / rate(app_http_requests_total[5m]) * 100",
                RefID:        "D",
                LegendFormat: "Error Rate %",
            },
        },
        FieldConfig: FieldConfig{
            Defaults: FieldDefaults{
                Unit: "short",
                Thresholds: Thresholds{
                    Steps: []ThresholdStep{
                        {Color: "green", Value: 0},
                        {Color: "yellow", Value: 80},
                        {Color: "red", Value: 95},
                    },
                },
            },
        },
    }
}

func createResponseTimePanel() Panel {
    return Panel{
        ID:    2,
        Title: "Response Time Distribution",
        Type:  "timeseries",
        GridPos: GridPos{
            X: 0, Y: 4, W: 12, H: 8,
        },
        Targets: []Target{
            {
                Expr:         "histogram_quantile(0.50, rate(app_http_request_duration_seconds_bucket[5m]))",
                RefID:        "A",
                LegendFormat: "P50",
            },
            {
                Expr:         "histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))",
                RefID:        "B",
                LegendFormat: "P95",
            },
            {
                Expr:         "histogram_quantile(0.99, rate(app_http_request_duration_seconds_bucket[5m]))",
                RefID:        "C",
                LegendFormat: "P99",
            },
        },
        FieldConfig: FieldConfig{
            Defaults: FieldDefaults{
                Unit: "s",
            },
        },
        Alert: &AlertConfig{
            Name:       "High Response Time",
            Message:    "P95 response time is above threshold",
            Frequency:  "30s",
            Conditions: []AlertCondition{
                {
                    Query: AlertQuery{
                        RefID: "B",
                        Model: map[string]interface{}{
                            "expr": "histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))",
                        },
                    },
                    Reducer: AlertReducer{
                        Type: "last",
                    },
                    Evaluator: AlertEvaluator{
                        Type:   "gt",
                        Params: []float64{0.5}, // 500ms threshold
                    },
                },
            },
        },
    }
}

func createMemoryPanel() Panel {
    return Panel{
        ID:    5,
        Title: "Memory Usage",
        Type:  "timeseries",
        GridPos: GridPos{
            X: 0, Y: 12, W: 12, H: 8,
        },
        Targets: []Target{
            {
                Expr:         "go_memory_heap_inuse_bytes",
                RefID:        "A",
                LegendFormat: "Heap In Use",
            },
            {
                Expr:         "go_memory_heap_bytes",
                RefID:        "B",
                LegendFormat: "Heap Size",
            },
            {
                Expr:         "go_memory_stack_bytes",
                RefID:        "C",
                LegendFormat: "Stack Size",
            },
            {
                Expr:         "rate(go_memory_allocations_bytes_total[5m])",
                RefID:        "D",
                LegendFormat: "Allocation Rate",
            },
        },
        FieldConfig: FieldConfig{
            Defaults: FieldDefaults{
                Unit: "bytes",
            },
        },
    }
}

func createGCPanel() Panel {
    return Panel{
        ID:    7,
        Title: "Garbage Collection",
        Type:  "timeseries",
        GridPos: GridPos{
            X: 12, Y: 12, W: 12, H: 8,
        },
        Targets: []Target{
            {
                Expr:         "rate(go_runtime_gc_duration_seconds_sum[5m])",
                RefID:        "A",
                LegendFormat: "GC Duration",
            },
            {
                Expr:         "rate(go_runtime_gc_runs_total[5m])",
                RefID:        "B",
                LegendFormat: "GC Frequency",
            },
            {
                Expr:         "go_runtime_gc_pause_ns",
                RefID:        "C",
                LegendFormat: "GC Pause",
            },
        },
        FieldConfig: FieldConfig{
            Defaults: FieldDefaults{
                Unit: "s",
            },
        },
    }
}

// Dashboard export and provisioning
func (gd *GrafanaDashboard) ExportJSON() ([]byte, error) {
    return json.MarshalIndent(gd, "", "  ")
}

func ProvisionDashboard(dashboardDir string) error {
    dashboard := CreateGoPerformanceDashboard()
    
    jsonData, err := dashboard.ExportJSON()
    if err != nil {
        return err
    }
    
    filename := fmt.Sprintf("%s/go-performance-dashboard.json", dashboardDir)
    return os.WriteFile(filename, jsonData, 0644)
}

// Alert rule definitions
func CreateGoPerformanceAlerts() []AlertRule {
    return []AlertRule{
        {
            Alert:       "HighResponseTime",
            Expr:        "histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m])) > 0.5",
            For:         "2m",
            Labels:      map[string]string{"severity": "warning"},
            Annotations: map[string]string{
                "summary":     "High response time detected",
                "description": "P95 response time is {{ $value }}s",
            },
        },
        {
            Alert:       "HighErrorRate", 
            Expr:        "rate(app_http_requests_total{status=~\"5..\"}[5m]) / rate(app_http_requests_total[5m]) > 0.05",
            For:         "1m",
            Labels:      map[string]string{"severity": "critical"},
            Annotations: map[string]string{
                "summary":     "High error rate detected",
                "description": "Error rate is {{ $value | humanizePercentage }}",
            },
        },
        {
            Alert:       "MemoryLeakSuspected",
            Expr:        "increase(go_memory_heap_inuse_bytes[1h]) > 100*1024*1024", // 100MB increase per hour
            For:         "5m",
            Labels:      map[string]string{"severity": "warning"},
            Annotations: map[string]string{
                "summary":     "Potential memory leak detected",
                "description": "Memory usage increased by {{ $value | humanizeBytes }} in the last hour",
            },
        },
        {
            Alert:       "GoroutineLeak",
            Expr:        "increase(go_runtime_goroutines_total[10m]) > 1000",
            For:         "2m",
            Labels:      map[string]string{"severity": "critical"},
            Annotations: map[string]string{
                "summary":     "Goroutine leak detected",
                "description": "Goroutine count increased by {{ $value }} in 10 minutes",
            },
        },
    }
}
```

This comprehensive tools and frameworks section provides practical implementations for essential Go performance monitoring, from basic profiling to advanced distributed monitoring systems with Prometheus and Grafana integration.
