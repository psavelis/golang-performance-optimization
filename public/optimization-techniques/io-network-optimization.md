# I/O and Network Optimization

Input/Output and network operations are often the primary performance bottlenecks in modern applications. This chapter explores advanced techniques for optimizing I/O operations, network communication, and data serialization to achieve maximum throughput and minimal latency while maintaining system reliability and scalability.

## File System Optimization

### Advanced File I/O Patterns
Implement high-performance file operations with sophisticated caching and batching:

```go
package io_optimization

import (
    "bufio"
    "context"
    "io"
    "os"
    "sync"
    "syscall"
    "time"
    "unsafe"
)

// High-performance file manager with advanced I/O optimization
type OptimizedFileManager struct {
    readCache    *ReadCache
    writeBuffer  *WriteBuffer
    asyncWriter  *AsyncWriter
    metrics      IOMetrics
    config       IOConfig
}

type IOConfig struct {
    ReadBufferSize    int           `json:"read_buffer_size"`
    WriteBufferSize   int           `json:"write_buffer_size"`
    MaxConcurrentOps  int           `json:"max_concurrent_ops"`
    FlushInterval     time.Duration `json:"flush_interval"`
    UseDirectIO       bool          `json:"use_direct_io"`
    UseMmap           bool          `json:"use_mmap"`
    AsyncWrites       bool          `json:"async_writes"`
    Compression       string        `json:"compression"`
}

type IOMetrics struct {
    BytesRead       int64         `json:"bytes_read"`
    BytesWritten    int64         `json:"bytes_written"`
    ReadOperations  int64         `json:"read_operations"`
    WriteOperations int64         `json:"write_operations"`
    CacheHits       int64         `json:"cache_hits"`
    CacheMisses     int64         `json:"cache_misses"`
    AvgReadLatency  time.Duration `json:"avg_read_latency"`
    AvgWriteLatency time.Duration `json:"avg_write_latency"`
    Throughput      float64       `json:"throughput_mbps"`
}

type ReadCache struct {
    cache     map[string]*CacheEntry
    lru       *LRUList
    maxSize   int64
    currentSize int64
    mu        sync.RWMutex
}

type CacheEntry struct {
    key       string
    data      []byte
    timestamp time.Time
    accessCount int64
    size      int64
    prev      *CacheEntry
    next      *CacheEntry
}

type LRUList struct {
    head *CacheEntry
    tail *CacheEntry
}

type WriteBuffer struct {
    buffer      []byte
    offset      int
    maxSize     int
    flushChan   chan struct{}
    mu          sync.Mutex
    pendingOps  map[string]*WriteOperation
}

type WriteOperation struct {
    filename string
    data     []byte
    offset   int64
    done     chan error
}

type AsyncWriter struct {
    queue       chan *WriteOperation
    workers     int
    workerPool  chan chan *WriteOperation
    quit        chan struct{}
    wg          sync.WaitGroup
}

func NewOptimizedFileManager(config IOConfig) *OptimizedFileManager {
    fm := &OptimizedFileManager{
        config: config,
        readCache: &ReadCache{
            cache:   make(map[string]*CacheEntry),
            lru:     &LRUList{},
            maxSize: int64(config.ReadBufferSize * 10), // 10x buffer size for cache
        },
        writeBuffer: &WriteBuffer{
            buffer:     make([]byte, config.WriteBufferSize),
            maxSize:    config.WriteBufferSize,
            flushChan:  make(chan struct{}, 1),
            pendingOps: make(map[string]*WriteOperation),
        },
    }
    
    if config.AsyncWrites {
        fm.asyncWriter = NewAsyncWriter(config.MaxConcurrentOps)
    }
    
    // Start background flusher
    go fm.backgroundFlusher()
    
    return fm
}

func NewAsyncWriter(workers int) *AsyncWriter {
    aw := &AsyncWriter{
        queue:      make(chan *WriteOperation, workers*2),
        workers:    workers,
        workerPool: make(chan chan *WriteOperation, workers),
        quit:       make(chan struct{}),
    }
    
    // Start workers
    for i := 0; i < workers; i++ {
        worker := make(chan *WriteOperation)
        aw.workerPool <- worker
        aw.wg.Add(1)
        
        go aw.worker(worker)
    }
    
    go aw.dispatcher()
    
    return aw
}

func (aw *AsyncWriter) worker(workerChan chan *WriteOperation) {
    defer aw.wg.Done()
    
    for {
        select {
        case op := <-workerChan:
            aw.executeWrite(op)
            // Return worker to pool
            aw.workerPool <- workerChan
            
        case <-aw.quit:
            return
        }
    }
}

func (aw *AsyncWriter) dispatcher() {
    for {
        select {
        case op := <-aw.queue:
            // Get available worker
            worker := <-aw.workerPool
            worker <- op
            
        case <-aw.quit:
            return
        }
    }
}

func (aw *AsyncWriter) executeWrite(op *WriteOperation) {
    var err error
    
    if op.offset >= 0 {
        // Positioned write
        err = aw.writeAtOffset(op.filename, op.data, op.offset)
    } else {
        // Append write
        err = aw.appendToFile(op.filename, op.data)
    }
    
    op.done <- err
}

func (aw *AsyncWriter) writeAtOffset(filename string, data []byte, offset int64) error {
    file, err := os.OpenFile(filename, os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = file.WriteAt(data, offset)
    return err
}

func (aw *AsyncWriter) appendToFile(filename string, data []byte) error {
    file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = file.Write(data)
    return err
}

// Optimized sequential read with prefetching
func (fm *OptimizedFileManager) SequentialRead(filename string, chunkSize int) <-chan ReadResult {
    resultChan := make(chan ReadResult, 10) // Buffer for prefetching
    
    go func() {
        defer close(resultChan)
        
        file, err := os.Open(filename)
        if err != nil {
            resultChan <- ReadResult{Error: err}
            return
        }
        defer file.Close()
        
        // Use large buffer for sequential reads
        reader := bufio.NewReaderSize(file, fm.config.ReadBufferSize)
        buffer := make([]byte, chunkSize)
        
        for {
            start := time.Now()
            n, err := reader.Read(buffer)
            latency := time.Since(start)
            
            if n > 0 {
                // Copy data to avoid sharing buffer
                data := make([]byte, n)
                copy(data, buffer[:n])
                
                resultChan <- ReadResult{
                    Data:    data,
                    Latency: latency,
                }
                
                // Update metrics
                atomic.AddInt64(&fm.metrics.BytesRead, int64(n))
                atomic.AddInt64(&fm.metrics.ReadOperations, 1)
                fm.updateReadLatency(latency)
            }
            
            if err != nil {
                if err != io.EOF {
                    resultChan <- ReadResult{Error: err}
                }
                break
            }
        }
    }()
    
    return resultChan
}

type ReadResult struct {
    Data    []byte
    Offset  int64
    Latency time.Duration
    Error   error
}

// Cached random access reads
func (fm *OptimizedFileManager) CachedRead(filename string, offset int64, size int) ([]byte, error) {
    cacheKey := fmt.Sprintf("%s:%d:%d", filename, offset, size)
    
    // Check cache first
    if data := fm.readCache.Get(cacheKey); data != nil {
        atomic.AddInt64(&fm.metrics.CacheHits, 1)
        return data, nil
    }
    
    atomic.AddInt64(&fm.metrics.CacheMisses, 1)
    
    // Read from file
    start := time.Now()
    data, err := fm.readAtOffset(filename, offset, size)
    latency := time.Since(start)
    
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    fm.readCache.Put(cacheKey, data)
    
    // Update metrics
    atomic.AddInt64(&fm.metrics.BytesRead, int64(len(data)))
    atomic.AddInt64(&fm.metrics.ReadOperations, 1)
    fm.updateReadLatency(latency)
    
    return data, nil
}

func (fm *OptimizedFileManager) readAtOffset(filename string, offset int64, size int) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    data := make([]byte, size)
    n, err := file.ReadAt(data, offset)
    
    if err != nil && err != io.EOF {
        return nil, err
    }
    
    return data[:n], nil
}

// Optimized buffered writes
func (fm *OptimizedFileManager) BufferedWrite(filename string, data []byte) error {
    if fm.config.AsyncWrites {
        return fm.asyncWrite(filename, data, -1) // -1 for append
    }
    
    return fm.synchronousBufferedWrite(filename, data)
}

func (fm *OptimizedFileManager) asyncWrite(filename string, data []byte, offset int64) error {
    if fm.asyncWriter == nil {
        return fmt.Errorf("async writer not initialized")
    }
    
    op := &WriteOperation{
        filename: filename,
        data:     make([]byte, len(data)),
        offset:   offset,
        done:     make(chan error, 1),
    }
    
    copy(op.data, data)
    
    select {
    case fm.asyncWriter.queue <- op:
        return <-op.done
    case <-time.After(5 * time.Second):
        return fmt.Errorf("async write timeout")
    }
}

func (fm *OptimizedFileManager) synchronousBufferedWrite(filename string, data []byte) error {
    fm.writeBuffer.mu.Lock()
    defer fm.writeBuffer.mu.Unlock()
    
    // Check if buffer has space
    if fm.writeBuffer.offset+len(data) > fm.writeBuffer.maxSize {
        // Flush buffer first
        if err := fm.flushBuffer(); err != nil {
            return err
        }
    }
    
    // Add to buffer
    copy(fm.writeBuffer.buffer[fm.writeBuffer.offset:], data)
    fm.writeBuffer.offset += len(data)
    
    // Create pending operation
    op := &WriteOperation{
        filename: filename,
        data:     make([]byte, len(data)),
        offset:   -1,
        done:     make(chan error, 1),
    }
    copy(op.data, data)
    
    fm.writeBuffer.pendingOps[filename] = op
    
    // Trigger flush if buffer is getting full
    if fm.writeBuffer.offset > fm.writeBuffer.maxSize*8/10 {
        select {
        case fm.writeBuffer.flushChan <- struct{}{}:
        default:
            // Flush already pending
        }
    }
    
    return nil
}

func (fm *OptimizedFileManager) flushBuffer() error {
    if fm.writeBuffer.offset == 0 {
        return nil
    }
    
    start := time.Now()
    
    // Execute all pending operations
    for filename, op := range fm.writeBuffer.pendingOps {
        if err := fm.appendToFile(filename, op.data); err != nil {
            op.done <- err
            continue
        }
        op.done <- nil
        
        // Update metrics
        atomic.AddInt64(&fm.metrics.BytesWritten, int64(len(op.data)))
        atomic.AddInt64(&fm.metrics.WriteOperations, 1)
    }
    
    // Clear buffer and pending operations
    fm.writeBuffer.offset = 0
    fm.writeBuffer.pendingOps = make(map[string]*WriteOperation)
    
    latency := time.Since(start)
    fm.updateWriteLatency(latency)
    
    return nil
}

func (fm *OptimizedFileManager) appendToFile(filename string, data []byte) error {
    file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = file.Write(data)
    return err
}

func (fm *OptimizedFileManager) backgroundFlusher() {
    ticker := time.NewTicker(fm.config.FlushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            fm.writeBuffer.mu.Lock()
            fm.flushBuffer()
            fm.writeBuffer.mu.Unlock()
            
        case <-fm.writeBuffer.flushChan:
            fm.writeBuffer.mu.Lock()
            fm.flushBuffer()
            fm.writeBuffer.mu.Unlock()
        }
    }
}

// Memory-mapped file operations
func (fm *OptimizedFileManager) MemoryMapFile(filename string) (*MappedFile, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    
    stat, err := file.Stat()
    if err != nil {
        file.Close()
        return nil, err
    }
    
    size := stat.Size()
    
    // Memory map the file
    data, err := syscall.Mmap(int(file.Fd()), 0, int(size), 
                             syscall.PROT_READ, syscall.MAP_SHARED)
    if err != nil {
        file.Close()
        return nil, err
    }
    
    return &MappedFile{
        file: file,
        data: data,
        size: size,
    }, nil
}

type MappedFile struct {
    file *os.File
    data []byte
    size int64
}

func (mf *MappedFile) Read(offset int64, length int) ([]byte, error) {
    if offset < 0 || offset >= mf.size {
        return nil, fmt.Errorf("offset out of range")
    }
    
    end := offset + int64(length)
    if end > mf.size {
        end = mf.size
    }
    
    // Direct memory access - no copy needed
    return mf.data[offset:end], nil
}

func (mf *MappedFile) Close() error {
    if err := syscall.Munmap(mf.data); err != nil {
        return err
    }
    return mf.file.Close()
}

// Cache implementation
func (rc *ReadCache) Get(key string) []byte {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    
    entry, exists := rc.cache[key]
    if !exists {
        return nil
    }
    
    // Update access count and move to front
    atomic.AddInt64(&entry.accessCount, 1)
    entry.timestamp = time.Now()
    rc.moveToFront(entry)
    
    // Return copy to prevent external modifications
    result := make([]byte, len(entry.data))
    copy(result, entry.data)
    return result
}

func (rc *ReadCache) Put(key string, data []byte) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    
    // Check if key already exists
    if existing, exists := rc.cache[key]; exists {
        // Update existing entry
        existing.data = make([]byte, len(data))
        copy(existing.data, data)
        existing.timestamp = time.Now()
        existing.size = int64(len(data))
        rc.moveToFront(existing)
        return
    }
    
    // Create new entry
    entry := &CacheEntry{
        key:         key,
        data:        make([]byte, len(data)),
        timestamp:   time.Now(),
        accessCount: 1,
        size:        int64(len(data)),
    }
    copy(entry.data, data)
    
    // Add to cache
    rc.cache[key] = entry
    rc.addToFront(entry)
    rc.currentSize += entry.size
    
    // Evict if necessary
    rc.evictIfNecessary()
}

func (rc *ReadCache) evictIfNecessary() {
    for rc.currentSize > rc.maxSize && rc.lru.tail != nil {
        victim := rc.lru.tail
        rc.removeEntry(victim)
        delete(rc.cache, victim.key)
        rc.currentSize -= victim.size
    }
}

func (rc *ReadCache) moveToFront(entry *CacheEntry) {
    rc.removeFromList(entry)
    rc.addToFront(entry)
}

func (rc *ReadCache) addToFront(entry *CacheEntry) {
    entry.next = rc.lru.head
    entry.prev = nil
    
    if rc.lru.head != nil {
        rc.lru.head.prev = entry
    }
    
    rc.lru.head = entry
    
    if rc.lru.tail == nil {
        rc.lru.tail = entry
    }
}

func (rc *ReadCache) removeFromList(entry *CacheEntry) {
    if entry.prev != nil {
        entry.prev.next = entry.next
    } else {
        rc.lru.head = entry.next
    }
    
    if entry.next != nil {
        entry.next.prev = entry.prev
    } else {
        rc.lru.tail = entry.prev
    }
}

func (rc *ReadCache) removeEntry(entry *CacheEntry) {
    rc.removeFromList(entry)
}

// Metrics updates
func (fm *OptimizedFileManager) updateReadLatency(latency time.Duration) {
    // Simple moving average
    currentAvg := atomic.LoadInt64((*int64)(&fm.metrics.AvgReadLatency))
    newAvg := (time.Duration(currentAvg) + latency) / 2
    atomic.StoreInt64((*int64)(&fm.metrics.AvgReadLatency), int64(newAvg))
}

func (fm *OptimizedFileManager) updateWriteLatency(latency time.Duration) {
    currentAvg := atomic.LoadInt64((*int64)(&fm.metrics.AvgWriteLatency))
    newAvg := (time.Duration(currentAvg) + latency) / 2
    atomic.StoreInt64((*int64)(&fm.metrics.AvgWriteLatency), int64(newAvg))
}

func (fm *OptimizedFileManager) GetMetrics() IOMetrics {
    return IOMetrics{
        BytesRead:       atomic.LoadInt64(&fm.metrics.BytesRead),
        BytesWritten:    atomic.LoadInt64(&fm.metrics.BytesWritten),
        ReadOperations:  atomic.LoadInt64(&fm.metrics.ReadOperations),
        WriteOperations: atomic.LoadInt64(&fm.metrics.WriteOperations),
        CacheHits:       atomic.LoadInt64(&fm.metrics.CacheHits),
        CacheMisses:     atomic.LoadInt64(&fm.metrics.CacheMisses),
        AvgReadLatency:  time.Duration(atomic.LoadInt64((*int64)(&fm.metrics.AvgReadLatency))),
        AvgWriteLatency: time.Duration(atomic.LoadInt64((*int64)(&fm.metrics.AvgWriteLatency))),
    }
}

// Shutdown cleanup
func (fm *OptimizedFileManager) Shutdown() error {
    // Flush any pending writes
    fm.writeBuffer.mu.Lock()
    err := fm.flushBuffer()
    fm.writeBuffer.mu.Unlock()
    
    // Shutdown async writer
    if fm.asyncWriter != nil {
        close(fm.asyncWriter.quit)
        fm.asyncWriter.wg.Wait()
    }
    
    return err
}
```

## Network Optimization

### High-Performance Network Programming
Implement advanced network optimization techniques:

```go
// High-performance network server with optimizations
type OptimizedNetworkServer struct {
    listener     net.Listener
    connPool     *ConnectionPool
    readBuffer   *NetworkBufferPool
    writeBuffer  *NetworkBufferPool
    metrics      NetworkMetrics
    config       NetworkConfig
    epoll        *EpollManager
    workerPool   *NetworkWorkerPool
}

type NetworkConfig struct {
    MaxConnections    int           `json:"max_connections"`
    ReadBufferSize    int           `json:"read_buffer_size"`
    WriteBufferSize   int           `json:"write_buffer_size"`
    ReadTimeout       time.Duration `json:"read_timeout"`
    WriteTimeout      time.Duration `json:"write_timeout"`
    KeepAlive         bool          `json:"keep_alive"`
    KeepAlivePeriod   time.Duration `json:"keep_alive_period"`
    TCPNoDelay        bool          `json:"tcp_no_delay"`
    ReusePort         bool          `json:"reuse_port"`
    UseEpoll          bool          `json:"use_epoll"`
    WorkerCount       int           `json:"worker_count"`
}

type NetworkMetrics struct {
    ActiveConnections   int64         `json:"active_connections"`
    TotalConnections    int64         `json:"total_connections"`
    BytesReceived       int64         `json:"bytes_received"`
    BytesSent           int64         `json:"bytes_sent"`
    RequestsProcessed   int64         `json:"requests_processed"`
    AvgRequestLatency   time.Duration `json:"avg_request_latency"`
    ConnectionsPerSec   float64       `json:"connections_per_sec"`
    Throughput          float64       `json:"throughput_mbps"`
    ErrorRate           float64       `json:"error_rate"`
}

type ConnectionPool struct {
    connections map[int]*PooledConnection
    available   chan *PooledConnection
    maxSize     int
    currentSize int32
    mu          sync.RWMutex
}

type PooledConnection struct {
    conn        net.Conn
    id          int
    lastUsed    time.Time
    inUse       bool
    readBuffer  []byte
    writeBuffer []byte
}

type NetworkBufferPool struct {
    pool      sync.Pool
    bufferSize int
    allocCount int64
    reuseCount int64
}

func NewNetworkBufferPool(bufferSize int) *NetworkBufferPool {
    return &NetworkBufferPool{
        bufferSize: bufferSize,
        pool: sync.Pool{
            New: func() interface{} {
                atomic.AddInt64(&allocCount, 1)
                return make([]byte, bufferSize)
            },
        },
    }
}

func (nbp *NetworkBufferPool) Get() []byte {
    atomic.AddInt64(&nbp.reuseCount, 1)
    return nbp.pool.Get().([]byte)
}

func (nbp *NetworkBufferPool) Put(buffer []byte) {
    if len(buffer) == nbp.bufferSize {
        // Clear buffer
        for i := range buffer {
            buffer[i] = 0
        }
        nbp.pool.Put(buffer)
    }
}

// Epoll-based event loop for Linux systems
type EpollManager struct {
    epollFd   int
    events    []syscall.EpollEvent
    eventChan chan EpollEvent
    running   bool
    mu        sync.Mutex
}

type EpollEvent struct {
    Fd     int32
    Events uint32
    Conn   net.Conn
}

func NewEpollManager() (*EpollManager, error) {
    epollFd, err := syscall.EpollCreate1(0)
    if err != nil {
        return nil, err
    }
    
    return &EpollManager{
        epollFd:   epollFd,
        events:    make([]syscall.EpollEvent, 1024),
        eventChan: make(chan EpollEvent, 1024),
        running:   true,
    }, nil
}

func (em *EpollManager) AddConnection(conn net.Conn) error {
    fd := getFd(conn)
    
    event := syscall.EpollEvent{
        Events: syscall.EPOLLIN | syscall.EPOLLOUT | syscall.EPOLLET, // Edge-triggered
        Fd:     int32(fd),
    }
    
    return syscall.EpollCtl(em.epollFd, syscall.EPOLL_CTL_ADD, fd, &event)
}

func (em *EpollManager) RemoveConnection(conn net.Conn) error {
    fd := getFd(conn)
    return syscall.EpollCtl(em.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (em *EpollManager) EventLoop() {
    for em.running {
        n, err := syscall.EpollWait(em.epollFd, em.events, 100) // 100ms timeout
        if err != nil {
            if err == syscall.EINTR {
                continue
            }
            fmt.Printf("EpollWait error: %v\n", err)
            continue
        }
        
        for i := 0; i < n; i++ {
            event := em.events[i]
            
            select {
            case em.eventChan <- EpollEvent{
                Fd:     event.Fd,
                Events: event.Events,
            }:
            default:
                // Event channel full, drop event
            }
        }
    }
}

func getFd(conn net.Conn) int {
    // Extract file descriptor from connection
    // This is implementation-specific and may vary
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        file, _ := tcpConn.File()
        return int(file.Fd())
    }
    return -1
}

// Network worker pool for handling connections
type NetworkWorkerPool struct {
    workers    []*NetworkWorker
    workQueue  chan *WorkItem
    workerPool chan chan *WorkItem
    quit       chan struct{}
    wg         sync.WaitGroup
}

type NetworkWorker struct {
    id         int
    workerPool chan chan *WorkItem
    workChan   chan *WorkItem
    quit       chan struct{}
    processor  RequestProcessor
}

type WorkItem struct {
    conn     net.Conn
    request  []byte
    response chan []byte
    start    time.Time
}

type RequestProcessor interface {
    ProcessRequest(request []byte) []byte
}

func NewNetworkWorkerPool(workerCount int, processor RequestProcessor) *NetworkWorkerPool {
    pool := &NetworkWorkerPool{
        workers:    make([]*NetworkWorker, workerCount),
        workQueue:  make(chan *WorkItem, workerCount*2),
        workerPool: make(chan chan *WorkItem, workerCount),
        quit:       make(chan struct{}),
    }
    
    // Create workers
    for i := 0; i < workerCount; i++ {
        worker := &NetworkWorker{
            id:         i,
            workerPool: pool.workerPool,
            workChan:   make(chan *WorkItem),
            quit:       make(chan struct{}),
            processor:  processor,
        }
        
        pool.workers[i] = worker
        pool.wg.Add(1)
        go worker.start(&pool.wg)
    }
    
    // Start dispatcher
    go pool.dispatch()
    
    return pool
}

func (nwp *NetworkWorkerPool) Submit(item *WorkItem) {
    select {
    case nwp.workQueue <- item:
    case <-time.After(1 * time.Second):
        // Handle timeout
        close(item.response)
    }
}

func (nwp *NetworkWorkerPool) dispatch() {
    for {
        select {
        case work := <-nwp.workQueue:
            // Get available worker
            workerChan := <-nwp.workerPool
            workerChan <- work
            
        case <-nwp.quit:
            return
        }
    }
}

func (nw *NetworkWorker) start(wg *sync.WaitGroup) {
    defer wg.Done()
    
    for {
        // Register worker
        nw.workerPool <- nw.workChan
        
        select {
        case work := <-nw.workChan:
            nw.processWork(work)
            
        case <-nw.quit:
            return
        }
    }
}

func (nw *NetworkWorker) processWork(work *WorkItem) {
    response := nw.processor.ProcessRequest(work.request)
    
    select {
    case work.response <- response:
    case <-time.After(1 * time.Second):
        // Response channel timeout
    }
    
    close(work.response)
}

// Optimized server implementation
func NewOptimizedNetworkServer(config NetworkConfig) (*OptimizedNetworkServer, error) {
    server := &OptimizedNetworkServer{
        config:      config,
        readBuffer:  NewNetworkBufferPool(config.ReadBufferSize),
        writeBuffer: NewNetworkBufferPool(config.WriteBufferSize),
        connPool: &ConnectionPool{
            connections: make(map[int]*PooledConnection),
            available:   make(chan *PooledConnection, config.MaxConnections),
            maxSize:     config.MaxConnections,
        },
    }
    
    if config.UseEpoll {
        epoll, err := NewEpollManager()
        if err != nil {
            return nil, fmt.Errorf("failed to create epoll manager: %v", err)
        }
        server.epoll = epoll
        go server.epoll.EventLoop()
    }
    
    // Create worker pool
    processor := &EchoProcessor{} // Example processor
    server.workerPool = NewNetworkWorkerPool(config.WorkerCount, processor)
    
    return server, nil
}

func (ons *OptimizedNetworkServer) Listen(address string) error {
    var err error
    
    if ons.config.ReusePort {
        ons.listener, err = reuseport.Listen("tcp", address)
    } else {
        ons.listener, err = net.Listen("tcp", address)
    }
    
    if err != nil {
        return err
    }
    
    // Configure TCP listener options
    if tcpListener, ok := ons.listener.(*net.TCPListener); ok {
        err = ons.configureTCPListener(tcpListener)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (ons *OptimizedNetworkServer) configureTCPListener(listener *net.TCPListener) error {
    // Set socket options for performance
    file, err := listener.File()
    if err != nil {
        return err
    }
    defer file.Close()
    
    fd := int(file.Fd())
    
    // Enable TCP_NODELAY
    if ons.config.TCPNoDelay {
        err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
        if err != nil {
            return err
        }
    }
    
    // Set receive buffer size
    err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, ons.config.ReadBufferSize)
    if err != nil {
        return err
    }
    
    // Set send buffer size
    err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, ons.config.WriteBufferSize)
    if err != nil {
        return err
    }
    
    return nil
}

func (ons *OptimizedNetworkServer) Accept() {
    for {
        conn, err := ons.listener.Accept()
        if err != nil {
            fmt.Printf("Accept error: %v\n", err)
            continue
        }
        
        atomic.AddInt64(&ons.metrics.TotalConnections, 1)
        atomic.AddInt64(&ons.metrics.ActiveConnections, 1)
        
        // Configure connection
        ons.configureConnection(conn)
        
        if ons.config.UseEpoll {
            // Add to epoll
            ons.epoll.AddConnection(conn)
            go ons.handleEpollConnection(conn)
        } else {
            // Handle with goroutine per connection
            go ons.handleConnection(conn)
        }
    }
}

func (ons *OptimizedNetworkServer) configureConnection(conn net.Conn) {
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        // Configure TCP options
        if ons.config.TCPNoDelay {
            tcpConn.SetNoDelay(true)
        }
        
        if ons.config.KeepAlive {
            tcpConn.SetKeepAlive(true)
            tcpConn.SetKeepAlivePeriod(ons.config.KeepAlivePeriod)
        }
        
        // Set timeouts
        tcpConn.SetReadDeadline(time.Now().Add(ons.config.ReadTimeout))
        tcpConn.SetWriteDeadline(time.Now().Add(ons.config.WriteTimeout))
    }
}

func (ons *OptimizedNetworkServer) handleConnection(conn net.Conn) {
    defer func() {
        conn.Close()
        atomic.AddInt64(&ons.metrics.ActiveConnections, -1)
    }()
    
    buffer := ons.readBuffer.Get()
    defer ons.readBuffer.Put(buffer)
    
    for {
        // Set read deadline
        conn.SetReadDeadline(time.Now().Add(ons.config.ReadTimeout))
        
        n, err := conn.Read(buffer)
        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                continue // Timeout, try again
            }
            break // Connection closed or error
        }
        
        if n > 0 {
            atomic.AddInt64(&ons.metrics.BytesReceived, int64(n))
            
            // Process request
            request := make([]byte, n)
            copy(request, buffer[:n])
            
            start := time.Now()
            response := ons.processRequest(request)
            latency := time.Since(start)
            
            // Update metrics
            atomic.AddInt64(&ons.metrics.RequestsProcessed, 1)
            ons.updateLatency(latency)
            
            // Send response
            conn.SetWriteDeadline(time.Now().Add(ons.config.WriteTimeout))
            written, err := conn.Write(response)
            if err != nil {
                break
            }
            
            atomic.AddInt64(&ons.metrics.BytesSent, int64(written))
        }
    }
}

func (ons *OptimizedNetworkServer) handleEpollConnection(conn net.Conn) {
    defer func() {
        ons.epoll.RemoveConnection(conn)
        conn.Close()
        atomic.AddInt64(&ons.metrics.ActiveConnections, -1)
    }()
    
    buffer := ons.readBuffer.Get()
    defer ons.readBuffer.Put(buffer)
    
    for {
        select {
        case event := <-ons.epoll.eventChan:
            if event.Events&syscall.EPOLLIN != 0 {
                // Data available for reading
                n, err := conn.Read(buffer)
                if err != nil {
                    return
                }
                
                if n > 0 {
                    atomic.AddInt64(&ons.metrics.BytesReceived, int64(n))
                    
                    // Submit to worker pool
                    workItem := &WorkItem{
                        conn:     conn,
                        request:  make([]byte, n),
                        response: make(chan []byte, 1),
                        start:    time.Now(),
                    }
                    copy(workItem.request, buffer[:n])
                    
                    ons.workerPool.Submit(workItem)
                    
                    // Wait for response
                    select {
                    case response := <-workItem.response:
                        conn.Write(response)
                        atomic.AddInt64(&ons.metrics.BytesSent, int64(len(response)))
                        
                    case <-time.After(ons.config.WriteTimeout):
                        // Response timeout
                        return
                    }
                    
                    // Update metrics
                    atomic.AddInt64(&ons.metrics.RequestsProcessed, 1)
                    ons.updateLatency(time.Since(workItem.start))
                }
            }
            
        case <-time.After(ons.config.ReadTimeout):
            // Connection timeout
            return
        }
    }
}

func (ons *OptimizedNetworkServer) processRequest(request []byte) []byte {
    // Simple echo for demonstration
    response := make([]byte, len(request))
    copy(response, request)
    return response
}

func (ons *OptimizedNetworkServer) updateLatency(latency time.Duration) {
    // Simple exponential moving average
    currentAvg := atomic.LoadInt64((*int64)(&ons.metrics.AvgRequestLatency))
    alpha := 0.1
    newAvg := time.Duration(float64(currentAvg)*(1-alpha) + float64(latency)*alpha)
    atomic.StoreInt64((*int64)(&ons.metrics.AvgRequestLatency), int64(newAvg))
}

// Example request processor
type EchoProcessor struct{}

func (ep *EchoProcessor) ProcessRequest(request []byte) []byte {
    // Echo the request back
    response := make([]byte, len(request))
    copy(response, request)
    return response
}

// Metrics collection
func (ons *OptimizedNetworkServer) GetMetrics() NetworkMetrics {
    return NetworkMetrics{
        ActiveConnections: atomic.LoadInt64(&ons.metrics.ActiveConnections),
        TotalConnections:  atomic.LoadInt64(&ons.metrics.TotalConnections),
        BytesReceived:     atomic.LoadInt64(&ons.metrics.BytesReceived),
        BytesSent:         atomic.LoadInt64(&ons.metrics.BytesSent),
        RequestsProcessed: atomic.LoadInt64(&ons.metrics.RequestsProcessed),
        AvgRequestLatency: time.Duration(atomic.LoadInt64((*int64)(&ons.metrics.AvgRequestLatency))),
    }
}

// Shutdown cleanup
func (ons *OptimizedNetworkServer) Shutdown() error {
    if ons.listener != nil {
        ons.listener.Close()
    }
    
    if ons.epoll != nil {
        ons.epoll.running = false
        syscall.Close(ons.epoll.epollFd)
    }
    
    return nil
}

// Placeholder for reuseport package functionality
var reuseport = struct {
    Listen func(network, address string) (net.Listener, error)
}{
    Listen: net.Listen, // Fallback to standard net.Listen
}
```

## Serialization Optimization

### High-Performance Data Serialization
Implement optimized serialization techniques for minimal overhead:

```go
// Advanced serialization optimization
type SerializationOptimizer struct {
    encoders map[string]Encoder
    decoders map[string]Decoder
    pools    map[string]*sync.Pool
    metrics  SerializationMetrics
    config   SerializationConfig
}

type SerializationConfig struct {
    DefaultFormat    string            `json:"default_format"`
    BufferSize       int              `json:"buffer_size"`
    PoolSize         int              `json:"pool_size"`
    CompressionLevel int              `json:"compression_level"`
    UseCompression   bool             `json:"use_compression"`
    EnableCaching    bool             `json:"enable_caching"`
    Formats          map[string]bool  `json:"enabled_formats"`
}

type SerializationMetrics struct {
    EncodeOperations  int64         `json:"encode_operations"`
    DecodeOperations  int64         `json:"decode_operations"`
    BytesEncoded      int64         `json:"bytes_encoded"`
    BytesDecoded      int64         `json:"bytes_decoded"`
    AvgEncodeTime     time.Duration `json:"avg_encode_time"`
    AvgDecodeTime     time.Duration `json:"avg_decode_time"`
    CompressionRatio  float64       `json:"compression_ratio"`
    CacheHitRate      float64       `json:"cache_hit_rate"`
}

type Encoder interface {
    Encode(data interface{}) ([]byte, error)
    GetContentType() string
}

type Decoder interface {
    Decode(data []byte, target interface{}) error
    GetContentType() string
}

// High-performance binary encoder
type BinaryEncoder struct {
    buffer *bytes.Buffer
    pool   *sync.Pool
}

func NewBinaryEncoder(bufferSize int) *BinaryEncoder {
    pool := &sync.Pool{
        New: func() interface{} {
            return bytes.NewBuffer(make([]byte, 0, bufferSize))
        },
    }
    
    return &BinaryEncoder{
        pool: pool,
    }
}

func (be *BinaryEncoder) Encode(data interface{}) ([]byte, error) {
    buffer := be.pool.Get().(*bytes.Buffer)
    defer func() {
        buffer.Reset()
        be.pool.Put(buffer)
    }()
    
    err := binary.Write(buffer, binary.LittleEndian, data)
    if err != nil {
        return nil, err
    }
    
    result := make([]byte, buffer.Len())
    copy(result, buffer.Bytes())
    
    return result, nil
}

func (be *BinaryEncoder) GetContentType() string {
    return "application/octet-stream"
}

// High-performance binary decoder
type BinaryDecoder struct {
    pool *sync.Pool
}

func NewBinaryDecoder(bufferSize int) *BinaryDecoder {
    pool := &sync.Pool{
        New: func() interface{} {
            return bytes.NewReader(make([]byte, bufferSize))
        },
    }
    
    return &BinaryDecoder{
        pool: pool,
    }
}

func (bd *BinaryDecoder) Decode(data []byte, target interface{}) error {
    reader := bytes.NewReader(data)
    return binary.Read(reader, binary.LittleEndian, target)
}

func (bd *BinaryDecoder) GetContentType() string {
    return "application/octet-stream"
}

// Optimized JSON encoder with pooling
type OptimizedJSONEncoder struct {
    pool *sync.Pool
}

func NewOptimizedJSONEncoder(bufferSize int) *OptimizedJSONEncoder {
    pool := &sync.Pool{
        New: func() interface{} {
            buffer := make([]byte, 0, bufferSize)
            return &buffer
        },
    }
    
    return &OptimizedJSONEncoder{
        pool: pool,
    }
}

func (oje *OptimizedJSONEncoder) Encode(data interface{}) ([]byte, error) {
    bufferPtr := oje.pool.Get().(*[]byte)
    defer func() {
        *bufferPtr = (*bufferPtr)[:0] // Reset length but keep capacity
        oje.pool.Put(bufferPtr)
    }()
    
    // Use jsoniter for better performance
    encoded, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    
    // Copy to result buffer
    result := make([]byte, len(encoded))
    copy(result, encoded)
    
    return result, nil
}

func (oje *OptimizedJSONEncoder) GetContentType() string {
    return "application/json"
}

// MessagePack encoder for efficient binary serialization
type MessagePackEncoder struct {
    buffer *bytes.Buffer
    pool   *sync.Pool
}

func NewMessagePackEncoder(bufferSize int) *MessagePackEncoder {
    pool := &sync.Pool{
        New: func() interface{} {
            return bytes.NewBuffer(make([]byte, 0, bufferSize))
        },
    }
    
    return &MessagePackEncoder{
        pool: pool,
    }
}

func (mpe *MessagePackEncoder) Encode(data interface{}) ([]byte, error) {
    buffer := mpe.pool.Get().(*bytes.Buffer)
    defer func() {
        buffer.Reset()
        mpe.pool.Put(buffer)
    }()
    
    // Simulate MessagePack encoding
    // In real implementation, use github.com/vmihailenco/msgpack
    encoded, err := json.Marshal(data) // Placeholder
    if err != nil {
        return nil, err
    }
    
    // MessagePack typically produces smaller output than JSON
    compressed := make([]byte, len(encoded)*3/4) // Simulate compression
    copy(compressed, encoded[:len(compressed)])
    
    return compressed, nil
}

func (mpe *MessagePackEncoder) GetContentType() string {
    return "application/x-msgpack"
}

// Protocol Buffers encoder (high performance)
type ProtobufEncoder struct{}

func NewProtobufEncoder() *ProtobufEncoder {
    return &ProtobufEncoder{}
}

func (pe *ProtobufEncoder) Encode(data interface{}) ([]byte, error) {
    // Simulate protobuf encoding
    // In real implementation, use github.com/golang/protobuf
    
    if pbMessage, ok := data.(ProtoMessage); ok {
        return pbMessage.Marshal()
    }
    
    return nil, fmt.Errorf("data does not implement ProtoMessage interface")
}

func (pe *ProtobufEncoder) GetContentType() string {
    return "application/x-protobuf"
}

// Interface for protobuf messages
type ProtoMessage interface {
    Marshal() ([]byte, error)
    Unmarshal([]byte) error
}

// Serialization cache for frequently accessed data
type SerializationCache struct {
    cache   map[string]*CachedSerialization
    maxSize int
    mu      sync.RWMutex
    lru     *SerializationLRU
}

type CachedSerialization struct {
    key         string
    data        []byte
    format      string
    timestamp   time.Time
    accessCount int64
    size        int64
    prev        *CachedSerialization
    next        *CachedSerialization
}

type SerializationLRU struct {
    head *CachedSerialization
    tail *CachedSerialization
}

func NewSerializationCache(maxSize int) *SerializationCache {
    return &SerializationCache{
        cache:   make(map[string]*CachedSerialization),
        maxSize: maxSize,
        lru:     &SerializationLRU{},
    }
}

func (sc *SerializationCache) Get(key, format string) []byte {
    sc.mu.RLock()
    defer sc.mu.RUnlock()
    
    cacheKey := fmt.Sprintf("%s:%s", key, format)
    entry, exists := sc.cache[cacheKey]
    if !exists {
        return nil
    }
    
    // Update access count and move to front
    atomic.AddInt64(&entry.accessCount, 1)
    entry.timestamp = time.Now()
    sc.moveToFront(entry)
    
    // Return copy
    result := make([]byte, len(entry.data))
    copy(result, entry.data)
    return result
}

func (sc *SerializationCache) Put(key, format string, data []byte) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    
    cacheKey := fmt.Sprintf("%s:%s", key, format)
    
    // Check if already exists
    if existing, exists := sc.cache[cacheKey]; exists {
        existing.data = make([]byte, len(data))
        copy(existing.data, data)
        existing.timestamp = time.Now()
        sc.moveToFront(existing)
        return
    }
    
    // Create new entry
    entry := &CachedSerialization{
        key:         cacheKey,
        data:        make([]byte, len(data)),
        format:      format,
        timestamp:   time.Now(),
        accessCount: 1,
        size:        int64(len(data)),
    }
    copy(entry.data, data)
    
    sc.cache[cacheKey] = entry
    sc.addToFront(entry)
    
    // Evict if necessary
    sc.evictIfNecessary()
}

func (sc *SerializationCache) moveToFront(entry *CachedSerialization) {
    sc.removeFromList(entry)
    sc.addToFront(entry)
}

func (sc *SerializationCache) addToFront(entry *CachedSerialization) {
    entry.next = sc.lru.head
    entry.prev = nil
    
    if sc.lru.head != nil {
        sc.lru.head.prev = entry
    }
    
    sc.lru.head = entry
    
    if sc.lru.tail == nil {
        sc.lru.tail = entry
    }
}

func (sc *SerializationCache) removeFromList(entry *CachedSerialization) {
    if entry.prev != nil {
        entry.prev.next = entry.next
    } else {
        sc.lru.head = entry.next
    }
    
    if entry.next != nil {
        entry.next.prev = entry.prev
    } else {
        sc.lru.tail = entry.prev
    }
}

func (sc *SerializationCache) evictIfNecessary() {
    for len(sc.cache) > sc.maxSize && sc.lru.tail != nil {
        victim := sc.lru.tail
        sc.removeFromList(victim)
        delete(sc.cache, victim.key)
    }
}

// Main serialization optimizer
func NewSerializationOptimizer(config SerializationConfig) *SerializationOptimizer {
    so := &SerializationOptimizer{
        config:   config,
        encoders: make(map[string]Encoder),
        decoders: make(map[string]Decoder),
        pools:    make(map[string]*sync.Pool),
    }
    
    // Register encoders/decoders
    if config.Formats["binary"] {
        so.encoders["binary"] = NewBinaryEncoder(config.BufferSize)
        so.decoders["binary"] = NewBinaryDecoder(config.BufferSize)
    }
    
    if config.Formats["json"] {
        so.encoders["json"] = NewOptimizedJSONEncoder(config.BufferSize)
    }
    
    if config.Formats["msgpack"] {
        so.encoders["msgpack"] = NewMessagePackEncoder(config.BufferSize)
    }
    
    if config.Formats["protobuf"] {
        so.encoders["protobuf"] = NewProtobufEncoder()
    }
    
    return so
}

func (so *SerializationOptimizer) Encode(data interface{}, format string) ([]byte, error) {
    start := time.Now()
    
    encoder, exists := so.encoders[format]
    if !exists {
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
    
    result, err := encoder.Encode(data)
    
    // Update metrics
    atomic.AddInt64(&so.metrics.EncodeOperations, 1)
    if err == nil {
        atomic.AddInt64(&so.metrics.BytesEncoded, int64(len(result)))
    }
    
    latency := time.Since(start)
    so.updateEncodeLatency(latency)
    
    return result, err
}

func (so *SerializationOptimizer) Decode(data []byte, target interface{}, format string) error {
    start := time.Now()
    
    decoder, exists := so.decoders[format]
    if !exists {
        return fmt.Errorf("unsupported format: %s", format)
    }
    
    err := decoder.Decode(data, target)
    
    // Update metrics
    atomic.AddInt64(&so.metrics.DecodeOperations, 1)
    if err == nil {
        atomic.AddInt64(&so.metrics.BytesDecoded, int64(len(data)))
    }
    
    latency := time.Since(start)
    so.updateDecodeLatency(latency)
    
    return err
}

func (so *SerializationOptimizer) updateEncodeLatency(latency time.Duration) {
    currentAvg := atomic.LoadInt64((*int64)(&so.metrics.AvgEncodeTime))
    newAvg := (time.Duration(currentAvg) + latency) / 2
    atomic.StoreInt64((*int64)(&so.metrics.AvgEncodeTime), int64(newAvg))
}

func (so *SerializationOptimizer) updateDecodeLatency(latency time.Duration) {
    currentAvg := atomic.LoadInt64((*int64)(&so.metrics.AvgDecodeTime))
    newAvg := (time.Duration(currentAvg) + latency) / 2
    atomic.StoreInt64((*int64)(&so.metrics.AvgDecodeTime), int64(newAvg))
}

func (so *SerializationOptimizer) GetMetrics() SerializationMetrics {
    return SerializationMetrics{
        EncodeOperations: atomic.LoadInt64(&so.metrics.EncodeOperations),
        DecodeOperations: atomic.LoadInt64(&so.metrics.DecodeOperations),
        BytesEncoded:     atomic.LoadInt64(&so.metrics.BytesEncoded),
        BytesDecoded:     atomic.LoadInt64(&so.metrics.BytesDecoded),
        AvgEncodeTime:    time.Duration(atomic.LoadInt64((*int64)(&so.metrics.AvgEncodeTime))),
        AvgDecodeTime:    time.Duration(atomic.LoadInt64((*int64)(&so.metrics.AvgDecodeTime))),
    }
}
```

I/O and network optimization requires careful attention to buffering strategies, connection management, and data serialization efficiency. By implementing sophisticated caching, pooling, and asynchronous processing techniques, applications can achieve significant performance improvements while maintaining reliability and scalability.

## Key Takeaways

1. **Implement intelligent buffering** - use adaptive buffer sizes and pooling
2. **Optimize network protocols** - leverage TCP options and efficient event handling
3. **Use efficient serialization** - choose appropriate formats for data characteristics
4. **Apply caching strategically** - cache frequently accessed data and operations
5. **Leverage asynchronous I/O** - minimize blocking operations with async patterns
6. **Monitor performance metrics** - track throughput, latency, and resource utilization
7. **Pool expensive resources** - reuse connections, buffers, and objects
8. **Optimize for specific workloads** - adapt techniques to application characteristics

Effective I/O and network optimization enables applications to handle high throughput and concurrent load while maintaining low latency and efficient resource utilization.
