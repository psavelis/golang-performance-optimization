# Parallel Benchmarks

Master parallel benchmarking techniques to accurately measure the performance characteristics of concurrent Go applications and identify scaling bottlenecks.

## Understanding Parallel Benchmarks

Parallel benchmarks test how your code performs under concurrent load, revealing scaling characteristics, contention issues, and race conditions that sequential benchmarks might miss.

## Basic Parallel Benchmark Structure

```go
func BenchmarkParallelExample(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // Code to benchmark in parallel
            result := expensiveOperation()
            _ = result
        }
    })
}
```

## Comparing Sequential vs Parallel Performance

```go
func BenchmarkMapOperations(b *testing.B) {
    // Test data
    data := make(map[string]int)
    for i := 0; i < 1000; i++ {
        data[fmt.Sprintf("key%d", i)] = i
    }
    
    // Sequential benchmark
    b.Run("Sequential", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            key := fmt.Sprintf("key%d", i%1000)
            _ = data[key]
        }
    })
    
    // Parallel benchmark
    b.Run("Parallel", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            i := 0
            for pb.Next() {
                key := fmt.Sprintf("key%d", i%1000)
                _ = data[key]
                i++
            }
        })
    })
}
```

## Thread-Safe Data Structure Benchmarks

```go
func BenchmarkConcurrentMap(b *testing.B) {
    // Regular map with mutex
    b.Run("MapWithMutex", func(b *testing.B) {
        var mu sync.RWMutex
        data := make(map[string]int)
        
        // Pre-populate
        for i := 0; i < 1000; i++ {
            data[fmt.Sprintf("key%d", i)] = i
        }
        
        b.ResetTimer()
        b.RunParallel(func(pb *testing.PB) {
            i := 0
            for pb.Next() {
                key := fmt.Sprintf("key%d", i%1000)
                
                mu.RLock()
                _ = data[key]
                mu.RUnlock()
                
                i++
            }
        })
    })
    
    // sync.Map
    b.Run("SyncMap", func(b *testing.B) {
        var data sync.Map
        
        // Pre-populate
        for i := 0; i < 1000; i++ {
            data.Store(fmt.Sprintf("key%d", i), i)
        }
        
        b.ResetTimer()
        b.RunParallel(func(pb *testing.PB) {
            i := 0
            for pb.Next() {
                key := fmt.Sprintf("key%d", i%1000)
                _, _ = data.Load(key)
                i++
            }
        })
    })
}
```

## Channel Performance Benchmarks

```go
func BenchmarkChannelOperations(b *testing.B) {
    b.Run("Unbuffered", func(b *testing.B) {
        ch := make(chan int)
        
        b.RunParallel(func(pb *testing.PB) {
            go func() {
                for pb.Next() {
                    ch <- 42
                }
            }()
            
            for pb.Next() {
                <-ch
            }
        })
    })
    
    b.Run("Buffered/Size100", func(b *testing.B) {
        ch := make(chan int, 100)
        
        b.RunParallel(func(pb *testing.PB) {
            go func() {
                for pb.Next() {
                    select {
                    case ch <- 42:
                    default:
                    }
                }
            }()
            
            for pb.Next() {
                select {
                case <-ch:
                default:
                }
            }
        })
    })
}
```

## Worker Pool Benchmarks

```go
type WorkerPool struct {
    tasks chan func()
    wg    sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
    wp := &WorkerPool{
        tasks: make(chan func(), 100),
    }
    
    for i := 0; i < workerCount; i++ {
        go wp.worker()
    }
    
    return wp
}

func (wp *WorkerPool) worker() {
    for task := range wp.tasks {
        task()
        wp.wg.Done()
    }
}

func (wp *WorkerPool) Submit(task func()) {
    wp.wg.Add(1)
    wp.tasks <- task
}

func (wp *WorkerPool) Wait() {
    wp.wg.Wait()
}

func BenchmarkWorkerPool(b *testing.B) {
    workerCounts := []int{1, 2, 4, 8, 16}
    
    for _, count := range workerCounts {
        b.Run(fmt.Sprintf("Workers%d", count), func(b *testing.B) {
            pool := NewWorkerPool(count)
            
            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    pool.Submit(func() {
                        // Simulate work
                        time.Sleep(time.Microsecond)
                    })
                }
            })
            
            pool.Wait()
        })
    }
}
```

## Lock Contention Benchmarks

```go
func BenchmarkLockContention(b *testing.B) {
    // Mutex contention
    b.Run("Mutex", func(b *testing.B) {
        var mu sync.Mutex
        var counter int64
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.Lock()
                counter++
                mu.Unlock()
            }
        })
    })
    
    // RWMutex for reads
    b.Run("RWMutex/Read", func(b *testing.B) {
        var mu sync.RWMutex
        var data int64 = 42
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.RLock()
                _ = data
                mu.RUnlock()
            }
        })
    })
    
    // Atomic operations
    b.Run("Atomic", func(b *testing.B) {
        var counter int64
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                atomic.AddInt64(&counter, 1)
            }
        })
    })
}
```

## Memory Pool Benchmarks

```go
func BenchmarkMemoryPools(b *testing.B) {
    // Regular allocation
    b.Run("RegularAlloc", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                data := make([]byte, 1024)
                // Simulate usage
                for i := 0; i < len(data); i++ {
                    data[i] = byte(i % 256)
                }
                _ = data
            }
        })
    })
    
    // sync.Pool
    b.Run("SyncPool", func(b *testing.B) {
        pool := &sync.Pool{
            New: func() interface{} {
                return make([]byte, 1024)
            },
        }
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                data := pool.Get().([]byte)
                // Simulate usage
                for i := 0; i < len(data); i++ {
                    data[i] = byte(i % 256)
                }
                pool.Put(data)
            }
        })
    })
    
    // Channel-based pool
    b.Run("ChannelPool", func(b *testing.B) {
        poolSize := runtime.NumCPU() * 2
        pool := make(chan []byte, poolSize)
        
        // Pre-populate pool
        for i := 0; i < poolSize; i++ {
            pool <- make([]byte, 1024)
        }
        
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                var data []byte
                select {
                case data = <-pool:
                default:
                    data = make([]byte, 1024)
                }
                
                // Simulate usage
                for i := 0; i < len(data); i++ {
                    data[i] = byte(i % 256)
                }
                
                select {
                case pool <- data:
                default:
                    // Pool full, let GC handle it
                }
            }
        })
    })
}
```

## HTTP Server Benchmarks

```go
func BenchmarkHTTPServer(b *testing.B) {
    // Simple handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    client := &http.Client{
        Transport: &http.Transport{
            MaxIdleConnsPerHost: 100,
        },
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := client.Get(server.URL)
            if err != nil {
                b.Fatal(err)
            }
            resp.Body.Close()
        }
    })
}
```

## Database Connection Pool Benchmarks

```go
func BenchmarkDBConnectionPool(b *testing.B) {
    // Setup database with different pool sizes
    poolSizes := []int{1, 5, 10, 25, 50}
    
    for _, poolSize := range poolSizes {
        b.Run(fmt.Sprintf("PoolSize%d", poolSize), func(b *testing.B) {
            db := setupTestDB()
            db.SetMaxOpenConns(poolSize)
            db.SetMaxIdleConns(poolSize / 2)
            defer db.Close()
            
            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    var count int
                    err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
                    if err != nil {
                        b.Fatal(err)
                    }
                }
            })
        })
    }
}
```

## CPU-bound vs I/O-bound Workloads

```go
func BenchmarkWorkloadTypes(b *testing.B) {
    // CPU-bound workload
    b.Run("CPUBound", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                // CPU-intensive calculation
                result := 0
                for i := 0; i < 10000; i++ {
                    result += i * i
                }
                _ = result
            }
        })
    })
    
    // I/O-bound workload
    b.Run("IOBound", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                // Simulate I/O delay
                time.Sleep(time.Microsecond * 100)
            }
        })
    })
    
    // Mixed workload
    b.Run("Mixed", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                // Some CPU work
                result := 0
                for i := 0; i < 1000; i++ {
                    result += i
                }
                
                // Some I/O wait
                time.Sleep(time.Microsecond * 10)
                
                _ = result
            }
        })
    })
}
```

## GOMAXPROCS Impact Benchmarks

```go
func BenchmarkGOMAXPROCS(b *testing.B) {
    originalGOMAXPROCS := runtime.GOMAXPROCS(0)
    defer runtime.GOMAXPROCS(originalGOMAXPROCS)
    
    maxProcs := []int{1, 2, 4, 8, runtime.NumCPU()}
    
    for _, procs := range maxProcs {
        runtime.GOMAXPROCS(procs)
        
        b.Run(fmt.Sprintf("GOMAXPROCS%d", procs), func(b *testing.B) {
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    // CPU-bound work
                    result := 0
                    for i := 0; i < 1000; i++ {
                        result += i * i
                    }
                    _ = result
                }
            })
        })
    }
}
```

## Best Practices for Parallel Benchmarks

### 1. Avoid Shared State When Possible

```go
// BAD: Shared counter causes contention
func BenchmarkSharedCounter(b *testing.B) {
    var counter int64
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            atomic.AddInt64(&counter, 1)
            // This measures contention, not your algorithm
        }
    })
}

// GOOD: Per-goroutine state
func BenchmarkLocalCounter(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        localCounter := 0
        for pb.Next() {
            localCounter++
            // This measures your algorithm
        }
    })
}
```

### 2. Use Appropriate Buffer Sizes

```go
func BenchmarkChannelBuffers(b *testing.B) {
    bufferSizes := []int{0, 1, 10, 100, 1000}
    
    for _, size := range bufferSizes {
        b.Run(fmt.Sprintf("Buffer%d", size), func(b *testing.B) {
            ch := make(chan int, size)
            
            b.RunParallel(func(pb *testing.PB) {
                go func() {
                    for pb.Next() {
                        select {
                        case ch <- 42:
                        default:
                        }
                    }
                }()
                
                for pb.Next() {
                    select {
                    case <-ch:
                    default:
                    }
                }
            })
        })
    }
}
```

### 3. Measure Real Scenarios

```go
// Benchmark realistic parallel usage
func BenchmarkRealisticParallel(b *testing.B) {
    // Setup realistic shared resource
    cache := NewLRUCache(1000)
    
    // Pre-populate cache
    for i := 0; i < 500; i++ {
        cache.Put(fmt.Sprintf("key%d", i), i)
    }
    
    b.RunParallel(func(pb *testing.PB) {
        localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
        
        for pb.Next() {
            if localRand.Float32() < 0.8 {
                // 80% reads
                key := fmt.Sprintf("key%d", localRand.Intn(1000))
                cache.Get(key)
            } else {
                // 20% writes
                key := fmt.Sprintf("key%d", localRand.Intn(1000))
                cache.Put(key, localRand.Intn(10000))
            }
        }
    })
}
```

## Interpreting Parallel Benchmark Results

```
BenchmarkMapOperations/Sequential-8    20000000    75.0 ns/op
BenchmarkMapOperations/Parallel-8      50000000    25.0 ns/op
```

Key metrics to analyze:
- **Throughput improvement**: 50M vs 20M operations
- **Per-operation latency**: 25ns vs 75ns
- **Scalability**: How performance changes with core count
- **Contention effects**: Degradation under high concurrency

Parallel benchmarks reveal the true concurrency characteristics of your Go applications, helping you optimize for real-world concurrent usage patterns.
