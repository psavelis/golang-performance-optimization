# Algorithm Optimization

Comprehensive guide to optimizing algorithms and data structures in Go for maximum performance and efficiency.

## Overview

Algorithm optimization focuses on choosing the right algorithms and data structures for your specific use case. This involves understanding time and space complexity, analyzing algorithmic trade-offs, and implementing efficient solutions.

## Algorithm Complexity Analysis

### Time Complexity Fundamentals

Understanding Big O notation and its practical implications:

```go
// O(1) - Constant time
func getFirst(slice []int) int {
    if len(slice) == 0 {
        return 0
    }
    return slice[0]
}

// O(n) - Linear time
func findElement(slice []int, target int) bool {
    for _, value := range slice {
        if value == target {
            return true
        }
    }
    return false
}

// O(n²) - Quadratic time
func bubbleSort(arr []int) {
    n := len(arr)
    for i := 0; i < n-1; i++ {
        for j := 0; j < n-i-1; j++ {
            if arr[j] > arr[j+1] {
                arr[j], arr[j+1] = arr[j+1], arr[j]
            }
        }
    }
}

// O(log n) - Logarithmic time
func binarySearch(arr []int, target int) int {
    left, right := 0, len(arr)-1
    
    for left <= right {
        mid := left + (right-left)/2
        
        if arr[mid] == target {
            return mid
        } else if arr[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    
    return -1
}
```

### Space Complexity Considerations

```go
// O(1) space - In-place operations
func reverseSliceInPlace(slice []int) {
    left, right := 0, len(slice)-1
    for left < right {
        slice[left], slice[right] = slice[right], slice[left]
        left++
        right--
    }
}

// O(n) space - Additional storage
func reverseSliceCopy(slice []int) []int {
    result := make([]int, len(slice))
    for i, v := range slice {
        result[len(slice)-1-i] = v
    }
    return result
}
```

## Data Structure Selection

### Array vs Slice vs Map Performance

```go
func BenchmarkDataStructures(b *testing.B) {
    // Array - fixed size, stack allocated
    b.Run("Array", func(b *testing.B) {
        var arr [1000]int
        for i := 0; i < b.N; i++ {
            arr[i%1000] = i
        }
    })
    
    // Slice - dynamic, heap allocated
    b.Run("Slice", func(b *testing.B) {
        slice := make([]int, 1000)
        for i := 0; i < b.N; i++ {
            slice[i%1000] = i
        }
    })
    
    // Map - hash table
    b.Run("Map", func(b *testing.B) {
        m := make(map[int]int, 1000)
        for i := 0; i < b.N; i++ {
            m[i%1000] = i
        }
    })
}
```

### Optimal Data Structure Choices

#### For Fast Lookups
```go
// Map for O(1) average case lookups
type FastLookup struct {
    data map[string]interface{}
}

func (f *FastLookup) Get(key string) (interface{}, bool) {
    value, exists := f.data[key]
    return value, exists
}

// Sorted slice for O(log n) lookups with less memory
type SortedLookup struct {
    keys   []string
    values []interface{}
}

func (s *SortedLookup) Get(key string) (interface{}, bool) {
    idx := sort.SearchStrings(s.keys, key)
    if idx < len(s.keys) && s.keys[idx] == key {
        return s.values[idx], true
    }
    return nil, false
}
```

#### For Ordered Data
```go
// Heap for priority operations
type PriorityQueue []*Item

type Item struct {
    value    string
    priority int
    index    int
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*Item)
    item.index = n
    *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    old[n-1] = nil
    item.index = -1
    *pq = old[0 : n-1]
    return item
}
```

## Algorithm Optimization Patterns

### 1. Memoization
Cache expensive computations:

```go
type FibonacciCalculator struct {
    cache map[int]int
    mutex sync.RWMutex
}

func NewFibonacciCalculator() *FibonacciCalculator {
    return &FibonacciCalculator{
        cache: make(map[int]int),
    }
}

func (f *FibonacciCalculator) Calculate(n int) int {
    // Check cache first
    f.mutex.RLock()
    if result, exists := f.cache[n]; exists {
        f.mutex.RUnlock()
        return result
    }
    f.mutex.RUnlock()
    
    // Calculate and cache
    var result int
    if n <= 1 {
        result = n
    } else {
        result = f.Calculate(n-1) + f.Calculate(n-2)
    }
    
    f.mutex.Lock()
    f.cache[n] = result
    f.mutex.Unlock()
    
    return result
}
```

### 2. Lazy Evaluation
Defer expensive operations until needed:

```go
type LazyProcessor struct {
    data     []int
    computed bool
    result   int
}

func (l *LazyProcessor) GetSum() int {
    if !l.computed {
        l.result = l.computeSum()
        l.computed = true
    }
    return l.result
}

func (l *LazyProcessor) computeSum() int {
    sum := 0
    for _, value := range l.data {
        sum += value
    }
    return sum
}
```

### 3. Divide and Conquer
Break problems into smaller subproblems:

```go
func mergeSort(arr []int) []int {
    if len(arr) <= 1 {
        return arr
    }
    
    mid := len(arr) / 2
    left := mergeSort(arr[:mid])
    right := mergeSort(arr[mid:])
    
    return merge(left, right)
}

func merge(left, right []int) []int {
    result := make([]int, 0, len(left)+len(right))
    i, j := 0, 0
    
    for i < len(left) && j < len(right) {
        if left[i] <= right[j] {
            result = append(result, left[i])
            i++
        } else {
            result = append(result, right[j])
            j++
        }
    }
    
    result = append(result, left[i:]...)
    result = append(result, right[j:]...)
    
    return result
}
```

## Optimization Techniques

### Loop Optimization

```go
// Inefficient - multiple iterations
func processDataSlow(data []int) (sum, max, min int) {
    // First pass for sum
    for _, value := range data {
        sum += value
    }
    
    // Second pass for max
    max = data[0]
    for _, value := range data {
        if value > max {
            max = value
        }
    }
    
    // Third pass for min
    min = data[0]
    for _, value := range data {
        if value < min {
            min = value
        }
    }
    
    return sum, max, min
}

// Optimized - single iteration
func processDataFast(data []int) (sum, max, min int) {
    if len(data) == 0 {
        return 0, 0, 0
    }
    
    sum = data[0]
    max = data[0]
    min = data[0]
    
    for i := 1; i < len(data); i++ {
        value := data[i]
        sum += value
        if value > max {
            max = value
        }
        if value < min {
            min = value
        }
    }
    
    return sum, max, min
}
```

### Early Termination

```go
// Early exit when condition is met
func findFirst(slice []string, predicate func(string) bool) (string, bool) {
    for _, item := range slice {
        if predicate(item) {
            return item, true
        }
    }
    return "", false
}

// Short-circuit evaluation
func allValid(items []Item) bool {
    for _, item := range items {
        if !item.IsValid() {
            return false // Exit immediately on first invalid
        }
    }
    return true
}
```

### Batch Processing

```go
// Process items in batches to improve cache locality
func processBatches(items []Item, batchSize int) []Result {
    results := make([]Result, 0, len(items))
    
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        batchResults := processBatch(batch)
        results = append(results, batchResults...)
    }
    
    return results
}

func processBatch(batch []Item) []Result {
    results := make([]Result, len(batch))
    for i, item := range batch {
        results[i] = item.Process()
    }
    return results
}
```

## Performance Benchmarking

### Comparative Algorithm Analysis

```go
func BenchmarkSortingAlgorithms(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        data := generateRandomData(size)
        
        b.Run(fmt.Sprintf("BubbleSort_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                b.StopTimer()
                testData := make([]int, len(data))
                copy(testData, data)
                b.StartTimer()
                
                bubbleSort(testData)
            }
        })
        
        b.Run(fmt.Sprintf("QuickSort_%d", size), func(b *testing.B) {
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
```

## Algorithm Selection Guidelines

### Choose Based on Use Case

1. **Frequent Reads, Infrequent Writes**
   - Sorted slices with binary search
   - Pre-computed data structures

2. **Frequent Writes, Infrequent Reads**
   - Hash maps for O(1) insertions
   - Append-only structures

3. **Memory-Constrained Environments**
   - In-place algorithms
   - Streaming algorithms
   - Compact data structures

4. **Time-Critical Operations**
   - Pre-computed lookup tables
   - Optimized hot paths
   - Cache-friendly algorithms

### Performance Trade-offs

```go
// Time-optimized: O(1) lookup, O(n) space
type FastCache struct {
    data map[string]string
}

// Space-optimized: O(log n) lookup, O(n) space (sorted)
type CompactCache struct {
    keys   []string
    values []string
}

// Balanced: O(1) average, O(n) worst case, moderate space
type BalancedCache struct {
    buckets [][]KeyValue
}
```

## Next Steps

Explore specific algorithm optimization areas:

1. **[Data Structures](data-structures.md)** - Choosing optimal data structures
2. **[Time Complexity](time-complexity.md)** - Understanding and optimizing algorithmic complexity
3. **[Space Complexity](space-complexity.md)** - Memory-efficient algorithm design

The key to algorithm optimization is understanding your data patterns, access patterns, and performance requirements, then choosing the most appropriate algorithmic approach.
