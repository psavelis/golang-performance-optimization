# I/O Optimization

Comprehensive guide to optimizing input/output operations in Go applications. This section covers buffer management, network optimization, streaming, and advanced I/O patterns for maximum performance.

## Table of Contents

- [Introduction](#introduction)
- [Buffer Management](buffer-management.md)
- [Network Optimization](network-optimization.md)
- [Streaming](streaming.md)
- [Core I/O Patterns](#core-io-patterns)
- [File System Optimization](#file-system-optimization)
- [Database I/O](#database-io)
- [Monitoring and Metrics](#monitoring-and-metrics)
- [Best Practices](#best-practices)

## Introduction

I/O operations are often the primary bottleneck in Go applications. This guide provides comprehensive strategies for optimizing all aspects of I/O performance, from basic file operations to complex network protocols.

### I/O Performance Framework

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "net"
    "os"
    "sync"
    "sync/atomic"
    "time"
)

// IOOptimizer provides comprehensive I/O optimization capabilities
type IOOptimizer struct {
    config       IOConfig
    metrics      *IOMetrics
    bufferPool   *BufferPool
    connPool     *ConnectionPool
    rateLimiter  *IORate Limiter
    monitor      *IOMonitor
    mu           sync.RWMutex
}

// IOConfig contains I/O optimization configuration
type IOConfig struct {
    BufferSize       int
    ReadBufferSize   int
    WriteBufferSize  int
    MaxConnections   int
    ConnectionTimeout time.Duration
    ReadTimeout      time.Duration
    WriteTimeout     time.Duration
    EnableMetrics    bool
    EnableCompression bool
    CompressionLevel int
    IOWorkers        int
    BatchSize        int
}

// IOMetrics tracks I/O performance metrics
type IOMetrics struct {
    BytesRead         int64
    BytesWritten      int64
    ReadOperations    int64
    WriteOperations   int64
    ReadLatency       time.Duration
    WriteLatency      time.Duration
    ThroughputRead    float64
    ThroughputWrite   float64
    ActiveConnections int32
    ErrorCount        int64
    CacheHitRate      float64
    BufferUtilization float64
}

// BufferPool manages reusable buffers for I/O operations
type BufferPool struct {
    pools       map[int]*sync.Pool
    sizes       []int
    metrics     BufferPoolMetrics
    allocations int64
    deallocations int64
}

// BufferPoolMetrics tracks buffer pool performance
type BufferPoolMetrics struct {
    PoolHits      int64
    PoolMisses    int64
    TotalBuffers  int64
    MemoryUsage   int64
    ReuseRate     float64
}

// ConnectionPool manages reusable network connections
type ConnectionPool struct {
    pools        map[string]*ConnPool
    factory      ConnectionFactory
    maxConns     int
    idleTimeout  time.Duration
    maxLifetime  time.Duration
    mu           sync.RWMutex
}

// ConnPool represents a pool of connections to a specific address
type ConnPool struct {
    conns        chan *PooledConnection
    factory      ConnectionFactory
    address      string
    maxConns     int
    currentConns int32
    mu           sync.Mutex
}

// PooledConnection wraps a connection with pool metadata
type PooledConnection struct {
    Conn        net.Conn
    CreatedAt   time.Time
    LastUsed    time.Time
    UseCount    int64
    Address     string
}

// ConnectionFactory creates new connections
type ConnectionFactory interface {
    CreateConnection(address string) (net.Conn, error)
    ValidateConnection(conn net.Conn) error
}

// IORateLimiter controls I/O operation rates
type IORate Limiter struct {
    readLimiter    *TokenBucket
    writeLimiter   *TokenBucket
    enabled        bool
    readLimit      int64
    writeLimit     int64
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
    capacity    int64
    tokens      int64
    refillRate  int64
    lastRefill  time.Time
    mu          sync.Mutex
}

// IOMonitor monitors I/O operations in real-time
type IOMonitor struct {
    operations   chan IOOperation
    metrics      *IOMetrics
    collectors   []MetricsCollector
    alerting     *IOAlerting
    running      bool
    mu           sync.RWMutex
}

// IOOperation represents a monitored I/O operation
type IOOperation struct {
    Type      OperationType
    Size      int64
    Duration  time.Duration
    Error     error
    Timestamp time.Time
    Context   map[string]interface{}
}

// OperationType defines types of I/O operations
type OperationType int

const (
    ReadOperation OperationType = iota
    WriteOperation
    SeekOperation
    FlushOperation
    CloseOperation
)

// MetricsCollector defines interface for metrics collection
type MetricsCollector interface {
    CollectMetrics(operation IOOperation)
    GetMetrics() map[string]interface{}
    Reset()
}

// IOAlerting provides alerting for I/O issues
type IOAlerting struct {
    thresholds   AlertThresholds
    alerts       chan Alert
    handlers     []AlertHandler
}

// AlertThresholds defines alerting thresholds
type AlertThresholds struct {
    MaxLatency       time.Duration
    MinThroughput    float64
    MaxErrorRate     float64
    MaxBufferUsage   float64
}

// Alert represents an I/O alert
type Alert struct {
    Type        AlertType
    Severity    Severity
    Message     string
    Metrics     map[string]interface{}
    Timestamp   time.Time
}

// AlertType defines alert types
type AlertType int

const (
    LatencyAlert AlertType = iota
    ThroughputAlert
    ErrorAlert
    ResourceAlert
)

// Severity defines alert severity levels
type Severity int

const (
    Info Severity = iota
    Warning
    Error
    Critical
)

// AlertHandler handles I/O alerts
type AlertHandler interface {
    HandleAlert(alert Alert) error
}

// NewIOOptimizer creates a new I/O optimizer
func NewIOOptimizer(config IOConfig) *IOOptimizer {
    return &IOOptimizer{
        config:      config,
        metrics:     &IOMetrics{},
        bufferPool:  NewBufferPool(),
        connPool:    NewConnectionPool(config.MaxConnections),
        rateLimiter: NewIORate Limiter(config),
        monitor:     NewIOMonitor(),
    }
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
    pool := &BufferPool{
        pools: make(map[int]*sync.Pool),
        sizes: []int{1024, 4096, 16384, 65536, 262144},
    }
    
    // Initialize pools for different buffer sizes
    for _, size := range pool.sizes {
        sz := size // Capture for closure
        pool.pools[size] = &sync.Pool{
            New: func() interface{} {
                atomic.AddInt64(&pool.allocations, 1)
                return make([]byte, sz)
            },
        }
    }
    
    return pool
}

// GetBuffer returns a buffer of appropriate size
func (bp *BufferPool) GetBuffer(size int) []byte {
    targetSize := bp.findOptimalSize(size)
    
    if pool, exists := bp.pools[targetSize]; exists {
        atomic.AddInt64(&bp.metrics.PoolHits, 1)
        return pool.Get().([]byte)[:size]
    }
    
    atomic.AddInt64(&bp.metrics.PoolMisses, 1)
    return make([]byte, size)
}

// PutBuffer returns a buffer to the pool
func (bp *BufferPool) PutBuffer(buf []byte) {
    size := cap(buf)
    targetSize := bp.findOptimalSize(size)
    
    if pool, exists := bp.pools[targetSize]; exists {
        // Reset buffer
        buf = buf[:cap(buf)]
        for i := range buf {
            buf[i] = 0
        }
        
        pool.Put(buf)
        atomic.AddInt64(&bp.deallocations, 1)
    }
}

// findOptimalSize finds the optimal buffer size for pooling
func (bp *BufferPool) findOptimalSize(size int) int {
    for _, poolSize := range bp.sizes {
        if size <= poolSize {
            return poolSize
        }
    }
    return bp.sizes[len(bp.sizes)-1]
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxConns int) *ConnectionPool {
    return &ConnectionPool{
        pools:       make(map[string]*ConnPool),
        maxConns:    maxConns,
        idleTimeout: 30 * time.Second,
        maxLifetime: 60 * time.Minute,
    }
}

// GetConnection gets a connection from the pool
func (cp *ConnectionPool) GetConnection(address string) (*PooledConnection, error) {
    cp.mu.RLock()
    pool, exists := cp.pools[address]
    cp.mu.RUnlock()
    
    if !exists {
        cp.mu.Lock()
        if pool, exists = cp.pools[address]; !exists {
            pool = &ConnPool{
                conns:    make(chan *PooledConnection, cp.maxConns),
                address:  address,
                maxConns: cp.maxConns,
                factory:  cp.factory,
            }
            cp.pools[address] = pool
        }
        cp.mu.Unlock()
    }
    
    return pool.getConnection()
}

// getConnection gets a connection from a specific pool
func (cp *ConnPool) getConnection() (*PooledConnection, error) {
    select {
    case conn := <-cp.conns:
        // Validate connection before use
        if cp.factory != nil && cp.factory.ValidateConnection(conn.Conn) != nil {
            conn.Conn.Close()
            return cp.createNewConnection()
        }
        
        conn.LastUsed = time.Now()
        atomic.AddInt64(&conn.UseCount, 1)
        return conn, nil
        
    default:
        // No idle connections available
        if atomic.LoadInt32(&cp.currentConns) < int32(cp.maxConns) {
            return cp.createNewConnection()
        }
        
        // Wait for a connection to become available
        select {
        case conn := <-cp.conns:
            conn.LastUsed = time.Now()
            atomic.AddInt64(&conn.UseCount, 1)
            return conn, nil
        case <-time.After(10 * time.Second):
            return nil, fmt.Errorf("connection pool timeout for %s", cp.address)
        }
    }
}

// createNewConnection creates a new pooled connection
func (cp *ConnPool) createNewConnection() (*PooledConnection, error) {
    if cp.factory == nil {
        return nil, fmt.Errorf("no connection factory configured")
    }
    
    conn, err := cp.factory.CreateConnection(cp.address)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection: %w", err)
    }
    
    atomic.AddInt32(&cp.currentConns, 1)
    
    return &PooledConnection{
        Conn:      conn,
        CreatedAt: time.Now(),
        LastUsed:  time.Now(),
        UseCount:  1,
        Address:   cp.address,
    }, nil
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(conn *PooledConnection) error {
    if conn == nil {
        return nil
    }
    
    cp.mu.RLock()
    pool, exists := cp.pools[conn.Address]
    cp.mu.RUnlock()
    
    if !exists {
        conn.Conn.Close()
        return fmt.Errorf("pool not found for address %s", conn.Address)
    }
    
    // Check if connection is still valid
    if time.Since(conn.CreatedAt) > cp.maxLifetime {
        conn.Conn.Close()
        atomic.AddInt32(&pool.currentConns, -1)
        return nil
    }
    
    select {
    case pool.conns <- conn:
        return nil
    default:
        // Pool is full, close the connection
        conn.Conn.Close()
        atomic.AddInt32(&pool.currentConns, -1)
        return nil
    }
}

// NewIORate Limiter creates a new I/O rate limiter
func NewIORate Limiter(config IOConfig) *IORate Limiter {
    return &IORate Limiter{
        readLimiter:  NewTokenBucket(1000000, 100000),  // 1MB capacity, 100KB/s refill
        writeLimiter: NewTokenBucket(1000000, 100000),  // 1MB capacity, 100KB/s refill
        enabled:      true,
        readLimit:    100000,
        writeLimit:   100000,
    }
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
    return &TokenBucket{
        capacity:   capacity,
        tokens:     capacity,
        refillRate: refillRate,
        lastRefill: time.Now(),
    }
}

// TakeTokens attempts to take tokens from the bucket
func (tb *TokenBucket) TakeTokens(tokens int64) bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    tb.refill()
    
    if tb.tokens >= tokens {
        tb.tokens -= tokens
        return true
    }
    
    return false
}

// refill refills the token bucket
func (tb *TokenBucket) refill() {
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)
    tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate
    
    if tokensToAdd > 0 {
        tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
        tb.lastRefill = now
    }
}

// WaitForTokens waits for tokens to become available
func (tb *TokenBucket) WaitForTokens(tokens int64) error {
    for !tb.TakeTokens(tokens) {
        time.Sleep(10 * time.Millisecond)
    }
    return nil
}

// OptimizedReader provides optimized reading capabilities
type OptimizedReader struct {
    reader      io.Reader
    buffer      []byte
    bufferPool  *BufferPool
    rateLimiter *IORate Limiter
    metrics     *IOMetrics
    monitor     *IOMonitor
}

// NewOptimizedReader creates a new optimized reader
func NewOptimizedReader(reader io.Reader, optimizer *IOOptimizer) *OptimizedReader {
    return &OptimizedReader{
        reader:      reader,
        buffer:      optimizer.bufferPool.GetBuffer(optimizer.config.ReadBufferSize),
        bufferPool:  optimizer.bufferPool,
        rateLimiter: optimizer.rateLimiter,
        metrics:     optimizer.metrics,
        monitor:     optimizer.monitor,
    }
}

// Read performs optimized reading with metrics and rate limiting
func (or *OptimizedReader) Read(p []byte) (int, error) {
    startTime := time.Now()
    
    // Apply rate limiting
    if or.rateLimiter.enabled {
        if err := or.rateLimiter.readLimiter.WaitForTokens(int64(len(p))); err != nil {
            return 0, err
        }
    }
    
    n, err := or.reader.Read(p)
    
    duration := time.Since(startTime)
    
    // Update metrics
    atomic.AddInt64(&or.metrics.BytesRead, int64(n))
    atomic.AddInt64(&or.metrics.ReadOperations, 1)
    
    // Calculate latency
    currentLatency := or.metrics.ReadLatency
    newLatency := time.Duration((currentLatency.Nanoseconds() + duration.Nanoseconds()) / 2)
    or.metrics.ReadLatency = newLatency
    
    // Monitor operation
    if or.monitor != nil {
        operation := IOOperation{
            Type:      ReadOperation,
            Size:      int64(n),
            Duration:  duration,
            Error:     err,
            Timestamp: startTime,
        }
        or.monitor.RecordOperation(operation)
    }
    
    if err != nil {
        atomic.AddInt64(&or.metrics.ErrorCount, 1)
    }
    
    return n, err
}

// Close closes the optimized reader and returns buffers
func (or *OptimizedReader) Close() error {
    if or.buffer != nil {
        or.bufferPool.PutBuffer(or.buffer)
        or.buffer = nil
    }
    
    if closer, ok := or.reader.(io.Closer); ok {
        return closer.Close()
    }
    
    return nil
}

// OptimizedWriter provides optimized writing capabilities
type OptimizedWriter struct {
    writer      io.Writer
    buffer      []byte
    bufferPool  *BufferPool
    rateLimiter *IORate Limiter
    metrics     *IOMetrics
    monitor     *IOMonitor
    buffered    *bufio.Writer
}

// NewOptimizedWriter creates a new optimized writer
func NewOptimizedWriter(writer io.Writer, optimizer *IOOptimizer) *OptimizedWriter {
    buffer := optimizer.bufferPool.GetBuffer(optimizer.config.WriteBufferSize)
    
    return &OptimizedWriter{
        writer:      writer,
        buffer:      buffer,
        bufferPool:  optimizer.bufferPool,
        rateLimiter: optimizer.rateLimiter,
        metrics:     optimizer.metrics,
        monitor:     optimizer.monitor,
        buffered:    bufio.NewWriterSize(writer, len(buffer)),
    }
}

// Write performs optimized writing with metrics and rate limiting
func (ow *OptimizedWriter) Write(p []byte) (int, error) {
    startTime := time.Now()
    
    // Apply rate limiting
    if ow.rateLimiter.enabled {
        if err := ow.rateLimiter.writeLimiter.WaitForTokens(int64(len(p))); err != nil {
            return 0, err
        }
    }
    
    n, err := ow.buffered.Write(p)
    
    duration := time.Since(startTime)
    
    // Update metrics
    atomic.AddInt64(&ow.metrics.BytesWritten, int64(n))
    atomic.AddInt64(&ow.metrics.WriteOperations, 1)
    
    // Calculate latency
    currentLatency := ow.metrics.WriteLatency
    newLatency := time.Duration((currentLatency.Nanoseconds() + duration.Nanoseconds()) / 2)
    ow.metrics.WriteLatency = newLatency
    
    // Monitor operation
    if ow.monitor != nil {
        operation := IOOperation{
            Type:      WriteOperation,
            Size:      int64(n),
            Duration:  duration,
            Error:     err,
            Timestamp: startTime,
        }
        ow.monitor.RecordOperation(operation)
    }
    
    if err != nil {
        atomic.AddInt64(&ow.metrics.ErrorCount, 1)
    }
    
    return n, err
}

// Flush flushes the buffered writer
func (ow *OptimizedWriter) Flush() error {
    startTime := time.Now()
    err := ow.buffered.Flush()
    duration := time.Since(startTime)
    
    // Monitor flush operation
    if ow.monitor != nil {
        operation := IOOperation{
            Type:      FlushOperation,
            Duration:  duration,
            Error:     err,
            Timestamp: startTime,
        }
        ow.monitor.RecordOperation(operation)
    }
    
    return err
}

// Close closes the optimized writer and returns buffers
func (ow *OptimizedWriter) Close() error {
    if err := ow.Flush(); err != nil {
        return err
    }
    
    if ow.buffer != nil {
        ow.bufferPool.PutBuffer(ow.buffer)
        ow.buffer = nil
    }
    
    if closer, ok := ow.writer.(io.Closer); ok {
        return closer.Close()
    }
    
    return nil
}

// NewIOMonitor creates a new I/O monitor
func NewIOMonitor() *IOMonitor {
    return &IOMonitor{
        operations: make(chan IOOperation, 10000),
        metrics:    &IOMetrics{},
        collectors: make([]MetricsCollector, 0),
        alerting:   NewIOAlerting(),
    }
}

// RecordOperation records an I/O operation for monitoring
func (iom *IOMonitor) RecordOperation(operation IOOperation) {
    select {
    case iom.operations <- operation:
    default:
        // Channel full, drop operation (non-blocking)
    }
}

// Start starts the I/O monitor
func (iom *IOMonitor) Start() error {
    iom.mu.Lock()
    defer iom.mu.Unlock()
    
    if iom.running {
        return fmt.Errorf("monitor already running")
    }
    
    iom.running = true
    go iom.monitorLoop()
    
    return nil
}

// monitorLoop processes I/O operations
func (iom *IOMonitor) monitorLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for iom.running {
        select {
        case operation := <-iom.operations:
            iom.processOperation(operation)
            
        case <-ticker.C:
            iom.updateThroughputMetrics()
            iom.checkAlerts()
        }
    }
}

// processOperation processes a single I/O operation
func (iom *IOMonitor) processOperation(operation IOOperation) {
    // Update collectors
    for _, collector := range iom.collectors {
        collector.CollectMetrics(operation)
    }
    
    // Check for immediate alerts
    if operation.Error != nil {
        alert := Alert{
            Type:      ErrorAlert,
            Severity:  Error,
            Message:   fmt.Sprintf("I/O error: %v", operation.Error),
            Timestamp: operation.Timestamp,
        }
        iom.alerting.SendAlert(alert)
    }
    
    if operation.Duration > iom.alerting.thresholds.MaxLatency {
        alert := Alert{
            Type:      LatencyAlert,
            Severity:  Warning,
            Message:   fmt.Sprintf("High I/O latency: %v", operation.Duration),
            Timestamp: operation.Timestamp,
        }
        iom.alerting.SendAlert(alert)
    }
}

// updateThroughputMetrics updates throughput metrics
func (iom *IOMonitor) updateThroughputMetrics() {
    // Calculate throughput based on recent operations
    // This is a simplified implementation
    iom.metrics.ThroughputRead = float64(atomic.LoadInt64(&iom.metrics.BytesRead))
    iom.metrics.ThroughputWrite = float64(atomic.LoadInt64(&iom.metrics.BytesWritten))
}

// checkAlerts checks for alert conditions
func (iom *IOMonitor) checkAlerts() {
    // Check throughput alerts
    if iom.metrics.ThroughputRead < iom.alerting.thresholds.MinThroughput {
        alert := Alert{
            Type:      ThroughputAlert,
            Severity:  Warning,
            Message:   "Low read throughput",
            Timestamp: time.Now(),
        }
        iom.alerting.SendAlert(alert)
    }
}

// NewIOAlerting creates a new I/O alerting system
func NewIOAlerting() *IOAlerting {
    return &IOAlerting{
        thresholds: AlertThresholds{
            MaxLatency:     100 * time.Millisecond,
            MinThroughput:  1000000, // 1MB/s
            MaxErrorRate:   0.01,    // 1%
            MaxBufferUsage: 0.8,     // 80%
        },
        alerts:   make(chan Alert, 1000),
        handlers: make([]AlertHandler, 0),
    }
}

// SendAlert sends an alert
func (ioa *IOAlerting) SendAlert(alert Alert) {
    select {
    case ioa.alerts <- alert:
    default:
        // Alert channel full, drop alert
    }
    
    // Notify handlers
    for _, handler := range ioa.handlers {
        go handler.HandleAlert(alert)
    }
}

// GetMetrics returns current I/O metrics
func (ioo *IOOptimizer) GetMetrics() IOMetrics {
    ioo.mu.RLock()
    defer ioo.mu.RUnlock()
    
    // Calculate derived metrics
    metrics := *ioo.metrics
    
    if metrics.ReadOperations > 0 {
        metrics.ThroughputRead = float64(metrics.BytesRead) / float64(metrics.ReadOperations)
    }
    
    if metrics.WriteOperations > 0 {
        metrics.ThroughputWrite = float64(metrics.BytesWritten) / float64(metrics.WriteOperations)
    }
    
    return metrics
}

// Helper function
func min(a, b int64) int64 {
    if a < b {
        return a
    }
    return b
}
```

## Core I/O Patterns

Essential I/O patterns for optimal performance in different scenarios.

### Streaming I/O

Optimized streaming patterns for continuous data processing.

### Batch I/O

Efficient batch processing for high-throughput scenarios.

### Async I/O

Non-blocking I/O patterns for concurrent applications.

## File System Optimization

Optimizing file system operations for better performance.

### Memory-Mapped Files

Using memory mapping for efficient file access.

### Direct I/O

Bypassing system caches for predictable performance.

### File Handle Management

Efficient management of file descriptors and handles.

## Database I/O

Optimizing database interactions for performance.

### Connection Pooling

Efficient database connection management.

### Query Optimization

Optimizing database queries and transactions.

### Batch Operations

Efficient batch database operations.

## Best Practices

1. **Buffer Management**: Use pooled buffers to reduce allocations
2. **Connection Reuse**: Pool connections for network operations  
3. **Rate Limiting**: Control I/O rates to prevent system overload
4. **Monitoring**: Continuously monitor I/O performance metrics
5. **Error Handling**: Implement robust error handling and retries
6. **Resource Cleanup**: Properly clean up I/O resources
7. **Compression**: Use compression for network and storage I/O
8. **Caching**: Implement intelligent caching strategies

## Related Sections

- [Buffer Management](buffer-management.md) - Advanced buffer management techniques
- [Network Optimization](network-optimization.md) - Network-specific optimizations
- [Streaming](streaming.md) - Streaming data processing patterns
- [Memory Optimization](../memory/README.md) - Memory-related optimizations
- [Concurrency](../concurrency/README.md) - Concurrent I/O patterns

This I/O optimization guide provides the foundation for building high-performance Go applications with efficient I/O operations.
