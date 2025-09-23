# Sub-benchmarks

Learn to organize complex benchmark suites using Go's sub-benchmark functionality for comprehensive performance testing across multiple scenarios and parameters.

## Understanding Sub-benchmarks

Sub-benchmarks allow you to group related benchmarks and test multiple scenarios within a single benchmark function. They provide better organization and enable parameterized testing.

## Basic Sub-benchmark Structure

```go
func BenchmarkStringOperations(b *testing.B) {
    b.Run("Concatenation", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := "hello" + "world"
            _ = result
        }
    })
    
    b.Run("StringBuilder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var sb strings.Builder
            sb.WriteString("hello")
            sb.WriteString("world")
            _ = sb.String()
        }
    })
    
    b.Run("Sprintf", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := fmt.Sprintf("%s%s", "hello", "world")
            _ = result
        }
    })
}
```

## Parameterized Benchmarks

### Testing Multiple Data Sizes

```go
func BenchmarkSorting(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
            data := generateRandomSlice(size)
            b.ResetTimer()
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                // Make a copy to sort (don't modify original)
                sortData := make([]int, len(data))
                copy(sortData, data)
                sort.Ints(sortData)
            }
        })
    }
}

func generateRandomSlice(size int) []int {
    slice := make([]int, size)
    for i := range slice {
        slice[i] = rand.Intn(size)
    }
    return slice
}
```

### Testing Different Algorithms

```go
func BenchmarkSearchAlgorithms(b *testing.B) {
    // Prepare test data
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        data := make([]int, size)
        for i := range data {
            data[i] = i
        }
        target := size / 2 // Middle element
        
        b.Run(fmt.Sprintf("Linear/Size%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = linearSearch(data, target)
            }
        })
        
        b.Run(fmt.Sprintf("Binary/Size%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = binarySearch(data, target)
            }
        })
    }
}

func linearSearch(data []int, target int) int {
    for i, v := range data {
        if v == target {
            return i
        }
    }
    return -1
}

func binarySearch(data []int, target int) int {
    left, right := 0, len(data)-1
    
    for left <= right {
        mid := (left + right) / 2
        if data[mid] == target {
            return mid
        } else if data[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    return -1
}
```

## Advanced Sub-benchmark Patterns

### Nested Sub-benchmarks

```go
func BenchmarkDataStructures(b *testing.B) {
    operations := []string{"Insert", "Lookup", "Delete"}
    sizes := []int{1000, 10000, 100000}
    
    for _, op := range operations {
        b.Run(op, func(b *testing.B) {
            for _, size := range sizes {
                b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
                    benchmarkOperation(b, op, size)
                })
            }
        })
    }
}

func benchmarkOperation(b *testing.B, operation string, size int) {
    switch operation {
    case "Insert":
        benchmarkInsert(b, size)
    case "Lookup":
        benchmarkLookup(b, size)
    case "Delete":
        benchmarkDelete(b, size)
    }
}
```

### Configuration-based Benchmarks

```go
type BenchmarkConfig struct {
    Name        string
    BufferSize  int
    WorkerCount int
    Duration    time.Duration
}

func BenchmarkWorkerPool(b *testing.B) {
    configs := []BenchmarkConfig{
        {"Small", 100, 1, time.Millisecond},
        {"Medium", 1000, 4, time.Millisecond * 10},
        {"Large", 10000, 16, time.Millisecond * 100},
    }
    
    for _, config := range configs {
        b.Run(config.Name, func(b *testing.B) {
            benchmarkWorkerPoolWithConfig(b, config)
        })
    }
}

func benchmarkWorkerPoolWithConfig(b *testing.B, config BenchmarkConfig) {
    pool := NewWorkerPool(config.WorkerCount, config.BufferSize)
    defer pool.Close()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        task := &Task{Duration: config.Duration}
        pool.Submit(task)
        task.Wait()
    }
}
```

## Memory-focused Sub-benchmarks

```go
func BenchmarkMemoryPatterns(b *testing.B) {
    sizes := []int{1024, 8192, 65536}
    
    for _, size := range sizes {
        sizeLabel := fmt.Sprintf("Size%d", size)
        
        b.Run(sizeLabel+"/Allocate", func(b *testing.B) {
            b.ReportAllocs()
            for i := 0; i < b.N; i++ {
                data := make([]byte, size)
                _ = data
            }
        })
        
        b.Run(sizeLabel+"/Pool", func(b *testing.B) {
            pool := &sync.Pool{
                New: func() interface{} {
                    return make([]byte, size)
                },
            }
            b.ReportAllocs()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                data := pool.Get().([]byte)
                // Use the data
                for j := 0; j < len(data); j++ {
                    data[j] = 0
                }
                pool.Put(data)
            }
        })
        
        b.Run(sizeLabel+"/Reuse", func(b *testing.B) {
            data := make([]byte, size)
            b.ReportAllocs()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                // Reuse existing allocation
                for j := 0; j < len(data); j++ {
                    data[j] = byte(i % 256)
                }
            }
        })
    }
}
```

## Comparison Benchmarks

```go
func BenchmarkJSONLibraries(b *testing.B) {
    testData := generateTestObject()
    
    libraries := map[string]func(interface{}) ([]byte, error){
        "stdlib":     json.Marshal,
        "easyjson":   testData.MarshalJSON, // Assuming easyjson generated method
        "jsoniter":   jsoniter.Marshal,
        "gojson":     gojson.Marshal,
    }
    
    for name, marshalFunc := range libraries {
        b.Run("Marshal/"+name, func(b *testing.B) {
            b.ReportAllocs()
            for i := 0; i < b.N; i++ {
                _, err := marshalFunc(testData)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
    
    // Benchmark unmarshaling as well
    jsonData, _ := json.Marshal(testData)
    
    for name, unmarshalFunc := range getUnmarshalFunctions() {
        b.Run("Unmarshal/"+name, func(b *testing.B) {
            b.ReportAllocs()
            for i := 0; i < b.N; i++ {
                var result TestObject
                err := unmarshalFunc(jsonData, &result)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}
```

## Database Operation Sub-benchmarks

```go
func BenchmarkDatabaseOps(b *testing.B) {
    db := setupTestDB()
    defer db.Close()
    
    // Prepare test data
    users := generateTestUsers(1000)
    
    b.Run("Insert", func(b *testing.B) {
        b.Run("Single", func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                user := users[i%len(users)]
                _, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
                                  user.Name, user.Email)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
        
        b.Run("Batch", func(b *testing.B) {
            batchSize := 100
            for i := 0; i < b.N; i++ {
                tx, _ := db.Begin()
                for j := 0; j < batchSize; j++ {
                    user := users[(i*batchSize+j)%len(users)]
                    tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
                            user.Name, user.Email)
                }
                tx.Commit()
            }
        })
        
        b.Run("Prepared", func(b *testing.B) {
            stmt, _ := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
            defer stmt.Close()
            
            for i := 0; i < b.N; i++ {
                user := users[i%len(users)]
                _, err := stmt.Exec(user.Name, user.Email)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    })
    
    b.Run("Select", func(b *testing.B) {
        b.Run("ByID", func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                var user User
                err := db.QueryRow("SELECT name, email FROM users WHERE id = ?", 
                                   i%1000+1).Scan(&user.Name, &user.Email)
                if err != nil && err != sql.ErrNoRows {
                    b.Fatal(err)
                }
            }
        })
        
        b.Run("Range", func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                rows, err := db.Query("SELECT name, email FROM users LIMIT 10 OFFSET ?", 
                                      i%100)
                if err != nil {
                    b.Fatal(err)
                }
                
                var users []User
                for rows.Next() {
                    var user User
                    rows.Scan(&user.Name, &user.Email)
                    users = append(users, user)
                }
                rows.Close()
            }
        })
    })
}
```

## Running Sub-benchmarks

```bash
# Run all sub-benchmarks
go test -bench=.

# Run specific sub-benchmark group
go test -bench=BenchmarkStringOperations

# Run specific sub-benchmark
go test -bench=BenchmarkStringOperations/StringBuilder

# Run with pattern matching
go test -bench=BenchmarkSorting/Size1000

# Get memory allocation info
go test -bench=. -benchmem

# Run multiple times for consistency
go test -bench=. -count=5
```

## Organizing Results

Sub-benchmark output is hierarchical:

```
BenchmarkStringOperations/Concatenation-8      100000000    10.5 ns/op    0 B/op    0 allocs/op
BenchmarkStringOperations/StringBuilder-8       50000000    25.4 ns/op   64 B/op    1 allocs/op
BenchmarkStringOperations/Sprintf-8             20000000    67.8 ns/op   64 B/op    2 allocs/op

BenchmarkSorting/Size10-8                       10000000    120 ns/op     0 B/op    0 allocs/op
BenchmarkSorting/Size100-8                       1000000   1200 ns/op     0 B/op    0 allocs/op
BenchmarkSorting/Size1000-8                       100000  15000 ns/op     0 B/op    0 allocs/op
```

## Best Practices for Sub-benchmarks

### 1. Meaningful Names

```go
// GOOD: Descriptive names
b.Run("SmallData", func(b *testing.B) { ... })
b.Run("LargeData", func(b *testing.B) { ... })

// BAD: Unclear names
b.Run("Test1", func(b *testing.B) { ... })
b.Run("Test2", func(b *testing.B) { ... })
```

### 2. Consistent Setup

```go
func BenchmarkConsistentSetup(b *testing.B) {
    // Common setup for all sub-benchmarks
    testData := prepareTestData()
    
    for _, scenario := range scenarios {
        b.Run(scenario.Name, func(b *testing.B) {
            // Scenario-specific setup
            config := scenario.Setup(testData)
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                scenario.Execute(config)
            }
        })
    }
}
```

### 3. Proper Resource Management

```go
func BenchmarkResourceManagement(b *testing.B) {
    for _, test := range tests {
        b.Run(test.name, func(b *testing.B) {
            resource := test.setup()
            defer resource.cleanup() // Ensure cleanup
            
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                test.operation(resource)
            }
        })
    }
}
```

Sub-benchmarks provide powerful organization capabilities that make complex performance testing manageable and results easy to interpret.
