# Basic Benchmarks

Master the fundamentals of Go benchmarking with practical examples and best practices for creating reliable performance measurements.

## Benchmark Fundamentals

Go's built-in testing package provides powerful benchmarking capabilities. A benchmark function must:

- Begin with the word "Benchmark"
- Take a `*testing.B` parameter
- Run the code under test `b.N` times

## Basic Benchmark Structure

```go
package main

import (
    "testing"
    "strings"
)

// Simple string concatenation benchmark
func BenchmarkStringConcatenation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := "hello" + "world"
        _ = result // Prevent compiler optimization
    }
}

// String builder benchmark for comparison
func BenchmarkStringBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        sb.WriteString("hello")
        sb.WriteString("world")
        _ = sb.String()
    }
}
```

## Essential Benchmark Patterns

### 1. Memory Allocation Measurement

```go
func BenchmarkSliceAllocation(b *testing.B) {
    b.ReportAllocs() // Report allocation statistics
    
    for i := 0; i < b.N; i++ {
        slice := make([]int, 1000)
        _ = slice
    }
}

// Pre-allocated slice for comparison
func BenchmarkSlicePreallocated(b *testing.B) {
    b.ReportAllocs()
    slice := make([]int, 1000)
    
    for i := 0; i < b.N; i++ {
        // Reuse pre-allocated slice
        for j := range slice {
            slice[j] = j
        }
    }
}
```

### 2. Setup and Teardown

```go
func BenchmarkDatabaseOperation(b *testing.B) {
    // Setup (not measured)
    db := setupTestDatabase()
    defer db.Close()
    
    b.ResetTimer() // Start measuring from here
    
    for i := 0; i < b.N; i++ {
        // Only this code is measured
        result := db.Query("SELECT * FROM users LIMIT 1")
        result.Close()
    }
}
```

### 3. Proper Resource Management

```go
func BenchmarkFileOperation(b *testing.B) {
    // Create test data
    testData := make([]byte, 1024)
    for i := range testData {
        testData[i] = byte(i % 256)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        b.StopTimer() // Exclude setup from measurement
        file, err := os.CreateTemp("", "benchmark")
        if err != nil {
            b.Fatal(err)
        }
        b.StartTimer() // Resume measurement
        
        // Measured operation
        _, err = file.Write(testData)
        if err != nil {
            b.Fatal(err)
        }
        
        b.StopTimer() // Exclude cleanup
        file.Close()
        os.Remove(file.Name())
        b.StartTimer()
    }
}
```

## Best Practices

### 1. Prevent Compiler Optimizations

```go
// BAD: Compiler might optimize away the operation
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := computeExpensiveFunction()
        // result is unused - might be optimized away
    }
}

// GOOD: Assign to package-level variable
var result int

func BenchmarkGood(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = computeExpensiveFunction()
    }
    result = r // Prevent optimization
}
```

### 2. Realistic Data Sizes

```go
// Test with various data sizes
func BenchmarkSortSmall(b *testing.B) {
    data := generateRandomSlice(100)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        sort.Ints(data)
    }
}

func BenchmarkSortLarge(b *testing.B) {
    data := generateRandomSlice(10000)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        sort.Ints(data)
    }
}
```

### 3. Stable Benchmark Environment

```go
func BenchmarkCriticalPath(b *testing.B) {
    // Stabilize goroutine scheduler
    runtime.GC()
    runtime.GC()
    
    // Ensure consistent CPU state
    runtime.GOMAXPROCS(1)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        criticalOperation()
    }
}
```

## Common Pitfalls

### 1. Including Setup in Measurement

```go
// BAD: Setup time included in benchmark
func BenchmarkBadSetup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        data := generateLargeDataset() // Expensive setup
        processData(data)
    }
}

// GOOD: Setup excluded from measurement
func BenchmarkGoodSetup(b *testing.B) {
    data := generateLargeDataset() // Setup once
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        processData(data)
    }
}
```

### 2. Incorrect Loop Structure

```go
// BAD: Timer operations inside loop
func BenchmarkBadLoop(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        setup()
        b.StartTimer()
        operation()
    }
}

// GOOD: Minimize timer operations
func BenchmarkGoodLoop(b *testing.B) {
    for i := 0; i < b.N; i++ {
        operation()
    }
}
```

### 3. Non-deterministic Operations

```go
// BAD: Random operations affect consistency
func BenchmarkBadRandom(b *testing.B) {
    for i := 0; i < b.N; i++ {
        data := generateRandomData() // Different each time
        processData(data)
    }
}

// GOOD: Consistent test data
func BenchmarkGoodConsistent(b *testing.B) {
    data := generateFixedTestData()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        processData(data)
    }
}
```

## Running Benchmarks

### Basic Execution

```bash
# Run all benchmarks
go test -bench=.

# Run specific benchmark
go test -bench=BenchmarkStringConcatenation

# Run with memory allocation stats
go test -bench=. -benchmem

# Multiple runs for stability
go test -bench=. -count=5
```

### Advanced Options

```bash
# Control benchmark time
go test -bench=. -benchtime=10s

# Control iterations
go test -bench=. -benchtime=1000x

# CPU profiling during benchmark
go test -bench=. -cpuprofile=cpu.prof

# Memory profiling during benchmark
go test -bench=. -memprofile=mem.prof
```

## Interpreting Results

```
BenchmarkStringConcatenation-8    50000000    25.4 ns/op    0 B/op    0 allocs/op
BenchmarkStringBuilder-8          30000000    41.2 ns/op   64 B/op    1 allocs/op
```

This output shows:
- `50000000`: Number of iterations
- `25.4 ns/op`: Nanoseconds per operation
- `0 B/op`: Bytes allocated per operation
- `0 allocs/op`: Allocations per operation
- `-8`: Number of CPUs used

## Example: Complete Benchmark Suite

```go
package main

import (
    "testing"
    "fmt"
    "strings"
)

// Benchmark different string joining methods
func BenchmarkStringJoinPlus(b *testing.B) {
    strings := []string{"hello", "world", "this", "is", "a", "test"}
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result := ""
        for _, s := range strings {
            result += s
        }
        _ = result
    }
}

func BenchmarkStringJoinBuilder(b *testing.B) {
    strings := []string{"hello", "world", "this", "is", "a", "test"}
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        var builder strings.Builder
        for _, s := range strings {
            builder.WriteString(s)
        }
        _ = builder.String()
    }
}

func BenchmarkStringJoinBuilderPrealloc(b *testing.B) {
    strings := []string{"hello", "world", "this", "is", "a", "test"}
    
    // Calculate total length
    totalLen := 0
    for _, s := range strings {
        totalLen += len(s)
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        var builder strings.Builder
        builder.Grow(totalLen) // Pre-allocate
        for _, s := range strings {
            builder.WriteString(s)
        }
        _ = builder.String()
    }
}

func BenchmarkStringJoinStdlib(b *testing.B) {
    strings := []string{"hello", "world", "this", "is", "a", "test"}
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result := strings.Join(strings, "")
        _ = result
    }
}
```

This comprehensive approach to basic benchmarking provides the foundation for accurate performance measurement in Go applications.
