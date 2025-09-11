# Algorithm Optimization

Algorithmic optimization is fundamental to building high-performance Go applications. This chapter explores advanced algorithmic techniques, data structure selection, complexity analysis, and implementation strategies that maximize computational efficiency while maintaining code readability and maintainability.

## Complexity Analysis and Algorithm Selection

### Big O Analysis in Practice
Understanding and applying complexity analysis to real-world scenarios:

```go
package algorithm_optimization

import (
    "container/heap"
    "math"
    "sort"
    "sync"
    "time"
)

// Complexity analyzer for algorithm selection
type ComplexityAnalyzer struct {
    measurements map[string][]Measurement
    mu          sync.RWMutex
}

type Measurement struct {
    InputSize   int           `json:"input_size"`
    Duration    time.Duration `json:"duration"`
    MemoryUsed  int64        `json:"memory_used"`
    Algorithm   string       `json:"algorithm"`
    Timestamp   time.Time    `json:"timestamp"`
}

func NewComplexityAnalyzer() *ComplexityAnalyzer {
    return &ComplexityAnalyzer{
        measurements: make(map[string][]Measurement),
    }
}

func (ca *ComplexityAnalyzer) Measure(algorithm string, inputSize int, fn func()) {
    var m1, m2 runtime.MemStats
    
    runtime.ReadMemStats(&m1)
    start := time.Now()
    
    fn()
    
    duration := time.Since(start)
    runtime.ReadMemStats(&m2)
    
    memoryUsed := int64(m2.TotalAlloc - m1.TotalAlloc)
    
    measurement := Measurement{
        InputSize:  inputSize,
        Duration:   duration,
        MemoryUsed: memoryUsed,
        Algorithm:  algorithm,
        Timestamp:  time.Now(),
    }
    
    ca.mu.Lock()
    ca.measurements[algorithm] = append(ca.measurements[algorithm], measurement)
    ca.mu.Unlock()
}

func (ca *ComplexityAnalyzer) AnalyzeComplexity(algorithm string) ComplexityResult {
    ca.mu.RLock()
    measurements := ca.measurements[algorithm]
    ca.mu.RUnlock()
    
    if len(measurements) < 3 {
        return ComplexityResult{Algorithm: algorithm, Confidence: 0}
    }
    
    return ca.fitComplexity(measurements)
}

type ComplexityResult struct {
    Algorithm   string  `json:"algorithm"`
    TimeClass   string  `json:"time_complexity"`
    SpaceClass  string  `json:"space_complexity"`
    Confidence  float64 `json:"confidence"`
    R2Score     float64 `json:"r2_score"`
    Constant    float64 `json:"constant_factor"`
}

func (ca *ComplexityAnalyzer) fitComplexity(measurements []Measurement) ComplexityResult {
    // Sort by input size
    sort.Slice(measurements, func(i, j int) bool {
        return measurements[i].InputSize < measurements[j].InputSize
    })
    
    // Test different complexity functions
    complexityFunctions := map[string]func(int) float64{
        "O(1)":        func(n int) float64 { return 1 },
        "O(log n)":    func(n int) float64 { return math.Log(float64(n)) },
        "O(n)":        func(n int) float64 { return float64(n) },
        "O(n log n)":  func(n int) float64 { return float64(n) * math.Log(float64(n)) },
        "O(n²)":       func(n int) float64 { return float64(n * n) },
        "O(n³)":       func(n int) float64 { return float64(n * n * n) },
        "O(2ⁿ)":       func(n int) float64 { return math.Pow(2, float64(n)) },
    }
    
    bestFit := ComplexityResult{Algorithm: measurements[0].Algorithm}
    
    for complexity, fn := range complexityFunctions {
        r2 := ca.calculateR2(measurements, fn)
        
        if r2 > bestFit.R2Score {
            bestFit.TimeClass = complexity
            bestFit.R2Score = r2
            bestFit.Confidence = r2
        }
    }
    
    return bestFit
}

func (ca *ComplexityAnalyzer) calculateR2(measurements []Measurement, fn func(int) float64) float64 {
    if len(measurements) < 2 {
        return 0
    }
    
    // Calculate mean of actual values
    var sumActual float64
    for _, m := range measurements {
        sumActual += float64(m.Duration.Nanoseconds())
    }
    meanActual := sumActual / float64(len(measurements))
    
    // Calculate sum of squares
    var ssRes, ssTot float64
    
    // Fit linear regression: duration = a * fn(n) + b
    var sumX, sumY, sumXY, sumX2 float64
    for _, m := range measurements {
        x := fn(m.InputSize)
        y := float64(m.Duration.Nanoseconds())
        
        sumX += x
        sumY += y
        sumXY += x * y
        sumX2 += x * x
    }
    
    n := float64(len(measurements))
    a := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
    b := (sumY - a*sumX) / n
    
    // Calculate R²
    for _, m := range measurements {
        actual := float64(m.Duration.Nanoseconds())
        predicted := a*fn(m.InputSize) + b
        
        ssRes += (actual - predicted) * (actual - predicted)
        ssTot += (actual - meanActual) * (actual - meanActual)
    }
    
    if ssTot == 0 {
        return 1.0
    }
    
    return 1.0 - (ssRes / ssTot)
}

// Advanced sorting algorithms with complexity analysis
type SortingBenchmark struct {
    analyzer *ComplexityAnalyzer
}

func NewSortingBenchmark() *SortingBenchmark {
    return &SortingBenchmark{
        analyzer: NewComplexityAnalyzer(),
    }
}

// Hybrid sorting algorithm that adapts based on input characteristics
func (sb *SortingBenchmark) AdaptiveSort(data []int) {
    size := len(data)
    
    // Algorithm selection based on size and characteristics
    if size <= 1 {
        return
    }
    
    if size <= 47 { // Tuned threshold
        sb.analyzer.Measure("InsertionSort", size, func() {
            insertionSort(data)
        })
        return
    }
    
    if sb.isNearlySorted(data) {
        sb.analyzer.Measure("TimSort", size, func() {
            timSort(data)
        })
        return
    }
    
    if sb.hasLimitedRange(data) {
        sb.analyzer.Measure("RadixSort", size, func() {
            radixSort(data)
        })
        return
    }
    
    // Default to introsort (hybrid quicksort/heapsort)
    sb.analyzer.Measure("IntroSort", size, func() {
        introSort(data, 0, size-1, 2*int(math.Log2(float64(size))))
    })
}

func (sb *SortingBenchmark) isNearlySorted(data []int) bool {
    inversions := 0
    threshold := len(data) / 10 // 10% inversions threshold
    
    for i := 1; i < len(data) && inversions <= threshold; i++ {
        if data[i] < data[i-1] {
            inversions++
        }
    }
    
    return inversions <= threshold
}

func (sb *SortingBenchmark) hasLimitedRange(data []int) bool {
    if len(data) < 100 {
        return false
    }
    
    min, max := data[0], data[0]
    for _, v := range data[1:] {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }
    
    // Use radix sort if range is reasonable compared to input size
    return (max - min) < len(data) * 10
}

// Introspective sort implementation
func introSort(data []int, low, high, maxDepth int) {
    for high-low > 16 {
        if maxDepth == 0 {
            heapSort(data[low:high+1])
            return
        }
        
        pivot := partition(data, low, high)
        
        // Recursively sort smaller partition first to limit stack depth
        if pivot-low < high-pivot {
            introSort(data, low, pivot-1, maxDepth-1)
            low = pivot + 1
        } else {
            introSort(data, pivot+1, high, maxDepth-1)
            high = pivot - 1
        }
        maxDepth--
    }
    
    // Use insertion sort for small subarrays
    insertionSortRange(data, low, high)
}

func partition(data []int, low, high int) int {
    // Median-of-three pivot selection
    mid := low + (high-low)/2
    if data[mid] < data[low] {
        data[low], data[mid] = data[mid], data[low]
    }
    if data[high] < data[low] {
        data[low], data[high] = data[high], data[low]
    }
    if data[high] < data[mid] {
        data[mid], data[high] = data[high], data[mid]
    }
    
    pivot := data[high]
    i := low - 1
    
    for j := low; j < high; j++ {
        if data[j] <= pivot {
            i++
            data[i], data[j] = data[j], data[i]
        }
    }
    
    data[i+1], data[high] = data[high], data[i+1]
    return i + 1
}

func heapSort(data []int) {
    n := len(data)
    
    // Build max heap
    for i := n/2 - 1; i >= 0; i-- {
        heapify(data, n, i)
    }
    
    // Extract elements one by one
    for i := n - 1; i > 0; i-- {
        data[0], data[i] = data[i], data[0]
        heapify(data, i, 0)
    }
}

func heapify(data []int, n, i int) {
    largest := i
    left := 2*i + 1
    right := 2*i + 2
    
    if left < n && data[left] > data[largest] {
        largest = left
    }
    
    if right < n && data[right] > data[largest] {
        largest = right
    }
    
    if largest != i {
        data[i], data[largest] = data[largest], data[i]
        heapify(data, n, largest)
    }
}

func insertionSort(data []int) {
    for i := 1; i < len(data); i++ {
        key := data[i]
        j := i - 1
        
        for j >= 0 && data[j] > key {
            data[j+1] = data[j]
            j--
        }
        
        data[j+1] = key
    }
}

func insertionSortRange(data []int, low, high int) {
    for i := low + 1; i <= high; i++ {
        key := data[i]
        j := i - 1
        
        for j >= low && data[j] > key {
            data[j+1] = data[j]
            j--
        }
        
        data[j+1] = key
    }
}

// TimSort implementation (adaptive merge sort)
func timSort(data []int) {
    n := len(data)
    if n < 2 {
        return
    }
    
    minMerge := getMinRunLength(n)
    
    // Sort individual runs of size minMerge using insertion sort
    for start := 0; start < n; start += minMerge {
        end := min(start+minMerge-1, n-1)
        insertionSortRange(data, start, end)
    }
    
    // Start merging runs
    size := minMerge
    for size < n {
        for start := 0; start < n; start += size*2 {
            mid := start + size - 1
            end := min(start+size*2-1, n-1)
            
            if mid < end {
                merge(data, start, mid, end)
            }
        }
        size *= 2
    }
}

func getMinRunLength(n int) int {
    r := 0
    for n >= 32 {
        r |= n & 1
        n >>= 1
    }
    return n + r
}

func merge(data []int, left, mid, right int) {
    // Create temporary arrays
    leftArr := make([]int, mid-left+1)
    rightArr := make([]int, right-mid)
    
    copy(leftArr, data[left:mid+1])
    copy(rightArr, data[mid+1:right+1])
    
    i, j, k := 0, 0, left
    
    // Merge the temporary arrays back
    for i < len(leftArr) && j < len(rightArr) {
        if leftArr[i] <= rightArr[j] {
            data[k] = leftArr[i]
            i++
        } else {
            data[k] = rightArr[j]
            j++
        }
        k++
    }
    
    // Copy remaining elements
    for i < len(leftArr) {
        data[k] = leftArr[i]
        i++
        k++
    }
    
    for j < len(rightArr) {
        data[k] = rightArr[j]
        j++
        k++
    }
}

// Radix sort for limited range integers
func radixSort(data []int) {
    if len(data) <= 1 {
        return
    }
    
    // Handle negative numbers by offsetting
    min := data[0]
    for _, v := range data {
        if v < min {
            min = v
        }
    }
    
    // Offset all values to make them non-negative
    for i := range data {
        data[i] -= min
    }
    
    // Find maximum to determine number of digits
    max := data[0]
    for _, v := range data {
        if v > max {
            max = v
        }
    }
    
    // Perform counting sort for each digit
    for exp := 1; max/exp > 0; exp *= 10 {
        countingSort(data, exp)
    }
    
    // Restore original values
    for i := range data {
        data[i] += min
    }
}

func countingSort(data []int, exp int) {
    output := make([]int, len(data))
    count := make([]int, 10)
    
    // Count occurrences of each digit
    for _, v := range data {
        count[(v/exp)%10]++
    }
    
    // Change count[i] to actual position
    for i := 1; i < 10; i++ {
        count[i] += count[i-1]
    }
    
    // Build output array
    for i := len(data) - 1; i >= 0; i-- {
        digit := (data[i] / exp) % 10
        output[count[digit]-1] = data[i]
        count[digit]--
    }
    
    // Copy output array to data
    copy(data, output)
}
```

## Advanced Data Structures

### Custom Data Structures for Specific Use Cases
Implement specialized data structures optimized for particular workloads:

```go
// Adaptive data structure that switches implementation based on usage patterns
type AdaptiveSet struct {
    implementation SetImplementation
    metrics       UsageMetrics
    threshold     AdaptiveThreshold
}

type SetImplementation interface {
    Add(item interface{}) bool
    Remove(item interface{}) bool
    Contains(item interface{}) bool
    Size() int
    Iterate() []interface{}
}

type UsageMetrics struct {
    AddCount      int64
    RemoveCount   int64
    ContainsCount int64
    IterateCount  int64
    Size          int64
    LastEvaluation time.Time
}

type AdaptiveThreshold struct {
    SmallSetSize    int
    MediumSetSize   int
    HighIterationRatio float64
    HighModificationRatio float64
}

// Array-based set for small collections
type ArraySet struct {
    items []interface{}
    mu    sync.RWMutex
}

func NewArraySet() *ArraySet {
    return &ArraySet{
        items: make([]interface{}, 0, 16),
    }
}

func (as *ArraySet) Add(item interface{}) bool {
    as.mu.Lock()
    defer as.mu.Unlock()
    
    // Check if item already exists
    for _, existing := range as.items {
        if existing == item {
            return false
        }
    }
    
    as.items = append(as.items, item)
    return true
}

func (as *ArraySet) Remove(item interface{}) bool {
    as.mu.Lock()
    defer as.mu.Unlock()
    
    for i, existing := range as.items {
        if existing == item {
            // Remove by swapping with last element
            as.items[i] = as.items[len(as.items)-1]
            as.items = as.items[:len(as.items)-1]
            return true
        }
    }
    
    return false
}

func (as *ArraySet) Contains(item interface{}) bool {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    for _, existing := range as.items {
        if existing == item {
            return true
        }
    }
    
    return false
}

func (as *ArraySet) Size() int {
    as.mu.RLock()
    defer as.mu.RUnlock()
    return len(as.items)
}

func (as *ArraySet) Iterate() []interface{} {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    result := make([]interface{}, len(as.items))
    copy(result, as.items)
    return result
}

// Tree-based set for medium collections with good iteration performance
type TreeSet struct {
    root *TreeNode
    size int
    mu   sync.RWMutex
}

type TreeNode struct {
    value  interface{}
    left   *TreeNode
    right  *TreeNode
    height int
}

func NewTreeSet() *TreeSet {
    return &TreeSet{}
}

func (ts *TreeSet) Add(item interface{}) bool {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    
    initialSize := ts.size
    ts.root = ts.insert(ts.root, item)
    return ts.size > initialSize
}

func (ts *TreeSet) insert(node *TreeNode, value interface{}) *TreeNode {
    if node == nil {
        ts.size++
        return &TreeNode{value: value, height: 1}
    }
    
    cmp := ts.compare(value, node.value)
    if cmp < 0 {
        node.left = ts.insert(node.left, value)
    } else if cmp > 0 {
        node.right = ts.insert(node.right, value)
    } else {
        return node // Duplicate
    }
    
    // Update height and rebalance
    node.height = 1 + max(ts.getHeight(node.left), ts.getHeight(node.right))
    return ts.rebalance(node)
}

func (ts *TreeSet) rebalance(node *TreeNode) *TreeNode {
    balance := ts.getBalance(node)
    
    // Left Heavy
    if balance > 1 {
        if ts.getBalance(node.left) < 0 {
            node.left = ts.rotateLeft(node.left)
        }
        return ts.rotateRight(node)
    }
    
    // Right Heavy
    if balance < -1 {
        if ts.getBalance(node.right) > 0 {
            node.right = ts.rotateRight(node.right)
        }
        return ts.rotateLeft(node)
    }
    
    return node
}

func (ts *TreeSet) rotateLeft(node *TreeNode) *TreeNode {
    newRoot := node.right
    node.right = newRoot.left
    newRoot.left = node
    
    node.height = 1 + max(ts.getHeight(node.left), ts.getHeight(node.right))
    newRoot.height = 1 + max(ts.getHeight(newRoot.left), ts.getHeight(newRoot.right))
    
    return newRoot
}

func (ts *TreeSet) rotateRight(node *TreeNode) *TreeNode {
    newRoot := node.left
    node.left = newRoot.right
    newRoot.right = node
    
    node.height = 1 + max(ts.getHeight(node.left), ts.getHeight(node.right))
    newRoot.height = 1 + max(ts.getHeight(newRoot.left), ts.getHeight(newRoot.right))
    
    return newRoot
}

func (ts *TreeSet) getHeight(node *TreeNode) int {
    if node == nil {
        return 0
    }
    return node.height
}

func (ts *TreeSet) getBalance(node *TreeNode) int {
    if node == nil {
        return 0
    }
    return ts.getHeight(node.left) - ts.getHeight(node.right)
}

func (ts *TreeSet) compare(a, b interface{}) int {
    // Simple comparison for demonstration
    aStr := fmt.Sprintf("%v", a)
    bStr := fmt.Sprintf("%v", b)
    
    if aStr < bStr {
        return -1
    } else if aStr > bStr {
        return 1
    }
    return 0
}

func (ts *TreeSet) Contains(item interface{}) bool {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    
    return ts.search(ts.root, item)
}

func (ts *TreeSet) search(node *TreeNode, value interface{}) bool {
    if node == nil {
        return false
    }
    
    cmp := ts.compare(value, node.value)
    if cmp == 0 {
        return true
    } else if cmp < 0 {
        return ts.search(node.left, value)
    } else {
        return ts.search(node.right, value)
    }
}

func (ts *TreeSet) Remove(item interface{}) bool {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    
    initialSize := ts.size
    ts.root = ts.delete(ts.root, item)
    return ts.size < initialSize
}

func (ts *TreeSet) delete(node *TreeNode, value interface{}) *TreeNode {
    if node == nil {
        return nil
    }
    
    cmp := ts.compare(value, node.value)
    if cmp < 0 {
        node.left = ts.delete(node.left, value)
    } else if cmp > 0 {
        node.right = ts.delete(node.right, value)
    } else {
        ts.size--
        
        if node.left == nil {
            return node.right
        } else if node.right == nil {
            return node.left
        }
        
        // Node with two children
        successor := ts.findMin(node.right)
        node.value = successor.value
        node.right = ts.delete(node.right, successor.value)
        ts.size++ // Adjust for the extra decrement
    }
    
    node.height = 1 + max(ts.getHeight(node.left), ts.getHeight(node.right))
    return ts.rebalance(node)
}

func (ts *TreeSet) findMin(node *TreeNode) *TreeNode {
    for node.left != nil {
        node = node.left
    }
    return node
}

func (ts *TreeSet) Size() int {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    return ts.size
}

func (ts *TreeSet) Iterate() []interface{} {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    
    var result []interface{}
    ts.inorderTraversal(ts.root, &result)
    return result
}

func (ts *TreeSet) inorderTraversal(node *TreeNode, result *[]interface{}) {
    if node != nil {
        ts.inorderTraversal(node.left, result)
        *result = append(*result, node.value)
        ts.inorderTraversal(node.right, result)
    }
}

// Hash-based set for large collections
type HashSet struct {
    buckets [][]interface{}
    size    int
    capacity int
    mu      sync.RWMutex
}

func NewHashSet() *HashSet {
    capacity := 16
    return &HashSet{
        buckets:  make([][]interface{}, capacity),
        capacity: capacity,
    }
}

func (hs *HashSet) Add(item interface{}) bool {
    hs.mu.Lock()
    defer hs.mu.Unlock()
    
    // Check load factor and resize if needed
    if float64(hs.size) > float64(hs.capacity)*0.75 {
        hs.resize()
    }
    
    hash := hs.hash(item)
    bucketIndex := hash % hs.capacity
    
    // Check if item already exists
    for _, existing := range hs.buckets[bucketIndex] {
        if existing == item {
            return false
        }
    }
    
    hs.buckets[bucketIndex] = append(hs.buckets[bucketIndex], item)
    hs.size++
    return true
}

func (hs *HashSet) resize() {
    oldBuckets := hs.buckets
    hs.capacity *= 2
    hs.buckets = make([][]interface{}, hs.capacity)
    hs.size = 0
    
    for _, bucket := range oldBuckets {
        for _, item := range bucket {
            hs.addWithoutResize(item)
        }
    }
}

func (hs *HashSet) addWithoutResize(item interface{}) {
    hash := hs.hash(item)
    bucketIndex := hash % hs.capacity
    
    hs.buckets[bucketIndex] = append(hs.buckets[bucketIndex], item)
    hs.size++
}

func (hs *HashSet) hash(item interface{}) int {
    str := fmt.Sprintf("%v", item)
    hash := 0
    for _, char := range str {
        hash = hash*31 + int(char)
    }
    if hash < 0 {
        hash = -hash
    }
    return hash
}

func (hs *HashSet) Contains(item interface{}) bool {
    hs.mu.RLock()
    defer hs.mu.RUnlock()
    
    hash := hs.hash(item)
    bucketIndex := hash % hs.capacity
    
    for _, existing := range hs.buckets[bucketIndex] {
        if existing == item {
            return true
        }
    }
    
    return false
}

func (hs *HashSet) Remove(item interface{}) bool {
    hs.mu.Lock()
    defer hs.mu.Unlock()
    
    hash := hs.hash(item)
    bucketIndex := hash % hs.capacity
    
    bucket := hs.buckets[bucketIndex]
    for i, existing := range bucket {
        if existing == item {
            // Remove by swapping with last element
            bucket[i] = bucket[len(bucket)-1]
            hs.buckets[bucketIndex] = bucket[:len(bucket)-1]
            hs.size--
            return true
        }
    }
    
    return false
}

func (hs *HashSet) Size() int {
    hs.mu.RLock()
    defer hs.mu.RUnlock()
    return hs.size
}

func (hs *HashSet) Iterate() []interface{} {
    hs.mu.RLock()
    defer hs.mu.RUnlock()
    
    result := make([]interface{}, 0, hs.size)
    for _, bucket := range hs.buckets {
        result = append(result, bucket...)
    }
    return result
}

// Adaptive set implementation
func NewAdaptiveSet() *AdaptiveSet {
    return &AdaptiveSet{
        implementation: NewArraySet(),
        threshold: AdaptiveThreshold{
            SmallSetSize:           20,
            MediumSetSize:          1000,
            HighIterationRatio:     0.3,
            HighModificationRatio:  0.1,
        },
    }
}

func (as *AdaptiveSet) Add(item interface{}) bool {
    atomic.AddInt64(&as.metrics.AddCount, 1)
    result := as.implementation.Add(item)
    as.updateSize()
    as.evaluateAndAdapt()
    return result
}

func (as *AdaptiveSet) Remove(item interface{}) bool {
    atomic.AddInt64(&as.metrics.RemoveCount, 1)
    result := as.implementation.Remove(item)
    as.updateSize()
    as.evaluateAndAdapt()
    return result
}

func (as *AdaptiveSet) Contains(item interface{}) bool {
    atomic.AddInt64(&as.metrics.ContainsCount, 1)
    return as.implementation.Contains(item)
}

func (as *AdaptiveSet) Size() int {
    return as.implementation.Size()
}

func (as *AdaptiveSet) Iterate() []interface{} {
    atomic.AddInt64(&as.metrics.IterateCount, 1)
    return as.implementation.Iterate()
}

func (as *AdaptiveSet) updateSize() {
    atomic.StoreInt64(&as.metrics.Size, int64(as.implementation.Size()))
}

func (as *AdaptiveSet) evaluateAndAdapt() {
    now := time.Now()
    if now.Sub(as.metrics.LastEvaluation) < 10*time.Second {
        return // Evaluate at most every 10 seconds
    }
    
    as.metrics.LastEvaluation = now
    
    size := atomic.LoadInt64(&as.metrics.Size)
    totalOps := atomic.LoadInt64(&as.metrics.AddCount) + 
               atomic.LoadInt64(&as.metrics.RemoveCount) +
               atomic.LoadInt64(&as.metrics.ContainsCount) +
               atomic.LoadInt64(&as.metrics.IterateCount)
    
    if totalOps < 100 {
        return // Not enough data
    }
    
    iterationRatio := float64(atomic.LoadInt64(&as.metrics.IterateCount)) / float64(totalOps)
    modificationRatio := float64(atomic.LoadInt64(&as.metrics.AddCount) + atomic.LoadInt64(&as.metrics.RemoveCount)) / float64(totalOps)
    
    newImplementation := as.selectOptimalImplementation(size, iterationRatio, modificationRatio)
    
    if fmt.Sprintf("%T", newImplementation) != fmt.Sprintf("%T", as.implementation) {
        as.migrate(newImplementation)
    }
}

func (as *AdaptiveSet) selectOptimalImplementation(size int64, iterationRatio, modificationRatio float64) SetImplementation {
    if size <= int64(as.threshold.SmallSetSize) {
        return NewArraySet()
    }
    
    if size <= int64(as.threshold.MediumSetSize) {
        if iterationRatio > as.threshold.HighIterationRatio {
            return NewTreeSet() // Good for iteration
        }
        return NewHashSet() // Good for lookups
    }
    
    // Large sets always use hash
    return NewHashSet()
}

func (as *AdaptiveSet) migrate(newImplementation SetImplementation) {
    // Migrate all items to new implementation
    items := as.implementation.Iterate()
    
    for _, item := range items {
        newImplementation.Add(item)
    }
    
    as.implementation = newImplementation
    
    // Reset metrics
    atomic.StoreInt64(&as.metrics.AddCount, 0)
    atomic.StoreInt64(&as.metrics.RemoveCount, 0)
    atomic.StoreInt64(&as.metrics.ContainsCount, 0)
    atomic.StoreInt64(&as.metrics.IterateCount, 0)
}
```

## Parallel Algorithms

### Divide-and-Conquer with Goroutines
Implement parallel algorithms that effectively utilize multiple CPU cores:

```go
package parallel_algorithms

import (
    "context"
    "runtime"
    "sync"
)

// Parallel merge sort with work stealing
type ParallelSorter struct {
    threshold   int
    workerCount int
    workQueue   chan SortTask
    wg          sync.WaitGroup
}

type SortTask struct {
    data   []int
    start  int
    end    int
    result chan<- []int
}

func NewParallelSorter() *ParallelSorter {
    workerCount := runtime.NumCPU()
    return &ParallelSorter{
        threshold:   10000, // Switch to sequential below this size
        workerCount: workerCount,
        workQueue:   make(chan SortTask, workerCount*2),
    }
}

func (ps *ParallelSorter) Sort(data []int) []int {
    if len(data) <= ps.threshold {
        // Use sequential sort for small arrays
        dataCopy := make([]int, len(data))
        copy(dataCopy, data)
        quickSort(dataCopy, 0, len(dataCopy)-1)
        return dataCopy
    }
    
    // Start workers
    for i := 0; i < ps.workerCount; i++ {
        go ps.worker()
    }
    
    result := make(chan []int, 1)
    ps.wg.Add(1)
    
    // Submit initial task
    ps.workQueue <- SortTask{
        data:   data,
        start:  0,
        end:    len(data) - 1,
        result: result,
    }
    
    // Wait for completion and close workers
    ps.wg.Wait()
    close(ps.workQueue)
    
    return <-result
}

func (ps *ParallelSorter) worker() {
    for task := range ps.workQueue {
        ps.processTask(task)
    }
}

func (ps *ParallelSorter) processTask(task SortTask) {
    defer ps.wg.Done()
    
    size := task.end - task.start + 1
    
    if size <= ps.threshold {
        // Sequential sort for small chunks
        chunk := make([]int, size)
        copy(chunk, task.data[task.start:task.end+1])
        quickSort(chunk, 0, len(chunk)-1)
        task.result <- chunk
        return
    }
    
    // Divide task
    mid := task.start + (task.end-task.start)/2
    
    leftResult := make(chan []int, 1)
    rightResult := make(chan []int, 1)
    
    ps.wg.Add(2)
    
    // Submit subtasks
    ps.workQueue <- SortTask{
        data:   task.data,
        start:  task.start,
        end:    mid,
        result: leftResult,
    }
    
    ps.workQueue <- SortTask{
        data:   task.data,
        start:  mid + 1,
        end:    task.end,
        result: rightResult,
    }
    
    // Merge results
    go func() {
        left := <-leftResult
        right := <-rightResult
        merged := mergeParallel(left, right)
        task.result <- merged
    }()
}

func mergeParallel(left, right []int) []int {
    result := make([]int, len(left)+len(right))
    i, j, k := 0, 0, 0
    
    for i < len(left) && j < len(right) {
        if left[i] <= right[j] {
            result[k] = left[i]
            i++
        } else {
            result[k] = right[j]
            j++
        }
        k++
    }
    
    for i < len(left) {
        result[k] = left[i]
        i++
        k++
    }
    
    for j < len(right) {
        result[k] = right[j]
        j++
        k++
    }
    
    return result
}

// Parallel map-reduce framework
type MapReduceJob struct {
    mapperCount  int
    reducerCount int
    chunkSize    int
}

func NewMapReduceJob(mappers, reducers, chunkSize int) *MapReduceJob {
    return &MapReduceJob{
        mapperCount:  mappers,
        reducerCount: reducers,
        chunkSize:    chunkSize,
    }
}

func (mrj *MapReduceJob) Execute(
    data []interface{},
    mapper func(interface{}) (string, interface{}),
    reducer func(string, []interface{}) interface{},
) map[string]interface{} {
    
    // Map phase
    mapOutput := mrj.mapPhase(data, mapper)
    
    // Shuffle phase
    shuffled := mrj.shufflePhase(mapOutput)
    
    // Reduce phase
    return mrj.reducePhase(shuffled, reducer)
}

func (mrj *MapReduceJob) mapPhase(
    data []interface{},
    mapper func(interface{}) (string, interface{}),
) []map[string][]interface{} {
    
    // Create chunks
    chunks := make([][]interface{}, 0)
    for i := 0; i < len(data); i += mrj.chunkSize {
        end := min(i+mrj.chunkSize, len(data))
        chunks = append(chunks, data[i:end])
    }
    
    // Process chunks in parallel
    resultChan := make(chan map[string][]interface{}, len(chunks))
    var wg sync.WaitGroup
    
    // Limit concurrent mappers
    semaphore := make(chan struct{}, mrj.mapperCount)
    
    for _, chunk := range chunks {
        wg.Add(1)
        go func(chunk []interface{}) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            result := make(map[string][]interface{})
            
            for _, item := range chunk {
                key, value := mapper(item)
                result[key] = append(result[key], value)
            }
            
            resultChan <- result
        }(chunk)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect results
    var results []map[string][]interface{}
    for result := range resultChan {
        results = append(results, result)
    }
    
    return results
}

func (mrj *MapReduceJob) shufflePhase(mapOutputs []map[string][]interface{}) map[string][]interface{} {
    shuffled := make(map[string][]interface{})
    
    // Combine all map outputs by key
    for _, mapOutput := range mapOutputs {
        for key, values := range mapOutput {
            shuffled[key] = append(shuffled[key], values...)
        }
    }
    
    return shuffled
}

func (mrj *MapReduceJob) reducePhase(
    shuffled map[string][]interface{},
    reducer func(string, []interface{}) interface{},
) map[string]interface{} {
    
    resultChan := make(chan KeyValue, len(shuffled))
    var wg sync.WaitGroup
    
    // Limit concurrent reducers
    semaphore := make(chan struct{}, mrj.reducerCount)
    
    for key, values := range shuffled {
        wg.Add(1)
        go func(key string, values []interface{}) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            result := reducer(key, values)
            resultChan <- KeyValue{Key: key, Value: result}
        }(key, values)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect final results
    finalResult := make(map[string]interface{})
    for kv := range resultChan {
        finalResult[kv.Key] = kv.Value
    }
    
    return finalResult
}

type KeyValue struct {
    Key   string
    Value interface{}
}

// Parallel graph algorithms
type Graph struct {
    vertices map[int][]int
    weights  map[Edge]int
    mu       sync.RWMutex
}

type Edge struct {
    From, To int
}

func NewGraph() *Graph {
    return &Graph{
        vertices: make(map[int][]int),
        weights:  make(map[Edge]int),
    }
}

func (g *Graph) AddEdge(from, to, weight int) {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    g.vertices[from] = append(g.vertices[from], to)
    g.weights[Edge{from, to}] = weight
}

// Parallel Dijkstra's algorithm
func (g *Graph) ParallelDijkstra(start int) map[int]int {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    distances := make(map[int]int)
    visited := make(map[int]bool)
    
    // Initialize distances
    for vertex := range g.vertices {
        distances[vertex] = math.MaxInt32
    }
    distances[start] = 0
    
    // Priority queue for processing
    pq := NewPriorityQueue()
    pq.Push(start, 0)
    
    // Worker pool for parallel processing
    workerCount := runtime.NumCPU()
    tasks := make(chan DijkstraTask, workerCount*2)
    results := make(chan DijkstraResult, workerCount*2)
    
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                g.processDijkstraTask(task, results)
            }
        }()
    }
    
    for !pq.IsEmpty() {
        current, currentDist := pq.Pop()
        
        if visited[current] {
            continue
        }
        
        visited[current] = true
        
        // Submit neighbors for parallel processing
        for _, neighbor := range g.vertices[current] {
            if !visited[neighbor] {
                edge := Edge{current, neighbor}
                weight := g.weights[edge]
                
                tasks <- DijkstraTask{
                    Current:     current,
                    Neighbor:    neighbor,
                    CurrentDist: currentDist,
                    EdgeWeight:  weight,
                }
            }
        }
        
        // Process results
        for i := 0; i < len(g.vertices[current]); i++ {
            select {
            case result := <-results:
                if result.NewDistance < distances[result.Vertex] {
                    distances[result.Vertex] = result.NewDistance
                    pq.Push(result.Vertex, result.NewDistance)
                }
            case <-time.After(100 * time.Millisecond):
                break
            }
        }
    }
    
    close(tasks)
    wg.Wait()
    close(results)
    
    return distances
}

type DijkstraTask struct {
    Current     int
    Neighbor    int
    CurrentDist int
    EdgeWeight  int
}

type DijkstraResult struct {
    Vertex      int
    NewDistance int
}

func (g *Graph) processDijkstraTask(task DijkstraTask, results chan<- DijkstraResult) {
    newDistance := task.CurrentDist + task.EdgeWeight
    
    results <- DijkstraResult{
        Vertex:      task.Neighbor,
        NewDistance: newDistance,
    }
}

// Simple priority queue implementation
type PriorityQueue struct {
    items []PQItem
    mu    sync.Mutex
}

type PQItem struct {
    Value    int
    Priority int
}

func NewPriorityQueue() *PriorityQueue {
    return &PriorityQueue{
        items: make([]PQItem, 0),
    }
}

func (pq *PriorityQueue) Push(value, priority int) {
    pq.mu.Lock()
    defer pq.mu.Unlock()
    
    item := PQItem{Value: value, Priority: priority}
    pq.items = append(pq.items, item)
    pq.heapifyUp(len(pq.items) - 1)
}

func (pq *PriorityQueue) Pop() (int, int) {
    pq.mu.Lock()
    defer pq.mu.Unlock()
    
    if len(pq.items) == 0 {
        return 0, 0
    }
    
    result := pq.items[0]
    lastIndex := len(pq.items) - 1
    pq.items[0] = pq.items[lastIndex]
    pq.items = pq.items[:lastIndex]
    
    if len(pq.items) > 0 {
        pq.heapifyDown(0)
    }
    
    return result.Value, result.Priority
}

func (pq *PriorityQueue) IsEmpty() bool {
    pq.mu.Lock()
    defer pq.mu.Unlock()
    return len(pq.items) == 0
}

func (pq *PriorityQueue) heapifyUp(index int) {
    for index > 0 {
        parentIndex := (index - 1) / 2
        if pq.items[index].Priority >= pq.items[parentIndex].Priority {
            break
        }
        pq.items[index], pq.items[parentIndex] = pq.items[parentIndex], pq.items[index]
        index = parentIndex
    }
}

func (pq *PriorityQueue) heapifyDown(index int) {
    for {
        leftChild := 2*index + 1
        rightChild := 2*index + 2
        smallest := index
        
        if leftChild < len(pq.items) && pq.items[leftChild].Priority < pq.items[smallest].Priority {
            smallest = leftChild
        }
        
        if rightChild < len(pq.items) && pq.items[rightChild].Priority < pq.items[smallest].Priority {
            smallest = rightChild
        }
        
        if smallest == index {
            break
        }
        
        pq.items[index], pq.items[smallest] = pq.items[smallest], pq.items[index]
        index = smallest
    }
}

// Utility functions
func quickSort(arr []int, low, high int) {
    if low < high {
        pivotIndex := partition(arr, low, high)
        quickSort(arr, low, pivotIndex-1)
        quickSort(arr, pivotIndex+1, high)
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

Algorithm optimization requires deep understanding of computational complexity, careful data structure selection, and strategic use of parallelism. By implementing adaptive algorithms that adjust to input characteristics and usage patterns, we can achieve optimal performance across diverse workloads.

## Key Takeaways

1. **Analyze complexity empirically** - measure real performance, not just theoretical bounds
2. **Select algorithms adaptively** - choose implementations based on input characteristics
3. **Implement hybrid approaches** - combine multiple algorithms for optimal performance
4. **Design cache-friendly structures** - consider memory hierarchy in data layout
5. **Parallelize effectively** - use divide-and-conquer with appropriate granularity
6. **Profile algorithm behavior** - understand performance characteristics under load
7. **Optimize for common cases** - handle frequent patterns efficiently
8. **Balance complexity and maintainability** - avoid premature optimization

Effective algorithmic optimization enables applications to scale efficiently while maintaining predictable performance across varying workloads and input sizes.
