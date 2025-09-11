# Escape Analysis

Escape analysis is a crucial compiler optimization in Go that determines whether variables can be allocated on the stack or must "escape" to the heap. Understanding and optimizing escape analysis is essential for building high-performance Go applications with minimal garbage collection pressure.

## Understanding Escape Analysis

Go's compiler performs escape analysis to determine:
- **Stack allocation** - Variables that remain within function scope
- **Heap allocation** - Variables that "escape" beyond function boundaries
- **Memory safety** - Ensuring pointers remain valid across function calls
- **Performance impact** - Stack allocation is faster and doesn't require GC
- **Optimization opportunities** - How to keep variables on the stack

### Escape Analysis Detector

```go
package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"
)

// EscapeAnalyzer provides comprehensive escape analysis insights
type EscapeAnalyzer struct {
    packagePath    string
    sourceFiles    map[string]*SourceFile
    escapeReports  []EscapeReport
    statistics     *EscapeStatistics
    buildFlags     []string
    mu             sync.RWMutex
}

type SourceFile struct {
    Path        string
    Content     string
    AST         *ast.File
    Functions   []*Function
    Variables   []*Variable
    EscapeData  map[string]*EscapeInfo
}

type Function struct {
    Name       string
    StartLine  int
    EndLine    int
    Parameters []*Variable
    Returns    []*Variable
    LocalVars  []*Variable
    Complexity int
}

type Variable struct {
    Name      string
    Type      string
    Line      int
    Escapes   bool
    Reason    string
    AllocSize int64
}

type EscapeInfo struct {
    Variable    string
    Function    string
    Line        int
    Escapes     bool
    Reason      string
    AllocSize   int64
    StackTrace  []string
}

type EscapeReport struct {
    FileName       string
    FunctionName   string
    TotalVars      int
    EscapingVars   int
    EscapeRate     float64
    HeapAllocs     int64
    StackAllocs    int64
    Optimizable    []OptimizationSuggestion
    CreatedAt      time.Time
}

type OptimizationSuggestion struct {
    Type        string
    Line        int
    Variable    string
    Current     string
    Suggested   string
    Impact      string
    Difficulty  string
}

type EscapeStatistics struct {
    TotalVariables     int
    EscapingVariables  int
    StackAllocations   int64
    HeapAllocations    int64
    TotalAllocSize     int64
    OptimizationCount  int
    PerformanceGain    float64
}

func NewEscapeAnalyzer(packagePath string) *EscapeAnalyzer {
    return &EscapeAnalyzer{
        packagePath: packagePath,
        sourceFiles: make(map[string]*SourceFile),
        statistics:  &EscapeStatistics{},
        buildFlags:  []string{"-gcflags", "-m -m"},
    }
}

func (ea *EscapeAnalyzer) AnalyzePackage() error {
    // Parse source files
    if err := ea.parseSourceFiles(); err != nil {
        return fmt.Errorf("failed to parse source files: %w", err)
    }
    
    // Run escape analysis
    if err := ea.runEscapeAnalysis(); err != nil {
        return fmt.Errorf("failed to run escape analysis: %w", err)
    }
    
    // Generate optimization suggestions
    ea.generateOptimizations()
    
    // Calculate statistics
    ea.calculateStatistics()
    
    return nil
}

func (ea *EscapeAnalyzer) parseSourceFiles() error {
    return filepath.Walk(ea.packagePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
            return nil
        }
        
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        
        fset := token.NewFileSet()
        astFile, err := parser.ParseFile(fset, path, content, parser.ParseComments)
        if err != nil {
            return err
        }
        
        sourceFile := &SourceFile{
            Path:       path,
            Content:    string(content),
            AST:        astFile,
            EscapeData: make(map[string]*EscapeInfo),
        }
        
        ea.parseAST(sourceFile, fset)
        
        ea.mu.Lock()
        ea.sourceFiles[path] = sourceFile
        ea.mu.Unlock()
        
        return nil
    })
}

func (ea *EscapeAnalyzer) parseAST(sourceFile *SourceFile, fset *token.FileSet) {
    ast.Inspect(sourceFile.AST, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.FuncDecl:
            if node.Body != nil {
                function := ea.parseFunctionDecl(node, fset)
                sourceFile.Functions = append(sourceFile.Functions, function)
            }
        case *ast.GenDecl:
            if node.Tok == token.VAR {
                for _, spec := range node.Specs {
                    if varSpec, ok := spec.(*ast.ValueSpec); ok {
                        variables := ea.parseVarSpec(varSpec, fset)
                        sourceFile.Variables = append(sourceFile.Variables, variables...)
                    }
                }
            }
        }
        return true
    })
}

func (ea *EscapeAnalyzer) parseFunctionDecl(funcDecl *ast.FuncDecl, fset *token.FileSet) *Function {
    pos := fset.Position(funcDecl.Pos())
    end := fset.Position(funcDecl.End())
    
    function := &Function{
        Name:      funcDecl.Name.Name,
        StartLine: pos.Line,
        EndLine:   end.Line,
    }
    
    // Parse parameters
    if funcDecl.Type.Params != nil {
        for _, field := range funcDecl.Type.Params.List {
            for _, name := range field.Names {
                param := &Variable{
                    Name: name.Name,
                    Type: ea.typeToString(field.Type),
                    Line: fset.Position(name.Pos()).Line,
                }
                function.Parameters = append(function.Parameters, param)
            }
        }
    }
    
    // Parse return values
    if funcDecl.Type.Results != nil {
        for _, field := range funcDecl.Type.Results.List {
            for _, name := range field.Names {
                ret := &Variable{
                    Name: name.Name,
                    Type: ea.typeToString(field.Type),
                    Line: fset.Position(name.Pos()).Line,
                }
                function.Returns = append(function.Returns, ret)
            }
        }
    }
    
    // Parse local variables
    ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.AssignStmt:
            for _, lhs := range node.Lhs {
                if ident, ok := lhs.(*ast.Ident); ok && ident.Obj != nil {
                    localVar := &Variable{
                        Name: ident.Name,
                        Line: fset.Position(ident.Pos()).Line,
                    }
                    function.LocalVars = append(function.LocalVars, localVar)
                }
            }
        case *ast.GenDecl:
            if node.Tok == token.VAR {
                for _, spec := range node.Specs {
                    if varSpec, ok := spec.(*ast.ValueSpec); ok {
                        variables := ea.parseVarSpec(varSpec, fset)
                        function.LocalVars = append(function.LocalVars, variables...)
                    }
                }
            }
        }
        return true
    })
    
    return function
}

func (ea *EscapeAnalyzer) parseVarSpec(varSpec *ast.ValueSpec, fset *token.FileSet) []*Variable {
    var variables []*Variable
    
    for i, name := range varSpec.Names {
        variable := &Variable{
            Name: name.Name,
            Line: fset.Position(name.Pos()).Line,
        }
        
        if varSpec.Type != nil {
            variable.Type = ea.typeToString(varSpec.Type)
        } else if i < len(varSpec.Values) {
            // Infer type from value
            variable.Type = ea.inferTypeFromExpr(varSpec.Values[i])
        }
        
        variables = append(variables, variable)
    }
    
    return variables
}

func (ea *EscapeAnalyzer) typeToString(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        return "*" + ea.typeToString(t.X)
    case *ast.ArrayType:
        return "[]" + ea.typeToString(t.Elt)
    case *ast.MapType:
        return "map[" + ea.typeToString(t.Key) + "]" + ea.typeToString(t.Value)
    case *ast.ChanType:
        return "chan " + ea.typeToString(t.Value)
    case *ast.InterfaceType:
        return "interface{}"
    case *ast.StructType:
        return "struct{}"
    case *ast.SelectorExpr:
        return ea.typeToString(t.X) + "." + t.Sel.Name
    default:
        return "unknown"
    }
}

func (ea *EscapeAnalyzer) inferTypeFromExpr(expr ast.Expr) string {
    switch e := expr.(type) {
    case *ast.BasicLit:
        switch e.Kind {
        case token.INT:
            return "int"
        case token.FLOAT:
            return "float64"
        case token.STRING:
            return "string"
        case token.CHAR:
            return "rune"
        }
    case *ast.CompositeLit:
        if e.Type != nil {
            return ea.typeToString(e.Type)
        }
    case *ast.CallExpr:
        if ident, ok := e.Fun.(*ast.Ident); ok {
            return ident.Name // Constructor call
        }
    }
    return "unknown"
}

func (ea *EscapeAnalyzer) runEscapeAnalysis() error {
    // Build with escape analysis flags
    cmd := exec.Command("go", "build")
    cmd.Args = append(cmd.Args, ea.buildFlags...)
    cmd.Args = append(cmd.Args, ea.packagePath)
    cmd.Dir = ea.packagePath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("go build failed: %w, output: %s", err, output)
    }
    
    // Parse escape analysis output
    ea.parseEscapeOutput(string(output))
    
    return nil
}

func (ea *EscapeAnalyzer) parseEscapeOutput(output string) {
    lines := strings.Split(output, "\n")
    
    // Regex patterns for escape analysis output
    escapePattern := regexp.MustCompile(`^(.+):(\d+):\d+: (.+) escapes to heap`)
    noEscapePattern := regexp.MustCompile(`^(.+):(\d+):\d+: (.+) does not escape`)
    allocPattern := regexp.MustCompile(`^(.+):(\d+):\d+: (.+) (\d+) bytes`)
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        
        if matches := escapePattern.FindStringSubmatch(line); matches != nil {
            fileName := matches[1]
            lineNum, _ := strconv.Atoi(matches[2])
            reason := matches[3]
            
            ea.recordEscapeInfo(fileName, lineNum, true, reason, 0)
        } else if matches := noEscapePattern.FindStringSubmatch(line); matches != nil {
            fileName := matches[1]
            lineNum, _ := strconv.Atoi(matches[2])
            reason := matches[3]
            
            ea.recordEscapeInfo(fileName, lineNum, false, reason, 0)
        } else if matches := allocPattern.FindStringSubmatch(line); matches != nil {
            fileName := matches[1]
            lineNum, _ := strconv.Atoi(matches[2])
            reason := matches[3]
            allocSize, _ := strconv.ParseInt(matches[4], 10, 64)
            
            ea.recordEscapeInfo(fileName, lineNum, true, reason, allocSize)
        }
    }
}

func (ea *EscapeAnalyzer) recordEscapeInfo(fileName string, lineNum int, escapes bool, reason string, allocSize int64) {
    ea.mu.Lock()
    defer ea.mu.Unlock()
    
    sourceFile, exists := ea.sourceFiles[fileName]
    if !exists {
        return
    }
    
    key := fmt.Sprintf("%s:%d", fileName, lineNum)
    
    escapeInfo := &EscapeInfo{
        Line:      lineNum,
        Escapes:   escapes,
        Reason:    reason,
        AllocSize: allocSize,
    }
    
    // Find the function and variable this line belongs to
    for _, function := range sourceFile.Functions {
        if lineNum >= function.StartLine && lineNum <= function.EndLine {
            escapeInfo.Function = function.Name
            
            // Find the specific variable
            for _, variable := range function.LocalVars {
                if variable.Line == lineNum {
                    escapeInfo.Variable = variable.Name
                    variable.Escapes = escapes
                    variable.Reason = reason
                    variable.AllocSize = allocSize
                    break
                }
            }
            break
        }
    }
    
    sourceFile.EscapeData[key] = escapeInfo
}

func (ea *EscapeAnalyzer) generateOptimizations() {
    ea.mu.Lock()
    defer ea.mu.Unlock()
    
    for filePath, sourceFile := range ea.sourceFiles {
        report := EscapeReport{
            FileName:  filePath,
            CreatedAt: time.Now(),
        }
        
        var totalVars, escapingVars int
        var heapAllocs, stackAllocs int64
        
        for _, function := range sourceFile.Functions {
            totalVars += len(function.LocalVars)
            
            for _, variable := range function.LocalVars {
                if variable.Escapes {
                    escapingVars++
                    heapAllocs += variable.AllocSize
                    
                    // Generate optimization suggestions
                    suggestions := ea.generateOptimizationSuggestions(variable, function)
                    report.Optimizable = append(report.Optimizable, suggestions...)
                } else {
                    stackAllocs += variable.AllocSize
                }
            }
        }
        
        if totalVars > 0 {
            report.TotalVars = totalVars
            report.EscapingVars = escapingVars
            report.EscapeRate = float64(escapingVars) / float64(totalVars) * 100
            report.HeapAllocs = heapAllocs
            report.StackAllocs = stackAllocs
            
            ea.escapeReports = append(ea.escapeReports, report)
        }
    }
}

func (ea *EscapeAnalyzer) generateOptimizationSuggestions(variable *Variable, function *Function) []OptimizationSuggestion {
    var suggestions []OptimizationSuggestion
    
    switch {
    case strings.Contains(variable.Reason, "returned"):
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:       "return_optimization",
            Line:       variable.Line,
            Variable:   variable.Name,
            Current:    "Variable escapes because it's returned",
            Suggested:  "Consider using value return instead of pointer, or use object pool",
            Impact:     "Reduces heap allocations",
            Difficulty: "Medium",
        })
        
    case strings.Contains(variable.Reason, "assigned to interface"):
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:       "interface_optimization",
            Line:       variable.Line,
            Variable:   variable.Name,
            Current:    "Variable escapes due to interface assignment",
            Suggested:  "Use concrete types where possible, or implement interface on value type",
            Impact:     "Eliminates heap allocation",
            Difficulty: "Hard",
        })
        
    case strings.Contains(variable.Reason, "too large"):
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:       "size_optimization",
            Line:       variable.Line,
            Variable:   variable.Name,
            Current:    "Variable too large for stack",
            Suggested:  "Reduce structure size or use smaller data types",
            Impact:     "May enable stack allocation",
            Difficulty: "Medium",
        })
        
    case strings.Contains(variable.Reason, "moved to heap"):
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:       "closure_optimization",
            Line:       variable.Line,
            Variable:   variable.Name,
            Current:    "Variable captured by closure",
            Suggested:  "Pass variables as parameters instead of capturing, or use value receivers",
            Impact:     "Reduces closure allocation overhead",
            Difficulty: "Easy",
        })
    }
    
    return suggestions
}

func (ea *EscapeAnalyzer) calculateStatistics() {
    ea.mu.Lock()
    defer ea.mu.Unlock()
    
    for _, report := range ea.escapeReports {
        ea.statistics.TotalVariables += report.TotalVars
        ea.statistics.EscapingVariables += report.EscapingVars
        ea.statistics.HeapAllocations += report.HeapAllocs
        ea.statistics.StackAllocations += report.StackAllocs
        ea.statistics.OptimizationCount += len(report.Optimizable)
    }
    
    ea.statistics.TotalAllocSize = ea.statistics.HeapAllocations + ea.statistics.StackAllocations
    
    if ea.statistics.TotalVariables > 0 {
        escapeRate := float64(ea.statistics.EscapingVariables) / float64(ea.statistics.TotalVariables)
        ea.statistics.PerformanceGain = (1.0 - escapeRate) * 100
    }
}

func (ea *EscapeAnalyzer) GetAnalysisReport() EscapeAnalysisReport {
    ea.mu.RLock()
    defer ea.mu.RUnlock()
    
    return EscapeAnalysisReport{
        PackagePath:     ea.packagePath,
        TotalFiles:      len(ea.sourceFiles),
        Statistics:      ea.statistics,
        FileReports:     ea.escapeReports,
        GeneratedAt:     time.Now(),
    }
}

type EscapeAnalysisReport struct {
    PackagePath string
    TotalFiles  int
    Statistics  *EscapeStatistics
    FileReports []EscapeReport
    GeneratedAt time.Time
}

func (ear EscapeAnalysisReport) String() string {
    result := fmt.Sprintf(`Escape Analysis Report for %s
Generated: %v
Files Analyzed: %d

=== STATISTICS ===
Total Variables: %d
Escaping Variables: %d (%.1f%%)
Heap Allocations: %d bytes
Stack Allocations: %d bytes
Total Allocation Size: %d bytes
Optimization Opportunities: %d
Estimated Performance Gain: %.1f%%`,
        ear.PackagePath,
        ear.GeneratedAt.Format(time.RFC3339),
        ear.TotalFiles,
        ear.Statistics.TotalVariables,
        ear.Statistics.EscapingVariables,
        float64(ear.Statistics.EscapingVariables)/float64(ear.Statistics.TotalVariables)*100,
        ear.Statistics.HeapAllocations,
        ear.Statistics.StackAllocations,
        ear.Statistics.TotalAllocSize,
        ear.Statistics.OptimizationCount,
        ear.Statistics.PerformanceGain)
    
    if len(ear.FileReports) > 0 {
        result += "\n\n=== FILE REPORTS ==="
        for _, report := range ear.FileReports {
            if report.EscapingVars > 0 {
                result += fmt.Sprintf("\n\n%s", report.String())
            }
        }
    }
    
    return result
}

func (er EscapeReport) String() string {
    result := fmt.Sprintf(`File: %s
Total Variables: %d
Escaping Variables: %d (%.1f%%)
Heap Allocations: %d bytes
Stack Allocations: %d bytes`,
        filepath.Base(er.FileName),
        er.TotalVars,
        er.EscapingVars, er.EscapeRate,
        er.HeapAllocs,
        er.StackAllocs)
    
    if len(er.Optimizable) > 0 {
        result += "\n\nOptimization Suggestions:"
        for i, suggestion := range er.Optimizable {
            if i >= 5 { // Show top 5 suggestions
                result += "\n  ..."
                break
            }
            result += fmt.Sprintf("\n  %d. Line %d (%s): %s",
                i+1, suggestion.Line, suggestion.Variable, suggestion.Suggested)
            result += fmt.Sprintf("\n     Impact: %s, Difficulty: %s",
                suggestion.Impact, suggestion.Difficulty)
        }
    }
    
    return result
}

func demonstrateEscapeAnalysis() {
    fmt.Println("=== ESCAPE ANALYSIS DEMONSTRATION ===")
    
    // Create sample code to analyze
    sampleCode := `package main

import "fmt"

// Example 1: Variable escapes because it's returned
func createLargeStruct() *LargeStruct {
    s := LargeStruct{data: make([]int, 1000)} // Escapes to heap
    return &s
}

// Example 2: Variable stays on stack
func processValue(s LargeStruct) int {
    local := s.data[0] // Stays on stack
    return local * 2
}

// Example 3: Interface assignment causes escape
func assignToInterface() interface{} {
    x := 42 // Escapes to heap due to interface{}
    return x
}

// Example 4: Closure capture causes escape
func createClosure() func() int {
    counter := 0 // Escapes due to closure capture
    return func() int {
        counter++
        return counter
    }
}

// Example 5: Large array escapes
func createLargeArray() {
    arr := [10000]int{} // Too large for stack, escapes to heap
    fmt.Println(len(arr))
}

type LargeStruct struct {
    data []int
    meta map[string]string
}

func main() {
    // These calls would be analyzed for escape behavior
    s := createLargeStruct()
    processValue(*s)
    assignToInterface()
    closure := createClosure()
    closure()
    createLargeArray()
}`
    
    // Write sample code to temporary file
    tmpDir := "/tmp/escape_analysis_demo"
    os.MkdirAll(tmpDir, 0755)
    
    tmpFile := filepath.Join(tmpDir, "main.go")
    err := os.WriteFile(tmpFile, []byte(sampleCode), 0644)
    if err != nil {
        fmt.Printf("Error writing sample code: %v\n", err)
        return
    }
    
    // Initialize escape analyzer
    analyzer := NewEscapeAnalyzer(tmpDir)
    
    // Run analysis
    if err := analyzer.AnalyzePackage(); err != nil {
        fmt.Printf("Analysis failed: %v\n", err)
        return
    }
    
    // Get and display report
    report := analyzer.GetAnalysisReport()
    fmt.Printf("\n%s\n", report)
    
    // Cleanup
    os.RemoveAll(tmpDir)
}
```

## Optimization Strategies

### 1. Stack Allocation Techniques

```go
// Prefer value receivers over pointer receivers for small types
type Point struct {
    X, Y float64
}

// Good: Value receiver keeps data on stack
func (p Point) Distance() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// Avoid: Pointer receiver may cause escape
func (p *Point) DistancePtr() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// Use arrays instead of slices for fixed-size data
func processFixedData() {
    // Good: Array stays on stack if small enough
    data := [100]int{}
    for i := range data {
        data[i] = i * i
    }
    
    // Avoid: Slice always allocated on heap
    // slice := make([]int, 100)
}
```

### 2. Interface Optimization

```go
// Use concrete types when possible
type Processor interface {
    Process(data []byte) error
}

type FileProcessor struct {
    path string
}

func (fp FileProcessor) Process(data []byte) error {
    // Implementation
    return nil
}

// Good: Direct type usage
func processDirectly(fp FileProcessor, data []byte) error {
    return fp.Process(data) // No interface escape
}

// Less optimal: Interface causes escape
func processViaInterface(p Processor, data []byte) error {
    return p.Process(data) // data may escape to heap
}
```

### 3. Return Value Optimization

```go
// Return values instead of pointers when possible
type Result struct {
    Value int
    Error string
}

// Good: Returns value, no heap allocation
func calculateValue(x, y int) Result {
    if y == 0 {
        return Result{Error: "division by zero"}
    }
    return Result{Value: x / y}
}

// Less optimal: Returns pointer, heap allocation
func calculateValuePtr(x, y int) *Result {
    result := &Result{} // Escapes to heap
    if y == 0 {
        result.Error = "division by zero"
    } else {
        result.Value = x / y
    }
    return result
}
```

### 4. Closure Optimization

```go
// Avoid capturing variables in closures
func createCounter() func() int {
    counter := 0 // This escapes due to closure capture
    return func() int {
        counter++
        return counter
    }
}

// Better: Pass state explicitly
type Counter struct {
    value int
}

func (c *Counter) Increment() int {
    c.value++
    return c.value
}

func createCounterStruct() *Counter {
    return &Counter{} // Only struct escapes, not captured variables
}
```

## Advanced Analysis Tools

### 1. Escape Analysis Profiler

```go
// Escape profiler for runtime analysis
type EscapeProfiler struct {
    samples    []EscapeSample
    heapStats  runtime.MemStats
    interval   time.Duration
    mu         sync.RWMutex
}

type EscapeSample struct {
    Timestamp   time.Time
    HeapObjects uint64
    HeapBytes   uint64
    StackBytes  uint64
    GCCycles    uint32
    AllocRate   float64 // bytes per second
}

func NewEscapeProfiler(interval time.Duration) *EscapeProfiler {
    return &EscapeProfiler{
        interval: interval,
    }
}

func (ep *EscapeProfiler) Start() {
    go ep.collectSamples()
}

func (ep *EscapeProfiler) collectSamples() {
    ticker := time.NewTicker(ep.interval)
    defer ticker.Stop()
    
    var lastStats runtime.MemStats
    runtime.ReadMemStats(&lastStats)
    lastTime := time.Now()
    
    for range ticker.C {
        var currentStats runtime.MemStats
        runtime.ReadMemStats(&currentStats)
        currentTime := time.Now()
        
        // Calculate allocation rate
        allocDiff := currentStats.TotalAlloc - lastStats.TotalAlloc
        timeDiff := currentTime.Sub(lastTime).Seconds()
        allocRate := float64(allocDiff) / timeDiff
        
        sample := EscapeSample{
            Timestamp:   currentTime,
            HeapObjects: currentStats.HeapObjects,
            HeapBytes:   currentStats.HeapAlloc,
            StackBytes:  currentStats.StackInuse,
            GCCycles:    currentStats.NumGC,
            AllocRate:   allocRate,
        }
        
        ep.mu.Lock()
        ep.samples = append(ep.samples, sample)
        
        // Keep only recent samples
        if len(ep.samples) > 1000 {
            ep.samples = ep.samples[len(ep.samples)-1000:]
        }
        ep.mu.Unlock()
        
        lastStats = currentStats
        lastTime = currentTime
    }
}

func (ep *EscapeProfiler) GetProfile() EscapeProfile {
    ep.mu.RLock()
    defer ep.mu.RUnlock()
    
    if len(ep.samples) == 0 {
        return EscapeProfile{}
    }
    
    var totalAllocRate float64
    var maxHeapBytes uint64
    var avgHeapObjects float64
    
    for _, sample := range ep.samples {
        totalAllocRate += sample.AllocRate
        if sample.HeapBytes > maxHeapBytes {
            maxHeapBytes = sample.HeapBytes
        }
        avgHeapObjects += float64(sample.HeapObjects)
    }
    
    avgAllocRate := totalAllocRate / float64(len(ep.samples))
    avgHeapObjects /= float64(len(ep.samples))
    
    return EscapeProfile{
        SampleCount:      len(ep.samples),
        AvgAllocRate:     avgAllocRate,
        MaxHeapBytes:     maxHeapBytes,
        AvgHeapObjects:   avgHeapObjects,
        EscapeEfficiency: ep.calculateEscapeEfficiency(),
    }
}

type EscapeProfile struct {
    SampleCount      int
    AvgAllocRate     float64
    MaxHeapBytes     uint64
    AvgHeapObjects   float64
    EscapeEfficiency float64
}

func (ep *EscapeProfiler) calculateEscapeEfficiency() float64 {
    // Calculate efficiency based on heap vs stack usage
    ep.mu.RLock()
    defer ep.mu.RUnlock()
    
    if len(ep.samples) == 0 {
        return 0
    }
    
    lastSample := ep.samples[len(ep.samples)-1]
    stackRatio := float64(lastSample.StackBytes) / float64(lastSample.HeapBytes+lastSample.StackBytes)
    
    return stackRatio * 100 // Percentage of stack allocation
}
```

### 2. Compiler Integration

```go
// Compiler flag optimizer
type CompilerOptimizer struct {
    flags        []string
    buildContext *build.Context
    targetArch   string
    targetOS     string
}

func NewCompilerOptimizer(targetArch, targetOS string) *CompilerOptimizer {
    return &CompilerOptimizer{
        buildContext: &build.Default,
        targetArch:   targetArch,
        targetOS:     targetOS,
    }
}

func (co *CompilerOptimizer) OptimizeForEscape() []string {
    flags := []string{
        "-gcflags", "-m -m",           // Verbose escape analysis
        "-gcflags", "-l=4",            // Aggressive inlining
        "-gcflags", "-N",              // Disable optimizations (for debugging)
        "-ldflags", "-s -w",           // Strip debug info
    }
    
    // Architecture-specific optimizations
    switch co.targetArch {
    case "amd64":
        flags = append(flags, "-gcflags", "-B") // Enable bounds check elimination
    case "arm64":
        flags = append(flags, "-gcflags", "-+") // Enable additional ARM optimizations
    }
    
    return flags
}

func (co *CompilerOptimizer) AnalyzeBuildOutput(output string) BuildAnalysis {
    analysis := BuildAnalysis{
        EscapeEvents:    co.countEscapeEvents(output),
        InlineEvents:    co.countInlineEvents(output),
        OptimizedFuncs:  co.countOptimizedFunctions(output),
        BuildTime:       co.extractBuildTime(output),
    }
    
    return analysis
}

type BuildAnalysis struct {
    EscapeEvents   int
    InlineEvents   int
    OptimizedFuncs int
    BuildTime      time.Duration
    Warnings       []string
}

func (co *CompilerOptimizer) countEscapeEvents(output string) int {
    escapePattern := regexp.MustCompile(`escapes to heap`)
    return len(escapePattern.FindAllString(output, -1))
}

func (co *CompilerOptimizer) countInlineEvents(output string) int {
    inlinePattern := regexp.MustCompile(`can inline`)
    return len(inlinePattern.FindAllString(output, -1))
}

func (co *CompilerOptimizer) countOptimizedFunctions(output string) int {
    // Count optimized function patterns
    return 0 // Placeholder
}

func (co *CompilerOptimizer) extractBuildTime(output string) time.Duration {
    // Extract build time from output
    return time.Second // Placeholder
}
```

## Best Practices

### 1. Design Patterns for Stack Allocation

```go
// Builder pattern with value semantics
type ConfigBuilder struct {
    config Config
}

func NewConfigBuilder() ConfigBuilder {
    return ConfigBuilder{
        config: Config{
            Timeout: time.Second * 30,
            Retries: 3,
        },
    }
}

func (cb ConfigBuilder) WithTimeout(timeout time.Duration) ConfigBuilder {
    cb.config.Timeout = timeout
    return cb // Return value, not pointer
}

func (cb ConfigBuilder) WithRetries(retries int) ConfigBuilder {
    cb.config.Retries = retries
    return cb
}

func (cb ConfigBuilder) Build() Config {
    return cb.config // Return value, not pointer
}

type Config struct {
    Timeout time.Duration
    Retries int
    URL     string
}
```

### 2. Memory Pool for Frequently Allocated Types

```go
// Object pool to reduce heap allocations
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
    return bp.pool.Get().([]byte)[:0] // Reset length, keep capacity
}

func (bp *BufferPool) Put(buf []byte) {
    if cap(buf) > 16*1024 { // Don't pool very large buffers
        return
    }
    bp.pool.Put(buf)
}

// Usage example
func processData(bp *BufferPool, input []byte) []byte {
    buf := bp.Get()
    defer bp.Put(buf)
    
    // Process input into buf
    buf = append(buf, input...)
    
    // Return copy to avoid escape
    result := make([]byte, len(buf))
    copy(result, buf)
    return result
}
```

### 3. Escape-Aware API Design

```go
// API designed to minimize escapes
type Processor struct {
    bufferPool *BufferPool
}

// Process data without allocating
func (p *Processor) Process(input []byte, output []byte) (int, error) {
    // Reuse provided buffer
    if cap(output) < len(input)*2 {
        return 0, fmt.Errorf("output buffer too small")
    }
    
    // Process in-place
    n := copy(output, input)
    return n, nil
}

// Alternative with callback to avoid return allocation
func (p *Processor) ProcessWithCallback(input []byte, callback func([]byte) error) error {
    buf := p.bufferPool.Get()
    defer p.bufferPool.Put(buf)
    
    // Process into temporary buffer
    buf = append(buf, input...)
    
    // Call callback with result
    return callback(buf)
}
```

## Next Steps

- Study [Inlining Optimization](inlining.md) techniques
- Learn [Build Optimization](build-optimization.md) strategies  
- Explore [Memory Layout](../memory/memory-layout.md) optimization
- Master [Stack vs Heap](../memory/stack-vs-heap.md) allocation

## Summary

Escape analysis optimization enables building high-performance Go applications by:

1. **Understanding allocation behavior** - Knowing when variables escape to heap
2. **Minimizing heap pressure** - Keeping allocations on stack when possible
3. **Reducing GC overhead** - Fewer heap allocations mean less garbage collection
4. **Optimizing API design** - Creating interfaces that promote stack allocation
5. **Measuring impact** - Quantifying the performance benefits of optimizations

Use these techniques to build memory-efficient Go applications with minimal garbage collection overhead.
