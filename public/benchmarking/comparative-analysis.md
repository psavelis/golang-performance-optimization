# Comparative Analysis

Comparative analysis is the cornerstone of effective performance optimization, enabling data-driven decisions about code changes, algorithm choices, and architectural decisions. This chapter covers advanced techniques for comparing benchmark results, statistical validation methods, and automated analysis workflows that ensure reliable performance insights.

## Statistical Foundations of Benchmark Comparison

### Understanding Variance and Significance
Performance measurements inherently contain noise from various sources. Proper statistical analysis separates real performance differences from random variation:

```go
package statistical_analysis

import (
    "fmt"
    "math"
    "sort"
    "testing"
)

// Statistical test for benchmark comparison
type BenchmarkComparator struct {
    confidenceLevel float64
    minSamples     int
    outlierThreshold float64
}

func NewBenchmarkComparator() *BenchmarkComparator {
    return &BenchmarkComparator{
        confidenceLevel: 0.95,  // 95% confidence
        minSamples:     10,     // Minimum samples for statistical validity
        outlierThreshold: 2.0,  // 2 standard deviations
    }
}

type Sample struct {
    Value     float64
    Timestamp time.Time
    Metadata  map[string]interface{}
}

type ComparisonResult struct {
    BaselineMean    float64 `json:"baseline_mean"`
    CurrentMean     float64 `json:"current_mean"`
    PercentChange   float64 `json:"percent_change"`
    PValue          float64 `json:"p_value"`
    Significant     bool    `json:"significant"`
    ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
    EffectSize      float64 `json:"effect_size"`
    Recommendation  string  `json:"recommendation"`
}

type ConfidenceInterval struct {
    Lower float64 `json:"lower"`
    Upper float64 `json:"upper"`
}

func (bc *BenchmarkComparator) Compare(baseline, current []Sample) ComparisonResult {
    // Remove outliers
    baselineClean := bc.removeOutliers(bc.extractValues(baseline))
    currentClean := bc.removeOutliers(bc.extractValues(current))
    
    if len(baselineClean) < bc.minSamples || len(currentClean) < bc.minSamples {
        return ComparisonResult{
            Recommendation: fmt.Sprintf("Insufficient samples for statistical analysis (need >= %d)", bc.minSamples),
        }
    }
    
    baselineMean := bc.calculateMean(baselineClean)
    currentMean := bc.calculateMean(currentClean)
    percentChange := (currentMean - baselineMean) / baselineMean * 100
    
    // Perform t-test
    tStat, pValue := bc.welchTTest(baselineClean, currentClean)
    significant := pValue < (1 - bc.confidenceLevel)
    
    // Calculate confidence interval for the difference
    ci := bc.calculateConfidenceInterval(baselineClean, currentClean)
    
    // Calculate effect size (Cohen's d)
    effectSize := bc.calculateCohenD(baselineClean, currentClean)
    
    result := ComparisonResult{
        BaselineMean:       baselineMean,
        CurrentMean:        currentMean,
        PercentChange:      percentChange,
        PValue:             pValue,
        Significant:        significant,
        ConfidenceInterval: ci,
        EffectSize:         effectSize,
        Recommendation:     bc.generateRecommendation(percentChange, significant, effectSize),
    }
    
    return result
}

func (bc *BenchmarkComparator) extractValues(samples []Sample) []float64 {
    values := make([]float64, len(samples))
    for i, sample := range samples {
        values[i] = sample.Value
    }
    return values
}

func (bc *BenchmarkComparator) removeOutliers(values []float64) []float64 {
    if len(values) < 3 {
        return values
    }
    
    mean := bc.calculateMean(values)
    stdDev := bc.calculateStdDev(values)
    threshold := bc.outlierThreshold * stdDev
    
    var clean []float64
    for _, value := range values {
        if math.Abs(value-mean) <= threshold {
            clean = append(clean, value)
        }
    }
    
    return clean
}

func (bc *BenchmarkComparator) calculateMean(values []float64) float64 {
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

func (bc *BenchmarkComparator) calculateStdDev(values []float64) float64 {
    mean := bc.calculateMean(values)
    variance := 0.0
    
    for _, v := range values {
        variance += math.Pow(v-mean, 2)
    }
    
    variance /= float64(len(values) - 1) // Sample standard deviation
    return math.Sqrt(variance)
}

// Welch's t-test for unequal variances
func (bc *BenchmarkComparator) welchTTest(sample1, sample2 []float64) (float64, float64) {
    n1, n2 := float64(len(sample1)), float64(len(sample2))
    mean1, mean2 := bc.calculateMean(sample1), bc.calculateMean(sample2)
    var1, var2 := bc.calculateVariance(sample1), bc.calculateVariance(sample2)
    
    // Calculate t-statistic
    se := math.Sqrt(var1/n1 + var2/n2)
    tStat := (mean1 - mean2) / se
    
    // Calculate degrees of freedom (Welch-Satterthwaite equation)
    df := math.Pow(var1/n1+var2/n2, 2) / (math.Pow(var1/n1, 2)/(n1-1) + math.Pow(var2/n2, 2)/(n2-1))
    
    // Calculate p-value (simplified approximation)
    pValue := bc.approximatePValue(math.Abs(tStat), df)
    
    return tStat, pValue
}

func (bc *BenchmarkComparator) calculateVariance(values []float64) float64 {
    mean := bc.calculateMean(values)
    variance := 0.0
    
    for _, v := range values {
        variance += math.Pow(v-mean, 2)
    }
    
    return variance / float64(len(values)-1)
}

func (bc *BenchmarkComparator) approximatePValue(tStat, df float64) float64 {
    // Simplified p-value approximation using normal distribution
    // For production use, implement proper t-distribution CDF
    z := tStat / math.Sqrt(1 + tStat*tStat/df)
    return 2 * (1 - bc.normalCDF(math.Abs(z)))
}

func (bc *BenchmarkComparator) normalCDF(x float64) float64 {
    // Approximation of standard normal CDF
    return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

func (bc *BenchmarkComparator) calculateConfidenceInterval(sample1, sample2 []float64) ConfidenceInterval {
    n1, n2 := float64(len(sample1)), float64(len(sample2))
    mean1, mean2 := bc.calculateMean(sample1), bc.calculateMean(sample2)
    var1, var2 := bc.calculateVariance(sample1), bc.calculateVariance(sample2)
    
    meanDiff := mean1 - mean2
    se := math.Sqrt(var1/n1 + var2/n2)
    
    // Critical value for 95% confidence (approximation)
    criticalValue := 1.96
    margin := criticalValue * se
    
    return ConfidenceInterval{
        Lower: meanDiff - margin,
        Upper: meanDiff + margin,
    }
}

func (bc *BenchmarkComparator) calculateCohenD(sample1, sample2 []float64) float64 {
    mean1, mean2 := bc.calculateMean(sample1), bc.calculateMean(sample2)
    var1, var2 := bc.calculateVariance(sample1), bc.calculateVariance(sample2)
    n1, n2 := float64(len(sample1)), float64(len(sample2))
    
    // Pooled standard deviation
    pooledVar := ((n1-1)*var1 + (n2-1)*var2) / (n1 + n2 - 2)
    pooledSD := math.Sqrt(pooledVar)
    
    return (mean1 - mean2) / pooledSD
}

func (bc *BenchmarkComparator) generateRecommendation(percentChange float64, significant bool, effectSize float64) string {
    if !significant {
        return "No statistically significant difference detected. More data may be needed."
    }
    
    magnitude := "small"
    if math.Abs(effectSize) > 0.5 {
        magnitude = "medium"
    }
    if math.Abs(effectSize) > 0.8 {
        magnitude = "large"
    }
    
    direction := "improvement"
    if percentChange > 0 {
        direction = "regression"
    }
    
    return fmt.Sprintf("Statistically significant %s detected (%.2f%% change, %s effect size)", 
        direction, math.Abs(percentChange), magnitude)
}

// Example benchmark comparison
func BenchmarkStringBuildingComparison(b *testing.B) {
    comparator := NewBenchmarkComparator()
    
    // Collect baseline samples
    var baselineSamples []Sample
    for i := 0; i < 20; i++ {
        start := time.Now()
        result := stringConcatenation()
        duration := time.Since(start)
        baselineSamples = append(baselineSamples, Sample{
            Value:     float64(duration.Nanoseconds()),
            Timestamp: time.Now(),
        })
        _ = result
    }
    
    // Collect current implementation samples
    var currentSamples []Sample
    for i := 0; i < 20; i++ {
        start := time.Now()
        result := stringBuilderApproach()
        duration := time.Since(start)
        currentSamples = append(currentSamples, Sample{
            Value:     float64(duration.Nanoseconds()),
            Timestamp: time.Now(),
        })
        _ = result
    }
    
    // Perform comparison
    comparison := comparator.Compare(baselineSamples, currentSamples)
    
    b.Logf("Comparison Result:")
    b.Logf("  Baseline Mean: %.2f ns", comparison.BaselineMean)
    b.Logf("  Current Mean: %.2f ns", comparison.CurrentMean)
    b.Logf("  Change: %.2f%%", comparison.PercentChange)
    b.Logf("  P-Value: %.4f", comparison.PValue)
    b.Logf("  Significant: %v", comparison.Significant)
    b.Logf("  Effect Size: %.3f", comparison.EffectSize)
    b.Logf("  Recommendation: %s", comparison.Recommendation)
    
    // Run the actual benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := stringBuilderApproach()
        _ = result
    }
}

func stringConcatenation() string {
    result := ""
    for i := 0; i < 100; i++ {
        result += fmt.Sprintf("item_%d ", i)
    }
    return result
}

func stringBuilderApproach() string {
    var builder strings.Builder
    builder.Grow(1000) // Pre-allocate
    for i := 0; i < 100; i++ {
        builder.WriteString(fmt.Sprintf("item_%d ", i))
    }
    return builder.String()
}
```

## Advanced Comparative Techniques

### Multi-dimensional Comparison
Compare benchmarks across multiple metrics simultaneously:

```go
package multidimensional

import (
    "encoding/json"
    "fmt"
    "math"
    "sort"
)

type BenchmarkMetrics struct {
    Name            string  `json:"name"`
    NsPerOp         float64 `json:"ns_per_op"`
    BytesPerOp      float64 `json:"bytes_per_op"`
    AllocsPerOp     float64 `json:"allocs_per_op"`
    ThroughputOps   float64 `json:"throughput_ops"`
    MemoryEfficiency float64 `json:"memory_efficiency"`
    CPUEfficiency   float64 `json:"cpu_efficiency"`
}

type MultiDimensionalComparison struct {
    WeightedScore   float64                    `json:"weighted_score"`
    MetricScores    map[string]float64         `json:"metric_scores"`
    NormalizedData  map[string]float64         `json:"normalized_data"`
    Recommendation  string                     `json:"recommendation"`
    TradeoffAnalysis TradeoffAnalysis          `json:"tradeoff_analysis"`
}

type TradeoffAnalysis struct {
    TimeVsMemory    string `json:"time_vs_memory"`
    AllocationRate  string `json:"allocation_rate"`
    OverallBalance  string `json:"overall_balance"`
}

type MetricWeights struct {
    Time        float64 `json:"time"`
    Memory      float64 `json:"memory"`
    Allocations float64 `json:"allocations"`
    Throughput  float64 `json:"throughput"`
}

type MultiDimensionalAnalyzer struct {
    weights MetricWeights
    baseline BenchmarkMetrics
}

func NewMultiDimensionalAnalyzer(baseline BenchmarkMetrics, weights MetricWeights) *MultiDimensionalAnalyzer {
    // Normalize weights
    total := weights.Time + weights.Memory + weights.Allocations + weights.Throughput
    weights.Time /= total
    weights.Memory /= total
    weights.Allocations /= total
    weights.Throughput /= total
    
    return &MultiDimensionalAnalyzer{
        weights:  weights,
        baseline: baseline,
    }
}

func (mda *MultiDimensionalAnalyzer) Compare(candidate BenchmarkMetrics) MultiDimensionalComparison {
    // Normalize metrics relative to baseline
    normalized := mda.normalizeMetrics(candidate)
    
    // Calculate individual metric scores
    scores := map[string]float64{
        "time":        mda.calculateTimeScore(normalized["time"]),
        "memory":      mda.calculateMemoryScore(normalized["memory"]),
        "allocations": mda.calculateAllocationScore(normalized["allocations"]),
        "throughput":  mda.calculateThroughputScore(normalized["throughput"]),
    }
    
    // Calculate weighted overall score
    weightedScore := scores["time"]*mda.weights.Time +
                    scores["memory"]*mda.weights.Memory +
                    scores["allocations"]*mda.weights.Allocations +
                    scores["throughput"]*mda.weights.Throughput
    
    // Analyze tradeoffs
    tradeoffs := mda.analyzeTradeoffs(candidate, normalized, scores)
    
    return MultiDimensionalComparison{
        WeightedScore:    weightedScore,
        MetricScores:     scores,
        NormalizedData:   normalized,
        Recommendation:   mda.generateRecommendation(weightedScore, scores),
        TradeoffAnalysis: tradeoffs,
    }
}

func (mda *MultiDimensionalAnalyzer) normalizeMetrics(candidate BenchmarkMetrics) map[string]float64 {
    return map[string]float64{
        "time":        candidate.NsPerOp / mda.baseline.NsPerOp,
        "memory":      candidate.BytesPerOp / mda.baseline.BytesPerOp,
        "allocations": candidate.AllocsPerOp / mda.baseline.AllocsPerOp,
        "throughput":  candidate.ThroughputOps / mda.baseline.ThroughputOps,
    }
}

func (mda *MultiDimensionalAnalyzer) calculateTimeScore(normalizedTime float64) float64 {
    // Lower time is better, so invert and scale
    if normalizedTime <= 0 {
        return 100
    }
    return math.Max(0, 100*(2-normalizedTime))
}

func (mda *MultiDimensionalAnalyzer) calculateMemoryScore(normalizedMemory float64) float64 {
    // Lower memory usage is better
    if normalizedMemory <= 0 {
        return 100
    }
    return math.Max(0, 100*(2-normalizedMemory))
}

func (mda *MultiDimensionalAnalyzer) calculateAllocationScore(normalizedAllocs float64) float64 {
    // Fewer allocations are better
    if normalizedAllocs <= 0 {
        return 100
    }
    return math.Max(0, 100*(2-normalizedAllocs))
}

func (mda *MultiDimensionalAnalyzer) calculateThroughputScore(normalizedThroughput float64) float64 {
    // Higher throughput is better
    return math.Min(100, normalizedThroughput*100)
}

func (mda *MultiDimensionalAnalyzer) analyzeTradeoffs(candidate BenchmarkMetrics, normalized map[string]float64, scores map[string]float64) TradeoffAnalysis {
    timeChange := (normalized["time"] - 1) * 100
    memoryChange := (normalized["memory"] - 1) * 100
    
    var timeVsMemory string
    if timeChange < -10 && memoryChange > 10 {
        timeVsMemory = "Faster execution at cost of higher memory usage"
    } else if timeChange > 10 && memoryChange < -10 {
        timeVsMemory = "Lower memory usage at cost of slower execution"
    } else if timeChange < -10 && memoryChange < -10 {
        timeVsMemory = "Optimal: both faster and more memory efficient"
    } else if timeChange > 10 && memoryChange > 10 {
        timeVsMemory = "Suboptimal: both slower and more memory intensive"
    } else {
        timeVsMemory = "Balanced performance characteristics"
    }
    
    var allocationRate string
    allocChange := (normalized["allocations"] - 1) * 100
    if allocChange < -50 {
        allocationRate = "Significantly reduced allocations (excellent for GC pressure)"
    } else if allocChange < -20 {
        allocationRate = "Moderately reduced allocations"
    } else if allocChange > 20 {
        allocationRate = "Increased allocations (may impact GC performance)"
    } else {
        allocationRate = "Similar allocation patterns"
    }
    
    var overallBalance string
    if scores["time"] > 80 && scores["memory"] > 80 {
        overallBalance = "Excellent overall performance"
    } else if scores["time"] > 60 && scores["memory"] > 60 {
        overallBalance = "Good balanced performance"
    } else if scores["time"] < 40 || scores["memory"] < 40 {
        overallBalance = "Performance concerns detected"
    } else {
        overallBalance = "Acceptable performance with room for improvement"
    }
    
    return TradeoffAnalysis{
        TimeVsMemory:   timeVsMemory,
        AllocationRate: allocationRate,
        OverallBalance: overallBalance,
    }
}

func (mda *MultiDimensionalAnalyzer) generateRecommendation(weightedScore float64, scores map[string]float64) string {
    if weightedScore > 90 {
        return "🚀 EXCELLENT: Significant improvement across all metrics"
    } else if weightedScore > 70 {
        return "✅ GOOD: Overall performance improvement with minor tradeoffs"
    } else if weightedScore > 50 {
        return "⚖️  MIXED: Some improvements offset by regressions - evaluate priorities"
    } else if weightedScore > 30 {
        return "⚠️  CONCERNING: Performance regression detected - reconsider approach"
    } else {
        return "❌ POOR: Significant performance degradation - revert changes"
    }
}

// Benchmark with multi-dimensional analysis
func BenchmarkMultiDimensionalAnalysis(b *testing.B) {
    // Define baseline metrics (from previous benchmark run)
    baseline := BenchmarkMetrics{
        Name:        "Baseline",
        NsPerOp:     1000.0,
        BytesPerOp:  512.0,
        AllocsPerOp: 4.0,
        ThroughputOps: 1000.0,
    }
    
    // Define weights based on application priorities
    weights := MetricWeights{
        Time:        0.4,  // 40% weight on execution time
        Memory:      0.3,  // 30% weight on memory usage
        Allocations: 0.2,  // 20% weight on allocation count
        Throughput:  0.1,  // 10% weight on throughput
    }
    
    analyzer := NewMultiDimensionalAnalyzer(baseline, weights)
    
    // Test different implementations
    implementations := []struct {
        name string
        fn   func() interface{}
    }{
        {"Optimized", optimizedImplementation},
        {"MemoryEfficient", memoryEfficientImplementation},
        {"HighThroughput", highThroughputImplementation},
    }
    
    for _, impl := range implementations {
        b.Run(impl.name, func(b *testing.B) {
            b.ReportAllocs()
            
            // Measure metrics
            start := time.Now()
            var result interface{}
            
            for i := 0; i < b.N; i++ {
                result = impl.fn()
            }
            
            duration := time.Since(start)
            
            // Calculate metrics (simplified for example)
            metrics := BenchmarkMetrics{
                Name:         impl.name,
                NsPerOp:      float64(duration.Nanoseconds()) / float64(b.N),
                BytesPerOp:   float64(testing.AllocsPerRun(func() { impl.fn() })) * 64, // Estimate
                AllocsPerOp:  float64(testing.AllocsPerRun(func() { impl.fn() })),
                ThroughputOps: float64(b.N) / duration.Seconds(),
            }
            
            // Perform multi-dimensional comparison
            comparison := analyzer.Compare(metrics)
            
            b.Logf("Multi-dimensional Analysis for %s:", impl.name)
            b.Logf("  Weighted Score: %.2f", comparison.WeightedScore)
            b.Logf("  Time Score: %.2f", comparison.MetricScores["time"])
            b.Logf("  Memory Score: %.2f", comparison.MetricScores["memory"])
            b.Logf("  Allocation Score: %.2f", comparison.MetricScores["allocations"])
            b.Logf("  Throughput Score: %.2f", comparison.MetricScores["throughput"])
            b.Logf("  Recommendation: %s", comparison.Recommendation)
            b.Logf("  Time vs Memory: %s", comparison.TradeoffAnalysis.TimeVsMemory)
            b.Logf("  Allocation Impact: %s", comparison.TradeoffAnalysis.AllocationRate)
            
            _ = result
        })
    }
}

func optimizedImplementation() interface{} {
    // Simulate optimized implementation
    data := make([]int, 100)
    for i := range data {
        data[i] = i * i
    }
    return data
}

func memoryEfficientImplementation() interface{} {
    // Simulate memory-efficient implementation
    var sum int
    for i := 0; i < 100; i++ {
        sum += i * i
    }
    return sum
}

func highThroughputImplementation() interface{} {
    // Simulate high-throughput implementation
    const size = 50 // Smaller size for higher throughput
    data := make([]int, size)
    for i := range data {
        data[i] = i * i
    }
    return data
}
```

## Regression Detection and Monitoring

### Automated Performance Regression Detection
Implement continuous monitoring to catch performance regressions:

```go
package regression

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "time"
)

type RegressionDetector struct {
    config     RegressionConfig
    repository BenchmarkRepository
    analyzer   TrendAnalyzer
    alerter    RegressionAlerter
}

type RegressionConfig struct {
    WindowSize         int     `json:"window_size"`
    SensitivityLevel   float64 `json:"sensitivity_level"`
    MinSamples         int     `json:"min_samples"`
    AlertThreshold     float64 `json:"alert_threshold"`
    BaselineWindow     int     `json:"baseline_window"`
}

type BenchmarkRepository interface {
    Store(result BenchmarkResult) error
    GetHistory(benchmarkName string, limit int) ([]BenchmarkResult, error)
    GetBaseline(benchmarkName string) (BenchmarkResult, error)
}

type TrendAnalyzer struct {
    windowSize int
    sensitivity float64
}

type RegressionAlert struct {
    BenchmarkName    string    `json:"benchmark_name"`
    AlertType        string    `json:"alert_type"`
    Severity         string    `json:"severity"`
    CurrentValue     float64   `json:"current_value"`
    BaselineValue    float64   `json:"baseline_value"`
    ChangePercent    float64   `json:"change_percent"`
    TrendDirection   string    `json:"trend_direction"`
    Confidence       float64   `json:"confidence"`
    Timestamp        time.Time `json:"timestamp"`
    Context          map[string]interface{} `json:"context"`
}

func NewRegressionDetector(config RegressionConfig) *RegressionDetector {
    return &RegressionDetector{
        config:     config,
        repository: NewFileBasedRepository("./benchmark_history"),
        analyzer:   TrendAnalyzer{
            windowSize:  config.WindowSize,
            sensitivity: config.SensitivityLevel,
        },
        alerter:    NewRegressionAlerter(),
    }
}

func (rd *RegressionDetector) AnalyzeBenchmark(result BenchmarkResult) error {
    // Store the new result
    if err := rd.repository.Store(result); err != nil {
        return fmt.Errorf("storing benchmark result: %w", err)
    }
    
    // Get historical data
    history, err := rd.repository.GetHistory(result.Name, rd.config.WindowSize*2)
    if err != nil {
        return fmt.Errorf("getting benchmark history: %w", err)
    }
    
    if len(history) < rd.config.MinSamples {
        return nil // Not enough data for analysis
    }
    
    // Detect trends and regressions
    alerts := rd.detectRegressions(result, history)
    
    // Send alerts if necessary
    for _, alert := range alerts {
        if err := rd.alerter.SendAlert(alert); err != nil {
            fmt.Printf("Failed to send alert: %v\n", err)
        }
    }
    
    return nil
}

func (rd *RegressionDetector) detectRegressions(current BenchmarkResult, history []BenchmarkResult) []RegressionAlert {
    var alerts []RegressionAlert
    
    // Sort history by timestamp
    sort.Slice(history, func(i, j int) bool {
        return history[i].Timestamp.Before(history[j].Timestamp)
    })
    
    // Detect timing regression
    if alert := rd.detectTimingRegression(current, history); alert != nil {
        alerts = append(alerts, *alert)
    }
    
    // Detect memory regression
    if alert := rd.detectMemoryRegression(current, history); alert != nil {
        alerts = append(alerts, *alert)
    }
    
    // Detect allocation regression
    if alert := rd.detectAllocationRegression(current, history); alert != nil {
        alerts = append(alerts, *alert)
    }
    
    // Detect trend-based regressions
    if alert := rd.detectTrendRegression(current, history); alert != nil {
        alerts = append(alerts, *alert)
    }
    
    return alerts
}

func (rd *RegressionDetector) detectTimingRegression(current BenchmarkResult, history []BenchmarkResult) *RegressionAlert {
    baseline := rd.calculateBaseline(history, func(r BenchmarkResult) float64 { return r.NsPerOp })
    changePercent := (current.NsPerOp - baseline) / baseline * 100
    
    if changePercent > rd.config.AlertThreshold {
        severity := "medium"
        if changePercent > rd.config.AlertThreshold*2 {
            severity = "high"
        }
        if changePercent > rd.config.AlertThreshold*3 {
            severity = "critical"
        }
        
        return &RegressionAlert{
            BenchmarkName:  current.Name,
            AlertType:      "timing_regression",
            Severity:       severity,
            CurrentValue:   current.NsPerOp,
            BaselineValue:  baseline,
            ChangePercent:  changePercent,
            TrendDirection: "increasing",
            Confidence:     rd.calculateConfidence(current, history),
            Timestamp:      time.Now(),
            Context: map[string]interface{}{
                "metric": "execution_time",
                "unit":   "ns/op",
            },
        }
    }
    
    return nil
}

func (rd *RegressionDetector) detectMemoryRegression(current BenchmarkResult, history []BenchmarkResult) *RegressionAlert {
    if current.BytesPerOp == 0 {
        return nil // No memory data
    }
    
    baseline := rd.calculateBaseline(history, func(r BenchmarkResult) float64 { return float64(r.BytesPerOp) })
    if baseline == 0 {
        return nil
    }
    
    changePercent := (float64(current.BytesPerOp) - baseline) / baseline * 100
    
    if changePercent > rd.config.AlertThreshold*0.5 { // More sensitive for memory
        return &RegressionAlert{
            BenchmarkName:  current.Name,
            AlertType:      "memory_regression",
            Severity:       rd.calculateSeverity(changePercent, rd.config.AlertThreshold*0.5),
            CurrentValue:   float64(current.BytesPerOp),
            BaselineValue:  baseline,
            ChangePercent:  changePercent,
            TrendDirection: "increasing",
            Confidence:     rd.calculateConfidence(current, history),
            Timestamp:      time.Now(),
            Context: map[string]interface{}{
                "metric": "memory_usage",
                "unit":   "bytes/op",
            },
        }
    }
    
    return nil
}

func (rd *RegressionDetector) detectAllocationRegression(current BenchmarkResult, history []BenchmarkResult) *RegressionAlert {
    if current.AllocsPerOp == 0 {
        return nil
    }
    
    baseline := rd.calculateBaseline(history, func(r BenchmarkResult) float64 { return float64(r.AllocsPerOp) })
    if baseline == 0 {
        return nil
    }
    
    changePercent := (float64(current.AllocsPerOp) - baseline) / baseline * 100
    
    if changePercent > rd.config.AlertThreshold*0.3 { // Very sensitive for allocations
        return &RegressionAlert{
            BenchmarkName:  current.Name,
            AlertType:      "allocation_regression",
            Severity:       rd.calculateSeverity(changePercent, rd.config.AlertThreshold*0.3),
            CurrentValue:   float64(current.AllocsPerOp),
            BaselineValue:  baseline,
            ChangePercent:  changePercent,
            TrendDirection: "increasing",
            Confidence:     rd.calculateConfidence(current, history),
            Timestamp:      time.Now(),
            Context: map[string]interface{}{
                "metric": "allocations",
                "unit":   "allocs/op",
            },
        }
    }
    
    return nil
}

func (rd *RegressionDetector) detectTrendRegression(current BenchmarkResult, history []BenchmarkResult) *RegressionAlert {
    if len(history) < rd.config.WindowSize {
        return nil
    }
    
    // Analyze recent trend
    recentHistory := history[len(history)-rd.config.WindowSize:]
    trend := rd.analyzer.CalculateTrend(recentHistory)
    
    if trend.Slope > rd.config.SensitivityLevel && trend.Confidence > 0.8 {
        return &RegressionAlert{
            BenchmarkName:  current.Name,
            AlertType:      "trend_regression",
            Severity:       rd.calculateTrendSeverity(trend.Slope),
            CurrentValue:   current.NsPerOp,
            BaselineValue:  recentHistory[0].NsPerOp,
            ChangePercent:  trend.Slope * 100,
            TrendDirection: "increasing",
            Confidence:     trend.Confidence,
            Timestamp:      time.Now(),
            Context: map[string]interface{}{
                "metric":     "execution_time_trend",
                "trend_type": "linear_regression",
                "window":     rd.config.WindowSize,
            },
        }
    }
    
    return nil
}

func (rd *RegressionDetector) calculateBaseline(history []BenchmarkResult, extractor func(BenchmarkResult) float64) float64 {
    if len(history) == 0 {
        return 0
    }
    
    // Use the median of recent stable values
    baselineWindow := rd.config.BaselineWindow
    if baselineWindow > len(history) {
        baselineWindow = len(history)
    }
    
    values := make([]float64, baselineWindow)
    for i := 0; i < baselineWindow; i++ {
        values[i] = extractor(history[len(history)-baselineWindow+i])
    }
    
    sort.Float64s(values)
    
    // Return median
    if len(values)%2 == 0 {
        return (values[len(values)/2-1] + values[len(values)/2]) / 2
    }
    return values[len(values)/2]
}

func (rd *RegressionDetector) calculateSeverity(changePercent, threshold float64) string {
    if changePercent > threshold*3 {
        return "critical"
    } else if changePercent > threshold*2 {
        return "high"
    } else if changePercent > threshold {
        return "medium"
    }
    return "low"
}

func (rd *RegressionDetector) calculateTrendSeverity(slope float64) string {
    if slope > 0.1 {
        return "critical"
    } else if slope > 0.05 {
        return "high"
    } else if slope > 0.02 {
        return "medium"
    }
    return "low"
}

func (rd *RegressionDetector) calculateConfidence(current BenchmarkResult, history []BenchmarkResult) float64 {
    if len(history) < 3 {
        return 0.5
    }
    
    // Calculate variance in recent measurements
    recent := history[len(history)-min(5, len(history)):]
    values := make([]float64, len(recent))
    for i, r := range recent {
        values[i] = r.NsPerOp
    }
    
    variance := calculateVariance(values)
    cv := math.Sqrt(variance) / calculateMean(values) // Coefficient of variation
    
    // Lower variance = higher confidence
    confidence := math.Max(0.1, 1.0-cv)
    return math.Min(0.99, confidence)
}

type TrendResult struct {
    Slope      float64 `json:"slope"`
    Intercept  float64 `json:"intercept"`
    RSquared   float64 `json:"r_squared"`
    Confidence float64 `json:"confidence"`
}

func (ta *TrendAnalyzer) CalculateTrend(history []BenchmarkResult) TrendResult {
    if len(history) < 2 {
        return TrendResult{}
    }
    
    // Convert to x,y pairs (time index, value)
    n := float64(len(history))
    var sumX, sumY, sumXY, sumXX float64
    
    for i, result := range history {
        x := float64(i)
        y := result.NsPerOp
        
        sumX += x
        sumY += y
        sumXY += x * y
        sumXX += x * x
    }
    
    // Calculate linear regression
    slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
    intercept := (sumY - slope*sumX) / n
    
    // Calculate R-squared
    var ssRes, ssTot float64
    meanY := sumY / n
    
    for i, result := range history {
        x := float64(i)
        y := result.NsPerOp
        predicted := slope*x + intercept
        
        ssRes += math.Pow(y-predicted, 2)
        ssTot += math.Pow(y-meanY, 2)
    }
    
    rSquared := 1 - (ssRes / ssTot)
    confidence := math.Max(0, rSquared) // Use R-squared as confidence measure
    
    return TrendResult{
        Slope:      slope,
        Intercept:  intercept,
        RSquared:   rSquared,
        Confidence: confidence,
    }
}

// File-based benchmark repository
type FileBasedRepository struct {
    basePath string
}

func NewFileBasedRepository(basePath string) *FileBasedRepository {
    os.MkdirAll(basePath, 0755)
    return &FileBasedRepository{basePath: basePath}
}

func (fbr *FileBasedRepository) Store(result BenchmarkResult) error {
    filename := filepath.Join(fbr.basePath, sanitizeFilename(result.Name)+".jsonl")
    
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    data, err := json.Marshal(result)
    if err != nil {
        return err
    }
    
    _, err = file.Write(append(data, '\n'))
    return err
}

func (fbr *FileBasedRepository) GetHistory(benchmarkName string, limit int) ([]BenchmarkResult, error) {
    filename := filepath.Join(fbr.basePath, sanitizeFilename(benchmarkName)+".jsonl")
    
    data, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }
    
    lines := strings.Split(string(data), "\n")
    var results []BenchmarkResult
    
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        
        var result BenchmarkResult
        if err := json.Unmarshal([]byte(line), &result); err != nil {
            continue // Skip malformed lines
        }
        
        results = append(results, result)
    }
    
    // Return the most recent results
    if len(results) > limit {
        results = results[len(results)-limit:]
    }
    
    return results, nil
}

func (fbr *FileBasedRepository) GetBaseline(benchmarkName string) (BenchmarkResult, error) {
    history, err := fbr.GetHistory(benchmarkName, 50)
    if err != nil || len(history) == 0 {
        return BenchmarkResult{}, err
    }
    
    // Return the median of recent results as baseline
    sort.Slice(history, func(i, j int) bool {
        return history[i].NsPerOp < history[j].NsPerOp
    })
    
    return history[len(history)/2], nil
}

func sanitizeFilename(name string) string {
    // Replace invalid filename characters
    invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
    result := name
    for _, char := range invalid {
        result = strings.ReplaceAll(result, char, "_")
    }
    return result
}

type RegressionAlerter struct {
    webhookURL string
    slackToken string
}

func NewRegressionAlerter() RegressionAlerter {
    return RegressionAlerter{
        webhookURL: os.Getenv("PERFORMANCE_WEBHOOK_URL"),
        slackToken: os.Getenv("SLACK_TOKEN"),
    }
}

func (ra RegressionAlerter) SendAlert(alert RegressionAlert) error {
    message := ra.formatAlert(alert)
    
    // Send to console (always)
    fmt.Printf("🚨 PERFORMANCE ALERT: %s\n", message)
    
    // Send to webhook if configured
    if ra.webhookURL != "" {
        return ra.sendWebhook(alert)
    }
    
    // Send to Slack if configured
    if ra.slackToken != "" {
        return ra.sendSlack(alert)
    }
    
    return nil
}

func (ra RegressionAlerter) formatAlert(alert RegressionAlert) string {
    emoji := "⚠️"
    switch alert.Severity {
    case "critical":
        emoji = "🔥"
    case "high":
        emoji = "🚨"
    case "medium":
        emoji = "⚠️"
    case "low":
        emoji = "ℹ️"
    }
    
    return fmt.Sprintf("%s %s: %s benchmark regression detected - %.2f%% change (%.2f → %.2f %s)",
        emoji, strings.ToUpper(alert.Severity), alert.BenchmarkName,
        alert.ChangePercent, alert.BaselineValue, alert.CurrentValue,
        alert.Context["unit"])
}

func (ra RegressionAlerter) sendWebhook(alert RegressionAlert) error {
    // Implementation for webhook notification
    // This would send HTTP POST to configured webhook URL
    return nil
}

func (ra RegressionAlerter) sendSlack(alert RegressionAlert) error {
    // Implementation for Slack notification
    // This would use Slack API to send message
    return nil
}

// Helper functions
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func calculateMean(values []float64) float64 {
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

func calculateVariance(values []float64) float64 {
    mean := calculateMean(values)
    variance := 0.0
    
    for _, v := range values {
        variance += math.Pow(v-mean, 2)
    }
    
    return variance / float64(len(values)-1)
}
```

## Benchstat Integration and Advanced Analysis

### Using benchstat for Statistical Comparison
Leverage Go's benchstat tool for rigorous statistical analysis:

```bash
# Compare benchmark results statistically
go test -bench=. -count=10 ./baseline > baseline.txt
go test -bench=. -count=10 ./optimized > optimized.txt
benchstat baseline.txt optimized.txt

# Example output:
# name                     old time/op    new time/op    delta
# StringBuilding-8           1.23µs ± 2%    0.89µs ± 3%  -27.64%  (p=0.000 n=10+10)
# 
# name                     old alloc/op   new alloc/op   delta
# StringBuilding-8            512B ± 0%      256B ± 0%  -50.00%  (p=0.000 n=10+10)
```

### Custom benchstat Integration
Integrate benchstat programmatically for automated analysis:

```go
package benchstat_integration

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
    "regexp"
    "strconv"
)

type BenchstatResult struct {
    Name           string  `json:"name"`
    OldValue       float64 `json:"old_value"`
    NewValue       float64 `json:"new_value"`
    Delta          float64 `json:"delta"`
    PValue         float64 `json:"p_value"`
    Significant    bool    `json:"significant"`
    Metric         string  `json:"metric"`
    Unit           string  `json:"unit"`
}

type BenchstatAnalyzer struct {
    benchstatPath string
}

func NewBenchstatAnalyzer() *BenchstatAnalyzer {
    return &BenchstatAnalyzer{
        benchstatPath: "benchstat", // Assumes benchstat is in PATH
    }
}

func (ba *BenchstatAnalyzer) Compare(baselineFile, currentFile string) ([]BenchstatResult, error) {
    cmd := exec.Command(ba.benchstatPath, baselineFile, currentFile)
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("running benchstat: %w", err)
    }
    
    return ba.parseBenchstatOutput(string(output))
}

func (ba *BenchstatAnalyzer) parseBenchstatOutput(output string) ([]BenchstatResult, error) {
    var results []BenchstatResult
    
    lines := strings.Split(output, "\n")
    var currentMetric string
    
    // Regex patterns for different types of benchstat output
    headerPattern := regexp.MustCompile(`^name\s+old\s+(\w+)\s+new\s+(\w+)\s+delta`)
    dataPattern := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+([+-]?\d+\.\d+%)\s+\(p=(\d+\.\d+)\s+n=\d+\+\d+\)`)
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        
        // Check for metric header
        if matches := headerPattern.FindStringSubmatch(line); matches != nil {
            currentMetric = matches[1] // Extract metric type (time, alloc, etc.)
            continue
        }
        
        // Parse data line
        if matches := dataPattern.FindStringSubmatch(line); matches != nil {
            result := BenchstatResult{
                Name:   matches[1],
                Metric: currentMetric,
            }
            
            // Parse old value
            if oldVal, err := ba.parseValue(matches[2]); err == nil {
                result.OldValue = oldVal
            }
            
            // Parse new value
            if newVal, err := ba.parseValue(matches[3]); err == nil {
                result.NewValue = newVal
            }
            
            // Parse delta percentage
            deltaStr := strings.TrimSuffix(matches[4], "%")
            if delta, err := strconv.ParseFloat(deltaStr, 64); err == nil {
                result.Delta = delta
            }
            
            // Parse p-value
            if pValue, err := strconv.ParseFloat(matches[5], 64); err == nil {
                result.PValue = pValue
                result.Significant = pValue < 0.05
            }
            
            // Extract unit from old value
            result.Unit = ba.extractUnit(matches[2])
            
            results = append(results, result)
        }
    }
    
    return results, nil
}

func (ba *BenchstatAnalyzer) parseValue(valueStr string) (float64, error) {
    // Remove unit suffix and parse number
    re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Z/µ]+)?`)
    matches := re.FindStringSubmatch(valueStr)
    
    if len(matches) < 2 {
        return 0, fmt.Errorf("invalid value format: %s", valueStr)
    }
    
    value, err := strconv.ParseFloat(matches[1], 64)
    if err != nil {
        return 0, err
    }
    
    // Convert based on unit
    if len(matches) > 2 {
        unit := matches[2]
        switch unit {
        case "µs":
            value *= 1000 // Convert to nanoseconds
        case "ms":
            value *= 1000000 // Convert to nanoseconds
        case "s":
            value *= 1000000000 // Convert to nanoseconds
        case "KB":
            value *= 1024 // Convert to bytes
        case "MB":
            value *= 1024 * 1024 // Convert to bytes
        case "GB":
            value *= 1024 * 1024 * 1024 // Convert to bytes
        }
    }
    
    return value, nil
}

func (ba *BenchstatAnalyzer) extractUnit(valueStr string) string {
    re := regexp.MustCompile(`\s*([a-zA-Z/µ]+)$`)
    matches := re.FindStringSubmatch(valueStr)
    
    if len(matches) > 1 {
        return matches[1]
    }
    
    return ""
}

func (ba *BenchstatAnalyzer) GenerateReport(results []BenchstatResult) string {
    var report strings.Builder
    
    report.WriteString("Benchstat Analysis Report\n")
    report.WriteString("========================\n\n")
    
    // Group by metric type
    metricGroups := make(map[string][]BenchstatResult)
    for _, result := range results {
        metricGroups[result.Metric] = append(metricGroups[result.Metric], result)
    }
    
    for metric, results := range metricGroups {
        report.WriteString(fmt.Sprintf("%s Comparison:\n", strings.Title(metric)))
        report.WriteString("-------------------\n")
        
        var improvements, regressions, neutral []BenchstatResult
        
        for _, result := range results {
            if result.Significant {
                if result.Delta < 0 {
                    improvements = append(improvements, result)
                } else {
                    regressions = append(regressions, result)
                }
            } else {
                neutral = append(neutral, result)
            }
        }
        
        if len(improvements) > 0 {
            report.WriteString("✅ Significant Improvements:\n")
            for _, result := range improvements {
                report.WriteString(fmt.Sprintf("  %s: %.2f%% faster\n", result.Name, -result.Delta))
            }
            report.WriteString("\n")
        }
        
        if len(regressions) > 0 {
            report.WriteString("❌ Significant Regressions:\n")
            for _, result := range regressions {
                report.WriteString(fmt.Sprintf("  %s: %.2f%% slower\n", result.Name, result.Delta))
            }
            report.WriteString("\n")
        }
        
        if len(neutral) > 0 {
            report.WriteString("➡️  No Significant Change:\n")
            for _, result := range neutral {
                report.WriteString(fmt.Sprintf("  %s: %.2f%% change (not significant)\n", result.Name, result.Delta))
            }
            report.WriteString("\n")
        }
    }
    
    // Overall summary
    totalTests := len(results)
    significantChanges := 0
    totalImprovement := 0.0
    totalRegression := 0.0
    
    for _, result := range results {
        if result.Significant {
            significantChanges++
            if result.Delta < 0 {
                totalImprovement += -result.Delta
            } else {
                totalRegression += result.Delta
            }
        }
    }
    
    report.WriteString("Summary:\n")
    report.WriteString("--------\n")
    report.WriteString(fmt.Sprintf("Total benchmarks: %d\n", totalTests))
    report.WriteString(fmt.Sprintf("Significant changes: %d (%.1f%%)\n", 
        significantChanges, float64(significantChanges)/float64(totalTests)*100))
    
    if totalImprovement > 0 {
        report.WriteString(fmt.Sprintf("Average improvement: %.2f%%\n", totalImprovement/float64(totalTests)))
    }
    
    if totalRegression > 0 {
        report.WriteString(fmt.Sprintf("Average regression: %.2f%%\n", totalRegression/float64(totalTests)))
    }
    
    return report.String()
}

// Example benchmark with benchstat integration
func BenchmarkWithBenchstatAnalysis(b *testing.B) {
    analyzer := NewBenchstatAnalyzer()
    
    // Run baseline benchmarks and save to file
    baselineCmd := exec.Command("go", "test", "-bench=StringBuilding", "-count=10")
    baselineOutput, err := baselineCmd.Output()
    if err != nil {
        b.Fatalf("Running baseline benchmarks: %v", err)
    }
    
    // Save baseline to file
    baselineFile := "baseline_results.txt"
    if err := os.WriteFile(baselineFile, baselineOutput, 0644); err != nil {
        b.Fatalf("Saving baseline results: %v", err)
    }
    defer os.Remove(baselineFile)
    
    // Run current benchmarks and save to file
    currentCmd := exec.Command("go", "test", "-bench=StringBuildingOptimized", "-count=10")
    currentOutput, err := currentCmd.Output()
    if err != nil {
        b.Fatalf("Running current benchmarks: %v", err)
    }
    
    currentFile := "current_results.txt"
    if err := os.WriteFile(currentFile, currentOutput, 0644); err != nil {
        b.Fatalf("Saving current results: %v", err)
    }
    defer os.Remove(currentFile)
    
    // Compare using benchstat
    results, err := analyzer.Compare(baselineFile, currentFile)
    if err != nil {
        b.Fatalf("Benchstat comparison failed: %v", err)
    }
    
    // Generate and log report
    report := analyzer.GenerateReport(results)
    b.Log(report)
    
    // The actual benchmark runs here
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := stringBuildingOptimized()
        _ = result
    }
}

func stringBuildingOptimized() string {
    var builder strings.Builder
    builder.Grow(1000)
    for i := 0; i < 100; i++ {
        builder.WriteString(fmt.Sprintf("item_%d ", i))
    }
    return builder.String()
}
```

Comparative analysis is the foundation of evidence-based performance optimization. By implementing rigorous statistical methods, automated regression detection, and comprehensive reporting, you can make confident decisions about code changes and ensure consistent performance improvements over time.

## Key Takeaways

1. **Use proper statistical methods** to distinguish real changes from noise
2. **Implement automated regression detection** to catch performance issues early
3. **Consider multiple metrics simultaneously** for comprehensive evaluation
4. **Leverage benchstat** for rigorous statistical comparison
5. **Monitor trends over time** to identify gradual performance degradation
6. **Generate actionable reports** that guide optimization decisions
7. **Set up continuous monitoring** to maintain performance standards
8. **Use confidence intervals and effect sizes** to understand the magnitude of changes

Effective comparative analysis transforms benchmark data into actionable insights, enabling data-driven performance optimization decisions with confidence and precision.
