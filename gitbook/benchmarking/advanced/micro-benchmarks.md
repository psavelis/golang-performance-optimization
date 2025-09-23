# Micro-benchmarks

Master the art of precise, focused benchmarking for specific code paths and optimizations in Go applications.

## Understanding Micro-benchmarks

Micro-benchmarks measure the performance of small, isolated pieces of code to understand their precise behavior and identify optimization opportunities.

### Basic Micro-benchmark Structure

```go
// Simple function to benchmark
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

// Basic micro-benchmark
func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        fibonacci(20)
    }
}

// Benchmark with different input sizes
func BenchmarkFibonacciSizes(b *testing.B) {
    sizes := []int{10, 15, 20, 25}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                fibonacci(size)
            }
        })
    }
}
```

### Memory Allocation Benchmarking

```go
// String concatenation micro-benchmarks
func BenchmarkStringConcat(b *testing.B) {
    parts := []string{"hello", "world", "from", "golang", "benchmark"}
    
    b.Run("Plus", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            result := ""
            for _, part := range parts {
                result += part
            }
            _ = result
        }
    })
    
    b.Run("Builder", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for _, part := range parts {
                builder.WriteString(part)
            }
            _ = builder.String()
        }
    })
    
    b.Run("BuilderPrealloc", func(b *testing.B) {
        b.ReportAllocs()
        totalLen := 0
        for _, part := range parts {
            totalLen += len(part)
        }
        
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            builder.Grow(totalLen)
            for _, part := range parts {
                builder.WriteString(part)
            }
            _ = builder.String()
        }
    })
    
    b.Run("Join", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            result := strings.Join(parts, "")
            _ = result
        }
    })
}
```

### CPU-Intensive Micro-benchmarks

```go
// Algorithm comparison micro-benchmarks
func BenchmarkSortingAlgorithms(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        data := generateRandomData(size)
        
        b.Run(fmt.Sprintf("QuickSort_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                b.StopTimer()
                testData := make([]int, len(data))
                copy(testData, data)
                b.StartTimer()
                
                quickSort(testData, 0, len(testData)-1)
            }
        })
        
        b.Run(fmt.Sprintf("HeapSort_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                b.StopTimer()
                testData := make([]int, len(data))
                copy(testData, data)
                b.StartTimer()
                
                heapSort(testData)
            }
        })
        
        b.Run(fmt.Sprintf("StdSort_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                b.StopTimer()
                testData := make([]int, len(data))
                copy(testData, data)
                b.StartTimer()
                
                sort.Ints(testData)
            }
        })
    }
}

func generateRandomData(size int) []int {
    rand.Seed(42) // Fixed seed for reproducible benchmarks
    data := make([]int, size)
    for i := range data {
        data[i] = rand.Intn(size * 10)
    }
    return data
}

func quickSort(arr []int, low, high int) {
    if low < high {
        pivot := partition(arr, low, high)
        quickSort(arr, low, pivot-1)
        quickSort(arr, pivot+1, high)
    }
}

func partition(arr []int, low, high int) int {
    pivot := arr[high]
    i := low - 1
    
    for j := low; j < high; j++ {
        if arr[j] <= pivot {
            i++
            arr[i], arr[j] = arr[j], arr[i]
        }
    }
    
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1
}

func heapSort(arr []int) {
    n := len(arr)
    
    // Build heap
    for i := n/2 - 1; i >= 0; i-- {
        heapify(arr, n, i)
    }
    
    // Extract elements
    for i := n - 1; i > 0; i-- {
        arr[0], arr[i] = arr[i], arr[0]
        heapify(arr, i, 0)
    }
}

func heapify(arr []int, n, i int) {
    largest := i
    left := 2*i + 1
    right := 2*i + 2
    
    if left < n && arr[left] > arr[largest] {
        largest = left
    }
    
    if right < n && arr[right] > arr[largest] {
        largest = right
    }
    
    if largest != i {
        arr[i], arr[largest] = arr[largest], arr[i]
        heapify(arr, n, largest)
    }
}
```

### Data Structure Performance Micro-benchmarks

```go
// Data structure access pattern benchmarks
func BenchmarkDataStructures(b *testing.B) {
    const size = 10000
    
    // Generate test data
    keys := make([]string, size)
    for i := 0; i < size; i++ {
        keys[i] = fmt.Sprintf("key_%06d", i)
    }
    
    // Map benchmarks
    b.Run("Map", func(b *testing.B) {
        m := make(map[string]int, size)
        for i, key := range keys {
            m[key] = i
        }
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            key := keys[i%size]
            _ = m[key]
        }
    })
    
    // Slice with linear search
    b.Run("SliceLinear", func(b *testing.B) {
        type keyValue struct {
            key   string
            value int
        }
        
        slice := make([]keyValue, size)
        for i, key := range keys {
            slice[i] = keyValue{key: key, value: i}
        }
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            target := keys[i%size]
            for _, kv := range slice {
                if kv.key == target {
                    _ = kv.value
                    break
                }
            }
        }
    })
    
    // Sorted slice with binary search
    b.Run("SliceBinary", func(b *testing.B) {
        sortedKeys := make([]string, len(keys))
        copy(sortedKeys, keys)
        sort.Strings(sortedKeys)
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            target := keys[i%size]
            idx := sort.SearchStrings(sortedKeys, target)
            if idx < len(sortedKeys) && sortedKeys[idx] == target {
                _ = idx
            }
        }
    })
}
```

### Memory Access Pattern Benchmarks

```go
// Cache-friendly vs cache-unfriendly access patterns
func BenchmarkMemoryAccess(b *testing.B) {
    const size = 1024 * 1024 // 1M elements
    
    // Sequential access benchmark
    b.Run("Sequential", func(b *testing.B) {
        data := make([]int64, size)
        for i := range data {
            data[i] = int64(i)
        }
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            sum := int64(0)
            for j := 0; j < size; j++ {
                sum += data[j]
            }
            _ = sum
        }
    })
    
    // Random access benchmark
    b.Run("Random", func(b *testing.B) {
        data := make([]int64, size)
        indices := make([]int, size)
        
        for i := range data {
            data[i] = int64(i)
            indices[i] = i
        }
        
        rand.Seed(42)
        rand.Shuffle(len(indices), func(i, j int) {
            indices[i], indices[j] = indices[j], indices[i]
        })
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            sum := int64(0)
            for j := 0; j < size; j++ {
                sum += data[indices[j]]
            }
            _ = sum
        }
    })
    
    // Strided access benchmark
    b.Run("Strided", func(b *testing.B) {
        data := make([]int64, size)
        for i := range data {
            data[i] = int64(i)
        }
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            sum := int64(0)
            for j := 0; j < size; j += 64 { // Skip cache lines
                sum += data[j]
            }
            _ = sum
        }
    })
}
```

### Function Call Overhead Benchmarks

```go
// Function call overhead analysis
func BenchmarkFunctionCalls(b *testing.B) {
    value := 42
    
    // Direct computation
    b.Run("Direct", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := value * 2 + 1
            _ = result
        }
    })
    
    // Function call
    b.Run("FunctionCall", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := simpleOperation(value)
            _ = result
        }
    })
    
    // Method call
    b.Run("MethodCall", func(b *testing.B) {
        calculator := Calculator{multiplier: 2}
        for i := 0; i < b.N; i++ {
            result := calculator.Calculate(value)
            _ = result
        }
    })
    
    // Interface call
    b.Run("InterfaceCall", func(b *testing.B) {
        var calculator Calculable = Calculator{multiplier: 2}
        for i := 0; i < b.N; i++ {
            result := calculator.Calculate(value)
            _ = result
        }
    })
    
    // Function pointer call
    b.Run("FunctionPointer", func(b *testing.B) {
        fn := simpleOperation
        for i := 0; i < b.N; i++ {
            result := fn(value)
            _ = result
        }
    })
}

func simpleOperation(x int) int {
    return x*2 + 1
}

type Calculator struct {
    multiplier int
}

func (c Calculator) Calculate(x int) int {
    return x*c.multiplier + 1
}

type Calculable interface {
    Calculate(int) int
}
```

### Lock Performance Micro-benchmarks

```go
// Lock contention and performance benchmarks
func BenchmarkLocks(b *testing.B) {
    var counter int64
    
    // No lock baseline
    b.Run("NoLock", func(b *testing.B) {
        localCounter := int64(0)
        for i := 0; i < b.N; i++ {
            localCounter++
        }
        _ = localCounter
    })
    
    // Mutex lock
    b.Run("Mutex", func(b *testing.B) {
        var mu sync.Mutex
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.Lock()
                counter++
                mu.Unlock()
            }
        })
    })
    
    // RWMutex read lock
    b.Run("RWMutexRead", func(b *testing.B) {
        var rwmu sync.RWMutex
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                rwmu.RLock()
                _ = counter
                rwmu.RUnlock()
            }
        })
    })
    
    // RWMutex write lock
    b.Run("RWMutexWrite", func(b *testing.B) {
        var rwmu sync.RWMutex
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                rwmu.Lock()
                counter++
                rwmu.Unlock()
            }
        })
    })
    
    // Atomic operations
    b.Run("Atomic", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                atomic.AddInt64(&counter, 1)
            }
        })
    })
    
    // Channel communication
    b.Run("Channel", func(b *testing.B) {
        ch := make(chan int64, 1)
        ch <- 0
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                val := <-ch
                ch <- val + 1
            }
        })
    })
}
```

### I/O Performance Micro-benchmarks

```go
// I/O operation micro-benchmarks
func BenchmarkIO(b *testing.B) {
    data := make([]byte, 4096) // 4KB buffer
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    // Memory copy
    b.Run("MemoryCopy", func(b *testing.B) {
        dest := make([]byte, len(data))
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            copy(dest, data)
        }
    })
    
    // File write
    b.Run("FileWrite", func(b *testing.B) {
        tmpFile, err := os.CreateTemp("", "benchmark")
        if err != nil {
            b.Fatal(err)
        }
        defer os.Remove(tmpFile.Name())
        defer tmpFile.Close()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            tmpFile.Seek(0, 0)
            tmpFile.Write(data)
        }
    })
    
    // Buffered file write
    b.Run("BufferedFileWrite", func(b *testing.B) {
        tmpFile, err := os.CreateTemp("", "benchmark")
        if err != nil {
            b.Fatal(err)
        }
        defer os.Remove(tmpFile.Name())
        defer tmpFile.Close()
        
        writer := bufio.NewWriter(tmpFile)
        defer writer.Flush()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            writer.Write(data)
        }
    })
    
    // Network write (to local server)
    b.Run("NetworkWrite", func(b *testing.B) {
        // Start a simple echo server
        listener, err := net.Listen("tcp", "localhost:0")
        if err != nil {
            b.Fatal(err)
        }
        defer listener.Close()
        
        go func() {
            for {
                conn, err := listener.Accept()
                if err != nil {
                    return
                }
                go func(c net.Conn) {
                    defer c.Close()
                    io.Copy(io.Discard, c)
                }(conn)
            }
        }()
        
        conn, err := net.Dial("tcp", listener.Addr().String())
        if err != nil {
            b.Fatal(err)
        }
        defer conn.Close()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            conn.Write(data)
        }
    })
}
```

## Advanced Micro-benchmark Techniques

### Custom Metrics and Reporting

```go
// Custom benchmark with additional metrics
func BenchmarkWithCustomMetrics(b *testing.B) {
    var totalAllocations int64
    var totalDuration time.Duration
    
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        start := time.Now()
        
        // Your code here
        data := make([]byte, 1024)
        for j := range data {
            data[j] = byte(j)
        }
        
        duration := time.Since(start)
        totalDuration += duration
        totalAllocations++
    }
    
    // Report custom metrics
    b.ReportMetric(float64(totalDuration.Nanoseconds())/float64(b.N), "ns/op")
    b.ReportMetric(float64(totalAllocations)/float64(b.N), "allocs/op")
}
```

### Benchmark Utilities

```go
// Utilities for consistent micro-benchmarking
func setupBenchmarkData(size int) []int {
    data := make([]int, size)
    for i := range data {
        data[i] = rand.Intn(1000)
    }
    return data
}

func runBenchmarkSizes(b *testing.B, sizes []int, benchFunc func(*testing.B, []int)) {
    for _, size := range sizes {
        data := setupBenchmarkData(size)
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            benchFunc(b, data)
        })
    }
}

// Example usage
func BenchmarkExampleWithSizes(b *testing.B) {
    sizes := []int{100, 1000, 10000, 100000}
    
    runBenchmarkSizes(b, sizes, func(b *testing.B, data []int) {
        for i := 0; i < b.N; i++ {
            sum := 0
            for _, v := range data {
                sum += v
            }
            _ = sum
        }
    })
}
```

## Best Practices for Micro-benchmarks

### 1. Isolate What You're Measuring
- Benchmark one thing at a time
- Use `b.StopTimer()` and `b.StartTimer()` to exclude setup
- Reset state between iterations when necessary

### 2. Use Realistic Data
- Use representative data sizes and patterns
- Include edge cases in separate benchmarks
- Consider cache effects and memory layout

### 3. Control for Variables
- Use fixed seeds for random data
- Run benchmarks multiple times
- Consider system load and other processes

### 4. Measure What Matters
- Include memory allocation metrics with `b.ReportAllocs()`
- Consider both CPU and memory performance
- Benchmark different input sizes to understand scaling

Micro-benchmarks are essential for understanding detailed performance characteristics and validating optimization efforts at the function and algorithm level.
