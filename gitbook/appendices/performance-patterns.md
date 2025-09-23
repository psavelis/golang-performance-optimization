# Performance Patterns

Collection of proven performance patterns, anti-patterns, and best practices for Go applications with practical examples and benchmarks.

## Core Performance Patterns

### Pre-allocation Pattern

**Pattern:** Allocate slices and maps with known capacity to avoid reallocations.

```go
// ❌ Anti-pattern: Growing slice
func processItemsBad(items []Item) []Result {
    var results []Result // Zero capacity
    
    for _, item := range items {
        result := processItem(item)
        results = append(results, result) // May trigger multiple reallocations
    }
    
    return results
}

// ✅ Pattern: Pre-allocated slice
func processItemsGood(items []Item) []Result {
    results := make([]Result, 0, len(items)) // Pre-allocate capacity
    
    for _, item := range items {
        result := processItem(item)
        results = append(results, result) // No reallocations
    }
    
    return results
}

// Benchmark comparison
func BenchmarkPreallocation(b *testing.B) {
    items := make([]Item, 1000)
    for i := range items {
        items[i] = Item{ID: i}
    }
    
    b.Run("WithoutPreallocation", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            results := processItemsBad(items)
            _ = results
        }
    })
    
    b.Run("WithPreallocation", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            results := processItemsGood(items)
            _ = results
        }
    })
}
```

**Results:** Pre-allocation typically reduces allocations by 10-100x and improves performance by 2-5x.

### String Builder Pattern

**Pattern:** Use `strings.Builder` for efficient string concatenation.

```go
// ❌ Anti-pattern: String concatenation
func buildMessageBad(parts []string) string {
    var message string
    
    for _, part := range parts {
        message += part + " " // Creates new string each time
    }
    
    return strings.TrimSpace(message)
}

// ✅ Pattern: strings.Builder
func buildMessageGood(parts []string) string {
    var builder strings.Builder
    
    // Pre-allocate if size is known
    totalLen := 0
    for _, part := range parts {
        totalLen += len(part) + 1 // +1 for space
    }
    builder.Grow(totalLen)
    
    for i, part := range parts {
        if i > 0 {
            builder.WriteByte(' ')
        }
        builder.WriteString(part)
    }
    
    return builder.String()
}

// ✅ Alternative: Join for simple cases
func buildMessageSimple(parts []string) string {
    return strings.Join(parts, " ")
}
```

### Object Pool Pattern

**Pattern:** Reuse expensive objects to reduce allocations and GC pressure.

```go
// Object pool for buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 4096) // 4KB initial capacity
    },
}

// ✅ Pattern: Object pooling
func processDataWithPool(data []byte) []byte {
    buffer := bufferPool.Get().([]byte)
    defer func() {
        buffer = buffer[:0] // Reset length, keep capacity
        bufferPool.Put(buffer)
    }()
    
    // Use buffer for processing
    buffer = append(buffer, data...)
    
    // Apply transformations
    for i := range buffer {
        buffer[i] = transform(buffer[i])
    }
    
    // Return copy since buffer will be reused
    result := make([]byte, len(buffer))
    copy(result, buffer)
    return result
}

// Custom pool with type safety
type BufferPool struct {
    pool sync.Pool
}

func NewBufferPool(initialSize int) *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, initialSize)
            },
        },
    }
}

func (bp *BufferPool) Get() []byte {
    return bp.pool.Get().([]byte)
}

func (bp *BufferPool) Put(buf []byte) {
    if cap(buf) > 64*1024 { // Don't keep very large buffers
        return
    }
    bp.pool.Put(buf[:0])
}
```

### Worker Pool Pattern

**Pattern:** Control concurrency with a fixed number of workers to prevent resource exhaustion.

```go
// ✅ Pattern: Worker pool
type WorkerPool struct {
    workers    int
    taskQueue  chan Task
    resultChan chan Result
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

type Task func() Result
type Result struct {
    Value interface{}
    Error error
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        workers:    workers,
        taskQueue:  make(chan Task, queueSize),
        resultChan: make(chan Result, queueSize),
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case task := <-wp.taskQueue:
            result := task()
            
            select {
            case wp.resultChan <- result:
            case <-wp.ctx.Done():
                return
            }
            
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(task Task) error {
    select {
    case wp.taskQueue <- task:
        return nil
    case <-wp.ctx.Done():
        return wp.ctx.Err()
    default:
        return errors.New("task queue full")
    }
}

func (wp *WorkerPool) Results() <-chan Result {
    return wp.resultChan
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    close(wp.taskQueue)
    wp.wg.Wait()
    close(wp.resultChan)
}

// Usage example
func processWithWorkerPool(tasks []Task) []Result {
    pool := NewWorkerPool(runtime.NumCPU(), len(tasks))
    pool.Start()
    defer pool.Stop()
    
    // Submit all tasks
    for _, task := range tasks {
        pool.Submit(task)
    }
    
    // Collect results
    results := make([]Result, 0, len(tasks))
    for i := 0; i < len(tasks); i++ {
        result := <-pool.Results()
        results = append(results, result)
    }
    
    return results
}
```

## Memory Optimization Patterns

### Zero Allocation JSON Parsing

**Pattern:** Parse JSON without allocations using streaming or pre-allocated structures.

```go
// ✅ Pattern: Pre-allocated struct with json.Decoder
type Event struct {
    ID        int64     `json:"id"`
    Type      string    `json:"type"`
    Timestamp time.Time `json:"timestamp"`
    Data      []byte    `json:"data"`
}

func parseEventsEfficient(r io.Reader) ([]Event, error) {
    decoder := json.NewDecoder(r)
    
    var events []Event
    
    // Expect array start
    token, err := decoder.Token()
    if err != nil {
        return nil, err
    }
    if delim, ok := token.(json.Delim); !ok || delim != '[' {
        return nil, errors.New("expected array")
    }
    
    // Parse each object
    for decoder.More() {
        var event Event
        if err := decoder.Decode(&event); err != nil {
            return nil, err
        }
        events = append(events, event)
    }
    
    // Expect array end
    if _, err := decoder.Token(); err != nil {
        return nil, err
    }
    
    return events, nil
}

// ✅ Pattern: Custom JSON parser for hot paths
func parseIDFromJSON(data []byte) (int64, error) {
    // Simple parser for {"id": 12345, ...}
    start := bytes.Index(data, []byte(`"id":`))
    if start == -1 {
        return 0, errors.New("id not found")
    }
    
    start += 5 // len(`"id":`)
    
    // Skip whitespace
    for start < len(data) && (data[start] == ' ' || data[start] == '\t') {
        start++
    }
    
    end := start
    for end < len(data) && data[end] >= '0' && data[end] <= '9' {
        end++
    }
    
    if end == start {
        return 0, errors.New("invalid id format")
    }
    
    return strconv.ParseInt(string(data[start:end]), 10, 64)
}
```

### Memory-Mapped File Pattern

**Pattern:** Use memory mapping for large file processing.

```go
// ✅ Pattern: Memory-mapped file processing
func processLargeFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    stat, err := file.Stat()
    if err != nil {
        return err
    }
    
    size := stat.Size()
    
    // Memory map the file (platform-specific implementation)
    data, err := syscall.Mmap(int(file.Fd()), 0, int(size), 
        syscall.PROT_READ, syscall.MAP_PRIVATE)
    if err != nil {
        return err
    }
    defer syscall.Munmap(data)
    
    // Process data without loading entire file into memory
    return processDataInChunks(data)
}

func processDataInChunks(data []byte) error {
    chunkSize := 64 * 1024 // 64KB chunks
    
    for i := 0; i < len(data); i += chunkSize {
        end := i + chunkSize
        if end > len(data) {
            end = len(data)
        }
        
        chunk := data[i:end]
        if err := processChunk(chunk); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Concurrency Patterns

### Pipeline Pattern

**Pattern:** Process data through stages with goroutines and channels.

```go
// ✅ Pattern: Pipeline processing
func processPipeline(input <-chan RawData) <-chan ProcessedData {
    // Stage 1: Validate
    validated := make(chan ValidatedData, 100)
    go func() {
        defer close(validated)
        for data := range input {
            if valid := validate(data); valid != nil {
                validated <- *valid
            }
        }
    }()
    
    // Stage 2: Transform
    transformed := make(chan TransformedData, 100)
    go func() {
        defer close(transformed)
        for data := range validated {
            transformed <- transform(data)
        }
    }()
    
    // Stage 3: Enrich
    processed := make(chan ProcessedData, 100)
    go func() {
        defer close(processed)
        for data := range transformed {
            processed <- enrich(data)
        }
    }()
    
    return processed
}

// ✅ Pattern: Fan-out/Fan-in
func fanOutFanIn(input <-chan Task, numWorkers int) <-chan Result {
    // Fan-out: distribute work to multiple workers
    workers := make([]<-chan Result, numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        worker := make(chan Result)
        workers[i] = worker
        
        go func(output chan<- Result) {
            defer close(output)
            for task := range input {
                output <- processTask(task)
            }
        }(worker)
    }
    
    // Fan-in: merge results from all workers
    return mergeChannels(workers...)
}

func mergeChannels(channels ...<-chan Result) <-chan Result {
    merged := make(chan Result)
    var wg sync.WaitGroup
    
    wg.Add(len(channels))
    for _, ch := range channels {
        go func(c <-chan Result) {
            defer wg.Done()
            for result := range c {
                merged <- result
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(merged)
    }()
    
    return merged
}
```

### Rate Limiting Pattern

**Pattern:** Control resource access rate to prevent system overload.

```go
// ✅ Pattern: Token bucket rate limiter
type RateLimiter struct {
    tokens    chan struct{}
    ticker    *time.Ticker
    maxTokens int
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
    rl := &RateLimiter{
        tokens:    make(chan struct{}, burst),
        ticker:    time.NewTicker(time.Second / time.Duration(rate)),
        maxTokens: burst,
    }
    
    // Fill initial tokens
    for i := 0; i < burst; i++ {
        rl.tokens <- struct{}{}
    }
    
    // Refill tokens
    go func() {
        defer rl.ticker.Stop()
        for range rl.ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default: // Bucket full
            }
        }
    }()
    
    return rl
}

func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Usage in HTTP handler
func rateLimitedHandler(rl *RateLimiter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !rl.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        // Process request
        handleRequest(w, r)
    }
}
```

## I/O Optimization Patterns

### Buffered Writer Pattern

**Pattern:** Batch writes to reduce system call overhead.

```go
// ✅ Pattern: Buffered batch writing
type BatchWriter struct {
    writer io.Writer
    buffer []byte
    size   int
    mu     sync.Mutex
}

func NewBatchWriter(w io.Writer, batchSize int) *BatchWriter {
    bw := &BatchWriter{
        writer: w,
        buffer: make([]byte, 0, batchSize),
        size:   batchSize,
    }
    
    // Periodic flush
    go func() {
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()
        
        for range ticker.C {
            bw.Flush()
        }
    }()
    
    return bw
}

func (bw *BatchWriter) Write(data []byte) error {
    bw.mu.Lock()
    defer bw.mu.Unlock()
    
    // If adding this data would exceed buffer size, flush first
    if len(bw.buffer)+len(data) > bw.size {
        if err := bw.flushLocked(); err != nil {
            return err
        }
    }
    
    bw.buffer = append(bw.buffer, data...)
    
    return nil
}

func (bw *BatchWriter) Flush() error {
    bw.mu.Lock()
    defer bw.mu.Unlock()
    return bw.flushLocked()
}

func (bw *BatchWriter) flushLocked() error {
    if len(bw.buffer) == 0 {
        return nil
    }
    
    _, err := bw.writer.Write(bw.buffer)
    bw.buffer = bw.buffer[:0] // Reset length, keep capacity
    return err
}
```

### Connection Pool Pattern

**Pattern:** Reuse network connections to reduce establishment overhead.

```go
// ✅ Pattern: Connection pool
type ConnectionPool struct {
    factory    func() (net.Conn, error)
    pool       chan net.Conn
    maxSize    int
    activeConn int32
    mu         sync.RWMutex
}

func NewConnectionPool(factory func() (net.Conn, error), maxSize int) *ConnectionPool {
    return &ConnectionPool{
        factory: factory,
        pool:    make(chan net.Conn, maxSize),
        maxSize: maxSize,
    }
}

func (cp *ConnectionPool) Get() (net.Conn, error) {
    select {
    case conn := <-cp.pool:
        // Test connection before returning
        if err := testConnection(conn); err != nil {
            conn.Close()
            return cp.createConnection()
        }
        return conn, nil
    default:
        return cp.createConnection()
    }
}

func (cp *ConnectionPool) Put(conn net.Conn) {
    if conn == nil {
        return
    }
    
    select {
    case cp.pool <- conn:
        // Successfully returned to pool
    default:
        // Pool full, close connection
        conn.Close()
        atomic.AddInt32(&cp.activeConn, -1)
    }
}

func (cp *ConnectionPool) createConnection() (net.Conn, error) {
    active := atomic.AddInt32(&cp.activeConn, 1)
    if int(active) > cp.maxSize {
        atomic.AddInt32(&cp.activeConn, -1)
        return nil, errors.New("connection pool exhausted")
    }
    
    return cp.factory()
}

func testConnection(conn net.Conn) error {
    // Set a short deadline for the test
    conn.SetDeadline(time.Now().Add(time.Second))
    defer conn.SetDeadline(time.Time{})
    
    // Try to write and read a byte
    if _, err := conn.Write([]byte{0}); err != nil {
        return err
    }
    
    buffer := make([]byte, 1)
    _, err := conn.Read(buffer)
    return err
}
```

## Anti-patterns to Avoid

### The "God Goroutine" Anti-pattern

```go
// ❌ Anti-pattern: Single goroutine doing everything
func godGoroutine(input <-chan Data) {
    for data := range input {
        // Validation
        if !isValid(data) {
            continue
        }
        
        // Transformation
        transformed := transform(data)
        
        // Database save
        saveToDatabase(transformed)
        
        // Send notification
        sendNotification(transformed)
        
        // Update cache
        updateCache(transformed)
        
        // Log metrics
        logMetrics(transformed)
    }
}

// ✅ Better: Separate concerns with pipelines
func processDataPipeline(input <-chan Data) {
    validated := validateStage(input)
    transformed := transformStage(validated)
    
    // Fan out to multiple processing stages
    go saveStage(transformed)
    go notificationStage(transformed)
    go cacheStage(transformed)
    go metricsStage(transformed)
}
```

### The "Premature Channels" Anti-pattern

```go
// ❌ Anti-pattern: Unnecessary channel complexity
func unnecessaryChannels() {
    ch1 := make(chan int)
    ch2 := make(chan int)
    ch3 := make(chan int)
    
    go func() {
        for i := 0; i < 10; i++ {
            ch1 <- i
        }
        close(ch1)
    }()
    
    go func() {
        for val := range ch1 {
            ch2 <- val * 2
        }
        close(ch2)
    }()
    
    go func() {
        for val := range ch2 {
            ch3 <- val + 1
        }
        close(ch3)
    }()
    
    for result := range ch3 {
        fmt.Println(result)
    }
}

// ✅ Better: Simple sequential processing when no concurrency needed
func simpleProcessing() {
    for i := 0; i < 10; i++ {
        result := (i * 2) + 1
        fmt.Println(result)
    }
}
```

These performance patterns provide proven approaches to common optimization challenges in Go, helping you write efficient, scalable, and maintainable code while avoiding common pitfalls.
