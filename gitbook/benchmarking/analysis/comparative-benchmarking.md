# Comparative Benchmarking

Comprehensive guide to comparative benchmarking in Go applications. This guide covers benchmark comparison methodologies, statistical analysis, regression detection, and automated comparison systems for evaluating performance differences across versions, configurations, and implementations.

## Table of Contents

- [Introduction](#introduction)
- [Comparison Framework](#comparison-framework)
- [Statistical Methods](#statistical-methods)
- [Benchmark Design](#benchmark-design)
- [Data Collection](#data-collection)
- [Analysis Techniques](#analysis-techniques)
- [Reporting Systems](#reporting-systems)
- [Automation](#automation)
- [Best Practices](#best-practices)

## Introduction

Comparative benchmarking enables systematic evaluation of performance differences between different versions, configurations, or implementations. This guide provides comprehensive strategies for conducting rigorous comparative benchmarks that produce statistically valid and actionable insights.

### Comparison Framework

```go
package main

import (
    "context"
    "fmt"
    "math"
    "sort"
    "sync"
    "time"
)

// ComparativeBenchmark manages comparative benchmarking operations
type ComparativeBenchmark struct {
    config         BenchmarkConfig
    subjects       map[string]*BenchmarkSubject
    comparisons    map[string]*Comparison
    analyzer       *StatisticalAnalyzer
    reporter       *ComparisonReporter
    storage        ComparisonStorage
    scheduler      *BenchmarkScheduler
    validator      *ResultValidator
    aggregator     *ResultAggregator
    notifier       *ComparisonNotifier
    metrics        *BenchmarkMetrics
    mu             sync.RWMutex
}

// BenchmarkConfig contains comparative benchmark configuration
type BenchmarkConfig struct {
    Name                string
    Description         string
    ComparisonMode      ComparisonMode
    StatisticalMethod   StatisticalMethod
    ConfidenceLevel     float64
    SignificanceLevel   float64
    MinSampleSize       int
    MaxSampleSize       int
    WarmupIterations    int
    MeasurementDuration time.Duration
    CooldownPeriod      time.Duration
    EnvironmentControl  EnvironmentControl
    RandomizationEnabled bool
    BaselineRequired    bool
    AutoValidation      bool
    ReportGeneration    bool
    NotificationEnabled bool
}

// ComparisonMode defines comparison modes
type ComparisonMode int

const (
    PairwiseComparison ComparisonMode = iota
    BaselineComparison
    AllPairsComparison
    TournamentComparison
    HierarchicalComparison
)

// StatisticalMethod defines statistical analysis methods
type StatisticalMethod int

const (
    TTestMethod StatisticalMethod = iota
    MannWhitneyMethod
    WilcoxonMethod
    KruskalWallisMethod
    BootstrapMethod
    PermutationMethod
    BayesianMethod
)

// EnvironmentControl defines environment control levels
type EnvironmentControl int

const (
    NoControl EnvironmentControl = iota
    BasicControl
    StrictControl
    IsolatedControl
)

// BenchmarkSubject represents a subject for comparison
type BenchmarkSubject struct {
    ID           string
    Name         string
    Description  string
    Version      string
    Configuration SubjectConfiguration
    Implementation SubjectImplementation
    Environment  SubjectEnvironment
    Metadata     SubjectMetadata
    Results      []BenchmarkResult
    Statistics   SubjectStatistics
    Status       SubjectStatus
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// SubjectConfiguration contains subject configuration
type SubjectConfiguration struct {
    Parameters   map[string]interface{}
    Flags        map[string]string
    Environment  map[string]string
    Resources    ResourceConfiguration
    Constraints  []Constraint
    Tags         map[string]string
}

// ResourceConfiguration defines resource constraints
type ResourceConfiguration struct {
    MaxMemory    int64
    MaxCPU       float64
    MaxDisk      int64
    MaxNetwork   int64
    MaxDuration  time.Duration
    Affinity     []string
}

// Constraint represents execution constraints
type Constraint struct {
    Type        ConstraintType
    Parameter   string
    Value       interface{}
    Operator    ComparisonOperator
    Enforcement EnforcementLevel
}

// ConstraintType defines constraint types
type ConstraintType int

const (
    ResourceConstraint ConstraintType = iota
    PerformanceConstraint
    EnvironmentConstraint
    TimeConstraint
    QualityConstraint
)

// ComparisonOperator defines comparison operators
type ComparisonOperator int

const (
    Equal ComparisonOperator = iota
    NotEqual
    GreaterThan
    LessThan
    GreaterThanOrEqual
    LessThanOrEqual
    InRange
    OutOfRange
)

// EnforcementLevel defines constraint enforcement levels
type EnforcementLevel int

const (
    Advisory EnforcementLevel = iota
    Warning
    Error
    Blocking
)

// SubjectImplementation contains implementation details
type SubjectImplementation struct {
    Language     string
    Framework    string
    Algorithm    string
    DataStructure string
    Optimization []string
    Source       SourceInfo
    Build        BuildInfo
    Dependencies []Dependency
}

// SourceInfo contains source code information
type SourceInfo struct {
    Repository string
    Branch     string
    Commit     string
    Tag        string
    Path       string
    Files      []string
}

// BuildInfo contains build information
type BuildInfo struct {
    Version     string
    BuildTime   time.Time
    BuildHost   string
    BuildFlags  []string
    Compiler    string
    Linker      string
    Optimization string
}

// Dependency represents a dependency
type Dependency struct {
    Name     string
    Version  string
    Type     DependencyType
    Source   string
    Checksum string
}

// DependencyType defines dependency types
type DependencyType int

const (
    LibraryDependency DependencyType = iota
    FrameworkDependency
    ToolDependency
    RuntimeDependency
)

// SubjectEnvironment describes execution environment
type SubjectEnvironment struct {
    OS           string
    Architecture string
    Platform     string
    Runtime      RuntimeInfo
    Hardware     HardwareInfo
    Network      NetworkInfo
    Storage      StorageInfo
    Container    ContainerInfo
    Cloud        CloudInfo
}

// RuntimeInfo contains runtime information
type RuntimeInfo struct {
    Version      string
    Vendor       string
    Mode         string
    GCSettings   map[string]interface{}
    JITSettings  map[string]interface{}
    MemorySettings map[string]interface{}
    ThreadSettings map[string]interface{}
}

// HardwareInfo contains hardware information
type HardwareInfo struct {
    CPU          CPUInfo
    Memory       MemoryInfo
    Storage      StorageInfo
    Network      NetworkInfo
    Accelerators []AcceleratorInfo
}

// CPUInfo contains CPU information
type CPUInfo struct {
    Model       string
    Vendor      string
    Architecture string
    Cores       int
    Threads     int
    Frequency   float64
    Cache       CacheInfo
    Features    []string
}

// CacheInfo contains cache information
type CacheInfo struct {
    L1Data        int64
    L1Instruction int64
    L2            int64
    L3            int64
    LineSize      int
}

// MemoryInfo contains memory information
type MemoryInfo struct {
    Total     int64
    Available int64
    Type      string
    Speed     int64
    Channels  int
    NUMA      bool
}

// StorageInfo contains storage information
type StorageInfo struct {
    Type       string
    Capacity   int64
    Speed      StorageSpeed
    Interface  string
    FileSystem string
}

// StorageSpeed contains storage speed information
type StorageSpeed struct {
    Sequential SequentialSpeed
    Random     RandomSpeed
    IOPS       IOPSInfo
}

// SequentialSpeed contains sequential speed information
type SequentialSpeed struct {
    Read  float64
    Write float64
    Unit  string
}

// RandomSpeed contains random speed information
type RandomSpeed struct {
    Read  float64
    Write float64
    Unit  string
}

// IOPSInfo contains IOPS information
type IOPSInfo struct {
    Read  int64
    Write int64
    Mixed int64
}

// NetworkInfo contains network information
type NetworkInfo struct {
    Interface string
    Speed     int64
    Latency   time.Duration
    Bandwidth float64
    Protocol  []string
}

// AcceleratorInfo contains accelerator information
type AcceleratorInfo struct {
    Type     string
    Model    string
    Memory   int64
    Cores    int
    Speed    float64
    Features []string
}

// ContainerInfo contains container information
type ContainerInfo struct {
    Runtime     string
    Image       string
    Tag         string
    Resources   ContainerResources
    Limits      ContainerLimits
    Networks    []string
    Volumes     []string
}

// ContainerResources contains container resource allocation
type ContainerResources struct {
    CPURequest    float64
    CPULimit      float64
    MemoryRequest int64
    MemoryLimit   int64
    Storage       int64
}

// ContainerLimits contains container limits
type ContainerLimits struct {
    PidsLimit    int64
    FilesLimit   int64
    ProcessLimit int64
    NetworkLimit NetworkLimits
}

// NetworkLimits contains network limits
type NetworkLimits struct {
    BandwidthIn  int64
    BandwidthOut int64
    ConnectionLimit int64
}

// CloudInfo contains cloud environment information
type CloudInfo struct {
    Provider      string
    Region        string
    Zone          string
    InstanceType  string
    InstanceSize  string
    Pricing       PricingInfo
    Spot          bool
    Reserved      bool
}

// PricingInfo contains pricing information
type PricingInfo struct {
    HourlyRate   float64
    Currency     string
    BillingModel string
    Discounts    []string
}

// SubjectMetadata contains subject metadata
type SubjectMetadata struct {
    Author      string
    Team        string
    Purpose     string
    Category    string
    Priority    int
    Tags        map[string]string
    Labels      map[string]string
    Annotations map[string]interface{}
    Links       []MetadataLink
}

// MetadataLink represents metadata links
type MetadataLink struct {
    Type        LinkType
    URL         string
    Description string
}

// LinkType defines link types
type LinkType int

const (
    DocumentationLink LinkType = iota
    SourceCodeLink
    IssueLink
    PullRequestLink
    DashboardLink
    ReportLink
)

// BenchmarkResult represents benchmark execution results
type BenchmarkResult struct {
    ID          string
    SubjectID   string
    RunID       string
    Timestamp   time.Time
    Duration    time.Duration
    Iterations  int64
    Metrics     map[string]MetricResult
    Environment EnvironmentSnapshot
    Errors      []BenchmarkError
    Warnings    []BenchmarkWarning
    Metadata    ResultMetadata
    Quality     ResultQuality
}

// MetricResult contains metric measurement results
type MetricResult struct {
    Name        string
    Value       float64
    Unit        string
    Type        MetricType
    Aggregation AggregationType
    Samples     []float64
    Statistics  MetricStatistics
    Distribution DistributionInfo
    Outliers    []OutlierInfo
}

// MetricType defines metric types
type MetricType int

const (
    LatencyMetric MetricType = iota
    ThroughputMetric
    MemoryMetric
    CPUMetric
    IOMetric
    NetworkMetric
    CustomMetric
    CompositeMetric
)

// AggregationType defines aggregation types
type AggregationType int

const (
    MeanAggregation AggregationType = iota
    MedianAggregation
    MinAggregation
    MaxAggregation
    SumAggregation
    CountAggregation
    PercentileAggregation
)

// MetricStatistics contains metric statistics
type MetricStatistics struct {
    Count      int64
    Mean       float64
    Median     float64
    StdDev     float64
    Variance   float64
    Min        float64
    Max        float64
    Range      float64
    IQR        float64
    Skewness   float64
    Kurtosis   float64
    Percentiles map[int]float64
}

// DistributionInfo contains distribution information
type DistributionInfo struct {
    Type        DistributionType
    Parameters  map[string]float64
    GoodnessOfFit float64
    Confidence  float64
    Histogram   HistogramData
}

// DistributionType defines distribution types
type DistributionType int

const (
    NormalDistribution DistributionType = iota
    LogNormalDistribution
    ExponentialDistribution
    GammaDistribution
    WeibullDistribution
    UniformDistribution
)

// HistogramData contains histogram data
type HistogramData struct {
    Bins   []float64
    Counts []int64
    Edges  []float64
    Width  float64
}

// OutlierInfo contains outlier information
type OutlierInfo struct {
    Index      int
    Value      float64
    ZScore     float64
    Probability float64
    Method     OutlierMethod
}

// OutlierMethod defines outlier detection methods
type OutlierMethod int

const (
    ZScoreMethod OutlierMethod = iota
    IQRMethod
    ModifiedZScoreMethod
    IsolationForestMethod
)

// EnvironmentSnapshot captures environment state during execution
type EnvironmentSnapshot struct {
    Timestamp    time.Time
    SystemLoad   SystemLoadInfo
    Memory       MemorySnapshot
    CPU          CPUSnapshot
    Storage      StorageSnapshot
    Network      NetworkSnapshot
    Processes    []ProcessInfo
    Temperature  TemperatureInfo
}

// SystemLoadInfo contains system load information
type SystemLoadInfo struct {
    LoadAverage LoadAverageInfo
    RunQueue    int
    Processes   ProcessStats
    Interrupts  int64
    ContextSwitches int64
}

// LoadAverageInfo contains load average information
type LoadAverageInfo struct {
    OneMinute     float64
    FiveMinutes   float64
    FifteenMinutes float64
}

// ProcessStats contains process statistics
type ProcessStats struct {
    Total     int
    Running   int
    Sleeping  int
    Stopped   int
    Zombie    int
}

// MemorySnapshot contains memory state
type MemorySnapshot struct {
    Total       int64
    Available   int64
    Used        int64
    Free        int64
    Cached      int64
    Buffers     int64
    SwapTotal   int64
    SwapUsed    int64
    SwapFree    int64
}

// CPUSnapshot contains CPU state
type CPUSnapshot struct {
    Usage       CPUUsageInfo
    Frequency   CPUFrequencyInfo
    Temperature CPUTemperatureInfo
    Throttling  bool
}

// CPUUsageInfo contains CPU usage information
type CPUUsageInfo struct {
    User      float64
    System    float64
    Idle      float64
    IOWait    float64
    IRQ       float64
    SoftIRQ   float64
    Steal     float64
    Guest     float64
}

// CPUFrequencyInfo contains CPU frequency information
type CPUFrequencyInfo struct {
    Current   []float64
    Min       []float64
    Max       []float64
    Governor  string
}

// CPUTemperatureInfo contains CPU temperature information
type CPUTemperatureInfo struct {
    Package []float64
    Cores   []float64
    Max     float64
    Critical float64
}

// StorageSnapshot contains storage state
type StorageSnapshot struct {
    Usage      StorageUsageInfo
    IO         StorageIOInfo
    Queue      StorageQueueInfo
}

// StorageUsageInfo contains storage usage information
type StorageUsageInfo struct {
    Used      int64
    Available int64
    Total     int64
    Inodes    InodeInfo
}

// InodeInfo contains inode information
type InodeInfo struct {
    Used      int64
    Available int64
    Total     int64
}

// StorageIOInfo contains storage I/O information
type StorageIOInfo struct {
    ReadOps     int64
    WriteOps    int64
    ReadBytes   int64
    WriteBytes  int64
    ReadTime    time.Duration
    WriteTime   time.Duration
    IOTime      time.Duration
}

// StorageQueueInfo contains storage queue information
type StorageQueueInfo struct {
    Depth     int
    WaitTime  time.Duration
    ServiceTime time.Duration
}

// NetworkSnapshot contains network state
type NetworkSnapshot struct {
    Interfaces []NetworkInterfaceInfo
    Connections NetworkConnectionInfo
    Traffic    NetworkTrafficInfo
}

// NetworkInterfaceInfo contains network interface information
type NetworkInterfaceInfo struct {
    Name      string
    State     string
    MTU       int
    Speed     int64
    RxBytes   int64
    TxBytes   int64
    RxPackets int64
    TxPackets int64
    RxErrors  int64
    TxErrors  int64
    RxDrops   int64
    TxDrops   int64
}

// NetworkConnectionInfo contains network connection information
type NetworkConnectionInfo struct {
    TCP   ConnectionStats
    UDP   ConnectionStats
    UNIX  ConnectionStats
}

// ConnectionStats contains connection statistics
type ConnectionStats struct {
    Established int
    Listen      int
    TimeWait    int
    CloseWait   int
    SynSent     int
    SynRecv     int
    FinWait1    int
    FinWait2    int
    Closing     int
    Closed      int
}

// NetworkTrafficInfo contains network traffic information
type NetworkTrafficInfo struct {
    Bandwidth   BandwidthInfo
    Latency     LatencyInfo
    PacketLoss  float64
    Jitter      time.Duration
}

// BandwidthInfo contains bandwidth information
type BandwidthInfo struct {
    Incoming float64
    Outgoing float64
    Peak     float64
    Average  float64
}

// LatencyInfo contains latency information
type LatencyInfo struct {
    Min     time.Duration
    Max     time.Duration
    Mean    time.Duration
    Median  time.Duration
    P95     time.Duration
    P99     time.Duration
}

// ProcessInfo contains process information
type ProcessInfo struct {
    PID      int32
    Name     string
    State    string
    CPU      float64
    Memory   int64
    Threads  int
    FDs      int
    Priority int
    Nice     int
}

// TemperatureInfo contains temperature information
type TemperatureInfo struct {
    CPU     []float64
    GPU     []float64
    System  float64
    Ambient float64
}

// BenchmarkError represents benchmark errors
type BenchmarkError struct {
    Type        ErrorType
    Message     string
    Code        string
    Timestamp   time.Time
    Stack       []string
    Context     map[string]interface{}
    Recoverable bool
}

// ErrorType defines error types
type ErrorType int

const (
    SetupError ErrorType = iota
    ExecutionError
    TeardownError
    ValidationError
    TimeoutError
    ResourceError
    EnvironmentError
)

// BenchmarkWarning represents benchmark warnings
type BenchmarkWarning struct {
    Type      WarningType
    Message   string
    Code      string
    Timestamp time.Time
    Severity  WarningSeverity
    Context   map[string]interface{}
}

// WarningType defines warning types
type WarningType int

const (
    PerformanceWarning WarningType = iota
    EnvironmentWarning
    ResourceWarning
    QualityWarning
    ConfigurationWarning
)

// WarningSeverity defines warning severity levels
type WarningSeverity int

const (
    LowSeverity WarningSeverity = iota
    MediumSeverity
    HighSeverity
)

// ResultMetadata contains result metadata
type ResultMetadata struct {
    Executor     string
    Framework    string
    Runner       string
    Configuration map[string]interface{}
    Tags         map[string]string
    Annotations  map[string]interface{}
    Session      SessionInfo
}

// SessionInfo contains session information
type SessionInfo struct {
    ID        string
    StartTime time.Time
    EndTime   time.Time
    Duration  time.Duration
    Status    SessionStatus
}

// SessionStatus defines session status
type SessionStatus int

const (
    SessionRunning SessionStatus = iota
    SessionCompleted
    SessionFailed
    SessionCancelled
)

// ResultQuality contains result quality metrics
type ResultQuality struct {
    Reliability  float64
    Stability    float64
    Reproducibility float64
    Accuracy     float64
    Precision    float64
    Completeness float64
    OverallScore float64
    Issues       []QualityIssue
}

// QualityIssue represents quality issues
type QualityIssue struct {
    Type        QualityIssueType
    Severity    IssueSeverity
    Description string
    Impact      float64
    Suggestion  string
}

// QualityIssueType defines quality issue types
type QualityIssueType int

const (
    VariabilityIssue QualityIssueType = iota
    OutlierIssue
    BiasIssue
    NoiseIssue
    TrendIssue
    EnvironmentIssue
)

// IssueSeverity defines issue severity levels
type IssueSeverity int

const (
    MinorIssue IssueSeverity = iota
    ModerateIssue
    MajorIssue
    CriticalIssue
)

// SubjectStatistics contains subject statistics
type SubjectStatistics struct {
    ResultCount    int64
    SuccessRate    float64
    AverageRuntime time.Duration
    LastRun        time.Time
    Trends         map[string]TrendInfo
    Quality        QualityMetrics
}

// TrendInfo contains trend information
type TrendInfo struct {
    Direction  TrendDirection
    Slope      float64
    Confidence float64
    StartTime  time.Time
    EndTime    time.Time
    Significance float64
}

// TrendDirection defines trend directions
type TrendDirection int

const (
    NoTrend TrendDirection = iota
    Improving
    Degrading
    Stable
    Volatile
)

// QualityMetrics contains quality metrics
type QualityMetrics struct {
    Consistency  float64
    Stability    float64
    Reliability  float64
    Accuracy     float64
    OverallScore float64
}

// SubjectStatus defines subject status
type SubjectStatus int

const (
    ActiveSubject SubjectStatus = iota
    InactiveSubject
    FailedSubject
    DisabledSubject
)

// Comparison represents a comparison between subjects
type Comparison struct {
    ID          string
    Name        string
    Description string
    Type        ComparisonType
    Subjects    []string
    Baseline    string
    Method      ComparisonMethod
    Configuration ComparisonConfiguration
    Results     *ComparisonResult
    Status      ComparisonStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CompletedAt *time.Time
}

// ComparisonType defines comparison types
type ComparisonType int

const (
    PerformanceComparison ComparisonType = iota
    RegressionComparison
    ScalabilityComparison
    StabilityComparison
    EfficiencyComparison
    QualityComparison
)

// ComparisonMethod defines comparison methods
type ComparisonMethod struct {
    Statistical   StatisticalMethod
    Aggregation   AggregationMethod
    Normalization NormalizationMethod
    Validation    ValidationMethod
    Confidence    float64
    Significance  float64
}

// AggregationMethod defines aggregation methods
type AggregationMethod int

const (
    SimpleAggregation AggregationMethod = iota
    WeightedAggregation
    RobustAggregation
    BootstrapAggregation
)

// NormalizationMethod defines normalization methods
type NormalizationMethod int

const (
    NoNormalization NormalizationMethod = iota
    BaselineNormalization
    ZScoreNormalization
    MinMaxNormalization
    QuantileNormalization
)

// ValidationMethod defines validation methods
type ValidationMethod int

const (
    CrossValidation ValidationMethod = iota
    BootstrapValidation
    PermutationValidation
    HoldoutValidation
)

// ComparisonConfiguration contains comparison configuration
type ComparisonConfiguration struct {
    IncludeMetrics  []string
    ExcludeMetrics  []string
    WeightMetrics   map[string]float64
    ThresholdValues map[string]float64
    FilterCriteria  []FilterCriterion
    GroupingRules   []GroupingRule
    SortingRules    []SortingRule
}

// FilterCriterion defines filtering criteria
type FilterCriterion struct {
    Field     string
    Operator  ComparisonOperator
    Value     interface{}
    Enabled   bool
}

// GroupingRule defines grouping rules
type GroupingRule struct {
    Field     string
    Function  GroupingFunction
    Enabled   bool
}

// GroupingFunction defines grouping functions
type GroupingFunction int

const (
    IdentityGrouping GroupingFunction = iota
    RangeGrouping
    PercentileGrouping
    ClusterGrouping
)

// SortingRule defines sorting rules
type SortingRule struct {
    Field     string
    Direction SortDirection
    Priority  int
    Enabled   bool
}

// SortDirection defines sort directions
type SortDirection int

const (
    AscendingSort SortDirection = iota
    DescendingSort
)

// ComparisonResult contains comparison results
type ComparisonResult struct {
    Overall      OverallComparison
    Pairwise     []PairwiseComparison
    Statistical  StatisticalComparison
    Practical    PracticalComparison
    Visual       VisualComparison
    Summary      ComparisonSummary
    Recommendations []ComparisonRecommendation
    Metadata     ComparisonMetadata
}

// OverallComparison contains overall comparison results
type OverallComparison struct {
    Winner       string
    Ranking      []RankingEntry
    Score        float64
    Confidence   float64
    Significance float64
    Effect       EffectSize
}

// RankingEntry represents ranking information
type RankingEntry struct {
    Subject    string
    Rank       int
    Score      float64
    Confidence float64
    Metrics    map[string]float64
}

// EffectSize represents effect size information
type EffectSize struct {
    CohensD        float64
    GlassD         float64
    HedgesG        float64
    R2             float64
    Eta2           float64
    Omega2         float64
    Interpretation EffectInterpretation
}

// EffectInterpretation defines effect size interpretations
type EffectInterpretation int

const (
    NegligibleEffect EffectInterpretation = iota
    SmallEffect
    MediumEffect
    LargeEffect
    VeryLargeEffect
)

// PairwiseComparison contains pairwise comparison results
type PairwiseComparison struct {
    Subject1     string
    Subject2     string
    Metrics      map[string]MetricComparison
    Overall      PairwiseResult
    Statistical  PairwiseStatistical
    Practical    PairwisePractical
}

// MetricComparison contains metric comparison results
type MetricComparison struct {
    Metric       string
    Value1       float64
    Value2       float64
    Difference   DifferenceInfo
    Statistical  MetricStatistical
    Practical    MetricPractical
}

// DifferenceInfo contains difference information
type DifferenceInfo struct {
    Absolute     float64
    Relative     float64
    Percentage   float64
    Direction    ComparisonDirection
    Magnitude    DifferenceMagnitude
}

// ComparisonDirection defines comparison directions
type ComparisonDirection int

const (
    NoSignificantDifference ComparisonDirection = iota
    Subject1Better
    Subject2Better
    Inconclusive
)

// DifferenceMagnitude defines difference magnitude
type DifferenceMagnitude int

const (
    NegligibleDifference DifferenceMagnitude = iota
    SmallDifference
    MediumDifference
    LargeDifference
    ExtremelyLargeDifference
)

// MetricStatistical contains metric statistical analysis
type MetricStatistical struct {
    Test         StatisticalTest
    Statistic    float64
    PValue       float64
    Significant  bool
    EffectSize   EffectSize
    Confidence   ConfidenceInterval
    Power        float64
}

// StatisticalTest defines statistical tests
type StatisticalTest int

const (
    TTest StatisticalTest = iota
    WelchTTest
    MannWhitneyTest
    WilcoxonTest
    PermutationTest
    BootstrapTest
)

// ConfidenceInterval represents confidence intervals
type ConfidenceInterval struct {
    Lower      float64
    Upper      float64
    Level      float64
    Method     CIMethod
}

// CIMethod defines confidence interval methods
type CIMethod int

const (
    ParametricCI CIMethod = iota
    BootstrapCI
    PermutationCI
    BayesianCI
)

// MetricPractical contains metric practical significance
type MetricPractical struct {
    Threshold    float64
    Significant  bool
    Business     BusinessSignificance
    Technical    TechnicalSignificance
    User         UserSignificance
}

// BusinessSignificance represents business significance
type BusinessSignificance struct {
    Impact       float64
    Cost         float64
    Revenue      float64
    ROI          float64
    Strategic    bool
    Priority     BusinessPriority
}

// BusinessPriority defines business priority
type BusinessPriority int

const (
    LowBusinessPriority BusinessPriority = iota
    MediumBusinessPriority
    HighBusinessPriority
    CriticalBusinessPriority
)

// TechnicalSignificance represents technical significance
type TechnicalSignificance struct {
    Performance  float64
    Scalability  float64
    Reliability  float64
    Maintainability float64
    Security     float64
    Complexity   float64
}

// UserSignificance represents user significance
type UserSignificance struct {
    Experience   float64
    Satisfaction float64
    Productivity float64
    Adoption     float64
    Retention    float64
    Feedback     float64
}

// PairwiseResult contains pairwise result summary
type PairwiseResult struct {
    Winner       string
    Confidence   float64
    Significance float64
    Effect       EffectSize
    Decision     ComparisonDecision
}

// ComparisonDecision defines comparison decisions
type ComparisonDecision int

const (
    Subject1Wins ComparisonDecision = iota
    Subject2Wins
    NoWinner
    Inconclusive
)

// PairwiseStatistical contains pairwise statistical results
type PairwiseStatistical struct {
    OverallTest  StatisticalResult
    MetricTests  map[string]StatisticalResult
    Corrections  []CorrectionInfo
}

// StatisticalResult contains statistical test results
type StatisticalResult struct {
    Test         StatisticalTest
    Statistic    float64
    PValue       float64
    Significant  bool
    EffectSize   EffectSize
    Confidence   ConfidenceInterval
    Assumptions  AssumptionResults
}

// AssumptionResults contains assumption validation results
type AssumptionResults struct {
    Normality       AssumptionResult
    Independence    AssumptionResult
    Homoscedasticity AssumptionResult
    Outliers        AssumptionResult
    Satisfied       bool
    Warnings        []string
}

// AssumptionResult contains individual assumption results
type AssumptionResult struct {
    Test        string
    Statistic   float64
    PValue      float64
    Satisfied   bool
    Method      string
}

// CorrectionInfo contains multiple comparison correction information
type CorrectionInfo struct {
    Method       CorrectionMethod
    OriginalP    float64
    CorrectedP   float64
    Significant  bool
    Procedure    string
}

// CorrectionMethod defines correction methods
type CorrectionMethod int

const (
    BonferroniCorrection CorrectionMethod = iota
    HolmCorrection
    FDRCorrection
    TukeyCorrection
    ScheffeCorrection
)

// PairwisePractical contains pairwise practical significance
type PairwisePractical struct {
    ThresholdsMet   map[string]bool
    BusinessImpact  BusinessSignificance
    TechnicalImpact TechnicalSignificance
    UserImpact      UserSignificance
    Recommendation  PracticalRecommendation
}

// PracticalRecommendation contains practical recommendations
type PracticalRecommendation struct {
    Action      RecommendationAction
    Confidence  float64
    Rationale   string
    Conditions  []string
    Timeline    time.Duration
    Resources   []RequiredResource
}

// RecommendationAction defines recommendation actions
type RecommendationAction int

const (
    AdoptSubject1 RecommendationAction = iota
    AdoptSubject2
    FurtherTesting
    NoAction
    Conditional
)

// RequiredResource represents required resources
type RequiredResource struct {
    Type        ResourceType
    Amount      float64
    Description string
    Critical    bool
}

// ResourceType defines resource types
type ResourceType int

const (
    ComputeResource ResourceType = iota
    MemoryResource
    StorageResource
    NetworkResource
    HumanResource
    TimeResource
    FinancialResource
)

// StatisticalComparison contains statistical comparison results
type StatisticalComparison struct {
    Method       StatisticalMethod
    OverallTest  GlobalTestResult
    PostHoc      []PostHocResult
    PowerAnalysis PowerAnalysisResult
    EffectSizes  []EffectSizeResult
}

// GlobalTestResult contains global test results
type GlobalTestResult struct {
    Test        GlobalTest
    Statistic   float64
    PValue      float64
    Significant bool
    DegreesOfFreedom int
    CriticalValue float64
}

// GlobalTest defines global tests
type GlobalTest int

const (
    ANOVA GlobalTest = iota
    KruskalWallis
    FriedmanTest
    CochranQ
)

// PostHocResult contains post-hoc test results
type PostHocResult struct {
    Test        PostHocTest
    Comparisons []PairwiseComparison
    Adjustments []CorrectionInfo
}

// PostHocTest defines post-hoc tests
type PostHocTest int

const (
    TukeyHSD PostHocTest = iota
    Scheffe
    Bonferroni
    DunnTest
    NemenyiTest
)

// PowerAnalysisResult contains power analysis results
type PowerAnalysisResult struct {
    ObservedPower   float64
    RequiredN       int
    DetectableEffect float64
    TypeIIError     float64
    Recommendations []PowerRecommendation
}

// PowerRecommendation contains power analysis recommendations
type PowerRecommendation struct {
    Type        PowerRecommendationType
    Description string
    Impact      float64
    Feasibility float64
}

// PowerRecommendationType defines power recommendation types
type PowerRecommendationType int

const (
    IncreaseSampleSize PowerRecommendationType = iota
    ReduceVariability
    IncreaseEffectSize
    AdjustAlpha
    ImproveDesign
)

// EffectSizeResult contains effect size results
type EffectSizeResult struct {
    Metric         string
    EffectSize     EffectSize
    Interpretation EffectInterpretation
    Confidence     ConfidenceInterval
}

// PracticalComparison contains practical comparison results
type PracticalComparison struct {
    Thresholds    map[string]ThresholdResult
    Business      BusinessComparison
    Technical     TechnicalComparison
    User          UserComparison
    Overall       PracticalResult
}

// ThresholdResult contains threshold analysis results
type ThresholdResult struct {
    Metric      string
    Threshold   float64
    Met         bool
    Margin      float64
    Confidence  float64
}

// BusinessComparison contains business comparison results
type BusinessComparison struct {
    ROI            ROIComparison
    Cost           CostComparison
    Revenue        RevenueComparison
    Risk           RiskComparison
    Strategic      StrategicComparison
    Overall        BusinessResult
}

// ROIComparison contains ROI comparison
type ROIComparison struct {
    Subject1 float64
    Subject2 float64
    Difference float64
    Threshold float64
    Significant bool
}

// CostComparison contains cost comparison
type CostComparison struct {
    Development    CostBreakdown
    Operation      CostBreakdown
    Maintenance    CostBreakdown
    Total          CostBreakdown
}

// CostBreakdown contains cost breakdown
type CostBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Percentage  float64
    Significant bool
}

// RevenueComparison contains revenue comparison
type RevenueComparison struct {
    Direct     RevenueBreakdown
    Indirect   RevenueBreakdown
    Potential  RevenueBreakdown
    Total      RevenueBreakdown
}

// RevenueBreakdown contains revenue breakdown
type RevenueBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Percentage  float64
    Significant bool
}

// RiskComparison contains risk comparison
type RiskComparison struct {
    Technical    RiskBreakdown
    Business     RiskBreakdown
    Operational  RiskBreakdown
    Security     RiskBreakdown
    Overall      RiskBreakdown
}

// RiskBreakdown contains risk breakdown
type RiskBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Acceptable  bool
    Mitigation  []string
}

// StrategicComparison contains strategic comparison
type StrategicComparison struct {
    Alignment     StrategicBreakdown
    Innovation    StrategicBreakdown
    Competitive   StrategicBreakdown
    Sustainability StrategicBreakdown
    Overall       StrategicBreakdown
}

// StrategicBreakdown contains strategic breakdown
type StrategicBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Strategic   bool
    Impact      float64
}

// BusinessResult contains business result summary
type BusinessResult struct {
    Winner        string
    Confidence    float64
    ROI           float64
    Payback       time.Duration
    Risk          float64
    Strategic     bool
    Recommendation BusinessRecommendation
}

// BusinessRecommendation contains business recommendations
type BusinessRecommendation struct {
    Action      BusinessAction
    Rationale   string
    Timeline    time.Duration
    Investment  float64
    ExpectedROI float64
    Risk        float64
}

// BusinessAction defines business actions
type BusinessAction int

const (
    ProceedWithSubject1 BusinessAction = iota
    ProceedWithSubject2
    RequiresMoreAnalysis
    NotRecommended
    ConditionalApproval
)

// TechnicalComparison contains technical comparison results
type TechnicalComparison struct {
    Performance     TechnicalBreakdown
    Scalability     TechnicalBreakdown
    Reliability     TechnicalBreakdown
    Maintainability TechnicalBreakdown
    Security        TechnicalBreakdown
    Overall         TechnicalResult
}

// TechnicalBreakdown contains technical breakdown
type TechnicalBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Threshold   float64
    Significant bool
    Impact      float64
}

// TechnicalResult contains technical result summary
type TechnicalResult struct {
    Winner         string
    Confidence     float64
    Architecture   ArchitecturalImpact
    Implementation ImplementationImpact
    Operations     OperationalImpact
    Recommendation TechnicalRecommendation
}

// ArchitecturalImpact contains architectural impact
type ArchitecturalImpact struct {
    Complexity   float64
    Coupling     float64
    Cohesion     float64
    Extensibility float64
    Modularity   float64
}

// ImplementationImpact contains implementation impact
type ImplementationImpact struct {
    Effort       float64
    Timeline     time.Duration
    Resources    int
    Complexity   float64
    Risk         float64
}

// OperationalImpact contains operational impact
type OperationalImpact struct {
    Deployment   float64
    Monitoring   float64
    Maintenance  float64
    Support      float64
    Automation   float64
}

// TechnicalRecommendation contains technical recommendations
type TechnicalRecommendation struct {
    Action        TechnicalAction
    Rationale     string
    Considerations []string
    Risks         []string
    Mitigations   []string
}

// TechnicalAction defines technical actions
type TechnicalAction int

const (
    AdoptTechnical1 TechnicalAction = iota
    AdoptTechnical2
    HybridApproach
    RequiresPrototype
    NeedMoreTesting
)

// UserComparison contains user comparison results
type UserComparison struct {
    Experience   UserBreakdown
    Performance  UserBreakdown
    Satisfaction UserBreakdown
    Productivity UserBreakdown
    Overall      UserResult
}

// UserBreakdown contains user breakdown
type UserBreakdown struct {
    Subject1    float64
    Subject2    float64
    Difference  float64
    Threshold   float64
    Significant bool
    Impact      float64
}

// UserResult contains user result summary
type UserResult struct {
    Winner         string
    Confidence     float64
    Experience     float64
    Adoption       float64
    Retention      float64
    Recommendation UserRecommendation
}

// UserRecommendation contains user recommendations
type UserRecommendation struct {
    Action      UserAction
    Rationale   string
    Training    bool
    Support     bool
    Timeline    time.Duration
}

// UserAction defines user actions
type UserAction int

const (
    DeployToUsers1 UserAction = iota
    DeployToUsers2
    GradualRollout
    UserTesting
    RequiresTraining
)

// PracticalResult contains practical result summary
type PracticalResult struct {
    Overall        PracticalDecision
    Confidence     float64
    Business       BusinessResult
    Technical      TechnicalResult
    User           UserResult
    Recommendation FinalRecommendation
}

// PracticalDecision defines practical decisions
type PracticalDecision int

const (
    ClearWinner PracticalDecision = iota
    ConditionalWinner
    NoWinner
    RequiresMoreData
    ContextDependent
)

// FinalRecommendation contains final recommendations
type FinalRecommendation struct {
    Primary     RecommendationAction
    Secondary   []RecommendationAction
    Conditions  []string
    Timeline    time.Duration
    Resources   []RequiredResource
    Risks       []string
    Mitigations []string
    Monitoring  []string
}

// VisualComparison contains visual comparison elements
type VisualComparison struct {
    Charts      []ChartInfo
    Tables      []TableInfo
    Summaries   []SummaryInfo
    Interactive []InteractiveInfo
}

// ChartInfo contains chart information
type ChartInfo struct {
    Type        ChartType
    Title       string
    Description string
    Data        interface{}
    Options     ChartOptions
}

// ChartType defines chart types
type ChartType int

const (
    BarChart ChartType = iota
    LineChart
    ScatterPlot
    BoxPlot
    ViolinPlot
    HeatMap
    RadarChart
    ParallelCoordinates
)

// ChartOptions contains chart options
type ChartOptions struct {
    Width       int
    Height      int
    Colors      []string
    Interactive bool
    Export      bool
    Annotations []string
}

// TableInfo contains table information
type TableInfo struct {
    Type        TableType
    Title       string
    Description string
    Headers     []string
    Rows        [][]interface{}
    Options     TableOptions
}

// TableType defines table types
type TableType int

const (
    SummaryTable TableType = iota
    DetailTable
    ComparisonTable
    StatisticalTable
    RankingTable
)

// TableOptions contains table options
type TableOptions struct {
    Sortable    bool
    Filterable  bool
    Paginated   bool
    Exportable  bool
    Highlighting bool
}

// SummaryInfo contains summary information
type SummaryInfo struct {
    Type        SummaryType
    Title       string
    Content     string
    Highlights  []string
    Insights    []string
}

// SummaryType defines summary types
type SummaryType int

const (
    ExecutiveSummary SummaryType = iota
    TechnicalSummary
    StatisticalSummary
    BusinessSummary
)

// InteractiveInfo contains interactive element information
type InteractiveInfo struct {
    Type        InteractiveType
    Title       string
    Description string
    Configuration interface{}
}

// InteractiveType defines interactive types
type InteractiveType int

const (
    Dashboard InteractiveType = iota
    Explorer
    Simulator
    Calculator
)

// ComparisonSummary contains comparison summary
type ComparisonSummary struct {
    Winner        string
    Confidence    float64
    KeyFindings   []string
    Implications  []string
    Limitations   []string
    NextSteps     []string
    Timeline      time.Duration
}

// ComparisonRecommendation contains comparison recommendations
type ComparisonRecommendation struct {
    Type        RecommendationType
    Priority    RecommendationPriority
    Action      string
    Rationale   string
    Impact      float64
    Effort      float64
    Timeline    time.Duration
    Dependencies []string
}

// RecommendationType defines recommendation types
type RecommendationType int

const (
    ImmediateAction RecommendationType = iota
    ShortTermAction
    LongTermAction
    ConditionalAction
    InvestigationAction
)

// RecommendationPriority defines recommendation priorities
type RecommendationPriority int

const (
    LowPriority RecommendationPriority = iota
    MediumPriority
    HighPriority
    CriticalPriority
)

// ComparisonMetadata contains comparison metadata
type ComparisonMetadata struct {
    Version     string
    Tool        string
    Framework   string
    Analyst     string
    Reviewer    string
    Approved    bool
    Tags        map[string]string
    References  []string
    Attachments []AttachmentInfo
}

// AttachmentInfo contains attachment information
type AttachmentInfo struct {
    Name        string
    Type        string
    Size        int64
    Description string
    URL         string
}

// ComparisonStatus defines comparison status
type ComparisonStatus int

const (
    PendingComparison ComparisonStatus = iota
    RunningComparison
    CompletedComparison
    FailedComparison
    CancelledComparison
)

// NewComparativeBenchmark creates a new comparative benchmark
func NewComparativeBenchmark(config BenchmarkConfig) *ComparativeBenchmark {
    return &ComparativeBenchmark{
        config:      config,
        subjects:    make(map[string]*BenchmarkSubject),
        comparisons: make(map[string]*Comparison),
        analyzer:    NewStatisticalAnalyzer(),
        reporter:    NewComparisonReporter(),
        scheduler:   NewBenchmarkScheduler(),
        validator:   NewResultValidator(),
        aggregator:  NewResultAggregator(),
        notifier:    NewComparisonNotifier(),
        metrics:     &BenchmarkMetrics{},
    }
}

// AddSubject adds a benchmark subject
func (cb *ComparativeBenchmark) AddSubject(subject *BenchmarkSubject) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if subject.ID == "" {
        return fmt.Errorf("subject ID is required")
    }
    
    if _, exists := cb.subjects[subject.ID]; exists {
        return fmt.Errorf("subject %s already exists", subject.ID)
    }
    
    subject.CreatedAt = time.Now()
    subject.Status = ActiveSubject
    cb.subjects[subject.ID] = subject
    
    return nil
}

// RunComparison executes a comparison between subjects
func (cb *ComparativeBenchmark) RunComparison(ctx context.Context, comparison *Comparison) (*ComparisonResult, error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Validate comparison
    if err := cb.validateComparison(comparison); err != nil {
        return nil, fmt.Errorf("comparison validation failed: %w", err)
    }
    
    comparison.Status = RunningComparison
    comparison.UpdatedAt = time.Now()
    
    // Collect benchmark results for all subjects
    subjectResults := make(map[string][]BenchmarkResult)
    for _, subjectID := range comparison.Subjects {
        subject := cb.subjects[subjectID]
        if subject == nil {
            continue
        }
        
        results, err := cb.runBenchmark(ctx, subject)
        if err != nil {
            return nil, fmt.Errorf("benchmark failed for subject %s: %w", subjectID, err)
        }
        
        subjectResults[subjectID] = results
    }
    
    // Perform statistical analysis
    result, err := cb.analyzer.CompareResults(subjectResults, comparison.Method)
    if err != nil {
        return nil, fmt.Errorf("statistical analysis failed: %w", err)
    }
    
    // Validate results
    if cb.config.AutoValidation {
        if err := cb.validator.ValidateResults(result); err != nil {
            return nil, fmt.Errorf("result validation failed: %w", err)
        }
    }
    
    // Generate report
    if cb.config.ReportGeneration {
        report, err := cb.reporter.GenerateReport(result)
        if err != nil {
            return nil, fmt.Errorf("report generation failed: %w", err)
        }
        result.Visual = report.Visual
    }
    
    comparison.Results = result
    comparison.Status = CompletedComparison
    comparison.CompletedAt = &[]time.Time{time.Now()}[0]
    comparison.UpdatedAt = time.Now()
    
    // Send notifications
    if cb.config.NotificationEnabled {
        cb.notifier.NotifyCompletion(comparison)
    }
    
    return result, nil
}

// Helper methods and implementations
func (cb *ComparativeBenchmark) validateComparison(comparison *Comparison) error {
    if len(comparison.Subjects) < 2 {
        return fmt.Errorf("at least 2 subjects required for comparison")
    }
    
    for _, subjectID := range comparison.Subjects {
        if _, exists := cb.subjects[subjectID]; !exists {
            return fmt.Errorf("subject %s not found", subjectID)
        }
    }
    
    return nil
}

func (cb *ComparativeBenchmark) runBenchmark(ctx context.Context, subject *BenchmarkSubject) ([]BenchmarkResult, error) {
    // Simplified benchmark execution
    // In a real implementation, this would execute the actual benchmark
    var results []BenchmarkResult
    
    for i := 0; i < cb.config.MinSampleSize; i++ {
        result := BenchmarkResult{
            ID:        fmt.Sprintf("%s-run-%d", subject.ID, i),
            SubjectID: subject.ID,
            Timestamp: time.Now(),
            Duration:  cb.config.MeasurementDuration,
            Metrics:   make(map[string]MetricResult),
        }
        
        // Simulate metric collection
        result.Metrics["latency"] = MetricResult{
            Name:  "latency",
            Value: float64(i*10 + 100), // Simulated values
            Unit:  "ms",
            Type:  LatencyMetric,
        }
        
        results = append(results, result)
    }
    
    return results, nil
}

// Component types and constructors
type BenchmarkMetrics struct{}
type StatisticalAnalyzer struct{}
type ComparisonReporter struct{}
type ComparisonStorage interface{}
type BenchmarkScheduler struct{}
type ResultValidator struct{}
type ResultAggregator struct{}
type ComparisonNotifier struct{}

func NewStatisticalAnalyzer() *StatisticalAnalyzer { return &StatisticalAnalyzer{} }
func NewComparisonReporter() *ComparisonReporter { return &ComparisonReporter{} }
func NewBenchmarkScheduler() *BenchmarkScheduler { return &BenchmarkScheduler{} }
func NewResultValidator() *ResultValidator { return &ResultValidator{} }
func NewResultAggregator() *ResultAggregator { return &ResultAggregator{} }
func NewComparisonNotifier() *ComparisonNotifier { return &ComparisonNotifier{} }

func (sa *StatisticalAnalyzer) CompareResults(results map[string][]BenchmarkResult, method ComparisonMethod) (*ComparisonResult, error) {
    return &ComparisonResult{}, nil
}
func (rv *ResultValidator) ValidateResults(result *ComparisonResult) error { return nil }
func (cr *ComparisonReporter) GenerateReport(result *ComparisonResult) (*ComparisonReport, error) {
    return &ComparisonReport{}, nil
}
func (cn *ComparisonNotifier) NotifyCompletion(comparison *Comparison) {}

type ComparisonReport struct {
    Visual VisualComparison
}

// Example usage
func ExampleComparativeBenchmark() {
    // Create benchmark configuration
    config := BenchmarkConfig{
        Name:                "Algorithm Comparison",
        Description:         "Compare sorting algorithm performance",
        ComparisonMode:      PairwiseComparison,
        StatisticalMethod:   TTestMethod,
        ConfidenceLevel:     0.95,
        SignificanceLevel:   0.05,
        MinSampleSize:       30,
        MaxSampleSize:       100,
        WarmupIterations:    10,
        MeasurementDuration: time.Second,
        CooldownPeriod:      time.Millisecond * 100,
        EnvironmentControl:  StrictControl,
        RandomizationEnabled: true,
        BaselineRequired:    false,
        AutoValidation:      true,
        ReportGeneration:    true,
        NotificationEnabled: true,
    }
    
    // Create comparative benchmark
    benchmark := NewComparativeBenchmark(config)
    
    // Add subjects
    quickSort := &BenchmarkSubject{
        ID:          "quicksort",
        Name:        "Quick Sort",
        Description: "Quick sort algorithm implementation",
        Version:     "1.0.0",
        Configuration: SubjectConfiguration{
            Parameters: map[string]interface{}{
                "pivot_strategy": "median_of_three",
                "cutoff":        10,
            },
        },
    }
    
    mergeSort := &BenchmarkSubject{
        ID:          "mergesort",
        Name:        "Merge Sort",
        Description: "Merge sort algorithm implementation",
        Version:     "1.0.0",
        Configuration: SubjectConfiguration{
            Parameters: map[string]interface{}{
                "merge_strategy": "in_place",
                "buffer_size":   1024,
            },
        },
    }
    
    // Add subjects to benchmark
    benchmark.AddSubject(quickSort)
    benchmark.AddSubject(mergeSort)
    
    // Create comparison
    comparison := &Comparison{
        ID:          "quicksort-vs-mergesort",
        Name:        "Quick Sort vs Merge Sort",
        Description: "Performance comparison of sorting algorithms",
        Type:        PerformanceComparison,
        Subjects:    []string{"quicksort", "mergesort"},
        Method: ComparisonMethod{
            Statistical:   TTestMethod,
            Aggregation:   SimpleAggregation,
            Normalization: NoNormalization,
            Validation:    CrossValidation,
            Confidence:    0.95,
            Significance:  0.05,
        },
        CreatedAt: time.Now(),
    }
    
    // Run comparison
    ctx := context.Background()
    result, err := benchmark.RunComparison(ctx, comparison)
    if err != nil {
        fmt.Printf("Comparison failed: %v\n", err)
        return
    }
    
    fmt.Println("Comparative Benchmark Results:")
    fmt.Printf("Winner: %s\n", result.Overall.Winner)
    fmt.Printf("Confidence: %.2f\n", result.Overall.Confidence)
    fmt.Printf("Statistical significance: %.4f\n", result.Overall.Significance)
    fmt.Printf("Effect size (Cohen's d): %.3f\n", result.Overall.Effect.CohensD)
    
    fmt.Printf("\nPairwise comparisons: %d\n", len(result.Pairwise))
    for _, pairwise := range result.Pairwise {
        fmt.Printf("  %s vs %s: Winner = %s (p = %.4f)\n",
            pairwise.Subject1,
            pairwise.Subject2,
            pairwise.Overall.Winner,
            pairwise.Statistical.OverallTest.PValue)
    }
    
    if len(result.Recommendations) > 0 {
        fmt.Println("\nRecommendations:")
        for _, rec := range result.Recommendations {
            fmt.Printf("  - %s: %s\n", rec.Type, rec.Action)
        }
    }
}
```

## Statistical Methods

Advanced statistical techniques for rigorous benchmark comparisons.

### Hypothesis Testing

Proper hypothesis testing frameworks for comparing benchmark results.

### Effect Size Analysis

Quantifying the practical significance of performance differences.

### Multiple Comparisons

Handling multiple comparisons with appropriate statistical corrections.

## Benchmark Design

Designing benchmarks specifically for comparative analysis.

### Controlled Experiments

Creating controlled experimental designs for fair comparisons.

### Randomization

Implementing proper randomization to eliminate bias.

### Sample Size Determination

Calculating appropriate sample sizes for statistical power.

## Best Practices

1. **Statistical Rigor**: Apply proper statistical methods with appropriate corrections
2. **Controlled Conditions**: Maintain consistent environmental conditions
3. **Adequate Sampling**: Ensure sufficient sample sizes for reliable results
4. **Effect Size**: Always report both statistical and practical significance
5. **Reproducibility**: Design benchmarks for reproducible results
6. **Documentation**: Document all comparison parameters and decisions
7. **Validation**: Validate results through multiple methods
8. **Interpretation**: Provide clear interpretation of statistical results

## Summary

Comparative benchmarking enables rigorous performance evaluation:

1. **Statistical Analysis**: Robust statistical methods for reliable comparisons
2. **Experimental Design**: Controlled experimental design for fair evaluation
3. **Effect Size Analysis**: Quantification of practical significance
4. **Comprehensive Reporting**: Detailed reporting with visual analysis
5. **Automated Validation**: Automated validation of comparison results
6. **Decision Support**: Clear recommendations based on statistical evidence

These techniques enable organizations to make data-driven decisions about performance optimizations and technology choices.
