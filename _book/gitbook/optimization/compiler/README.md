# Compiler Optimization

Master Go compiler optimizations, build flags, and compilation techniques to maximize runtime performance through intelligent code generation.

## Go Compiler Architecture

### Compilation Pipeline

The Go compiler performs multiple optimization passes:

```go
// Source code → SSA IR → Machine code
// 
// 1. Parsing and type checking
// 2. SSA (Static Single Assignment) generation
// 3. Optimization passes
// 4. Code generation
// 5. Linking

// Example: Understanding compiler optimizations
package main

import "fmt"

// This function will be inlined by the compiler
func add(a, b int) int {
    return a + b
}

// This function demonstrates escape analysis
func createSlice() []int {
    // Compiler determines if this escapes to heap
    s := make([]int, 10)
    return s
}

func main() {
    // Compiler optimizes this call through inlining
    result := add(5, 3)
    fmt.Println(result)
    
    // Escape analysis determines allocation location
    data := createSlice()
    fmt.Println(len(data))
}
```

### SSA (Static Single Assignment) Form

Understanding SSA helps optimize code for the compiler:

```go
// Original code
func example(x, y int) int {
    x = x + 1
    if x > 10 {
        x = x * 2
    }
    return x + y
}

// SSA form (conceptual):
// x1 = x0 + 1
// if x1 > 10 goto B2 else B3
// B2: x2 = x1 * 2; goto B4
// B3: x2 = x1; goto B4  
// B4: result = x2 + y0; return result
```

## Build Flags and Optimization Levels

### Essential Build Flags

```bash
# Basic optimization (default)
go build -o myapp main.go

# Disable optimizations (debugging)
go build -gcflags="-N -l" -o myapp-debug main.go

# Enable all optimizations
go build -ldflags="-s -w" -o myapp-optimized main.go

# Link-time optimization
go build -ldflags="-X main.version=1.0.0 -s -w" -o myapp main.go

# Profile-guided optimization (Go 1.20+)
go build -pgo=cpu.prof -o myapp-pgo main.go
```

### Advanced Build Configuration

```go
// Build constraints for optimization
//go:build optimization
// +build optimization

package main

// Compiler directives
//go:noinline
func expensiveFunction() {
    // Force no inlining for profiling
}

//go:nosplit
func lowLevelFunction() {
    // Prevent stack splitting
}

//go:noescape
func noescape(p *int) {
    // Hint that pointer doesn't escape
}

//go:linkname fastFunction runtime.fastFunction
func fastFunction()

// Build with optimization:
// go build -tags optimization main.go
```

### Custom Build Scripts

```bash
#!/bin/bash
# build-optimized.sh

# Set optimization flags
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# Build with maximum optimizations
go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD)" \
    -o app-optimized \
    ./cmd/app

# Strip additional debug info (if available)
if command -v strip &> /dev/null; then
    strip app-optimized
fi

# Compress binary (if upx available)
if command -v upx &> /dev/null; then
    upx --best app-optimized
fi

echo "Optimized binary created: app-optimized"
ls -lh app-optimized
```

## Inlining Optimization

### Understanding Function Inlining

```go
// Inlining candidates - small, simple functions
//go:inline
func fastMath(x int) int {
    return x*x + 2*x + 1 // Will be inlined
}

// Too complex for inlining
func complexFunction(data []int) int {
    sum := 0
    for i, v := range data {
        if i%2 == 0 {
            sum += v * v
        } else {
            sum += v * 2
        }
    }
    return sum // Unlikely to be inlined due to complexity
}

// Interface calls - harder to inline
type Calculator interface {
    Calculate(int) int
}

type SimpleCalc struct{}

func (sc SimpleCalc) Calculate(x int) int {
    return x * 2 // May not be inlined due to interface
}

// Direct function calls - easier to inline
func directCall(x int) int {
    return fastMath(x) // fastMath will be inlined here
}

// Benchmark inlining effects
func BenchmarkInlining(b *testing.B) {
    b.Run("Inlined", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = fastMath(i)
        }
    })
    
    b.Run("NotInlined", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = complexFunction([]int{i, i + 1, i + 2})
        }
    })
    
    b.Run("InterfaceCall", func(b *testing.B) {
        calc := SimpleCalc{}
        for i := 0; i < b.N; i++ {
            _ = calc.Calculate(i)
        }
    })
}
```

### Controlling Inlining

```go
// Force inlining for critical paths
//go:inline
func criticalPath(x, y int) int {
    return x*y + x - y
}

// Prevent inlining for debugging
//go:noinline
func debugFunction(data []byte) []byte {
    // Keep this separate for profiling
    result := make([]byte, len(data))
    copy(result, data)
    return result
}

// Mid-level function - let compiler decide
func processData(input []int) []int {
    output := make([]int, len(input))
    for i, v := range input {
        output[i] = criticalPath(v, i) // Will be inlined
    }
    return output
}
```

## Escape Analysis Optimization

### Understanding Escape Analysis

```go
// Check escape analysis with: go build -gcflags="-m" main.go

// Stack allocation - no escape
func stackAllocation() {
    x := 42        // Stays on stack
    y := &x       // Local pointer, stays on stack
    _ = *y
} // x and y are deallocated when function returns

// Heap allocation - escapes
func heapAllocation() *int {
    x := 42        // Escapes to heap
    return &x      // Pointer returned, must be on heap
}

// Interface allocation - may escape
func interfaceAllocation() interface{} {
    x := 42        // Escapes to heap
    return x       // Boxed in interface
}

// Slice allocation analysis
func sliceAnalysis() {
    // Stack allocation
    s1 := make([]int, 10)     // Small slice, stays on stack
    _ = s1
    
    // Heap allocation
    s2 := make([]int, 10000)  // Large slice, goes to heap
    _ = s2
    
    // Escape through return
    s3 := make([]int, 5)
    _ = s3
    // If s3 is returned, it would escape
}

// Optimizing for stack allocation
func optimizedStackUsage() {
    const maxStackSize = 1000
    
    // Use stack allocation when possible
    var buffer [maxStackSize]byte
    
    // Process in chunks to stay on stack
    for i := 0; i < len(buffer); i += 100 {
        end := i + 100
        if end > len(buffer) {
            end = len(buffer)
        }
        processChunk(buffer[i:end])
    }
}

func processChunk(chunk []byte) {
    // Process chunk without escaping
    for i, b := range chunk {
        chunk[i] = b ^ 0xFF
    }
}
```

### Minimizing Heap Allocations

```go
// Inefficient - multiple heap allocations
func inefficientProcessing(data []string) []string {
    var results []string
    for _, item := range data {
        // String concatenation creates new allocations
        processed := "prefix_" + item + "_suffix"
        results = append(results, processed)
    }
    return results
}

// Optimized - minimize allocations
func optimizedProcessing(data []string) []string {
    // Pre-allocate result slice
    results := make([]string, 0, len(data))
    
    // Reuse string builder
    var builder strings.Builder
    
    for _, item := range data {
        builder.Reset()
        builder.Grow(7 + len(item) + 7) // "prefix_" + item + "_suffix"
        
        builder.WriteString("prefix_")
        builder.WriteString(item)
        builder.WriteString("_suffix")
        
        results = append(results, builder.String())
    }
    
    return results
}

// Pool-based optimization
var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

func poolBasedProcessing(data []string) []string {
    results := make([]string, 0, len(data))
    
    builder := stringBuilderPool.Get().(*strings.Builder)
    defer func() {
        builder.Reset()
        stringBuilderPool.Put(builder)
    }()
    
    for _, item := range data {
        builder.Reset()
        builder.WriteString("prefix_")
        builder.WriteString(item)
        builder.WriteString("_suffix")
        results = append(results, builder.String())
    }
    
    return results
}
```

## Dead Code Elimination

### Conditional Compilation

```go
// Build tags for dead code elimination
//go:build production
// +build production

package config

const (
    DebugMode = false
    LogLevel  = "error"
)

// debug.go - separate file
//go:build !production
// +build !production

package config

const (
    DebugMode = true
    LogLevel  = "debug"
)

// Usage - debug code eliminated in production builds
func processRequest(req Request) Response {
    if DebugMode {
        // This entire block eliminated in production
        logDebug("Processing request: %+v", req)
        validateRequest(req)
    }
    
    return handleRequest(req)
}

// Build production: go build -tags production
// Build debug: go build (default)
```

### Compiler-Assisted Dead Code Elimination

```go
// Constants allow compile-time optimization
const EnableLogging = false
const MaxCacheSize = 1000

func optimizedFunction(data []int) []int {
    // Compiler eliminates this branch if EnableLogging = false
    if EnableLogging {
        fmt.Printf("Processing %d items\n", len(data))
    }
    
    // Compiler can optimize fixed-size loops
    cache := make([]int, MaxCacheSize)
    
    for i := 0; i < MaxCacheSize && i < len(data); i++ {
        cache[i] = data[i] * 2
    }
    
    return cache[:min(MaxCacheSize, len(data))]
}

// Interface elimination through devirtualization
type Processor interface {
    Process(int) int
}

type SimpleProcessor struct{}

func (sp SimpleProcessor) Process(x int) int {
    return x * 2
}

// When type is known at compile time, interface call can be optimized
func processWithKnownType(data []int) []int {
    processor := SimpleProcessor{} // Concrete type known
    
    results := make([]int, len(data))
    for i, v := range data {
        results[i] = processor.Process(v) // Can be inlined/devirtualized
    }
    
    return results
}
```

## Profile-Guided Optimization (PGO)

### Collecting Profiles for PGO

```go
// main.go - application with profiling
package main

import (
    "os"
    "runtime/pprof"
)

func main() {
    // CPU profiling for PGO
    if cpuProfile := os.Getenv("CPUPROFILE"); cpuProfile != "" {
        f, err := os.Create(cpuProfile)
        if err != nil {
            panic(err)
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            panic(err)
        }
        defer pprof.StopCPUProfile()
    }
    
    // Your application workload
    runApplicationWorkload()
}

func runApplicationWorkload() {
    // Simulate typical workload
    for i := 0; i < 1000000; i++ {
        processData(generateData(i))
    }
}

func generateData(seed int) []int {
    data := make([]int, 1000)
    for i := range data {
        data[i] = (seed + i) % 1000
    }
    return data
}

func processData(data []int) int {
    sum := 0
    for _, v := range data {
        if v%2 == 0 {
            sum += v * v
        } else {
            sum += v * 3
        }
    }
    return sum
}
```

### Building with PGO

```bash
#!/bin/bash
# Build process with PGO

# 1. Build instrumented binary
go build -o app-instrumented main.go

# 2. Run with profiling to collect profile
CPUPROFILE=cpu.prof ./app-instrumented

# 3. Build optimized binary with profile
go build -pgo=cpu.prof -o app-optimized main.go

# 4. Compare performance
echo "Benchmarking instrumented version:"
./app-instrumented &
time ./app-instrumented

echo "Benchmarking PGO optimized version:"
time ./app-optimized
```

### PGO Optimization Analysis

```go
// Analyze PGO effectiveness
func BenchmarkPGOEffectiveness(b *testing.B) {
    data := generateTestData(10000)
    
    b.Run("HotPath", func(b *testing.B) {
        // Function that should be optimized by PGO
        for i := 0; i < b.N; i++ {
            _ = hotPathFunction(data)
        }
    })
    
    b.Run("ColdPath", func(b *testing.B) {
        // Function rarely called in profile
        for i := 0; i < b.N; i++ {
            _ = coldPathFunction(data)
        }
    })
}

// Hot path - frequently called in profile
func hotPathFunction(data []int) int {
    sum := 0
    for _, v := range data {
        sum += complexCalculation(v)
    }
    return sum
}

// Cold path - rarely called in profile
func coldPathFunction(data []int) int {
    product := 1
    for _, v := range data {
        if v > 0 {
            product *= v
            if product > 1000000 {
                break
            }
        }
    }
    return product
}

func complexCalculation(x int) int {
    // Complex enough to benefit from optimization
    result := x
    for i := 0; i < 10; i++ {
        result = result*result + result + 1
        result = result % 1000000
    }
    return result
}
```

## Advanced Compiler Optimizations

### Loop Optimizations

```go
// Loop unrolling candidate
func processArrayOptimized(data []int) {
    // Compiler may unroll this loop
    for i := 0; i < len(data); i += 4 {
        // Process 4 elements at once for vectorization
        if i+3 < len(data) {
            data[i] *= 2
            data[i+1] *= 2
            data[i+2] *= 2
            data[i+3] *= 2
        } else {
            // Handle remaining elements
            for j := i; j < len(data); j++ {
                data[j] *= 2
            }
            break
        }
    }
}

// Loop invariant code motion
func loopInvariantOptimization(matrix [][]int, multiplier int) {
    // Compiler moves invariant calculations out of loop
    n := len(matrix)
    m := len(matrix[0])
    
    // These calculations are loop invariant
    threshold := multiplier * 100
    factor := multiplier + 5
    
    for i := 0; i < n; i++ {
        for j := 0; j < m; j++ {
            if matrix[i][j] > threshold {
                matrix[i][j] *= factor
            }
        }
    }
}

// Strength reduction
func strengthReduction(n int) []int {
    result := make([]int, n)
    
    // Compiler optimizes multiplication to addition
    for i := 0; i < n; i++ {
        result[i] = i * 7 // May be optimized to addition
    }
    
    return result
}
```

### Bounds Check Elimination

```go
// Bounds check elimination patterns
func boundsCheckElimination(data []int) {
    n := len(data)
    
    // Pattern 1: Compiler eliminates bounds checks
    for i := 0; i < n; i++ {
        data[i] = i // No bounds check needed
    }
    
    // Pattern 2: Use _ to hint no bounds check needed
    for i := range data {
        _ = data[i] // Eliminates bounds check
        data[i] = i * 2
    }
    
    // Pattern 3: Manual bounds check removal
    if len(data) >= 10 {
        // Compiler knows these are safe
        data[0] = 1
        data[9] = 2
    }
}

// Slice bounds optimization
func sliceBoundsOptimization(data []int) []int {
    if len(data) < 100 {
        return data
    }
    
    // Compiler eliminates bounds checks for this slice
    subset := data[10:90] // Known to be within bounds
    
    for i := range subset {
        subset[i] *= 2 // No bounds check needed
    }
    
    return subset
}
```

## Measuring Compiler Optimization Impact

### Compilation Analysis Tools

```bash
# View compiler optimizations
go build -gcflags="-m -m" main.go 2>&1 | grep -E "(inlin|escap|alloc)"

# View SSA intermediate representation
go build -gcflags="-S" main.go > assembly.txt

# Analyze binary size
go build -ldflags="-s -w" -o optimized main.go
go build -o unoptimized main.go
ls -lh optimized unoptimized

# Profile-guided optimization report
go build -pgo=cpu.prof -gcflags="-m" main.go 2>&1 | grep "PGO"
```

### Performance Validation

```go
// Benchmark compiler optimizations
func BenchmarkCompilerOptimizations(b *testing.B) {
    data := make([]int, 1000)
    for i := range data {
        data[i] = i
    }
    
    b.Run("InlinedFunction", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = inlinedMath(data[i%len(data)])
        }
    })
    
    b.Run("NonInlinedFunction", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = nonInlinedMath(data[i%len(data)])
        }
    })
    
    b.Run("BoundsCheckEliminated", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            processSafeBounds(data)
        }
    })
    
    b.Run("BoundsCheckPresent", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            processUnsafeBounds(data)
        }
    })
}

//go:inline
func inlinedMath(x int) int {
    return x*x + 2*x + 1
}

//go:noinline
func nonInlinedMath(x int) int {
    return x*x + 2*x + 1
}

func processSafeBounds(data []int) {
    for i := 0; i < len(data); i++ {
        data[i] *= 2 // Bounds check eliminated
    }
}

func processUnsafeBounds(data []int) {
    for i := 0; i < 2000; i++ {
        if i < len(data) {
            data[i] *= 2 // Bounds check required
        }
    }
}
```

## Best Practices for Compiler Optimization

### 1. Write Optimizer-Friendly Code
- Use constants instead of variables when values don't change
- Prefer small, simple functions for inlining
- Use concrete types instead of interfaces when possible
- Minimize pointer indirection

### 2. Leverage Build Flags Appropriately
- Use `-ldflags="-s -w"` for production builds
- Apply PGO for performance-critical applications
- Use build tags for conditional compilation
- Profile before and after optimization

### 3. Understand Escape Analysis
- Keep allocations on the stack when possible
- Avoid returning pointers to local variables unnecessarily
- Use value receivers for small structs
- Pre-allocate slices and maps when size is known

### 4. Monitor and Validate
- Use benchmarks to measure optimization impact
- Profile with `go tool pprof` to verify improvements
- Monitor binary size and startup performance
- Test optimized builds thoroughly

The key to effective compiler optimization is understanding how the Go compiler works, writing code that enables optimizations, and continuously measuring the impact of your optimization efforts.
