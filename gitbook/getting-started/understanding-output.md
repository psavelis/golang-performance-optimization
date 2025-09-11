# Understanding Profile Output

Learning to interpret profiling data is crucial for effective performance optimization. This chapter teaches you how to read, analyze, and extract actionable insights from Go profiles.

## Profile Data Fundamentals

### Types of Profile Information

Go profiles contain different types of performance data:

```bash
# CPU Profile shows:
# - Function execution time
# - Call stack relationships  
# - Sampling frequency data
# - Instruction-level details

# Memory Profile shows:
# - Allocation patterns
# - Object counts and sizes
# - Call stacks for allocations
# - Memory usage over time
```

### Reading Profile Headers

Understanding profile metadata helps contextualize the data:

```
(pprof) top
File: your-app
Type: cpu
Time: May 15, 2024 at 3:45pm (PDT)
Duration: 30.13s, Total samples = 28450ms (94.46%)
Entering interactive mode (type "help" for commands, "o" for options)
```

**Key Information:**
- **File**: Binary that was profiled
- **Type**: Profile type (cpu, heap, goroutine, etc.)
- **Duration**: How long profiling ran
- **Total samples**: Actual data collected
- **Sample percentage**: Coverage of execution time

## CPU Profile Analysis

### Understanding the Top Output

```bash
(pprof) top10
Showing nodes accounting for 25.84s, 90.87% of 28.45s total
Dropped 45 nodes (cum <= 0.14s)
Showing top 10 nodes out of 127
      flat  flat%   sum%        cum   cum%
     8.45s 29.70% 29.70%      8.45s 29.70%  crypto/sha256.block
     4.12s 14.48% 44.18%      4.12s 14.48%  main.processData
     3.89s 13.67% 57.85%      5.23s 18.38%  encoding/json.Marshal
     2.34s  8.23% 66.08%      2.34s  8.23%  runtime.mallocgc
     1.56s  5.48% 71.56%      1.89s  6.64%  strings.(*Builder).Grow
     1.23s  4.32% 75.88%      1.23s  4.32%  runtime.memmove
     0.98s  3.45% 79.33%      0.98s  3.45%  syscall.Syscall
     0.87s  3.06% 82.39%      2.45s  8.61%  main.(*Parser).Parse
     0.76s  2.67% 85.06%      0.76s  2.67%  runtime.nextFreeFast
     0.64s  2.25% 87.31%      0.64s  2.25%  hash/crc32.update
```

### Column Interpretation

| Column | Meaning | When to Focus |
|--------|---------|---------------|
| **flat** | Time spent only in this function | High values indicate direct bottlenecks |
| **flat%** | Percentage of total time | Functions >5% are optimization candidates |
| **sum%** | Cumulative percentage | Shows data completeness |
| **cum** | Time including called functions | High values indicate important call paths |
| **cum%** | Cumulative percentage | Helps identify system-level bottlenecks |

### Identifying Optimization Opportunities

#### 🔴 **High Flat Time** - Direct Optimization Targets
```bash
# Example: Function consuming 29.7% of CPU
8.45s 29.70% crypto/sha256.block

# Action: Optimize this specific function
# - Algorithm improvements
# - Implementation efficiency  
# - Caching strategies
```

#### 🟡 **High Cumulative Time** - System Bottlenecks
```bash
# Example: JSON marshaling with high cumulative time
3.89s 13.67% 57.85% 5.23s 18.38% encoding/json.Marshal

# Analysis: Function uses 13.67% directly, but 18.38% including calls
# Action: Investigate what's calling Marshal and why
```

#### 🟢 **Runtime Functions** - Infrastructure Overhead
```bash
# Example: Memory allocation overhead
2.34s 8.23% runtime.mallocgc

# Analysis: High allocation overhead
# Action: Reduce allocations in calling code
```

## Source Code Analysis

### Function-Level Analysis

```bash
(pprof) list main.processData
Total: 28.45s
ROUTINE ======================== main.processData in /app/main.go
     4.12s      4.12s (flat, cum) 14.48% of Total
         .          .     45:func processData(data []byte) Result {
         .          .     46:	var result Result
      2.1s       2.1s     47:	items := strings.Split(string(data), ",")  // 🚨 Expensive
         .          .     48:	for _, item := range items {
     1.2s       1.2s     49:		processed := heavyComputation(item)      // 🚨 Expensive  
     0.82s      0.82s     50:		result.Items = append(result.Items, processed)
         .          .     51:	}
         .          .     52:	return result
         .          .     53:}
```

**Analysis Insights:**
- **Line 47**: `strings.Split(string(data), ",")` - Converting []byte to string causes allocation
- **Line 49**: `heavyComputation` - Called in loop, potential optimization target
- **Line 50**: `append` operations - May cause slice reallocations

### Assembly-Level Analysis

```bash
(pprof) disasm main.processData
Total: 28.45s
ROUTINE ======================== main.processData
     4.12s      4.12s (flat, cum) 14.48% of Total
         .          .    488690: MOVQ 0x18(SP), AX        
     1.2s       1.2s    488698: CALL runtime.convT2Estring    // String conversion
         .          .    48869d: MOVQ AX, 0x20(SP)        
     2.1s       2.1s    4886a5: CALL strings.Split            // Split operation
         .          .    4886aa: MOVQ 0x28(SP), CX
```

**When to Use Assembly Analysis:**
- Micro-optimization of critical paths
- Understanding compiler optimization decisions
- Investigating unexpected performance characteristics

## Memory Profile Analysis

### Heap Profile Interpretation

```bash
(pprof) top heap.prof
File: your-app
Type: inuse_space
Time: May 15, 2024 at 3:45pm (PDT)
Showing nodes accounting for 45.23MB, 89.12% of 50.78MB total
      flat  flat%   sum%        cum   cum%
   15.67MB 30.86% 30.86%    15.67MB 30.86%  main.loadData
    8.34MB 16.43% 47.29%     8.34MB 16.43%  encoding/json.Unmarshal
    6.78MB 13.35% 60.64%     6.78MB 13.35%  strings.Join
    4.89MB  9.63% 70.27%     4.89MB  9.63%  make([]string, 1000)
    3.45MB  6.79% 77.06%     3.45MB  6.79%  regexp.Compile
```

### Memory Profile Types

#### inuse_space vs alloc_space
```bash
# Current memory usage (inuse_space)
go tool pprof http://localhost:6060/debug/pprof/heap

# Total allocations since start (alloc_space)
go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap
```

#### inuse_objects vs alloc_objects
```bash
# Current object count
go tool pprof -inuse_objects http://localhost:6060/debug/pprof/heap

# Total objects allocated
go tool pprof -alloc_objects http://localhost:6060/debug/pprof/heap
```

### Memory Analysis Patterns

#### 🔴 **Memory Leaks** - Growing inuse_space
```bash
# Pattern: inuse_space continuously growing
# Symptoms: 
# - High inuse_space values
# - Growing over time
# - Not released after operations complete

# Investigation:
(pprof) list suspiciousFunction
# Look for:
# - Global variables holding references
# - Goroutines not terminating
# - Caches without expiration
```

#### 🟡 **High Allocation Rate** - High alloc_space
```bash
# Pattern: High alloc_space, normal inuse_space
# Symptoms:
# - High GC pressure
# - CPU spent in runtime.mallocgc
# - Frequent garbage collection

# Investigation:
# Focus on reducing allocation frequency
```

#### 🟢 **Large Objects** - High inuse_objects
```bash
# Pattern: Few large allocations
# Symptoms:
# - High inuse_space per object
# - Large individual allocations

# Investigation:
# Consider streaming or chunking strategies
```

## Goroutine Profile Analysis

### Goroutine States

```bash
(pprof) top goroutine.prof
File: your-app
Type: goroutine
Time: May 15, 2024 at 3:45pm (PDT)
Showing nodes accounting for 1247 goroutines, 100% of 1247 total
      flat  flat%   sum%        cum   cum%
       456 36.57% 36.57%        456 36.57%  runtime.gopark
       234 18.78% 55.35%        234 18.78%  net/http.(*conn).serve
       178 14.28% 69.63%        178 14.28%  main.worker
       123  9.86% 79.49%        123  9.86%  runtime.chanrecv
        89  7.14% 86.63%         89  7.14%  sync.(*WaitGroup).Wait
```

### Common Goroutine Patterns

#### ✅ **Healthy Patterns**
```go
// Worker pool with bounded goroutines
func healthyWorkerPool() {
    const numWorkers = 8
    jobs := make(chan Job, 100)
    
    for i := 0; i < numWorkers; i++ {
        go func() {
            for job := range jobs {
                processJob(job)
            }
        }()
    }
}
```

#### 🚨 **Problematic Patterns**
```go
// Goroutine leak - no termination condition
func leakyGoroutine() {
    go func() {
        for {
            // This goroutine never terminates!
            doWork()
            time.Sleep(1 * time.Second)
        }
    }()
}

// Unbounded goroutine creation
func unboundedGoroutines(requests []Request) {
    for _, req := range requests {
        go processRequest(req) // Creates one goroutine per request!
    }
}
```

### Goroutine Analysis Commands

```bash
# View goroutine details
(pprof) list runtime.gopark

# Show call stacks for blocked goroutines
(pprof) peek runtime.gopark

# Web visualization
(pprof) web

# Focus on specific function
(pprof) focus main.worker
```

## Blocking Profile Analysis

### Understanding Blocking Events

```bash
(pprof) top block.prof
File: your-app
Type: delay
Time: May 15, 2024 at 3:45pm (PDT)
Showing nodes accounting for 12.45s, 87.23% of 14.27s total
      flat  flat%   sum%        cum   cum%
     4.67s 32.73% 32.73%      4.67s 32.73%  sync.(*Mutex).Lock
     3.45s 24.18% 56.91%      3.45s 24.18%  runtime.chanrecv
     2.34s 16.40% 73.31%      2.34s 16.40%  runtime.chansend
     1.23s  8.62% 81.93%      1.23s  8.62%  sync.(*RWMutex).RLock
     0.76s  5.33% 87.26%      0.76s  5.33%  runtime.selectgo
```

### Blocking Analysis Patterns

#### 🔴 **Mutex Contention**
```bash
# High sync.(*Mutex).Lock times indicate:
# - Lock held too long
# - Too many goroutines competing
# - Critical section too large

# Solutions:
# - Reduce critical section size
# - Use RWMutex for read-heavy workloads
# - Consider lock-free alternatives
```

#### 🟡 **Channel Blocking**
```bash
# High runtime.chanrecv/chansend times indicate:
# - Channel capacity too small
# - Slow consumers
# - Unbuffered channels in high-throughput scenarios

# Solutions:
# - Increase channel buffer size
# - Add more consumers
# - Use select with default case
```

## Differential Analysis

### Comparing Before/After Profiles

```bash
# Collect baseline
go tool pprof -output=before.prof http://localhost:6060/debug/pprof/profile

# Make optimizations

# Collect optimized profile
go tool pprof -output=after.prof http://localhost:6060/debug/pprof/profile

# Compare
go tool pprof -base=before.prof after.prof
```

### Interpreting Differential Results

```bash
(pprof) top
File: after.prof
Type: cpu
Time: May 15, 2024 at 4:15pm (PDT)
Duration: 30s, Total samples = 15.67s (52.23%)
Showing nodes accounting for -12.78s, -81.57% of 15.67s total
      flat  flat%   sum%        cum   cum%
    -8.45s -53.92% -53.92%     -8.45s -53.92%  main.slowFunction ✅
     2.34s  14.93% -38.99%      2.34s  14.93%  main.newOptimization
    -1.23s  -7.85% -46.84%     -1.23s  -7.85%  strings.Join ✅
     0.89s   5.68% -41.16%      0.89s   5.68%  runtime.mallocgc
```

**Interpretation:**
- **Negative values**: Time reduced (good!)
- **Positive values**: New or increased time
- **Focus**: Verify optimizations worked and no new bottlenecks appeared

## Common Profile Interpretation Mistakes

### ❌ **Anti-patterns to Avoid**

#### 1. **Optimizing Low-Impact Functions**
```bash
# Don't optimize functions with <5% flat time
0.01s  0.04%  someFunction  # Not worth optimizing
```

#### 2. **Ignoring Runtime Overhead**
```bash
# High runtime.mallocgc means allocation problem
2.34s  8.23%  runtime.mallocgc  # Fix calling code, not runtime
```

#### 3. **Misreading Cumulative vs Flat**
```bash
# High cum%, low flat% = optimize called functions
     .15s  0.5%   15.6%  main.orchestrator  # Don't optimize this
    8.45s 29.7%   29.7%  main.actualWork    # Optimize this instead
```

#### 4. **Single Sample Analysis**
```bash
# Always collect multiple samples for statistical significance
go test -bench=. -count=5 -cpuprofile=cpu.prof
```

## Actionable Analysis Workflow

### 1. **Initial Assessment**
```bash
# Get overview
(pprof) top10

# Identify major bottlenecks (>5% flat time)
# Check for runtime overhead patterns
# Look for obvious algorithmic issues
```

### 2. **Deep Dive Analysis**
```bash
# Examine high-impact functions
(pprof) list functionName

# Check call relationships
(pprof) peek functionName

# Generate visual analysis
(pprof) web
```

### 3. **Hypothesis Formation**
```bash
# Based on profile data, form specific hypotheses:
# - Algorithm inefficiency?
# - Excessive allocations?
# - I/O bottlenecks?
# - Concurrency issues?
```

### 4. **Targeted Optimization**
```bash
# Apply specific optimizations based on evidence
# Measure impact with differential profiling
# Verify no new bottlenecks introduced
```

### 5. **Validation**
```bash
# Collect new profiles
# Compare before/after
# Validate in production-like environment
```

## Profile Interpretation Cheat Sheet

### Quick Reference

| Pattern | Indication | Action |
|---------|------------|--------|
| High `strings.Join` | String concatenation | Use `strings.Builder` |
| High `json.Marshal` | JSON overhead | Consider alternatives |
| High `runtime.mallocgc` | Allocation pressure | Reduce allocations |
| High `sync.(*Mutex).Lock` | Lock contention | Optimize critical sections |
| High `runtime.chanrecv` | Channel blocking | Increase buffer/consumers |
| High `crypto/*` | Cryptographic overhead | Cache/optimize crypto operations |
| High `regexp.*` | Regex compilation | Compile once, reuse |
| High `syscall.*` | System call overhead | Batch operations |

## Advanced Analysis Techniques

### Focus and Ignore Patterns

```bash
# Focus on specific package
(pprof) focus main.*

# Ignore runtime functions
(pprof) ignore runtime.*

# Focus on custom code only
(pprof) focus -ignore runtime

# Complex patterns
(pprof) focus "main\.(process|handle).*" -ignore "test.*"
```

### Tag-Based Analysis

```bash
# Filter by tags (if using tagged profiles)
(pprof) tagfocus service:api
(pprof) tagignore env:test
```

### Sample Filtering

```bash
# Show only samples above threshold
(pprof) sample_index=cpu:samples:count
(pprof) show_from=1000
```

By mastering profile interpretation, you can quickly identify performance bottlenecks and make data-driven optimization decisions. The key is to focus on high-impact opportunities and validate every optimization with measurement.

---

**Next**: [Performance Fundamentals](../fundamentals/README.md) - Learn the Go runtime internals that drive performance characteristics
