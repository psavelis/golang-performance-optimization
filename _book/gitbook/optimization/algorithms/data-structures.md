# Data Structures Optimization

Master the selection and optimization of data structures in Go for maximum performance across different use cases and access patterns.

## Data Structure Performance Characteristics

### Built-in Go Data Structures

#### Arrays and Slices

```go
// Array - fixed size, stack allocated (when small)
var fixedArray [1000]int

// Slice - dynamic size, heap allocated
dynamicSlice := make([]int, 0, 1000) // capacity pre-allocation

// Performance comparison
func BenchmarkArrayVsSlice(b *testing.B) {
    b.Run("Array", func(b *testing.B) {
        var arr [1000]int
        for i := 0; i < b.N; i++ {
            for j := 0; j < 1000; j++ {
                arr[j] = j
            }
        }
    })
    
    b.Run("Slice", func(b *testing.B) {
        slice := make([]int, 1000)
        for i := 0; i < b.N; i++ {
            for j := 0; j < 1000; j++ {
                slice[j] = j
            }
        }
    })
}
```

#### Maps vs Slices for Lookups

```go
func BenchmarkLookupStrategies(b *testing.B) {
    // Test data
    keys := make([]string, 1000)
    for i := range keys {
        keys[i] = fmt.Sprintf("key_%d", i)
    }
    
    // Map lookup - O(1) average
    b.Run("Map", func(b *testing.B) {
        m := make(map[string]int, len(keys))
        for i, key := range keys {
            m[key] = i
        }
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _ = m[keys[i%len(keys)]]
        }
    })
    
    // Linear search - O(n)
    b.Run("LinearSearch", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            target := keys[i%len(keys)]
            for j, key := range keys {
                if key == target {
                    _ = j
                    break
                }
            }
        }
    })
    
    // Binary search on sorted slice - O(log n)
    b.Run("BinarySearch", func(b *testing.B) {
        sortedKeys := make([]string, len(keys))
        copy(sortedKeys, keys)
        sort.Strings(sortedKeys)
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            target := keys[i%len(keys)]
            sort.SearchStrings(sortedKeys, target)
        }
    })
}
```

## Specialized Data Structures

### Ring Buffer (Circular Buffer)

```go
type RingBuffer struct {
    data   []interface{}
    head   int
    tail   int
    size   int
    length int
}

func NewRingBuffer(size int) *RingBuffer {
    return &RingBuffer{
        data: make([]interface{}, size),
        size: size,
    }
}

func (rb *RingBuffer) Push(item interface{}) bool {
    if rb.length == rb.size {
        return false // Buffer full
    }
    
    rb.data[rb.tail] = item
    rb.tail = (rb.tail + 1) % rb.size
    rb.length++
    return true
}

func (rb *RingBuffer) Pop() (interface{}, bool) {
    if rb.length == 0 {
        return nil, false // Buffer empty
    }
    
    item := rb.data[rb.head]
    rb.head = (rb.head + 1) % rb.size
    rb.length--
    return item, true
}

func (rb *RingBuffer) IsFull() bool {
    return rb.length == rb.size
}

func (rb *RingBuffer) IsEmpty() bool {
    return rb.length == 0
}
```

### Trie (Prefix Tree)

```go
type TrieNode struct {
    children map[rune]*TrieNode
    isEnd    bool
    value    interface{}
}

type Trie struct {
    root *TrieNode
}

func NewTrie() *Trie {
    return &Trie{
        root: &TrieNode{
            children: make(map[rune]*TrieNode),
        },
    }
}

func (t *Trie) Insert(key string, value interface{}) {
    node := t.root
    
    for _, char := range key {
        if _, exists := node.children[char]; !exists {
            node.children[char] = &TrieNode{
                children: make(map[rune]*TrieNode),
            }
        }
        node = node.children[char]
    }
    
    node.isEnd = true
    node.value = value
}

func (t *Trie) Search(key string) (interface{}, bool) {
    node := t.root
    
    for _, char := range key {
        if child, exists := node.children[char]; exists {
            node = child
        } else {
            return nil, false
        }
    }
    
    return node.value, node.isEnd
}

func (t *Trie) StartsWith(prefix string) []string {
    node := t.root
    
    // Navigate to prefix
    for _, char := range prefix {
        if child, exists := node.children[char]; exists {
            node = child
        } else {
            return nil
        }
    }
    
    // Collect all words with this prefix
    var results []string
    t.collectWords(node, prefix, &results)
    return results
}

func (t *Trie) collectWords(node *TrieNode, prefix string, results *[]string) {
    if node.isEnd {
        *results = append(*results, prefix)
    }
    
    for char, child := range node.children {
        t.collectWords(child, prefix+string(char), results)
    }
}
```

### Skip List

```go
const MaxLevel = 16

type SkipListNode struct {
    key     int
    value   interface{}
    forward []*SkipListNode
}

type SkipList struct {
    header *SkipListNode
    level  int
}

func NewSkipList() *SkipList {
    header := &SkipListNode{
        forward: make([]*SkipListNode, MaxLevel),
    }
    
    return &SkipList{
        header: header,
        level:  1,
    }
}

func (sl *SkipList) randomLevel() int {
    level := 1
    for rand.Float32() < 0.5 && level < MaxLevel {
        level++
    }
    return level
}

func (sl *SkipList) Insert(key int, value interface{}) {
    update := make([]*SkipListNode, MaxLevel)
    current := sl.header
    
    // Find insertion point
    for i := sl.level - 1; i >= 0; i-- {
        for current.forward[i] != nil && current.forward[i].key < key {
            current = current.forward[i]
        }
        update[i] = current
    }
    
    current = current.forward[0]
    
    if current == nil || current.key != key {
        newLevel := sl.randomLevel()
        
        if newLevel > sl.level {
            for i := sl.level; i < newLevel; i++ {
                update[i] = sl.header
            }
            sl.level = newLevel
        }
        
        newNode := &SkipListNode{
            key:     key,
            value:   value,
            forward: make([]*SkipListNode, newLevel),
        }
        
        for i := 0; i < newLevel; i++ {
            newNode.forward[i] = update[i].forward[i]
            update[i].forward[i] = newNode
        }
    } else {
        current.value = value
    }
}

func (sl *SkipList) Search(key int) (interface{}, bool) {
    current := sl.header
    
    for i := sl.level - 1; i >= 0; i-- {
        for current.forward[i] != nil && current.forward[i].key < key {
            current = current.forward[i]
        }
    }
    
    current = current.forward[0]
    
    if current != nil && current.key == key {
        return current.value, true
    }
    
    return nil, false
}
```

## Cache-Friendly Data Structures

### Structure of Arrays (SoA) vs Array of Structures (AoS)

```go
// Array of Structures (AoS) - cache unfriendly for selective access
type Point3D struct {
    X, Y, Z float64
}

type AoSPoints struct {
    points []Point3D
}

func (aos *AoSPoints) SumX() float64 {
    sum := 0.0
    for _, point := range aos.points {
        sum += point.X // Loads entire struct, wastes cache on Y, Z
    }
    return sum
}

// Structure of Arrays (SoA) - cache friendly for selective access
type SoAPoints struct {
    X []float64
    Y []float64
    Z []float64
}

func (soa *SoAPoints) SumX() float64 {
    sum := 0.0
    for _, x := range soa.X { // Only loads X values, better cache utilization
        sum += x
    }
    return sum
}

func BenchmarkDataLayout(b *testing.B) {
    const size = 10000
    
    // AoS setup
    aosPoints := &AoSPoints{
        points: make([]Point3D, size),
    }
    for i := range aosPoints.points {
        aosPoints.points[i] = Point3D{
            X: float64(i),
            Y: float64(i * 2),
            Z: float64(i * 3),
        }
    }
    
    // SoA setup
    soaPoints := &SoAPoints{
        X: make([]float64, size),
        Y: make([]float64, size),
        Z: make([]float64, size),
    }
    for i := 0; i < size; i++ {
        soaPoints.X[i] = float64(i)
        soaPoints.Y[i] = float64(i * 2)
        soaPoints.Z[i] = float64(i * 3)
    }
    
    b.Run("AoS", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = aosPoints.SumX()
        }
    })
    
    b.Run("SoA", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = soaPoints.SumX()
        }
    })
}
```

### Memory Pool for Object Reuse

```go
type ObjectPool struct {
    pool sync.Pool
    new  func() interface{}
}

func NewObjectPool(newFunc func() interface{}) *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{
            New: newFunc,
        },
        new: newFunc,
    }
}

func (op *ObjectPool) Get() interface{} {
    return op.pool.Get()
}

func (op *ObjectPool) Put(obj interface{}) {
    op.pool.Put(obj)
}

// Example usage with byte buffers
var bufferPool = NewObjectPool(func() interface{} {
    return make([]byte, 0, 1024)
})

func processWithPool(data []byte) []byte {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0]) // Reset length but keep capacity
    
    // Process data using buf
    buf = append(buf, data...)
    buf = append(buf, []byte(" processed")...)
    
    // Return copy since we're returning the buffer to pool
    result := make([]byte, len(buf))
    copy(result, buf)
    return result
}
```

## Hash Table Optimization

### Custom Hash Function

```go
type FastStringMap struct {
    buckets [][]KeyValue
    size    int
    mask    int
}

type KeyValue struct {
    Key   string
    Value interface{}
}

func NewFastStringMap(initialSize int) *FastStringMap {
    // Ensure power of 2 for fast modulo
    size := 1
    for size < initialSize {
        size <<= 1
    }
    
    return &FastStringMap{
        buckets: make([][]KeyValue, size),
        size:    size,
        mask:    size - 1,
    }
}

// Fast hash function for strings
func (fsm *FastStringMap) hash(key string) uint32 {
    hash := uint32(2166136261) // FNV offset basis
    for i := 0; i < len(key); i++ {
        hash ^= uint32(key[i])
        hash *= 16777619 // FNV prime
    }
    return hash
}

func (fsm *FastStringMap) Set(key string, value interface{}) {
    hash := fsm.hash(key)
    bucketIndex := hash & uint32(fsm.mask)
    bucket := fsm.buckets[bucketIndex]
    
    // Check if key exists
    for i := range bucket {
        if bucket[i].Key == key {
            bucket[i].Value = value
            return
        }
    }
    
    // Add new key-value pair
    fsm.buckets[bucketIndex] = append(bucket, KeyValue{
        Key:   key,
        Value: value,
    })
}

func (fsm *FastStringMap) Get(key string) (interface{}, bool) {
    hash := fsm.hash(key)
    bucketIndex := hash & uint32(fsm.mask)
    bucket := fsm.buckets[bucketIndex]
    
    for i := range bucket {
        if bucket[i].Key == key {
            return bucket[i].Value, true
        }
    }
    
    return nil, false
}
```

## Performance Analysis

### Benchmarking Data Structure Operations

```go
func BenchmarkDataStructureOperations(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Map_Size_%d", size), func(b *testing.B) {
            m := make(map[int]int, size)
            
            b.Run("Insert", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    m[i%size] = i
                }
            })
            
            b.Run("Lookup", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    _ = m[i%size]
                }
            })
        })
        
        b.Run(fmt.Sprintf("Slice_Size_%d", size), func(b *testing.B) {
            slice := make([]int, size)
            
            b.Run("Access", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    _ = slice[i%size]
                }
            })
            
            b.Run("Append", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    slice = append(slice, i)
                }
            })
        })
    }
}
```

## Data Structure Selection Guidelines

### Decision Matrix

| Use Case | Best Choice | Time Complexity | Space Efficiency |
|----------|-------------|-----------------|------------------|
| Fast lookups by key | Map | O(1) avg | Medium |
| Ordered iteration | Slice | O(1) access | High |
| Prefix matching | Trie | O(m) where m=key length | Low |
| Range queries | Skip List | O(log n) | Medium |
| FIFO operations | Ring Buffer | O(1) | High |
| Frequent insertions/deletions | Linked List | O(1) at position | High |

### Performance Trade-offs

```go
// Example: Choosing between different set implementations

// Map-based set - fast membership testing
type MapSet struct {
    data map[string]struct{}
}

func (ms *MapSet) Add(item string) {
    ms.data[item] = struct{}{}
}

func (ms *MapSet) Contains(item string) bool {
    _, exists := ms.data[item]
    return exists
}

// Slice-based set - memory efficient for small sets
type SliceSet struct {
    data []string
}

func (ss *SliceSet) Add(item string) {
    if !ss.Contains(item) {
        ss.data = append(ss.data, item)
    }
}

func (ss *SliceSet) Contains(item string) bool {
    for _, existing := range ss.data {
        if existing == item {
            return true
        }
    }
    return false
}

// Bit set - extremely memory efficient for integer sets
type BitSet struct {
    bits []uint64
}

func NewBitSet(size int) *BitSet {
    return &BitSet{
        bits: make([]uint64, (size+63)/64),
    }
}

func (bs *BitSet) Set(bit int) {
    bs.bits[bit/64] |= 1 << (bit % 64)
}

func (bs *BitSet) IsSet(bit int) bool {
    return bs.bits[bit/64]&(1<<(bit%64)) != 0
}
```

## Memory Layout Optimization

### Struct Field Ordering

```go
// Inefficient - 24 bytes due to padding
type BadStruct struct {
    a bool    // 1 byte + 7 bytes padding
    b int64   // 8 bytes
    c bool    // 1 byte + 7 bytes padding
}

// Efficient - 16 bytes with proper ordering
type GoodStruct struct {
    b int64   // 8 bytes
    a bool    // 1 byte
    c bool    // 1 byte + 6 bytes padding
}

// Even better - 10 bytes with packed booleans
type BestStruct struct {
    b int64   // 8 bytes
    flags byte // Pack multiple booleans into single byte
}

func (bs *BestStruct) SetA(value bool) {
    if value {
        bs.flags |= 1
    } else {
        bs.flags &^= 1
    }
}

func (bs *BestStruct) GetA() bool {
    return bs.flags&1 != 0
}

func (bs *BestStruct) SetC(value bool) {
    if value {
        bs.flags |= 2
    } else {
        bs.flags &^= 2
    }
}

func (bs *BestStruct) GetC() bool {
    return bs.flags&2 != 0
}
```

## Best Practices

### 1. Choose Data Structures Based on Access Patterns
- **Random access**: Slices, arrays
- **Key-based lookup**: Maps
- **Ordered traversal**: Sorted slices, trees
- **LIFO operations**: Slices (as stacks)
- **FIFO operations**: Ring buffers, queues

### 2. Pre-allocate When Size is Known
```go
// Inefficient - multiple reallocations
var result []int
for i := 0; i < 1000; i++ {
    result = append(result, i)
}

// Efficient - single allocation
result := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    result = append(result, i)
}
```

### 3. Use Sync.Pool for Expensive Objects
```go
var expensiveObjectPool = sync.Pool{
    New: func() interface{} {
        return &ExpensiveObject{
            data: make([]byte, 64*1024), // Large allocation
        }
    },
}

func processData(input []byte) []byte {
    obj := expensiveObjectPool.Get().(*ExpensiveObject)
    defer expensiveObjectPool.Put(obj)
    
    return obj.Process(input)
}
```

### 4. Consider Cache Locality
- Group related data together
- Use structure of arrays for bulk operations
- Minimize pointer chasing
- Align data structures to cache line boundaries

The key to effective data structure optimization is understanding your specific use case, access patterns, and performance requirements, then selecting and customizing the most appropriate data structure for optimal performance.
