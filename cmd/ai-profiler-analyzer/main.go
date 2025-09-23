package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

type TopEntry struct {
    Name    string  `json:"name"`
    Flat    float64 `json:"flat"`          // in samples (best effort)
    FlatPct float64 `json:"flat_pct"`
    Cum     float64 `json:"cum"`
    CumPct  float64 `json:"cum_pct"`
}

type BenchmarkDelta struct {
    Name     string  `json:"name"`
    Metric   string  `json:"metric"`
    Old      string  `json:"old"`
    New      string  `json:"new"`
    DeltaRaw string  `json:"delta_raw"`
    DeltaPct float64 `json:"delta_pct"`
    Worse    bool    `json:"worse"`
}

type AnalysisReport struct {
    GeneratedAt      time.Time        `json:"generated_at"`
    CPUImprovements  []string         `json:"cpu_improvements"`
    CPURegressions   []string         `json:"cpu_regressions"`
    MemImprovements  []string         `json:"mem_improvements"`
    MemRegressions   []string         `json:"mem_regressions"`
    NewHotspots      []string         `json:"new_hotspots"`
    RemovedHotspots  []string         `json:"removed_hotspots"`
    BenchDeltas      []BenchmarkDelta `json:"bench_deltas"`
    Rating           map[string]int   `json:"rating"`
    Summary          string           `json:"summary"`
    Recommendations  []string         `json:"recommendations"`
}

// Simple pprof top line regex capturing numeric columns then name.
var topLineRe = regexp.MustCompile(`^\s*([0-9\.]+)(ms|s)?\s+([0-9\.]+)%\s+([0-9\.]*)?\s*([0-9\.]*)?%?\s+(.+)$`)

func parseTopFile(path string) (map[string]TopEntry, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    res := map[string]TopEntry{}
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "Showing") || strings.HasPrefix(line, "Total:") || strings.HasPrefix(line, "flat") || strings.HasPrefix(line, "Flat") {
            continue
        }
        m := topLineRe.FindStringSubmatch(line)
        if len(m) == 0 { continue }
        flatVal, flatUnit := m[1], m[2]
        flat, _ := strconv.ParseFloat(flatVal, 64)
        if flatUnit == "s" { flat *= 1000 } // to ms
        flatPct, _ := strconv.ParseFloat(m[3], 64)
        name := strings.TrimSpace(m[6])
        res[name] = TopEntry{Name: name, Flat: flat, FlatPct: flatPct}
    }
    return res, scanner.Err()
}

var benchLineRe = regexp.MustCompile(`^(Benchmark\S+)\s+\d+\s+([0-9\.a-zA-Z/]+)\s+->?\s*([0-9\.a-zA-Z/]+)?\s*([-+][0-9\.]+)%?`)

// Fallback benchstat-like parser; we also parse native benchstat diff lines (format may vary).
var benchStatLineRe = regexp.MustCompile(`^(Benchmark\S+)\s+([0-9\.]+)ns/op\s+([0-9\.]+)ns/op\s+([+-][0-9\.]+)%.*$`)

func parseBenchDiff(path string) ([]BenchmarkDelta, error) {
    b, err := os.ReadFile(path)
    if err != nil { return nil, err }
    lines := strings.Split(string(b), "\n")
    out := []BenchmarkDelta{}
    for _, l := range lines {
        l = strings.TrimSpace(l)
        if l == "" || strings.HasPrefix(l, "name") { continue }
        if m := benchStatLineRe.FindStringSubmatch(l); len(m) == 5 {
            oldV, newV, delta := m[2], m[3], m[4]
            dpct, _ := strconv.ParseFloat(strings.TrimSuffix(delta, "%"), 64)
            out = append(out, BenchmarkDelta{Name: m[1], Metric: "time/op", Old: oldV, New: newV, DeltaRaw: delta, DeltaPct: dpct, Worse: dpct > 2})
            continue
        }
    }
    return out, nil
}

func diffHotspots(orig, opt map[string]TopEntry) (improve, regress, added, removed []string) {
    for name, o := range orig {
        if n, ok := opt[name]; ok {
            // improvement if flat pct decreased by >5%
            delta := n.FlatPct - o.FlatPct
            if delta < -5 {
                improve = append(improve, fmt.Sprintf("%s: %.2f%% -> %.2f%% (%.2f%% improvement)", name, o.FlatPct, n.FlatPct, -delta))
            } else if delta > 5 {
                regress = append(regress, fmt.Sprintf("%s: %.2f%% -> %.2f%% (%.2f%% regression)", name, o.FlatPct, n.FlatPct, delta))
            }
        } else {
            removed = append(removed, fmt.Sprintf("%s (was %.2f%% flat)", name, o.FlatPct))
        }
    }
    for name, n := range opt {
        if _, ok := orig[name]; !ok {
            if n.FlatPct > 1.0 { // ignore negligible
                added = append(added, fmt.Sprintf("%s (new %.2f%% flat)", name, n.FlatPct))
            }
        }
    }
    sort.Strings(improve); sort.Strings(regress); sort.Strings(added); sort.Strings(removed)
    return
}

func computeRating(cpuImp, cpuReg, memImp, memReg int) map[string]int {
    rating := map[string]int{"cpu":5,"memory":5,"overall":5}
    if cpuImp > cpuReg { rating["cpu"] += 2 } else if cpuReg > cpuImp { rating["cpu"] -= 2 }
    if memImp > memReg { rating["memory"] += 2 } else if memReg > memImp { rating["memory"] -= 2 }
    // overall heuristic
    rating["overall"] = (rating["cpu"] + rating["memory"]) / 2
    for k,v := range rating { if v < 1 { rating[k] = 1 }; if v > 10 { rating[k] = 10 } }
    return rating
}

func writeFiles(outDir string, rep AnalysisReport) error {
    if err := os.MkdirAll(outDir, 0o755); err != nil { return err }
    jb, _ := json.MarshalIndent(rep, "", "  ")
    if err := os.WriteFile(filepath.Join(outDir, "report.json"), jb, 0o644); err != nil { return err }
    var sb strings.Builder
    sb.WriteString("# AI-Assisted Performance Analysis\n\n")
    sb.WriteString(rep.Summary + "\n\n")
    writeSection := func(title string, items []string) {
        if len(items)==0 { return }
        sb.WriteString("## "+title+"\n")
        for _, it := range items { sb.WriteString("- "+it+"\n") }
        sb.WriteString("\n")
    }
    sb.WriteString("## Ratings\n")
    for k,v := range rep.Rating { sb.WriteString(fmt.Sprintf("- %s: %d/10\n", k, v)) }
    sb.WriteString("\n")
    writeSection("CPU Improvements", rep.CPUImprovements)
    writeSection("CPU Regressions", rep.CPURegressions)
    writeSection("Memory Improvements", rep.MemImprovements)
    writeSection("Memory Regressions", rep.MemRegressions)
    writeSection("New Hotspots", rep.NewHotspots)
    writeSection("Removed Hotspots", rep.RemovedHotspots)
    if len(rep.BenchDeltas) > 0 {
        sb.WriteString("## Benchmark Deltas\n")
        for _, d := range rep.BenchDeltas { sb.WriteString(fmt.Sprintf("- %s %s delta %s (%0.2f%%)%s\n", d.Name, d.Metric, d.DeltaRaw, d.DeltaPct, ternary(d.Worse, " regression", ""))) }
        sb.WriteString("\n")
    }
    writeSection("Recommendations", rep.Recommendations)
    return os.WriteFile(filepath.Join(outDir, "report.md"), []byte(sb.String()), 0o644)
}

func ternary[T any](cond bool, a,b T) T { if cond { return a }; return b }

func maybeCallExternal(endpoint string, rep *AnalysisReport) {
    if endpoint == "" { return }
    jb, _ := json.Marshal(rep)
    req, _ := http.NewRequest("POST", endpoint, strings.NewReader(string(jb)))
    req.Header.Set("Content-Type", "application/json")
    cli := &http.Client{ Timeout: 15 * time.Second }
    resp, err := cli.Do(req)
    if err != nil { return }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    // Expect optional {"summary_override":"...","recommendations":[...]} structure
    var ext map[string]any
    if err := json.Unmarshal(body, &ext); err == nil {
        if s, ok := ext["summary_override"].(string); ok && s != "" { rep.Summary = s }
        if rec, ok := ext["recommendations"].([]any); ok {
            rep.Recommendations = []string{}
            for _, r := range rec { if rs, ok := r.(string); ok { rep.Recommendations = append(rep.Recommendations, rs) } }
        }
    }
}

func main() {
    // Inputs (artifact directories)
    profilesDir := getenv("PROFILES_DIR", ".docs/artifacts/ci")
    benchDir := getenv("BENCH_DIR", ".docs/artifacts/benchdiff")
    outDir := filepath.Join(profilesDir, "ai_analysis")

    origCPU, _ := parseTopFile(filepath.Join(profilesDir, "generator_cpu_top.txt"))
    optCPU, _ := parseTopFile(filepath.Join(profilesDir, "generator_optimized_cpu_top.txt"))
    origMem, _ := parseTopFile(filepath.Join(profilesDir, "generator_mem_top.txt"))
    optMem, _ := parseTopFile(filepath.Join(profilesDir, "generator_optimized_mem_top.txt"))
    cpuImp, cpuReg, cpuAdded, cpuRemoved := diffHotspots(origCPU, optCPU)
    memImp, memReg, memAdded, memRemoved := diffHotspots(origMem, optMem)
    benchOriginal, _ := parseBenchDiff(filepath.Join(benchDir, "bench_original.diff"))
    benchOptimized, _ := parseBenchDiff(filepath.Join(benchDir, "bench_optimized.diff"))
    benchAll := append(benchOriginal, benchOptimized...)

    rating := computeRating(len(cpuImp), len(cpuReg), len(memImp), len(memReg))
    summary := buildSummary(len(cpuImp), len(cpuReg), len(memImp), len(memReg), len(cpuAdded)+len(memAdded), len(cpuRemoved)+len(memRemoved))
    recs := buildRecommendations(cpuReg, memReg, cpuAdded, benchAll)

    rep := AnalysisReport{
        GeneratedAt: time.Now(),
        CPUImprovements: cpuImp,
        CPURegressions: cpuReg,
        MemImprovements: memImp,
        MemRegressions: memReg,
        NewHotspots: append(cpuAdded, memAdded...),
        RemovedHotspots: append(cpuRemoved, memRemoved...),
        BenchDeltas: benchAll,
        Rating: rating,
        Summary: summary,
        Recommendations: recs,
    }

    maybeCallExternal(os.Getenv("AI_ANALYSIS_ENDPOINT"), &rep)
    if err := writeFiles(outDir, rep); err != nil { fmt.Println("failed writing report:", err) }
    fmt.Println("AI analysis report generated at", outDir)
    if os.Getenv("AI_GATE_ENABLE") == "true" {
        gateFail := evaluateGates(&rep)
        if gateFail != "" {
            fmt.Println("Gate failure:\n" + gateFail)
            os.Exit(2)
        }
        fmt.Println("All performance gates passed")
    }
}

func buildSummary(ci, cr, mi, mr, added, removed int) string {
    totalImp := ci + mi
    totalReg := cr + mr
    verb := "net neutral"
    if totalImp > totalReg { verb = "net improvement" } else if totalReg > totalImp { verb = "net regression" }
    return fmt.Sprintf("Overall %s: CPU %+d / -%d, Memory %+d / -%d, New hotspots %d, Removed %d", verb, ci, cr, mi, mr, added, removed)
}

func buildRecommendations(cpuReg, memReg, cpuAdded []string, bench []BenchmarkDelta) []string {
    rec := []string{}
    if len(cpuReg) > 0 { rec = append(rec, "Investigate CPU regressions (consider profiling those functions with deeper block/mutex traces)") }
    if len(memReg) > 0 { rec = append(rec, "Memory regressions detected; look for transient allocations and enable escape analysis (-gcflags '-m') for hotspots.") }
    if len(cpuAdded) > 0 { rec = append(rec, "New CPU hotspots emerged; validate they are intentional and consider algorithmic review.") }
    for _, b := range bench { if b.Worse { rec = append(rec, fmt.Sprintf("Benchmark %s shows %0.2f%% slowdown; replicate locally with 'go test -bench=%s -count=10'", b.Name, b.DeltaPct, b.Name)) } }
    if len(rec) == 0 { rec = append(rec, "No significant issues detected; consider tightening thresholds or expanding benchmark coverage.") }
    return rec
}

func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

func evaluateGates(rep *AnalysisReport) string {
    var issues []string
    cpuMax := envFloat("AI_GATE_CPU_REG_MAX", 8)
    memMax := envFloat("AI_GATE_MEM_REG_MAX", 10)
    benchMax := envFloat("AI_GATE_BENCH_SLOW_MAX", 5)
    minRating := int(envFloat("AI_GATE_MIN_RATING", 0))
    // Parse regression lines to extract percentage inside parentheses
    pctRe := regexp.MustCompile(`\(([-+0-9\.]+)% regression\)$`)
    scan := func(lines []string, max float64, label string) {
        for _, l := range lines {
            if m := pctRe.FindStringSubmatch(l); len(m)==2 { if v,err:=strconv.ParseFloat(m[1],64); err==nil && v>max { issues = append(issues, fmt.Sprintf("%s regression %.2f%% > %.2f%%: %s", label, v, max, l)) } }
        }
    }
    scan(rep.CPURegressions, cpuMax, "CPU")
    scan(rep.MemRegressions, memMax, "MEM")
    for _, b := range rep.BenchDeltas { if b.Worse && b.DeltaPct > benchMax { issues = append(issues, fmt.Sprintf("Benchmark %s slowdown %.2f%% > %.2f%%", b.Name, b.DeltaPct, benchMax)) } }
    if minRating > 0 && rep.Rating["overall"] < minRating { issues = append(issues, fmt.Sprintf("Overall rating %d < %d", rep.Rating["overall"], minRating)) }
    if len(issues) == 0 { return "" }
    return strings.Join(issues, "\n")
}

func envFloat(k string, def float64) float64 { if v:=os.Getenv(k); v!="" { if f,err:=strconv.ParseFloat(v,64); err==nil { return f } }; return def }
