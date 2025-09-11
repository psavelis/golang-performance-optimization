# Load Testing

Comprehensive guide to load testing Go applications, from basic HTTP load testing to complex distributed system validation.

## Load Testing Fundamentals

Load testing validates application performance under realistic user loads and identifies performance bottlenecks before they impact production.

### Basic HTTP Load Testing

```go
// Simple HTTP load test framework
type LoadTest struct {
    URL         string
    Method      string
    Headers     map[string]string
    Body        []byte
    Clients     int
    Duration    time.Duration
    RampUpTime  time.Duration
    RequestRate int // requests per second (0 = unlimited)
}

type LoadTestResult struct {
    TotalRequests     int64
    SuccessRequests   int64
    FailedRequests    int64
    AverageLatency    time.Duration
    P50Latency        time.Duration
    P95Latency        time.Duration
    P99Latency        time.Duration
    MinLatency        time.Duration
    MaxLatency        time.Duration
    RequestsPerSecond float64
    BytesRead         int64
    BytesWritten      int64
    Errors            map[string]int64
}

type RequestResult struct {
    Timestamp    time.Time
    Latency      time.Duration
    StatusCode   int
    Error        error
    BytesRead    int64
    BytesWritten int64
}

func (lt *LoadTest) Run() (*LoadTestResult, error) {
    results := make(chan RequestResult, lt.Clients*100)
    ctx, cancel := context.WithTimeout(context.Background(), lt.Duration)
    defer cancel()
    
    // Start result collector
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        lt.collectResults(ctx, results)
    }()
    
    // Calculate client start intervals for ramp-up
    var clientInterval time.Duration
    if lt.RampUpTime > 0 && lt.Clients > 1 {
        clientInterval = lt.RampUpTime / time.Duration(lt.Clients-1)
    }
    
    // Start load generators
    for i := 0; i < lt.Clients; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()
            
            // Stagger client starts during ramp-up
            if clientInterval > 0 {
                time.Sleep(time.Duration(clientID) * clientInterval)
            }
            
            lt.runClient(ctx, results)
        }(i)
    }
    
    wg.Wait()
    close(results)
    
    return lt.analyzeResults(), nil
}

func (lt *LoadTest) runClient(ctx context.Context, results chan<- RequestResult) {
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    var rateLimiter <-chan time.Time
    if lt.RequestRate > 0 {
        ticker := time.NewTicker(time.Second / time.Duration(lt.RequestRate))
        defer ticker.Stop()
        rateLimiter = ticker.C
    }
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            if rateLimiter != nil {
                <-rateLimiter
            }
            
            result := lt.makeRequest(client)
            select {
            case results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (lt *LoadTest) makeRequest(client *http.Client) RequestResult {
    start := time.Now()
    
    req, err := http.NewRequest(lt.Method, lt.URL, bytes.NewReader(lt.Body))
    if err != nil {
        return RequestResult{
            Timestamp: start,
            Error:     err,
        }
    }
    
    // Add headers
    for key, value := range lt.Headers {
        req.Header.Set(key, value)
    }
    
    resp, err := client.Do(req)
    latency := time.Since(start)
    
    result := RequestResult{
        Timestamp:    start,
        Latency:      latency,
        BytesWritten: int64(len(lt.Body)),
    }
    
    if err != nil {
        result.Error = err
        return result
    }
    
    result.StatusCode = resp.StatusCode
    
    // Read response body
    bodyBytes, err := io.ReadAll(resp.Body)
    resp.Body.Close()
    
    if err != nil {
        result.Error = err
    } else {
        result.BytesRead = int64(len(bodyBytes))
    }
    
    return result
}

func (lt *LoadTest) collectResults(ctx context.Context, results <-chan RequestResult) {
    // Results are collected and analyzed in analyzeResults()
}

func (lt *LoadTest) analyzeResults() *LoadTestResult {
    // Implementation would collect and analyze all results
    // Calculate percentiles, averages, error rates, etc.
    return &LoadTestResult{}
}
```

### Advanced Load Testing Scenarios

```go
// Multi-endpoint load testing
type MultiEndpointLoadTest struct {
    Endpoints []EndpointConfig
    Clients   int
    Duration  time.Duration
    Profile   LoadProfile
}

type EndpointConfig struct {
    URL     string
    Method  string
    Headers map[string]string
    Body    []byte
    Weight  int // Relative frequency
}

type LoadProfile string

const (
    ConstantLoad LoadProfile = "constant"
    RampUpLoad   LoadProfile = "rampup"
    SpikeLoad    LoadProfile = "spike"
    StepLoad     LoadProfile = "step"
)

func (mlt *MultiEndpointLoadTest) Run() error {
    // Create weighted endpoint selector
    selector := NewWeightedSelector(mlt.Endpoints)
    
    ctx, cancel := context.WithTimeout(context.Background(), mlt.Duration)
    defer cancel()
    
    var wg sync.WaitGroup
    
    // Start clients with different load profiles
    switch mlt.Profile {
    case ConstantLoad:
        mlt.runConstantLoad(ctx, &wg, selector)
    case RampUpLoad:
        mlt.runRampUpLoad(ctx, &wg, selector)
    case SpikeLoad:
        mlt.runSpikeLoad(ctx, &wg, selector)
    case StepLoad:
        mlt.runStepLoad(ctx, &wg, selector)
    }
    
    wg.Wait()
    return nil
}

func (mlt *MultiEndpointLoadTest) runConstantLoad(ctx context.Context, wg *sync.WaitGroup, selector *WeightedSelector) {
    for i := 0; i < mlt.Clients; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mlt.runConstantClient(ctx, selector)
        }()
    }
}

func (mlt *MultiEndpointLoadTest) runRampUpLoad(ctx context.Context, wg *sync.WaitGroup, selector *WeightedSelector) {
    rampUpDuration := mlt.Duration / 3 // First third for ramp-up
    clientInterval := rampUpDuration / time.Duration(mlt.Clients)
    
    for i := 0; i < mlt.Clients; i++ {
        wg.Add(1)
        go func(delay time.Duration) {
            defer wg.Done()
            
            // Wait for ramp-up delay
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return
            }
            
            mlt.runConstantClient(ctx, selector)
        }(time.Duration(i) * clientInterval)
    }
}

func (mlt *MultiEndpointLoadTest) runSpikeLoad(ctx context.Context, wg *sync.WaitGroup, selector *WeightedSelector) {
    baseClients := mlt.Clients / 4
    spikeClients := mlt.Clients - baseClients
    
    // Start base load
    for i := 0; i < baseClients; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mlt.runConstantClient(ctx, selector)
        }()
    }
    
    // Wait for spike time (middle of test)
    spikeDelay := mlt.Duration / 2
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        
        select {
        case <-time.After(spikeDelay):
        case <-ctx.Done():
            return
        }
        
        // Start spike clients
        var spikeWg sync.WaitGroup
        for i := 0; i < spikeClients; i++ {
            spikeWg.Add(1)
            go func() {
                defer spikeWg.Done()
                mlt.runConstantClient(ctx, selector)
            }()
        }
        spikeWg.Wait()
    }()
}

func (mlt *MultiEndpointLoadTest) runStepLoad(ctx context.Context, wg *sync.WaitGroup, selector *WeightedSelector) {
    steps := 4
    clientsPerStep := mlt.Clients / steps
    stepDuration := mlt.Duration / time.Duration(steps)
    
    for step := 0; step < steps; step++ {
        stepClients := (step + 1) * clientsPerStep
        
        wg.Add(1)
        go func(clients int, delay time.Duration) {
            defer wg.Done()
            
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return
            }
            
            var stepWg sync.WaitGroup
            for i := 0; i < clients; i++ {
                stepWg.Add(1)
                go func() {
                    defer stepWg.Done()
                    mlt.runConstantClient(ctx, selector)
                }()
            }
            stepWg.Wait()
        }(stepClients, time.Duration(step)*stepDuration)
    }
}

func (mlt *MultiEndpointLoadTest) runConstantClient(ctx context.Context, selector *WeightedSelector) {
    client := &http.Client{Timeout: 30 * time.Second}
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            endpoint := selector.Select()
            mlt.makeRequest(client, endpoint)
        }
    }
}

func (mlt *MultiEndpointLoadTest) makeRequest(client *http.Client, endpoint EndpointConfig) {
    req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewReader(endpoint.Body))
    if err != nil {
        return
    }
    
    for key, value := range endpoint.Headers {
        req.Header.Set(key, value)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return
    }
    
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
}
```

### Weighted Endpoint Selection

```go
// Weighted random selector for realistic load distribution
type WeightedSelector struct {
    endpoints []EndpointConfig
    weights   []int
    totalWeight int
}

func NewWeightedSelector(endpoints []EndpointConfig) *WeightedSelector {
    totalWeight := 0
    weights := make([]int, len(endpoints))
    
    for i, endpoint := range endpoints {
        weight := endpoint.Weight
        if weight <= 0 {
            weight = 1
        }
        weights[i] = weight
        totalWeight += weight
    }
    
    return &WeightedSelector{
        endpoints:   endpoints,
        weights:     weights,
        totalWeight: totalWeight,
    }
}

func (ws *WeightedSelector) Select() EndpointConfig {
    r := rand.Intn(ws.totalWeight)
    
    for i, weight := range ws.weights {
        r -= weight
        if r < 0 {
            return ws.endpoints[i]
        }
    }
    
    return ws.endpoints[0] // Fallback
}
```

### Database Load Testing

```go
// Database connection pool load testing
type DatabaseLoadTest struct {
    DSN         string
    MaxConns    int
    Duration    time.Duration
    Queries     []QueryTest
    Clients     int
}

type QueryTest struct {
    SQL    string
    Args   []interface{}
    Weight int
}

func (dlt *DatabaseLoadTest) Run() error {
    db, err := sql.Open("postgres", dlt.DSN)
    if err != nil {
        return err
    }
    defer db.Close()
    
    db.SetMaxOpenConns(dlt.MaxConns)
    db.SetMaxIdleConns(dlt.MaxConns / 2)
    db.SetConnMaxLifetime(time.Hour)
    
    ctx, cancel := context.WithTimeout(context.Background(), dlt.Duration)
    defer cancel()
    
    selector := NewQuerySelector(dlt.Queries)
    
    var wg sync.WaitGroup
    for i := 0; i < dlt.Clients; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            dlt.runDatabaseClient(ctx, db, selector)
        }()
    }
    
    wg.Wait()
    return nil
}

func (dlt *DatabaseLoadTest) runDatabaseClient(ctx context.Context, db *sql.DB, selector *QuerySelector) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            query := selector.Select()
            dlt.executeQuery(ctx, db, query)
        }
    }
}

func (dlt *DatabaseLoadTest) executeQuery(ctx context.Context, db *sql.DB, query QueryTest) {
    start := time.Now()
    
    rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
    if err != nil {
        log.Printf("Query error: %v", err)
        return
    }
    defer rows.Close()
    
    // Process results
    for rows.Next() {
        // Scan into discard to simulate real processing
        var dummy interface{}
        rows.Scan(&dummy)
    }
    
    latency := time.Since(start)
    log.Printf("Query executed in %v", latency)
}

type QuerySelector struct {
    queries []QueryTest
    weights []int
    total   int
}

func NewQuerySelector(queries []QueryTest) *QuerySelector {
    total := 0
    weights := make([]int, len(queries))
    
    for i, query := range queries {
        weight := query.Weight
        if weight <= 0 {
            weight = 1
        }
        weights[i] = weight
        total += weight
    }
    
    return &QuerySelector{
        queries: queries,
        weights: weights,
        total:   total,
    }
}

func (qs *QuerySelector) Select() QueryTest {
    r := rand.Intn(qs.total)
    
    for i, weight := range qs.weights {
        r -= weight
        if r < 0 {
            return qs.queries[i]
        }
    }
    
    return qs.queries[0]
}
```

### Load Testing with Real User Behavior

```go
// Realistic user session simulation
type UserSession struct {
    UserID    string
    SessionID string
    Actions   []UserAction
}

type UserAction struct {
    Type     ActionType
    URL      string
    Method   string
    Headers  map[string]string
    Body     []byte
    ThinkTime time.Duration // Time between actions
}

type ActionType string

const (
    Login       ActionType = "login"
    Browse      ActionType = "browse"
    Search      ActionType = "search"
    Purchase    ActionType = "purchase"
    Logout      ActionType = "logout"
)

func GenerateUserSession(userID string) UserSession {
    sessionID := fmt.Sprintf("session_%s_%d", userID, time.Now().Unix())
    
    actions := []UserAction{
        {
            Type:      Login,
            URL:       "/api/login",
            Method:    "POST",
            Body:      []byte(`{"username":"` + userID + `","password":"test123"}`),
            ThinkTime: time.Duration(rand.Intn(3)+1) * time.Second,
        },
        {
            Type:      Browse,
            URL:       "/api/products",
            Method:    "GET",
            ThinkTime: time.Duration(rand.Intn(5)+2) * time.Second,
        },
        {
            Type:      Search,
            URL:       "/api/search?q=laptop",
            Method:    "GET",
            ThinkTime: time.Duration(rand.Intn(10)+5) * time.Second,
        },
    }
    
    // 30% chance of purchase
    if rand.Float32() < 0.3 {
        actions = append(actions, UserAction{
            Type:      Purchase,
            URL:       "/api/purchase",
            Method:    "POST",
            Body:      []byte(`{"product_id":"12345","quantity":1}`),
            ThinkTime: time.Duration(rand.Intn(5)+1) * time.Second,
        })
    }
    
    actions = append(actions, UserAction{
        Type:      Logout,
        URL:       "/api/logout",
        Method:    "POST",
        ThinkTime: 0,
    })
    
    return UserSession{
        UserID:    userID,
        SessionID: sessionID,
        Actions:   actions,
    }
}

type UserBehaviorLoadTest struct {
    BaseURL       string
    ConcurrentUsers int
    Duration      time.Duration
    NewUserRate   int // New users per second
}

func (ublt *UserBehaviorLoadTest) Run() error {
    ctx, cancel := context.WithTimeout(context.Background(), ublt.Duration)
    defer cancel()
    
    userIDCounter := int64(0)
    
    var wg sync.WaitGroup
    
    // Start existing users
    for i := 0; i < ublt.ConcurrentUsers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    userID := fmt.Sprintf("user_%d", atomic.AddInt64(&userIDCounter, 1))
                    session := GenerateUserSession(userID)
                    ublt.runUserSession(ctx, session)
                    
                    // Wait before starting new session
                    time.Sleep(time.Duration(rand.Intn(30)+10) * time.Second)
                }
            }
        }()
    }
    
    // Start new users at specified rate
    if ublt.NewUserRate > 0 {
        ticker := time.NewTicker(time.Second / time.Duration(ublt.NewUserRate))
        defer ticker.Stop()
        
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for {
                select {
                case <-ctx.Done():
                    return
                case <-ticker.C:
                    wg.Add(1)
                    go func() {
                        defer wg.Done()
                        userID := fmt.Sprintf("new_user_%d", atomic.AddInt64(&userIDCounter, 1))
                        session := GenerateUserSession(userID)
                        ublt.runUserSession(ctx, session)
                    }()
                }
            }
        }()
    }
    
    wg.Wait()
    return nil
}

func (ublt *UserBehaviorLoadTest) runUserSession(ctx context.Context, session UserSession) {
    client := &http.Client{
        Timeout: 30 * time.Second,
        Jar:     &cookieJar{}, // Maintain session cookies
    }
    
    for _, action := range session.Actions {
        select {
        case <-ctx.Done():
            return
        default:
            ublt.executeAction(client, action)
            
            if action.ThinkTime > 0 {
                select {
                case <-time.After(action.ThinkTime):
                case <-ctx.Done():
                    return
                }
            }
        }
    }
}

func (ublt *UserBehaviorLoadTest) executeAction(client *http.Client, action UserAction) {
    url := ublt.BaseURL + action.URL
    
    req, err := http.NewRequest(action.Method, url, bytes.NewReader(action.Body))
    if err != nil {
        return
    }
    
    for key, value := range action.Headers {
        req.Header.Set(key, value)
    }
    
    if action.Type == Login || action.Type == Purchase {
        req.Header.Set("Content-Type", "application/json")
    }
    
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Action %s failed: %v", action.Type, err)
        return
    }
    
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
}

type cookieJar struct {
    cookies []*http.Cookie
}

func (cj *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
    cj.cookies = append(cj.cookies, cookies...)
}

func (cj *cookieJar) Cookies(u *url.URL) []*http.Cookie {
    return cj.cookies
}
```

## Load Testing Best Practices

### 1. Realistic Test Scenarios
- Model actual user behavior patterns
- Include think time between requests
- Use realistic data sizes and payloads
- Test different user types and workflows

### 2. Gradual Load Increase
- Start with baseline load
- Gradually increase to target load
- Include spike testing
- Test sustained load over time

### 3. Comprehensive Monitoring
- Monitor both client and server metrics
- Track response times, error rates, and throughput
- Monitor system resources (CPU, memory, disk, network)
- Use distributed monitoring for multi-service systems

### 4. Environment Considerations
- Test in production-like environments
- Account for network latency and bandwidth
- Consider database and cache performance
- Test with realistic data volumes

Load testing is essential for validating application performance under realistic conditions and ensuring systems can handle expected user loads without degradation.
