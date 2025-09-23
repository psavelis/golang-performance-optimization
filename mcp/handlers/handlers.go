package handlers

import (
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "regexp"
  "sort"
  "strings"
  "time"
)

// Generic request interface type for unstructured dispatch after schema validation.
// Each handler receives raw JSON bytes plus a now() injection for testability.

type Handler func(input json.RawMessage, now time.Time) (any, error)

// Registry maps capability id -> handler.
var Registry = map[string]Handler{}

func init() {
  // Observability
  Registry["prometheus.query_assistant"] = prometheusQueryAssistant
  Registry["prometheus.alert_explainer"] = unsupported("Alert explainer prototype")
  Registry["pyroscope.profile_explorer"] = pyroscopeProfileExplorer
  Registry["pyroscope.diff_analyzer"] = unsupported("Diff analyzer not yet implemented")
  Registry["grafana.dashboard_author"] = unsupported("Dashboard authoring generator placeholder")

  // Tracing
  Registry["tracing.trace_summarizer"] = tracingTraceSummarizer
  Registry["tracing.trace_profile_correlator"] = unsupported("Trace/profile correlation runtime not wired here")
  Registry["tracing.span_anomaly_detector"] = unsupported("Stat anomaly model TBD")
  Registry["tracing.latency_regression_detector"] = unsupported("Latency regression diff placeholder")
  Registry["tracing.critical_path_extractor"] = unsupported("Critical path extraction not yet implemented")

  // Performance
  Registry["perf.allocation_hotspot_classifier"] = unsupported("Allocation hotspot classification placeholder")
  Registry["perf.flamegraph_diff_explainer"] = unsupported("Flamegraph diff explanation placeholder")
  Registry["perf.gc_pressure_advisor"] = unsupported("GC advisor logic TBD")
  Registry["perf.concurrency_contention_advisor"] = unsupported("Contention advisor placeholder")
  Registry["perf.io_latency_bottleneck_finder"] = unsupported("IO bottleneck finder placeholder")
  Registry["perf.query_plan_regression_analyzer"] = unsupported("Query plan diff analyzer placeholder")
  Registry["perf.benchmark_trend_analyzer"] = unsupported("Benchmark trend analysis placeholder")
  Registry["perf.regression_gatekeeper"] = perfRegressionGatekeeper

  // Cross-cutting
  Registry["cross.golden_signals_aggregator"] = unsupported("Golden signals aggregator placeholder")
  Registry["cross.slo_error_budget_monitor"] = unsupported("SLO monitor placeholder")
  Registry["cross.capacity_forecasting_engine"] = unsupported("Capacity forecasting placeholder")
  Registry["cross.cost_efficiency_analyzer"] = unsupported("Cost efficiency analyzer placeholder")
  Registry["cross.performance_risk_scoring"] = unsupported("Performance risk scoring placeholder")
  Registry["cross.optimization_recommendation_orchestrator"] = unsupported("Optimization orchestrator placeholder")
}

// --- Implemented Handlers --------------------------------------------------

// prometheusQueryAssistant performs basic heuristics: detect rate() misuse, duplicate label matchers,
// naive cardinality estimates based on curly brace selectors, and returns a mock optimized expression.
func prometheusQueryAssistant(input json.RawMessage, now time.Time) (any, error) {
  var req struct {
    Expression string `json:"expression"`
    RangeMinutes int `json:"range_minutes"`
    StepSeconds int `json:"step_seconds"`
    Expand []string `json:"expand"`
    Optimize bool `json:"optimize"`
  }
  if err := json.Unmarshal(input, &req); err != nil { return nil, err }
  expr := req.Expression
  diagnostics := []string{}
  if strings.Count(expr, "rate(") > 1 && strings.Contains(expr, "avg(") {
    diagnostics = append(diagnostics, "Multiple nested rate() calls inside aggregation may be redundant")
  }
  if strings.Contains(expr, "{" ) && strings.Contains(expr, "}") {
    selector := expr[strings.Index(expr,"{")+1:strings.Index(expr,"}")]
    parts := strings.Split(selector, ",")
    seen := map[string]int{}
    for _, p := range parts { if i:=strings.Index(p,"="); i>0 { k:=strings.TrimSpace(p[:i]); seen[k]++ } }
    for k,c := range seen { if c>1 { diagnostics = append(diagnostics, fmt.Sprintf("Label %s specified %d times", k,c)) } }
  }
  // Cardinality guess: count of commas +1 in first selector * expansions length
  cardinalityBefore := 0
  if i:=strings.Index(expr,"{"); i>=0 { if j:=strings.Index(expr,"}"); j>i { cardinalityBefore = 1 + strings.Count(expr[i:j], ",") } }
  cardinalityAfter := cardinalityBefore
  optimized := expr
  if req.Optimize && !strings.Contains(expr, "sum by") && strings.Contains(expr, "rate(") {
    optimized = "sum by (job) (" + expr + ")"
    cardinalityAfter = int(math.Ceil(float64(cardinalityBefore)/2.0))
    diagnostics = append(diagnostics, "Wrapped in sum by (job) to reduce series fan-out")
  }
  // expansions (mock) simply replicate expression with replaced label if requested
  var expansions []map[string]any
  for _, e := range req.Expand { expansions = append(expansions, map[string]any{"kind": e, "query": expr + " # expand:" + e }) }
  if len(expansions) == 0 { expansions = []map[string]any{} }
  out := map[string]any{
    "original": expr,
    "optimized": optimized,
    "series_count": cardinalityAfter,
    "samples_total": 0,
    "stats": map[string]any{"min": 0, "max": 0, "p95": 0, "avg": 0},
    "cardinality": map[string]any{"before": cardinalityBefore, "after": cardinalityAfter},
    "diagnostics": diagnostics,
    "expansions": expansions,
    "error": nil,
  }
  return withLinks(out, "gitbook/production/metrics/prometheus-examples.md"), nil
}

// pyroscopeProfileExplorer returns the top symbols (mock heuristics) based on input parameters.
func pyroscopeProfileExplorer(input json.RawMessage, now time.Time) (any, error) {
  var req struct {
    Service string `json:"service"`
    ProfileType string `json:"profile_type"`
    RangeMinutes int `json:"range_minutes"`
    GroupBy []string `json:"group_by"`
    MinFlatPct float64 `json:"min_flat_pct"`
    MaxSymbols int `json:"max_symbols"`
    TagFilters map[string]string `json:"tag_filters"`
  }
  if err := json.Unmarshal(input, &req); err != nil { return nil, err }
  // Generate deterministic placeholder symbols
  symbols := []map[string]any{}
  baseFns := []string{"main.loop","service.handler","runtime.mallocgc","db.exec","cache.get","http.client"}
  for i, fn := range baseFns {
    flat := 0.5 + float64(i)*0.3
    if flat < req.MinFlatPct { continue }
    symbols = append(symbols, map[string]any{"fn": fn, "flat_pct": flat, "cum_pct": flat + 0.2})
    if len(symbols) >= req.MaxSymbols { break }
  }
  recommendation := "Investigate runtime.mallocgc if flat_pct > 1% persistently"
  out := map[string]any{
    "service": req.Service,
    "profile_type": req.ProfileType,
    "window": fmt.Sprintf("last_%dm", req.RangeMinutes),
    "symbols": symbols,
    "total_samples": 0,
    "filters": req.TagFilters,
    "recommendation": recommendation,
    "error": nil,
  }
  return withLinks(out, "gitbook/production/continuous-profiling/pyroscope-examples.md"), nil
}

// tracingTraceSummarizer: simple extraction of service/span patterns from a list of trace JSON objects.
func tracingTraceSummarizer(input json.RawMessage, now time.Time) (any, error) {
  var req struct { 
    Traces []json.RawMessage `json:"traces"`
    Max int `json:"max"`
  }
  if err := json.Unmarshal(input, &req); err != nil { return nil, err }
  type span struct { Name string `json:"name"`; Duration int64 `json:"duration_ms"`; Service string `json:"service"` }
  svcLatency := map[string][]int64{}
  reSvc := regexp.MustCompile(`service":"([a-zA-Z0-9_.-]+)`) // naive fallback if spans omitted
  for _, tr := range req.Traces {
    // Extract spans array if present
    var container struct { Spans []span `json:"spans"` }
    if json.Unmarshal(tr, &container) == nil && len(container.Spans) > 0 {
      for _, sp := range container.Spans { svcLatency[sp.Service] = append(svcLatency[sp.Service], sp.Duration) }
    } else {
      // fallback: regex service occurrences
      matches := reSvc.FindAllStringSubmatch(string(tr), -1)
      for _, m := range matches { svcLatency[m[1]] = append(svcLatency[m[1]], 0) }
    }
  }
  type agg struct { Service string; Count int; P95 int64 }
  aggs := []agg{}
  for svc, arr := range svcLatency {
    sort.Slice(arr, func(i,j int) bool { return arr[i]<arr[j] })
    p95 := int64(0)
    if len(arr)>0 { p95 = arr[int(float64(len(arr))*0.95)-1] }
    aggs = append(aggs, agg{Service: svc, Count: len(arr), P95: p95})
  }
  sort.Slice(aggs, func(i,j int) bool { return aggs[i].P95 > aggs[j].P95 })
  if req.Max>0 && len(aggs)>req.Max { aggs = aggs[:req.Max] }
  hotspots := []map[string]any{}
  for _, a := range aggs { hotspots = append(hotspots, map[string]any{"service": a.Service, "p95_ms": a.P95, "count": a.Count}) }
  out := map[string]any{
    "hotspots": hotspots,
    "trace_count": len(req.Traces),
    "generated_at": now.Format(time.RFC3339),
    "error": nil,
  }
  return withLinks(out, "gitbook/production/tracing/local-setup.md"), nil
}

// perfRegressionGatekeeper compares benchmark lines (candidate vs baseline) and produces decision.
func perfRegressionGatekeeper(input json.RawMessage, now time.Time) (any, error) {
  var req struct {
    Baseline string `json:"baseline_report"`
    Candidate string `json:"candidate_report"`
    Thresholds struct { NsPerOpPct float64 `json:"ns_per_op_pct"`; AllocsPerOpPct float64 `json:"allocs_per_op_pct"` } `json:"thresholds"`
    HardFailPct float64 `json:"hard_fail_pct"`
    MinSample int `json:"min_sample"`
  }
  if err := json.Unmarshal(input, &req); err != nil { return nil, err }
  if req.Baseline == "" || req.Candidate == "" { return nil, errors.New("baseline_report and candidate_report required") }
  parse := func(s string) map[string]map[string]float64 {
    lines := strings.Split(s, "\n")
    out := map[string]map[string]float64{}
    for _, ln := range lines {
      if !strings.HasPrefix(ln, "Benchmark") { continue }
      f := strings.Fields(ln)
      if len(f) < 5 { continue }
      name := f[0]
      var nsPerOp float64
      for i, fld := range f {
        if strings.HasSuffix(fld, "ns/op") && i>0 {
          fmt.Sscanf(f[i-1], "%f", &nsPerOp)
        }
      }
      if _, ok := out[name]; !ok { out[name] = map[string]float64{} }
      out[name]["ns_per_op"] = nsPerOp
      // crude allocs parse
      for i, fld := range f { if strings.HasSuffix(fld, "allocs/op") && i>0 { var v float64; fmt.Sscanf(f[i-1], "%f", &v); out[name]["allocs_per_op"] = v } }
    }
    return out
  }
  b := parse(req.Baseline)
  c := parse(req.Candidate)
  cases := []map[string]any{}
  decision := "pass"
  for name, bmetrics := range b {
    cm, ok := c[name]
    if !ok { continue }
    deltaNs := pctDelta(bmetrics["ns_per_op"], cm["ns_per_op"])
    deltaAllocs := pctDelta(bmetrics["allocs_per_op"], cm["allocs_per_op"])
    status := "ok"
    if deltaNs > req.Thresholds.NsPerOpPct || deltaAllocs > req.Thresholds.AllocsPerOpPct { status = "warn" }
    if deltaNs > req.HardFailPct { status = "fail"; decision = "fail" }
    cases = append(cases, map[string]any{"name": name, "metric": "ns_per_op", "delta_pct": deltaNs, "status": status})
    cases = append(cases, map[string]any{"name": name, "metric": "allocs_per_op", "delta_pct": deltaAllocs, "status": status})
  }
  out := map[string]any{
    "cases": cases,
    "decision": decision,
    "summary": map[string]any{"total_cases": len(cases), "hard_fail": decision=="fail"},
    "recommendation": recommendationForDecision(decision),
    "error": nil,
  }
  return withLinks(out, "gitbook/production/continuous-profiling/regression-gates-and-workflow.md"), nil
}

// --- Helpers ----------------------------------------------------------------

func unsupported(msg string) Handler { return func(_ json.RawMessage, _ time.Time) (any, error) { return map[string]any{"error": msg}, nil } }

func withLinks(obj map[string]any, doc string) map[string]any {
  obj["__doc_link"] = doc
  return obj
}

func pctDelta(base, cand float64) float64 { if base==0 { return 0 }; return (cand-base)/base*100.0 }

func recommendationForDecision(dec string) string {
  switch dec {
  case "fail": return "Investigate regressions; consider bisect or revert"
  case "pass": return "No hard regressions detected"
  default: return "Review warnings for potential optimization"
  }
}
