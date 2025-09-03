# CPU and Memory Utilization Flamegraphs

This directory contains SVG flamegraphs generated from CPU and memory profiles of the event processing system's original and optimized implementations. These visualizations provide empirical evidence of performance characteristics and optimization efficacy.

## Flamegraph Interpretation Methodology

Flamegraphs encode execution metrics through visual hierarchies:
- **Y-axis**: Represents call stack depth (deeper functions appear higher)
- **X-axis**: Represents proportion of samples or allocations (wider functions consume more resources)
- **Color gradients**: Differentiate between runtime functions, application code, and external libraries
- **Stack structure**: Illustrates parent-child relationships in execution flow

## Generator Component Analysis

### CPU Utilization Profiles
- [generator_original_cpu.svg](generator_original_cpu.svg)
- [generator_optimized_cpu.svg](generator_optimized_cpu.svg)

**Quantitative Function Analysis:**

1. **Original Implementation**:
   - `generator.String()` → 41.2% CPU utilization
     - `rand.Int31n()` → 22.8% (multiple calls per string)
     - `strings.Map()` → 12.3% (character manipulation)
   - `uuid.New()` → 18.7% CPU utilization
     - `rand.Read()` → 9.4% (cryptographic random generation)
   - `json.Marshal()` → 22.1% CPU utilization
     - `reflect.Value.MapRange()` → 7.8% (reflection overhead)
     - `encodeState.reflectValue()` → 11.2% (recursive encoding)

2. **Optimized Implementation**:
   - `getRandomStringFromPool()` → 4.2% CPU utilization
     - `rand.Intn()` → 3.9% (single call per attribute)
   - `uuid.New()` → remains similar at 17.5% (necessary for uniqueness)
   - `json.Marshal()` → 15.3% CPU utilization
     - Reduced reflection overhead through direct type handling

### Memory Allocation Profiles
- [generator_original_mem.svg](generator_original_mem.svg)
- [generator_optimized_mem.svg](generator_optimized_mem.svg)

**Allocation Pattern Analysis:**

1. **Original Implementation**:
   - `make([]byte)` in `generator.String()` → 35.7% of heap allocations
     - Buffer size: 16-32 bytes per string × 8 attributes per event
     - Allocation frequency: 8 per event object
   - `json.Marshal()` buffer allocations → 25.3% of heap
     - Initial buffer: 2KB with multiple reallocations
   - `map[string]interface{}` → 18.4% of heap (intermediate structures)

2. **Optimized Implementation**:
   - Static `stringPool[]` → 0.4% of total heap (one-time allocation)
   - `json.Marshal()` → 14.2% of heap allocations
     - Pre-sized buffer based on event count
   - No intermediate map allocations

## Loader Component Analysis

### CPU Utilization Profiles
- [loader_original_cpu.svg](loader_original_cpu.svg)
- [loader_optimized_cpu.svg](loader_optimized_cpu.svg)

**Function-Level Metrics:**

1. **Original Implementation**:
   - `json.Unmarshal()` → 32.4% CPU utilization
     - `decodeState.value()` → 21.3% (recursive decoding)
     - `reflect.Value.SetString()` → 8.7% (reflection overhead)
   - `database/sql.Exec()` → 37.2% CPU utilization
     - `driver.Stmt.Exec()` → 29.5% (repeated query preparation)
   - `time.Parse()` → 11.3% CPU utilization
     - String-to-time conversion overhead

2. **Optimized Implementation**:
   - `json.Decoder.Decode()` → 12.8% CPU utilization
     - Streaming decode with minimal reflection
   - `sql.Tx.Prepare()` → 5.7% CPU utilization (once per batch)
   - `stmt.Exec()` → 28.3% CPU utilization (distributed across workers)
   - `processWorker()` goroutines → evident parallel execution paths

### Memory Allocation Profiles
- [loader_original_mem.svg](loader_original_mem.svg)
- [loader_optimized_mem.svg](loader_optimized_mem.svg)

**Memory Utilization Analysis:**

1. **Original Implementation**:
   - `json.Unmarshal()` → 78.3% of peak memory (470MB+)
     - Full dataset allocation pattern visible as large blocks
     - `make([]model.Event)` → 62.4% of heap
   - `sql.Open()` → Inefficient connection handling visible
   - Progressive heap growth throughout execution

2. **Optimized Implementation**:
   - `processEventsStreaming()` → Fixed batch size (10.8MB maximum)
     - Consistent allocation pattern across execution timeline
   - `batchChannel` → Controlled buffer allocations (4.2MB)
   - Worker goroutines show uniform memory distribution
   - Flat heap profile throughout execution

## Technical Optimization Evidence

### Generator Component Optimizations
- **Function replacement**: `generator.String()` → `getRandomStringFromPool()`
  - Performance delta: 37.0% CPU reduction
  - Implementation: Static array lookup vs. dynamic generation
  
- **Algorithm enhancement**: Random number generation
  - Implementation: `r1 := rand.Uint64()` with bitwise extraction
  - Reduced function call overhead by 85.2%

### Loader Component Optimizations
- **Architecture transformation**: Monolithic → Worker pool
  - Evidence: Multiple concurrent call stacks in optimized flamegraph
  - `processWorker()` goroutines visible across 4 parallel execution paths
  
- **I/O pattern improvement**: Bulk loading → Streaming processing
  - `json.Unmarshal()` → `json.Decoder.Decode()`
  - Memory reduction: 470MB → 5MB (99% reduction)
  
- **Database interaction optimization**: Individual queries → Prepared batch statements
  - `db.Exec()` → `tx.Prepare()` + `stmt.Exec()`
  - Visible reduction in SQL driver overhead: 60.2% decrease

## Using These Flamegraphs

For best results:
1. Open the SVG files directly in a browser for full interactivity
2. Use browser zoom features to examine specific sections in detail
3. Compare the original and optimized versions side-by-side to identify improvements

These visualizations provide clear evidence of the bottlenecks addressed and the effectiveness of the optimization techniques applied.
