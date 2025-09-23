# Function Inlining Optimization

Function inlining is one of the most powerful compiler optimizations in Go. It replaces function calls with the actual function body, eliminating call overhead and enabling further optimizations. This chapter provides comprehensive coverage of Go's inlining behavior, analysis techniques, and optimization strategies.

## Table of Contents

- [Understanding Function Inlining](#understanding-function-inlining)
- [Go Compiler Inlining Rules](#go-compiler-inlining-rules)
- [Inlining Analysis Tools](#inlining-analysis-tools)
- [Optimization Strategies](#optimization-strategies)
- [Advanced Techniques](#advanced-techniques)
- [Performance Impact](#performance-impact)
- [Best Practices](#best-practices)

## Understanding Function Inlining

Function inlining replaces function calls with the function's body at compile time, eliminating the overhead of function calls and enabling additional optimizations.

### Benefits of Inlining

```go
// Before inlining
func add(a, b int) int {
    return a + b
}

func compute() int {
    return add(5, 3) // Function call overhead
}

// After inlining (conceptually)
func compute() int {
    return 5 + 3 // Direct computation
}
```

### Inlining Costs and Benefits

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

// InliningAnalyzer analyzes function inlining behavior
type InliningAnalyzer struct {
    functions    map[string]*FunctionMetrics
    callGraph    map[string][]string
    inlineDecisions map[string]bool
}

// FunctionMetrics tracks function characteristics for inlining analysis
type FunctionMetrics struct {
    Name          string
    Size          int
    CallCount     int64
    InlineCount   int64
    Complexity    int
    HasLoops      bool
    HasCalls      bool
    Cost          float64
    Benefit       float64
    ShouldInline  bool
}

// NewInliningAnalyzer creates a new inlining analyzer
func NewInliningAnalyzer() *InliningAnalyzer {
    return &InliningAnalyzer{
        functions:       make(map[string]*FunctionMetrics),
        callGraph:       make(map[string][]string),
        inlineDecisions: make(map[string]bool),
    }
}

// AnalyzeFunction analyzes a function for inlining potential
func (ia *InliningAnalyzer) AnalyzeFunction(name string, body []byte) *FunctionMetrics {
    metrics := &FunctionMetrics{
        Name: name,
        Size: len(body),
    }
    
    // Analyze function complexity
    metrics.Complexity = ia.calculateComplexity(body)
    metrics.HasLoops = ia.hasLoops(body)
    metrics.HasCalls = ia.hasFunctionCalls(body)
    
    // Calculate inlining cost
    metrics.Cost = ia.calculateInliningCost(metrics)
    
    // Determine if function should be inlined
    metrics.ShouldInline = ia.shouldInline(metrics)
    
    ia.functions[name] = metrics
    return metrics
}

// calculateComplexity estimates function complexity
func (ia *InliningAnalyzer) calculateComplexity(body []byte) int {
    complexity := 1 // Base complexity
    
    // Count control flow statements
    controlFlowKeywords := []string{"if", "for", "switch", "select", "defer", "go"}
    for _, keyword := range controlFlowKeywords {
        complexity += countOccurrences(body, keyword)
    }
    
    return complexity
}

// hasLoops checks if function contains loops
func (ia *InliningAnalyzer) hasLoops(body []byte) bool {
    loopKeywords := []string{"for", "range"}
    for _, keyword := range loopKeywords {
        if countOccurrences(body, keyword) > 0 {
            return true
        }
    }
    return false
}

// hasFunctionCalls checks if function makes other function calls
func (ia *InliningAnalyzer) hasFunctionCalls(body []byte) bool {
    // Simplified detection - look for function call patterns
    return countOccurrences(body, "(") > 0
}

// calculateInliningCost calculates the cost of inlining a function
func (ia *InliningAnalyzer) calculateInliningCost(metrics *FunctionMetrics) float64 {
    cost := float64(metrics.Size) * 0.1 // Base size cost
    
    // Add complexity penalties
    cost += float64(metrics.Complexity) * 2.0
    
    // Penalty for loops (harder to optimize)
    if metrics.HasLoops {
        cost += 10.0
    }
    
    // Penalty for function calls (may prevent further optimization)
    if metrics.HasCalls {
        cost += 5.0
    }
    
    return cost
}

// shouldInline determines if a function should be inlined
func (ia *InliningAnalyzer) shouldInline(metrics *FunctionMetrics) bool {
    // Go compiler heuristics (simplified)
    if metrics.Size > 80 { // Budget threshold
        return false
    }
    
    if metrics.Complexity > 10 {
        return false
    }
    
    if metrics.HasLoops && metrics.Size > 40 {
        return false
    }
    
    return true
}

// countOccurrences counts keyword occurrences in code
func countOccurrences(data []byte, keyword string) int {
    count := 0
    keywordBytes := []byte(keyword)
    
    for i := 0; i <= len(data)-len(keywordBytes); i++ {
        if string(data[i:i+len(keywordBytes)]) == keyword {
            count++
        }
    }
    
    return count
}
```

## Go Compiler Inlining Rules

Understanding the Go compiler's inlining decisions helps optimize code for better performance.

### Inlining Budget System

```go
// InliningBudgetTracker tracks compiler inlining decisions
type InliningBudgetTracker struct {
    totalBudget    int
    usedBudget     int
    functionBudgets map[string]int
    decisions      []InliningDecision
}

// InliningDecision represents a compiler inlining decision
type InliningDecision struct {
    Function     string
    Caller       string
    Cost         int
    Inlined      bool
    Reason       string
    Timestamp    time.Time
}

// NewInliningBudgetTracker creates a new budget tracker
func NewInliningBudgetTracker(budget int) *InliningBudgetTracker {
    return &InliningBudgetTracker{
        totalBudget:     budget,
        functionBudgets: make(map[string]int),
        decisions:       make([]InliningDecision, 0),
    }
}

// TrackInliningDecision records an inlining decision
func (ibt *InliningBudgetTracker) TrackInliningDecision(decision InliningDecision) {
    ibt.decisions = append(ibt.decisions, decision)
    
    if decision.Inlined {
        ibt.usedBudget += decision.Cost
        ibt.functionBudgets[decision.Function] += decision.Cost
    }
}

// CanInline checks if a function can be inlined within budget
func (ibt *InliningBudgetTracker) CanInline(function string, cost int) bool {
    return ibt.usedBudget+cost <= ibt.totalBudget
}

// GetInliningReport generates a comprehensive inlining report
func (ibt *InliningBudgetTracker) GetInliningReport() InliningReport {
    report := InliningReport{
        TotalBudget:     ibt.totalBudget,
        UsedBudget:      ibt.usedBudget,
        RemainingBudget: ibt.totalBudget - ibt.usedBudget,
        InlinedFunctions: make(map[string]int),
        RejectedFunctions: make([]string, 0),
    }
    
    for _, decision := range ibt.decisions {
        if decision.Inlined {
            report.InlinedFunctions[decision.Function] = decision.Cost
        } else {
            report.RejectedFunctions = append(report.RejectedFunctions, decision.Function)
        }
    }
    
    return report
}

// InliningReport provides detailed inlining analysis
type InliningReport struct {
    TotalBudget       int
    UsedBudget        int
    RemainingBudget   int
    InlinedFunctions  map[string]int
    RejectedFunctions []string
    Recommendations   []string
}

// GenerateRecommendations provides optimization recommendations
func (ir *InliningReport) GenerateRecommendations() []string {
    recommendations := make([]string, 0)
    
    if ir.UsedBudget > int(float64(ir.TotalBudget)*0.9) {
        recommendations = append(recommendations, 
            "Consider reducing function complexity to improve inlining")
    }
    
    if len(ir.RejectedFunctions) > len(ir.InlinedFunctions) {
        recommendations = append(recommendations,
            "Many functions rejected for inlining - review function sizes")
    }
    
    if ir.RemainingBudget < ir.TotalBudget/4 {
        recommendations = append(recommendations,
            "Inlining budget nearly exhausted - prioritize hot path functions")
    }
    
    return recommendations
}
```

## Inlining Analysis Tools

Tools for analyzing and optimizing function inlining behavior.

### Compiler Flag Analysis

```go
// CompilerFlagAnalyzer analyzes inlining using compiler flags
type CompilerFlagAnalyzer struct {
    buildFlags   []string
    gcFlags      []string
    output       []string
    inlineInfo   map[string]InlineInfo
}

// InlineInfo contains inlining information for a function
type InlineInfo struct {
    Function    string
    File        string
    Line        int
    Inlined     bool
    Cost        int
    Reason      string
    CallSite    string
}

// NewCompilerFlagAnalyzer creates a new compiler flag analyzer
func NewCompilerFlagAnalyzer() *CompilerFlagAnalyzer {
    return &CompilerFlagAnalyzer{
        buildFlags: []string{},
        gcFlags:    []string{},
        output:     []string{},
        inlineInfo: make(map[string]InlineInfo),
    }
}

// EnableInliningDiagnostics enables compiler inlining diagnostics
func (cfa *CompilerFlagAnalyzer) EnableInliningDiagnostics() {
    cfa.gcFlags = append(cfa.gcFlags, "-m=2") // Verbose inlining info
    cfa.buildFlags = append(cfa.buildFlags, "-gcflags=-m=2")
}

// AnalyzeInlining analyzes inlining decisions for a package
func (cfa *CompilerFlagAnalyzer) AnalyzeInlining(packagePath string) error {
    // Execute build command with inlining flags
    cmd := fmt.Sprintf("go build -gcflags='-m=2' %s", packagePath)
    
    // In a real implementation, you would execute this command
    // and parse the output to extract inlining decisions
    
    // Simulate compiler output parsing
    cfa.parseCompilerOutput([]string{
        "can inline add",
        "inlining call to add",
        "cannot inline compute: function too complex",
    })
    
    return nil
}

// parseCompilerOutput parses compiler inlining output
func (cfa *CompilerFlagAnalyzer) parseCompilerOutput(output []string) {
    for _, line := range output {
        if info := cfa.parseInlineLine(line); info != nil {
            cfa.inlineInfo[info.Function] = *info
        }
    }
}

// parseInlineLine parses a single compiler output line
func (cfa *CompilerFlagAnalyzer) parseInlineLine(line string) *InlineInfo {
    // Simplified parsing - real implementation would be more robust
    if contains(line, "can inline") {
        function := extractFunctionName(line)
        return &InlineInfo{
            Function: function,
            Inlined:  true,
            Reason:   "meets inlining criteria",
        }
    }
    
    if contains(line, "cannot inline") {
        function := extractFunctionName(line)
        reason := extractReason(line)
        return &InlineInfo{
            Function: function,
            Inlined:  false,
            Reason:   reason,
        }
    }
    
    return nil
}

// Helper functions for parsing
func contains(s, substr string) bool {
    return len(s) > 0 && len(substr) > 0 // Simplified
}

func extractFunctionName(line string) string {
    // Simplified extraction
    return "function_name"
}

func extractReason(line string) string {
    // Simplified extraction
    return "unknown reason"
}
```

### Runtime Inlining Profiler

```go
// RuntimeInliningProfiler profiles inlining effectiveness at runtime
type RuntimeInliningProfiler struct {
    functionCalls    map[string]*CallMetrics
    inlineHits       map[string]int64
    inlineMisses     map[string]int64
    hotFunctions     []string
    sampling         bool
    sampleRate       float64
}

// CallMetrics tracks function call performance
type CallMetrics struct {
    Name           string
    CallCount      int64
    TotalDuration  time.Duration
    AvgDuration    time.Duration
    IsInlined      bool
    CallSites      map[string]int64
}

// NewRuntimeInliningProfiler creates a new runtime inlining profiler
func NewRuntimeInliningProfiler(sampleRate float64) *RuntimeInliningProfiler {
    return &RuntimeInliningProfiler{
        functionCalls: make(map[string]*CallMetrics),
        inlineHits:    make(map[string]int64),
        inlineMisses:  make(map[string]int64),
        hotFunctions:  make([]string, 0),
        sampling:      sampleRate < 1.0,
        sampleRate:    sampleRate,
    }
}

// StartProfiling begins inlining profiling
func (rip *RuntimeInliningProfiler) StartProfiling() {
    go rip.samplingLoop()
}

// samplingLoop continuously samples function calls
func (rip *RuntimeInliningProfiler) samplingLoop() {
    ticker := time.NewTicker(10 * time.Millisecond)
    defer ticker.Stop()
    
    for range ticker.C {
        if rip.sampling {
            rip.sampleInlining()
        }
    }
}

// sampleInlining samples current inlining effectiveness
func (rip *RuntimeInliningProfiler) sampleInlining() {
    // In a real implementation, this would use runtime introspection
    // to determine if functions are being called inline or not
    
    // Simulate sampling
    rip.recordInlineHit("fastFunction")
    rip.recordInlineMiss("complexFunction")
}

// recordInlineHit records a successful inline
func (rip *RuntimeInliningProfiler) recordInlineHit(function string) {
    rip.inlineHits[function]++
}

// recordInlineMiss records a failed inline (function call)
func (rip *RuntimeInliningProfiler) recordInlineMiss(function string) {
    rip.inlineMisses[function]++
}

// GetInliningEffectiveness calculates inlining effectiveness
func (rip *RuntimeInliningProfiler) GetInliningEffectiveness() map[string]float64 {
    effectiveness := make(map[string]float64)
    
    for function := range rip.inlineHits {
        hits := rip.inlineHits[function]
        misses := rip.inlineMisses[function]
        total := hits + misses
        
        if total > 0 {
            effectiveness[function] = float64(hits) / float64(total)
        }
    }
    
    return effectiveness
}

// IdentifyInliningOpportunities identifies functions that should be inlined
func (rip *RuntimeInliningProfiler) IdentifyInliningOpportunities() []InliningOpportunity {
    opportunities := make([]InliningOpportunity, 0)
    effectiveness := rip.GetInliningEffectiveness()
    
    for function, eff := range effectiveness {
        if eff < 0.5 && rip.inlineMisses[function] > 1000 {
            opportunities = append(opportunities, InliningOpportunity{
                Function:        function,
                CallCount:       rip.inlineMisses[function],
                CurrentEffectiveness: eff,
                PotentialGain:   rip.calculatePotentialGain(function),
                Recommendation:  rip.generateRecommendation(function, eff),
            })
        }
    }
    
    return opportunities
}

// InliningOpportunity represents an optimization opportunity
type InliningOpportunity struct {
    Function             string
    CallCount            int64
    CurrentEffectiveness float64
    PotentialGain        float64
    Recommendation       string
}

// calculatePotentialGain estimates performance gain from inlining
func (rip *RuntimeInliningProfiler) calculatePotentialGain(function string) float64 {
    callCount := rip.inlineMisses[function]
    
    // Simplified calculation - real implementation would be more sophisticated
    return float64(callCount) * 0.001 // Assume 1ms gain per 1000 calls
}

// generateRecommendation generates optimization recommendations
func (rip *RuntimeInliningProfiler) generateRecommendation(function string, effectiveness float64) string {
    if effectiveness < 0.1 {
        return fmt.Sprintf("Function %s never inlined - reduce complexity or size", function)
    } else if effectiveness < 0.5 {
        return fmt.Sprintf("Function %s partially inlined - review inlining barriers", function)
    }
    return "Function already well optimized"
}
```

## Optimization Strategies

Strategies for improving function inlining effectiveness.

### Function Size Optimization

```go
// FunctionSizeOptimizer helps optimize function sizes for better inlining
type FunctionSizeOptimizer struct {
    functions        map[string]*FunctionAnalysis
    optimizations    []OptimizationSuggestion
    maxInlineSize    int
}

// FunctionAnalysis contains detailed function analysis
type FunctionAnalysis struct {
    Name              string
    OriginalSize      int
    OptimizedSize     int
    Complexity        int
    Dependencies      []string
    InlineCandidate   bool
    Optimizations     []string
}

// OptimizationSuggestion suggests code improvements
type OptimizationSuggestion struct {
    Function        string
    Type           string
    Description    string
    ExpectedGain   int
    Difficulty     string
    CodeExample    string
}

// NewFunctionSizeOptimizer creates a new function size optimizer
func NewFunctionSizeOptimizer(maxSize int) *FunctionSizeOptimizer {
    return &FunctionSizeOptimizer{
        functions:     make(map[string]*FunctionAnalysis),
        optimizations: make([]OptimizationSuggestion, 0),
        maxInlineSize: maxSize,
    }
}

// AnalyzeFunction analyzes a function for size optimization
func (fso *FunctionSizeOptimizer) AnalyzeFunction(name string, code []byte) *FunctionAnalysis {
    analysis := &FunctionAnalysis{
        Name:         name,
        OriginalSize: len(code),
        Dependencies: make([]string, 0),
        Optimizations: make([]string, 0),
    }
    
    // Analyze function characteristics
    analysis.Complexity = fso.calculateComplexity(code)
    analysis.Dependencies = fso.findDependencies(code)
    
    // Determine if function is inline candidate
    analysis.InlineCandidate = analysis.OriginalSize <= fso.maxInlineSize
    
    // Generate optimization suggestions
    if !analysis.InlineCandidate {
        fso.generateOptimizations(analysis, code)
    }
    
    fso.functions[name] = analysis
    return analysis
}

// calculateComplexity calculates function complexity
func (fso *FunctionSizeOptimizer) calculateComplexity(code []byte) int {
    // Simplified complexity calculation
    return len(code) / 10 // Rough approximation
}

// findDependencies finds function dependencies
func (fso *FunctionSizeOptimizer) findDependencies(code []byte) []string {
    // Simplified dependency analysis
    return []string{} // Would extract actual function calls
}

// generateOptimizations generates optimization suggestions
func (fso *FunctionSizeOptimizer) generateOptimizations(analysis *FunctionAnalysis, code []byte) {
    if analysis.OriginalSize > fso.maxInlineSize*2 {
        fso.optimizations = append(fso.optimizations, OptimizationSuggestion{
            Function:     analysis.Name,
            Type:        "function-splitting",
            Description: "Split large function into smaller inline-able functions",
            ExpectedGain: analysis.OriginalSize / 2,
            Difficulty:  "medium",
            CodeExample: fso.generateSplittingExample(),
        })
    }
    
    if fso.hasRedundantCode(code) {
        fso.optimizations = append(fso.optimizations, OptimizationSuggestion{
            Function:     analysis.Name,
            Type:        "redundancy-removal",
            Description: "Remove redundant code and computations",
            ExpectedGain: 20,
            Difficulty:  "easy",
            CodeExample: fso.generateRedundancyExample(),
        })
    }
    
    if fso.hasComplexExpressions(code) {
        fso.optimizations = append(fso.optimizations, OptimizationSuggestion{
            Function:     analysis.Name,
            Type:        "expression-simplification",
            Description: "Simplify complex expressions",
            ExpectedGain: 15,
            Difficulty:  "easy",
            CodeExample: fso.generateSimplificationExample(),
        })
    }
}

// hasRedundantCode checks for redundant code patterns
func (fso *FunctionSizeOptimizer) hasRedundantCode(code []byte) bool {
    // Simplified check - would use more sophisticated analysis
    return len(code) > 100
}

// hasComplexExpressions checks for complex expressions
func (fso *FunctionSizeOptimizer) hasComplexExpressions(code []byte) bool {
    // Simplified check
    return contains(string(code), "&&") || contains(string(code), "||")
}

// generateSplittingExample generates function splitting example
func (fso *FunctionSizeOptimizer) generateSplittingExample() string {
    return `
// Before: Large function
func processData(data []Item) Result {
    // Validation logic (20 lines)
    // Processing logic (30 lines)
    // Output formatting (25 lines)
}

// After: Split into inline-able functions
func processData(data []Item) Result {
    if !validateData(data) { return nil }
    processed := processItems(data)
    return formatOutput(processed)
}

//go:inline
func validateData(data []Item) bool { /* 20 lines */ }

//go:inline  
func processItems(data []Item) []Item { /* 30 lines */ }

//go:inline
func formatOutput(data []Item) Result { /* 25 lines */ }
`
}

// generateRedundancyExample generates redundancy removal example
func (fso *FunctionSizeOptimizer) generateRedundancyExample() string {
    return `
// Before: Redundant computations
func calculate(x, y float64) float64 {
    a := math.Sqrt(x*x + y*y)
    b := math.Sqrt(x*x + y*y) // Redundant
    return a + b
}

// After: Remove redundancy
func calculate(x, y float64) float64 {
    dist := math.Sqrt(x*x + y*y)
    return 2 * dist
}
`
}

// generateSimplificationExample generates expression simplification example
func (fso *FunctionSizeOptimizer) generateSimplificationExample() string {
    return `
// Before: Complex expression
func isValid(x, y, z int) bool {
    return (x > 0 && y > 0 && z > 0) && 
           (x < 100 && y < 100 && z < 100) && 
           (x+y+z < 200)
}

// After: Simplified with early returns
func isValid(x, y, z int) bool {
    if x <= 0 || y <= 0 || z <= 0 { return false }
    if x >= 100 || y >= 100 || z >= 100 { return false }
    return x+y+z < 200
}
`
}

// GetOptimizationReport generates optimization report
func (fso *FunctionSizeOptimizer) GetOptimizationReport() OptimizationReport {
    totalFunctions := len(fso.functions)
    inlineCandidates := 0
    needsOptimization := 0
    
    for _, analysis := range fso.functions {
        if analysis.InlineCandidate {
            inlineCandidates++
        } else {
            needsOptimization++
        }
    }
    
    return OptimizationReport{
        TotalFunctions:      totalFunctions,
        InlineCandidates:    inlineCandidates,
        NeedsOptimization:   needsOptimization,
        Optimizations:       fso.optimizations,
        InlinePercentage:    float64(inlineCandidates) / float64(totalFunctions) * 100,
    }
}

// OptimizationReport provides function optimization analysis
type OptimizationReport struct {
    TotalFunctions      int
    InlineCandidates    int
    NeedsOptimization   int
    Optimizations       []OptimizationSuggestion
    InlinePercentage    float64
}
```

## Advanced Techniques

Advanced techniques for controlling and optimizing inlining behavior.

### Custom Inlining Directives

```go
// CustomInliningController provides fine-grained inlining control
type CustomInliningController struct {
    inlineDirectives  map[string]InlineDirective
    profiles         map[string]InlineProfile
    currentProfile   string
}

// InlineDirective specifies inlining behavior for a function
type InlineDirective struct {
    Function     string
    ForceInline  bool
    NoInline     bool
    Conditions   []InlineCondition
    Priority     int
}

// InlineCondition specifies conditional inlining
type InlineCondition struct {
    Type        string // "caller", "call_depth", "hot_path"
    Value       interface{}
    Operator    string // "equals", "greater_than", "less_than"
}

// InlineProfile contains inlining configuration for different scenarios
type InlineProfile struct {
    Name            string
    BudgetMultiplier float64
    MaxFunctionSize int
    AggressiveInline bool
    Directives      []InlineDirective
}

// NewCustomInliningController creates a new inlining controller
func NewCustomInliningController() *CustomInliningController {
    controller := &CustomInliningController{
        inlineDirectives: make(map[string]InlineDirective),
        profiles:        make(map[string]InlineProfile),
    }
    
    // Set up default profiles
    controller.setupDefaultProfiles()
    return controller
}

// setupDefaultProfiles creates default inlining profiles
func (cic *CustomInliningController) setupDefaultProfiles() {
    // Performance profile - aggressive inlining
    cic.profiles["performance"] = InlineProfile{
        Name:             "performance",
        BudgetMultiplier: 2.0,
        MaxFunctionSize:  120,
        AggressiveInline: true,
    }
    
    // Size profile - conservative inlining
    cic.profiles["size"] = InlineProfile{
        Name:             "size",
        BudgetMultiplier: 0.5,
        MaxFunctionSize:  60,
        AggressiveInline: false,
    }
    
    // Debug profile - minimal inlining
    cic.profiles["debug"] = InlineProfile{
        Name:             "debug",
        BudgetMultiplier: 0.1,
        MaxFunctionSize:  20,
        AggressiveInline: false,
    }
}

// SetInlineDirective sets a custom inlining directive
func (cic *CustomInliningController) SetInlineDirective(directive InlineDirective) {
    cic.inlineDirectives[directive.Function] = directive
}

// ForceInline forces a function to be inlined
func (cic *CustomInliningController) ForceInline(function string) {
    cic.SetInlineDirective(InlineDirective{
        Function:    function,
        ForceInline: true,
        Priority:    10,
    })
}

// PreventInline prevents a function from being inlined
func (cic *CustomInliningController) PreventInline(function string) {
    cic.SetInlineDirective(InlineDirective{
        Function: function,
        NoInline: true,
        Priority: 10,
    })
}

// SetConditionalInline sets conditional inlining rules
func (cic *CustomInliningController) SetConditionalInline(function string, conditions []InlineCondition) {
    cic.SetInlineDirective(InlineDirective{
        Function:   function,
        Conditions: conditions,
        Priority:   5,
    })
}

// UseProfile activates an inlining profile
func (cic *CustomInliningController) UseProfile(profileName string) error {
    if _, exists := cic.profiles[profileName]; !exists {
        return fmt.Errorf("profile %s not found", profileName)
    }
    
    cic.currentProfile = profileName
    return nil
}

// ShouldInline determines if a function should be inlined
func (cic *CustomInliningController) ShouldInline(function, caller string, callDepth int, isHotPath bool) bool {
    // Check explicit directives first
    if directive, exists := cic.inlineDirectives[function]; exists {
        if directive.NoInline {
            return false
        }
        
        if directive.ForceInline {
            return true
        }
        
        // Check conditions
        if cic.evaluateConditions(directive.Conditions, caller, callDepth, isHotPath) {
            return true
        }
    }
    
    // Use current profile rules
    profile := cic.profiles[cic.currentProfile]
    return cic.evaluateProfileRules(profile, function, caller, callDepth, isHotPath)
}

// evaluateConditions evaluates inlining conditions
func (cic *CustomInliningController) evaluateConditions(conditions []InlineCondition, caller string, callDepth int, isHotPath bool) bool {
    for _, condition := range conditions {
        if !cic.evaluateCondition(condition, caller, callDepth, isHotPath) {
            return false
        }
    }
    return true
}

// evaluateCondition evaluates a single inlining condition
func (cic *CustomInliningController) evaluateCondition(condition InlineCondition, caller string, callDepth int, isHotPath bool) bool {
    switch condition.Type {
    case "caller":
        return cic.compareValue(caller, condition.Value, condition.Operator)
    case "call_depth":
        return cic.compareValue(callDepth, condition.Value, condition.Operator)
    case "hot_path":
        return cic.compareValue(isHotPath, condition.Value, condition.Operator)
    }
    return false
}

// compareValue compares values using the specified operator
func (cic *CustomInliningController) compareValue(actual, expected interface{}, operator string) bool {
    switch operator {
    case "equals":
        return actual == expected
    case "greater_than":
        if a, ok := actual.(int); ok {
            if e, ok := expected.(int); ok {
                return a > e
            }
        }
    case "less_than":
        if a, ok := actual.(int); ok {
            if e, ok := expected.(int); ok {
                return a < e
            }
        }
    }
    return false
}

// evaluateProfileRules evaluates profile-based inlining rules
func (cic *CustomInliningController) evaluateProfileRules(profile InlineProfile, function, caller string, callDepth int, isHotPath bool) bool {
    // Simplified profile evaluation
    if profile.AggressiveInline && isHotPath {
        return true
    }
    
    if callDepth > 3 && !profile.AggressiveInline {
        return false
    }
    
    return true
}

// GenerateCompilerFlags generates compiler flags for current configuration
func (cic *CustomInliningController) GenerateCompilerFlags() []string {
    flags := make([]string, 0)
    
    profile := cic.profiles[cic.currentProfile]
    
    // Budget adjustment
    if profile.BudgetMultiplier != 1.0 {
        budget := int(80 * profile.BudgetMultiplier) // Base budget of 80
        flags = append(flags, fmt.Sprintf("-gcflags=-l=%d", budget))
    }
    
    // Function-specific directives
    for _, directive := range cic.inlineDirectives {
        if directive.ForceInline {
            flags = append(flags, fmt.Sprintf("-gcflags=-inline=%s", directive.Function))
        } else if directive.NoInline {
            flags = append(flags, fmt.Sprintf("-gcflags=-noinline=%s", directive.Function))
        }
    }
    
    return flags
}
```

## Performance Impact

Measuring and analyzing the performance impact of inlining optimizations.

### Inlining Performance Metrics

```go
// InliningPerformanceAnalyzer measures inlining performance impact
type InliningPerformanceAnalyzer struct {
    benchmarks       map[string]*BenchmarkResult
    baselineMetrics  *PerformanceMetrics
    optimizedMetrics *PerformanceMetrics
    comparisons      []PerformanceComparison
}

// BenchmarkResult contains benchmark results for inlining analysis
type BenchmarkResult struct {
    Name           string
    Iterations     int
    NsPerOp        int64
    AllocsPerOp    int64
    BytesPerOp     int64
    InlineVersion  bool
    Timestamp      time.Time
}

// PerformanceMetrics contains comprehensive performance metrics
type PerformanceMetrics struct {
    TotalCPUTime     time.Duration
    TotalAllocations int64
    TotalBytes       int64
    FunctionCalls    int64
    InlinedCalls     int64
    CacheHits        int64
    CacheMisses      int64
}

// PerformanceComparison compares inlined vs non-inlined performance
type PerformanceComparison struct {
    Function        string
    BaselineTime    time.Duration
    OptimizedTime   time.Duration
    Improvement     float64
    AllocReduction  int64
    Recommendation  string
}

// NewInliningPerformanceAnalyzer creates a new performance analyzer
func NewInliningPerformanceAnalyzer() *InliningPerformanceAnalyzer {
    return &InliningPerformanceAnalyzer{
        benchmarks:  make(map[string]*BenchmarkResult),
        comparisons: make([]PerformanceComparison, 0),
    }
}

// RunInliningBenchmark runs performance benchmarks for inlining analysis
func (ipa *InliningPerformanceAnalyzer) RunInliningBenchmark(functionName string, iterations int) *BenchmarkResult {
    // Benchmark without inlining
    baselineResult := ipa.runBenchmark(functionName, iterations, false)
    
    // Benchmark with inlining
    optimizedResult := ipa.runBenchmark(functionName, iterations, true)
    
    // Store results
    ipa.benchmarks[functionName+"_baseline"] = baselineResult
    ipa.benchmarks[functionName+"_optimized"] = optimizedResult
    
    // Create comparison
    comparison := ipa.createComparison(baselineResult, optimizedResult)
    ipa.comparisons = append(ipa.comparisons, comparison)
    
    return optimizedResult
}

// runBenchmark runs a single benchmark
func (ipa *InliningPerformanceAnalyzer) runBenchmark(functionName string, iterations int, inlined bool) *BenchmarkResult {
    start := time.Now()
    var totalAllocs int64
    var totalBytes int64
    
    // Get initial memory stats
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Run benchmark iterations
    for i := 0; i < iterations; i++ {
        // In a real implementation, this would call the actual function
        // For simulation, we'll just add some CPU work
        ipa.simulateWork(inlined)
    }
    
    // Get final memory stats
    runtime.ReadMemStats(&m2)
    totalAllocs = int64(m2.Mallocs - m1.Mallocs)
    totalBytes = int64(m2.TotalAlloc - m1.TotalAlloc)
    
    duration := time.Since(start)
    
    return &BenchmarkResult{
        Name:          functionName,
        Iterations:    iterations,
        NsPerOp:       duration.Nanoseconds() / int64(iterations),
        AllocsPerOp:   totalAllocs / int64(iterations),
        BytesPerOp:    totalBytes / int64(iterations),
        InlineVersion: inlined,
        Timestamp:     time.Now(),
    }
}

// simulateWork simulates function execution with different inlining behavior
func (ipa *InliningPerformanceAnalyzer) simulateWork(inlined bool) {
    if inlined {
        // Simulated inlined version - more efficient
        for i := 0; i < 100; i++ {
            _ = i * i
        }
    } else {
        // Simulated non-inlined version - with call overhead
        for i := 0; i < 100; i++ {
            ipa.helperFunction(i)
        }
    }
}

// helperFunction simulates a helper function call
func (ipa *InliningPerformanceAnalyzer) helperFunction(x int) int {
    return x * x
}

// createComparison creates a performance comparison
func (ipa *InliningPerformanceAnalyzer) createComparison(baseline, optimized *BenchmarkResult) PerformanceComparison {
    improvement := float64(baseline.NsPerOp-optimized.NsPerOp) / float64(baseline.NsPerOp) * 100
    allocReduction := baseline.AllocsPerOp - optimized.AllocsPerOp
    
    var recommendation string
    if improvement > 10 {
        recommendation = "Significant performance gain - prioritize inlining"
    } else if improvement > 5 {
        recommendation = "Moderate performance gain - consider inlining"
    } else if improvement > 0 {
        recommendation = "Minor performance gain - evaluate cost/benefit"
    } else {
        recommendation = "No performance gain - avoid inlining"
    }
    
    return PerformanceComparison{
        Function:       baseline.Name,
        BaselineTime:   time.Duration(baseline.NsPerOp),
        OptimizedTime:  time.Duration(optimized.NsPerOp),
        Improvement:    improvement,
        AllocReduction: allocReduction,
        Recommendation: recommendation,
    }
}

// AnalyzeInliningImpact analyzes overall inlining performance impact
func (ipa *InliningPerformanceAnalyzer) AnalyzeInliningImpact() InliningImpactReport {
    report := InliningImpactReport{
        TotalFunctions: len(ipa.comparisons),
        Comparisons:    ipa.comparisons,
    }
    
    var totalImprovement float64
    improvedFunctions := 0
    
    for _, comp := range ipa.comparisons {
        if comp.Improvement > 0 {
            improvedFunctions++
            totalImprovement += comp.Improvement
        }
    }
    
    if improvedFunctions > 0 {
        report.AverageImprovement = totalImprovement / float64(improvedFunctions)
    }
    
    report.ImprovedFunctions = improvedFunctions
    report.ImprovementRate = float64(improvedFunctions) / float64(len(ipa.comparisons)) * 100
    
    return report
}

// InliningImpactReport provides comprehensive inlining impact analysis
type InliningImpactReport struct {
    TotalFunctions      int
    ImprovedFunctions   int
    AverageImprovement  float64
    ImprovementRate     float64
    Comparisons         []PerformanceComparison
    Recommendations     []string
}

// GenerateRecommendations generates optimization recommendations
func (iir *InliningImpactReport) GenerateRecommendations() []string {
    recommendations := make([]string, 0)
    
    if iir.ImprovementRate > 80 {
        recommendations = append(recommendations, 
            "High inlining success rate - consider more aggressive inlining")
    } else if iir.ImprovementRate < 40 {
        recommendations = append(recommendations,
            "Low inlining success rate - review function characteristics")
    }
    
    if iir.AverageImprovement > 20 {
        recommendations = append(recommendations,
            "Significant performance gains observed - prioritize inlining optimization")
    }
    
    return recommendations
}
```

## Best Practices

Best practices for effective function inlining optimization.

### Inlining Guidelines

```go
// InliningBestPractices provides guidelines for effective inlining
type InliningBestPractices struct {
    guidelines     []Guideline
    antipatterns   []Antipattern
    checklist      []ChecklistItem
}

// Guideline represents an inlining best practice
type Guideline struct {
    Title       string
    Description string
    Example     string
    Priority    string
    Category    string
}

// Antipattern represents patterns to avoid for inlining
type Antipattern struct {
    Pattern     string
    Problem     string
    Solution    string
    Example     string
}

// ChecklistItem represents an inlining optimization checklist item
type ChecklistItem struct {
    Item        string
    Description string
    Automated   bool
    Priority    string
}

// NewInliningBestPractices creates best practices guide
func NewInliningBestPractices() *InliningBestPractices {
    bp := &InliningBestPractices{
        guidelines:   make([]Guideline, 0),
        antipatterns: make([]Antipattern, 0),
        checklist:    make([]ChecklistItem, 0),
    }
    
    bp.setupGuidelines()
    bp.setupAntipatterns()
    bp.setupChecklist()
    
    return bp
}

// setupGuidelines sets up inlining guidelines
func (ibp *InliningBestPractices) setupGuidelines() {
    ibp.guidelines = []Guideline{
        {
            Title:       "Keep Functions Small",
            Description: "Functions under 80 nodes are more likely to be inlined",
            Priority:    "High",
            Category:    "Function Design",
            Example: `
// Good: Small, focused function
func add(a, b int) int {
    return a + b
}

// Bad: Large function unlikely to be inlined
func processComplexData(data []Item) Result {
    // 200+ lines of code
}`,
        },
        {
            Title:       "Minimize Function Calls",
            Description: "Functions that call other functions are less likely to be inlined",
            Priority:    "Medium",
            Category:    "Function Design",
            Example: `
// Good: Self-contained function
func validate(x int) bool {
    return x > 0 && x < 100
}

// Less optimal: Function with calls
func validate(x int) bool {
    return isPositive(x) && isWithinRange(x, 100)
}`,
        },
        {
            Title:       "Avoid Complex Control Flow",
            Description: "Minimize loops and complex conditionals in inline candidates",
            Priority:    "Medium",
            Category:    "Control Flow",
            Example: `
// Good: Simple conditional
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// Less optimal: Complex control flow
func processWithLoop(items []int) int {
    sum := 0
    for _, item := range items {
        if item > 0 {
            sum += item
        }
    }
    return sum
}`,
        },
    }
}

// setupAntipatterns sets up common antipatterns
func (ibp *InliningBestPractices) setupAntipatterns() {
    ibp.antipatterns = []Antipattern{
        {
            Pattern:  "Large Function Bodies",
            Problem:  "Functions over compiler budget won't be inlined",
            Solution: "Split into smaller functions or extract to separate compilation unit",
            Example: `
// Antipattern: Too large for inlining
func massiveFunction() {
    // 300+ lines of code
}

// Solution: Split into smaller functions
func processData() {
    validate()
    transform()
    output()
}`,
        },
        {
            Pattern:  "Recursive Functions",
            Problem:  "Recursive functions are rarely inlined effectively",
            Solution: "Use iterative approach for hot path functions",
            Example: `
// Antipattern: Recursive function
func factorial(n int) int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n-1)
}

// Solution: Iterative version
func factorial(n int) int {
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result
}`,
        },
    }
}

// setupChecklist sets up optimization checklist
func (ibp *InliningBestPractices) setupChecklist() {
    ibp.checklist = []ChecklistItem{
        {
            Item:        "Profile hot functions",
            Description: "Identify frequently called functions for inlining priority",
            Automated:   true,
            Priority:    "High",
        },
        {
            Item:        "Check function sizes",
            Description: "Ensure functions are under inlining budget",
            Automated:   true,
            Priority:    "High",
        },
        {
            Item:        "Analyze call patterns",
            Description: "Review function call patterns and dependencies",
            Automated:   false,
            Priority:    "Medium",
        },
        {
            Item:        "Measure performance impact",
            Description: "Benchmark before and after inlining changes",
            Automated:   true,
            Priority:    "High",
        },
    }
}

// GetGuidelines returns inlining guidelines by category
func (ibp *InliningBestPractices) GetGuidelines(category string) []Guideline {
    if category == "" {
        return ibp.guidelines
    }
    
    filtered := make([]Guideline, 0)
    for _, guideline := range ibp.guidelines {
        if guideline.Category == category {
            filtered = append(filtered, guideline)
        }
    }
    return filtered
}

// GetAntipatterns returns common antipatterns
func (ibp *InliningBestPractices) GetAntipatterns() []Antipattern {
    return ibp.antipatterns
}

// GetChecklist returns optimization checklist
func (ibp *InliningBestPractices) GetChecklist() []ChecklistItem {
    return ibp.checklist
}

// ValidateFunction validates a function against best practices
func (ibp *InliningBestPractices) ValidateFunction(name string, code []byte) ValidationResult {
    result := ValidationResult{
        Function: name,
        Passed:   make([]string, 0),
        Failed:   make([]string, 0),
        Warnings: make([]string, 0),
    }
    
    // Check function size
    if len(code) <= 80 {
        result.Passed = append(result.Passed, "Function size within inlining budget")
    } else {
        result.Failed = append(result.Failed, "Function too large for inlining")
    }
    
    // Check for function calls
    if !ibp.hasFunctionCalls(code) {
        result.Passed = append(result.Passed, "No function calls detected")
    } else {
        result.Warnings = append(result.Warnings, "Function contains calls to other functions")
    }
    
    // Check for complex control flow
    if !ibp.hasComplexControlFlow(code) {
        result.Passed = append(result.Passed, "Simple control flow")
    } else {
        result.Warnings = append(result.Warnings, "Complex control flow may hinder inlining")
    }
    
    result.Score = ibp.calculateScore(result)
    return result
}

// ValidationResult contains function validation results
type ValidationResult struct {
    Function string
    Passed   []string
    Failed   []string
    Warnings []string
    Score    float64
}

// hasFunctionCalls checks for function calls in code
func (ibp *InliningBestPractices) hasFunctionCalls(code []byte) bool {
    // Simplified check - would use AST analysis in real implementation
    return contains(string(code), "(") && contains(string(code), ")")
}

// hasComplexControlFlow checks for complex control flow
func (ibp *InliningBestPractices) hasComplexControlFlow(code []byte) bool {
    codeStr := string(code)
    return contains(codeStr, "for") || contains(codeStr, "switch") || contains(codeStr, "select")
}

// calculateScore calculates validation score
func (ibp *InliningBestPractices) calculateScore(result ValidationResult) float64 {
    total := len(result.Passed) + len(result.Failed) + len(result.Warnings)
    if total == 0 {
        return 100.0
    }
    
    score := float64(len(result.Passed))*100 + float64(len(result.Warnings))*50
    return score / float64(total)
}
```

## Summary

Function inlining is a critical optimization technique in Go that can significantly improve performance by eliminating function call overhead and enabling further optimizations. Key takeaways:

1. **Understanding Compiler Behavior**: The Go compiler uses sophisticated heuristics to determine which functions to inline based on size, complexity, and budget constraints.

2. **Analysis Tools**: Use compiler flags, profiling tools, and custom analyzers to understand inlining decisions and identify optimization opportunities.

3. **Optimization Strategies**: Focus on keeping functions small, minimizing complexity, and reducing dependencies to improve inlining effectiveness.

4. **Performance Measurement**: Always benchmark the impact of inlining optimizations to ensure they provide real performance benefits.

5. **Best Practices**: Follow established guidelines for function design, avoid common antipatterns, and use systematic approaches to optimize inlining behavior.

Effective inlining optimization requires a balance between performance gains and code maintainability, with careful analysis and measurement guiding optimization decisions.
