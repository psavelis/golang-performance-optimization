# Benchmark Analysis

Master the art of interpreting benchmark results to make data-driven optimization decisions and avoid common pitfalls in performance analysis.

## Understanding Benchmark Output

### Basic Benchmark Results Format

```
BenchmarkExample-8    1000000    1234 ns/op    456 B/op    7 allocs/op
│                │    │          │             │           │
│                │    │          │             │           └─ Allocations per operation
│                │    │          │             └─ Bytes allocated per operation
│                │    │          └─ Nanoseconds per operation
│                │    └─ Number of iterations
│                └─ Number of CPU cores used
└─ Benchmark name
```

### Memory Statistics

```go
func BenchmarkStringOperations(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result := fmt.Sprintf("iteration_%d", i)
        _ = result
    }
}
```

Example output:
```
BenchmarkStringOperations-8    5000000    245 ns/op    16 B/op    1 allocs/op
```

## Statistical Analysis of Results

### Running Multiple Iterations

```bash
# Run benchmark multiple times for statistical significance
go test -bench=BenchmarkExample -count=10

# Use benchstat for analysis
go get golang.org/x/perf/cmd/benchstat
go test -bench=BenchmarkExample -count=10 > before.txt
# Make changes
go test -bench=BenchmarkExample -count=10 > after.txt
benchstat before.txt after.txt
```

### Sample benchstat Output

```
name                old time/op    new time/op    delta
StringOperations-8    245ns ± 2%     198ns ± 3%   -19.18%  (p=0.000 n=10+10)

name                old alloc/op   new alloc/op   delta
StringOperations-8    16.0B ± 0%     12.0B ± 0%   -25.00%  (p=0.000 n=10+10)

name                old allocs/op  new allocs/op  delta
StringOperations-8    1.00 ± 0%      1.00 ± 0%     ~     (all equal)
```

## Common Analysis Patterns

### Performance Regression Detection

```go
func BenchmarkSuite(b *testing.B) {
    testCases := []struct {
        name string
        size int
    }{
        {"Small", 10},
        {"Medium", 100},
        {"Large", 1000},
        {"XLarge", 10000},
    }
    
    for _, tc := range testCases {
        b.Run(tc.name, func(b *testing.B) {
            data := generateTestData(tc.size)
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                result := processData(data)
                _ = result
            }
        })
    }
}
```

### Scalability Analysis

```go
func BenchmarkScalability(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000, 100000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            data := make([]int, size)
            for i := range data {
                data[i] = rand.Intn(1000)
            }
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                sort.Ints(data)
            }
        })
    }
}
```

Example results showing O(n log n) complexity:
```
BenchmarkScalability/Size_10-8       5000000    245 ns/op
BenchmarkScalability/Size_100-8       500000   2834 ns/op
BenchmarkScalability/Size_1000-8       50000  34521 ns/op
BenchmarkScalability/Size_10000-8       5000 456789 ns/op
```

## Memory Analysis

### Understanding Allocation Patterns

```go
func BenchmarkMemoryPatterns(b *testing.B) {
    b.Run("PreAllocated", func(b *testing.B) {
        b.ReportAllocs()
        slice := make([]int, 0, 1000) // Pre-allocate capacity
        
        for i := 0; i < b.N; i++ {
            slice = slice[:0] // Reset length, keep capacity
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
    
    b.Run("GrowthPattern", func(b *testing.B) {
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            var slice []int // Start with zero capacity
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
}
```

Results comparison:
```
BenchmarkMemoryPatterns/PreAllocated-8    500000   2456 ns/op    8192 B/op    1 allocs/op
BenchmarkMemoryPatterns/GrowthPattern-8    100000  12834 ns/op   24576 B/op   11 allocs/op
```

### String vs StringBuilder

```go
func BenchmarkStringBuilding(b *testing.B) {
    words := []string{"hello", "world", "benchmark", "analysis"}
    
    b.Run("StringConcatenation", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var result string
            for _, word := range words {
                result += word + " "
            }
            _ = result
        }
    })
    
    b.Run("StringBuilder", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            builder.Grow(50) // Pre-allocate expected size
            for _, word := range words {
                builder.WriteString(word)
                builder.WriteString(" ")
            }
            _ = builder.String()
        }
    })
    
    b.Run("ByteBuffer", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var buf bytes.Buffer
            buf.Grow(50) // Pre-allocate expected size
            for _, word := range words {
                buf.WriteString(word)
                buf.WriteString(" ")
            }
            _ = buf.String()
        }
    })
}
```

## CPU Profiling Integration

### Profile-Guided Analysis

```go
func BenchmarkWithProfile(b *testing.B) {
    // Enable CPU profiling
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            b.Fatal(err)
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            b.Fatal(err)
        }
        defer pprof.StopCPUProfile()
    }
    
    // Your benchmark code
    for i := 0; i < b.N; i++ {
        result := expensiveFunction()
        _ = result
    }
}
```

Command line usage:
```bash
go test -bench=BenchmarkWithProfile -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Comparative Analysis Techniques

### Before/After Performance Analysis

```go
// benchmark_test.go
func BenchmarkOptimizationComparison(b *testing.B) {
    data := generateLargeTestData()
    
    b.Run("Original", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := originalAlgorithm(data)
            _ = result
        }
    })
    
    b.Run("Optimized", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := optimizedAlgorithm(data)
            _ = result
        }
    })
}
```

### Cross-Platform Performance

```go
func BenchmarkCrossPlatform(b *testing.B) {
    b.Run(fmt.Sprintf("GOOS_%s_GOARCH_%s", runtime.GOOS, runtime.GOARCH), func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := platformSensitiveOperation()
            _ = result
        }
    })
}
```

## Advanced Analysis Patterns

### Performance Regression Testing

```go
// Create a benchmark suite for regression testing
func BenchmarkRegressionSuite(b *testing.B) {
    benchmarks := []struct {
        name     string
        function func() interface{}
        maxTime  time.Duration
        maxAlloc int64
    }{
        {"CriticalPath", criticalPathFunction, 100 * time.Microsecond, 1024},
        {"DataProcessing", dataProcessingFunction, 1 * time.Millisecond, 4096},
        {"NetworkIO", networkIOFunction, 10 * time.Millisecond, 8192},
    }
    
    for _, bm := range benchmarks {
        b.Run(bm.name, func(b *testing.B) {
            b.ReportAllocs()
            
            start := time.Now()
            var totalAlloc int64
            
            for i := 0; i < b.N; i++ {
                before := getAllocatedBytes()
                result := bm.function()
                after := getAllocatedBytes()
                
                totalAlloc += after - before
                _ = result
            }
            
            avgTime := time.Since(start) / time.Duration(b.N)
            avgAlloc := totalAlloc / int64(b.N)
            
            if avgTime > bm.maxTime {
                b.Errorf("Performance regression: %v > %v", avgTime, bm.maxTime)
            }
            
            if avgAlloc > bm.maxAlloc {
                b.Errorf("Memory regression: %d > %d bytes", avgAlloc, bm.maxAlloc)
            }
        })
    }
}
```

### Latency Distribution Analysis

```go
func BenchmarkLatencyDistribution(b *testing.B) {
    var latencies []time.Duration
    
    b.Run("LatencyMeasurement", func(b *testing.B) {
        latencies = make([]time.Duration, b.N)
        
        for i := 0; i < b.N; i++ {
            start := time.Now()
            result := operationWithVariableLatency()
            latencies[i] = time.Since(start)
            _ = result
        }
    })
    
    // Analyze latency distribution
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })
    
    p50 := latencies[len(latencies)*50/100]
    p95 := latencies[len(latencies)*95/100]
    p99 := latencies[len(latencies)*99/100]
    
    b.Logf("Latency P50: %v, P95: %v, P99: %v", p50, p95, p99)
}
```

## Benchmark Result Validation

### Consistency Checks

```go
func BenchmarkConsistency(b *testing.B) {
    const expectedResult = 42
    
    b.Run("ConsistencyCheck", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := deterministicFunction()
            if result != expectedResult {
                b.Fatalf("Inconsistent result: got %d, want %d", result, expectedResult)
            }
        }
    })
}
```

### Variance Analysis

```go
func BenchmarkVarianceAnalysis(b *testing.B) {
    measurements := make([]float64, 100)
    
    for run := 0; run < len(measurements); run++ {
        start := time.Now()
        
        for i := 0; i < 10000; i++ {
            result := functionUnderTest()
            _ = result
        }
        
        measurements[run] = float64(time.Since(start).Nanoseconds()) / 10000.0
    }
    
    mean := calculateMean(measurements)
    stddev := calculateStdDev(measurements, mean)
    cv := stddev / mean // Coefficient of variation
    
    b.Logf("Mean: %.2f ns, StdDev: %.2f ns, CV: %.2f%%", mean, stddev, cv*100)
    
    if cv > 0.1 { // More than 10% variation
        b.Logf("Warning: High variance detected (CV: %.2f%%)", cv*100)
    }
}
```

## Automated Analysis Tools

### Custom Benchmark Analysis

```go
type BenchmarkResult struct {
    Name        string
    Iterations  int
    NsPerOp     float64
    BytesPerOp  int64
    AllocsPerOp int64
    MBPerSec    float64
}

func AnalyzeBenchmarkResults(results []BenchmarkResult) {
    for _, result := range results {
        efficiency := result.MBPerSec / float64(result.AllocsPerOp)
        
        fmt.Printf("Benchmark: %s\n", result.Name)
        fmt.Printf("  Performance: %.2f ns/op\n", result.NsPerOp)
        fmt.Printf("  Memory efficiency: %.2f MB/s per allocation\n", efficiency)
        
        if result.NsPerOp > 1000 {
            fmt.Printf("  ⚠️  Slow operation detected\n")
        }
        
        if result.AllocsPerOp > 0 {
            fmt.Printf("  💾 Memory allocations: %d allocs/op\n", result.AllocsPerOp)
        }
        
        fmt.Println()
    }
}
```

## Common Analysis Pitfalls

### 1. Insufficient Sample Size

```go
// BAD: Too few iterations for reliable results
func BenchmarkUnreliable(b *testing.B) {
    if b.N < 1000 {
        b.N = 1000 // Force minimum iterations
    }
    
    for i := 0; i < b.N; i++ {
        result := randomVariableFunction()
        _ = result
    }
}
```

### 2. Ignoring Warmup Effects

```go
// GOOD: Account for JIT warmup and cache effects
func BenchmarkWithWarmup(b *testing.B) {
    // Warmup phase
    for i := 0; i < 1000; i++ {
        _ = functionUnderTest()
    }
    
    b.ResetTimer() // Start measuring after warmup
    
    for i := 0; i < b.N; i++ {
        result := functionUnderTest()
        _ = result
    }
}
```

### 3. Environmental Factors

```go
func BenchmarkEnvironmentalAware(b *testing.B) {
    // Check system load
    if getSystemLoad() > 0.8 {
        b.Skip("System under high load, skipping benchmark")
    }
    
    // Disable GC during critical measurements
    gcPercent := debug.SetGCPercent(-1)
    defer debug.SetGCPercent(gcPercent)
    
    runtime.GC() // Clean slate
    
    for i := 0; i < b.N; i++ {
        result := memoryIntensiveFunction()
        _ = result
    }
}
```

Proper benchmark analysis transforms raw performance data into actionable insights, enabling you to make informed optimization decisions and maintain performance standards over time.
