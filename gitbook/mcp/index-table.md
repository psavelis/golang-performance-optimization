# MCP Index Table

| Domain | Name (ID) | Primary Input | Key Output | Typical Use Case | Complexity | Maturity |
|--------|-----------|---------------|------------|------------------|------------|----------|
| Observability | prometheus.query_assistant | expression | stats + optimized query | Faster PromQL iteration | Low | Stable |
| Observability | prometheus.alert_explainer | alert_name | deviation + contributors | On-call triage | Med | Beta |
| Observability | pyroscope.profile_explorer | service | top symbols | Hotspot triage | Low | Stable |
| Observability | pyroscope.diff_analyzer | baseline/compare ranges | symbol deltas | Regression analysis | Med | Beta |
| Observability | grafana.dashboard_author | panels spec | dashboard JSON | Standard dashboards | Med | Beta |
| Observability | logs.aggregation_slicer | query | clusters | Incident log reduction | High | Alpha |
| Tracing | tracing.trace_summarizer | trace_id | narrative | Trace triage | Low | Stable |
| Tracing | tracing.span_anomaly_detector | service+op | anomaly flag | Early latency detection | Med | Beta |
| Tracing | tracing.latency_regression_detector | deploy SHAs | delta pct | Post-deploy validation | High | Alpha |
| Tracing | tracing.trace_profile_correlator | trace_id+service | hotspots mapping | Root cause focus | High | Alpha |
| Tracing | tracing.critical_path_extractor | trace_id | segments list | Latency bottleneck | Med | Beta |
| Performance | perf.benchmark_trend_analyzer | history path | trends | Slow regression guard | Med | Beta |
| Performance | perf.regression_gatekeeper | bench reports | decision | CI gating | Low | Stable |
| Performance | perf.flamegraph_diff_explainer | pprof pair | package deltas | Optimization targeting | Med | Beta |
| Performance | perf.allocation_hotspot_classifier | alloc profile | top + recs | Memory optimization | Low | Stable |
| Performance | perf.gc_pressure_advisor | runtime metrics | risk & recs | GC tuning | Low | Beta |
| Performance | perf.concurrency_contention_advisor | mutex/block profiles | hotspots | Throughput unlock | Med | Alpha |
| Performance | perf.io_latency_bottleneck_finder | spanmetrics | bottlenecks | External dep focus | Med | Beta |
| Performance | perf.query_plan_regression_analyzer | plans | regressions | DB performance | High | Alpha |
| Cross-Cutting | cross.slo_error_budget_monitor | metrics | burn + risk | SLO governance | Low | Stable |
| Cross-Cutting | cross.capacity_forecasting_engine | metric | forecast | Capacity planning | Med | Beta |
| Cross-Cutting | cross.cost_efficiency_analyzer | cost+metric | cost KPIs | FinOps alignment | Med | Alpha |
| Cross-Cutting | cross.performance_risk_scoring | signals map | risk score | Prioritization | Low | Alpha |
| Cross-Cutting | cross.golden_signals_aggregator | services list | unified table | Health overview | Low | Beta |
| Cross-Cutting | cross.optimization_recommendation_orchestrator | sources | ranked actions | Backlog curation | High | Alpha |

Legend:
- Complexity: effort & integration depth
- Maturity: Stable (battle tested) / Beta (needs tuning) / Alpha (conceptual or early impl)
