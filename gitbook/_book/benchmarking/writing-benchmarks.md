# Writing Effective Benchmarks

Writing effective benchmarks is both an art and a science. Well-designed benchmarks provide accurate, actionable insights into performance characteristics, while poorly designed ones can mislead optimization efforts and waste development time. This chapter covers the principles, patterns, and practices for creating benchmarks that drive meaningful performance improvements.

## Fundamental Principles of Benchmark Design

### Accuracy and Precision
Benchmarks must accurately measure what you intend to measure:

```go
package benchmark_examples

import (
    "crypto/rand"
    "strings"
    "testing"
    "time"
)

// BAD: Measures more than intended
func BenchmarkStringConcatBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        parts := []string{"Hello", "World", "From", "Go"} // Setup in loop
        result := strings.Join(parts, " ")                // Actual work
        _ = result
    }
}

// GOOD: Measures only the intended operation
func BenchmarkStringConcatGood(b *testing.B) {
    parts := []string{"Hello", "World", "From", "Go"} // Setup outside loop
    
    b.ResetTimer() // Exclude setup time
    
    for i := 0; i < b.N; i++ {
        result := strings.Join(parts, " ") // Only measure this
        _ = result
    }
}

// BETTER: Multiple scenarios with realistic data
func BenchmarkStringConcat(b *testing.B) {
    scenarios := []struct {
        name  string
        parts []string
    }{
        {"Small", []string{"Hello", "World"}},
        {"Medium", []string{"The", "quick", "brown", "fox", "jumps"}},
        {"Large", make([]string, 100)},
    }
    
    // Initialize large scenario
    for i := range scenarios[2].parts {
        scenarios[2].parts[i] = "word"
    }
    
    for _, scenario := range scenarios {
        b.Run(scenario.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                result := strings.Join(scenario.parts, " ")
                _ = result
            }
        })
    }
}
```

### Repeatability and Stability
Benchmarks should produce consistent results across runs:

```go
func BenchmarkWithStabilization(b *testing.B) {
    // Warm up to stabilize performance
    warmupData := generateTestData(1000)
    for i := 0; i < 100; i++ {
        _ = processData(warmupData)
    }
    
    // Force garbage collection to start with clean state
    runtime.GC()
    runtime.GC() // Call twice to ensure full collection
    
    // Allow scheduler to settle
    time.Sleep(10 * time.Millisecond)
    
    // Prepare test data
    testData := generateTestData(1000)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result := processData(testData)
        runtime.KeepAlive(result) // Prevent optimization
    }
}

// Advanced stabilization with multiple runs
func BenchmarkWithVarianceControl(b *testing.B) {
    const warmupRuns = 100
    const stabilityThreshold = 0.05 // 5% coefficient of variation
    
    testData := generateTestData(1000)
    
    // Warmup and stability check
    var warmupTimes []time.Duration
    for i := 0; i < warmupRuns; i++ {
        start := time.Now()
        _ = processData(testData)
        warmupTimes = append(warmupTimes, time.Since(start))
    }
    
    // Check if performance has stabilized
    cv := calculateCoefficientOfVariation(warmupTimes[warmupRuns-20:]) // Last 20 runs
    if cv > stabilityThreshold {
        b.Logf("Warning: High performance variance (CV=%.3f), results may be unreliable", cv)
    }
    
    runtime.GC()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result := processData(testData)
        runtime.KeepAlive(result)
    }
}

func calculateCoefficientOfVariation(durations []time.Duration) float64 {
    if len(durations) < 2 {
        return 0
    }
    
    var sum, sumSquares float64
    for _, d := range durations {
        val := float64(d.Nanoseconds())
        sum += val
        sumSquares += val * val
    }
    
    n := float64(len(durations))
    mean := sum / n
    variance := (sumSquares - sum*sum/n) / (n - 1)
    stdDev := math.Sqrt(variance)
    
    return stdDev / mean
}
```

### Realistic Workloads
Benchmarks should reflect real-world usage patterns:

```go
package realistic_benchmarks

import (
    "crypto/rand"
    "encoding/json"
    "fmt"
    "math/rand"
    "testing"
    "time"
)

// Model realistic data structures
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    Profile   Profile   `json:"profile"`
    Tags      []string  `json:"tags"`
}

type Profile struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Bio       string `json:"bio"`
    Avatar    string `json:"avatar"`
    Settings  map[string]interface{} `json:"settings"`
}

// Generate realistic test data with proper distributions
func generateRealisticUsers(count int) []User {
    users := make([]User, count)
    
    for i := range users {
        users[i] = User{
            ID:        i + 1,
            Username:  generateUsername(),
            Email:     generateEmail(),
            CreatedAt: generateCreationTime(),
            Profile:   generateProfile(),
            Tags:      generateTags(),
        }
    }
    
    return users
}

func generateUsername() string {
    // Realistic username patterns
    patterns := []string{
        "user_%d",
        "%s_%s",
        "%s%d",
        "%s.%s",
    }
    
    names := []string{"john", "alice", "bob", "charlie", "diana"}
    pattern := patterns[rand.Intn(len(patterns))]
    
    switch pattern {
    case "user_%d":
        return fmt.Sprintf(pattern, rand.Intn(100000))
    case "%s_%s":
        return fmt.Sprintf(pattern, names[rand.Intn(len(names))], names[rand.Intn(len(names))])
    case "%s%d":
        return fmt.Sprintf(pattern, names[rand.Intn(len(names))], rand.Intn(1000))
    case "%s.%s":
        return fmt.Sprintf(pattern, names[rand.Intn(len(names))], names[rand.Intn(len(names))])
    default:
        return "user"
    }
}

func generateEmail() string {
    domains := []string{"gmail.com", "yahoo.com", "outlook.com", "company.com"}
    username := generateUsername()
    domain := domains[rand.Intn(len(domains))]
    return fmt.Sprintf("%s@%s", username, domain)
}

func generateCreationTime() time.Time {
    // Users created over the last 2 years
    now := time.Now()
    twoYearsAgo := now.AddDate(-2, 0, 0)
    duration := now.Sub(twoYearsAgo)
    randomDuration := time.Duration(rand.Int63n(int64(duration)))
    return twoYearsAgo.Add(randomDuration)
}

func generateProfile() Profile {
    firstNames := []string{"John", "Alice", "Bob", "Charlie", "Diana"}
    lastNames := []string{"Smith", "Johnson", "Brown", "Davis", "Wilson"}
    
    return Profile{
        FirstName: firstNames[rand.Intn(len(firstNames))],
        LastName:  lastNames[rand.Intn(len(lastNames))],
        Bio:       generateBio(),
        Avatar:    fmt.Sprintf("https://avatar.service.com/%d.jpg", rand.Intn(1000)),
        Settings:  generateSettings(),
    }
}

func generateBio() string {
    bios := []string{
        "Software developer passionate about Go",
        "Building the future, one line of code at a time",
        "Coffee enthusiast and code optimizer",
        "Distributed systems engineer",
        "",  // Some users have empty bios
    }
    return bios[rand.Intn(len(bios))]
}

func generateSettings() map[string]interface{} {
    return map[string]interface{}{
        "theme":           []string{"light", "dark"}[rand.Intn(2)],
        "notifications":   rand.Intn(2) == 1,
        "privacy_level":   rand.Intn(3) + 1,
        "language":        []string{"en", "es", "fr", "de"}[rand.Intn(4)],
        "timezone":        generateTimezone(),
    }
}

func generateTimezone() string {
    timezones := []string{
        "America/New_York", "America/Los_Angeles", "Europe/London",
        "Europe/Berlin", "Asia/Tokyo", "Asia/Shanghai", "Australia/Sydney",
    }
    return timezones[rand.Intn(len(timezones))]
}

func generateTags() []string {
    allTags := []string{
        "developer", "golang", "python", "javascript", "docker",
        "kubernetes", "microservices", "cloud", "devops", "startup",
        "freelancer", "remote", "open-source", "tech-lead", "architect",
    }
    
    // Users have 0-5 tags
    numTags := rand.Intn(6)
    if numTags == 0 {
        return nil
    }
    
    // Shuffle and take first numTags
    shuffled := make([]string, len(allTags))
    copy(shuffled, allTags)
    rand.Shuffle(len(shuffled), func(i, j int) {
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    })
    
    return shuffled[:numTags]
}

// Benchmark JSON serialization with realistic data
func BenchmarkJSONSerialization(b *testing.B) {
    sizes := []int{1, 10, 100, 1000, 10000}
    
    for _, size := range sizes {
        users := generateRealisticUsers(size)
        
        b.Run(fmt.Sprintf("Users%d", size), func(b *testing.B) {
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                data, err := json.Marshal(users)
                if err != nil {
                    b.Fatal(err)
                }
                _ = data
            }
        })
    }
}

// Benchmark with realistic query patterns
func BenchmarkUserSearch(b *testing.B) {
    users := generateRealisticUsers(10000)
    
    // Create realistic search queries
    searchQueries := []struct {
        name  string
        query func([]User) []User
    }{
        {
            "ByUsername",
            func(users []User) []User {
                target := "alice"
                var results []User
                for _, user := range users {
                    if strings.Contains(strings.ToLower(user.Username), target) {
                        results = append(results, user)
                    }
                }
                return results
            },
        },
        {
            "ByTag",
            func(users []User) []User {
                target := "golang"
                var results []User
                for _, user := range users {
                    for _, tag := range user.Tags {
                        if tag == target {
                            results = append(results, user)
                            break
                        }
                    }
                }
                return results
            },
        },
        {
            "RecentUsers",
            func(users []User) []User {
                cutoff := time.Now().AddDate(0, -6, 0) // Last 6 months
                var results []User
                for _, user := range users {
                    if user.CreatedAt.After(cutoff) {
                        results = append(results, user)
                    }
                }
                return results
            },
        },
    }
    
    for _, query := range searchQueries {
        b.Run(query.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                results := query.query(users)
                _ = results
            }
        })
    }
}
```

## Memory Allocation Benchmarking

### Tracking Allocations
Monitor memory allocation patterns to identify optimization opportunities:

```go
func BenchmarkAllocationPatterns(b *testing.B) {
    b.ReportAllocs() // Enable allocation reporting
    
    for i := 0; i < b.N; i++ {
        result := allocatingFunction()
        _ = result
    }
}

// Compare allocation strategies
func BenchmarkStringBuilding(b *testing.B) {
    words := []string{"Hello", "World", "From", "Go", "Benchmarking"}
    
    b.Run("Concatenation", func(b *testing.B) {
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
            for _, word := range words {
                builder.WriteString(word)
                builder.WriteString(" ")
            }
            result := builder.String()
            _ = result
        }
    })
    
    b.Run("StringBuilderPrealloc", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            builder.Grow(50) // Pre-allocate based on expected size
            for _, word := range words {
                builder.WriteString(word)
                builder.WriteString(" ")
            }
            result := builder.String()
            _ = result
        }
    })
    
    b.Run("Join", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            result := strings.Join(words, " ")
            _ = result
        }
    })
}

// Benchmark memory pool effectiveness
func BenchmarkMemoryPools(b *testing.B) {
    // Without pool
    b.Run("WithoutPool", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            buffer := make([]byte, 1024)
            processBuffer(buffer)
        }
    })
    
    // With sync.Pool
    pool := sync.Pool{
        New: func() interface{} {
            return make([]byte, 1024)
        },
    }
    
    b.Run("WithPool", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            buffer := pool.Get().([]byte)
            processBuffer(buffer)
            pool.Put(buffer)
        }
    })
}

func processBuffer(buffer []byte) {
    // Simulate work with the buffer
    for i := range buffer {
        buffer[i] = byte(i % 256)
    }
}
```

### Zero-Allocation Benchmarking
Identify and measure zero-allocation code paths:

```go
func BenchmarkZeroAlloc(b *testing.B) {
    data := generateTestData(1000)
    
    b.Run("AllocatingVersion", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            result := processDataWithAlloc(data)
            _ = result
        }
    })
    
    b.Run("ZeroAllocVersion", func(b *testing.B) {
        b.ReportAllocs()
        // Pre-allocate reusable buffer
        buffer := make([]int, 1000)
        
        for i := 0; i < b.N; i++ {
            result := processDataZeroAlloc(data, buffer)
            _ = result
        }
    })
}

func processDataWithAlloc(data []int) []int {
    // Creates new slice on each call
    result := make([]int, len(data))
    for i, v := range data {
        result[i] = v * 2
    }
    return result
}

func processDataZeroAlloc(data []int, buffer []int) []int {
    // Reuses provided buffer
    if len(buffer) < len(data) {
        panic("buffer too small")
    }
    
    for i, v := range data {
        buffer[i] = v * 2
    }
    return buffer[:len(data)]
}

// Benchmark string operations without allocations
func BenchmarkStringOperations(b *testing.B) {
    input := "Hello, World! This is a test string for benchmarking."
    
    b.Run("WithAllocation", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            result := strings.ToUpper(input)
            _ = result
        }
    })
    
    b.Run("ZeroAllocation", func(b *testing.B) {
        b.ReportAllocs()
        buffer := make([]byte, len(input))
        
        for i := 0; i < b.N; i++ {
            uppercaseInPlace([]byte(input), buffer)
        }
    })
}

func uppercaseInPlace(src, dst []byte) {
    for i, b := range src {
        if b >= 'a' && b <= 'z' {
            dst[i] = b - 32
        } else {
            dst[i] = b
        }
    }
}
```

## Advanced Benchmarking Patterns

### Parallel Benchmarks
Test concurrent performance characteristics:

```go
func BenchmarkConcurrentOperations(b *testing.B) {
    data := generateTestData(10000)
    
    b.Run("Sequential", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            for _, item := range data {
                _ = expensiveOperation(item)
            }
        }
    })
    
    b.Run("Parallel", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                item := data[rand.Intn(len(data))]
                _ = expensiveOperation(item)
            }
        })
    })
    
    // Test different concurrency levels
    concurrencyLevels := []int{1, 2, 4, 8, 16, 32}
    for _, level := range concurrencyLevels {
        b.Run(fmt.Sprintf("Workers%d", level), func(b *testing.B) {
            benchmarkWithWorkers(b, data, level)
        })
    }
}

func benchmarkWithWorkers(b *testing.B, data []int, workers int) {
    work := make(chan int, len(data))
    var wg sync.WaitGroup
    
    // Pre-fill work channel
    for _, item := range data {
        work <- item
    }
    close(work)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        // Refill work channel for each iteration
        workCopy := make(chan int, len(data))
        for _, item := range data {
            workCopy <- item
        }
        close(workCopy)
        
        // Start workers
        for w := 0; w < workers; w++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for item := range workCopy {
                    _ = expensiveOperation(item)
                }
            }()
        }
        
        wg.Wait()
    }
}

func expensiveOperation(n int) int {
    // Simulate CPU-intensive work
    result := 0
    for i := 0; i < n%1000; i++ {
        result += i * i
    }
    return result
}
```

### Comparative Benchmarks
Compare multiple implementations systematically:

```go
func BenchmarkAlgorithmComparison(b *testing.B) {
    sizes := []int{100, 1000, 10000, 100000}
    
    algorithms := map[string]func([]int){
        "BubbleSort": bubbleSort,
        "QuickSort":  quickSort,
        "MergeSort":  mergeSort,
        "HeapSort":   heapSort,
        "SliceSort":  func(data []int) { sort.Ints(data) },
    }
    
    for _, size := range sizes {
        testData := generateRandomSlice(size)
        
        for name, algorithm := range algorithms {
            b.Run(fmt.Sprintf("%s/Size%d", name, size), func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    // Create fresh copy for each iteration
                    b.StopTimer()
                    data := make([]int, len(testData))
                    copy(data, testData)
                    b.StartTimer()
                    
                    algorithm(data)
                }
            })
        }
    }
}

// Benchmark data structure operations
func BenchmarkDataStructures(b *testing.B) {
    operations := []struct {
        name string
        size int
    }{
        {"Small", 100},
        {"Medium", 1000},
        {"Large", 10000},
    }
    
    for _, op := range operations {
        b.Run(fmt.Sprintf("Map/%s", op.name), func(b *testing.B) {
            benchmarkMap(b, op.size)
        })
        
        b.Run(fmt.Sprintf("Slice/%s", op.name), func(b *testing.B) {
            benchmarkSlice(b, op.size)
        })
        
        b.Run(fmt.Sprintf("SyncMap/%s", op.name), func(b *testing.B) {
            benchmarkSyncMap(b, op.size)
        })
    }
}

func benchmarkMap(b *testing.B, size int) {
    keys := generateKeys(size)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m := make(map[string]int, size)
        for j, key := range keys {
            m[key] = j
        }
        
        // Benchmark lookups
        for _, key := range keys[:10] { // Sample 10 lookups
            _ = m[key]
        }
    }
}

func benchmarkSlice(b *testing.B, size int) {
    keys := generateKeys(size)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice := make([]KeyValue, size)
        for j, key := range keys {
            slice[j] = KeyValue{Key: key, Value: j}
        }
        
        // Benchmark linear search
        for _, searchKey := range keys[:10] {
            for _, kv := range slice {
                if kv.Key == searchKey {
                    _ = kv.Value
                    break
                }
            }
        }
    }
}

func benchmarkSyncMap(b *testing.B, size int) {
    keys := generateKeys(size)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var m sync.Map
        for j, key := range keys {
            m.Store(key, j)
        }
        
        // Benchmark lookups
        for _, key := range keys[:10] {
            _, _ = m.Load(key)
        }
    }
}

type KeyValue struct {
    Key   string
    Value int
}

func generateKeys(count int) []string {
    keys := make([]string, count)
    for i := range keys {
        keys[i] = fmt.Sprintf("key_%d", i)
    }
    return keys
}

func generateRandomSlice(size int) []int {
    slice := make([]int, size)
    for i := range slice {
        slice[i] = rand.Intn(size * 10)
    }
    return slice
}
```

## Benchmark Analysis and Reporting

### Custom Benchmark Analysis
Create sophisticated analysis tools for benchmark results:

```go
package analysis

import (
    "encoding/json"
    "fmt"
    "math"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

type BenchmarkAnalyzer struct {
    results []BenchmarkResult
    config  AnalysisConfig
}

type AnalysisConfig struct {
    ConfidenceLevel    float64 `json:"confidence_level"`
    SignificanceThreshold float64 `json:"significance_threshold"`
    MinIterations      int     `json:"min_iterations"`
    OutlierThreshold   float64 `json:"outlier_threshold"`
}

type BenchmarkResult struct {
    Name        string    `json:"name"`
    Iterations  int       `json:"iterations"`
    NsPerOp     float64   `json:"ns_per_op"`
    BytesPerOp  int64     `json:"bytes_per_op"`
    AllocsPerOp int64     `json:"allocs_per_op"`
    Timestamp   time.Time `json:"timestamp"`
}

type AnalysisReport struct {
    Summary        Summary                    `json:"summary"`
    Comparisons    []Comparison              `json:"comparisons"`
    Regressions    []Regression              `json:"regressions"`
    Improvements   []Improvement             `json:"improvements"`
    Recommendations []Recommendation          `json:"recommendations"`
}

type Summary struct {
    TotalBenchmarks int                 `json:"total_benchmarks"`
    FastestBenchmark string             `json:"fastest_benchmark"`
    SlowestBenchmark string             `json:"slowest_benchmark"`
    Statistics      map[string]Statistics `json:"statistics"`
}

type Statistics struct {
    Mean         float64 `json:"mean"`
    Median       float64 `json:"median"`
    StdDev       float64 `json:"std_dev"`
    Min          float64 `json:"min"`
    Max          float64 `json:"max"`
    P95          float64 `json:"p95"`
    P99          float64 `json:"p99"`
    Outliers     int     `json:"outliers"`
}

type Comparison struct {
    Benchmark1  string  `json:"benchmark1"`
    Benchmark2  string  `json:"benchmark2"`
    Speedup     float64 `json:"speedup"`
    Significant bool    `json:"significant"`
}

type Regression struct {
    Benchmark    string    `json:"benchmark"`
    ChangePercent float64  `json:"change_percent"`
    Severity     string    `json:"severity"`
    Timestamp    time.Time `json:"timestamp"`
}

type Improvement struct {
    Benchmark    string    `json:"benchmark"`
    ChangePercent float64  `json:"change_percent"`
    Timestamp    time.Time `json:"timestamp"`
}

type Recommendation struct {
    Type        string `json:"type"`
    Benchmark   string `json:"benchmark"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
}

func NewBenchmarkAnalyzer(config AnalysisConfig) *BenchmarkAnalyzer {
    return &BenchmarkAnalyzer{
        config: config,
    }
}

func (ba *BenchmarkAnalyzer) LoadResults(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("reading results file: %w", err)
    }
    
    return json.Unmarshal(data, &ba.results)
}

func (ba *BenchmarkAnalyzer) ParseGoTestOutput(output string) error {
    lines := strings.Split(output, "\n")
    
    // Regex to parse benchmark lines
    re := regexp.MustCompile(`^(Benchmark\w+(?:/\w+)*)-(\d+)\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)
    
    for _, line := range lines {
        matches := re.FindStringSubmatch(strings.TrimSpace(line))
        if len(matches) < 5 {
            continue
        }
        
        result := BenchmarkResult{
            Name:      matches[1],
            Timestamp: time.Now(),
        }
        
        if iterations, err := strconv.Atoi(matches[3]); err == nil {
            result.Iterations = iterations
        }
        
        if nsPerOp, err := strconv.ParseFloat(matches[4], 64); err == nil {
            result.NsPerOp = nsPerOp
        }
        
        if len(matches) > 5 && matches[5] != "" {
            if bytesPerOp, err := strconv.ParseInt(matches[5], 10, 64); err == nil {
                result.BytesPerOp = bytesPerOp
            }
        }
        
        if len(matches) > 6 && matches[6] != "" {
            if allocsPerOp, err := strconv.ParseInt(matches[6], 10, 64); err == nil {
                result.AllocsPerOp = allocsPerOp
            }
        }
        
        ba.results = append(ba.results, result)
    }
    
    return nil
}

func (ba *BenchmarkAnalyzer) Analyze() AnalysisReport {
    report := AnalysisReport{
        Summary:         ba.generateSummary(),
        Comparisons:     ba.generateComparisons(),
        Regressions:     ba.detectRegressions(),
        Improvements:    ba.detectImprovements(),
        Recommendations: ba.generateRecommendations(),
    }
    
    return report
}

func (ba *BenchmarkAnalyzer) generateSummary() Summary {
    if len(ba.results) == 0 {
        return Summary{}
    }
    
    summary := Summary{
        TotalBenchmarks: len(ba.results),
        Statistics:      make(map[string]Statistics),
    }
    
    // Find fastest and slowest
    fastest := ba.results[0]
    slowest := ba.results[0]
    
    for _, result := range ba.results {
        if result.NsPerOp < fastest.NsPerOp {
            fastest = result
        }
        if result.NsPerOp > slowest.NsPerOp {
            slowest = result
        }
    }
    
    summary.FastestBenchmark = fastest.Name
    summary.SlowestBenchmark = slowest.Name
    
    // Calculate statistics for timing
    times := make([]float64, len(ba.results))
    for i, result := range ba.results {
        times[i] = result.NsPerOp
    }
    
    summary.Statistics["timing"] = ba.calculateStatistics(times)
    
    // Calculate statistics for memory allocations
    if ba.hasMemoryData() {
        var bytes, allocs []float64
        for _, result := range ba.results {
            if result.BytesPerOp > 0 {
                bytes = append(bytes, float64(result.BytesPerOp))
            }
            if result.AllocsPerOp > 0 {
                allocs = append(allocs, float64(result.AllocsPerOp))
            }
        }
        
        if len(bytes) > 0 {
            summary.Statistics["bytes"] = ba.calculateStatistics(bytes)
        }
        if len(allocs) > 0 {
            summary.Statistics["allocs"] = ba.calculateStatistics(allocs)
        }
    }
    
    return summary
}

func (ba *BenchmarkAnalyzer) calculateStatistics(values []float64) Statistics {
    if len(values) == 0 {
        return Statistics{}
    }
    
    sort.Float64s(values)
    
    mean := ba.calculateMean(values)
    stdDev := ba.calculateStdDev(values, mean)
    outliers := ba.countOutliers(values, mean, stdDev)
    
    return Statistics{
        Mean:     mean,
        Median:   ba.calculateMedian(values),
        StdDev:   stdDev,
        Min:      values[0],
        Max:      values[len(values)-1],
        P95:      ba.calculatePercentile(values, 0.95),
        P99:      ba.calculatePercentile(values, 0.99),
        Outliers: outliers,
    }
}

func (ba *BenchmarkAnalyzer) calculateMean(values []float64) float64 {
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

func (ba *BenchmarkAnalyzer) calculateMedian(sortedValues []float64) float64 {
    n := len(sortedValues)
    if n%2 == 0 {
        return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
    }
    return sortedValues[n/2]
}

func (ba *BenchmarkAnalyzer) calculateStdDev(values []float64, mean float64) float64 {
    variance := 0.0
    for _, v := range values {
        variance += math.Pow(v-mean, 2)
    }
    variance /= float64(len(values))
    return math.Sqrt(variance)
}

func (ba *BenchmarkAnalyzer) calculatePercentile(sortedValues []float64, percentile float64) float64 {
    index := percentile * float64(len(sortedValues)-1)
    lower := int(index)
    upper := lower + 1
    
    if upper >= len(sortedValues) {
        return sortedValues[len(sortedValues)-1]
    }
    
    weight := index - float64(lower)
    return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func (ba *BenchmarkAnalyzer) countOutliers(values []float64, mean, stdDev float64) int {
    threshold := ba.config.OutlierThreshold * stdDev
    count := 0
    
    for _, v := range values {
        if math.Abs(v-mean) > threshold {
            count++
        }
    }
    
    return count
}

func (ba *BenchmarkAnalyzer) generateComparisons() []Comparison {
    var comparisons []Comparison
    
    // Group benchmarks by base name (remove size suffixes)
    groups := ba.groupBenchmarks()
    
    for _, group := range groups {
        if len(group) < 2 {
            continue
        }
        
        // Compare each pair in the group
        for i := 0; i < len(group); i++ {
            for j := i + 1; j < len(group); j++ {
                comparison := Comparison{
                    Benchmark1: group[i].Name,
                    Benchmark2: group[j].Name,
                    Speedup:    group[i].NsPerOp / group[j].NsPerOp,
                    Significant: ba.isSignificantDifference(group[i], group[j]),
                }
                comparisons = append(comparisons, comparison)
            }
        }
    }
    
    // Sort by speedup (most significant first)
    sort.Slice(comparisons, func(i, j int) bool {
        return math.Abs(comparisons[i].Speedup-1.0) > math.Abs(comparisons[j].Speedup-1.0)
    })
    
    return comparisons
}

func (ba *BenchmarkAnalyzer) groupBenchmarks() map[string][]BenchmarkResult {
    groups := make(map[string][]BenchmarkResult)
    
    for _, result := range ba.results {
        // Extract base name (remove size/variant suffixes)
        baseName := ba.extractBaseName(result.Name)
        groups[baseName] = append(groups[baseName], result)
    }
    
    return groups
}

func (ba *BenchmarkAnalyzer) extractBaseName(fullName string) string {
    // Remove common suffixes like /Size100, /Small, etc.
    re := regexp.MustCompile(`/\w+\d+$|/\w+$`)
    return re.ReplaceAllString(fullName, "")
}

func (ba *BenchmarkAnalyzer) isSignificantDifference(a, b BenchmarkResult) bool {
    // Simple significance test based on relative difference
    diff := math.Abs(a.NsPerOp-b.NsPerOp) / math.Min(a.NsPerOp, b.NsPerOp)
    return diff > ba.config.SignificanceThreshold
}

func (ba *BenchmarkAnalyzer) detectRegressions() []Regression {
    // This would compare against historical data
    // For now, return empty slice
    return []Regression{}
}

func (ba *BenchmarkAnalyzer) detectImprovements() []Improvement {
    // This would compare against historical data
    // For now, return empty slice
    return []Improvement{}
}

func (ba *BenchmarkAnalyzer) generateRecommendations() []Recommendation {
    var recommendations []Recommendation
    
    for _, result := range ba.results {
        // High allocation count
        if result.AllocsPerOp > 10 {
            recommendations = append(recommendations, Recommendation{
                Type:        "memory",
                Benchmark:   result.Name,
                Description: fmt.Sprintf("High allocation count (%d allocs/op). Consider using object pooling or reducing allocations.", result.AllocsPerOp),
                Priority:    "high",
            })
        }
        
        // High memory usage
        if result.BytesPerOp > 10000 {
            recommendations = append(recommendations, Recommendation{
                Type:        "memory",
                Benchmark:   result.Name,
                Description: fmt.Sprintf("High memory usage (%d bytes/op). Consider optimizing data structures or algorithms.", result.BytesPerOp),
                Priority:    "medium",
            })
        }
        
        // Very slow operations
        if result.NsPerOp > 1000000 { // > 1ms
            recommendations = append(recommendations, Recommendation{
                Type:        "performance",
                Benchmark:   result.Name,
                Description: fmt.Sprintf("Slow operation (%.2f ms/op). Consider algorithmic optimization or caching.", result.NsPerOp/1000000),
                Priority:    "high",
            })
        }
    }
    
    return recommendations
}

func (ba *BenchmarkAnalyzer) hasMemoryData() bool {
    for _, result := range ba.results {
        if result.BytesPerOp > 0 || result.AllocsPerOp > 0 {
            return true
        }
    }
    return false
}

func (ar AnalysisReport) GenerateReport() string {
    var report strings.Builder
    
    report.WriteString("Benchmark Analysis Report\n")
    report.WriteString("========================\n\n")
    
    // Summary
    report.WriteString(fmt.Sprintf("Total Benchmarks: %d\n", ar.Summary.TotalBenchmarks))
    report.WriteString(fmt.Sprintf("Fastest: %s\n", ar.Summary.FastestBenchmark))
    report.WriteString(fmt.Sprintf("Slowest: %s\n", ar.Summary.SlowestBenchmark))
    report.WriteString("\n")
    
    // Statistics
    if stats, exists := ar.Summary.Statistics["timing"]; exists {
        report.WriteString("Timing Statistics:\n")
        report.WriteString(fmt.Sprintf("  Mean: %.2f ns/op\n", stats.Mean))
        report.WriteString(fmt.Sprintf("  Median: %.2f ns/op\n", stats.Median))
        report.WriteString(fmt.Sprintf("  Std Dev: %.2f ns/op\n", stats.StdDev))
        report.WriteString(fmt.Sprintf("  P95: %.2f ns/op\n", stats.P95))
        report.WriteString(fmt.Sprintf("  P99: %.2f ns/op\n", stats.P99))
        if stats.Outliers > 0 {
            report.WriteString(fmt.Sprintf("  Outliers: %d\n", stats.Outliers))
        }
        report.WriteString("\n")
    }
    
    // Top comparisons
    if len(ar.Comparisons) > 0 {
        report.WriteString("Top Performance Comparisons:\n")
        for i, comp := range ar.Comparisons {
            if i >= 5 { // Show top 5
                break
            }
            
            significance := ""
            if comp.Significant {
                significance = " (significant)"
            }
            
            if comp.Speedup > 1.0 {
                report.WriteString(fmt.Sprintf("  %s is %.2fx faster than %s%s\n",
                    comp.Benchmark1, comp.Speedup, comp.Benchmark2, significance))
            } else {
                report.WriteString(fmt.Sprintf("  %s is %.2fx slower than %s%s\n",
                    comp.Benchmark1, 1.0/comp.Speedup, comp.Benchmark2, significance))
            }
        }
        report.WriteString("\n")
    }
    
    // Recommendations
    if len(ar.Recommendations) > 0 {
        report.WriteString("Optimization Recommendations:\n")
        for _, rec := range ar.Recommendations {
            priority := strings.ToUpper(rec.Priority)
            report.WriteString(fmt.Sprintf("  [%s] %s: %s\n", priority, rec.Benchmark, rec.Description))
        }
    }
    
    return report.String()
}
```

Writing effective benchmarks is crucial for making data-driven performance decisions. By following these principles and patterns, you can create benchmarks that provide accurate, actionable insights into your application's performance characteristics and guide optimization efforts effectively.

## Key Takeaways

1. **Design benchmarks to measure only what matters** - exclude setup and irrelevant operations
2. **Use realistic data and workloads** that reflect production usage patterns
3. **Ensure benchmark stability** through proper warmup and variance control
4. **Track memory allocations** alongside timing for complete performance picture
5. **Use sub-benchmarks** to test different scenarios systematically
6. **Analyze results statistically** to ensure significance and reliability
7. **Generate actionable recommendations** from benchmark data
8. **Integrate benchmarks into development workflow** for continuous performance monitoring

Effective benchmarking transforms performance optimization from art to science, enabling confident, measurable improvements to your Go applications.
