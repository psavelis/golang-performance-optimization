# Build Optimization

Comprehensive guide to optimizing Go builds for performance, including compiler flags, build strategies, and deployment optimization. This guide covers advanced build techniques for production applications.

## Table of Contents

- [Introduction](#introduction)
- [Compiler Flags](#compiler-flags)
- [Build Strategies](#build-strategies)
- [Link-Time Optimization](#link-time-optimization)
- [Binary Optimization](#binary-optimization)
- [Cross-Platform Builds](#cross-platform-builds)
- [Build Caching](#build-caching)
- [Performance Analysis](#performance-analysis)
- [Deployment Optimization](#deployment-optimization)
- [Best Practices](#best-practices)

## Introduction

Build optimization focuses on maximizing runtime performance while minimizing binary size and build time through strategic compiler configuration and build process optimization.

### Build Optimization Framework

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
    "time"
)

// BuildOptimizer provides comprehensive build optimization capabilities
type BuildOptimizer struct {
    config         BuildConfig
    metrics        *BuildMetrics
    flagAnalyzer   *CompilerFlagAnalyzer
    binaryAnalyzer *BinaryAnalyzer
    cacheManager   *BuildCacheManager
    mu             sync.RWMutex
}

// BuildConfig contains build optimization configuration
type BuildConfig struct {
    TargetOS         string
    TargetArch       string
    OptimizationLevel int
    EnableInlining   bool
    EnableEscapeAnalysis bool
    DebugMode        bool
    StaticLinking    bool
    StripSymbols     bool
    CompressUPX      bool
    LDFlags          []string
    GCFlags          []string
    BuildTags        []string
    CGOEnabled       bool
    ModulePath       string
    OutputPath       string
}

// BuildMetrics tracks build performance metrics
type BuildMetrics struct {
    BuildTime        time.Duration
    BinarySize       int64
    CompileTime      time.Duration
    LinkTime         time.Duration
    OptimizationTime time.Duration
    CacheHitRate     float64
    MemoryUsage      int64
    CPUUsage         float64
}

// CompilerFlagAnalyzer analyzes and optimizes compiler flags
type CompilerFlagAnalyzer struct {
    baseFlags    []string
    optFlags     map[string]string
    perfProfile  map[string]time.Duration
    flagImpact   map[string]BuildImpact
}

// BuildImpact measures the impact of compiler flags
type BuildImpact struct {
    BuildTimeChange  time.Duration
    BinarySizeChange int64
    RuntimePerfChange float64
    CompilationComplexity int
}

// BinaryAnalyzer analyzes compiled binary characteristics
type BinaryAnalyzer struct {
    symbolTable    map[string]SymbolInfo
    sectionSizes   map[string]int64
    dependencies   []string
    optimizations  []OptimizationApplied
}

// SymbolInfo contains information about binary symbols
type SymbolInfo struct {
    Name         string
    Size         int64
    Type         string
    Section      string
    Visibility   string
    Inlined      bool
    Optimized    bool
}

// OptimizationApplied tracks applied optimizations
type OptimizationApplied struct {
    Type        string
    Target      string
    Improvement float64
    SizeImpact  int64
}

// NewBuildOptimizer creates a new build optimizer
func NewBuildOptimizer(config BuildConfig) *BuildOptimizer {
    return &BuildOptimizer{
        config:         config,
        metrics:        &BuildMetrics{},
        flagAnalyzer:   NewCompilerFlagAnalyzer(),
        binaryAnalyzer: NewBinaryAnalyzer(),
        cacheManager:   NewBuildCacheManager(),
    }
}

// NewCompilerFlagAnalyzer creates a new compiler flag analyzer
func NewCompilerFlagAnalyzer() *CompilerFlagAnalyzer {
    return &CompilerFlagAnalyzer{
        baseFlags:   []string{"-trimpath", "-buildvcs=false"},
        optFlags:    make(map[string]string),
        perfProfile: make(map[string]time.Duration),
        flagImpact:  make(map[string]BuildImpact),
    }
}

// NewBinaryAnalyzer creates a new binary analyzer
func NewBinaryAnalyzer() *BinaryAnalyzer {
    return &BinaryAnalyzer{
        symbolTable:   make(map[string]SymbolInfo),
        sectionSizes:  make(map[string]int64),
        dependencies:  make([]string, 0),
        optimizations: make([]OptimizationApplied, 0),
    }
}

// OptimizeBuild performs comprehensive build optimization
func (bo *BuildOptimizer) OptimizeBuild(ctx context.Context) (*BuildResult, error) {
    bo.mu.Lock()
    defer bo.mu.Unlock()
    
    startTime := time.Now()
    defer func() {
        bo.metrics.BuildTime = time.Since(startTime)
    }()
    
    // Phase 1: Analyze and optimize compiler flags
    optimizedFlags, err := bo.optimizeCompilerFlags(ctx)
    if err != nil {
        return nil, fmt.Errorf("flag optimization failed: %w", err)
    }
    
    // Phase 2: Perform optimized build
    buildResult, err := bo.performOptimizedBuild(ctx, optimizedFlags)
    if err != nil {
        return nil, fmt.Errorf("optimized build failed: %w", err)
    }
    
    // Phase 3: Analyze binary and apply post-build optimizations
    if err := bo.analyzeBinary(buildResult.BinaryPath); err != nil {
        return nil, fmt.Errorf("binary analysis failed: %w", err)
    }
    
    // Phase 4: Apply post-build optimizations
    if err := bo.applyPostBuildOptimizations(buildResult); err != nil {
        return nil, fmt.Errorf("post-build optimization failed: %w", err)
    }
    
    // Update metrics
    bo.updateBuildMetrics(buildResult)
    
    return buildResult, nil
}

// BuildResult contains the results of an optimized build
type BuildResult struct {
    BinaryPath       string
    BuildTime        time.Duration
    BinarySize       int64
    OptimizationLevel int
    FlagsUsed        []string
    Optimizations    []OptimizationApplied
    Metrics          BuildMetrics
}

// optimizeCompilerFlags analyzes and optimizes compiler flags
func (bo *BuildOptimizer) optimizeCompilerFlags(ctx context.Context) ([]string, error) {
    flags := make([]string, 0)
    
    // Base optimization flags
    flags = append(flags, bo.flagAnalyzer.baseFlags...)
    
    // Optimization level flags
    switch bo.config.OptimizationLevel {
    case 0: // Debug build
        flags = append(flags, "-gcflags=-N -l") // Disable optimizations and inlining
    case 1: // Basic optimization
        flags = append(flags, "-ldflags=-s") // Strip symbol table
    case 2: // Standard optimization
        flags = append(flags, "-ldflags=-s -w") // Strip symbol table and DWARF
    case 3: // Aggressive optimization
        flags = append(flags, "-ldflags=-s -w -extldflags=-static") // Static linking
        if bo.config.EnableInlining {
            flags = append(flags, "-gcflags=-l=4") // Aggressive inlining
        }
    }
    
    // Target-specific flags
    flags = append(flags, fmt.Sprintf("GOOS=%s", bo.config.TargetOS))
    flags = append(flags, fmt.Sprintf("GOARCH=%s", bo.config.TargetArch))
    
    // CGO flags
    if !bo.config.CGOEnabled {
        flags = append(flags, "CGO_ENABLED=0")
    }
    
    // Build tags
    if len(bo.config.BuildTags) > 0 {
        flags = append(flags, "-tags", strings.Join(bo.config.BuildTags, ","))
    }
    
    // Custom LD flags
    if len(bo.config.LDFlags) > 0 {
        ldflags := strings.Join(bo.config.LDFlags, " ")
        flags = append(flags, fmt.Sprintf("-ldflags=%s", ldflags))
    }
    
    // Custom GC flags
    if len(bo.config.GCFlags) > 0 {
        gcflags := strings.Join(bo.config.GCFlags, " ")
        flags = append(flags, fmt.Sprintf("-gcflags=%s", gcflags))
    }
    
    return flags, nil
}

// performOptimizedBuild executes the build with optimized flags
func (bo *BuildOptimizer) performOptimizedBuild(ctx context.Context, flags []string) (*BuildResult, error) {
    startTime := time.Now()
    
    // Construct build command
    args := []string{"build"}
    args = append(args, flags...)
    args = append(args, "-o", bo.config.OutputPath)
    args = append(args, bo.config.ModulePath)
    
    cmd := exec.CommandContext(ctx, "go", args...)
    
    // Set environment variables
    env := os.Environ()
    for _, flag := range flags {
        if strings.Contains(flag, "=") {
            env = append(env, flag)
        }
    }
    cmd.Env = env
    
    // Execute build
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("build failed: %s", string(output))
    }
    
    buildTime := time.Since(startTime)
    
    // Get binary size
    stat, err := os.Stat(bo.config.OutputPath)
    if err != nil {
        return nil, fmt.Errorf("failed to stat binary: %w", err)
    }
    
    return &BuildResult{
        BinaryPath:        bo.config.OutputPath,
        BuildTime:         buildTime,
        BinarySize:        stat.Size(),
        OptimizationLevel: bo.config.OptimizationLevel,
        FlagsUsed:         flags,
        Metrics:           *bo.metrics,
    }, nil
}

// analyzeBinary analyzes the compiled binary
func (bo *BuildOptimizer) analyzeBinary(binaryPath string) error {
    // Analyze binary sections
    if err := bo.binaryAnalyzer.analyzeSections(binaryPath); err != nil {
        return fmt.Errorf("section analysis failed: %w", err)
    }
    
    // Analyze symbols
    if err := bo.binaryAnalyzer.analyzeSymbols(binaryPath); err != nil {
        return fmt.Errorf("symbol analysis failed: %w", err)
    }
    
    // Analyze dependencies
    if err := bo.binaryAnalyzer.analyzeDependencies(binaryPath); err != nil {
        return fmt.Errorf("dependency analysis failed: %w", err)
    }
    
    return nil
}

// analyzeSections analyzes binary sections
func (ba *BinaryAnalyzer) analyzeSections(binaryPath string) error {
    cmd := exec.Command("objdump", "-h", binaryPath)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("objdump failed: %w", err)
    }
    
    // Parse objdump output
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, ".text") ||
           strings.Contains(line, ".data") ||
           strings.Contains(line, ".rodata") ||
           strings.Contains(line, ".bss") {
            
            fields := strings.Fields(line)
            if len(fields) >= 3 {
                sectionName := fields[1]
                // Parse size (simplified)
                ba.sectionSizes[sectionName] = 0 // Placeholder
            }
        }
    }
    
    return nil
}

// analyzeSymbols analyzes binary symbols
func (ba *BinaryAnalyzer) analyzeSymbols(binaryPath string) error {
    cmd := exec.Command("nm", "-S", binaryPath)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("nm failed: %w", err)
    }
    
    // Parse nm output
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) >= 3 {
            symbolName := fields[len(fields)-1]
            ba.symbolTable[symbolName] = SymbolInfo{
                Name:       symbolName,
                Type:       fields[1],
                Visibility: "global", // Simplified
            }
        }
    }
    
    return nil
}

// analyzeDependencies analyzes binary dependencies
func (ba *BinaryAnalyzer) analyzeDependencies(binaryPath string) error {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "linux":
        cmd = exec.Command("ldd", binaryPath)
    case "darwin":
        cmd = exec.Command("otool", "-L", binaryPath)
    default:
        return fmt.Errorf("dependency analysis not supported on %s", runtime.GOOS)
    }
    
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("dependency analysis failed: %w", err)
    }
    
    // Parse dependency output
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.Contains(line, binaryPath) {
            ba.dependencies = append(ba.dependencies, line)
        }
    }
    
    return nil
}

// applyPostBuildOptimizations applies optimizations after build
func (bo *BuildOptimizer) applyPostBuildOptimizations(result *BuildResult) error {
    // Strip additional symbols if requested
    if bo.config.StripSymbols {
        if err := bo.stripSymbols(result.BinaryPath); err != nil {
            return fmt.Errorf("symbol stripping failed: %w", err)
        }
    }
    
    // Compress with UPX if requested
    if bo.config.CompressUPX {
        if err := bo.compressWithUPX(result.BinaryPath); err != nil {
            return fmt.Errorf("UPX compression failed: %w", err)
        }
    }
    
    return nil
}

// stripSymbols removes symbols from binary
func (bo *BuildOptimizer) stripSymbols(binaryPath string) error {
    cmd := exec.Command("strip", binaryPath)
    return cmd.Run()
}

// compressWithUPX compresses binary with UPX
func (bo *BuildOptimizer) compressWithUPX(binaryPath string) error {
    cmd := exec.Command("upx", "--best", binaryPath)
    return cmd.Run()
}

// updateBuildMetrics updates build metrics
func (bo *BuildOptimizer) updateBuildMetrics(result *BuildResult) {
    stat, err := os.Stat(result.BinaryPath)
    if err == nil {
        bo.metrics.BinarySize = stat.Size()
    }
    
    bo.metrics.BuildTime = result.BuildTime
    bo.metrics.CompileTime = result.BuildTime // Simplified
}
```

## Link-Time Optimization

Advanced link-time optimization techniques for Go binaries.

### Link-Time Optimizer

```go
// LinkTimeOptimizer performs link-time optimizations
type LinkTimeOptimizer struct {
    config     LTOConfig
    analyzer   *LinkAnalyzer
    optimizer  *SymbolOptimizer
    deadCodeEliminator *DeadCodeEliminator
}

// LTOConfig contains link-time optimization configuration
type LTOConfig struct {
    EnableDeadCodeElimination bool
    EnableSymbolMerging       bool
    EnableSectionMerging      bool
    OptimizeImports          bool
    InlineThreshold          int
    SymbolVisibility         string
}

// LinkAnalyzer analyzes linking patterns
type LinkAnalyzer struct {
    symbolReferences map[string][]string
    importGraph      map[string][]string
    callGraph        map[string][]string
    sizeAnalysis     map[string]int64
}

// SymbolOptimizer optimizes symbol usage
type SymbolOptimizer struct {
    symbolTable    map[string]Symbol
    mergeCanditates []SymbolMerge
    optimizations  []SymbolOptimization
}

// Symbol represents a binary symbol
type Symbol struct {
    Name         string
    Size         int64
    Type         SymbolType
    References   int
    Section      string
    Visibility   Visibility
    CanInline    bool
    CanMerge     bool
}

// SymbolType defines symbol types
type SymbolType int

const (
    FunctionSymbol SymbolType = iota
    VariableSymbol
    ConstantSymbol
    TypeSymbol
)

// Visibility defines symbol visibility
type Visibility int

const (
    Public Visibility = iota
    Private
    Internal
)

// SymbolMerge represents a symbol merge opportunity
type SymbolMerge struct {
    Symbols     []string
    MergedName  string
    SizeReduction int64
    Complexity  int
}

// SymbolOptimization represents an applied optimization
type SymbolOptimization struct {
    Type        OptimizationType
    Target      string
    Improvement float64
    SizeImpact  int64
}

// OptimizationType defines optimization types
type OptimizationType int

const (
    DeadCodeElimination OptimizationType = iota
    SymbolMerging
    SectionMerging
    ImportOptimization
    InlineExpansion
)

// DeadCodeEliminator removes unused code
type DeadCodeEliminator struct {
    reachabilityGraph map[string]bool
    eliminatedSymbols []string
    sizeReduction     int64
}

// NewLinkTimeOptimizer creates a new link-time optimizer
func NewLinkTimeOptimizer(config LTOConfig) *LinkTimeOptimizer {
    return &LinkTimeOptimizer{
        config:             config,
        analyzer:           NewLinkAnalyzer(),
        optimizer:          NewSymbolOptimizer(),
        deadCodeEliminator: NewDeadCodeEliminator(),
    }
}

// NewLinkAnalyzer creates a new link analyzer
func NewLinkAnalyzer() *LinkAnalyzer {
    return &LinkAnalyzer{
        symbolReferences: make(map[string][]string),
        importGraph:      make(map[string][]string),
        callGraph:        make(map[string][]string),
        sizeAnalysis:     make(map[string]int64),
    }
}

// NewSymbolOptimizer creates a new symbol optimizer
func NewSymbolOptimizer() *SymbolOptimizer {
    return &SymbolOptimizer{
        symbolTable:     make(map[string]Symbol),
        mergeCanditates: make([]SymbolMerge, 0),
        optimizations:   make([]SymbolOptimization, 0),
    }
}

// NewDeadCodeEliminator creates a new dead code eliminator
func NewDeadCodeEliminator() *DeadCodeEliminator {
    return &DeadCodeEliminator{
        reachabilityGraph: make(map[string]bool),
        eliminatedSymbols: make([]string, 0),
    }
}

// OptimizeLinks performs link-time optimizations
func (lto *LinkTimeOptimizer) OptimizeLinks(objectFiles []string) (*LTOResult, error) {
    result := &LTOResult{
        ObjectFiles:   objectFiles,
        Optimizations: make([]SymbolOptimization, 0),
    }
    
    // Phase 1: Analyze linking patterns
    if err := lto.analyzer.analyzeLinks(objectFiles); err != nil {
        return nil, fmt.Errorf("link analysis failed: %w", err)
    }
    
    // Phase 2: Dead code elimination
    if lto.config.EnableDeadCodeElimination {
        eliminated, err := lto.deadCodeEliminator.eliminate(lto.analyzer)
        if err != nil {
            return nil, fmt.Errorf("dead code elimination failed: %w", err)
        }
        result.Optimizations = append(result.Optimizations, eliminated...)
    }
    
    // Phase 3: Symbol optimization
    if lto.config.EnableSymbolMerging {
        optimized, err := lto.optimizer.optimizeSymbols(lto.analyzer)
        if err != nil {
            return nil, fmt.Errorf("symbol optimization failed: %w", err)
        }
        result.Optimizations = append(result.Optimizations, optimized...)
    }
    
    // Phase 4: Import optimization
    if lto.config.OptimizeImports {
        if err := lto.optimizeImports(); err != nil {
            return nil, fmt.Errorf("import optimization failed: %w", err)
        }
    }
    
    result.TotalSizeReduction = lto.calculateSizeReduction(result.Optimizations)
    
    return result, nil
}

// LTOResult contains link-time optimization results
type LTOResult struct {
    ObjectFiles         []string
    Optimizations       []SymbolOptimization
    TotalSizeReduction  int64
    OptimizationCount   int
}

// analyzeLinks analyzes linking patterns
func (la *LinkAnalyzer) analyzeLinks(objectFiles []string) error {
    for _, file := range objectFiles {
        if err := la.analyzeObjectFile(file); err != nil {
            return fmt.Errorf("failed to analyze %s: %w", file, err)
        }
    }
    
    // Build call graph
    la.buildCallGraph()
    
    return nil
}

// analyzeObjectFile analyzes a single object file
func (la *LinkAnalyzer) analyzeObjectFile(filename string) error {
    // Use nm to extract symbols
    cmd := exec.Command("nm", filename)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("nm failed for %s: %w", filename, err)
    }
    
    // Parse nm output
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) >= 3 {
            symbolName := fields[len(fields)-1]
            la.symbolReferences[symbolName] = append(la.symbolReferences[symbolName], filename)
        }
    }
    
    return nil
}

// buildCallGraph builds the function call graph
func (la *LinkAnalyzer) buildCallGraph() {
    // Simplified call graph building
    // In practice, this would analyze debug information or use Go's SSA
    for symbol := range la.symbolReferences {
        if strings.HasPrefix(symbol, "main.") {
            la.callGraph[symbol] = []string{} // Placeholder
        }
    }
}

// eliminate performs dead code elimination
func (dce *DeadCodeEliminator) eliminate(analyzer *LinkAnalyzer) ([]SymbolOptimization, error) {
    optimizations := make([]SymbolOptimization, 0)
    
    // Mark reachable symbols starting from entry points
    dce.markReachable("main.main", analyzer)
    dce.markReachable("main.init", analyzer)
    
    // Eliminate unreachable symbols
    for symbol := range analyzer.symbolReferences {
        if !dce.reachabilityGraph[symbol] {
            dce.eliminatedSymbols = append(dce.eliminatedSymbols, symbol)
            
            optimizations = append(optimizations, SymbolOptimization{
                Type:        DeadCodeElimination,
                Target:      symbol,
                Improvement: 100.0, // Complete elimination
                SizeImpact:  analyzer.sizeAnalysis[symbol],
            })
            
            dce.sizeReduction += analyzer.sizeAnalysis[symbol]
        }
    }
    
    return optimizations, nil
}

// markReachable marks symbols as reachable
func (dce *DeadCodeEliminator) markReachable(symbol string, analyzer *LinkAnalyzer) {
    if dce.reachabilityGraph[symbol] {
        return // Already marked
    }
    
    dce.reachabilityGraph[symbol] = true
    
    // Mark all called functions as reachable
    if callees, exists := analyzer.callGraph[symbol]; exists {
        for _, callee := range callees {
            dce.markReachable(callee, analyzer)
        }
    }
}

// optimizeSymbols performs symbol-level optimizations
func (so *SymbolOptimizer) optimizeSymbols(analyzer *LinkAnalyzer) ([]SymbolOptimization, error) {
    optimizations := make([]SymbolOptimization, 0)
    
    // Find merge candidates
    mergeCandidates := so.findMergeCandidates(analyzer)
    
    // Apply symbol merging
    for _, candidate := range mergeCandidates {
        opt := SymbolOptimization{
            Type:        SymbolMerging,
            Target:      strings.Join(candidate.Symbols, ","),
            Improvement: float64(candidate.SizeReduction) / float64(candidate.SizeReduction) * 100,
            SizeImpact:  candidate.SizeReduction,
        }
        optimizations = append(optimizations, opt)
    }
    
    return optimizations, nil
}

// findMergeCandidates identifies symbols that can be merged
func (so *SymbolOptimizer) findMergeCandidates(analyzer *LinkAnalyzer) []SymbolMerge {
    candidates := make([]SymbolMerge, 0)
    
    // Group similar symbols
    symbolGroups := make(map[string][]string)
    
    for symbol := range analyzer.symbolReferences {
        // Group by prefix (simplified heuristic)
        prefix := so.getSymbolPrefix(symbol)
        symbolGroups[prefix] = append(symbolGroups[prefix], symbol)
    }
    
    // Find merge opportunities
    for prefix, symbols := range symbolGroups {
        if len(symbols) > 1 && so.canMergeSymbols(symbols) {
            totalSize := int64(0)
            for _, symbol := range symbols {
                totalSize += analyzer.sizeAnalysis[symbol]
            }
            
            mergedSize := totalSize * 80 / 100 // Assume 20% reduction
            
            candidates = append(candidates, SymbolMerge{
                Symbols:       symbols,
                MergedName:    prefix + "_merged",
                SizeReduction: totalSize - mergedSize,
            })
        }
    }
    
    return candidates
}

// getSymbolPrefix extracts symbol prefix for grouping
func (so *SymbolOptimizer) getSymbolPrefix(symbol string) string {
    parts := strings.Split(symbol, ".")
    if len(parts) > 1 {
        return strings.Join(parts[:len(parts)-1], ".")
    }
    return symbol
}

// canMergeSymbols determines if symbols can be merged
func (so *SymbolOptimizer) canMergeSymbols(symbols []string) bool {
    // Simplified merge criteria
    for _, symbol := range symbols {
        if strings.Contains(symbol, "init") || 
           strings.Contains(symbol, "main") {
            return false
        }
    }
    return true
}

// optimizeImports optimizes import dependencies
func (lto *LinkTimeOptimizer) optimizeImports() error {
    // Simplified import optimization
    // In practice, this would analyze Go module dependencies
    return nil
}

// calculateSizeReduction calculates total size reduction
func (lto *LinkTimeOptimizer) calculateSizeReduction(optimizations []SymbolOptimization) int64 {
    total := int64(0)
    for _, opt := range optimizations {
        total += opt.SizeImpact
    }
    return total
}
```

## Binary Optimization

Post-build binary optimization techniques.

### Binary Size Analyzer

```go
// BinarySizeAnalyzer provides detailed binary size analysis
type BinarySizeAnalyzer struct {
    binaryPath    string
    sections      map[string]SectionInfo
    symbols       map[string]SymbolInfo
    totalSize     int64
    breakdown     SizeBreakdown
}

// SectionInfo contains information about binary sections
type SectionInfo struct {
    Name         string
    Size         int64
    Offset       int64
    Type         SectionType
    Permissions  string
    Alignment    int64
}

// SectionType defines binary section types
type SectionType int

const (
    TextSection SectionType = iota
    DataSection
    BSSSection
    RODataSection
    SymtabSection
    StrtabSection
)

// SizeBreakdown provides detailed size breakdown
type SizeBreakdown struct {
    CodeSize      int64
    DataSize      int64
    SymbolsSize   int64
    DebugSize     int64
    OtherSize     int64
    Overhead      int64
}

// NewBinarySizeAnalyzer creates a new binary size analyzer
func NewBinarySizeAnalyzer(binaryPath string) *BinarySizeAnalyzer {
    return &BinarySizeAnalyzer{
        binaryPath: binaryPath,
        sections:   make(map[string]SectionInfo),
        symbols:    make(map[string]SymbolInfo),
    }
}

// AnalyzeSize performs comprehensive size analysis
func (bsa *BinarySizeAnalyzer) AnalyzeSize() (*SizeAnalysisResult, error) {
    // Get binary file size
    stat, err := os.Stat(bsa.binaryPath)
    if err != nil {
        return nil, fmt.Errorf("failed to stat binary: %w", err)
    }
    bsa.totalSize = stat.Size()
    
    // Analyze sections
    if err := bsa.analyzeSections(); err != nil {
        return nil, fmt.Errorf("section analysis failed: %w", err)
    }
    
    // Analyze symbols
    if err := bsa.analyzeSymbols(); err != nil {
        return nil, fmt.Errorf("symbol analysis failed: %w", err)
    }
    
    // Calculate breakdown
    bsa.calculateBreakdown()
    
    // Generate optimization suggestions
    suggestions := bsa.generateOptimizationSuggestions()
    
    return &SizeAnalysisResult{
        TotalSize:     bsa.totalSize,
        Sections:      bsa.sections,
        Symbols:       bsa.symbols,
        Breakdown:     bsa.breakdown,
        Suggestions:   suggestions,
    }, nil
}

// SizeAnalysisResult contains size analysis results
type SizeAnalysisResult struct {
    TotalSize     int64
    Sections      map[string]SectionInfo
    Symbols       map[string]SymbolInfo
    Breakdown     SizeBreakdown
    Suggestions   []OptimizationSuggestion
}

// OptimizationSuggestion represents a size optimization suggestion
type OptimizationSuggestion struct {
    Type            SuggestionType
    Description     string
    PotentialSaving int64
    Difficulty      DifficultyLevel
    Implementation  string
}

// SuggestionType defines suggestion types
type SuggestionType int

const (
    StripSymbols SuggestionType = iota
    RemoveDebugInfo
    CompressBinary
    OptimizeImports
    EliminateDeadCode
    ReduceStaticData
)

// DifficultyLevel defines implementation difficulty
type DifficultyLevel int

const (
    Easy DifficultyLevel = iota
    Medium
    Hard
)

// analyzeSections analyzes binary sections
func (bsa *BinarySizeAnalyzer) analyzeSections() error {
    cmd := exec.Command("objdump", "-h", bsa.binaryPath)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("objdump failed: %w", err)
    }
    
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, "CONTENTS") {
            fields := strings.Fields(line)
            if len(fields) >= 6 {
                name := fields[1]
                sizeHex := fields[2]
                
                // Parse size from hex
                size := parseInt64(sizeHex, 16)
                
                sectionType := bsa.determineSectionType(name)
                
                bsa.sections[name] = SectionInfo{
                    Name: name,
                    Size: size,
                    Type: sectionType,
                }
            }
        }
    }
    
    return nil
}

// analyzeSymbols analyzes binary symbols
func (bsa *BinarySizeAnalyzer) analyzeSymbols() error {
    cmd := exec.Command("nm", "-S", "--size-sort", bsa.binaryPath)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("nm failed: %w", err)
    }
    
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) >= 4 {
            sizeHex := fields[1]
            symbolType := fields[2]
            name := fields[3]
            
            size := parseInt64(sizeHex, 16)
            
            bsa.symbols[name] = SymbolInfo{
                Name: name,
                Size: size,
                Type: symbolType,
            }
        }
    }
    
    return nil
}

// determineSectionType determines the type of a binary section
func (bsa *BinarySizeAnalyzer) determineSectionType(name string) SectionType {
    switch {
    case strings.Contains(name, ".text"):
        return TextSection
    case strings.Contains(name, ".data"):
        return DataSection
    case strings.Contains(name, ".bss"):
        return BSSSection
    case strings.Contains(name, ".rodata"):
        return RODataSection
    case strings.Contains(name, ".symtab"):
        return SymtabSection
    case strings.Contains(name, ".strtab"):
        return StrtabSection
    default:
        return TextSection
    }
}

// calculateBreakdown calculates size breakdown
func (bsa *BinarySizeAnalyzer) calculateBreakdown() {
    for _, section := range bsa.sections {
        switch section.Type {
        case TextSection:
            bsa.breakdown.CodeSize += section.Size
        case DataSection, BSSSection:
            bsa.breakdown.DataSize += section.Size
        case SymtabSection, StrtabSection:
            bsa.breakdown.SymbolsSize += section.Size
        default:
            bsa.breakdown.OtherSize += section.Size
        }
    }
    
    // Calculate overhead
    totalSectionSize := bsa.breakdown.CodeSize + bsa.breakdown.DataSize + 
                       bsa.breakdown.SymbolsSize + bsa.breakdown.OtherSize
    bsa.breakdown.Overhead = bsa.totalSize - totalSectionSize
}

// generateOptimizationSuggestions generates optimization suggestions
func (bsa *BinarySizeAnalyzer) generateOptimizationSuggestions() []OptimizationSuggestion {
    suggestions := make([]OptimizationSuggestion, 0)
    
    // Strip symbols suggestion
    if bsa.breakdown.SymbolsSize > 0 {
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:            StripSymbols,
            Description:     "Strip symbol table to reduce binary size",
            PotentialSaving: bsa.breakdown.SymbolsSize,
            Difficulty:      Easy,
            Implementation:  "go build -ldflags='-s -w'",
        })
    }
    
    // UPX compression suggestion
    if bsa.totalSize > 1024*1024 { // > 1MB
        estimatedSaving := bsa.totalSize * 30 / 100 // Estimate 30% compression
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:            CompressBinary,
            Description:     "Compress binary with UPX",
            PotentialSaving: estimatedSaving,
            Difficulty:      Easy,
            Implementation:  "upx --best binary",
        })
    }
    
    // Large symbol suggestions
    for name, symbol := range bsa.symbols {
        if symbol.Size > 10240 { // > 10KB
            suggestions = append(suggestions, OptimizationSuggestion{
                Type:            EliminateDeadCode,
                Description:     fmt.Sprintf("Optimize large symbol: %s", name),
                PotentialSaving: symbol.Size / 2, // Estimate 50% reduction
                Difficulty:      Medium,
                Implementation:  "Code review and optimization",
            })
        }
    }
    
    return suggestions
}

// parseInt64 parses int64 from string with base
func parseInt64(s string, base int) int64 {
    val, err := strconv.ParseInt(s, base, 64)
    if err != nil {
        return 0
    }
    return val
}
```

## Build Caching

Advanced build caching strategies for faster builds.

### Build Cache Manager

```go
// BuildCacheManager manages build caching for faster builds
type BuildCacheManager struct {
    cacheDir      string
    hashStrategy  HashStrategy
    cache         map[string]CacheEntry
    statistics    CacheStatistics
    config        CacheConfig
    mu            sync.RWMutex
}

// CacheEntry represents a cached build artifact
type CacheEntry struct {
    Hash         string
    Path         string
    Size         int64
    BuildTime    time.Duration
    CreatedAt    time.Time
    AccessCount  int64
    LastAccessed time.Time
    Dependencies []string
}

// CacheStatistics tracks cache performance
type CacheStatistics struct {
    HitCount        int64
    MissCount       int64
    TotalRequests   int64
    CacheSize       int64
    SpaceSaved      int64
    TimeSaved       time.Duration
}

// CacheConfig contains cache configuration
type CacheConfig struct {
    MaxSize       int64
    MaxAge        time.Duration
    HashAlgorithm string
    Compression   bool
    EvictionPolicy EvictionPolicy
}

// HashStrategy defines how build inputs are hashed
type HashStrategy interface {
    HashBuildInputs(inputs BuildInputs) (string, error)
    HashDependencies(deps []string) (string, error)
}

// BuildInputs contains all inputs that affect build output
type BuildInputs struct {
    SourceFiles   map[string]string // file path -> content hash
    BuildFlags    []string
    Environment   map[string]string
    GoVersion     string
    Dependencies  []Dependency
}

// Dependency represents a build dependency
type Dependency struct {
    Module  string
    Version string
    Hash    string
}

// EvictionPolicy defines cache eviction strategies
type EvictionPolicy int

const (
    LRUEviction EvictionPolicy = iota
    LFUEviction
    TTLEviction
    SizeBasedEviction
)

// SHA256HashStrategy implements SHA256-based hashing
type SHA256HashStrategy struct{}

// NewBuildCacheManager creates a new build cache manager
func NewBuildCacheManager(cacheDir string, config CacheConfig) *BuildCacheManager {
    return &BuildCacheManager{
        cacheDir:     cacheDir,
        hashStrategy: &SHA256HashStrategy{},
        cache:        make(map[string]CacheEntry),
        config:       config,
    }
}

// HashBuildInputs hashes build inputs using SHA256
func (hs *SHA256HashStrategy) HashBuildInputs(inputs BuildInputs) (string, error) {
    hasher := sha256.New()
    
    // Hash source files
    for path, contentHash := range inputs.SourceFiles {
        hasher.Write([]byte(path))
        hasher.Write([]byte(contentHash))
    }
    
    // Hash build flags
    for _, flag := range inputs.BuildFlags {
        hasher.Write([]byte(flag))
    }
    
    // Hash environment
    for key, value := range inputs.Environment {
        hasher.Write([]byte(key))
        hasher.Write([]byte(value))
    }
    
    // Hash Go version
    hasher.Write([]byte(inputs.GoVersion))
    
    // Hash dependencies
    for _, dep := range inputs.Dependencies {
        hasher.Write([]byte(dep.Module))
        hasher.Write([]byte(dep.Version))
        hasher.Write([]byte(dep.Hash))
    }
    
    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// HashDependencies hashes dependency list
func (hs *SHA256HashStrategy) HashDependencies(deps []string) (string, error) {
    hasher := sha256.New()
    
    for _, dep := range deps {
        hasher.Write([]byte(dep))
    }
    
    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// GetCachedBuild retrieves a cached build if available
func (bcm *BuildCacheManager) GetCachedBuild(inputs BuildInputs) (*CacheEntry, bool) {
    bcm.mu.RLock()
    defer bcm.mu.RUnlock()
    
    hash, err := bcm.hashStrategy.HashBuildInputs(inputs)
    if err != nil {
        return nil, false
    }
    
    atomic.AddInt64(&bcm.statistics.TotalRequests, 1)
    
    if entry, exists := bcm.cache[hash]; exists {
        // Check if entry is still valid
        if bcm.isValidEntry(entry) {
            // Update access statistics
            entry.AccessCount++
            entry.LastAccessed = time.Now()
            bcm.cache[hash] = entry
            
            atomic.AddInt64(&bcm.statistics.HitCount, 1)
            atomic.AddInt64(&bcm.statistics.TimeSaved, int64(entry.BuildTime))
            
            return &entry, true
        } else {
            // Entry expired, remove it
            delete(bcm.cache, hash)
        }
    }
    
    atomic.AddInt64(&bcm.statistics.MissCount, 1)
    return nil, false
}

// StoreBuild stores a build result in cache
func (bcm *BuildCacheManager) StoreBuild(inputs BuildInputs, artifactPath string, buildTime time.Duration) error {
    bcm.mu.Lock()
    defer bcm.mu.Unlock()
    
    hash, err := bcm.hashStrategy.HashBuildInputs(inputs)
    if err != nil {
        return fmt.Errorf("failed to hash inputs: %w", err)
    }
    
    // Get artifact size
    stat, err := os.Stat(artifactPath)
    if err != nil {
        return fmt.Errorf("failed to stat artifact: %w", err)
    }
    
    // Copy artifact to cache
    cachedPath := filepath.Join(bcm.cacheDir, hash)
    if err := bcm.copyFile(artifactPath, cachedPath); err != nil {
        return fmt.Errorf("failed to copy artifact: %w", err)
    }
    
    // Create cache entry
    entry := CacheEntry{
        Hash:         hash,
        Path:         cachedPath,
        Size:         stat.Size(),
        BuildTime:    buildTime,
        CreatedAt:    time.Now(),
        LastAccessed: time.Now(),
        Dependencies: bcm.extractDependencies(inputs),
    }
    
    // Check cache size limits
    if err := bcm.evictIfNecessary(entry.Size); err != nil {
        return fmt.Errorf("cache eviction failed: %w", err)
    }
    
    bcm.cache[hash] = entry
    atomic.AddInt64(&bcm.statistics.CacheSize, entry.Size)
    
    return nil
}

// isValidEntry checks if cache entry is still valid
func (bcm *BuildCacheManager) isValidEntry(entry CacheEntry) bool {
    // Check age
    if bcm.config.MaxAge > 0 && time.Since(entry.CreatedAt) > bcm.config.MaxAge {
        return false
    }
    
    // Check if file still exists
    if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
        return false
    }
    
    return true
}

// evictIfNecessary evicts cache entries if necessary
func (bcm *BuildCacheManager) evictIfNecessary(newEntrySize int64) error {
    currentSize := atomic.LoadInt64(&bcm.statistics.CacheSize)
    
    if currentSize+newEntrySize <= bcm.config.MaxSize {
        return nil
    }
    
    switch bcm.config.EvictionPolicy {
    case LRUEviction:
        return bcm.evictLRU(newEntrySize)
    case LFUEviction:
        return bcm.evictLFU(newEntrySize)
    case TTLEviction:
        return bcm.evictTTL(newEntrySize)
    case SizeBasedEviction:
        return bcm.evictBySize(newEntrySize)
    default:
        return bcm.evictLRU(newEntrySize)
    }
}

// evictLRU evicts least recently used entries
func (bcm *BuildCacheManager) evictLRU(targetSize int64) error {
    entries := make([]CacheEntry, 0, len(bcm.cache))
    for _, entry := range bcm.cache {
        entries = append(entries, entry)
    }
    
    // Sort by last accessed time (oldest first)
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].LastAccessed.Before(entries[j].LastAccessed)
    })
    
    freedSize := int64(0)
    for _, entry := range entries {
        if freedSize >= targetSize {
            break
        }
        
        if err := bcm.removeEntry(entry.Hash); err != nil {
            return fmt.Errorf("failed to remove entry %s: %w", entry.Hash, err)
        }
        freedSize += entry.Size
    }
    
    return nil
}

// evictLFU evicts least frequently used entries
func (bcm *BuildCacheManager) evictLFU(targetSize int64) error {
    entries := make([]CacheEntry, 0, len(bcm.cache))
    for _, entry := range bcm.cache {
        entries = append(entries, entry)
    }
    
    // Sort by access count (lowest first)
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].AccessCount < entries[j].AccessCount
    })
    
    freedSize := int64(0)
    for _, entry := range entries {
        if freedSize >= targetSize {
            break
        }
        
        if err := bcm.removeEntry(entry.Hash); err != nil {
            return fmt.Errorf("failed to remove entry %s: %w", entry.Hash, err)
        }
        freedSize += entry.Size
    }
    
    return nil
}

// evictTTL evicts expired entries
func (bcm *BuildCacheManager) evictTTL(targetSize int64) error {
    now := time.Now()
    freedSize := int64(0)
    
    for hash, entry := range bcm.cache {
        if now.Sub(entry.CreatedAt) > bcm.config.MaxAge {
            if err := bcm.removeEntry(hash); err != nil {
                return fmt.Errorf("failed to remove expired entry %s: %w", hash, err)
            }
            freedSize += entry.Size
            
            if freedSize >= targetSize {
                break
            }
        }
    }
    
    return nil
}

// evictBySize evicts largest entries first
func (bcm *BuildCacheManager) evictBySize(targetSize int64) error {
    entries := make([]CacheEntry, 0, len(bcm.cache))
    for _, entry := range bcm.cache {
        entries = append(entries, entry)
    }
    
    // Sort by size (largest first)
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Size > entries[j].Size
    })
    
    freedSize := int64(0)
    for _, entry := range entries {
        if freedSize >= targetSize {
            break
        }
        
        if err := bcm.removeEntry(entry.Hash); err != nil {
            return fmt.Errorf("failed to remove entry %s: %w", entry.Hash, err)
        }
        freedSize += entry.Size
    }
    
    return nil
}

// removeEntry removes a cache entry
func (bcm *BuildCacheManager) removeEntry(hash string) error {
    entry, exists := bcm.cache[hash]
    if !exists {
        return nil
    }
    
    // Remove file
    if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove cached file: %w", err)
    }
    
    // Remove from cache
    delete(bcm.cache, hash)
    atomic.AddInt64(&bcm.statistics.CacheSize, -entry.Size)
    
    return nil
}

// copyFile copies a file from src to dst
func (bcm *BuildCacheManager) copyFile(src, dst string) error {
    // Create cache directory if needed
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return fmt.Errorf("failed to create cache directory: %w", err)
    }
    
    srcFile, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("failed to open source file: %w", err)
    }
    defer srcFile.Close()
    
    dstFile, err := os.Create(dst)
    if err != nil {
        return fmt.Errorf("failed to create destination file: %w", err)
    }
    defer dstFile.Close()
    
    _, err = io.Copy(dstFile, srcFile)
    return err
}

// extractDependencies extracts dependency list from inputs
func (bcm *BuildCacheManager) extractDependencies(inputs BuildInputs) []string {
    deps := make([]string, len(inputs.Dependencies))
    for i, dep := range inputs.Dependencies {
        deps[i] = fmt.Sprintf("%s@%s", dep.Module, dep.Version)
    }
    return deps
}

// GetStatistics returns cache statistics
func (bcm *BuildCacheManager) GetStatistics() CacheStatistics {
    bcm.mu.RLock()
    defer bcm.mu.RUnlock()
    return bcm.statistics
}

// CleanupExpired removes expired cache entries
func (bcm *BuildCacheManager) CleanupExpired() error {
    bcm.mu.Lock()
    defer bcm.mu.Unlock()
    
    now := time.Now()
    for hash, entry := range bcm.cache {
        if now.Sub(entry.CreatedAt) > bcm.config.MaxAge {
            if err := bcm.removeEntry(hash); err != nil {
                return fmt.Errorf("failed to remove expired entry %s: %w", hash, err)
            }
        }
    }
    
    return nil
}
```

## Summary

Build optimization provides comprehensive strategies for maximizing Go application performance:

1. **Compiler Flags**: Strategic use of optimization flags and build parameters
2. **Link-Time Optimization**: Advanced linking strategies for performance and size
3. **Binary Optimization**: Post-build optimization techniques
4. **Build Caching**: Intelligent caching for faster development cycles
5. **Cross-Platform Builds**: Optimized builds for multiple target platforms
6. **Performance Analysis**: Continuous monitoring and optimization of build processes

These techniques enable production-ready Go applications with optimal performance characteristics and efficient development workflows.
