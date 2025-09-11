# Production Performance

Master production-ready performance engineering, monitoring, testing, and scalability patterns for Go applications in real-world deployments.

## Production Performance Overview

Production performance engineering involves:

- **Continuous profiling** - Real-time performance monitoring
- **Performance testing** - Systematic validation of performance characteristics
- **Scalability planning** - Designing for growth and load variations
- **Incident response** - Rapid diagnosis and resolution of performance issues

## Performance in Production Context

### Production vs Development Performance

```go
// Production configuration example
type ProductionConfig struct {
    // Runtime configuration
    GOMAXPROCS       int    `env:"GOMAXPROCS" default:"0"`        // Use all CPUs
    GOMEMLIMIT       string `env:"GOMEMLIMIT" default:""`         // Memory limit
    GOGC             int    `env:"GOGC" default:"100"`            // GC target percentage
    
    // Application configuration
    PoolSize         int    `env:"POOL_SIZE" default:"100"`       // Connection pool size
    CacheSize        int    `env:"CACHE_SIZE" default:"10000"`    // Cache entries
    RequestTimeout   time.Duration `env:"REQUEST_TIMEOUT" default:"30s"`
    
    // Monitoring configuration
    ProfilingEnabled bool   `env:"PROFILING_ENABLED" default:"true"`
    MetricsAddr      string `env:"METRICS_ADDR" default:":8080"`
    LogLevel         string `env:"LOG_LEVEL" default:"info"`
}

func NewProductionConfig() (*ProductionConfig, error) {
    config := &ProductionConfig{}
    
    // Load from environment
    if err := env.Parse(config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    // Apply production optimizations
    if config.GOMAXPROCS > 0 {
        runtime.GOMAXPROCS(config.GOMAXPROCS)
    }
    
    if config.GOMEMLIMIT != "" {
        if limit, err := parseMemoryLimit(config.GOMEMLIMIT); err == nil {
            debug.SetMemoryLimit(limit)
        }
    }
    
    debug.SetGCPercent(config.GOGC)
    
    return config, nil
}
```

### Production Deployment Patterns

```go
// Graceful shutdown pattern
type Server struct {
    httpServer *http.Server
    cleanup    []func() error
    logger     *log.Logger
}

func (s *Server) Start() error {
    // Start HTTP server
    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.logger.Printf("HTTP server error: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    <-sigChan
    s.logger.Println("Shutting down server...")
    
    return s.Shutdown()
}

func (s *Server) Shutdown() error {
    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown HTTP server gracefully
    if err := s.httpServer.Shutdown(ctx); err != nil {
        s.logger.Printf("HTTP server shutdown error: %v", err)
    }
    
    // Run cleanup functions
    for _, cleanup := range s.cleanup {
        if err := cleanup(); err != nil {
            s.logger.Printf("Cleanup error: %v", err)
        }
    }
    
    return nil
}

// Health check endpoints
func (s *Server) setupHealthChecks() {
    http.HandleFunc("/health", s.healthCheck)
    http.HandleFunc("/ready", s.readinessCheck)
    http.HandleFunc("/metrics", s.metricsHandler)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
    // Basic liveness check
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (s *Server) readinessCheck(w http.ResponseWriter, r *http.Request) {
    // Check dependencies (database, cache, etc.)
    if !s.isDatabaseReady() || !s.isCacheReady() {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Service Unavailable"))
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
}
```

## Performance Monitoring in Production

### Built-in Profiling Integration

```go
import (
    _ "net/http/pprof" // Enable pprof endpoints
)

func setupProfiling() {
    // Production-safe profiling setup
    mux := http.NewServeMux()
    
    // Restrict pprof access in production
    if os.Getenv("ENABLE_PPROF") == "true" {
        mux.Handle("/debug/pprof/", http.DefaultServeMux)
    }
    
    // Custom metrics endpoint
    mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        metrics := map[string]interface{}{
            "goroutines":     runtime.NumGoroutine(),
            "memory_alloc":   m.Alloc,
            "memory_sys":     m.Sys,
            "gc_runs":        m.NumGC,
            "heap_objects":   m.HeapObjects,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics)
    })
    
    // Start metrics server on separate port
    go func() {
        log.Printf("Metrics server starting on :6060")
        log.Fatal(http.ListenAndServe(":6060", mux))
    }()
}
```

### Custom Metrics Collection

```go
// Production metrics system
type MetricsCollector struct {
    requestDuration *prometheus.HistogramVec
    requestCount    *prometheus.CounterVec
    activeRequests  prometheus.Gauge
    errorCount      *prometheus.CounterVec
    goroutineCount  prometheus.Gauge
    memoryUsage     prometheus.Gauge
}

func NewMetricsCollector() *MetricsCollector {
    mc := &MetricsCollector{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_request_duration_seconds",
                Help:    "HTTP request duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method", "endpoint", "status"},
        ),
        
        requestCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "endpoint", "status"},
        ),
        
        activeRequests: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "http_active_requests",
                Help: "Number of active HTTP requests",
            },
        ),
        
        errorCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_errors_total",
                Help: "Total number of HTTP errors",
            },
            []string{"method", "endpoint", "error_type"},
        ),
        
        goroutineCount: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "go_goroutines",
                Help: "Number of goroutines",
            },
        ),
        
        memoryUsage: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "go_memory_usage_bytes",
                Help: "Memory usage in bytes",
            },
        ),
    }
    
    // Register metrics
    prometheus.MustRegister(
        mc.requestDuration,
        mc.requestCount,
        mc.activeRequests,
        mc.errorCount,
        mc.goroutineCount,
        mc.memoryUsage,
    )
    
    // Start background metrics collection
    go mc.collectRuntimeMetrics()
    
    return mc
}

func (mc *MetricsCollector) collectRuntimeMetrics() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        mc.goroutineCount.Set(float64(runtime.NumGoroutine()))
        mc.memoryUsage.Set(float64(m.Alloc))
    }
}

// Middleware for request metrics
func (mc *MetricsCollector) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        mc.activeRequests.Inc()
        defer mc.activeRequests.Dec()
        
        // Wrap response writer to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        next.ServeHTTP(ww, r)
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(ww.statusCode)
        
        mc.requestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
        mc.requestCount.WithLabelValues(r.Method, r.URL.Path, status).Inc()
        
        if ww.statusCode >= 400 {
            errorType := "client_error"
            if ww.statusCode >= 500 {
                errorType = "server_error"
            }
            mc.errorCount.WithLabelValues(r.Method, r.URL.Path, errorType).Inc()
        }
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

## Performance Testing in Production

### Load Testing Framework

```go
// Production load testing
type LoadTester struct {
    target     string
    clients    int
    duration   time.Duration
    rampUp     time.Duration
    results    chan TestResult
    metrics    *MetricsCollector
}

type TestResult struct {
    Timestamp    time.Time
    Duration     time.Duration
    StatusCode   int
    Error        error
    RequestSize  int64
    ResponseSize int64
}

func NewLoadTester(target string, clients int, duration time.Duration) *LoadTester {
    return &LoadTester{
        target:   target,
        clients:  clients,
        duration: duration,
        rampUp:   duration / 10, // 10% ramp-up time
        results:  make(chan TestResult, clients*100),
        metrics:  NewMetricsCollector(),
    }
}

func (lt *LoadTester) Run() error {
    ctx, cancel := context.WithTimeout(context.Background(), lt.duration)
    defer cancel()
    
    // Start result collector
    go lt.collectResults()
    
    // Ramp up clients gradually
    clientInterval := lt.rampUp / time.Duration(lt.clients)
    
    var wg sync.WaitGroup
    for i := 0; i < lt.clients; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()
            lt.runClient(ctx, clientID)
        }(i)
        
        // Stagger client start times
        time.Sleep(clientInterval)
    }
    
    wg.Wait()
    close(lt.results)
    
    return nil
}

func (lt *LoadTester) runClient(ctx context.Context, clientID int) {
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            result := lt.makeRequest(client)
            select {
            case lt.results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (lt *LoadTester) makeRequest(client *http.Client) TestResult {
    start := time.Now()
    
    resp, err := client.Get(lt.target)
    duration := time.Since(start)
    
    result := TestResult{
        Timestamp: start,
        Duration:  duration,
        Error:     err,
    }
    
    if err != nil {
        return result
    }
    
    result.StatusCode = resp.StatusCode
    result.ResponseSize = resp.ContentLength
    
    // Read response body to simulate real usage
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    
    return result
}

func (lt *LoadTester) collectResults() {
    var (
        totalRequests   int64
        successRequests int64
        totalDuration   time.Duration
        minDuration     time.Duration = time.Hour
        maxDuration     time.Duration
        durations       []time.Duration
    )
    
    for result := range lt.results {
        totalRequests++
        totalDuration += result.Duration
        
        if result.Error == nil && result.StatusCode < 400 {
            successRequests++
        }
        
        if result.Duration < minDuration {
            minDuration = result.Duration
        }
        if result.Duration > maxDuration {
            maxDuration = result.Duration
        }
        
        durations = append(durations, result.Duration)
    }
    
    // Calculate percentiles
    sort.Slice(durations, func(i, j int) bool {
        return durations[i] < durations[j]
    })
    
    p50 := durations[len(durations)*50/100]
    p95 := durations[len(durations)*95/100]
    p99 := durations[len(durations)*99/100]
    
    fmt.Printf("Load Test Results:\n")
    fmt.Printf("Total Requests: %d\n", totalRequests)
    fmt.Printf("Success Rate: %.2f%%\n", float64(successRequests)/float64(totalRequests)*100)
    fmt.Printf("Average Duration: %v\n", totalDuration/time.Duration(totalRequests))
    fmt.Printf("Min Duration: %v\n", minDuration)
    fmt.Printf("Max Duration: %v\n", maxDuration)
    fmt.Printf("P50 Duration: %v\n", p50)
    fmt.Printf("P95 Duration: %v\n", p95)
    fmt.Printf("P99 Duration: %v\n", p99)
}
```

### Chaos Engineering

```go
// Chaos testing for production resilience
type ChaosTest struct {
    name        string
    probability float64
    impact      ChaosImpact
}

type ChaosImpact interface {
    Apply() error
    Restore() error
}

// Network latency injection
type NetworkLatency struct {
    delay time.Duration
}

func (nl *NetworkLatency) Apply() error {
    // Inject network latency (implementation depends on infrastructure)
    return injectNetworkDelay(nl.delay)
}

func (nl *NetworkLatency) Restore() error {
    return removeNetworkDelay()
}

// Memory pressure simulation
type MemoryPressure struct {
    size int64
    data []byte
}

func (mp *MemoryPressure) Apply() error {
    mp.data = make([]byte, mp.size)
    // Fill with random data to prevent optimization
    rand.Read(mp.data)
    return nil
}

func (mp *MemoryPressure) Restore() error {
    mp.data = nil
    runtime.GC()
    return nil
}

// CPU stress simulation
type CPUStress struct {
    workers int
    done    chan bool
}

func (cs *CPUStress) Apply() error {
    cs.done = make(chan bool)
    
    for i := 0; i < cs.workers; i++ {
        go func() {
            for {
                select {
                case <-cs.done:
                    return
                default:
                    // Busy loop to consume CPU
                    for j := 0; j < 1000000; j++ {
                        _ = j * j
                    }
                }
            }
        }()
    }
    
    return nil
}

func (cs *CPUStress) Restore() error {
    close(cs.done)
    return nil
}

// Chaos test runner
func runChaosTests(tests []ChaosTest, duration time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), duration)
    defer cancel()
    
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            for _, test := range tests {
                if rand.Float64() < test.probability {
                    log.Printf("Applying chaos test: %s", test.name)
                    
                    if err := test.impact.Apply(); err != nil {
                        log.Printf("Failed to apply chaos test %s: %v", test.name, err)
                        continue
                    }
                    
                    // Let chaos run for a short time
                    time.Sleep(30 * time.Second)
                    
                    if err := test.impact.Restore(); err != nil {
                        log.Printf("Failed to restore from chaos test %s: %v", test.name, err)
                    }
                    
                    log.Printf("Restored from chaos test: %s", test.name)
                }
            }
        }
    }
}
```

## Production Optimization Strategies

### Resource Management

```go
// Production resource management
type ResourceManager struct {
    pools map[string]*sync.Pool
    mu    sync.RWMutex
}

func NewResourceManager() *ResourceManager {
    rm := &ResourceManager{
        pools: make(map[string]*sync.Pool),
    }
    
    // Pre-configure common pools
    rm.RegisterPool("buffer", func() interface{} {
        return make([]byte, 0, 4096)
    })
    
    rm.RegisterPool("strings", func() interface{} {
        return &strings.Builder{}
    })
    
    rm.RegisterPool("json_encoder", func() interface{} {
        var buf bytes.Buffer
        return json.NewEncoder(&buf)
    })
    
    return rm
}

func (rm *ResourceManager) RegisterPool(name string, factory func() interface{}) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    rm.pools[name] = &sync.Pool{New: factory}
}

func (rm *ResourceManager) Get(name string) interface{} {
    rm.mu.RLock()
    pool, exists := rm.pools[name]
    rm.mu.RUnlock()
    
    if !exists {
        return nil
    }
    
    return pool.Get()
}

func (rm *ResourceManager) Put(name string, obj interface{}) {
    rm.mu.RLock()
    pool, exists := rm.pools[name]
    rm.mu.RUnlock()
    
    if exists {
        pool.Put(obj)
    }
}
```

### Circuit Breaker Pattern

```go
// Circuit breaker for production resilience
type CircuitBreaker struct {
    name          string
    maxFailures   int
    resetTimeout  time.Duration
    state         CircuitState
    failures      int
    lastFailTime  time.Time
    mu            sync.RWMutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        name:         name,
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        Closed,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case Open:
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = HalfOpen
            cb.failures = 0
        } else {
            return fmt.Errorf("circuit breaker %s is open", cb.name)
        }
    case HalfOpen:
        // Allow one request to test if service is recovered
    case Closed:
        // Normal operation
    }
    
    err := fn()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = Open
        }
        
        return err
    }
    
    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = Closed
    
    return nil
}

func (cb *CircuitBreaker) State() CircuitState {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}
```

## Observability and Debugging

### Distributed Tracing

```go
// Distributed tracing for production debugging
type TraceSpan struct {
    ID       string
    ParentID string
    Name     string
    Start    time.Time
    End      time.Time
    Tags     map[string]string
    Logs     []TraceLog
}

type TraceLog struct {
    Timestamp time.Time
    Message   string
    Level     string
}

type Tracer struct {
    spans map[string]*TraceSpan
    mu    sync.RWMutex
}

func NewTracer() *Tracer {
    return &Tracer{
        spans: make(map[string]*TraceSpan),
    }
}

func (t *Tracer) StartSpan(name string, parentID string) *TraceSpan {
    span := &TraceSpan{
        ID:       generateSpanID(),
        ParentID: parentID,
        Name:     name,
        Start:    time.Now(),
        Tags:     make(map[string]string),
    }
    
    t.mu.Lock()
    t.spans[span.ID] = span
    t.mu.Unlock()
    
    return span
}

func (t *Tracer) FinishSpan(spanID string) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if span, exists := t.spans[spanID]; exists {
        span.End = time.Now()
        // Send span to tracing backend
        t.exportSpan(span)
    }
}

func (t *Tracer) exportSpan(span *TraceSpan) {
    // Export to Jaeger, Zipkin, or other tracing systems
    // Implementation depends on chosen tracing backend
}

// Context propagation
type contextKey string

const traceContextKey contextKey = "trace"

func WithTrace(ctx context.Context, span *TraceSpan) context.Context {
    return context.WithValue(ctx, traceContextKey, span)
}

func SpanFromContext(ctx context.Context) (*TraceSpan, bool) {
    span, ok := ctx.Value(traceContextKey).(*TraceSpan)
    return span, ok
}
```

## Best Practices for Production Performance

### 1. Monitoring and Alerting
- Implement comprehensive metrics collection
- Set up alerting for performance degradation
- Use distributed tracing for debugging
- Monitor business metrics alongside technical metrics

### 2. Graceful Degradation
- Implement circuit breakers for external dependencies
- Use timeouts and retries with backoff
- Provide fallback mechanisms
- Design for partial failures

### 3. Resource Management
- Use connection pooling for databases and external services
- Implement proper resource cleanup
- Monitor resource usage continuously
- Set appropriate limits and quotas

### 4. Testing and Validation
- Conduct regular load testing
- Implement chaos engineering practices
- Validate performance in staging environments
- Use canary deployments for performance validation

Production performance engineering requires a holistic approach combining monitoring, testing, optimization, and operational excellence to ensure applications perform reliably under real-world conditions.
