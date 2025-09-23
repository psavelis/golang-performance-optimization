# Goroutine Scheduler

Understanding Go's goroutine scheduler is crucial for optimizing concurrent applications. This chapter explores the scheduler's architecture, scheduling algorithms, and performance optimization techniques.

## Scheduler Architecture

### GMP Model

Go's scheduler implements the **G-M-P model**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Scheduler (GMP)                      │
├─────────────────────────────────────────────────────────────┤
│  G (Goroutines)  │  M (OS Threads)  │  P (Processors)      │
│                  │                  │                      │
│  • User tasks    │  • OS threads    │  • Logical CPUs      │
│  • Lightweight  │  • 1:1 mapping   │  • Local run queues  │
│  • 2KB stack    │  • Expensive     │  • Work stealing     │
└─────────────────────────────────────────────────────────────┘
```

#### Components Breakdown

**G (Goroutines)**
- Lightweight user-space threads
- Initial stack: 2KB (grows up to 1GB)
- Contains execution state and stack

**M (Machine/OS Threads)**
- Operating system threads
- Expensive to create/destroy
- Typically matches number of CPU cores

**P (Processors)**
- Logical processors (not physical CPUs)
- Default: `runtime.GOMAXPROCS(0)` = number of CPU cores
- Each P has a local run queue

### Scheduler Data Structures

```go
// Simplified scheduler structures (actual implementation in runtime)

type g struct {
    stack       stack     // Stack bounds
    stackguard0 uintptr   // Stack overflow guard
    m           *m        // Current M
    sched       gobuf     // Scheduling context
    atomicstatus uint32   // G status
    goid        int64     // Goroutine ID
}

type m struct {
    g0          *g        // System goroutine
    curg        *g        // Current goroutine
    p           puintptr  // Attached P
    nextp       puintptr  // Next P to run
    spinning    bool      // M is spinning for work
}

type p struct {
    m           muintptr  // Attached M
    runqhead    uint32    // Local run queue head
    runqtail    uint32    // Local run queue tail
    runq        [256]guintptr  // Local run queue
    runnext     guintptr  // Next G to run
}
```

## Scheduling Algorithms

### Work Stealing

The scheduler uses work stealing to balance load:

```go
// Work stealing algorithm (simplified)
func schedule() {
    gp := runqget(_p_)  // Try local run queue first
    if gp == nil {
        gp = findrunnable()  // Global queue or steal
    }
    execute(gp)
}

func findrunnable() *g {
    // 1. Check global run queue
    if gp := globrunqget(_p_, 0); gp != nil {
        return gp
    }
    
    // 2. Steal from other P's
    for i := 0; i < 4; i++ {  // Try 4 times
        for enum := stealOrder.start(fastrand()); !enum.done(); enum.next() {
            p2 := &allp[enum.position()]
            if gp := runqsteal(_p_, p2, stealRunNextG); gp != nil {
                return gp
            }
        }
    }
    
    return nil
}
```

### Scheduling Fairness

The scheduler implements fairness mechanisms:

```go
// Global run queue prevents starvation
func globrunqget(p *p, max int32) *g {
    if sched.runqsize == 0 {
        return nil
    }
    
    // Take from global queue fairly
    n := sched.runqsize/gomaxprocs + 1
    if n > sched.runqsize {
        n = sched.runqsize
    }
    if max > 0 && n > max {
        n = max
    }
    
    gp := sched.runq.pop()
    n--
    
    // Move batch to local queue
    for ; n > 0; n-- {
        gp1 := sched.runq.pop()
        runqput(p, gp1, false)
    }
    
    return gp
}
```

## Goroutine States

### State Transitions

```go
// Goroutine states (from runtime/runtime2.go)
const (
    _Gidle     = iota // Allocated but not initialized
    _Grunnable        // On run queue, ready to run
    _Grunning         // Currently running
    _Gsyscall         // In system call
    _Gwaiting         // Waiting (blocked)
    _Gdead            // Unused, available for reuse
    _Gscan            // Being scanned by GC
)

// State transition examples
func goroutineStates() {
    // _Gidle -> _Grunnable: go func() { ... }()
    go func() {
        // _Grunnable -> _Grunning: scheduled to run
        
        // _Grunning -> _Gwaiting: channel operation
        ch := make(chan int)
        <-ch  // Blocks, enters _Gwaiting
        
        // _Gwaiting -> _Grunnable: channel receives data
        
        // _Grunning -> _Gsyscall: system call
        time.Sleep(time.Millisecond)
        
        // _Gsyscall -> _Grunning: returns from syscall
        
    }() // _Grunning -> _Gdead: function returns
}
```

### Preemption

Go implements cooperative and preemptive scheduling:

```go
// Cooperative preemption points
func cooperativePreemption() {
    for i := 0; i < 1000000; i++ {
        // Function calls are preemption points
        someFunction()
        
        // Channel operations are preemption points
        select {
        case <-time.After(time.Nanosecond):
        default:
        }
        
        // Garbage collection can trigger preemption
        if i%10000 == 0 {
            runtime.Gosched()  // Voluntary yield
        }
    }
}

// Preemptive scheduling (Go 1.14+)
func preemptiveScheduling() {
    // Signal-based preemption for tight loops
    for {
        // This loop can now be preempted even without
        // function calls or channel operations
        computation()
    }
}
```

## Performance Optimization

### GOMAXPROCS Tuning

```go
import (
    "runtime"
    "sync"
    "time"
)

// Optimal GOMAXPROCS depends on workload
func tuneGOMAXPROCS() {
    numCPU := runtime.NumCPU()
    
    // CPU-bound workloads: typically NumCPU
    runtime.GOMAXPROCS(numCPU)
    
    // I/O-bound workloads: may benefit from more
    // runtime.GOMAXPROCS(numCPU * 2)
    
    // Network services: often NumCPU works well
    // runtime.GOMAXPROCS(numCPU)
}

// Benchmark different GOMAXPROCS values
func benchmarkGOMAXPROCS(work func()) {
    for procs := 1; procs <= runtime.NumCPU()*2; procs++ {
        runtime.GOMAXPROCS(procs)
        
        start := time.Now()
        work()
        duration := time.Since(start)
        
        fmt.Printf("GOMAXPROCS=%d: %v\n", procs, duration)
    }
}
```

### Goroutine Pool Patterns

Implement goroutine pools for better resource management:

```go
// Worker pool implementation
type WorkerPool struct {
    workerCount int
    jobQueue    chan Job
    workers     []Worker
    quit        chan bool
}

type Job struct {
    ID   int
    Data interface{}
}

type Worker struct {
    id          int
    jobQueue    chan Job
    workerPool  chan chan Job
    quit        chan bool
}

func NewWorkerPool(workerCount, jobQueueSize int) *WorkerPool {
    pool := &WorkerPool{
        workerCount: workerCount,
        jobQueue:    make(chan Job, jobQueueSize),
        workers:     make([]Worker, workerCount),
        quit:        make(chan bool),
    }
    
    // Start workers
    for i := 0; i < workerCount; i++ {
        worker := Worker{
            id:         i,
            jobQueue:   make(chan Job),
            workerPool: make(chan chan Job),
            quit:       make(chan bool),
        }
        pool.workers[i] = worker
        worker.Start()
    }
    
    // Start dispatcher
    go pool.dispatch()
    
    return pool
}

func (w *Worker) Start() {
    go func() {
        for {
            // Register worker in pool
            w.workerPool <- w.jobQueue
            
            select {
            case job := <-w.jobQueue:
                // Process job
                processJob(job)
                
            case <-w.quit:
                return
            }
        }
    }()
}

func (p *WorkerPool) dispatch() {
    for {
        select {
        case job := <-p.jobQueue:
            // Get available worker
            go func(job Job) {
                worker := <-p.workers[0].workerPool
                worker <- job
            }(job)
            
        case <-p.quit:
            return
        }
    }
}
```

### Goroutine Lifecycle Management

```go
// Graceful goroutine shutdown
type GoroutineManager struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func NewGoroutineManager() *GoroutineManager {
    ctx, cancel := context.WithCancel(context.Background())
    return &GoroutineManager{
        ctx:    ctx,
        cancel: cancel,
    }
}

func (gm *GoroutineManager) Start(fn func(context.Context)) {
    gm.wg.Add(1)
    go func() {
        defer gm.wg.Done()
        fn(gm.ctx)
    }()
}

func (gm *GoroutineManager) Stop() {
    gm.cancel()
    gm.wg.Wait()
}

// Usage example
func gracefulShutdown() {
    gm := NewGoroutineManager()
    
    // Start multiple workers
    for i := 0; i < 10; i++ {
        gm.Start(func(ctx context.Context) {
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    doWork()
                }
            }
        })
    }
    
    // Graceful shutdown on signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    gm.Stop()
}
```

## Scheduler Debugging

### Runtime Scheduler Traces

```go
import (
    "os"
    "runtime/trace"
)

// Enable scheduler tracing
func enableSchedulerTrace() {
    f, err := os.Create("trace.out")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    err = trace.Start(f)
    if err != nil {
        panic(err)
    }
    defer trace.Stop()
    
    // Your application code here
    runApplication()
}

// Analyze with: go tool trace trace.out
```

### Scheduler Statistics

```go
import "runtime"

// Monitor scheduler performance
func schedulerStats() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
    fmt.Printf("OS Threads: %d\n", runtime.GOMAXPROCS(0))
    fmt.Printf("CGO Calls: %d\n", runtime.NumCgoCall())
    
    // Scheduler-specific stats
    fmt.Printf("Heap Objects: %d\n", stats.HeapObjects)
    fmt.Printf("Stack Inuse: %d bytes\n", stats.StackInuse)
    fmt.Printf("Goroutine Stack: %d bytes\n", stats.StackSys)
}

// Monitor goroutine leaks
func monitorGoroutines() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            count := runtime.NumGoroutine()
            if count > 1000 {  // Threshold
                fmt.Printf("WARNING: High goroutine count: %d\n", count)
                
                // Get stack traces
                buf := make([]byte, 1024*1024)
                n := runtime.Stack(buf, true)
                fmt.Printf("Stack traces:\n%s\n", buf[:n])
            }
        }
    }
}
```

### Scheduler-Aware Programming

```go
// Avoid blocking all threads
func avoidBlocking() {
    // BAD: Can block all OS threads
    for i := 0; i < runtime.GOMAXPROCS(0); i++ {
        go func() {
            // Blocking system call without runtime.LockOSThread()
            cgoBlockingCall()
        }()
    }
    
    // GOOD: Limit blocking operations
    semaphore := make(chan struct{}, runtime.GOMAXPROCS(0)/2)
    
    for i := 0; i < 100; i++ {
        go func() {
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Blocking operation with limited concurrency
            cgoBlockingCall()
        }()
    }
}

// CPU-intensive work cooperation
func cpuIntensiveWork() {
    for i := 0; i < 1000000; i++ {
        // Yield occasionally for other goroutines
        if i%10000 == 0 {
            runtime.Gosched()
        }
        
        // CPU-intensive computation
        complexCalculation()
    }
}

// I/O-bound optimization
func ioBoundOptimization() {
    // Use buffered channels to reduce scheduling overhead
    results := make(chan Result, 100)
    
    // Batch I/O operations
    go func() {
        batch := make([]Request, 0, 10)
        
        for req := range requests {
            batch = append(batch, req)
            
            if len(batch) == 10 {
                processBatch(batch, results)
                batch = batch[:0]
            }
        }
        
        // Process remaining
        if len(batch) > 0 {
            processBatch(batch, results)
        }
    }()
}
```

## Scheduler Anti-Patterns

### ❌ **Common Mistakes**

1. **Creating too many goroutines**
   ```go
   // DON'T: Unbounded goroutine creation
   for _, item := range millionItems {
       go processItem(item)
   }
   
   // DO: Use worker pools
   pool := NewWorkerPool(runtime.NumCPU(), 1000)
   for _, item := range millionItems {
       pool.Submit(item)
   }
   ```

2. **Blocking all threads**
   ```go
   // DON'T: Block all OS threads
   for i := 0; i < runtime.GOMAXPROCS(0); i++ {
       go func() {
           blockingSystemCall()  // Blocks thread
       }()
   }
   
   // DO: Limit blocking operations
   sem := make(chan struct{}, runtime.GOMAXPROCS(0)/2)
   ```

3. **Goroutine leaks**
   ```go
   // DON'T: Goroutines that never exit
   go func() {
       for {
           select {
           case data := <-ch:
               process(data)
               // No exit condition
           }
       }
   }()
   
   // DO: Always provide exit mechanism
   go func() {
       for {
           select {
           case data := <-ch:
               process(data)
           case <-ctx.Done():
               return
           }
       }
   }()
   ```

## Performance Best Practices

### ✅ **Scheduler Optimization Guidelines**

1. **Right-size GOMAXPROCS**
   ```go
   // For CPU-bound: GOMAXPROCS = NumCPU
   // For I/O-bound: May need higher values
   runtime.GOMAXPROCS(runtime.NumCPU())
   ```

2. **Use goroutine pools for predictable loads**
   ```go
   pool := NewWorkerPool(optimalWorkerCount, queueSize)
   ```

3. **Minimize goroutine creation overhead**
   ```go
   // Reuse goroutines instead of creating new ones
   ```

4. **Avoid unnecessary context switches**
   ```go
   // Batch work to reduce scheduling overhead
   ```

5. **Monitor and profile scheduler behavior**
   ```go
   go tool trace trace.out
   ```

### Performance Measurement

```go
// Benchmark scheduler overhead
func BenchmarkSchedulerOverhead(b *testing.B) {
    b.Run("DirectCall", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            doWork()
        }
    })
    
    b.Run("Goroutine", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            done := make(chan bool)
            go func() {
                doWork()
                done <- true
            }()
            <-done
        }
    })
    
    b.Run("WorkerPool", func(b *testing.B) {
        pool := NewWorkerPool(runtime.NumCPU(), 1000)
        defer pool.Close()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            pool.Submit(workItem{})
        }
    })
}
```

## Advanced Scheduler Topics

### Custom Schedulers

```go
// Implementing custom scheduling for specific workloads
type CustomScheduler struct {
    queues      []chan Task
    workers     []Worker
    loadBalance LoadBalancer
}

type LoadBalancer interface {
    SelectQueue(task Task) int
}

// Priority-based load balancing
type PriorityLoadBalancer struct {
    priorities []int
}

func (p *PriorityLoadBalancer) SelectQueue(task Task) int {
    // Route high-priority tasks to dedicated queues
    if task.Priority > 5 {
        return 0  // High-priority queue
    }
    return task.ID % (len(p.priorities) - 1) + 1
}
```

### Scheduler Integration with GC

```go
// Coordinate with garbage collector
func gcAwareScheduling() {
    // Monitor GC pressure
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    gcPressure := float64(m.PauseTotalNs) / float64(time.Now().UnixNano())
    
    if gcPressure > 0.01 {  // 1% of time in GC
        // Reduce allocation-heavy goroutines
        runtime.GOMAXPROCS(runtime.NumCPU() / 2)
    } else {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }
}
```

Understanding the goroutine scheduler enables you to write highly concurrent, efficient Go applications that scale well across multiple CPU cores while avoiding common scheduling pitfalls.

---

**Next**: [Garbage Collector](garbage-collector.md) - Learn about Go's garbage collection and memory management
