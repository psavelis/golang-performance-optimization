# Bottleneck Analysis

## Methodology

Analysis was conducted using Go's built-in profiler (`pprof`) with CPU and memory profiling. Flame graphs were generated to visualize hotspots and memory allocation patterns.

## Critical Issues Identified

### Generator Bottlenecks

#### 1. String Generation Hotspot (P0 - Critical)
**Location**: `pkg/generator/string.go:14-18`
```go
// BOTTLENECK: String concatenation in tight loop
func RandomString() string {
    var str string
    for i := 0; i <= int(strLen); i++ {
        str = str + string(letterRunes[...]) // Creates new string each iteration
    }
}
```
**Impact**: 15.96% CPU, 21.50MB memory allocations
**Root Cause**: String concatenation creates new string objects each iteration

#### 2. Memory Explosion (P0 - Critical)
**Location**: `cmd/generator/main.go:48-53`
```go
// BOTTLENECK: Storing all events before marshaling
events := []*model.Event{}
for i := 0; i < numEvents; i++ {
    events = append(events, generateEvent()) // 100MB+ in memory
}
json.Marshal(events) // 490MB JSON file
```
**Impact**: 100.75MB peak memory, 48MB in JSON marshaling
**Root Cause**: Entire dataset held in memory simultaneously

#### 3. Excessive Random Generation (P1 - High)
**Location**: `cmd/generator/main.go:76-91`
```go
// BOTTLENECK: ~300 rand() calls per event
EventSource:     rand.Intn(88005553535),    // Individual calls
CallingNumber:   rand.Intn(88005553535),    // No value reuse
Location:        generator.RandomString(),   // Up to 40 more calls
// ... 8 more RandomString() calls
```
**Impact**: 300M+ random calls for 1M events
**Root Cause**: No optimization of random number generation

### Loader Bottlenecks

#### 1. Memory Bomb (P0 - Critical)
**Location**: `cmd/loader/main.go:44-50`
```go
// BOTTLENECK: Entire file loaded into memory
eventRaw, err := os.ReadFile(inputFile)     // 490MB RAM
json.Unmarshal(eventRaw, &events)          // Additional 100MB structures
```
**Impact**: 57.95% of memory allocations, 490MB+ RAM usage
**Root Cause**: Bulk file processing instead of streaming

#### 2. Individual Database Operations (P0 - Critical)
**Location**: `cmd/loader/main.go:67-72`
```go
// BOTTLENECK: 100K individual INSERT statements
for _, e := range events {
    err = load(tx, e)  // Individual network round-trip
}
```
**Impact**: 61.35% CPU in syscalls, 38+ seconds for 100K events
**Root Cause**: No batch processing, 100K database round-trips

#### 3. Hot Path String Formatting (P1 - High)
**Location**: `cmd/loader/main.go:89-91`
```go
// BOTTLENECK: String formatting per event
q = fmt.Sprintf(q, timeToTimestampNoTz(&event.EventDate))
```
**Impact**: String allocation/formatting for every INSERT
**Root Cause**: SQL template processing instead of parameterized queries

#### 4. Timestamp Conversion Overhead (P1 - High)
**Location**: `cmd/loader/main.go:118-120`
```go
// BOTTLENECK: Complex SQL generation
func timeToTimestampNoTz(t *time.Time) string {
    return fmt.Sprintf("to_timestamp(cast(%d as bigint))::date", t.Unix())
}
```
**Impact**: SQL string generation for every timestamp
**Root Cause**: Converting Go time to SQL string vs direct parameter

#### 5. Monolithic Transaction (P1 - High)
**Location**: `cmd/loader/main.go:58-65`
```go
// BOTTLENECK: Single transaction for all events
tx, err := db.Begin()
// ... process 100K events ...
tx.Commit() // All-or-nothing
```
**Impact**: Prevents parallelization, memory pressure
**Root Cause**: No batch processing strategy

## Profiling Data

### CPU Profile (Original Generator)
```
15.96% pkg/generator.RandomString    # String concatenation
 8.51% encoding/json.Marshal         # JSON buffer growth
 2.13% main.generateEvent            # Event creation overhead
```

### CPU Profile (Original Loader) 
```
61.35% runtime.kevent               # Syscalls from individual DB ops
32.69% database/sql.(*Tx).Exec      # Per-event SQL execution
19.23% syscall.syscall              # System call overhead
```

### Memory Profile (Original Generator)
```
100.75MB main.main                  # Event array + JSON marshaling
 21.50MB generator.RandomString     # String concatenation allocations
 48.00MB encoding/json buffer       # JSON marshaling growth
```

### Memory Profile (Original Loader)
```
6.35MB main.main                    # JSON unmarshaling
5.17MB encoding/json.array          # Event array allocation
3.07MB reflect.New                  # Struct instantiation
```

## Priority Matrix

| Issue | Impact | Effort | Expected Gain |
|-------|--------|--------|---------------|
| String Generation | Critical | Low | 15-20% CPU |
| Memory Explosion | Critical | Medium | 98% memory |
| Individual DB Ops | Critical | High | 5-10x performance |
| Hot Path Formatting | High | Low | 10-15% CPU |
| Bulk File Loading | High | Medium | 95% memory |

## Validation Methodology

Each bottleneck was validated through:
1. **Profiling Data** - Quantified CPU/memory impact
2. **Flame Graph Analysis** - Visual confirmation of hotspots  
3. **Benchmark Testing** - Before/after performance measurement
4. **Memory Analysis** - Allocation pattern examination

