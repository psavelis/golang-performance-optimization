# Your First Profile

Now that your environment is set up, let's generate and analyze your first Go profiles. This hands-on tutorial will guide you through the complete profiling workflow.

## Sample Application

We'll use a realistic application that demonstrates common performance patterns. Create this sample program:

```go
// main.go - Sample application for profiling
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Event represents a business event
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Data      string    `json:"data"`
}

func main() {
	// Start profiling server
	go func() {
		log.Println("🔍 Profiling server at http://localhost:6060/debug/pprof/")
		log.Fatal(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("🚀 Starting performance demo...")
	
	// Demonstrate different performance patterns
	demonstrateProfilingTargets()
	
	fmt.Println("✅ Demo complete. Check profiles at http://localhost:6060/debug/pprof/")
	
	// Keep server running for profiling
	select {}
}

func demonstrateProfilingTargets() {
	// CPU-intensive operations
	fmt.Println("📊 Running CPU-intensive operations...")
	runCPUIntensiveOperations()
	
	// Memory allocation patterns
	fmt.Println("💾 Demonstrating memory allocation patterns...")
	runMemoryIntensiveOperations()
	
	// Concurrent operations
	fmt.Println("🔄 Running concurrent operations...")
	runConcurrentOperations()
	
	// Blocking operations
	fmt.Println("⏸️  Demonstrating blocking operations...")
	runBlockingOperations()
}

// CPU-intensive operations
func runCPUIntensiveOperations() {
	start := time.Now()
	
	// String manipulation (CPU intensive)
	for i := 0; i < 100000; i++ {
		_ = generateRandomString(20)
	}
	
	// Mathematical operations
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += fibonacci(20)
	}
	
	fmt.Printf("   CPU operations completed in %v\n", time.Since(start))
}

// Memory allocation operations
func runMemoryIntensiveOperations() {
	start := time.Now()
	
	// Large slice allocations
	events := make([]*Event, 0, 50000)
	
	for i := 0; i < 50000; i++ {
		event := &Event{
			ID:        generateRandomString(10),
			Type:      "user_action",
			Timestamp: time.Now(),
			UserID:    generateRandomString(8),
			Data:      generateRandomString(100),
		}
		events = append(events, event)
	}
	
	// JSON serialization (allocation intensive)
	for i := 0; i < 1000; i++ {
		_, _ = json.Marshal(events[i*10:(i+1)*10])
	}
	
	fmt.Printf("   Memory operations completed in %v\n", time.Since(start))
	runtime.GC() // Force garbage collection
}

// Concurrent operations
func runConcurrentOperations() {
	start := time.Now()
	
	var wg sync.WaitGroup
	ch := make(chan string, 100)
	
	// Producer goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				ch <- fmt.Sprintf("message-%d-%d", id, j)
			}
		}(i)
	}
	
	// Consumer goroutine
	go func() {
		processed := 0
		for range ch {
			processed++
			if processed >= 10000 {
				break
			}
		}
	}()
	
	wg.Wait()
	close(ch)
	
	fmt.Printf("   Concurrent operations completed in %v\n", time.Since(start))
}

// Blocking operations
func runBlockingOperations() {
	start := time.Now()
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	shared := 0
	
	// Mutex contention simulation
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock()
				shared++
				time.Sleep(time.Microsecond) // Simulate work under lock
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	fmt.Printf("   Blocking operations completed in %v (shared: %d)\n", 
		time.Since(start), shared)
}

// Helper functions
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(length)
	
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
```

Create a `go.mod` file:

```go
// go.mod
module profiling-demo

go 1.24
```

## Generating Profiles

### Method 1: Live Profiling with pprof Endpoint

Start the application:

```bash
go run main.go
```

In another terminal, generate profiles:

```bash
# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Mutex profile (if enabled)
go tool pprof http://localhost:6060/debug/pprof/mutex

# Block profile (if enabled)
go tool pprof http://localhost:6060/debug/pprof/block
```

### Method 2: Programmatic Profiling

Add profiling directly to your code:

```go
// Add to imports
import (
	"os"
	"runtime/pprof"
)

// Add profiling functions
func startCPUProfile() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()
	
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	
	// Your application code here
	demonstrateProfilingTargets()
}

func writeMemProfile() {
	f, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()
	
	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
```

### Method 3: Benchmark Profiling

Create a benchmark file:

```go
// main_test.go
package main

import (
	"testing"
)

func BenchmarkCPUIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runCPUIntensiveOperations()
	}
}

func BenchmarkMemoryIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runMemoryIntensiveOperations()
	}
}

func BenchmarkConcurrent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runConcurrentOperations()
	}
}

func BenchmarkBlocking(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runBlockingOperations()
	}
}
```

Run benchmarks with profiling:

```bash
# Generate profiles from benchmarks
go test -bench=. -benchmem \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  -blockprofile=block.prof \
  -mutexprofile=mutex.prof
```

## Analyzing Your First Profiles

### CPU Profile Analysis

Analyze the CPU profile:

```bash
# Interactive analysis
go tool pprof cpu.prof

# Commands within pprof:
# (pprof) top10          # Show top 10 CPU consumers
# (pprof) list main.generateRandomString  # Show source code
# (pprof) web            # Generate SVG visualization
# (pprof) quit           # Exit

# Web interface (recommended)
go tool pprof -http=:8080 cpu.prof
```

**What to Look For:**

1. **Top Functions**: Functions consuming the most CPU time
2. **Call Stack**: How functions call each other
3. **Self vs Cumulative**: Direct vs. indirect CPU usage
4. **Hotspots**: Concentration of CPU usage in specific areas

### Memory Profile Analysis

Analyze memory allocation:

```bash
# Heap analysis
go tool pprof -http=:8080 mem.prof

# Command-line analysis
go tool pprof mem.prof
# (pprof) top10
# (pprof) list main.generateRandomString
# (pprof) svg
```

**Key Metrics:**

- **Alloc Space**: Total bytes allocated
- **Alloc Objects**: Total objects allocated  
- **In-Use Space**: Current memory usage
- **In-Use Objects**: Current object count

### Understanding Profile Output

#### CPU Profile Example Output

```
(pprof) top10
Showing nodes accounting for 2.84s, 94.67% of 3.00s total
Dropped 45 nodes (cum <= 0.01s)
Showing top 10 nodes out of 25
      flat  flat%   sum%        cum   cum%
     1.20s 40.00% 40.00%      1.20s 40.00%  main.fibonacci
     0.89s 29.67% 69.67%      0.89s 29.67%  main.generateRandomString
     0.35s 11.67% 81.33%      0.35s 11.67%  runtime.mallocgc
     0.21s  7.00% 88.33%      0.25s  8.33%  strings.(*Builder).Grow
     0.19s  6.33% 94.67%      0.19s  6.33%  runtime.memclrNoHeapPointers
```

**Understanding the Columns:**
- **flat**: Time spent in this function only
- **flat%**: Percentage of total time
- **sum%**: Cumulative percentage
- **cum**: Cumulative time (including called functions)
- **cum%**: Cumulative percentage

#### Memory Profile Example Output

```
(pprof) top10
Showing nodes accounting for 156.73MB, 98.12% of 159.73MB total
Dropped 12 nodes (cum <= 0.80MB)
      flat  flat%   sum%        cum   cum%
   89.23MB 55.87% 55.87%    89.23MB 55.87%  main.generateRandomString
   45.12MB 28.25% 84.12%    45.12MB 28.25%  main.(*Event)
   22.38MB 14.01% 98.12%    22.38MB 14.01%  encoding/json.Marshal
```

## Visual Analysis with Web Interface

The web interface (`-http=:8080`) provides powerful visualizations:

### Graph View
- **Nodes**: Functions (size = resource usage)
- **Edges**: Call relationships (thickness = usage)
- **Colors**: Heat map of resource consumption

### Flamegraph View
- **Width**: Time spent in function
- **Height**: Call stack depth
- **Color**: Different functions/packages

### Source View
- **Line-by-line analysis** of hot functions
- **Annotation** with resource usage per line
- **Context** showing surrounding code

## Common Patterns in First Profiles

### Expected Patterns

1. **String Operations**: Often show up as CPU hotspots
2. **JSON Serialization**: Memory allocation intensive
3. **Goroutine Creation**: May show in goroutine profiles
4. **Mutex Contention**: Visible in blocking profiles

### Red Flags to Look For

1. **Excessive Allocations**: High allocation rates
2. **Deep Call Stacks**: Recursive functions
3. **Blocking Operations**: Long-held locks
4. **Memory Leaks**: Growing heap usage

## Practice Exercises

### Exercise 1: Identify the Hotspot
Run the sample application and identify:
1. Which function consumes the most CPU?
2. Which function allocates the most memory?
3. How many goroutines are running?

### Exercise 2: Compare Before/After
1. Profile the original application
2. Optimize one function (hint: `generateRandomString`)
3. Profile again and compare results

### Exercise 3: Enable Mutex Profiling
Add this to enable mutex profiling:

```go
import "runtime"

func init() {
    runtime.SetMutexProfileFraction(1)
    runtime.SetBlockProfileRate(1)
}
```

## Next Steps

Congratulations! You've generated and analyzed your first Go profiles. You should now be able to:

✅ Generate CPU, memory, goroutine, and blocking profiles  
✅ Use both web and command-line interfaces  
✅ Identify performance hotspots  
✅ Understand basic profile output  

Ready to dive deeper? Continue to [Understanding Output](understanding-output.md) to master profile interpretation and analysis techniques.

## Quick Reference

### Essential Commands

```bash
# Live profiling
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Benchmark profiling  
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

# Analysis
go tool pprof -http=:8080 cpu.prof
go tool pprof -http=:8080 mem.prof

# Common pprof commands
# top10, list <function>, web, svg, png, pdf
```

### Profile Types Quick Guide

| Profile Type | What It Shows | When to Use |
|-------------|---------------|-------------|
| **CPU** | Function execution time | Performance optimization |
| **Heap** | Memory allocation | Memory usage analysis |
| **Goroutine** | Goroutine states | Concurrency debugging |
| **Mutex** | Lock contention | Synchronization issues |
| **Block** | Blocking operations | I/O and sync bottlenecks |

Your profiling journey has just begun! 🚀
