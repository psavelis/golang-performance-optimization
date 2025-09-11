# Go Runtime Flags

Comprehensive reference of Go runtime flags and environment variables for performance tuning and debugging in production environments.

## Environment Variables

### GOGC - Garbage Collection Target

Controls the aggressiveness of the garbage collector by setting the target percentage of heap growth before triggering a collection.

```bash
# Default value (100%) - GC when heap doubles
export GOGC=100

# More aggressive GC (50%) - GC when heap grows by 50%
export GOGC=50

# Less aggressive GC (200%) - GC when heap triples
export GOGC=200

# Disable GC completely (for debugging only)
export GOGC=off
```

**Example Usage:**
```go
package main

import (
    "runtime"
    "runtime/debug"
    "time"
)

func demonstrateGOGC() {
    // Check current GOGC setting
    fmt.Printf("Current GOGC: %d\n", debug.SetGCPercent(-1))
    
    // Temporarily change GOGC
    oldGOGC := debug.SetGCPercent(50)
    defer debug.SetGCPercent(oldGOGC)
    
    // Allocate memory to trigger GC
    var data [][]byte
    for i := 0; i < 1000; i++ {
        data = append(data, make([]byte, 1024*1024)) // 1MB chunks
    }
}
```

### GOMAXPROCS - Maximum OS Threads

Sets the maximum number of OS threads that can execute Go code simultaneously.

```bash
# Use all available CPU cores (default)
export GOMAXPROCS=0

# Limit to 4 cores
export GOMAXPROCS=4

# Single threaded execution
export GOMAXPROCS=1
```

**Runtime Adjustment:**
```go
package main

import (
    "fmt"
    "runtime"
)

func demonstrateGOMAXPROCS() {
    // Get current GOMAXPROCS
    current := runtime.GOMAXPROCS(0)
    fmt.Printf("Current GOMAXPROCS: %d\n", current)
    
    // Set to half the available cores
    newValue := runtime.NumCPU() / 2
    if newValue < 1 {
        newValue = 1
    }
    
    old := runtime.GOMAXPROCS(newValue)
    fmt.Printf("Changed from %d to %d\n", old, newValue)
    
    // Restore original value
    runtime.GOMAXPROCS(old)
}
```

### GOMEMLIMIT - Memory Limit

Sets a soft memory limit for the Go runtime, helping to prevent OOM kills in containerized environments.

```bash
# Set 1GB memory limit
export GOMEMLIMIT=1GiB

# Set 512MB memory limit
export GOMEMLIMIT=512MiB

# Disable memory limit
export GOMEMLIMIT=off
```

**Programmatic Control:**
```go
package main

import (
    "runtime/debug"
    "fmt"
)

func demonstrateGOMEMLIMIT() {
    // Set memory limit programmatically
    // This returns the previous limit
    oldLimit := debug.SetMemoryLimit(1024 * 1024 * 1024) // 1GB
    
    fmt.Printf("Previous memory limit: %d bytes\n", oldLimit)
    
    // Get current limit
    currentLimit := debug.SetMemoryLimit(-1) // -1 returns current without changing
    fmt.Printf("Current memory limit: %d bytes\n", currentLimit)
}
```

### GOTRACEBACK - Stack Trace Detail

Controls the verbosity of stack traces during panics.

```bash
# Default behavior
export GOTRACEBACK=single

# All goroutines
export GOTRACEBACK=all

# System goroutines too
export GOTRACEBACK=system

# Crash dumps
export GOTRACEBACK=crash

# No stack traces
export GOTRACEBACK=none
```

### GODEBUG - Debug Options

Enables various debugging and profiling options.

```bash
# Enable GC trace information
export GODEBUG=gctrace=1

# Detailed GC information
export GODEBUG=gctrace=2

# Memory allocator debugging
export GODEBUG=allocfreetrace=1

# Schedule trace
export GODEBUG=schedtrace=1000  # Print every 1000ms

# Multiple options
export GODEBUG=gctrace=1,schedtrace=1000,allocfreetrace=1
```

**GODEBUG Options Reference:**

| Option | Description | Example Values |
|--------|-------------|----------------|
| `gctrace` | GC trace information | `1`, `2` |
| `schedtrace` | Scheduler trace (ms interval) | `1000`, `5000` |
| `allocfreetrace` | Memory allocation traces | `1` |
| `cgocheck` | Cgo pointer checking | `0`, `1`, `2` |
| `efence` | Electric fence malloc debugging | `1` |
| `gccheckmark` | GC mark phase checking | `1` |
| `gcpacertrace` | GC pacer trace | `1` |
| `gcrescanstacks` | GC stack rescan debugging | `1` |
| `gcstoptheworld` | Stop-the-world GC debugging | `1`, `2` |
| `madvdontneed` | MADV_DONTNEED behavior | `0`, `1` |
| `memprofilerate` | Memory profiling rate | `1`, `512000` |
| `scavtrace` | Scavenger trace | `1` |

## Runtime Flags

### pprof Integration Flags

```go
package main

import (
    "flag"
    "log"
    "net/http"
    _ "net/http/pprof"
    "runtime"
)

var (
    cpuProfile     = flag.String("cpuprofile", "", "CPU profile output file")
    memProfile     = flag.String("memprofile", "", "Memory profile output file")
    blockProfile   = flag.String("blockprofile", "", "Block profile output file")
    mutexProfile   = flag.String("mutexprofile", "", "Mutex profile output file")
    traceFile      = flag.String("trace", "", "Execution trace output file")
    pprofAddr      = flag.String("pprof", "", "pprof HTTP server address")
)

func main() {
    flag.Parse()
    
    // Enable profiling
    if *pprofAddr != "" {
        go func() {
            log.Printf("Starting pprof server on %s", *pprofAddr)
            log.Println(http.ListenAndServe(*pprofAddr, nil))
        }()
    }
    
    // Set profiling rates
    runtime.SetBlockProfileRate(1)
    runtime.SetMutexProfileFraction(1)
    
    // Your application code here
    runApplication()
}
```

### Performance Tuning Flags

```bash
#!/bin/bash
# Production performance tuning script

# Memory settings
export GOGC=100              # Default GC target
export GOMEMLIMIT=8GiB       # Container memory limit
export GOMAXPROCS=0          # Use all available cores

# Debug settings for production
export GODEBUG=gctrace=1,scavtrace=1

# Application-specific flags
./myapp \
    -pprof=:6060 \
    -cpuprofile=/tmp/cpu.prof \
    -memprofile=/tmp/mem.prof \
    -blockprofile=/tmp/block.prof \
    -mutexprofile=/tmp/mutex.prof \
    -trace=/tmp/trace.out
```

## GC Tuning Examples

### Low Latency Applications

```bash
# Prioritize low latency over throughput
export GOGC=50               # More frequent, shorter GC pauses
export GOMEMLIMIT=4GiB       # Prevent memory pressure
export GODEBUG=gctrace=1     # Monitor GC behavior
```

### High Throughput Applications

```bash
# Prioritize throughput over latency
export GOGC=200              # Less frequent GC, higher throughput
export GOMEMLIMIT=16GiB      # More memory headroom
export GOMAXPROCS=0          # Use all cores
```

### Memory-Constrained Environments

```bash
# Optimize for low memory usage
export GOGC=50               # Aggressive GC
export GOMEMLIMIT=1GiB       # Strict memory limit
export GODEBUG=gctrace=1,scavtrace=1  # Monitor memory behavior
```

## Container Optimization

### Docker Configuration

```bash
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .

# Set runtime environment
ENV GOGC=100
ENV GOMEMLIMIT=1GiB
ENV GOMAXPROCS=0

# Optional debug settings
ENV GODEBUG=gctrace=1

EXPOSE 8080 6060
CMD ["./app", "-pprof=:6060"]
```

### Kubernetes Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-runtime-config
data:
  GOGC: "100"
  GOMEMLIMIT: "2GiB"
  GOMAXPROCS: "0"
  GODEBUG: "gctrace=1"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-app
  template:
    metadata:
      labels:
        app: go-app
    spec:
      containers:
      - name: app
        image: my-go-app:latest
        envFrom:
        - configMapRef:
            name: go-runtime-config
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        ports:
        - containerPort: 8080
        - containerPort: 6060  # pprof
```

## Runtime Monitoring

### GC Trace Analysis

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// Parse GC trace output (from GODEBUG=gctrace=1)
// Example: gc 1 @0.002s 0%: 0.018+0.46+0.040 ms clock, 0.14+0.16/0.38/0.052+0.32 ms cpu, 4->4->2 MB, 5 MB goal, 8 P

type GCTrace struct {
    Cycle     int     // GC cycle number
    Timestamp float64 // Time since program start
    CPUUsage  float64 // CPU percentage used by GC
    WallTime  float64 // Wall clock time
    CPUTime   float64 // CPU time
    HeapSizes string  // Before->After->Live heap sizes
    Goal      string  // GC goal
    Procs     int     // Number of processors
}

func parseGCTrace(line string) (*GCTrace, error) {
    // Regular expression to parse GC trace line
    re := regexp.MustCompile(`gc (\d+) @([\d.]+)s (\d+)%: ([\d.+]+) ms clock, ([\d.+/]+) ms cpu, ([\d->]+) MB, (\d+) MB goal, (\d+) P`)
    
    matches := re.FindStringSubmatch(line)
    if len(matches) != 9 {
        return nil, fmt.Errorf("invalid GC trace format")
    }
    
    cycle, _ := strconv.Atoi(matches[1])
    timestamp, _ := strconv.ParseFloat(matches[2], 64)
    cpuUsage, _ := strconv.ParseFloat(matches[3], 64)
    wallTime, _ := strconv.ParseFloat(strings.Split(matches[4], "+")[0], 64)
    procs, _ := strconv.Atoi(matches[8])
    
    return &GCTrace{
        Cycle:     cycle,
        Timestamp: timestamp,
        CPUUsage:  cpuUsage,
        WallTime:  wallTime,
        HeapSizes: matches[6],
        Goal:      matches[7] + " MB",
        Procs:     procs,
    }, nil
}

func analyzeGCLog(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    var traces []GCTrace
    
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, "gc ") {
            if trace, err := parseGCTrace(line); err == nil {
                traces = append(traces, *trace)
            }
        }
    }
    
    if len(traces) == 0 {
        return fmt.Errorf("no GC traces found")
    }
    
    // Analyze GC patterns
    var totalPause float64
    var maxPause float64
    var gcFrequency float64
    
    for _, trace := range traces {
        totalPause += trace.WallTime
        if trace.WallTime > maxPause {
            maxPause = trace.WallTime
        }
    }
    
    if len(traces) > 1 {
        duration := traces[len(traces)-1].Timestamp - traces[0].Timestamp
        gcFrequency = float64(len(traces)) / duration
    }
    
    fmt.Printf("GC Analysis Summary:\n")
    fmt.Printf("  Total GC cycles: %d\n", len(traces))
    fmt.Printf("  Average pause: %.2f ms\n", totalPause/float64(len(traces)))
    fmt.Printf("  Maximum pause: %.2f ms\n", maxPause)
    fmt.Printf("  GC frequency: %.2f cycles/second\n", gcFrequency)
    
    return nil
}
```

### Schedule Trace Analysis

```bash
# Enable scheduler tracing
export GODEBUG=schedtrace=1000

# Example output:
# SCHED 1000ms: gomaxprocs=8 idleprocs=6 threads=10 spinningthreads=0 idlethreads=4 runqueue=0 [0 0 0 0 0 0 0 0]
```

**Schedule Trace Fields:**
- `gomaxprocs`: GOMAXPROCS setting
- `idleprocs`: Idle processors
- `threads`: OS threads
- `spinningthreads`: Spinning threads looking for work
- `idlethreads`: Idle threads
- `runqueue`: Global runqueue length
- `[...]`: Per-processor runqueue lengths

These runtime flags and environment variables provide fine-grained control over Go's runtime behavior, enabling precise performance tuning for different deployment scenarios and workload characteristics.
