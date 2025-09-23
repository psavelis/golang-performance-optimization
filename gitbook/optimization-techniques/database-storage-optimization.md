# Database and Storage Optimization

Database and storage operations are critical performance bottlenecks in most applications. This chapter explores advanced techniques for optimizing database connections, query performance, caching strategies, and storage access patterns to achieve maximum throughput while maintaining data consistency and reliability.

## Connection Pool Optimization

### Advanced Database Connection Management
Implement sophisticated connection pooling with adaptive scaling and intelligent load balancing:

```go
package database_optimization

import (
    "context"
    "database/sql"
    "database/sql/driver"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// Advanced connection pool with dynamic scaling and load balancing
type OptimizedConnectionPool struct {
    config          PoolConfig
    connections     []*PooledConnection
    available       chan *PooledConnection
    totalConns      int32
    activeConns     int32
    metrics         PoolMetrics
    healthChecker   *HealthChecker
    loadBalancer    *LoadBalancer
    circuitBreaker  *CircuitBreaker
    mu              sync.RWMutex
    shutdown        chan struct{}
    scalingTicker   *time.Ticker
}

type PoolConfig struct {
    MinConnections      int           `json:"min_connections"`
    MaxConnections      int           `json:"max_connections"`
    ConnectionTimeout   time.Duration `json:"connection_timeout"`
    IdleTimeout         time.Duration `json:"idle_timeout"`
    LifetimeTimeout     time.Duration `json:"lifetime_timeout"`
    HealthCheckInterval time.Duration `json:"health_check_interval"`
    ScalingInterval     time.Duration `json:"scaling_interval"`
    AcquisitionTimeout  time.Duration `json:"acquisition_timeout"`
    RetryAttempts       int           `json:"retry_attempts"`
    RetryDelay          time.Duration `json:"retry_delay"`
    EnableCircuitBreaker bool         `json:"enable_circuit_breaker"`
    MaxIdleTime         time.Duration `json:"max_idle_time"`
}

type PoolMetrics struct {
    TotalConnections     int32         `json:"total_connections"`
    ActiveConnections    int32         `json:"active_connections"`
    IdleConnections      int32         `json:"idle_connections"`
    ConnectionsCreated   int64         `json:"connections_created"`
    ConnectionsDestroyed int64         `json:"connections_destroyed"`
    ConnectionErrors     int64         `json:"connection_errors"`
    AcquisitionTime      time.Duration `json:"avg_acquisition_time"`
    QueryLatency         time.Duration `json:"avg_query_latency"`
    ThroughputQPS        float64       `json:"throughput_qps"`
    ConnectionUtilization float64      `json:"connection_utilization"`
}

type PooledConnection struct {
    id           int32
    conn         *sql.Conn
    db           *sql.DB
    created      time.Time
    lastUsed     time.Time
    inUse        bool
    healthy      bool
    queryCount   int64
    errorCount   int64
    totalLatency time.Duration
    mu           sync.Mutex
}

type HealthChecker struct {
    pool        *OptimizedConnectionPool
    interval    time.Duration
    timeout     time.Duration
    testQuery   string
    running     bool
    mu          sync.Mutex
}

type LoadBalancer struct {
    strategy    LoadBalanceStrategy
    connections []*PooledConnection
    weights     map[int32]float64
    mu          sync.RWMutex
}

type LoadBalanceStrategy int

const (
    RoundRobin LoadBalanceStrategy = iota
    LeastConnections
    WeightedRoundRobin
    LatencyBased
    AdaptiveLatency
)

type CircuitBreaker struct {
    maxFailures     int
    timeout         time.Duration
    failures        int32
    lastFailureTime time.Time
    state           CircuitState
    mu              sync.RWMutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func NewOptimizedConnectionPool(dsn string, config PoolConfig) (*OptimizedConnectionPool, error) {
    pool := &OptimizedConnectionPool{
        config:      config,
        connections: make([]*PooledConnection, 0, config.MaxConnections),
        available:   make(chan *PooledConnection, config.MaxConnections),
        shutdown:    make(chan struct{}),
        healthChecker: &HealthChecker{
            interval:  config.HealthCheckInterval,
            timeout:   config.ConnectionTimeout,
            testQuery: "SELECT 1",
        },
        loadBalancer: &LoadBalancer{
            strategy: AdaptiveLatency,
            weights:  make(map[int32]float64),
        },
    }
    
    if config.EnableCircuitBreaker {
        pool.circuitBreaker = &CircuitBreaker{
            maxFailures: 5,
            timeout:     30 * time.Second,
            state:       Closed,
        }
    }
    
    pool.healthChecker.pool = pool
    pool.loadBalancer.connections = pool.connections
    
    // Initialize minimum connections
    for i := 0; i < config.MinConnections; i++ {
        conn, err := pool.createConnection(dsn)
        if err != nil {
            pool.Close()
            return nil, fmt.Errorf("failed to create initial connection %d: %v", i, err)
        }
        
        pool.connections = append(pool.connections, conn)
        pool.available <- conn
        atomic.AddInt32(&pool.totalConns, 1)
    }
    
    // Start background tasks
    go pool.healthChecker.start()
    go pool.startAutoScaling()
    go pool.startMetricsCollection()
    
    return pool, nil
}

func (pool *OptimizedConnectionPool) createConnection(dsn string) (*PooledConnection, error) {
    db, err := sql.Open("postgres", dsn) // Example with PostgreSQL
    if err != nil {
        return nil, err
    }
    
    // Configure database connection
    db.SetMaxOpenConns(1) // One connection per PooledConnection
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(pool.config.LifetimeTimeout)
    db.SetConnMaxIdleTime(pool.config.IdleTimeout)
    
    ctx, cancel := context.WithTimeout(context.Background(), pool.config.ConnectionTimeout)
    defer cancel()
    
    conn, err := db.Conn(ctx)
    if err != nil {
        db.Close()
        return nil, err
    }
    
    // Test connection
    if err := conn.PingContext(ctx); err != nil {
        conn.Close()
        db.Close()
        return nil, err
    }
    
    pooledConn := &PooledConnection{
        id:       atomic.AddInt32(&pool.totalConns, 1),
        conn:     conn,
        db:       db,
        created:  time.Now(),
        lastUsed: time.Now(),
        healthy:  true,
    }
    
    atomic.AddInt64(&pool.metrics.ConnectionsCreated, 1)
    
    return pooledConn, nil
}

func (pool *OptimizedConnectionPool) AcquireConnection(ctx context.Context) (*PooledConnection, error) {
    start := time.Now()
    
    // Check circuit breaker
    if pool.circuitBreaker != nil && !pool.circuitBreaker.CanProceed() {
        return nil, fmt.Errorf("circuit breaker is open")
    }
    
    // Try to get available connection
    select {
    case conn := <-pool.available:
        if conn.isHealthy() {
            conn.acquire()
            atomic.AddInt32(&pool.activeConns, 1)
            
            // Update metrics
            acquisitionTime := time.Since(start)
            pool.updateAcquisitionTime(acquisitionTime)
            
            return conn, nil
        }
        
        // Connection unhealthy, destroy and create new one
        pool.destroyConnection(conn)
        
    case <-ctx.Done():
        return nil, ctx.Err()
        
    case <-time.After(pool.config.AcquisitionTimeout):
        return nil, fmt.Errorf("connection acquisition timeout")
    }
    
    // No available connections, try to create new one
    if atomic.LoadInt32(&pool.totalConns) < int32(pool.config.MaxConnections) {
        conn, err := pool.createConnection(pool.getDSN())
        if err != nil {
            atomic.AddInt64(&pool.metrics.ConnectionErrors, 1)
            if pool.circuitBreaker != nil {
                pool.circuitBreaker.RecordFailure()
            }
            return nil, err
        }
        
        conn.acquire()
        pool.connections = append(pool.connections, conn)
        atomic.AddInt32(&pool.activeConns, 1)
        
        return conn, nil
    }
    
    return nil, fmt.Errorf("connection pool exhausted")
}

func (pool *OptimizedConnectionPool) ReleaseConnection(conn *PooledConnection) {
    if conn == nil {
        return
    }
    
    conn.release()
    atomic.AddInt32(&pool.activeConns, -1)
    
    // Check if connection should be recycled
    if pool.shouldRecycleConnection(conn) {
        pool.destroyConnection(conn)
        
        // Create replacement connection asynchronously
        go func() {
            if newConn, err := pool.createConnection(pool.getDSN()); err == nil {
                select {
                case pool.available <- newConn:
                    pool.connections = append(pool.connections, newConn)
                default:
                    // Channel full, destroy the connection
                    pool.destroyConnection(newConn)
                }
            }
        }()
        
        return
    }
    
    // Return to available pool
    select {
    case pool.available <- conn:
        // Successfully returned to pool
    default:
        // Pool full, destroy connection
        pool.destroyConnection(conn)
    }
}

func (conn *PooledConnection) acquire() {
    conn.mu.Lock()
    defer conn.mu.Unlock()
    
    conn.inUse = true
    conn.lastUsed = time.Now()
}

func (conn *PooledConnection) release() {
    conn.mu.Lock()
    defer conn.mu.Unlock()
    
    conn.inUse = false
    conn.lastUsed = time.Now()
}

func (conn *PooledConnection) isHealthy() bool {
    conn.mu.Lock()
    defer conn.mu.Unlock()
    
    if !conn.healthy {
        return false
    }
    
    // Check if connection is too old
    if time.Since(conn.created) > time.Hour {
        return false
    }
    
    // Check if connection has been idle too long
    if time.Since(conn.lastUsed) > 30*time.Minute {
        return false
    }
    
    return true
}

func (pool *OptimizedConnectionPool) shouldRecycleConnection(conn *PooledConnection) bool {
    conn.mu.Lock()
    defer conn.mu.Unlock()
    
    // Recycle if too many errors
    if conn.errorCount > 10 {
        return true
    }
    
    // Recycle if average latency is too high
    if conn.queryCount > 0 {
        avgLatency := conn.totalLatency / time.Duration(conn.queryCount)
        if avgLatency > 100*time.Millisecond {
            return true
        }
    }
    
    // Recycle if connection is too old
    if time.Since(conn.created) > pool.config.LifetimeTimeout {
        return true
    }
    
    return false
}

func (pool *OptimizedConnectionPool) destroyConnection(conn *PooledConnection) {
    if conn == nil {
        return
    }
    
    conn.mu.Lock()
    defer conn.mu.Unlock()
    
    if conn.conn != nil {
        conn.conn.Close()
    }
    
    if conn.db != nil {
        conn.db.Close()
    }
    
    atomic.AddInt32(&pool.totalConns, -1)
    atomic.AddInt64(&pool.metrics.ConnectionsDestroyed, 1)
    
    // Remove from connections slice
    pool.mu.Lock()
    for i, c := range pool.connections {
        if c.id == conn.id {
            pool.connections = append(pool.connections[:i], pool.connections[i+1:]...)
            break
        }
    }
    pool.mu.Unlock()
}

// Health checker implementation
func (hc *HealthChecker) start() {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()
    
    hc.mu.Lock()
    hc.running = true
    hc.mu.Unlock()
    
    for {
        select {
        case <-ticker.C:
            hc.checkConnections()
            
        case <-hc.pool.shutdown:
            hc.mu.Lock()
            hc.running = false
            hc.mu.Unlock()
            return
        }
    }
}

func (hc *HealthChecker) checkConnections() {
    hc.pool.mu.RLock()
    connections := make([]*PooledConnection, len(hc.pool.connections))
    copy(connections, hc.pool.connections)
    hc.pool.mu.RUnlock()
    
    for _, conn := range connections {
        if conn.inUse {
            continue // Skip connections in use
        }
        
        healthy := hc.testConnection(conn)
        
        conn.mu.Lock()
        conn.healthy = healthy
        conn.mu.Unlock()
        
        if !healthy {
            // Remove unhealthy connection from available pool
            select {
            case <-hc.pool.available:
                // Connection removed from available pool
            default:
                // Connection not in available pool
            }
            
            // Schedule for replacement
            go func(c *PooledConnection) {
                hc.pool.destroyConnection(c)
                
                // Create replacement
                if newConn, err := hc.pool.createConnection(hc.pool.getDSN()); err == nil {
                    select {
                    case hc.pool.available <- newConn:
                        hc.pool.connections = append(hc.pool.connections, newConn)
                    default:
                        hc.pool.destroyConnection(newConn)
                    }
                }
            }(conn)
        }
    }
}

func (hc *HealthChecker) testConnection(conn *PooledConnection) bool {
    if conn.conn == nil {
        return false
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
    defer cancel()
    
    _, err := conn.conn.ExecContext(ctx, hc.testQuery)
    return err == nil
}

// Auto-scaling implementation
func (pool *OptimizedConnectionPool) startAutoScaling() {
    pool.scalingTicker = time.NewTicker(pool.config.ScalingInterval)
    defer pool.scalingTicker.Stop()
    
    for {
        select {
        case <-pool.scalingTicker.C:
            pool.autoScale()
            
        case <-pool.shutdown:
            return
        }
    }
}

func (pool *OptimizedConnectionPool) autoScale() {
    currentTotal := atomic.LoadInt32(&pool.totalConns)
    currentActive := atomic.LoadInt32(&pool.activeConns)
    
    utilization := float64(currentActive) / float64(currentTotal)
    
    // Scale up if utilization is high
    if utilization > 0.8 && currentTotal < int32(pool.config.MaxConnections) {
        target := min(currentTotal+2, int32(pool.config.MaxConnections))
        pool.scaleUp(int(target - currentTotal))
    }
    
    // Scale down if utilization is low
    if utilization < 0.3 && currentTotal > int32(pool.config.MinConnections) {
        target := max(currentTotal-1, int32(pool.config.MinConnections))
        pool.scaleDown(int(currentTotal - target))
    }
    
    // Update metrics
    atomic.StoreInt32(&pool.metrics.TotalConnections, currentTotal)
    atomic.StoreInt32(&pool.metrics.ActiveConnections, currentActive)
    atomic.StoreInt32(&pool.metrics.IdleConnections, currentTotal-currentActive)
    
    pool.metrics.ConnectionUtilization = utilization
}

func (pool *OptimizedConnectionPool) scaleUp(count int) {
    for i := 0; i < count; i++ {
        go func() {
            if conn, err := pool.createConnection(pool.getDSN()); err == nil {
                select {
                case pool.available <- conn:
                    pool.mu.Lock()
                    pool.connections = append(pool.connections, conn)
                    pool.mu.Unlock()
                default:
                    pool.destroyConnection(conn)
                }
            }
        }()
    }
}

func (pool *OptimizedConnectionPool) scaleDown(count int) {
    for i := 0; i < count; i++ {
        select {
        case conn := <-pool.available:
            if !conn.inUse {
                pool.destroyConnection(conn)
            } else {
                // Put back if in use
                pool.available <- conn
            }
        default:
            return // No available connections to remove
        }
    }
}

// Circuit breaker implementation
func (cb *CircuitBreaker) CanProceed() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case Closed:
        return true
    case Open:
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = HalfOpen
            return true
        }
        return false
    case HalfOpen:
        return true
    }
    
    return false
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures = 0
    cb.state = Closed
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures++
    cb.lastFailureTime = time.Now()
    
    if cb.failures >= int32(cb.maxFailures) {
        cb.state = Open
    }
}

// Load balancer implementation
func (lb *LoadBalancer) SelectConnection() *PooledConnection {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    if len(lb.connections) == 0 {
        return nil
    }
    
    switch lb.strategy {
    case LeastConnections:
        return lb.selectLeastConnections()
    case LatencyBased:
        return lb.selectLowestLatency()
    case AdaptiveLatency:
        return lb.selectAdaptiveLatency()
    default:
        return lb.selectRoundRobin()
    }
}

func (lb *LoadBalancer) selectLeastConnections() *PooledConnection {
    var selected *PooledConnection
    minQueries := int64(-1)
    
    for _, conn := range lb.connections {
        if !conn.inUse && conn.healthy {
            queries := atomic.LoadInt64(&conn.queryCount)
            if minQueries == -1 || queries < minQueries {
                minQueries = queries
                selected = conn
            }
        }
    }
    
    return selected
}

func (lb *LoadBalancer) selectLowestLatency() *PooledConnection {
    var selected *PooledConnection
    lowestLatency := time.Duration(-1)
    
    for _, conn := range lb.connections {
        if !conn.inUse && conn.healthy {
            conn.mu.Lock()
            var avgLatency time.Duration
            if conn.queryCount > 0 {
                avgLatency = conn.totalLatency / time.Duration(conn.queryCount)
            }
            conn.mu.Unlock()
            
            if lowestLatency == -1 || avgLatency < lowestLatency {
                lowestLatency = avgLatency
                selected = conn
            }
        }
    }
    
    return selected
}

func (lb *LoadBalancer) selectAdaptiveLatency() *PooledConnection {
    // Combine latency and connection count with adaptive weights
    var selected *PooledConnection
    bestScore := float64(-1)
    
    for _, conn := range lb.connections {
        if !conn.inUse && conn.healthy {
            conn.mu.Lock()
            queries := conn.queryCount
            var avgLatency time.Duration
            if queries > 0 {
                avgLatency = conn.totalLatency / time.Duration(queries)
            }
            conn.mu.Unlock()
            
            // Calculate score (lower is better)
            latencyScore := float64(avgLatency.Nanoseconds()) / 1e6 // Convert to milliseconds
            loadScore := float64(queries) / 1000                   // Normalize query count
            
            combinedScore := latencyScore*0.7 + loadScore*0.3
            
            if bestScore == -1 || combinedScore < bestScore {
                bestScore = combinedScore
                selected = conn
            }
        }
    }
    
    return selected
}

func (lb *LoadBalancer) selectRoundRobin() *PooledConnection {
    // Simple round-robin selection
    for _, conn := range lb.connections {
        if !conn.inUse && conn.healthy {
            return conn
        }
    }
    return nil
}

// Metrics collection
func (pool *OptimizedConnectionPool) startMetricsCollection() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            pool.collectMetrics()
            
        case <-pool.shutdown:
            return
        }
    }
}

func (pool *OptimizedConnectionPool) collectMetrics() {
    var totalQueries int64
    var totalLatency time.Duration
    
    pool.mu.RLock()
    for _, conn := range pool.connections {
        conn.mu.Lock()
        totalQueries += conn.queryCount
        totalLatency += conn.totalLatency
        conn.mu.Unlock()
    }
    pool.mu.RUnlock()
    
    if totalQueries > 0 {
        avgLatency := totalLatency / time.Duration(totalQueries)
        atomic.StoreInt64((*int64)(&pool.metrics.QueryLatency), int64(avgLatency))
    }
    
    // Calculate QPS (rough estimate)
    qps := float64(totalQueries) / time.Since(time.Now().Add(-time.Second)).Seconds()
    pool.metrics.ThroughputQPS = qps
}

// Utility functions
func (pool *OptimizedConnectionPool) updateAcquisitionTime(duration time.Duration) {
    // Exponential moving average
    currentAvg := atomic.LoadInt64((*int64)(&pool.metrics.AcquisitionTime))
    alpha := 0.1
    newAvg := time.Duration(float64(currentAvg)*(1-alpha) + float64(duration)*alpha)
    atomic.StoreInt64((*int64)(&pool.metrics.AcquisitionTime), int64(newAvg))
}

func (pool *OptimizedConnectionPool) getDSN() string {
    // Return the data source name - implementation specific
    return "postgres://user:password@localhost/db?sslmode=disable"
}

func (pool *OptimizedConnectionPool) GetMetrics() PoolMetrics {
    return PoolMetrics{
        TotalConnections:     atomic.LoadInt32(&pool.metrics.TotalConnections),
        ActiveConnections:    atomic.LoadInt32(&pool.metrics.ActiveConnections),
        IdleConnections:      atomic.LoadInt32(&pool.metrics.IdleConnections),
        ConnectionsCreated:   atomic.LoadInt64(&pool.metrics.ConnectionsCreated),
        ConnectionsDestroyed: atomic.LoadInt64(&pool.metrics.ConnectionsDestroyed),
        ConnectionErrors:     atomic.LoadInt64(&pool.metrics.ConnectionErrors),
        AcquisitionTime:      time.Duration(atomic.LoadInt64((*int64)(&pool.metrics.AcquisitionTime))),
        QueryLatency:         time.Duration(atomic.LoadInt64((*int64)(&pool.metrics.QueryLatency))),
        ThroughputQPS:        pool.metrics.ThroughputQPS,
        ConnectionUtilization: pool.metrics.ConnectionUtilization,
    }
}

func (pool *OptimizedConnectionPool) Close() error {
    close(pool.shutdown)
    
    // Close all connections
    pool.mu.Lock()
    for _, conn := range pool.connections {
        pool.destroyConnection(conn)
    }
    pool.connections = pool.connections[:0]
    pool.mu.Unlock()
    
    return nil
}

func min(a, b int32) int32 {
    if a < b {
        return a
    }
    return b
}

func max(a, b int32) int32 {
    if a > b {
        return a
    }
    return b
}
```

## Query Optimization

### Intelligent Query Execution and Caching
Implement advanced query optimization with prepared statements, caching, and intelligent execution planning:

```go
// Advanced query optimizer with caching and execution planning
type QueryOptimizer struct {
    cache           *QueryCache
    preparedStmts   *PreparedStatementManager
    analyzer        *QueryAnalyzer
    executor        *QueryExecutor
    metrics         QueryMetrics
    config          QueryConfig
}

type QueryConfig struct {
    CacheSize           int           `json:"cache_size"`
    CacheTTL            time.Duration `json:"cache_ttl"`
    PreparedStmtCache   int           `json:"prepared_stmt_cache"`
    QueryTimeout        time.Duration `json:"query_timeout"`
    SlowQueryThreshold  time.Duration `json:"slow_query_threshold"`
    EnableQueryRewrite  bool          `json:"enable_query_rewrite"`
    EnableResultCache   bool          `json:"enable_result_cache"`
    MaxRetries          int           `json:"max_retries"`
    RetryDelay          time.Duration `json:"retry_delay"`
}

type QueryMetrics struct {
    QueriesExecuted     int64         `json:"queries_executed"`
    SlowQueries         int64         `json:"slow_queries"`
    CacheHits           int64         `json:"cache_hits"`
    CacheMisses         int64         `json:"cache_misses"`
    PreparedStmtHits    int64         `json:"prepared_stmt_hits"`
    QueryErrors         int64         `json:"query_errors"`
    AvgExecutionTime    time.Duration `json:"avg_execution_time"`
    TotalExecutionTime  time.Duration `json:"total_execution_time"`
    ThroughputQPS       float64       `json:"throughput_qps"`
}

type QueryCache struct {
    results     map[string]*CachedResult
    access      map[string]time.Time
    maxSize     int
    currentSize int
    ttl         time.Duration
    mu          sync.RWMutex
}

type CachedResult struct {
    data      interface{}
    timestamp time.Time
    size      int
    hits      int64
}

type PreparedStatementManager struct {
    statements map[string]*PreparedStatement
    cache      *lru.Cache
    mu         sync.RWMutex
}

type PreparedStatement struct {
    stmt        *sql.Stmt
    query       string
    created     time.Time
    lastUsed    time.Time
    useCount    int64
    parameters  []interface{}
}

type QueryAnalyzer struct {
    patterns    map[string]*QueryPattern
    optimizer   *QueryOptimizer
    rewriter    *QueryRewriter
}

type QueryPattern struct {
    pattern     string
    frequency   int64
    avgTime     time.Duration
    complexity  QueryComplexity
    parameters  []ParameterInfo
}

type QueryComplexity int

const (
    Simple QueryComplexity = iota
    Medium
    Complex
    VeryComplex
)

type ParameterInfo struct {
    name     string
    dataType string
    nullable bool
    index    int
}

type QueryRewriter struct {
    rules       []RewriteRule
    cache       map[string]string
    mu          sync.RWMutex
}

type RewriteRule struct {
    pattern     string
    replacement string
    conditions  []RewriteCondition
}

type RewriteCondition struct {
    field    string
    operator string
    value    interface{}
}

type QueryExecutor struct {
    pool        *OptimizedConnectionPool
    batcher     *QueryBatcher
    monitor     *QueryMonitor
    retryPolicy *RetryPolicy
}

type QueryBatcher struct {
    batches     map[string]*QueryBatch
    maxBatchSize int
    flushInterval time.Duration
    mu          sync.RWMutex
}

type QueryBatch struct {
    queries     []BatchedQuery
    parameters  [][]interface{}
    callbacks   []func(interface{}, error)
    deadline    time.Time
}

type BatchedQuery struct {
    sql        string
    params     []interface{}
    resultType interface{}
    priority   QueryPriority
}

type QueryPriority int

const (
    Low QueryPriority = iota
    Normal
    High
    Critical
)

type QueryMonitor struct {
    slowQueries []SlowQuery
    maxHistory  int
    mu          sync.RWMutex
}

type SlowQuery struct {
    sql          string
    duration     time.Duration
    timestamp    time.Time
    parameters   []interface{}
    stackTrace   []string
}

func NewQueryOptimizer(config QueryConfig, pool *OptimizedConnectionPool) *QueryOptimizer {
    qo := &QueryOptimizer{
        config: config,
        cache: &QueryCache{
            results:     make(map[string]*CachedResult),
            access:      make(map[string]time.Time),
            maxSize:     config.CacheSize,
            ttl:         config.CacheTTL,
        },
        preparedStmts: &PreparedStatementManager{
            statements: make(map[string]*PreparedStatement),
        },
        analyzer: &QueryAnalyzer{
            patterns: make(map[string]*QueryPattern),
        },
    }
    
    qo.executor = &QueryExecutor{
        pool: pool,
        batcher: &QueryBatcher{
            batches:       make(map[string]*QueryBatch),
            maxBatchSize:  100,
            flushInterval: 10 * time.Millisecond,
        },
        monitor: &QueryMonitor{
            maxHistory: 1000,
        },
    }
    
    if config.EnableQueryRewrite {
        qo.analyzer.rewriter = &QueryRewriter{
            cache: make(map[string]string),
        }
        qo.initializeRewriteRules()
    }
    
    // Start background tasks
    go qo.startCacheEviction()
    go qo.startBatchFlushing()
    go qo.startMetricsCollection()
    
    return qo
}

func (qo *QueryOptimizer) ExecuteQuery(ctx context.Context, sql string, params ...interface{}) (interface{}, error) {
    start := time.Now()
    
    // Generate cache key
    cacheKey := qo.generateCacheKey(sql, params...)
    
    // Check cache first
    if qo.config.EnableResultCache {
        if result := qo.cache.Get(cacheKey); result != nil {
            atomic.AddInt64(&qo.metrics.CacheHits, 1)
            return result, nil
        }
        atomic.AddInt64(&qo.metrics.CacheMisses, 1)
    }
    
    // Analyze query
    pattern := qo.analyzer.AnalyzeQuery(sql)
    
    // Rewrite query if enabled
    optimizedSQL := sql
    if qo.config.EnableQueryRewrite && qo.analyzer.rewriter != nil {
        optimizedSQL = qo.analyzer.rewriter.RewriteQuery(sql)
    }
    
    // Execute query
    result, err := qo.executeWithRetry(ctx, optimizedSQL, params...)
    
    duration := time.Since(start)
    
    // Update metrics
    atomic.AddInt64(&qo.metrics.QueriesExecuted, 1)
    qo.updateExecutionTime(duration)
    
    if duration > qo.config.SlowQueryThreshold {
        atomic.AddInt64(&qo.metrics.SlowQueries, 1)
        qo.recordSlowQuery(sql, duration, params...)
    }
    
    if err != nil {
        atomic.AddInt64(&qo.metrics.QueryErrors, 1)
        return nil, err
    }
    
    // Cache result if enabled
    if qo.config.EnableResultCache && result != nil {
        qo.cache.Put(cacheKey, result)
    }
    
    // Update query pattern
    pattern.frequency++
    pattern.avgTime = (pattern.avgTime + duration) / 2
    
    return result, nil
}

func (qo *QueryOptimizer) executeWithRetry(ctx context.Context, sql string, params ...interface{}) (interface{}, error) {
    var lastErr error
    
    for attempt := 0; attempt <= qo.config.MaxRetries; attempt++ {
        if attempt > 0 {
            // Wait before retry
            select {
            case <-time.After(qo.config.RetryDelay):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
        
        result, err := qo.executeSingle(ctx, sql, params...)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !qo.isRetryableError(err) {
            break
        }
    }
    
    return nil, lastErr
}

func (qo *QueryOptimizer) executeSingle(ctx context.Context, sql string, params ...interface{}) (interface{}, error) {
    // Get connection from pool
    conn, err := qo.executor.pool.AcquireConnection(ctx)
    if err != nil {
        return nil, err
    }
    defer qo.executor.pool.ReleaseConnection(conn)
    
    // Check for prepared statement
    stmt, err := qo.getOrCreatePreparedStatement(conn, sql)
    if err != nil {
        // Fallback to direct execution
        return qo.executeDirectly(ctx, conn, sql, params...)
    }
    
    atomic.AddInt64(&qo.metrics.PreparedStmtHits, 1)
    
    // Execute prepared statement
    queryCtx, cancel := context.WithTimeout(ctx, qo.config.QueryTimeout)
    defer cancel()
    
    rows, err := stmt.stmt.QueryContext(queryCtx, params...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    return qo.scanResults(rows)
}

func (qo *QueryOptimizer) executeDirectly(ctx context.Context, conn *PooledConnection, sql string, params ...interface{}) (interface{}, error) {
    queryCtx, cancel := context.WithTimeout(ctx, qo.config.QueryTimeout)
    defer cancel()
    
    rows, err := conn.conn.QueryContext(queryCtx, sql, params...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    return qo.scanResults(rows)
}

func (qo *QueryOptimizer) getOrCreatePreparedStatement(conn *PooledConnection, sql string) (*PreparedStatement, error) {
    qo.preparedStmts.mu.RLock()
    existing, exists := qo.preparedStmts.statements[sql]
    qo.preparedStmts.mu.RUnlock()
    
    if exists {
        existing.lastUsed = time.Now()
        atomic.AddInt64(&existing.useCount, 1)
        return existing, nil
    }
    
    // Create new prepared statement
    stmt, err := conn.conn.PrepareContext(context.Background(), sql)
    if err != nil {
        return nil, err
    }
    
    prepared := &PreparedStatement{
        stmt:     stmt,
        query:    sql,
        created:  time.Now(),
        lastUsed: time.Now(),
        useCount: 1,
    }
    
    qo.preparedStmts.mu.Lock()
    qo.preparedStmts.statements[sql] = prepared
    qo.preparedStmts.mu.Unlock()
    
    return prepared, nil
}

func (qo *QueryOptimizer) scanResults(rows *sql.Rows) (interface{}, error) {
    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }
    
    var results []map[string]interface{}
    
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        
        results = append(results, row)
    }
    
    return results, rows.Err()
}

// Query analysis implementation
func (qa *QueryAnalyzer) AnalyzeQuery(sql string) *QueryPattern {
    // Normalize query for pattern matching
    normalized := qa.normalizeQuery(sql)
    
    qa.optimizer.analyzer.mu.RLock()
    pattern, exists := qa.patterns[normalized]
    qa.optimizer.analyzer.mu.RUnlock()
    
    if !exists {
        pattern = &QueryPattern{
            pattern:    normalized,
            frequency:  0,
            complexity: qa.assessComplexity(sql),
            parameters: qa.extractParameters(sql),
        }
        
        qa.optimizer.analyzer.mu.Lock()
        qa.patterns[normalized] = pattern
        qa.optimizer.analyzer.mu.Unlock()
    }
    
    return pattern
}

func (qa *QueryAnalyzer) normalizeQuery(sql string) string {
    // Remove extra whitespace and normalize case
    sql = strings.ToUpper(strings.TrimSpace(sql))
    
    // Replace parameter placeholders with generic markers
    re := regexp.MustCompile(`\$\d+|\?`)
    sql = re.ReplaceAllString(sql, "?")
    
    // Remove literal values
    re = regexp.MustCompile(`'[^']*'|"[^"]*"|\d+`)
    sql = re.ReplaceAllString(sql, "?")
    
    return sql
}

func (qa *QueryAnalyzer) assessComplexity(sql string) QueryComplexity {
    sql = strings.ToUpper(sql)
    
    complexity := Simple
    
    // Check for joins
    if strings.Contains(sql, "JOIN") {
        complexity = Medium
    }
    
    // Check for subqueries
    if strings.Count(sql, "(") > 1 {
        complexity = Medium
    }
    
    // Check for complex operations
    complexOps := []string{"UNION", "INTERSECT", "EXCEPT", "WITH", "RECURSIVE"}
    for _, op := range complexOps {
        if strings.Contains(sql, op) {
            complexity = Complex
            break
        }
    }
    
    // Check for window functions
    if strings.Contains(sql, "OVER") {
        complexity = Complex
    }
    
    // Check for multiple tables and complex conditions
    if strings.Count(sql, "FROM") > 1 || strings.Count(sql, "WHERE") > 1 {
        if complexity < Complex {
            complexity = Complex
        } else {
            complexity = VeryComplex
        }
    }
    
    return complexity
}

func (qa *QueryAnalyzer) extractParameters(sql string) []ParameterInfo {
    var params []ParameterInfo
    
    // Extract parameter placeholders
    re := regexp.MustCompile(`\$(\d+)`)
    matches := re.FindAllStringSubmatch(sql, -1)
    
    for i, match := range matches {
        if len(match) > 1 {
            params = append(params, ParameterInfo{
                name:     fmt.Sprintf("param%d", i+1),
                dataType: "unknown",
                nullable: true,
                index:    i,
            })
        }
    }
    
    return params
}

// Query rewriting implementation
func (qr *QueryRewriter) RewriteQuery(sql string) string {
    qr.mu.RLock()
    cached, exists := qr.cache[sql]
    qr.mu.RUnlock()
    
    if exists {
        return cached
    }
    
    rewritten := sql
    
    // Apply rewrite rules
    for _, rule := range qr.rules {
        if qr.matchesRule(sql, rule) {
            rewritten = qr.applyRule(rewritten, rule)
        }
    }
    
    // Cache the result
    qr.mu.Lock()
    qr.cache[sql] = rewritten
    qr.mu.Unlock()
    
    return rewritten
}

func (qr *QueryRewriter) matchesRule(sql string, rule RewriteRule) bool {
    matched, _ := regexp.MatchString(rule.pattern, sql)
    return matched
}

func (qr *QueryRewriter) applyRule(sql string, rule RewriteRule) string {
    re := regexp.MustCompile(rule.pattern)
    return re.ReplaceAllString(sql, rule.replacement)
}

func (qo *QueryOptimizer) initializeRewriteRules() {
    rules := []RewriteRule{
        {
            pattern:     `SELECT \* FROM`,
            replacement: `SELECT column_list FROM`,
        },
        {
            pattern:     `WHERE.*IN\s*\(SELECT`,
            replacement: `WHERE EXISTS (SELECT 1 FROM`,
        },
        {
            pattern:     `ORDER BY.*LIMIT 1`,
            replacement: `ORDER BY $1 FETCH FIRST 1 ROW ONLY`,
        },
    }
    
    qo.analyzer.rewriter.rules = rules
}

// Cache implementation
func (qc *QueryCache) Get(key string) interface{} {
    qc.mu.RLock()
    defer qc.mu.RUnlock()
    
    result, exists := qc.results[key]
    if !exists {
        return nil
    }
    
    // Check if expired
    if time.Since(result.timestamp) > qc.ttl {
        return nil
    }
    
    // Update access time
    qc.access[key] = time.Now()
    atomic.AddInt64(&result.hits, 1)
    
    return result.data
}

func (qc *QueryCache) Put(key string, data interface{}) {
    qc.mu.Lock()
    defer qc.mu.Unlock()
    
    size := qc.estimateSize(data)
    
    // Evict if necessary
    for qc.currentSize+size > qc.maxSize && len(qc.results) > 0 {
        qc.evictLRU()
    }
    
    qc.results[key] = &CachedResult{
        data:      data,
        timestamp: time.Now(),
        size:      size,
        hits:      0,
    }
    
    qc.access[key] = time.Now()
    qc.currentSize += size
}

func (qc *QueryCache) evictLRU() {
    var oldestKey string
    var oldestTime time.Time
    
    for key, accessTime := range qc.access {
        if oldestKey == "" || accessTime.Before(oldestTime) {
            oldestKey = key
            oldestTime = accessTime
        }
    }
    
    if oldestKey != "" {
        if result, exists := qc.results[oldestKey]; exists {
            qc.currentSize -= result.size
        }
        delete(qc.results, oldestKey)
        delete(qc.access, oldestKey)
    }
}

func (qc *QueryCache) estimateSize(data interface{}) int {
    // Simple size estimation
    return 1024 // Placeholder
}

// Background tasks
func (qo *QueryOptimizer) startCacheEviction() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        qo.cache.mu.Lock()
        
        // Remove expired entries
        now := time.Now()
        for key, result := range qo.cache.results {
            if now.Sub(result.timestamp) > qo.cache.ttl {
                qo.cache.currentSize -= result.size
                delete(qo.cache.results, key)
                delete(qo.cache.access, key)
            }
        }
        
        qo.cache.mu.Unlock()
    }
}

func (qo *QueryOptimizer) startBatchFlushing() {
    ticker := time.NewTicker(qo.executor.batcher.flushInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        qo.flushBatches()
    }
}

func (qo *QueryOptimizer) flushBatches() {
    qo.executor.batcher.mu.Lock()
    defer qo.executor.batcher.mu.Unlock()
    
    now := time.Now()
    
    for key, batch := range qo.executor.batcher.batches {
        if now.After(batch.deadline) || len(batch.queries) >= qo.executor.batcher.maxBatchSize {
            go qo.executeBatch(batch)
            delete(qo.executor.batcher.batches, key)
        }
    }
}

func (qo *QueryOptimizer) executeBatch(batch *QueryBatch) {
    // Execute batched queries
    for i, query := range batch.queries {
        result, err := qo.executeSingle(context.Background(), query.sql, query.params...)
        if len(batch.callbacks) > i && batch.callbacks[i] != nil {
            batch.callbacks[i](result, err)
        }
    }
}

func (qo *QueryOptimizer) startMetricsCollection() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        qo.updateThroughputMetrics()
    }
}

func (qo *QueryOptimizer) updateThroughputMetrics() {
    queries := atomic.LoadInt64(&qo.metrics.QueriesExecuted)
    
    // Calculate QPS (simple implementation)
    qps := float64(queries) / time.Since(time.Now().Add(-time.Second)).Seconds()
    qo.metrics.ThroughputQPS = qps
}

// Utility functions
func (qo *QueryOptimizer) generateCacheKey(sql string, params ...interface{}) string {
    key := sql
    for _, param := range params {
        key += fmt.Sprintf(":%v", param)
    }
    
    // Use hash for shorter keys
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:])
}

func (qo *QueryOptimizer) isRetryableError(err error) bool {
    // Check for temporary/retryable errors
    if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
        return true
    }
    
    // Database-specific retryable errors
    errStr := err.Error()
    retryablePatterns := []string{
        "connection reset",
        "connection refused",
        "timeout",
        "temporary",
    }
    
    for _, pattern := range retryablePatterns {
        if strings.Contains(strings.ToLower(errStr), pattern) {
            return true
        }
    }
    
    return false
}

func (qo *QueryOptimizer) recordSlowQuery(sql string, duration time.Duration, params ...interface{}) {
    slowQuery := SlowQuery{
        sql:        sql,
        duration:   duration,
        timestamp:  time.Now(),
        parameters: params,
        stackTrace: qo.getCaller(),
    }
    
    qo.executor.monitor.mu.Lock()
    qo.executor.monitor.slowQueries = append(qo.executor.monitor.slowQueries, slowQuery)
    
    // Keep only recent slow queries
    if len(qo.executor.monitor.slowQueries) > qo.executor.monitor.maxHistory {
        qo.executor.monitor.slowQueries = qo.executor.monitor.slowQueries[1:]
    }
    qo.executor.monitor.mu.Unlock()
}

func (qo *QueryOptimizer) getCaller() []string {
    // Get stack trace for debugging
    // Simplified implementation
    return []string{"caller information"}
}

func (qo *QueryOptimizer) updateExecutionTime(duration time.Duration) {
    // Update average execution time using exponential moving average
    currentAvg := atomic.LoadInt64((*int64)(&qo.metrics.AvgExecutionTime))
    alpha := 0.1
    newAvg := time.Duration(float64(currentAvg)*(1-alpha) + float64(duration)*alpha)
    atomic.StoreInt64((*int64)(&qo.metrics.AvgExecutionTime), int64(newAvg))
    
    // Update total execution time
    atomic.AddInt64((*int64)(&qo.metrics.TotalExecutionTime), int64(duration))
}

func (qo *QueryOptimizer) GetMetrics() QueryMetrics {
    return QueryMetrics{
        QueriesExecuted:    atomic.LoadInt64(&qo.metrics.QueriesExecuted),
        SlowQueries:        atomic.LoadInt64(&qo.metrics.SlowQueries),
        CacheHits:          atomic.LoadInt64(&qo.metrics.CacheHits),
        CacheMisses:        atomic.LoadInt64(&qo.metrics.CacheMisses),
        PreparedStmtHits:   atomic.LoadInt64(&qo.metrics.PreparedStmtHits),
        QueryErrors:        atomic.LoadInt64(&qo.metrics.QueryErrors),
        AvgExecutionTime:   time.Duration(atomic.LoadInt64((*int64)(&qo.metrics.AvgExecutionTime))),
        TotalExecutionTime: time.Duration(atomic.LoadInt64((*int64)(&qo.metrics.TotalExecutionTime))),
        ThroughputQPS:      qo.metrics.ThroughputQPS,
    }
}

// Required imports and placeholder implementations
import (
    "crypto/sha256"
    "encoding/hex"
    "net"
    "regexp"
    "strings"
)

// Placeholder for LRU cache
type lru struct {
    Cache interface{}
}
```

Database and storage optimization requires sophisticated connection management, intelligent query execution, and comprehensive caching strategies. By implementing advanced connection pooling, query optimization, and result caching, applications can achieve significant performance improvements while maintaining data consistency and system reliability.

## Key Takeaways

1. **Implement intelligent connection pooling** - use adaptive scaling and load balancing
2. **Optimize query execution** - leverage prepared statements and intelligent caching
3. **Apply sophisticated caching** - cache query results and execution plans
4. **Monitor performance metrics** - track query latency, throughput, and resource utilization
5. **Use circuit breakers** - protect against database failures and cascading issues
6. **Implement query analysis** - identify patterns and optimize frequently executed queries
7. **Apply intelligent retry policies** - handle temporary failures gracefully
8. **Leverage batch processing** - group similar operations for better efficiency

Effective database optimization enables applications to handle high query loads while maintaining low latency and optimal resource utilization across the entire data access layer.
