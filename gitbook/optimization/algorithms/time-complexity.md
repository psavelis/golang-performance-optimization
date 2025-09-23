# Time Complexity Optimization

Master algorithmic time complexity analysis and optimization techniques to minimize execution time and improve application performance.

## Understanding Time Complexity

### Big O Notation Fundamentals

Time complexity describes how an algorithm's runtime scales with input size:

```go
// O(1) - Constant time
func getElement(slice []int, index int) int {
    return slice[index] // Always takes same time regardless of slice size
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

// O(n) - Linear time
func linearSearch(slice []int, target int) int {
    for i, value := range slice {
        if value == target {
            return i
        }
    }
    return -1
}

// O(n log n) - Linearithmic time
func mergeSort(arr []int) []int {
    if len(arr) <= 1 {
        return arr
    }
    
    mid := len(arr) / 2
    left := mergeSort(arr[:mid])
    right := mergeSort(arr[mid:])
    
    return merge(left, right)
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

// O(2^n) - Exponential time (inefficient)
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
```

### Analyzing Real-World Algorithms

```go
func BenchmarkComplexityComparison(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        data := generateSortedData(size)
        target := data[size/2] // Middle element
        
        b.Run(fmt.Sprintf("Linear_n_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                linearSearch(data, target)
            }
        })
        
        b.Run(fmt.Sprintf("Binary_log_n_%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                binarySearch(data, target)
            }
        })
        
        b.Run(fmt.Sprintf("Hash_1_%d", size), func(b *testing.B) {
            hashMap := make(map[int]int, size)
            for i, v := range data {
                hashMap[v] = i
            }
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = hashMap[target]
            }
        })
    }
}
```

## Optimization Techniques by Complexity Class

### From O(n²) to O(n) Optimizations

#### Nested Loop Elimination

```go
// O(n²) - Inefficient nested loops
func findPairsSlow(nums []int, target int) [][]int {
    var pairs [][]int
    
    for i := 0; i < len(nums); i++ {
        for j := i + 1; j < len(nums); j++ {
            if nums[i]+nums[j] == target {
                pairs = append(pairs, []int{nums[i], nums[j]})
            }
        }
    }
    
    return pairs
}

// O(n) - Hash map optimization
func findPairsFast(nums []int, target int) [][]int {
    seen := make(map[int]int)
    var pairs [][]int
    
    for i, num := range nums {
        complement := target - num
        if _, exists := seen[complement]; exists {
            pairs = append(pairs, []int{complement, num})
        }
        seen[num] = i
    }
    
    return pairs
}
```

#### String Processing Optimization

```go
// O(n²) - String concatenation
func buildStringSlow(parts []string) string {
    result := ""
    for _, part := range parts {
        result += part // Each concatenation creates new string
    }
    return result
}

// O(n) - StringBuilder optimization
func buildStringFast(parts []string) string {
    var builder strings.Builder
    builder.Grow(estimateSize(parts)) // Pre-allocate capacity
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    
    return builder.String()
}

func estimateSize(parts []string) int {
    total := 0
    for _, part := range parts {
        total += len(part)
    }
    return total
}
```

### From O(n) to O(log n) Optimizations

#### Binary Search Applications

```go
// Generic binary search for any comparable type
func binarySearchGeneric[T any](arr []T, target T, compare func(T, T) int) int {
    left, right := 0, len(arr)-1
    
    for left <= right {
        mid := left + (right-left)/2
        cmp := compare(arr[mid], target)
        
        if cmp == 0 {
            return mid
        } else if cmp < 0 {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    
    return -1
}

// Finding insertion point
func searchInsertPosition(nums []int, target int) int {
    left, right := 0, len(nums)
    
    for left < right {
        mid := left + (right-left)/2
        if nums[mid] < target {
            left = mid + 1
        } else {
            right = mid
        }
    }
    
    return left
}

// Binary search on ranges
func findFirstOccurrence(nums []int, target int) int {
    left, right := 0, len(nums)-1
    result := -1
    
    for left <= right {
        mid := left + (right-left)/2
        if nums[mid] == target {
            result = mid
            right = mid - 1 // Continue searching left
        } else if nums[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    
    return result
}
```

#### Tree-Based Data Structures

```go
type AVLNode struct {
    key    int
    value  interface{}
    height int
    left   *AVLNode
    right  *AVLNode
}

type AVLTree struct {
    root *AVLNode
}

func (tree *AVLTree) Height(node *AVLNode) int {
    if node == nil {
        return 0
    }
    return node.height
}

func (tree *AVLTree) Balance(node *AVLNode) int {
    if node == nil {
        return 0
    }
    return tree.Height(node.left) - tree.Height(node.right)
}

func (tree *AVLTree) UpdateHeight(node *AVLNode) {
    if node != nil {
        leftHeight := tree.Height(node.left)
        rightHeight := tree.Height(node.right)
        if leftHeight > rightHeight {
            node.height = leftHeight + 1
        } else {
            node.height = rightHeight + 1
        }
    }
}

func (tree *AVLTree) RotateRight(y *AVLNode) *AVLNode {
    x := y.left
    T2 := x.right
    
    x.right = y
    y.left = T2
    
    tree.UpdateHeight(y)
    tree.UpdateHeight(x)
    
    return x
}

func (tree *AVLTree) RotateLeft(x *AVLNode) *AVLNode {
    y := x.right
    T2 := y.left
    
    y.left = x
    x.right = T2
    
    tree.UpdateHeight(x)
    tree.UpdateHeight(y)
    
    return y
}

func (tree *AVLTree) Insert(node *AVLNode, key int, value interface{}) *AVLNode {
    // Standard BST insertion
    if node == nil {
        return &AVLNode{
            key:    key,
            value:  value,
            height: 1,
        }
    }
    
    if key < node.key {
        node.left = tree.Insert(node.left, key, value)
    } else if key > node.key {
        node.right = tree.Insert(node.right, key, value)
    } else {
        node.value = value
        return node
    }
    
    // Update height
    tree.UpdateHeight(node)
    
    // Get balance factor
    balance := tree.Balance(node)
    
    // Left Left Case
    if balance > 1 && key < node.left.key {
        return tree.RotateRight(node)
    }
    
    // Right Right Case
    if balance < -1 && key > node.right.key {
        return tree.RotateLeft(node)
    }
    
    // Left Right Case
    if balance > 1 && key > node.left.key {
        node.left = tree.RotateLeft(node.left)
        return tree.RotateRight(node)
    }
    
    // Right Left Case
    if balance < -1 && key < node.right.key {
        node.right = tree.RotateRight(node.right)
        return tree.RotateLeft(node)
    }
    
    return node
}
```

### From O(2^n) to O(n) with Dynamic Programming

#### Memoization Pattern

```go
// O(2^n) - Exponential time without memoization
func fibonacciSlow(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacciSlow(n-1) + fibonacciSlow(n-2)
}

// O(n) - Linear time with memoization
type FibCalculator struct {
    memo map[int]int
}

func NewFibCalculator() *FibCalculator {
    return &FibCalculator{
        memo: make(map[int]int),
    }
}

func (fc *FibCalculator) Calculate(n int) int {
    if n <= 1 {
        return n
    }
    
    if val, exists := fc.memo[n]; exists {
        return val
    }
    
    result := fc.Calculate(n-1) + fc.Calculate(n-2)
    fc.memo[n] = result
    return result
}

// O(n) - Bottom-up dynamic programming
func fibonacciFast(n int) int {
    if n <= 1 {
        return n
    }
    
    prev2, prev1 := 0, 1
    
    for i := 2; i <= n; i++ {
        current := prev1 + prev2
        prev2, prev1 = prev1, current
    }
    
    return prev1
}
```

#### Longest Common Subsequence

```go
// O(2^n) - Exponential time naive approach
func lcsSlow(text1, text2 string, i, j int) int {
    if i == len(text1) || j == len(text2) {
        return 0
    }
    
    if text1[i] == text2[j] {
        return 1 + lcsSlow(text1, text2, i+1, j+1)
    }
    
    return max(lcsSlow(text1, text2, i+1, j), lcsSlow(text1, text2, i, j+1))
}

// O(m*n) - Dynamic programming optimization
func lcsFast(text1, text2 string) int {
    m, n := len(text1), len(text2)
    dp := make([][]int, m+1)
    for i := range dp {
        dp[i] = make([]int, n+1)
    }
    
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if text1[i-1] == text2[j-1] {
                dp[i][j] = dp[i-1][j-1] + 1
            } else {
                dp[i][j] = max(dp[i-1][j], dp[i][j-1])
            }
        }
    }
    
    return dp[m][n]
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

## Advanced Optimization Techniques

### Divide and Conquer

```go
// O(n log n) - Merge sort implementation
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

// O(n log n) - Quick select for finding kth element
func quickSelect(arr []int, k int) int {
    if len(arr) == 1 {
        return arr[0]
    }
    
    pivot := partition(arr)
    
    if k == pivot {
        return arr[k]
    } else if k < pivot {
        return quickSelect(arr[:pivot], k)
    } else {
        return quickSelect(arr[pivot+1:], k-pivot-1)
    }
}

func partition(arr []int) int {
    pivot := arr[len(arr)-1]
    i := 0
    
    for j := 0; j < len(arr)-1; j++ {
        if arr[j] <= pivot {
            arr[i], arr[j] = arr[j], arr[i]
            i++
        }
    }
    
    arr[i], arr[len(arr)-1] = arr[len(arr)-1], arr[i]
    return i
}
```

### Greedy Algorithms

```go
// O(n log n) - Activity selection problem
type Activity struct {
    start, end int
}

func maxActivities(activities []Activity) []Activity {
    // Sort by end time
    sort.Slice(activities, func(i, j int) bool {
        return activities[i].end < activities[j].end
    })
    
    var result []Activity
    lastEnd := -1
    
    for _, activity := range activities {
        if activity.start >= lastEnd {
            result = append(result, activity)
            lastEnd = activity.end
        }
    }
    
    return result
}

// O(n log n) - Huffman coding for compression
type HuffmanNode struct {
    char   rune
    freq   int
    left   *HuffmanNode
    right  *HuffmanNode
}

type PriorityQueue []*HuffmanNode

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].freq < pq[j].freq
}

func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
    *pq = append(*pq, x.(*HuffmanNode))
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    *pq = old[0 : n-1]
    return item
}

func buildHuffmanTree(frequencies map[rune]int) *HuffmanNode {
    pq := &PriorityQueue{}
    heap.Init(pq)
    
    for char, freq := range frequencies {
        heap.Push(pq, &HuffmanNode{char: char, freq: freq})
    }
    
    for pq.Len() > 1 {
        left := heap.Pop(pq).(*HuffmanNode)
        right := heap.Pop(pq).(*HuffmanNode)
        
        merged := &HuffmanNode{
            freq:  left.freq + right.freq,
            left:  left,
            right: right,
        }
        
        heap.Push(pq, merged)
    }
    
    return heap.Pop(pq).(*HuffmanNode)
}
```

## Amortized Analysis

### Dynamic Array with Amortized O(1) Append

```go
type DynamicArray struct {
    data     []int
    size     int
    capacity int
}

func NewDynamicArray() *DynamicArray {
    return &DynamicArray{
        data:     make([]int, 1),
        capacity: 1,
    }
}

func (da *DynamicArray) Append(value int) {
    if da.size == da.capacity {
        // Double the capacity - amortized O(1)
        newCapacity := da.capacity * 2
        newData := make([]int, newCapacity)
        copy(newData, da.data[:da.size])
        da.data = newData
        da.capacity = newCapacity
    }
    
    da.data[da.size] = value
    da.size++
}

func (da *DynamicArray) Get(index int) int {
    if index < 0 || index >= da.size {
        panic("index out of bounds")
    }
    return da.data[index]
}

func (da *DynamicArray) Size() int {
    return da.size
}
```

## Performance Testing and Analysis

### Benchmarking Time Complexity

```go
func BenchmarkTimeComplexity(b *testing.B) {
    complexities := map[string]func(int) int{
        "O(1)":     constant,
        "O(log n)": logarithmic,
        "O(n)":     linear,
        "O(n²)":    quadratic,
    }
    
    sizes := []int{100, 500, 1000, 5000}
    
    for name, fn := range complexities {
        for _, size := range sizes {
            b.Run(fmt.Sprintf("%s_n_%d", name, size), func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    fn(size)
                }
            })
        }
    }
}

func constant(n int) int {
    return 42
}

func logarithmic(n int) int {
    count := 0
    for i := 1; i < n; i *= 2 {
        count++
    }
    return count
}

func linear(n int) int {
    sum := 0
    for i := 0; i < n; i++ {
        sum += i
    }
    return sum
}

func quadratic(n int) int {
    sum := 0
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            sum += i * j
        }
    }
    return sum
}
```

### Growth Rate Analysis

```go
func analyzeGrowthRate() {
    sizes := []int{10, 100, 1000, 10000}
    
    fmt.Println("Input Size | O(1) | O(log n) | O(n) | O(n log n) | O(n²)")
    fmt.Println("-----------|------|----------|------|------------|-------")
    
    for _, n := range sizes {
        logN := math.Log2(float64(n))
        nLogN := float64(n) * logN
        n2 := float64(n * n)
        
        fmt.Printf("%-11d| %-4d | %-8.1f | %-4d | %-10.0f | %-6.0f\n",
            n, 1, logN, n, nLogN, n2)
    }
}
```

## Optimization Decision Framework

### Algorithm Selection Guidelines

```go
type AlgorithmChoice struct {
    name       string
    complexity string
    bestFor    string
    example    string
}

var algorithmGuide = []AlgorithmChoice{
    {
        name:       "Hash Table",
        complexity: "O(1) average",
        bestFor:    "Fast lookups, insertions, deletions",
        example:    "Caching, indexing, set operations",
    },
    {
        name:       "Binary Search",
        complexity: "O(log n)",
        bestFor:    "Searching sorted data",
        example:    "Finding elements, insertion points",
    },
    {
        name:       "Merge Sort",
        complexity: "O(n log n)",
        bestFor:    "Stable sorting, guaranteed performance",
        example:    "Large datasets, external sorting",
    },
    {
        name:       "Quick Sort",
        complexity: "O(n log n) average",
        bestFor:    "In-place sorting, cache-friendly",
        example:    "General purpose sorting",
    },
    {
        name:       "Dynamic Programming",
        complexity: "O(n) to O(n³)",
        bestFor:    "Optimization problems with overlapping subproblems",
        example:    "Fibonacci, LCS, knapsack",
    },
}

func selectAlgorithm(dataSize int, operations []string) string {
    // Example algorithm selection logic
    if contains(operations, "lookup") && dataSize > 1000 {
        return "Hash Table"
    }
    if contains(operations, "sort") && dataSize > 10000 {
        return "Merge Sort"
    }
    if contains(operations, "search") && isSorted(operations) {
        return "Binary Search"
    }
    return "Linear approach"
}
```

## Best Practices for Time Optimization

### 1. Profile Before Optimizing
```go
func profileExample() {
    // CPU profiling
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Your algorithm here
    timeIntensiveAlgorithm()
}
```

### 2. Use Appropriate Data Structures
- Maps for O(1) lookups
- Slices for sequential access
- Heaps for priority operations
- Trees for ordered operations

### 3. Apply Algorithmic Optimizations
- Reduce nested loops
- Use divide and conquer
- Apply memoization for recursive problems
- Consider greedy approaches when applicable

### 4. Measure and Validate
- Benchmark different approaches
- Use profiling tools
- Test with realistic data sizes
- Monitor production performance

The key to time complexity optimization is understanding the fundamental algorithmic patterns, recognizing optimization opportunities, and systematically applying the most appropriate techniques for your specific use case.
