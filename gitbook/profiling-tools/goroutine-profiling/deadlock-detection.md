# Deadlock Detection

Deadlock detection is crucial for maintaining application stability in concurrent Go systems. This comprehensive guide covers advanced techniques for detecting, analyzing, and preventing deadlocks involving goroutines, channels, mutexes, and other synchronization primitives.

## Understanding Deadlocks

Deadlocks occur when:
- **Circular waiting** - Goroutines wait for each other in a cycle
- **Hold and wait** - Goroutines hold resources while waiting for others
- **No preemption** - Resources cannot be forcibly taken from goroutines
- **Mutual exclusion** - Resources can only be used by one goroutine at a time

### Advanced Deadlock Detection System

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "sort"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

// DeadlockDetector provides comprehensive deadlock detection and analysis
type DeadlockDetector struct {
    resources        map[string]*Resource
    goroutines       map[int]*GoroutineState
    dependencies     *DependencyGraph
    detectionConfig  *DetectionConfig
    statistics       *DeadlockStatistics
    alertCallbacks   []DeadlockCallback
    enabled          bool
    mu               sync.RWMutex
}

type Resource struct {
    ID          string
    Type        ResourceType
    Owner       int  // Goroutine ID that owns this resource
    Waiters     []int // Goroutine IDs waiting for this resource
    CreatedAt   time.Time
    LastAccess  time.Time
    AccessCount int64
    Properties  map[string]interface{}
}

type ResourceType int

const (
    ResourceMutex ResourceType = iota
    ResourceRWMutex
    ResourceChannel
    ResourceSemaphore
    ResourceWaitGroup
    ResourceCustom
)

func (rt ResourceType) String() string {
    switch rt {
    case ResourceMutex:
        return "Mutex"
    case ResourceRWMutex:
        return "RWMutex"
    case ResourceChannel:
        return "Channel"
    case ResourceSemaphore:
        return "Semaphore"
    case ResourceWaitGroup:
        return "WaitGroup"
    case ResourceCustom:
        return "Custom"
    default:
        return "Unknown"
    }
}

type GoroutineState struct {
    ID            int
    Name          string
    State         string
    WaitingFor    []string // Resource IDs
    Holding       []string // Resource IDs
    StackTrace    []string
    CreatedAt     time.Time
    LastActivity  time.Time
    BlockedSince  time.Time
    Priority      int
    Context       string
}

type DependencyGraph struct {
    nodes map[string]*GraphNode
    edges map[string]map[string]*GraphEdge
    mu    sync.RWMutex
}

type GraphNode struct {
    ID       string
    Type     NodeType
    Data     interface{}
    InDegree int
    OutDegree int
}

type NodeType int

const (
    NodeGoroutine NodeType = iota
    NodeResource
)

type GraphEdge struct {
    From       string
    To         string
    Type       EdgeType
    Weight     float64
    CreatedAt  time.Time
    Properties map[string]interface{}
}

type EdgeType int

const (
    EdgeWaitsFor EdgeType = iota
    EdgeHolds
    EdgeRequests
)

type DetectionConfig struct {
    ScanInterval        time.Duration
    DeadlockThreshold   time.Duration
    MaxGoroutines       int
    MaxResources        int
    EnableCycleDetection bool
    EnableHeuristics     bool
    AlertOnSuspicion     bool
}

type DeadlockStatistics struct {
    TotalScans         int64
    DetectedDeadlocks  int64
    FalsePositives     int64
    PreventedDeadlocks int64
    AverageDetectionTime time.Duration
    MaxGraphSize       int
    LastScanTime       time.Time
}

type DeadlockCallback func(deadlock DetectedDeadlock)

type DetectedDeadlock struct {
    ID                string
    Type              DeadlockType
    ParticipantCount  int
    Participants      []DeadlockParticipant
    Cycle             []CycleStep
    DetectedAt        time.Time
    Confidence        float64
    Severity          DeadlockSeverity
    Evidence          []DeadlockEvidence
    Resolution        []ResolutionSuggestion
}

type DeadlockType int

const (
    DeadlockCircular DeadlockType = iota
    DeadlockResource
    DeadlockChannel
    DeadlockLivelock
    DeadlockStarvation
)

func (dt DeadlockType) String() string {
    switch dt {
    case DeadlockCircular:
        return "Circular"
    case DeadlockResource:
        return "Resource"
    case DeadlockChannel:
        return "Channel"
    case DeadlockLivelock:
        return "Livelock"
    case DeadlockStarvation:
        return "Starvation"
    default:
        return "Unknown"
    }
}

type DeadlockParticipant struct {
    GoroutineID int
    GoroutineName string
    Resources   []string
    State       string
    StackTrace  []string
}

type CycleStep struct {
    From         string
    To           string
    ResourceType string
    Action       string
    Timestamp    time.Time
}

type DeadlockSeverity int

const (
    SeverityLow DeadlockSeverity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)

func (ds DeadlockSeverity) String() string {
    switch ds {
    case SeverityLow:
        return "Low"
    case SeverityMedium:
        return "Medium"
    case SeverityHigh:
        return "High"
    case SeverityCritical:
        return "Critical"
    default:
        return "Unknown"
    }
}

type DeadlockEvidence struct {
    Type        string
    Description string
    Timestamp   time.Time
    Data        interface{}
}

type ResolutionSuggestion struct {
    Strategy    string
    Description string
    Priority    int
    Difficulty  string
    Impact      string
}

func NewDeadlockDetector() *DeadlockDetector {
    return &DeadlockDetector{
        resources:   make(map[string]*Resource),
        goroutines:  make(map[int]*GoroutineState),
        dependencies: &DependencyGraph{
            nodes: make(map[string]*GraphNode),
            edges: make(map[string]map[string]*GraphEdge),
        },
        detectionConfig: &DetectionConfig{
            ScanInterval:         time.Second * 5,
            DeadlockThreshold:    time.Second * 30,
            MaxGoroutines:        10000,
            MaxResources:         1000,
            EnableCycleDetection: true,
            EnableHeuristics:     true,
            AlertOnSuspicion:     true,
        },
        statistics: &DeadlockStatistics{},
    }
}

func (dd *DeadlockDetector) Enable() {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    if dd.enabled {
        return
    }
    
    dd.enabled = true
    go dd.detectionLoop()
}

func (dd *DeadlockDetector) Disable() {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    dd.enabled = false
}

func (dd *DeadlockDetector) RegisterCallback(callback DeadlockCallback) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    dd.alertCallbacks = append(dd.alertCallbacks, callback)
}

func (dd *DeadlockDetector) RegisterResource(id string, resourceType ResourceType, properties map[string]interface{}) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    resource := &Resource{
        ID:         id,
        Type:       resourceType,
        CreatedAt:  time.Now(),
        Properties: properties,
    }
    
    dd.resources[id] = resource
    
    // Add to dependency graph
    dd.dependencies.mu.Lock()
    dd.dependencies.nodes[id] = &GraphNode{
        ID:   id,
        Type: NodeResource,
        Data: resource,
    }
    dd.dependencies.mu.Unlock()
}

func (dd *DeadlockDetector) UnregisterResource(id string) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    delete(dd.resources, id)
    
    // Remove from dependency graph
    dd.dependencies.mu.Lock()
    delete(dd.dependencies.nodes, id)
    delete(dd.dependencies.edges, id)
    
    // Remove edges to this node
    for fromID, toMap := range dd.dependencies.edges {
        delete(toMap, id)
        if len(toMap) == 0 {
            delete(dd.dependencies.edges, fromID)
        }
    }
    dd.dependencies.mu.Unlock()
}

func (dd *DeadlockDetector) RecordResourceAcquisition(goroutineID int, resourceID string) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    // Update resource ownership
    if resource, exists := dd.resources[resourceID]; exists {
        resource.Owner = goroutineID
        resource.LastAccess = time.Now()
        atomic.AddInt64(&resource.AccessCount, 1)
        
        // Remove from waiters if present
        for i, waiter := range resource.Waiters {
            if waiter == goroutineID {
                resource.Waiters = append(resource.Waiters[:i], resource.Waiters[i+1:]...)
                break
            }
        }
    }
    
    // Update goroutine state
    if goroutine, exists := dd.goroutines[goroutineID]; exists {
        goroutine.Holding = append(goroutine.Holding, resourceID)
        goroutine.LastActivity = time.Now()
        
        // Remove from waiting
        for i, waiting := range goroutine.WaitingFor {
            if waiting == resourceID {
                goroutine.WaitingFor = append(goroutine.WaitingFor[:i], goroutine.WaitingFor[i+1:]...)
                break
            }
        }
    }
    
    // Update dependency graph
    dd.updateDependencyGraph()
}

func (dd *DeadlockDetector) RecordResourceRelease(goroutineID int, resourceID string) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    // Update resource ownership
    if resource, exists := dd.resources[resourceID]; exists {
        if resource.Owner == goroutineID {
            resource.Owner = 0 // No owner
            resource.LastAccess = time.Now()
        }
    }
    
    // Update goroutine state
    if goroutine, exists := dd.goroutines[goroutineID]; exists {
        for i, holding := range goroutine.Holding {
            if holding == resourceID {
                goroutine.Holding = append(goroutine.Holding[:i], goroutine.Holding[i+1:]...)
                break
            }
        }
        goroutine.LastActivity = time.Now()
    }
    
    // Update dependency graph
    dd.updateDependencyGraph()
}

func (dd *DeadlockDetector) RecordResourceWait(goroutineID int, resourceID string) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    // Update resource waiters
    if resource, exists := dd.resources[resourceID]; exists {
        // Add to waiters if not already present
        found := false
        for _, waiter := range resource.Waiters {
            if waiter == goroutineID {
                found = true
                break
            }
        }
        if !found {
            resource.Waiters = append(resource.Waiters, goroutineID)
        }
    }
    
    // Update goroutine state
    if goroutine, exists := dd.goroutines[goroutineID]; exists {
        goroutine.WaitingFor = append(goroutine.WaitingFor, resourceID)
        goroutine.BlockedSince = time.Now()
        goroutine.LastActivity = time.Now()
    }
    
    // Update dependency graph
    dd.updateDependencyGraph()
}

func (dd *DeadlockDetector) UpdateGoroutineState(goroutineID int, name, state string, stackTrace []string) {
    dd.mu.Lock()
    defer dd.mu.Unlock()
    
    now := time.Now()
    
    if goroutine, exists := dd.goroutines[goroutineID]; exists {
        goroutine.Name = name
        goroutine.State = state
        goroutine.StackTrace = stackTrace
        goroutine.LastActivity = now
    } else {
        goroutine := &GoroutineState{
            ID:           goroutineID,
            Name:         name,
            State:        state,
            StackTrace:   stackTrace,
            CreatedAt:    now,
            LastActivity: now,
            WaitingFor:   make([]string, 0),
            Holding:      make([]string, 0),
        }
        dd.goroutines[goroutineID] = goroutine
        
        // Add to dependency graph
        dd.dependencies.mu.Lock()
        goroutineKey := fmt.Sprintf("goroutine:%d", goroutineID)
        dd.dependencies.nodes[goroutineKey] = &GraphNode{
            ID:   goroutineKey,
            Type: NodeGoroutine,
            Data: goroutine,
        }
        dd.dependencies.mu.Unlock()
    }
}

func (dd *DeadlockDetector) updateDependencyGraph() {
    dd.dependencies.mu.Lock()
    defer dd.dependencies.mu.Unlock()
    
    // Clear existing edges
    dd.dependencies.edges = make(map[string]map[string]*GraphEdge)
    
    // Build new edges based on current state
    for goroutineID, goroutine := range dd.goroutines {
        goroutineKey := fmt.Sprintf("goroutine:%d", goroutineID)
        
        // Edges from goroutine to resources it's waiting for
        for _, resourceID := range goroutine.WaitingFor {
            dd.addEdge(goroutineKey, resourceID, EdgeWaitsFor, 1.0)
        }
        
        // Edges from resources to goroutines that hold them
        for _, resourceID := range goroutine.Holding {
            dd.addEdge(resourceID, goroutineKey, EdgeHolds, 1.0)
        }
    }
}

func (dd *DeadlockDetector) addEdge(from, to string, edgeType EdgeType, weight float64) {
    if dd.dependencies.edges[from] == nil {
        dd.dependencies.edges[from] = make(map[string]*GraphEdge)
    }
    
    edge := &GraphEdge{
        From:      from,
        To:        to,
        Type:      edgeType,
        Weight:    weight,
        CreatedAt: time.Now(),
    }
    
    dd.dependencies.edges[from][to] = edge
    
    // Update node degrees
    if fromNode, exists := dd.dependencies.nodes[from]; exists {
        fromNode.OutDegree++
    }
    if toNode, exists := dd.dependencies.nodes[to]; exists {
        toNode.InDegree++
    }
}

func (dd *DeadlockDetector) detectionLoop() {
    ticker := time.NewTicker(dd.detectionConfig.ScanInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        dd.mu.RLock()
        enabled := dd.enabled
        dd.mu.RUnlock()
        
        if !enabled {
            return
        }
        
        dd.scanForDeadlocks()
    }
}

func (dd *DeadlockDetector) scanForDeadlocks() {
    start := time.Now()
    atomic.AddInt64(&dd.statistics.TotalScans, 1)
    
    deadlocks := dd.detectDeadlocks()
    
    for _, deadlock := range deadlocks {
        atomic.AddInt64(&dd.statistics.DetectedDeadlocks, 1)
        dd.notifyCallbacks(deadlock)
    }
    
    detectionTime := time.Since(start)
    dd.updateDetectionTime(detectionTime)
    dd.statistics.LastScanTime = time.Now()
}

func (dd *DeadlockDetector) detectDeadlocks() []DetectedDeadlock {
    dd.mu.RLock()
    defer dd.mu.RUnlock()
    
    var deadlocks []DetectedDeadlock
    
    if dd.detectionConfig.EnableCycleDetection {
        cycles := dd.detectCycles()
        for _, cycle := range cycles {
            deadlock := dd.analyzeCycle(cycle)
            if deadlock.Confidence >= 0.7 { // 70% confidence threshold
                deadlocks = append(deadlocks, deadlock)
            }
        }
    }
    
    if dd.detectionConfig.EnableHeuristics {
        heuristicDeadlocks := dd.detectHeuristicDeadlocks()
        deadlocks = append(deadlocks, heuristicDeadlocks...)
    }
    
    return deadlocks
}

func (dd *DeadlockDetector) detectCycles() [][]string {
    dd.dependencies.mu.RLock()
    defer dd.dependencies.mu.RUnlock()
    
    visited := make(map[string]bool)
    recursionStack := make(map[string]bool)
    var cycles [][]string
    
    for nodeID := range dd.dependencies.nodes {
        if !visited[nodeID] {
            path := make([]string, 0)
            if cycle := dd.dfsDetectCycle(nodeID, visited, recursionStack, path); cycle != nil {
                cycles = append(cycles, cycle)
            }
        }
    }
    
    return cycles
}

func (dd *DeadlockDetector) dfsDetectCycle(nodeID string, visited, recursionStack map[string]bool, path []string) []string {
    visited[nodeID] = true
    recursionStack[nodeID] = true
    path = append(path, nodeID)
    
    if edges, exists := dd.dependencies.edges[nodeID]; exists {
        for nextNodeID := range edges {
            if !visited[nextNodeID] {
                if cycle := dd.dfsDetectCycle(nextNodeID, visited, recursionStack, path); cycle != nil {
                    return cycle
                }
            } else if recursionStack[nextNodeID] {
                // Found cycle - extract cycle from path
                cycleStart := -1
                for i, pathNode := range path {
                    if pathNode == nextNodeID {
                        cycleStart = i
                        break
                    }
                }
                if cycleStart >= 0 {
                    cycle := make([]string, len(path)-cycleStart)
                    copy(cycle, path[cycleStart:])
                    cycle = append(cycle, nextNodeID) // Close the cycle
                    return cycle
                }
            }
        }
    }
    
    recursionStack[nodeID] = false
    return nil
}

func (dd *DeadlockDetector) analyzeCycle(cycle []string) DetectedDeadlock {
    var participants []DeadlockParticipant
    var steps []CycleStep
    confidence := 1.0
    
    // Analyze each step in the cycle
    for i := 0; i < len(cycle)-1; i++ {
        from := cycle[i]
        to := cycle[i+1]
        
        step := CycleStep{
            From:      from,
            To:        to,
            Timestamp: time.Now(),
        }
        
        // Determine action type
        if dd.dependencies.edges[from] != nil {
            if edge, exists := dd.dependencies.edges[from][to]; exists {
                switch edge.Type {
                case EdgeWaitsFor:
                    step.Action = "waits_for"
                case EdgeHolds:
                    step.Action = "holds"
                case EdgeRequests:
                    step.Action = "requests"
                }
            }
        }
        
        steps = append(steps, step)
        
        // Add participant if it's a goroutine
        if strings.HasPrefix(from, "goroutine:") {
            var goroutineID int
            fmt.Sscanf(from, "goroutine:%d", &goroutineID)
            
            if goroutine, exists := dd.goroutines[goroutineID]; exists {
                participant := DeadlockParticipant{
                    GoroutineID:   goroutineID,
                    GoroutineName: goroutine.Name,
                    Resources:     append(goroutine.Holding, goroutine.WaitingFor...),
                    State:         goroutine.State,
                    StackTrace:    goroutine.StackTrace,
                }
                participants = append(participants, participant)
            }
        }
    }
    
    // Calculate confidence based on cycle characteristics
    if len(participants) < 2 {
        confidence *= 0.5 // Low confidence for single participant
    }
    
    // Check if all participants are actually blocked
    blockedCount := 0
    for _, participant := range participants {
        if participant.State == "blocked" || participant.State == "waiting" {
            blockedCount++
        }
    }
    
    if blockedCount == len(participants) {
        confidence *= 1.0 // High confidence if all blocked
    } else {
        confidence *= 0.7 // Lower confidence if not all blocked
    }
    
    deadlock := DetectedDeadlock{
        ID:               fmt.Sprintf("deadlock_%d", time.Now().UnixNano()),
        Type:             DeadlockCircular,
        ParticipantCount: len(participants),
        Participants:     participants,
        Cycle:            steps,
        DetectedAt:       time.Now(),
        Confidence:       confidence,
        Severity:         dd.calculateSeverity(participants),
        Evidence:         dd.gatherEvidence(participants, steps),
        Resolution:       dd.generateResolutionSuggestions(participants, steps),
    }
    
    return deadlock
}

func (dd *DeadlockDetector) detectHeuristicDeadlocks() []DetectedDeadlock {
    var deadlocks []DetectedDeadlock
    
    // Detect long-waiting goroutines
    threshold := dd.detectionConfig.DeadlockThreshold
    now := time.Now()
    
    for _, goroutine := range dd.goroutines {
        if len(goroutine.WaitingFor) > 0 && now.Sub(goroutine.BlockedSince) > threshold {
            // Check if this looks like a deadlock situation
            confidence := dd.calculateHeuristicConfidence(goroutine)
            
            if confidence >= 0.6 {
                participant := DeadlockParticipant{
                    GoroutineID:   goroutine.ID,
                    GoroutineName: goroutine.Name,
                    Resources:     append(goroutine.Holding, goroutine.WaitingFor...),
                    State:         goroutine.State,
                    StackTrace:    goroutine.StackTrace,
                }
                
                deadlock := DetectedDeadlock{
                    ID:               fmt.Sprintf("heuristic_deadlock_%d", time.Now().UnixNano()),
                    Type:             DeadlockResource,
                    ParticipantCount: 1,
                    Participants:     []DeadlockParticipant{participant},
                    DetectedAt:       now,
                    Confidence:       confidence,
                    Severity:         SeverityMedium,
                    Evidence:         dd.gatherHeuristicEvidence(goroutine),
                }
                
                deadlocks = append(deadlocks, deadlock)
            }
        }
    }
    
    return deadlocks
}

func (dd *DeadlockDetector) calculateHeuristicConfidence(goroutine *GoroutineState) float64 {
    confidence := 0.0
    
    // Time-based confidence
    waitTime := time.Since(goroutine.BlockedSince)
    if waitTime > time.Hour {
        confidence += 0.4
    } else if waitTime > time.Minute*30 {
        confidence += 0.3
    } else if waitTime > time.Minute*10 {
        confidence += 0.2
    } else if waitTime > dd.detectionConfig.DeadlockThreshold {
        confidence += 0.1
    }
    
    // Stack trace analysis
    stackStr := strings.Join(goroutine.StackTrace, " ")
    if strings.Contains(stackStr, "lock") || strings.Contains(stackStr, "sync") {
        confidence += 0.2
    }
    
    if strings.Contains(stackStr, "chan") || strings.Contains(stackStr, "select") {
        confidence += 0.2
    }
    
    // Resource analysis
    if len(goroutine.WaitingFor) > 0 && len(goroutine.Holding) > 0 {
        confidence += 0.2 // Holding resources while waiting
    }
    
    return confidence
}

func (dd *DeadlockDetector) calculateSeverity(participants []DeadlockParticipant) DeadlockSeverity {
    if len(participants) >= 5 {
        return SeverityCritical
    } else if len(participants) >= 3 {
        return SeverityHigh
    } else if len(participants) >= 2 {
        return SeverityMedium
    } else {
        return SeverityLow
    }
}

func (dd *DeadlockDetector) gatherEvidence(participants []DeadlockParticipant, steps []CycleStep) []DeadlockEvidence {
    var evidence []DeadlockEvidence
    
    // Cycle evidence
    evidence = append(evidence, DeadlockEvidence{
        Type:        "cycle_detected",
        Description: fmt.Sprintf("Circular dependency detected with %d steps", len(steps)),
        Timestamp:   time.Now(),
        Data:        steps,
    })
    
    // Participant evidence
    for _, participant := range participants {
        evidence = append(evidence, DeadlockEvidence{
            Type:        "blocked_goroutine",
            Description: fmt.Sprintf("Goroutine %d (%s) is blocked", participant.GoroutineID, participant.GoroutineName),
            Timestamp:   time.Now(),
            Data:        participant,
        })
    }
    
    return evidence
}

func (dd *DeadlockDetector) gatherHeuristicEvidence(goroutine *GoroutineState) []DeadlockEvidence {
    var evidence []DeadlockEvidence
    
    waitTime := time.Since(goroutine.BlockedSince)
    evidence = append(evidence, DeadlockEvidence{
        Type:        "long_wait",
        Description: fmt.Sprintf("Goroutine blocked for %v", waitTime),
        Timestamp:   time.Now(),
        Data:        waitTime,
    })
    
    if len(goroutine.StackTrace) > 0 {
        evidence = append(evidence, DeadlockEvidence{
            Type:        "stack_trace",
            Description: "Stack trace shows blocking operation",
            Timestamp:   time.Now(),
            Data:        goroutine.StackTrace,
        })
    }
    
    return evidence
}

func (dd *DeadlockDetector) generateResolutionSuggestions(participants []DeadlockParticipant, steps []CycleStep) []ResolutionSuggestion {
    var suggestions []ResolutionSuggestion
    
    // Ordering suggestion
    suggestions = append(suggestions, ResolutionSuggestion{
        Strategy:    "lock_ordering",
        Description: "Implement consistent lock ordering to prevent circular dependencies",
        Priority:    1,
        Difficulty:  "Medium",
        Impact:      "High",
    })
    
    // Timeout suggestion
    suggestions = append(suggestions, ResolutionSuggestion{
        Strategy:    "timeouts",
        Description: "Add timeouts to lock acquisition to prevent indefinite blocking",
        Priority:    2,
        Difficulty:  "Easy",
        Impact:      "Medium",
    })
    
    // Context cancellation
    suggestions = append(suggestions, ResolutionSuggestion{
        Strategy:    "context_cancellation",
        Description: "Use context.Context for cancellable operations",
        Priority:    3,
        Difficulty:  "Easy",
        Impact:      "High",
    })
    
    // Resource reduction
    suggestions = append(suggestions, ResolutionSuggestion{
        Strategy:    "reduce_resources",
        Description: "Minimize the number of resources held simultaneously",
        Priority:    4,
        Difficulty:  "Hard",
        Impact:      "High",
    })
    
    return suggestions
}

func (dd *DeadlockDetector) updateDetectionTime(duration time.Duration) {
    // Update average detection time
    totalScans := atomic.LoadInt64(&dd.statistics.TotalScans)
    if totalScans == 1 {
        dd.statistics.AverageDetectionTime = duration
    } else {
        // Exponential moving average
        alpha := 0.1
        current := float64(dd.statistics.AverageDetectionTime)
        new := float64(duration)
        dd.statistics.AverageDetectionTime = time.Duration(alpha*new + (1-alpha)*current)
    }
}

func (dd *DeadlockDetector) notifyCallbacks(deadlock DetectedDeadlock) {
    dd.mu.RLock()
    callbacks := make([]DeadlockCallback, len(dd.alertCallbacks))
    copy(callbacks, dd.alertCallbacks)
    dd.mu.RUnlock()
    
    for _, callback := range callbacks {
        go func(cb DeadlockCallback) {
            defer func() {
                if r := recover(); r != nil {
                    // Prevent callback panics from affecting detector
                }
            }()
            cb(deadlock)
        }(callback)
    }
}

func (dd *DeadlockDetector) GetStatistics() DeadlockStatistics {
    dd.mu.RLock()
    defer dd.mu.RUnlock()
    
    stats := *dd.statistics
    
    // Update graph size
    dd.dependencies.mu.RLock()
    currentGraphSize := len(dd.dependencies.nodes)
    dd.dependencies.mu.RUnlock()
    
    if currentGraphSize > stats.MaxGraphSize {
        stats.MaxGraphSize = currentGraphSize
    }
    
    return stats
}

func (dd *DeadlockDetector) GetDetailedReport() DeadlockReport {
    dd.mu.RLock()
    defer dd.mu.RUnlock()
    
    stats := dd.GetStatistics()
    
    // Get current state
    var activeGoroutines []GoroutineState
    var blockedGoroutines []GoroutineState
    var resourceStates []ResourceState
    
    for _, goroutine := range dd.goroutines {
        if goroutine.State == "blocked" || len(goroutine.WaitingFor) > 0 {
            blockedGoroutines = append(blockedGoroutines, *goroutine)
        } else {
            activeGoroutines = append(activeGoroutines, *goroutine)
        }
    }
    
    for _, resource := range dd.resources {
        state := ResourceState{
            ID:          resource.ID,
            Type:        resource.Type.String(),
            Owner:       resource.Owner,
            WaiterCount: len(resource.Waiters),
            AccessCount: resource.AccessCount,
            LastAccess:  resource.LastAccess,
        }
        resourceStates = append(resourceStates, state)
    }
    
    // Sort by various criteria
    sort.Slice(blockedGoroutines, func(i, j int) bool {
        return blockedGoroutines[i].BlockedSince.Before(blockedGoroutines[j].BlockedSince)
    })
    
    sort.Slice(resourceStates, func(i, j int) bool {
        return resourceStates[i].WaiterCount > resourceStates[j].WaiterCount
    })
    
    return DeadlockReport{
        Statistics:        stats,
        ActiveGoroutines:  activeGoroutines,
        BlockedGoroutines: blockedGoroutines,
        ResourceStates:    resourceStates,
        GraphSize:         len(dd.dependencies.nodes),
        GeneratedAt:       time.Now(),
    }
}

type ResourceState struct {
    ID          string
    Type        string
    Owner       int
    WaiterCount int
    AccessCount int64
    LastAccess  time.Time
}

type DeadlockReport struct {
    Statistics        DeadlockStatistics
    ActiveGoroutines  []GoroutineState
    BlockedGoroutines []GoroutineState
    ResourceStates    []ResourceState
    GraphSize         int
    GeneratedAt       time.Time
}

func (dr DeadlockReport) String() string {
    result := fmt.Sprintf(`Deadlock Detection Report
Generated: %v

=== STATISTICS ===
Total Scans: %d
Detected Deadlocks: %d
False Positives: %d
Prevented Deadlocks: %d
Average Detection Time: %v
Max Graph Size: %d
Last Scan: %v

=== CURRENT STATE ===
Active Goroutines: %d
Blocked Goroutines: %d
Total Resources: %d
Dependency Graph Size: %d`,
        dr.GeneratedAt.Format(time.RFC3339),
        dr.Statistics.TotalScans,
        dr.Statistics.DetectedDeadlocks,
        dr.Statistics.FalsePositives,
        dr.Statistics.PreventedDeadlocks,
        dr.Statistics.AverageDetectionTime,
        dr.Statistics.MaxGraphSize,
        dr.Statistics.LastScanTime.Format(time.RFC3339),
        len(dr.ActiveGoroutines),
        len(dr.BlockedGoroutines),
        len(dr.ResourceStates),
        dr.GraphSize)
    
    if len(dr.BlockedGoroutines) > 0 {
        result += "\n\n=== BLOCKED GOROUTINES ==="
        for i, goroutine := range dr.BlockedGoroutines {
            if i >= 5 {
                result += "\n  ..."
                break
            }
            
            blockedTime := time.Since(goroutine.BlockedSince)
            result += fmt.Sprintf("\n  %d. ID:%d Name:%s Blocked:%v",
                i+1, goroutine.ID, goroutine.Name, blockedTime)
            
            if len(goroutine.WaitingFor) > 0 {
                result += fmt.Sprintf(" (waiting for: %v)", goroutine.WaitingFor)
            }
        }
    }
    
    if len(dr.ResourceStates) > 0 {
        result += "\n\n=== RESOURCE CONTENTION ==="
        for i, resource := range dr.ResourceStates {
            if i >= 5 || resource.WaiterCount == 0 {
                if i >= 5 {
                    result += "\n  ..."
                }
                break
            }
            
            result += fmt.Sprintf("\n  %d. %s (%s) - %d waiters, owner: %d",
                i+1, resource.ID, resource.Type, resource.WaiterCount, resource.Owner)
        }
    }
    
    return result
}

func (dd DetectedDeadlock) String() string {
    return fmt.Sprintf(`DEADLOCK DETECTED
ID: %s
Type: %s
Participants: %d
Confidence: %.1f%%
Severity: %s
Detected: %v

Participants:`,
        dd.ID, dd.Type, dd.ParticipantCount, dd.Confidence*100, dd.Severity, dd.DetectedAt.Format(time.RFC3339))
}

func demonstrateDeadlockDetection() {
    fmt.Println("=== DEADLOCK DETECTION DEMONSTRATION ===")
    
    detector := NewDeadlockDetector()
    detector.RegisterCallback(func(deadlock DetectedDeadlock) {
        fmt.Printf("\n🚨 DEADLOCK ALERT: %s\n", deadlock)
    })
    
    detector.Enable()
    defer detector.Disable()
    
    // Register resources
    detector.RegisterResource("mutex_a", ResourceMutex, nil)
    detector.RegisterResource("mutex_b", ResourceMutex, nil)
    detector.RegisterResource("channel_c", ResourceChannel, map[string]interface{}{"buffer_size": 0})
    
    // Simulate goroutines
    detector.UpdateGoroutineState(1, "goroutine_1", "running", []string{"main.go:10", "sync.(*Mutex).Lock"})
    detector.UpdateGoroutineState(2, "goroutine_2", "running", []string{"main.go:20", "sync.(*Mutex).Lock"})
    
    // Simulate deadlock scenario
    fmt.Println("Simulating classic deadlock scenario...")
    
    // Goroutine 1 acquires mutex_a, waits for mutex_b
    detector.RecordResourceAcquisition(1, "mutex_a")
    detector.RecordResourceWait(1, "mutex_b")
    detector.UpdateGoroutineState(1, "goroutine_1", "blocked", []string{"main.go:15", "sync.(*Mutex).Lock"})
    
    // Goroutine 2 acquires mutex_b, waits for mutex_a
    detector.RecordResourceAcquisition(2, "mutex_b")
    detector.RecordResourceWait(2, "mutex_a")
    detector.UpdateGoroutineState(2, "goroutine_2", "blocked", []string{"main.go:25", "sync.(*Mutex).Lock"})
    
    // Wait for detection
    time.Sleep(time.Second * 2)
    
    // Get detailed report
    report := detector.GetDetailedReport()
    fmt.Printf("\n%s\n", report)
}
```

## Prevention Strategies

### 1. Lock Ordering

```go
// Implement consistent lock ordering
type OrderedMutexes struct {
    mutexes map[string]*sync.Mutex
    order   []string
}

func (om *OrderedMutexes) LockMultiple(names ...string) {
    // Sort names to ensure consistent ordering
    sorted := make([]string, len(names))
    copy(sorted, names)
    sort.Strings(sorted)
    
    // Lock in sorted order
    for _, name := range sorted {
        if mutex, exists := om.mutexes[name]; exists {
            mutex.Lock()
        }
    }
}

func (om *OrderedMutexes) UnlockMultiple(names ...string) {
    // Unlock in reverse order
    sorted := make([]string, len(names))
    copy(sorted, names)
    sort.Sort(sort.Reverse(sort.StringSlice(sorted)))
    
    for _, name := range sorted {
        if mutex, exists := om.mutexes[name]; exists {
            mutex.Unlock()
        }
    }
}
```

### 2. Timeout-based Locking

```go
// Implement timeout-based resource acquisition
type TimeoutMutex struct {
    mu sync.Mutex
    ch chan struct{}
}

func NewTimeoutMutex() *TimeoutMutex {
    return &TimeoutMutex{
        ch: make(chan struct{}, 1),
    }
}

func (tm *TimeoutMutex) LockWithTimeout(timeout time.Duration) error {
    acquired := make(chan struct{})
    
    go func() {
        tm.mu.Lock()
        close(acquired)
    }()
    
    select {
    case <-acquired:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("lock acquisition timeout")
    }
}

func (tm *TimeoutMutex) Unlock() {
    tm.mu.Unlock()
}
```

### 3. Context-aware Synchronization

```go
// Context-aware channel operations
func SafeChannelSend(ctx context.Context, ch chan<- interface{}, value interface{}) error {
    select {
    case ch <- value:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(time.Second * 30):
        return fmt.Errorf("send timeout")
    }
}

func SafeChannelReceive(ctx context.Context, ch <-chan interface{}) (interface{}, error) {
    select {
    case value := <-ch:
        return value, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-time.After(time.Second * 30):
        return nil, fmt.Errorf("receive timeout")
    }
}
```

## Next Steps

- Study [Goroutine Analysis](goroutine-analysis.md) techniques
- Learn [Goroutine Leak Detection](goroutine-leaks.md) methods
- Explore [Mutex Contention](../concurrency-profiling/mutex-contention.md) analysis  
- Master [Lock-Free Programming](../../optimization/concurrency/lock-free.md)

## Summary

Deadlock detection enables building robust concurrent systems by:

1. **Early detection** - Identifying deadlocks before they cause system failure
2. **Root cause analysis** - Understanding why deadlocks occur
3. **Prevention strategies** - Implementing patterns that avoid deadlocks
4. **Runtime monitoring** - Continuously watching for deadlock conditions
5. **Automated resolution** - Providing suggestions for fixing deadlocks

Use these techniques to build deadlock-free Go applications that maintain high availability and performance.
