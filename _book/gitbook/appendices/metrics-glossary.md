# Performance Metrics Glossary

Comprehensive glossary of performance metrics, terms, and concepts used in Go performance engineering and optimization.

## Core Performance Metrics

### Time-based Metrics

**ns/op (Nanoseconds per Operation)**
- Time taken to execute a single operation
- Lower values indicate better performance
- Primary metric for CPU-bound operations
- Example: `1234 ns/op` means 1.234 microseconds per operation

**Latency**
- Time from request initiation to response completion
- Often measured in percentiles (P50, P95, P99)
- Includes network, processing, and queuing delays
- Critical for user-facing applications

**Throughput**
- Number of operations completed per unit time
- Measured in ops/sec, requests/sec, or bytes/sec
- Higher values indicate better performance
- Often inversely related to latency

**Response Time**
- Total time to process a request
- Includes all processing stages
- Measured from client perspective
- Different from server-side latency

### Memory Metrics

**B/op (Bytes per Operation)**
- Memory allocated per operation
- Lower values indicate better memory efficiency
- Includes both stack and heap allocations
- Example: `1024 B/op` means 1KB allocated per operation

**allocs/op (Allocations per Operation)**
- Number of heap allocations per operation
- Lower values reduce GC pressure
- Zero allocations ideal for hot paths
- Example: `5 allocs/op` means 5 heap allocations per operation

**Heap Size**
- Total memory allocated on heap
- Includes both live and garbage objects
- Measured in bytes (MB, GB)
- Affects GC frequency and pause times

**RSS (Resident Set Size)**
- Physical memory currently used by process
- Includes heap, stack, and system memory
- Operating system perspective
- Can exceed heap size significantly

### Garbage Collection Metrics

**GC Pause Time**
- Duration of stop-the-world GC phases
- Measured in milliseconds or microseconds
- Lower values improve application responsiveness
- Target: <10ms for low-latency applications

**GC Frequency**
- How often garbage collection cycles occur
- Measured in collections per second
- Affected by allocation rate and GOGC setting
- Higher frequency can impact throughput

**GC Overhead**
- Percentage of CPU time spent in garbage collection
- Typically 1-5% for well-tuned applications
- High overhead indicates memory pressure
- Formula: (GC CPU Time / Total CPU Time) × 100

**Mark Assist Time**
- Time spent by mutator goroutines helping GC
- Occurs during concurrent mark phase
- High values indicate allocation pressure
- Counted as part of GC overhead

### Concurrency Metrics

**Goroutine Count**
- Number of active goroutines
- Growing count may indicate goroutine leaks
- Typical range: hundreds to low thousands
- Monitor for unbounded growth

**Block Time**
- Time goroutines spend blocked on synchronization
- Includes mutex, channel, and select operations
- High values indicate contention
- Measured per operation or total

**Mutex Contention**
- Time spent waiting for mutex locks
- Indicates synchronization bottlenecks
- Measured in nanoseconds
- Goal: minimize through design changes

**Channel Operations**
- Send/receive operations per second
- Buffer utilization and blocking frequency
- Coordination overhead measurement
- Optimize for your communication patterns

## Advanced Performance Concepts

### CPU Performance

**CPU Utilization**
- Percentage of available CPU time used
- 100% utilization may indicate CPU bottleneck
- Consider both user and system time
- Multi-core systems: per-core analysis important

**Instructions per Cycle (IPC)**
- CPU efficiency metric
- Higher values indicate better CPU utilization
- Affected by cache misses and branch mispredictions
- Typical range: 0.5-4.0 for modern CPUs

**Cache Hit Rate**
- Percentage of memory accesses served by cache
- L1/L2/L3 cache levels have different characteristics
- Higher rates improve performance significantly
- Goal: >95% for L1, >90% for L2

**Branch Prediction Accuracy**
- Percentage of correctly predicted branches
- Modern CPUs achieve >95% accuracy
- Unpredictable code patterns hurt performance
- Affects pipeline efficiency

### Memory Performance

**Memory Bandwidth**
- Rate of data transfer to/from memory
- Measured in GB/s
- Critical for data-intensive applications
- Limited by hardware specifications

**Memory Allocation Rate**
- Bytes allocated per second
- High rates increase GC pressure
- Monitor for allocation hotspots
- Optimize through pooling and reuse

**Escape Analysis Success Rate**
- Percentage of allocations kept on stack
- Higher rates reduce heap pressure
- Compiler optimization effectiveness
- Measured through escape analysis output

**Memory Fragmentation**
- Unused memory between allocated blocks
- Reduces effective memory utilization
- More common in long-running applications
- Go's GC helps reduce fragmentation

### Network Performance

**Network Latency**
- Time for data to travel between nodes
- Physical distance and routing affect latency
- Measured in milliseconds
- Cannot be optimized below physical limits

**Network Throughput**
- Data transfer rate over network
- Measured in Mbps or Gbps
- Limited by bandwidth and protocol overhead
- Optimize through batching and compression

**Connection Pool Utilization**
- Percentage of pool connections in use
- High utilization may indicate undersized pool
- Low utilization suggests oversized pool
- Monitor for optimal sizing

**Request Queue Depth**
- Number of pending network requests
- High depth indicates bottleneck
- Can cause timeout errors
- Balance with connection limits

## Profiling Metrics

### CPU Profiling

**Flat Time**
- Time spent directly in function
- Excludes time in called functions
- Identifies computational hotspots
- Sum of flat times equals total time

**Cumulative Time**
- Time spent in function and callees
- Includes all downstream function calls
- Identifies expensive call paths
- Can exceed flat time significantly

**Sample Count**
- Number of profiling samples in function
- Higher counts indicate more time spent
- Sampling rate affects precision
- Minimum sample count for statistical significance

### Memory Profiling

**Allocation Sites**
- Code locations where memory is allocated
- Identified by file and line number
- Ranked by total allocation size
- Target optimization efforts here

**Live Objects**
- Objects currently in memory
- Not yet garbage collected
- Different from total allocations
- Indicates memory usage patterns

**Allocation Stack Traces**
- Call chains leading to allocations
- Help identify allocation sources
- Essential for optimization targeting
- May be incomplete due to sampling

### Execution Tracing

**Event Timeline**
- Chronological sequence of runtime events
- Includes goroutine state changes
- Shows concurrent execution patterns
- Visualized in trace viewer

**Goroutine Lifecycle**
- Creation, execution, and termination events
- State transitions (running, blocked, waiting)
- Scheduling decisions and delays
- Coordination patterns

**System Call Tracking**
- Operating system interactions
- I/O operations and their durations
- Blocking vs non-blocking operations
- System resource utilization

## Benchmark Metrics

### Statistical Measures

**Mean (Average)**
- Sum of values divided by count
- Sensitive to outliers
- Most common central tendency measure
- May not represent typical performance

**Median (P50)**
- Middle value in sorted dataset
- Less sensitive to outliers
- Better represents typical performance
- 50th percentile measurement

**Standard Deviation**
- Measure of variability
- Low values indicate consistent performance
- High values suggest performance variability
- Important for performance guarantees

**Coefficient of Variation**
- Standard deviation divided by mean
- Normalized measure of variability
- Useful for comparing different metrics
- Values <0.1 indicate good consistency

### Percentile Metrics

**P95 (95th Percentile)**
- 95% of measurements below this value
- Common SLA metric
- Excludes worst 5% of performance
- Balance between coverage and outliers

**P99 (99th Percentile)**
- 99% of measurements below this value
- Captures near-worst-case performance
- Important for critical applications
- More sensitive to outliers than P95

**P99.9 (99.9th Percentile)**
- Captures worst-case scenarios
- Critical for high-availability systems
- Small number of samples at this level
- Difficult to optimize cost-effectively

## Performance Testing Terminology

### Load Testing

**Concurrent Users**
- Number of simultaneous active users
- Different from total registered users
- Affects resource utilization
- Key sizing parameter

**Request Rate**
- Requests per second (RPS)
- Primary load testing metric
- Must match expected production load
- Consider peak and average rates

**Think Time**
- Delay between user requests
- Simulates real user behavior
- Affects total system load
- Balance realism with test intensity

### Performance Boundaries

**Saturation Point**
- Load level where performance degrades
- Response time increases significantly
- Throughput plateaus or decreases
- Critical for capacity planning

**Breaking Point**
- Load level where system fails
- Error rates increase substantially
- System becomes unstable
- Maximum sustainable load

**Knee Point**
- Load level where efficiency drops
- Linear performance relationship breaks
- Early warning of saturation
- Optimal operating point

## Optimization Terminology

### Algorithmic Complexity

**Big O Notation**
- Asymptotic upper bound on growth rate
- O(1), O(log n), O(n), O(n²), etc.
- Describes scalability characteristics
- Independent of constant factors

**Time Complexity**
- How runtime scales with input size
- Critical for algorithm selection
- Measured in Big O notation
- Consider both average and worst case

**Space Complexity**
- How memory usage scales with input
- Important for memory-constrained systems
- Includes both auxiliary and input space
- Trade-off with time complexity

### Code Optimization

**Hot Path**
- Frequently executed code sections
- Prime candidates for optimization
- Identified through profiling
- Small improvements have large impact

**Cold Path**
- Rarely executed code sections
- Lower optimization priority
- Often error handling or edge cases
- May sacrifice performance for clarity

**Critical Path**
- Longest sequence of dependent operations
- Determines minimum execution time
- Cannot be parallelized effectively
- Focus of serial optimization efforts

Understanding these metrics and concepts enables effective performance analysis, optimization targeting, and capacity planning for Go applications across development and production environments.
