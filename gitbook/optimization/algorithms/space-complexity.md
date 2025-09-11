# Space Complexity Optimization

Master memory-efficient algorithm design and space complexity analysis to minimize memory usage while maintaining optimal performance.

## Understanding Space Complexity

### Space Complexity Fundamentals

Space complexity measures the memory footprint of algorithms relative to input size:

```go
// O(1) - Constant space
func reverseArrayInPlace(arr []int) {
    left, right := 0, len(arr)-1
    for left < right {
        arr[left], arr[right] = arr[right], arr[left]
        left++
        right--
    }
    // Only uses a fixed amount of extra memory (left, right)
}

// O(n) - Linear space
func reverseArrayCopy(arr []int) []int {
    result := make([]int, len(arr))
    for i, v := range arr {
        result[len(arr)-1-i] = v
    }
    return result
    // Creates new array of size n
}

// O(log n) - Logarithmic space (recursive calls)
func binarySearchRecursive(arr []int, target, left, right int) int {
    if left > right {
        return -1
    }
    
    mid := left + (right-left)/2
    if arr[mid] == target {
        return mid
    } else if arr[mid] < target {
        return binarySearchRecursive(arr, target, mid+1, right)
    } else {
        return binarySearchRecursive(arr, target, left, mid-1)
    }
    // Each recursive call uses stack space - O(log n) depth
}

// O(n) space - Linear recursion depth
func fibonacciRecursive(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacciRecursive(n-1) + fibonacciRecursive(n-2)
    // Maximum call stack depth is n
}
```

### Memory Usage Analysis

```go
func analyzeMemoryUsage() {
    var m runtime.MemStats
    
    // Baseline memory
    runtime.GC()
    runtime.ReadMemStats(&m)
    baseline := m.Alloc
    
    // Allocate memory
    data := make([][]int, 1000)
    for i := range data {
        data[i] = make([]int, 1000)
    }
    
    // Measure memory after allocation
    runtime.ReadMemStats(&m)
    allocated := m.Alloc - baseline
    
    fmt.Printf("Memory allocated: %d bytes\n", allocated)
    fmt.Printf("Memory per element: %.2f bytes\n", float64(allocated)/(1000*1000))
}
```

## Space Optimization Techniques

### In-Place Algorithms

#### Array Manipulation

```go
// In-place array rotation - O(1) space
func rotateArrayInPlace(arr []int, k int) {
    n := len(arr)
    k = k % n
    
    // Reverse entire array
    reverse(arr, 0, n-1)
    
    // Reverse first k elements
    reverse(arr, 0, k-1)
    
    // Reverse remaining elements
    reverse(arr, k, n-1)
}

func reverse(arr []int, start, end int) {
    for start < end {
        arr[start], arr[end] = arr[end], arr[start]
        start++
        end--
    }
}

// In-place removal of duplicates - O(1) space
func removeDuplicatesInPlace(arr []int) int {
    if len(arr) == 0 {
        return 0
    }
    
    writeIndex := 1
    for readIndex := 1; readIndex < len(arr); readIndex++ {
        if arr[readIndex] != arr[readIndex-1] {
            arr[writeIndex] = arr[readIndex]
            writeIndex++
        }
    }
    
    return writeIndex
}
```

#### String Processing

```go
// In-place string reversal using byte slice
func reverseStringInPlace(s string) string {
    runes := []rune(s) // Convert to rune slice for Unicode support
    left, right := 0, len(runes)-1
    
    for left < right {
        runes[left], runes[right] = runes[right], runes[left]
        left++
        right--
    }
    
    return string(runes)
}

// Space-efficient palindrome check - O(1) space
func isPalindrome(s string) bool {
    left, right := 0, len(s)-1
    
    for left < right {
        // Skip non-alphanumeric characters
        for left < right && !isAlphanumeric(s[left]) {
            left++
        }
        for left < right && !isAlphanumeric(s[right]) {
            right--
        }
        
        if toLower(s[left]) != toLower(s[right]) {
            return false
        }
        
        left++
        right--
    }
    
    return true
}

func isAlphanumeric(c byte) bool {
    return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func toLower(c byte) byte {
    if c >= 'A' && c <= 'Z' {
        return c + 32
    }
    return c
}
```

### Iterative vs Recursive Space Usage

#### Fibonacci: Recursive vs Iterative

```go
// O(n) space - recursive with memoization
func fibonacciMemo(n int, memo map[int]int) int {
    if n <= 1 {
        return n
    }
    
    if val, exists := memo[n]; exists {
        return val
    }
    
    result := fibonacciMemo(n-1, memo) + fibonacciMemo(n-2, memo)
    memo[n] = result
    return result
}

// O(1) space - iterative approach
func fibonacciIterative(n int) int {
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

// Space comparison benchmark
func BenchmarkFibonacciSpace(b *testing.B) {
    n := 40
    
    b.Run("Recursive", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            memo := make(map[int]int)
            fibonacciMemo(n, memo)
        }
    })
    
    b.Run("Iterative", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            fibonacciIterative(n)
        }
    })
}
```

#### Tree Traversal: Recursive vs Iterative

```go
type TreeNode struct {
    Val   int
    Left  *TreeNode
    Right *TreeNode
}

// O(h) space where h is height - recursive
func inorderRecursive(root *TreeNode) []int {
    var result []int
    
    var dfs func(*TreeNode)
    dfs = func(node *TreeNode) {
        if node == nil {
            return
        }
        
        dfs(node.Left)
        result = append(result, node.Val)
        dfs(node.Right)
    }
    
    dfs(root)
    return result
}

// O(h) space - iterative with explicit stack
func inorderIterative(root *TreeNode) []int {
    var result []int
    var stack []*TreeNode
    current := root
    
    for current != nil || len(stack) > 0 {
        // Go to leftmost node
        for current != nil {
            stack = append(stack, current)
            current = current.Left
        }
        
        // Pop from stack and process
        current = stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        result = append(result, current.Val)
        
        // Move to right subtree
        current = current.Right
    }
    
    return result
}

// Morris Traversal - O(1) space
func inorderMorris(root *TreeNode) []int {
    var result []int
    current := root
    
    for current != nil {
        if current.Left == nil {
            result = append(result, current.Val)
            current = current.Right
        } else {
            // Find inorder predecessor
            predecessor := current.Left
            for predecessor.Right != nil && predecessor.Right != current {
                predecessor = predecessor.Right
            }
            
            if predecessor.Right == nil {
                // Make thread
                predecessor.Right = current
                current = current.Left
            } else {
                // Remove thread
                predecessor.Right = nil
                result = append(result, current.Val)
                current = current.Right
            }
        }
    }
    
    return result
}
```

## Memory Pool and Object Reuse

### Buffer Pooling

```go
// Efficient buffer management with sync.Pool
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024) // Start with 1KB capacity
    },
}

func processDataWithPool(data []byte) []byte {
    // Get buffer from pool
    buf := bufferPool.Get().([]byte)
    defer func() {
        // Reset and return to pool
        buf = buf[:0] // Reset length but keep capacity
        bufferPool.Put(buf)
    }()
    
    // Use buffer for processing
    buf = append(buf, data...)
    buf = append(buf, []byte(" processed")...)
    
    // Return copy since we're returning buffer to pool
    result := make([]byte, len(buf))
    copy(result, buf)
    return result
}

// String builder pooling
var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

func concatenateStringsWithPool(parts []string) string {
    builder := stringBuilderPool.Get().(*strings.Builder)
    defer func() {
        builder.Reset()
        stringBuilderPool.Put(builder)
    }()
    
    // Estimate total size to avoid reallocations
    totalSize := 0
    for _, part := range parts {
        totalSize += len(part)
    }
    builder.Grow(totalSize)
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    
    return builder.String()
}
```

### Object Recycling

```go
// Reusable slice structure
type SlicePool struct {
    pool sync.Pool
}

func NewSlicePool(initialCap int) *SlicePool {
    return &SlicePool{
        pool: sync.Pool{
            New: func() interface{} {
                return make([]int, 0, initialCap)
            },
        },
    }
}

func (sp *SlicePool) Get() []int {
    return sp.pool.Get().([]int)
}

func (sp *SlicePool) Put(slice []int) {
    // Reset slice but keep capacity
    slice = slice[:0]
    sp.pool.Put(slice)
}

// Usage example
func processWithSlicePool(pool *SlicePool, data []int) []int {
    temp := pool.Get()
    defer pool.Put(temp)
    
    // Process data using temp slice
    for _, v := range data {
        if v > 0 {
            temp = append(temp, v*2)
        }
    }
    
    // Return copy
    result := make([]int, len(temp))
    copy(result, temp)
    return result
}
```

## Streaming and Lazy Evaluation

### Streaming Data Processing

```go
// Memory-efficient file processing
func processLargeFileStreaming(filename string, processor func(string) string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    outputFile, err := os.Create(filename + ".processed")
    if err != nil {
        return err
    }
    defer outputFile.Close()
    
    scanner := bufio.NewScanner(file)
    writer := bufio.NewWriter(outputFile)
    defer writer.Flush()
    
    // Process line by line - O(1) space relative to file size
    for scanner.Scan() {
        line := scanner.Text()
        processed := processor(line)
        writer.WriteString(processed + "\n")
    }
    
    return scanner.Err()
}

// Channel-based streaming
func processStreamWithChannels(input <-chan string, output chan<- string) {
    defer close(output)
    
    for data := range input {
        // Process each item individually - constant memory
        processed := strings.ToUpper(data)
        output <- processed
    }
}
```

### Lazy Data Structures

```go
// Lazy list implementation
type LazyList struct {
    head interface{}
    tail func() *LazyList
}

func (ll *LazyList) Head() interface{} {
    return ll.head
}

func (ll *LazyList) Tail() *LazyList {
    if ll.tail == nil {
        return nil
    }
    return ll.tail()
}

func (ll *LazyList) Take(n int) []interface{} {
    var result []interface{}
    current := ll
    
    for i := 0; i < n && current != nil; i++ {
        result = append(result, current.Head())
        current = current.Tail()
    }
    
    return result
}

// Infinite sequence generator - constant space
func infiniteNumbers(start int) *LazyList {
    return &LazyList{
        head: start,
        tail: func() *LazyList {
            return infiniteNumbers(start + 1)
        },
    }
}

// Generator pattern for memory efficiency
func fibonacciGenerator() func() int {
    a, b := 0, 1
    return func() int {
        result := a
        a, b = b, a+b
        return result
    }
}
```

## Cache-Conscious Data Layouts

### Structure of Arrays (SoA) vs Array of Structures (AoS)

```go
// Array of Structures - less cache friendly
type Particle struct {
    X, Y, Z    float64 // Position
    VX, VY, VZ float64 // Velocity
}

type AoSParticles struct {
    particles []Particle
}

func (aos *AoSParticles) UpdatePositions(dt float64) {
    for i := range aos.particles {
        // Loads entire struct for each particle
        aos.particles[i].X += aos.particles[i].VX * dt
        aos.particles[i].Y += aos.particles[i].VY * dt
        aos.particles[i].Z += aos.particles[i].VZ * dt
    }
}

// Structure of Arrays - more cache friendly
type SoAParticles struct {
    X, Y, Z    []float64 // Positions
    VX, VY, VZ []float64 // Velocities
    count      int
}

func NewSoAParticles(capacity int) *SoAParticles {
    return &SoAParticles{
        X:  make([]float64, capacity),
        Y:  make([]float64, capacity),
        Z:  make([]float64, capacity),
        VX: make([]float64, capacity),
        VY: make([]float64, capacity),
        VZ: make([]float64, capacity),
    }
}

func (soa *SoAParticles) UpdatePositions(dt float64) {
    // Better cache locality - processes similar data together
    for i := 0; i < soa.count; i++ {
        soa.X[i] += soa.VX[i] * dt
    }
    for i := 0; i < soa.count; i++ {
        soa.Y[i] += soa.VY[i] * dt
    }
    for i := 0; i < soa.count; i++ {
        soa.Z[i] += soa.VZ[i] * dt
    }
}

// Benchmark cache performance
func BenchmarkDataLayout(b *testing.B) {
    const numParticles = 10000
    
    // AoS setup
    aosParticles := &AoSParticles{
        particles: make([]Particle, numParticles),
    }
    
    // SoA setup
    soaParticles := NewSoAParticles(numParticles)
    soaParticles.count = numParticles
    
    b.Run("AoS", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            aosParticles.UpdatePositions(0.016) // 60 FPS
        }
    })
    
    b.Run("SoA", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            soaParticles.UpdatePositions(0.016) // 60 FPS
        }
    })
}
```

## Memory-Efficient Data Structures

### Compressed Data Structures

```go
// Bit vector for boolean arrays
type BitVector struct {
    bits []uint64
    size int
}

func NewBitVector(size int) *BitVector {
    numWords := (size + 63) / 64
    return &BitVector{
        bits: make([]uint64, numWords),
        size: size,
    }
}

func (bv *BitVector) Set(index int) {
    wordIndex := index / 64
    bitIndex := index % 64
    bv.bits[wordIndex] |= 1 << bitIndex
}

func (bv *BitVector) Clear(index int) {
    wordIndex := index / 64
    bitIndex := index % 64
    bv.bits[wordIndex] &^= 1 << bitIndex
}

func (bv *BitVector) Get(index int) bool {
    wordIndex := index / 64
    bitIndex := index % 64
    return bv.bits[wordIndex]&(1<<bitIndex) != 0
}

func (bv *BitVector) PopCount() int {
    count := 0
    for _, word := range bv.bits {
        count += bits.OnesCount64(word)
    }
    return count
}

// Packed integer arrays
type PackedInts struct {
    data     []uint64
    bitsPerValue int
    mask     uint64
    size     int
}

func NewPackedInts(bitsPerValue, capacity int) *PackedInts {
    valuesPerWord := 64 / bitsPerValue
    numWords := (capacity + valuesPerWord - 1) / valuesPerWord
    
    return &PackedInts{
        data:         make([]uint64, numWords),
        bitsPerValue: bitsPerValue,
        mask:         (1 << bitsPerValue) - 1,
    }
}

func (pi *PackedInts) Set(index int, value uint64) {
    valuesPerWord := 64 / pi.bitsPerValue
    wordIndex := index / valuesPerWord
    bitOffset := (index % valuesPerWord) * pi.bitsPerValue
    
    // Clear existing value
    pi.data[wordIndex] &^= pi.mask << bitOffset
    
    // Set new value
    pi.data[wordIndex] |= (value & pi.mask) << bitOffset
}

func (pi *PackedInts) Get(index int) uint64 {
    valuesPerWord := 64 / pi.bitsPerValue
    wordIndex := index / valuesPerWord
    bitOffset := (index % valuesPerWord) * pi.bitsPerValue
    
    return (pi.data[wordIndex] >> bitOffset) & pi.mask
}
```

### Sparse Data Structures

```go
// Sparse matrix using map
type SparseMatrix struct {
    data map[int]map[int]float64
    rows, cols int
}

func NewSparseMatrix(rows, cols int) *SparseMatrix {
    return &SparseMatrix{
        data: make(map[int]map[int]float64),
        rows: rows,
        cols: cols,
    }
}

func (sm *SparseMatrix) Set(row, col int, value float64) {
    if value == 0 {
        // Remove zero values to save space
        if rowMap, exists := sm.data[row]; exists {
            delete(rowMap, col)
            if len(rowMap) == 0 {
                delete(sm.data, row)
            }
        }
        return
    }
    
    if _, exists := sm.data[row]; !exists {
        sm.data[row] = make(map[int]float64)
    }
    sm.data[row][col] = value
}

func (sm *SparseMatrix) Get(row, col int) float64 {
    if rowMap, exists := sm.data[row]; exists {
        return rowMap[col] // Returns 0 if not found
    }
    return 0
}

func (sm *SparseMatrix) NonZeroCount() int {
    count := 0
    for _, rowMap := range sm.data {
        count += len(rowMap)
    }
    return count
}
```

## Performance Monitoring

### Memory Usage Tracking

```go
func trackMemoryUsage(name string, fn func()) {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    fn()
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("%s:\n", name)
    fmt.Printf("  Allocs: %d\n", m2.Allocs-m1.Allocs)
    fmt.Printf("  Total Allocated: %d bytes\n", m2.TotalAlloc-m1.TotalAlloc)
    fmt.Printf("  Heap Objects: %d\n", m2.HeapObjects-m1.HeapObjects)
    fmt.Printf("  Heap Size: %d bytes\n", m2.HeapSys-m1.HeapSys)
}

// Benchmark memory allocations
func BenchmarkMemoryAllocations(b *testing.B) {
    b.Run("SliceAppend", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var slice []int
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
    
    b.Run("SlicePrealloc", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            slice := make([]int, 0, 1000)
            for j := 0; j < 1000; j++ {
                slice = append(slice, j)
            }
        }
    })
}
```

## Best Practices for Space Optimization

### 1. Choose Appropriate Data Structures
- Use bit vectors for boolean arrays
- Use sparse structures for mostly empty data
- Use packed integers for small value ranges
- Use sync.Pool for temporary objects

### 2. Minimize Allocations
- Pre-allocate slices and maps when size is known
- Reuse buffers and temporary objects
- Use in-place algorithms when possible
- Avoid unnecessary string concatenations

### 3. Memory Layout Optimization
- Order struct fields by size (largest first)
- Use SoA layout for bulk operations
- Align data to cache line boundaries
- Pack boolean flags into bit fields

### 4. Lazy and Streaming Approaches
- Use generators for large sequences
- Process data in chunks
- Implement lazy evaluation for expensive computations
- Use channels for pipeline processing

The key to effective space optimization is understanding your data access patterns, choosing appropriate algorithms and data structures, and continuously monitoring memory usage in production environments.
