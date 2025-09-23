# Microservices Performance Case Study

This case study examines the optimization of a distributed microservices architecture handling millions of transactions per day, demonstrating how systematic performance engineering transformed a struggling system into a high-performance platform.

## System Overview

**Architecture**: 15 microservices handling financial transactions
**Scale**: 50M+ transactions/day, 2,000+ requests/second peak
**Technology Stack**: Go 1.21+, gRPC, PostgreSQL, Redis, Kubernetes
**Initial Problem**: System failing under 30% of target load

### Performance Crisis

**Symptoms Observed:**
- Services crashing under moderate load (500 RPS vs 2,000 RPS target)
- Database connection exhaustion
- Memory leaks causing container restarts every 2 hours
- Inter-service communication latency >1 second
- Cascade failures bringing down entire system

**Business Impact:**
- $2M+ revenue loss due to transaction failures
- Customer satisfaction drop to 65% (from 94%)
- Engineering team 80% occupied with firefighting
- Regulatory compliance violations due to audit trail gaps

## Initial Performance Assessment

### System-Wide Profiling

```bash
# Distributed tracing revealed bottlenecks across services
kubectl exec -it payment-service -- go tool pprof :6060/debug/pprof/profile

# Critical findings across the system:
# 1. Payment Service: 87% CPU in JSON processing
# 2. User Service: 92% memory growth rate (memory leak)
# 3. Transaction Service: 78% time waiting on database
# 4. Notification Service: 15,000+ goroutines (should be <100)
# 5. Gateway Service: 67% time in request routing
```

### Service-by-Service Analysis

**Payment Service:**
```go
// CPU profile showed JSON serialization bottleneck
Type: cpu
Duration: 30s
Showing nodes accounting for 26.1s, 87.0% of 30s total

flat  flat%   sum%        cum   cum%
8.7s  29.0%  29.0%      8.7s  29.0%  encoding/json.(*encodeState).marshal
5.2s  17.3%  46.3%      5.2s  17.3%  encoding/json.valueEncoder
3.1s  10.3%  56.6%      3.1s  10.3%  reflect.Value.Interface
2.8s   9.3%  65.9%      2.8s   9.3%  runtime.mapaccess2_faststr
2.3s   7.7%  73.6%      2.3s   7.7%  encoding/json.(*decodeState).object
```

**User Service Memory Leak:**
```bash
# Memory profile revealed goroutine leak in session management
go tool pprof :6060/debug/pprof/heap

Type: inuse_space
Showing nodes accounting for 2.1GB, 94.2% of 2.2GB total

flat  flat%   sum%        cum   cum%
892MB  40.5%  40.5%      892MB  40.5%  sessionManager.(*SessionStore).background
445MB  20.2%  60.7%      445MB  20.2%  http.(*persistConn).writeLoop
287MB  13.0%  73.7%      287MB  13.0%  encoding/json.Marshal
```

**Transaction Service Database Issues:**
```sql
-- Query analysis revealed N+1 problems and missing indexes
SELECT pg_stat_statements.calls, pg_stat_statements.total_time, 
       pg_stat_statements.mean_time, pg_stat_statements.query
FROM pg_stat_statements 
ORDER BY mean_time DESC;

-- Top problematic queries:
-- 1. Individual transaction lookups: 2,847ms avg (called 45,000x/hour)
-- 2. User balance calculations: 1,234ms avg (called 12,000x/hour)  
-- 3. Audit trail inserts: 856ms avg (called 89,000x/hour)
```

## Comprehensive Optimization Strategy

### 1. Payment Service - JSON Processing Optimization

**Problem**: Standard JSON processing consuming 87% CPU time.

**Solution**: Implemented high-performance protocol buffer communication with streaming:

```go
// Before: Slow JSON-based API
type PaymentRequest struct {
    UserID      string          `json:"user_id"`
    Amount      decimal.Decimal `json:"amount"`
    Currency    string          `json:"currency"`
    Description string          `json:"description"`
    Metadata    map[string]string `json:"metadata"`
}

func (s *PaymentService) ProcessPayment(w http.ResponseWriter, r *http.Request) {
    var req PaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    
    // Processing...
    
    response := PaymentResponse{
        TransactionID: generateID(),
        Status:        "completed",
        ProcessedAt:   time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// After: High-performance gRPC with Protocol Buffers
service PaymentService {
    rpc ProcessPayment(PaymentRequest) returns (PaymentResponse);
    rpc ProcessPaymentStream(stream PaymentRequest) returns (stream PaymentResponse);
}

message PaymentRequest {
    string user_id = 1;
    int64 amount_cents = 2;
    string currency = 3;
    string description = 4;
    map<string, string> metadata = 5;
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
    // Direct protobuf processing - no reflection overhead
    userID := req.GetUserId()
    amount := decimal.NewFromInt(req.GetAmountCents()).Div(decimal.NewFromInt(100))
    
    // Process payment...
    
    return &pb.PaymentResponse{
        TransactionId: s.generateID(),
        Status:        pb.PaymentStatus_COMPLETED,
        ProcessedAt:   timestamppb.Now(),
    }, nil
}

// Performance improvement: 15.2x faster serialization
// CPU usage: 87% → 5.7% for JSON processing
// Latency: 234ms → 15ms per request
```

### 2. User Service - Memory Leak Resolution

**Problem**: Session management goroutines accumulating indefinitely.

**Solution**: Implemented bounded session pool with proper lifecycle management:

```go
// Before: Leaking session goroutines
type SessionManager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

func (sm *SessionManager) CreateSession(userID string) *Session {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    session := &Session{
        ID:        generateSessionID(),
        UserID:    userID,
        CreatedAt: time.Now(),
    }
    
    sm.sessions[session.ID] = session
    
    // BUG: Goroutine never cleaned up!
    go func() {
        for {
            select {
            case <-time.After(1 * time.Minute):
                session.LastActivity = time.Now()
            }
        }
    }()
    
    return session
}

// After: Bounded session pool with lifecycle management
type OptimizedSessionManager struct {
    sessions    map[string]*Session
    cleanupChan chan string
    workerPool  *WorkerPool
    metrics     SessionMetrics
    mu          sync.RWMutex
}

type Session struct {
    ID           string
    UserID       string
    CreatedAt    time.Time
    LastActivity time.Time
    ExpiresAt    time.Time
    mu           sync.RWMutex
}

func NewOptimizedSessionManager(poolSize int) *OptimizedSessionManager {
    sm := &OptimizedSessionManager{
        sessions:    make(map[string]*Session),
        cleanupChan: make(chan string, 1000),
        workerPool:  NewWorkerPool(poolSize),
    }
    
    // Single cleanup goroutine instead of per-session goroutines
    go sm.cleanupWorker()
    
    // Periodic batch cleanup
    go sm.periodicCleanup()
    
    return sm
}

func (sm *OptimizedSessionManager) CreateSession(userID string) *Session {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    sessionID := sm.generateSessionID()
    expiresAt := time.Now().Add(24 * time.Hour)
    
    session := &Session{
        ID:           sessionID,
        UserID:       userID,
        CreatedAt:    time.Now(),
        LastActivity: time.Now(),
        ExpiresAt:    expiresAt,
    }
    
    sm.sessions[sessionID] = session
    
    // Schedule cleanup instead of creating goroutine
    sm.workerPool.ScheduleTask(WorkTask{
        ID:       sessionID,
        RunAt:    expiresAt,
        Handler:  sm.expireSession,
    })
    
    atomic.AddInt64(&sm.metrics.SessionsCreated, 1)
    return session
}

func (sm *OptimizedSessionManager) cleanupWorker() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case sessionID := <-sm.cleanupChan:
            sm.removeSession(sessionID)
            
        case <-ticker.C:
            sm.batchCleanup()
        }
    }
}

func (sm *OptimizedSessionManager) batchCleanup() {
    sm.mu.Lock()
    now := time.Now()
    expired := make([]string, 0, 100)
    
    for id, session := range sm.sessions {
        session.mu.RLock()
        isExpired := now.After(session.ExpiresAt)
        session.mu.RUnlock()
        
        if isExpired {
            expired = append(expired, id)
        }
        
        // Limit batch size
        if len(expired) >= 100 {
            break
        }
    }
    sm.mu.Unlock()
    
    // Remove expired sessions
    for _, id := range expired {
        sm.removeSession(id)
    }
    
    atomic.AddInt64(&sm.metrics.SessionsExpired, int64(len(expired)))
}

func (sm *OptimizedSessionManager) removeSession(sessionID string) {
    sm.mu.Lock()
    delete(sm.sessions, sessionID)
    sm.mu.Unlock()
    
    atomic.AddInt64(&sm.metrics.SessionsRemoved, 1)
}

func (sm *OptimizedSessionManager) expireSession(sessionID string) {
    select {
    case sm.cleanupChan <- sessionID:
    default:
        // Cleanup channel full, schedule for next batch
        go func() {
            time.Sleep(1 * time.Minute)
            sm.cleanupChan <- sessionID
        }()
    }
}

// Memory improvement: 97% reduction in goroutine count
// Memory usage: 2.1GB → 85MB steady state
// Container restart frequency: Every 2 hours → Never
```

### 3. Transaction Service - Database Optimization

**Problem**: Database queries averaging 2+ seconds with connection exhaustion.

**Solution**: Implemented connection pooling, query optimization, and caching:

```go
// Database optimization layer
type OptimizedTransactionService struct {
    db          *sql.DB
    cache       *redis.Client
    stmtCache   map[string]*sql.Stmt
    queryStats  map[string]*QueryStats
    mu          sync.RWMutex
}

type QueryStats struct {
    ExecutionCount int64
    TotalTime      time.Duration
    AverageTime    time.Duration
    ErrorCount     int64
}

func NewOptimizedTransactionService(dbConfig DatabaseConfig) (*OptimizedTransactionService, error) {
    db, err := sql.Open("postgres", dbConfig.DSN)
    if err != nil {
        return nil, err
    }
    
    // Optimized connection pool settings
    db.SetMaxOpenConns(100)        // Increased from 10
    db.SetMaxIdleConns(50)         // Increased from 2
    db.SetConnMaxLifetime(30 * time.Minute)
    db.SetConnMaxIdleTime(5 * time.Minute)
    
    // Redis for caching
    rdb := redis.NewClient(&redis.Options{
        Addr:         dbConfig.RedisAddr,
        PoolSize:     50,
        MinIdleConns: 10,
        MaxRetries:   3,
    })
    
    return &OptimizedTransactionService{
        db:         db,
        cache:      rdb,
        stmtCache:  make(map[string]*sql.Stmt),
        queryStats: make(map[string]*QueryStats),
    }, nil
}

// Optimized transaction lookup with caching
func (s *OptimizedTransactionService) GetTransaction(ctx context.Context, txID string) (*Transaction, error) {
    start := time.Now()
    
    // Check cache first
    cacheKey := fmt.Sprintf("tx:%s", txID)
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var tx Transaction
        if err := json.Unmarshal([]byte(cached), &tx); err == nil {
            s.recordQueryStats("GetTransaction", time.Since(start), false)
            return &tx, nil
        }
    }
    
    // Prepared statement for database query
    stmt, err := s.getOrCreateStmt("GetTransaction", `
        SELECT id, user_id, amount, currency, status, created_at, updated_at
        FROM transactions 
        WHERE id = $1 AND deleted_at IS NULL
    `)
    if err != nil {
        return nil, err
    }
    
    var tx Transaction
    err = stmt.QueryRowContext(ctx, txID).Scan(
        &tx.ID, &tx.UserID, &tx.Amount, &tx.Currency,
        &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
    )
    
    queryTime := time.Since(start)
    
    if err != nil {
        s.recordQueryStats("GetTransaction", queryTime, true)
        return nil, err
    }
    
    // Cache successful result
    if txData, err := json.Marshal(tx); err == nil {
        s.cache.Set(ctx, cacheKey, txData, 5*time.Minute)
    }
    
    s.recordQueryStats("GetTransaction", queryTime, false)
    return &tx, nil
}

// Batch transaction processing to reduce N+1 queries
func (s *OptimizedTransactionService) GetUserTransactions(ctx context.Context, userID string, limit int) ([]*Transaction, error) {
    cacheKey := fmt.Sprintf("user_txs:%s:%d", userID, limit)
    
    // Check cache
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var transactions []*Transaction
        if err := json.Unmarshal([]byte(cached), &transactions); err == nil {
            return transactions, nil
        }
    }
    
    // Optimized query with proper indexing
    stmt, err := s.getOrCreateStmt("GetUserTransactions", `
        SELECT id, user_id, amount, currency, status, created_at, updated_at
        FROM transactions 
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC 
        LIMIT $2
    `)
    if err != nil {
        return nil, err
    }
    
    rows, err := stmt.QueryContext(ctx, userID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var transactions []*Transaction
    for rows.Next() {
        var tx Transaction
        err := rows.Scan(
            &tx.ID, &tx.UserID, &tx.Amount, &tx.Currency,
            &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        transactions = append(transactions, &tx)
    }
    
    // Cache results
    if txData, err := json.Marshal(transactions); err == nil {
        s.cache.Set(ctx, cacheKey, txData, 2*time.Minute)
    }
    
    return transactions, nil
}

func (s *OptimizedTransactionService) getOrCreateStmt(name, query string) (*sql.Stmt, error) {
    s.mu.RLock()
    stmt, exists := s.stmtCache[name]
    s.mu.RUnlock()
    
    if exists {
        return stmt, nil
    }
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Double-check after acquiring write lock
    if stmt, exists := s.stmtCache[name]; exists {
        return stmt, nil
    }
    
    stmt, err := s.db.Prepare(query)
    if err != nil {
        return nil, err
    }
    
    s.stmtCache[name] = stmt
    return stmt, nil
}

func (s *OptimizedTransactionService) recordQueryStats(queryName string, duration time.Duration, isError bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    stats, exists := s.queryStats[queryName]
    if !exists {
        stats = &QueryStats{}
        s.queryStats[queryName] = stats
    }
    
    stats.ExecutionCount++
    stats.TotalTime += duration
    stats.AverageTime = stats.TotalTime / time.Duration(stats.ExecutionCount)
    
    if isError {
        stats.ErrorCount++
    }
}

// Database schema optimizations
const schema = `
-- Added compound indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_user_created 
ON transactions(user_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_status_created 
ON transactions(status, created_at DESC) WHERE deleted_at IS NULL;

-- Partitioning for large transaction tables
CREATE TABLE transactions_2024_q1 PARTITION OF transactions 
FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');

CREATE TABLE transactions_2024_q2 PARTITION OF transactions 
FOR VALUES FROM ('2024-04-01') TO ('2024-07-01');
`

// Performance improvements:
// - Query latency: 2,847ms → 12ms (237x improvement)
// - Connection usage: 100% → 35% utilization
// - Cache hit rate: 89% for frequent queries
// - Database CPU: 78% → 15% reduction
```

## Results and Impact

### Performance Metrics

**System-Wide Improvements:**

| Service | Metric | Before | After | Improvement |
|---------|--------|--------|-------|-------------|
| Payment | Latency (p95) | 234ms | 15ms | 93% reduction |
| Payment | CPU Usage | 87% | 12% | 86% reduction |
| User | Memory Usage | 2.1GB | 85MB | 96% reduction |
| User | Goroutine Count | 15,000 | 45 | 99.7% reduction |
| Transaction | Query Latency | 2.8s | 12ms | 99.6% reduction |
| Transaction | Throughput | 45 QPS | 1,200 QPS | 26.7x increase |
| Gateway | Request Routing | 67% CPU | 8% CPU | 88% reduction |

**System Capacity:**

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Peak RPS | 500 | 2,500 | 5x increase |
| Concurrent Users | 1,000 | 8,000 | 8x increase |
| Daily Transactions | 15M | 75M | 5x increase |
| System Availability | 95.2% | 99.97% | 4.77% points |
| MTTR | 45 minutes | 3 minutes | 93% reduction |

### Business Impact

**Financial Results:**
- **Revenue Recovery**: $2.1M monthly revenue loss eliminated
- **Infrastructure Savings**: 60% reduction in cloud costs ($180K/month)
- **Operational Efficiency**: 80% reduction in support tickets
- **Development Velocity**: 3x faster feature delivery

**Customer Satisfaction:**
- **User Experience**: 65% → 96% satisfaction score
- **Transaction Success Rate**: 85% → 99.8%
- **API Response Time**: 2.3s → 47ms average
- **Customer Retention**: 15% improvement

### Implementation Timeline

**Phase 1 (Weeks 1-3): Critical Stabilization**
- Payment service protobuf migration
- User service memory leak fix  
- Database connection pool optimization
- Emergency capacity scaling

**Phase 2 (Weeks 4-6): Performance Optimization**
- Query optimization and indexing
- Caching layer implementation
- Goroutine pool optimization
- Load balancing improvements

**Phase 3 (Weeks 7-9): Monitoring and Validation**
- Comprehensive monitoring setup
- Load testing and capacity planning
- Performance regression testing
- Production validation

**Phase 4 (Weeks 10-12): Long-term Optimization**
- Advanced caching strategies
- Database partitioning
- Auto-scaling implementation
- Performance culture establishment

### Monitoring and Observability

**Performance Dashboard:**
```go
// Real-time performance monitoring
type PerformanceMonitor struct {
    metrics map[string]*ServiceMetrics
    alerts  *AlertManager
    mu      sync.RWMutex
}

type ServiceMetrics struct {
    RequestCount    int64
    ErrorCount      int64
    LatencyP50      time.Duration
    LatencyP95      time.Duration
    LatencyP99      time.Duration
    MemoryUsage     int64
    CPUUsage        float64
    GoroutineCount  int
    LastUpdated     time.Time
}

func (pm *PerformanceMonitor) RecordMetrics(service string, latency time.Duration, isError bool) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    metrics, exists := pm.metrics[service]
    if !exists {
        metrics = &ServiceMetrics{}
        pm.metrics[service] = metrics
    }
    
    atomic.AddInt64(&metrics.RequestCount, 1)
    if isError {
        atomic.AddInt64(&metrics.ErrorCount, 1)
    }
    
    // Update latency percentiles (simplified)
    metrics.updateLatencyPercentiles(latency)
    metrics.LastUpdated = time.Now()
    
    // Check alerts
    pm.alerts.CheckThresholds(service, metrics)
}

// Performance SLA monitoring
const performanceSLA = `
Service Level Objectives:
- API Latency P95: <100ms
- Error Rate: <0.1%
- Availability: >99.9%
- Memory Growth: <1% per hour
- CPU Usage: <80% average
`
```

**Automated Performance Testing:**
```bash
#!/bin/bash
# Continuous performance validation pipeline

# Run load tests
k6 run --vus 1000 --duration 10m performance-tests/api-load-test.js

# Validate SLA compliance
if [ $P95_LATENCY -gt 100 ]; then
    echo "SLA VIOLATION: P95 latency ${P95_LATENCY}ms exceeds 100ms"
    exit 1
fi

# Memory leak detection
kubectl exec payment-service -- go tool pprof -top heap > memory-profile.txt
if grep -q "growing" memory-profile.txt; then
    echo "MEMORY LEAK DETECTED"
    exit 1
fi

echo "Performance validation passed"
```

This comprehensive microservices optimization case study demonstrates how systematic performance engineering can transform a failing distributed system into a high-performance platform capable of handling enterprise-scale loads while maintaining reliability and cost efficiency.
