# Stack vs Heap Optimization

Comprehensive guide to understanding and optimizing memory allocation between stack and heap in Go. This guide covers escape analysis, allocation strategies, and performance optimization techniques.

## Table of Contents

- [Introduction](#introduction)
- [Memory Model Fundamentals](#memory-model-fundamentals)
- [Escape Analysis](#escape-analysis)
- [Stack Allocation Optimization](#stack-allocation-optimization)
- [Heap Allocation Control](#heap-allocation-control)
- [Performance Analysis](#performance-analysis)
- [Optimization Strategies](#optimization-strategies)
- [Profiling and Monitoring](#profiling-and-monitoring)
- [Best Practices](#best-practices)

## Introduction

Understanding stack vs heap allocation is crucial for Go performance optimization. This guide provides comprehensive strategies for optimizing memory allocation patterns to minimize garbage collection pressure and maximize performance.

### Memory Allocation Framework

```go
package main

import (
    "fmt"
    "reflect"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// MemoryAnalyzer provides comprehensive memory allocation analysis
type MemoryAnalyzer struct {
    config       AnalysisConfig
    profiler     *AllocationProfiler
    escapeTracker *EscapeTracker
    optimizer    *AllocationOptimizer
    monitor      *MemoryMonitor
    metrics      *AllocationMetrics
    mu           sync.RWMutex
}

// AnalysisConfig contains memory analysis configuration
type AnalysisConfig struct {
    TrackAllocations  bool
    TrackEscapes      bool
    SampleRate       float64
    ProfileDuration  time.Duration
    OptimizationLevel int
    EnableProfiling  bool
    ReportThreshold  int64
}

// AllocationProfiler profiles memory allocations
type AllocationProfiler struct {
    stackAllocations map[string]*AllocationInfo
    heapAllocations  map[string]*AllocationInfo
    escapeAnalysis   map[string]*EscapeInfo
    callSites        map[string]*CallSiteInfo
    hotspots         []*AllocationHotspot
    baseline         runtime.MemStats
}

// AllocationInfo tracks allocation details
type AllocationInfo struct {
    Location     string
    Type         reflect.Type
    Size         int64
    Count        int64
    TotalSize    int64
    AllocType    AllocationType
    CallStack    []uintptr
    FirstSeen    time.Time
    LastSeen     time.Time
    EscapeReason string
}

// AllocationType defines allocation location
type AllocationType int

const (
    StackAllocation AllocationType = iota
    HeapAllocation
    EscapedAllocation
    InlineAllocation
)

// EscapeInfo tracks escape analysis information
type EscapeInfo struct {
    Function     string
    Variable     string
    Type         string
    Size         int64
    EscapeReason EscapeReason
    CallSite     string
    Frequency    int64
    CanOptimize  bool
}

// EscapeReason defines why a variable escapes
type EscapeReason int

const (
    UnknownEscape EscapeReason = iota
    ReturnedPointer
    StoredInGlobal
    PassedToInterface
    ClosureCaptured
    ChannelSend
    SliceGrowth
    MapOperation
    InterfaceConversion
    GoroutineCapture
)

// CallSiteInfo tracks allocation call sites
type CallSiteInfo struct {
    Function      string
    File          string
    Line          int
    AllocCount    int64
    TotalSize     int64
    HeapAllocRate float64
    Optimizable   bool
}

// AllocationHotspot identifies allocation hotspots
type AllocationHotspot struct {
    Location       string
    Function       string
    Type           string
    AllocationsPerSec float64
    BytesPerSec    float64
    HeapPressure   float64
    Severity       HotspotSeverity
}

// HotspotSeverity defines hotspot severity levels
type HotspotSeverity int

const (
    LowSeverity HotspotSeverity = iota
    MediumSeverity
    HighSeverity
    CriticalSeverity
)

// EscapeTracker tracks variable escapes
type EscapeTracker struct {
    escapes       map[string]*EscapeEvent
    patterns      map[string]*EscapePattern
    suggestions   []*OptimizationSuggestion
    analysis      *EscapeAnalysisResult
}

// EscapeEvent represents an escape event
type EscapeEvent struct {
    Variable     string
    Function     string
    Type         string
    Size         int64
    Reason       EscapeReason
    StackTrace   []uintptr
    Timestamp    time.Time
    Preventable  bool
}

// EscapePattern identifies common escape patterns
type EscapePattern struct {
    Pattern      string
    Frequency    int64
    TotalSize    int64
    Preventable  bool
    Optimization string
}

// OptimizationSuggestion provides optimization recommendations
type OptimizationSuggestion struct {
    Type            SuggestionType
    Location        string
    Description     string
    PotentialSaving int64
    Difficulty      OptimizationDifficulty
    Implementation  string
    Example         string
}

// SuggestionType defines suggestion types
type SuggestionType int

const (
    AvoidEscape SuggestionType = iota
    UseStackAllocation
    ReducePointerChasing
    OptimizeDataStructure
    ImplementObjectPool
    UseValueReceiver
    AvoidSliceGrowth
    OptimizeInterface
)

// OptimizationDifficulty defines implementation difficulty
type OptimizationDifficulty int

const (
    EasyOptimization OptimizationDifficulty = iota
    MediumOptimization
    HardOptimization
)

// AllocationOptimizer optimizes allocation patterns
type AllocationOptimizer struct {
    strategies   []OptimizationStrategy
    patterns     map[string]*AllocationPattern
    transforms   []CodeTransform
    results      *OptimizationResult
}

// OptimizationStrategy defines optimization strategies
type OptimizationStrategy interface {
    Analyze(info *AllocationInfo) *OptimizationSuggestion
    Apply(code string) (string, error)
    Estimate(info *AllocationInfo) int64
}

// AllocationPattern identifies allocation patterns
type AllocationPattern struct {
    Name          string
    Pattern       string
    Frequency     int64
    AverageSize   int64
    StackEligible bool
    Optimization  string
}

// CodeTransform represents code transformations
type CodeTransform struct {
    Pattern     string
    Replacement string
    Description string
    Impact      TransformImpact
}

// TransformImpact measures transformation impact
type TransformImpact struct {
    AllocationReduction int64
    PerformanceGain     float64
    ComplexityChange    int
}

// OptimizationResult contains optimization results
type OptimizationResult struct {
    TotalSavings     int64
    AllocationReduction float64
    HeapPressureReduction float64
    AppliedOptimizations []string
    FailedOptimizations  []string
}

// MemoryMonitor monitors memory allocation patterns
type MemoryMonitor struct {
    allocations  chan AllocationEvent
    metrics      *AllocationMetrics
    collectors   []AllocationCollector
    alerting     *AllocationAlerting
    running      bool
}

// AllocationEvent represents a memory allocation event
type AllocationEvent struct {
    Type        AllocationType
    Size        int64
    Location    string
    Function    string
    StackTrace  []uintptr
    Timestamp   time.Time
}

// AllocationMetrics tracks allocation metrics
type AllocationMetrics struct {
    StackAllocations    int64
    HeapAllocations     int64
    TotalStackBytes     int64
    TotalHeapBytes      int64
    EscapeRate          float64
    AllocationRate      float64
    GCPressure          float64
    MemoryEfficiency    float64
}

// AllocationCollector collects allocation metrics
type AllocationCollector interface {
    CollectAllocation(event AllocationEvent)
    GetMetrics() map[string]interface{}
    Reset()
}

// AllocationAlerting provides alerting for allocation issues
type AllocationAlerting struct {
    thresholds   AllocationThresholds
    alerts       chan AllocationAlert
    handlers     []AllocationAlertHandler
}

// AllocationThresholds defines alerting thresholds
type AllocationThresholds struct {
    MaxEscapeRate      float64
    MaxAllocationRate  float64
    MaxHeapPressure    float64
    MaxGCFrequency     int
}

// AllocationAlert represents an allocation alert
type AllocationAlert struct {
    Type        AlertType
    Severity    AlertSeverity
    Message     string
    Metrics     map[string]interface{}
    Timestamp   time.Time
}

// AlertType defines alert types
type AlertType int

const (
    EscapeRateAlert AlertType = iota
    AllocationRateAlert
    HeapPressureAlert
    GCFrequencyAlert
)

// AlertSeverity defines alert severity
type AlertSeverity int

const (
    InfoAlert AlertSeverity = iota
    WarningAlert
    ErrorAlert
    CriticalAlert
)

// AllocationAlertHandler handles allocation alerts
type AllocationAlertHandler interface {
    HandleAlert(alert AllocationAlert) error
}

// NewMemoryAnalyzer creates a new memory analyzer
func NewMemoryAnalyzer(config AnalysisConfig) *MemoryAnalyzer {
    return &MemoryAnalyzer{
        config:        config,
        profiler:      NewAllocationProfiler(),
        escapeTracker: NewEscapeTracker(),
        optimizer:     NewAllocationOptimizer(),
        monitor:       NewMemoryMonitor(),
        metrics:       &AllocationMetrics{},
    }
}

// NewAllocationProfiler creates a new allocation profiler
func NewAllocationProfiler() *AllocationProfiler {
    profiler := &AllocationProfiler{
        stackAllocations: make(map[string]*AllocationInfo),
        heapAllocations:  make(map[string]*AllocationInfo),
        escapeAnalysis:   make(map[string]*EscapeInfo),
        callSites:        make(map[string]*CallSiteInfo),
        hotspots:         make([]*AllocationHotspot, 0),
    }
    
    runtime.ReadMemStats(&profiler.baseline)
    return profiler
}

// AnalyzeAllocations performs comprehensive allocation analysis
func (ma *MemoryAnalyzer) AnalyzeAllocations() (*AllocationAnalysisResult, error) {
    ma.mu.Lock()
    defer ma.mu.Unlock()
    
    // Start profiling
    if err := ma.startProfiling(); err != nil {
        return nil, fmt.Errorf("failed to start profiling: %w", err)
    }
    
    // Collect allocation data
    if err := ma.collectAllocationData(); err != nil {
        return nil, fmt.Errorf("failed to collect allocation data: %w", err)
    }
    
    // Analyze escape patterns
    escapeAnalysis, err := ma.escapeTracker.analyzeEscapes()
    if err != nil {
        return nil, fmt.Errorf("escape analysis failed: %w", err)
    }
    
    // Identify optimization opportunities
    optimizations := ma.optimizer.identifyOptimizations(ma.profiler)
    
    // Generate recommendations
    suggestions := ma.generateSuggestions(escapeAnalysis, optimizations)
    
    return &AllocationAnalysisResult{
        StackAllocations:  ma.profiler.stackAllocations,
        HeapAllocations:   ma.profiler.heapAllocations,
        EscapeAnalysis:    escapeAnalysis,
        Optimizations:     optimizations,
        Suggestions:       suggestions,
        Metrics:          *ma.metrics,
    }, nil
}

// AllocationAnalysisResult contains allocation analysis results
type AllocationAnalysisResult struct {
    StackAllocations  map[string]*AllocationInfo
    HeapAllocations   map[string]*AllocationInfo
    EscapeAnalysis    *EscapeAnalysisResult
    Optimizations     []*OptimizationSuggestion
    Suggestions       []*OptimizationSuggestion
    Metrics           AllocationMetrics
}

// EscapeAnalysisResult contains escape analysis results
type EscapeAnalysisResult struct {
    TotalEscapes     int64
    EscapesByReason  map[EscapeReason]int64
    PreventableEscapes int64
    Patterns         map[string]*EscapePattern
    Impact           EscapeImpact
}

// EscapeImpact measures escape impact
type EscapeImpact struct {
    MemoryOverhead   int64
    GCPressure       float64
    AllocationRate   float64
    PerformanceImpact float64
}

// startProfiling starts allocation profiling
func (ma *MemoryAnalyzer) startProfiling() error {
    if !ma.config.EnableProfiling {
        return nil
    }
    
    // Set up memory profiling
    runtime.MemProfileRate = int(1.0 / ma.config.SampleRate)
    
    // Start monitoring
    if err := ma.monitor.Start(); err != nil {
        return fmt.Errorf("failed to start monitor: %w", err)
    }
    
    return nil
}

// collectAllocationData collects allocation data
func (ma *MemoryAnalyzer) collectAllocationData() error {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    // Calculate allocation metrics
    ma.metrics.HeapAllocations = int64(memStats.Mallocs)
    ma.metrics.TotalHeapBytes = int64(memStats.TotalAlloc)
    ma.metrics.AllocationRate = float64(memStats.Mallocs-ma.profiler.baseline.Mallocs) / 
                               ma.config.ProfileDuration.Seconds()
    
    // Profile current allocations
    if err := ma.profileCurrentAllocations(); err != nil {
        return fmt.Errorf("failed to profile allocations: %w", err)
    }
    
    return nil
}

// profileCurrentAllocations profiles current memory allocations
func (ma *MemoryAnalyzer) profileCurrentAllocations() error {
    // Get memory profile
    profile := runtime.MemProfile(nil, true)
    
    for _, record := range profile {
        location := fmt.Sprintf("%p", record.Stack()[0])
        
        info := &AllocationInfo{
            Location:  location,
            Size:      int64(record.AllocBytes),
            Count:     int64(record.AllocObjects),
            TotalSize: int64(record.AllocBytes),
            CallStack: record.Stack(),
            FirstSeen: time.Now(),
            LastSeen:  time.Now(),
        }
        
        // Determine allocation type
        if ma.isStackAllocation(record) {
            info.AllocType = StackAllocation
            ma.profiler.stackAllocations[location] = info
        } else {
            info.AllocType = HeapAllocation
            ma.profiler.heapAllocations[location] = info
        }
    }
    
    return nil
}

// isStackAllocation determines if allocation is on stack
func (ma *MemoryAnalyzer) isStackAllocation(record runtime.MemProfileRecord) bool {
    // Simplified heuristic - in practice, this would use escape analysis
    return record.AllocBytes < 1024 && record.InUseObjects == 0
}

// NewEscapeTracker creates a new escape tracker
func NewEscapeTracker() *EscapeTracker {
    return &EscapeTracker{
        escapes:     make(map[string]*EscapeEvent),
        patterns:    make(map[string]*EscapePattern),
        suggestions: make([]*OptimizationSuggestion, 0),
    }
}

// analyzeEscapes analyzes variable escapes
func (et *EscapeTracker) analyzeEscapes() (*EscapeAnalysisResult, error) {
    result := &EscapeAnalysisResult{
        EscapesByReason:    make(map[EscapeReason]int64),
        Patterns:          make(map[string]*EscapePattern),
    }
    
    // Analyze escape events
    preventableCount := int64(0)
    totalEscapes := int64(len(et.escapes))
    
    for _, escape := range et.escapes {
        result.EscapesByReason[escape.Reason]++
        
        if escape.Preventable {
            preventableCount++
        }
        
        // Update patterns
        pattern := et.getEscapePattern(escape)
        if existing, found := et.patterns[pattern]; found {
            existing.Frequency++
            existing.TotalSize += escape.Size
        } else {
            et.patterns[pattern] = &EscapePattern{
                Pattern:     pattern,
                Frequency:   1,
                TotalSize:   escape.Size,
                Preventable: escape.Preventable,
            }
        }
    }
    
    result.TotalEscapes = totalEscapes
    result.PreventableEscapes = preventableCount
    result.Patterns = et.patterns
    
    // Calculate impact
    result.Impact = et.calculateEscapeImpact()
    
    return result, nil
}

// getEscapePattern extracts escape pattern from event
func (et *EscapeTracker) getEscapePattern(escape *EscapeEvent) string {
    switch escape.Reason {
    case ReturnedPointer:
        return "returned_pointer"
    case StoredInGlobal:
        return "stored_global"
    case PassedToInterface:
        return "interface_conversion"
    case ClosureCaptured:
        return "closure_capture"
    case ChannelSend:
        return "channel_operation"
    case SliceGrowth:
        return "slice_growth"
    case MapOperation:
        return "map_operation"
    case InterfaceConversion:
        return "interface_conversion"
    case GoroutineCapture:
        return "goroutine_capture"
    default:
        return "unknown"
    }
}

// calculateEscapeImpact calculates the impact of escapes
func (et *EscapeTracker) calculateEscapeImpact() EscapeImpact {
    totalMemory := int64(0)
    totalEscapes := int64(0)
    
    for _, escape := range et.escapes {
        totalMemory += escape.Size
        totalEscapes++
    }
    
    // Simplified impact calculation
    gcPressure := float64(totalMemory) / (1024 * 1024) // MB
    allocationRate := float64(totalEscapes) / time.Minute.Seconds()
    
    return EscapeImpact{
        MemoryOverhead:    totalMemory,
        GCPressure:       gcPressure,
        AllocationRate:   allocationRate,
        PerformanceImpact: gcPressure * 0.1, // Simplified calculation
    }
}

// NewAllocationOptimizer creates a new allocation optimizer
func NewAllocationOptimizer() *AllocationOptimizer {
    optimizer := &AllocationOptimizer{
        strategies: make([]OptimizationStrategy, 0),
        patterns:   make(map[string]*AllocationPattern),
        transforms: make([]CodeTransform, 0),
    }
    
    // Initialize optimization strategies
    optimizer.initializeStrategies()
    
    return optimizer
}

// initializeStrategies initializes optimization strategies
func (ao *AllocationOptimizer) initializeStrategies() {
    // Add common optimization strategies
    ao.strategies = append(ao.strategies, &StackAllocationStrategy{})
    ao.strategies = append(ao.strategies, &PointerAvoidanceStrategy{})
    ao.strategies = append(ao.strategies, &InterfaceOptimizationStrategy{})
    ao.strategies = append(ao.strategies, &SliceOptimizationStrategy{})
}

// identifyOptimizations identifies optimization opportunities
func (ao *AllocationOptimizer) identifyOptimizations(profiler *AllocationProfiler) []*OptimizationSuggestion {
    suggestions := make([]*OptimizationSuggestion, 0)
    
    // Analyze heap allocations for optimization opportunities
    for _, allocation := range profiler.heapAllocations {
        for _, strategy := range ao.strategies {
            if suggestion := strategy.Analyze(allocation); suggestion != nil {
                suggestions = append(suggestions, suggestion)
            }
        }
    }
    
    // Sort suggestions by potential savings
    ao.sortSuggestionsByImpact(suggestions)
    
    return suggestions
}

// sortSuggestionsByImpact sorts suggestions by potential impact
func (ao *AllocationOptimizer) sortSuggestionsByImpact(suggestions []*OptimizationSuggestion) {
    for i := 0; i < len(suggestions)-1; i++ {
        for j := i + 1; j < len(suggestions); j++ {
            if suggestions[i].PotentialSaving < suggestions[j].PotentialSaving {
                suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
            }
        }
    }
}

// StackAllocationStrategy optimizes for stack allocation
type StackAllocationStrategy struct{}

// Analyze analyzes allocation for stack optimization
func (sas *StackAllocationStrategy) Analyze(info *AllocationInfo) *OptimizationSuggestion {
    if info.AllocType == HeapAllocation && info.Size < 1024 {
        return &OptimizationSuggestion{
            Type:            UseStackAllocation,
            Location:        info.Location,
            Description:     "Small allocation can be moved to stack",
            PotentialSaving: info.TotalSize,
            Difficulty:      EasyOptimization,
            Implementation:  "Use value types instead of pointers",
            Example:         "var x MyStruct instead of x := &MyStruct{}",
        }
    }
    return nil
}

// Apply applies stack allocation optimization
func (sas *StackAllocationStrategy) Apply(code string) (string, error) {
    // Simplified code transformation
    return code, nil
}

// Estimate estimates optimization impact
func (sas *StackAllocationStrategy) Estimate(info *AllocationInfo) int64 {
    return info.TotalSize
}

// PointerAvoidanceStrategy optimizes pointer usage
type PointerAvoidanceStrategy struct{}

// Analyze analyzes allocation for pointer optimization
func (pas *PointerAvoidanceStrategy) Analyze(info *AllocationInfo) *OptimizationSuggestion {
    if info.EscapeReason == "returned_pointer" {
        return &OptimizationSuggestion{
            Type:            AvoidEscape,
            Location:        info.Location,
            Description:     "Avoid returning pointer to local variable",
            PotentialSaving: info.TotalSize,
            Difficulty:      MediumOptimization,
            Implementation:  "Return value instead of pointer",
            Example:         "func() MyStruct instead of func() *MyStruct",
        }
    }
    return nil
}

// Apply applies pointer avoidance optimization
func (pas *PointerAvoidanceStrategy) Apply(code string) (string, error) {
    return code, nil
}

// Estimate estimates optimization impact
func (pas *PointerAvoidanceStrategy) Estimate(info *AllocationInfo) int64 {
    return info.TotalSize * 2 // Account for GC overhead
}

// InterfaceOptimizationStrategy optimizes interface usage
type InterfaceOptimizationStrategy struct{}

// Analyze analyzes allocation for interface optimization
func (ios *InterfaceOptimizationStrategy) Analyze(info *AllocationInfo) *OptimizationSuggestion {
    if info.EscapeReason == "interface_conversion" && info.Size < 256 {
        return &OptimizationSuggestion{
            Type:            OptimizeInterface,
            Location:        info.Location,
            Description:     "Small type escapes through interface conversion",
            PotentialSaving: info.TotalSize,
            Difficulty:      MediumOptimization,
            Implementation:  "Use concrete types or type assertions",
            Example:         "Avoid fmt.Printf with small types",
        }
    }
    return nil
}

// Apply applies interface optimization
func (ios *InterfaceOptimizationStrategy) Apply(code string) (string, error) {
    return code, nil
}

// Estimate estimates optimization impact
func (ios *InterfaceOptimizationStrategy) Estimate(info *AllocationInfo) int64 {
    return info.TotalSize
}

// SliceOptimizationStrategy optimizes slice usage
type SliceOptimizationStrategy struct{}

// Analyze analyzes allocation for slice optimization
func (sos *SliceOptimizationStrategy) Analyze(info *AllocationInfo) *OptimizationSuggestion {
    if info.EscapeReason == "slice_growth" {
        return &OptimizationSuggestion{
            Type:            AvoidSliceGrowth,
            Location:        info.Location,
            Description:     "Slice growth causes allocations",
            PotentialSaving: info.TotalSize / 2, // Estimate savings
            Difficulty:      EasyOptimization,
            Implementation:  "Pre-allocate slice with known capacity",
            Example:         "make([]T, 0, expectedSize)",
        }
    }
    return nil
}

// Apply applies slice optimization
func (sos *SliceOptimizationStrategy) Apply(code string) (string, error) {
    return code, nil
}

// Estimate estimates optimization impact
func (sos *SliceOptimizationStrategy) Estimate(info *AllocationInfo) int64 {
    return info.TotalSize / 2
}

// generateSuggestions generates optimization suggestions
func (ma *MemoryAnalyzer) generateSuggestions(escapeAnalysis *EscapeAnalysisResult, 
                                              optimizations []*OptimizationSuggestion) []*OptimizationSuggestion {
    suggestions := make([]*OptimizationSuggestion, 0)
    
    // Add optimization suggestions
    suggestions = append(suggestions, optimizations...)
    
    // Add escape-specific suggestions
    for reason, count := range escapeAnalysis.EscapesByReason {
        if count > 100 { // Threshold for suggesting optimization
            suggestion := ma.createEscapeSuggestion(reason, count)
            if suggestion != nil {
                suggestions = append(suggestions, suggestion)
            }
        }
    }
    
    return suggestions
}

// createEscapeSuggestion creates suggestion for escape reason
func (ma *MemoryAnalyzer) createEscapeSuggestion(reason EscapeReason, count int64) *OptimizationSuggestion {
    switch reason {
    case ReturnedPointer:
        return &OptimizationSuggestion{
            Type:            AvoidEscape,
            Description:     "High frequency of returned pointer escapes",
            PotentialSaving: count * 64, // Estimate
            Difficulty:      MediumOptimization,
            Implementation:  "Return values instead of pointers where possible",
        }
    case InterfaceConversion:
        return &OptimizationSuggestion{
            Type:            OptimizeInterface,
            Description:     "High frequency of interface conversion escapes",
            PotentialSaving: count * 32, // Estimate
            Difficulty:      MediumOptimization,
            Implementation:  "Use type assertions or avoid unnecessary conversions",
        }
    case SliceGrowth:
        return &OptimizationSuggestion{
            Type:            AvoidSliceGrowth,
            Description:     "High frequency of slice growth allocations",
            PotentialSaving: count * 128, // Estimate
            Difficulty:      EasyOptimization,
            Implementation:  "Pre-allocate slices with appropriate capacity",
        }
    default:
        return nil
    }
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor() *MemoryMonitor {
    return &MemoryMonitor{
        allocations: make(chan AllocationEvent, 10000),
        metrics:     &AllocationMetrics{},
        collectors:  make([]AllocationCollector, 0),
        alerting:    NewAllocationAlerting(),
    }
}

// Start starts the memory monitor
func (mm *MemoryMonitor) Start() error {
    mm.running = true
    go mm.monitorLoop()
    return nil
}

// monitorLoop processes allocation events
func (mm *MemoryMonitor) monitorLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for mm.running {
        select {
        case event := <-mm.allocations:
            mm.processAllocationEvent(event)
            
        case <-ticker.C:
            mm.updateMetrics()
            mm.checkAlerts()
        }
    }
}

// processAllocationEvent processes a single allocation event
func (mm *MemoryMonitor) processAllocationEvent(event AllocationEvent) {
    // Update allocation counters
    switch event.Type {
    case StackAllocation:
        atomic.AddInt64(&mm.metrics.StackAllocations, 1)
        atomic.AddInt64(&mm.metrics.TotalStackBytes, event.Size)
    case HeapAllocation:
        atomic.AddInt64(&mm.metrics.HeapAllocations, 1)
        atomic.AddInt64(&mm.metrics.TotalHeapBytes, event.Size)
    }
    
    // Notify collectors
    for _, collector := range mm.collectors {
        collector.CollectAllocation(event)
    }
}

// updateMetrics updates allocation metrics
func (mm *MemoryMonitor) updateMetrics() {
    total := mm.metrics.StackAllocations + mm.metrics.HeapAllocations
    if total > 0 {
        mm.metrics.EscapeRate = float64(mm.metrics.HeapAllocations) / float64(total)
    }
    
    mm.metrics.AllocationRate = float64(total) / time.Minute.Seconds()
}

// checkAlerts checks for allocation alerts
func (mm *MemoryMonitor) checkAlerts() {
    if mm.metrics.EscapeRate > mm.alerting.thresholds.MaxEscapeRate {
        alert := AllocationAlert{
            Type:      EscapeRateAlert,
            Severity:  WarningAlert,
            Message:   fmt.Sprintf("High escape rate: %.2f%%", mm.metrics.EscapeRate*100),
            Timestamp: time.Now(),
        }
        mm.alerting.SendAlert(alert)
    }
}

// NewAllocationAlerting creates a new allocation alerting system
func NewAllocationAlerting() *AllocationAlerting {
    return &AllocationAlerting{
        thresholds: AllocationThresholds{
            MaxEscapeRate:     0.3,  // 30%
            MaxAllocationRate: 1000, // 1000 allocs/min
            MaxHeapPressure:   0.8,  // 80%
            MaxGCFrequency:    10,   // 10 GC/min
        },
        alerts:   make(chan AllocationAlert, 1000),
        handlers: make([]AllocationAlertHandler, 0),
    }
}

// SendAlert sends an allocation alert
func (aa *AllocationAlerting) SendAlert(alert AllocationAlert) {
    select {
    case aa.alerts <- alert:
    default:
        // Alert channel full
    }
    
    // Notify handlers
    for _, handler := range aa.handlers {
        go handler.HandleAlert(alert)
    }
}

// GetMetrics returns current allocation metrics
func (ma *MemoryAnalyzer) GetMetrics() AllocationMetrics {
    ma.mu.RLock()
    defer ma.mu.RUnlock()
    return *ma.metrics
}

// OptimizeFunction provides function-level optimization suggestions
func (ma *MemoryAnalyzer) OptimizeFunction(funcName string) ([]*OptimizationSuggestion, error) {
    suggestions := make([]*OptimizationSuggestion, 0)
    
    // Analyze function allocations
    for _, allocation := range ma.profiler.heapAllocations {
        if allocation.Location == funcName {
            for _, strategy := range ma.optimizer.strategies {
                if suggestion := strategy.Analyze(allocation); suggestion != nil {
                    suggestions = append(suggestions, suggestion)
                }
            }
        }
    }
    
    return suggestions, nil
}

// GetStackUsage analyzes stack usage patterns
func (ma *MemoryAnalyzer) GetStackUsage() map[string]int64 {
    usage := make(map[string]int64)
    
    for location, allocation := range ma.profiler.stackAllocations {
        usage[location] = allocation.TotalSize
    }
    
    return usage
}

// GetHeapUsage analyzes heap usage patterns
func (ma *MemoryAnalyzer) GetHeapUsage() map[string]int64 {
    usage := make(map[string]int64)
    
    for location, allocation := range ma.profiler.heapAllocations {
        usage[location] = allocation.TotalSize
    }
    
    return usage
}

// GetEscapeHotspots identifies escape hotspots
func (ma *MemoryAnalyzer) GetEscapeHotspots() []*AllocationHotspot {
    hotspots := make([]*AllocationHotspot, 0)
    
    for location, allocation := range ma.profiler.heapAllocations {
        if allocation.Count > 1000 { // Threshold
            hotspot := &AllocationHotspot{
                Location:          location,
                AllocationsPerSec: float64(allocation.Count) / time.Minute.Seconds(),
                BytesPerSec:      float64(allocation.TotalSize) / time.Minute.Seconds(),
                HeapPressure:     float64(allocation.TotalSize) / (1024 * 1024),
                Severity:         ma.calculateHotspotSeverity(allocation),
            }
            hotspots = append(hotspots, hotspot)
        }
    }
    
    return hotspots
}

// calculateHotspotSeverity calculates hotspot severity
func (ma *MemoryAnalyzer) calculateHotspotSeverity(allocation *AllocationInfo) HotspotSeverity {
    if allocation.TotalSize > 10*1024*1024 { // > 10MB
        return CriticalSeverity
    } else if allocation.TotalSize > 1024*1024 { // > 1MB
        return HighSeverity
    } else if allocation.TotalSize > 100*1024 { // > 100KB
        return MediumSeverity
    }
    return LowSeverity
}
```

## Performance Analysis

Advanced techniques for analyzing stack vs heap performance characteristics.

### Allocation Profiling

Detailed allocation profiling to understand memory patterns.

### Escape Analysis Tools

Tools and techniques for understanding escape analysis results.

### Performance Benchmarking

Benchmarking methodologies for comparing stack and heap performance.

## Optimization Strategies

Proven strategies for optimizing stack vs heap allocation.

### Value vs Pointer Types

Guidelines for choosing between value and pointer types.

### Interface Optimization

Optimizing interface usage to reduce allocations.

### Slice and Map Optimization

Efficient slice and map usage patterns.

## Best Practices

1. **Prefer Stack Allocation**: Use value types when possible
2. **Avoid Unnecessary Pointers**: Return values instead of pointers for small types
3. **Pre-allocate Slices**: Use make([]T, 0, capacity) for known sizes
4. **Minimize Interface Conversions**: Avoid unnecessary interface{} usage
5. **Profile Regularly**: Use allocation profiling to identify hotspots
6. **Understand Escape Analysis**: Learn when and why variables escape
7. **Optimize Hot Paths**: Focus optimization efforts on frequently called code
8. **Use Object Pools**: For frequently allocated objects that escape

## Summary

Stack vs heap optimization is fundamental to Go performance:

1. **Understanding**: Know when variables escape to heap
2. **Analysis**: Use profiling tools to identify allocation patterns
3. **Optimization**: Apply systematic optimization strategies
4. **Monitoring**: Continuously monitor allocation behavior
5. **Measurement**: Benchmark optimization impact

These techniques enable developers to write memory-efficient Go applications with minimal garbage collection pressure.
