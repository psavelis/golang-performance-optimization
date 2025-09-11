# Flamegraph Analysis

Flamegraphs are powerful visualization tools that transform CPU profiling data into intuitive hierarchical charts. This guide covers creating, interpreting, and leveraging flamegraphs for deep performance analysis.

## Introduction to Flamegraphs

Flamegraphs represent CPU usage as a hierarchical visualization where:
- **Width** represents time spent (wider = more time)
- **Height** represents call stack depth
- **Color** typically indicates different modules or languages
- **Interactive** elements allow drilling down into specific functions

### Why Flamegraphs Matter

```go
// Traditional text output is hard to interpret:
// main.processData  2.34s  78.52%
// main.calculateSum 1.20s  40.27%
// main.validateInput 0.64s  21.48%

// Flamegraph shows the relationship visually:
// [    main.processData (2.34s)    ]
//   [calculateSum 1.20s][validateInput 0.64s]
//     [math ops][validation logic]
```

## Generating Flamegraphs

### Using go tool pprof

```bash
# Generate SVG flamegraph
go tool pprof -http=:8080 cpu.prof

# Generate static SVG
go tool pprof -svg cpu.prof > flamegraph.svg

# Generate interactive HTML (if available)
go tool pprof -web cpu.prof
```

### Using Brendan Gregg's FlameGraph Tools

```bash
# Install FlameGraph tools
git clone https://github.com/brendangregg/FlameGraph
cd FlameGraph

# Convert pprof to FlameGraph format
go tool pprof -raw cpu.prof | ./flamegraph.pl > flame.svg

# With folded format
go tool pprof -traces cpu.prof | ./flamegraph.pl > flame.svg
```

### Programmatic Flamegraph Generation

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "runtime/pprof"
    "time"
)

type FlameGraphGenerator struct {
    profileDir    string
    flameGraphDir string
    toolsPath     string
}

func NewFlameGraphGenerator(profileDir, flameGraphDir, toolsPath string) *FlameGraphGenerator {
    return &FlameGraphGenerator{
        profileDir:    profileDir,
        flameGraphDir: flameGraphDir,
        toolsPath:     toolsPath,
    }
}

func (fg *FlameGraphGenerator) GenerateFromProfile(profilePath string) (string, error) {
    // Ensure output directory exists
    if err := os.MkdirAll(fg.flameGraphDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create flamegraph directory: %v", err)
    }

    // Generate timestamp for unique filename
    timestamp := time.Now().Format("20060102_150405")
    outputFile := filepath.Join(fg.flameGraphDir, fmt.Sprintf("flamegraph_%s.svg", timestamp))

    // Convert profile to flamegraph format
    cmd := exec.Command("go", "tool", "pprof", "-raw", profilePath)
    rawOutput, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to convert profile: %v", err)
    }

    // Generate flamegraph SVG
    flameCmd := exec.Command(filepath.Join(fg.toolsPath, "flamegraph.pl"))
    flameCmd.Stdin = strings.NewReader(string(rawOutput))

    svgOutput, err := flameCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to generate flamegraph: %v", err)
    }

    // Write SVG file
    if err := os.WriteFile(outputFile, svgOutput, 0644); err != nil {
        return "", fmt.Errorf("failed to write flamegraph: %v", err)
    }

    return outputFile, nil
}

func (fg *FlameGraphGenerator) GenerateInteractive(profilePath string) (string, error) {
    // Generate interactive HTML flamegraph
    timestamp := time.Now().Format("20060102_150405")
    outputFile := filepath.Join(fg.flameGraphDir, fmt.Sprintf("interactive_%s.html", timestamp))

    // Use pprof's web interface programmatically
    cmd := exec.Command("go", "tool", "pprof", "-http", ":0", profilePath)
    // Implementation depends on specific requirements
    
    return outputFile, nil
}

// HTTP handler for real-time flamegraph generation
func (fg *FlameGraphGenerator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    profilePath := r.URL.Query().Get("profile")
    if profilePath == "" {
        http.Error(w, "profile parameter required", http.StatusBadRequest)
        return
    }

    // Validate profile path
    fullPath := filepath.Join(fg.profileDir, profilePath)
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        http.Error(w, "profile not found", http.StatusNotFound)
        return
    }

    // Generate flamegraph
    svgPath, err := fg.GenerateFromProfile(fullPath)
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to generate flamegraph: %v", err), http.StatusInternalServerError)
        return
    }

    // Serve SVG file
    http.ServeFile(w, r, svgPath)
}
```

## Advanced Flamegraph Techniques

### Differential Flamegraphs

```go
package main

import (
    "fmt"
    "os/exec"
    "path/filepath"
)

func generateDifferentialFlameGraph(beforeProfile, afterProfile, outputPath string) error {
    // Generate baseline flamegraph data
    baselineCmd := exec.Command("go", "tool", "pprof", "-raw", beforeProfile)
    baselineData, err := baselineCmd.Output()
    if err != nil {
        return fmt.Errorf("failed to process baseline profile: %v", err)
    }

    // Generate comparison flamegraph data
    comparisonCmd := exec.Command("go", "tool", "pprof", "-raw", afterProfile)
    comparisonData, err := comparisonCmd.Output()
    if err != nil {
        return fmt.Errorf("failed to process comparison profile: %v", err)
    }

    // Generate differential flamegraph
    diffCmd := exec.Command("difffolded.pl", string(baselineData), string(comparisonData))
    diffData, err := diffCmd.Output()
    if err != nil {
        return fmt.Errorf("failed to generate diff: %v", err)
    }

    // Create flamegraph from diff data
    flameCmd := exec.Command("flamegraph.pl", "--title", "Differential Flamegraph")
    flameCmd.Stdin = strings.NewReader(string(diffData))
    
    svgData, err := flameCmd.Output()
    if err != nil {
        return fmt.Errorf("failed to generate flamegraph: %v", err)
    }

    // Write output
    return os.WriteFile(outputPath, svgData, 0644)
}

// Usage example
func compareTwoOptimizations() {
    beforeProfile := "cpu_before_optimization.prof"
    afterProfile := "cpu_after_optimization.prof"
    outputFile := "differential_flamegraph.svg"

    if err := generateDifferentialFlameGraph(beforeProfile, afterProfile, outputFile); err != nil {
        fmt.Printf("Error generating differential flamegraph: %v\n", err)
        return
    }

    fmt.Printf("Differential flamegraph saved to: %s\n", outputFile)
}
```

### Custom Flamegraph Colors

```go
package main

import (
    "fmt"
    "os/exec"
    "strings"
)

type FlameGraphConfig struct {
    Title      string
    Colors     string
    Width      int
    Height     int
    FontSize   int
    Reverse    bool
    Inverted   bool
}

func generateCustomFlameGraph(profilePath string, config FlameGraphConfig) (string, error) {
    // Build flamegraph command with custom options
    args := []string{}
    
    if config.Title != "" {
        args = append(args, "--title", config.Title)
    }
    
    if config.Colors != "" {
        args = append(args, "--colors", config.Colors)
    }
    
    if config.Width > 0 {
        args = append(args, "--width", fmt.Sprintf("%d", config.Width))
    }
    
    if config.Height > 0 {
        args = append(args, "--height", fmt.Sprintf("%d", config.Height))
    }
    
    if config.FontSize > 0 {
        args = append(args, "--fontsize", fmt.Sprintf("%d", config.FontSize))
    }
    
    if config.Reverse {
        args = append(args, "--reverse")
    }
    
    if config.Inverted {
        args = append(args, "--inverted")
    }

    // Get raw profile data
    profileCmd := exec.Command("go", "tool", "pprof", "-raw", profilePath)
    profileData, err := profileCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get profile data: %v", err)
    }

    // Generate flamegraph with custom options
    flameCmd := exec.Command("flamegraph.pl", args...)
    flameCmd.Stdin = strings.NewReader(string(profileData))
    
    svgData, err := flameCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to generate flamegraph: %v", err)
    }

    return string(svgData), nil
}

// Predefined color schemes
func getColorSchemes() map[string]string {
    return map[string]string{
        "hot":     "hot",      // Hot colors (red, orange, yellow)
        "mem":     "mem",      // Memory allocation colors
        "io":      "io",       // I/O operation colors
        "java":    "java",     // Java-style colors
        "js":      "js",       // JavaScript colors
        "perl":    "perl",     // Perl colors
        "chain":   "chain",    // Color chains
        "aqua":    "aqua",     // Aqua theme
    }
}

// Generate multiple themed flamegraphs
func generateThemedFlameGraphs(profilePath string) {
    themes := getColorSchemes()
    
    for themeName, colorScheme := range themes {
        config := FlameGraphConfig{
            Title:    fmt.Sprintf("CPU Profile - %s Theme", strings.Title(themeName)),
            Colors:   colorScheme,
            Width:    1200,
            Height:   800,
            FontSize: 12,
        }
        
        svgData, err := generateCustomFlameGraph(profilePath, config)
        if err != nil {
            fmt.Printf("Failed to generate %s flamegraph: %v\n", themeName, err)
            continue
        }
        
        outputFile := fmt.Sprintf("flamegraph_%s.svg", themeName)
        if err := os.WriteFile(outputFile, []byte(svgData), 0644); err != nil {
            fmt.Printf("Failed to write %s flamegraph: %v\n", themeName, err)
            continue
        }
        
        fmt.Printf("Generated %s flamegraph: %s\n", themeName, outputFile)
    }
}
```

## Interpreting Flamegraphs

### Reading Flamegraph Patterns

```go
package main

import (
    "fmt"
    "runtime/pprof"
    "time"
)

// Example: CPU hotspot pattern
func demonstrateCPUHotspot() {
    f, _ := os.Create("hotspot_demo.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // This will show as a wide bar in the flamegraph
    for i := 0; i < 1000000; i++ {
        _ = expensiveComputation(i)
    }
}

func expensiveComputation(n int) int {
    // This function will appear as the widest bar
    sum := 0
    for i := 0; i < n%1000; i++ {
        sum += i * i
    }
    return sum
}

// Example: Deep call stack pattern
func demonstrateDeepStack() {
    f, _ := os.Create("deep_stack_demo.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // This will show as a tall tower in the flamegraph
    recursiveFunction(100)
}

func recursiveFunction(depth int) int {
    if depth <= 0 {
        // Actual work at the bottom of the stack
        return performWork()
    }
    return recursiveFunction(depth - 1)
}

func performWork() int {
    sum := 0
    for i := 0; i < 10000; i++ {
        sum += i
    }
    return sum
}

// Example: Multiple execution paths
func demonstrateMultiplePaths() {
    f, _ := os.Create("multiple_paths_demo.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // This will show as multiple separate stacks
    go pathA()
    go pathB()
    go pathC()
    
    time.Sleep(5 * time.Second)
}

func pathA() {
    for i := 0; i < 100000; i++ {
        _ = algorithmA(i)
    }
}

func pathB() {
    for i := 0; i < 200000; i++ {
        _ = algorithmB(i)
    }
}

func pathC() {
    for i := 0; i < 150000; i++ {
        _ = algorithmC(i)
    }
}

func algorithmA(n int) int { return n * 2 }
func algorithmB(n int) int { return n * 3 }
func algorithmC(n int) int { return n * 4 }
```

### Common Flamegraph Patterns

#### 1. The Tower (Deep Recursion)
```
[                main                ]
[           recursiveFunc            ]
[        recursiveFunc               ]
[     recursiveFunc                  ]
[  actualWork                        ]
```

#### 2. The Plateau (CPU Hotspot)
```
[           main.processData         ]
[    expensiveComputation            ]
```

#### 3. The Mountain Range (Multiple Hotspots)
```
[              main                  ]
[  funcA  ][  funcB  ][    funcC    ]
```

#### 4. The Icicle (Inverted View)
```
[  actualWork                        ]
[     recursiveFunc                  ]
[        recursiveFunc               ]
[           recursiveFunc            ]
[                main                ]
```

## Interactive Flamegraph Analysis

### Web-Based Analysis Tool

```go
package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
)

type FlameGraphServer struct {
    profileDir string
    templates  *template.Template
}

type ProfileInfo struct {
    Name     string `json:"name"`
    Path     string `json:"path"`
    Size     int64  `json:"size"`
    Modified string `json:"modified"`
}

func NewFlameGraphServer(profileDir string) *FlameGraphServer {
    return &FlameGraphServer{
        profileDir: profileDir,
        templates:  parseTemplates(),
    }
}

func (fgs *FlameGraphServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/":
        fgs.handleIndex(w, r)
    case "/profiles":
        fgs.handleProfiles(w, r)
    case "/flamegraph":
        fgs.handleFlameGraph(w, r)
    case "/compare":
        fgs.handleCompare(w, r)
    default:
        http.NotFound(w, r)
    }
}

func (fgs *FlameGraphServer) handleIndex(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Title string
    }{
        Title: "FlameGraph Analysis Tool",
    }
    
    fgs.templates.ExecuteTemplate(w, "index.html", data)
}

func (fgs *FlameGraphServer) handleProfiles(w http.ResponseWriter, r *http.Request) {
    profiles, err := fgs.listProfiles()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(profiles)
}

func (fgs *FlameGraphServer) handleFlameGraph(w http.ResponseWriter, r *http.Request) {
    profileName := r.URL.Query().Get("profile")
    if profileName == "" {
        http.Error(w, "profile parameter required", http.StatusBadRequest)
        return
    }

    width := 1200
    if w, err := strconv.Atoi(r.URL.Query().Get("width")); err == nil && w > 0 {
        width = w
    }

    height := 800
    if h, err := strconv.Atoi(r.URL.Query().Get("height")); err == nil && h > 0 {
        height = h
    }

    title := r.URL.Query().Get("title")
    if title == "" {
        title = fmt.Sprintf("FlameGraph - %s", profileName)
    }

    // Generate flamegraph with specified parameters
    config := FlameGraphConfig{
        Title:  title,
        Width:  width,
        Height: height,
        Colors: r.URL.Query().Get("colors"),
    }

    profilePath := filepath.Join(fgs.profileDir, profileName)
    svgData, err := generateCustomFlameGraph(profilePath, config)
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to generate flamegraph: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "image/svg+xml")
    w.Write([]byte(svgData))
}

func (fgs *FlameGraphServer) handleCompare(w http.ResponseWriter, r *http.Request) {
    beforeProfile := r.URL.Query().Get("before")
    afterProfile := r.URL.Query().Get("after")
    
    if beforeProfile == "" || afterProfile == "" {
        http.Error(w, "both 'before' and 'after' parameters required", http.StatusBadRequest)
        return
    }

    beforePath := filepath.Join(fgs.profileDir, beforeProfile)
    afterPath := filepath.Join(fgs.profileDir, afterProfile)
    
    // Generate differential flamegraph
    outputPath := fmt.Sprintf("/tmp/diff_%s_%s.svg", beforeProfile, afterProfile)
    if err := generateDifferentialFlameGraph(beforePath, afterPath, outputPath); err != nil {
        http.Error(w, fmt.Sprintf("failed to generate differential flamegraph: %v", err), http.StatusInternalServerError)
        return
    }

    // Serve the generated file
    http.ServeFile(w, r, outputPath)
}

func (fgs *FlameGraphServer) listProfiles() ([]ProfileInfo, error) {
    files, err := filepath.Glob(filepath.Join(fgs.profileDir, "*.prof"))
    if err != nil {
        return nil, err
    }

    profiles := make([]ProfileInfo, 0, len(files))
    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            continue
        }

        profiles = append(profiles, ProfileInfo{
            Name:     filepath.Base(file),
            Path:     file,
            Size:     info.Size(),
            Modified: info.ModTime().Format("2006-01-02 15:04:05"),
        })
    }

    return profiles, nil
}

func parseTemplates() *template.Template {
    indexTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .flamegraph { border: 1px solid #ccc; margin: 20px 0; }
        .controls { margin: 20px 0; padding: 20px; background: #f5f5f5; border-radius: 5px; }
        .control-group { margin: 10px 0; }
        label { display: inline-block; width: 100px; }
        input, select { margin: 0 10px; padding: 5px; }
        button { padding: 10px 20px; background: #007cba; color: white; border: none; border-radius: 3px; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        
        <div class="controls">
            <h3>Generate FlameGraph</h3>
            <div class="control-group">
                <label>Profile:</label>
                <select id="profileSelect">
                    <option value="">Select a profile...</option>
                </select>
            </div>
            <div class="control-group">
                <label>Width:</label>
                <input type="number" id="width" value="1200" min="800" max="2400">
                <label>Height:</label>
                <input type="number" id="height" value="800" min="400" max="1200">
            </div>
            <div class="control-group">
                <label>Colors:</label>
                <select id="colors">
                    <option value="">Default</option>
                    <option value="hot">Hot</option>
                    <option value="mem">Memory</option>
                    <option value="io">I/O</option>
                    <option value="java">Java</option>
                    <option value="js">JavaScript</option>
                </select>
            </div>
            <div class="control-group">
                <button onclick="generateFlameGraph()">Generate FlameGraph</button>
                <button onclick="compareProfiles()">Compare Profiles</button>
            </div>
        </div>
        
        <div id="flamegraphContainer" class="flamegraph"></div>
    </div>

    <script>
        // Load available profiles
        fetch('/profiles')
            .then(response => response.json())
            .then(profiles => {
                const select = document.getElementById('profileSelect');
                profiles.forEach(profile => {
                    const option = document.createElement('option');
                    option.value = profile.name;
                    option.textContent = profile.name + ' (' + formatSize(profile.size) + ')';
                    select.appendChild(option);
                });
            });

        function generateFlameGraph() {
            const profile = document.getElementById('profileSelect').value;
            const width = document.getElementById('width').value;
            const height = document.getElementById('height').value;
            const colors = document.getElementById('colors').value;
            
            if (!profile) {
                alert('Please select a profile');
                return;
            }
            
            const params = new URLSearchParams({
                profile: profile,
                width: width,
                height: height,
                colors: colors
            });
            
            const container = document.getElementById('flamegraphContainer');
            container.innerHTML = '<embed src="/flamegraph?' + params + '" type="image/svg+xml" width="100%" height="' + height + 'px">';
        }
        
        function compareProfiles() {
            // Implementation for profile comparison
            alert('Profile comparison feature coming soon!');
        }
        
        function formatSize(bytes) {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
            return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
        }
    </script>
</body>
</html>
`
    
    return template.Must(template.New("index.html").Parse(indexTemplate))
}

// Start the flamegraph server
func startFlameGraphServer() {
    server := NewFlameGraphServer("./profiles")
    
    fmt.Println("FlameGraph server starting on http://localhost:8080")
    http.ListenAndServe(":8080", server)
}
```

## Automated Flamegraph Analysis

### Pattern Detection

```go
package main

import (
    "bufio"
    "fmt"
    "os/exec"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

type FlameAnalysis struct {
    HotFunctions    []FunctionMetric
    DeepStacks      []StackInfo
    RecursivePatterns []RecursionInfo
    CPUDistribution map[string]float64
}

type FunctionMetric struct {
    Name    string
    Samples int
    Percent float64
}

type StackInfo struct {
    Depth   int
    Pattern string
    Samples int
}

type RecursionInfo struct {
    Function string
    MaxDepth int
    Samples  int
}

func analyzeFlameGraph(profilePath string) (*FlameAnalysis, error) {
    // Get raw profile data
    cmd := exec.Command("go", "tool", "pprof", "-top", profilePath)
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get profile data: %v", err)
    }

    analysis := &FlameAnalysis{
        CPUDistribution: make(map[string]float64),
    }

    // Parse top functions
    analysis.HotFunctions = parseTopFunctions(string(output))
    
    // Get trace data for stack analysis
    traceCmd := exec.Command("go", "tool", "pprof", "-traces", profilePath)
    traceOutput, err := traceCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get trace data: %v", err)
    }

    // Analyze stack patterns
    analysis.DeepStacks = analyzeStackDepth(string(traceOutput))
    analysis.RecursivePatterns = analyzeRecursion(string(traceOutput))
    analysis.CPUDistribution = analyzeCPUDistribution(string(output))

    return analysis, nil
}

func parseTopFunctions(output string) []FunctionMetric {
    lines := strings.Split(output, "\n")
    functions := []FunctionMetric{}
    
    // Regex to parse pprof top output
    re := regexp.MustCompile(`^\s*(\d+(?:\.\d+)?[a-zA-Z]*)\s+(\d+(?:\.\d+)?)%.*?(\S+)$`)
    
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if len(matches) >= 4 {
            samples := parseValue(matches[1])
            percent, _ := strconv.ParseFloat(matches[2], 64)
            name := matches[3]
            
            functions = append(functions, FunctionMetric{
                Name:    name,
                Samples: int(samples),
                Percent: percent,
            })
        }
    }
    
    return functions
}

func analyzeStackDepth(traceOutput string) []StackInfo {
    stacks := make(map[int]int) // depth -> count
    
    traces := strings.Split(traceOutput, "\n\n")
    for _, trace := range traces {
        lines := strings.Split(trace, "\n")
        depth := 0
        
        for _, line := range lines {
            if strings.TrimSpace(line) != "" && !strings.Contains(line, "samples:") {
                depth++
            }
        }
        
        if depth > 0 {
            stacks[depth]++
        }
    }
    
    var stackInfos []StackInfo
    for depth, count := range stacks {
        stackInfos = append(stackInfos, StackInfo{
            Depth:   depth,
            Samples: count,
        })
    }
    
    sort.Slice(stackInfos, func(i, j int) bool {
        return stackInfos[i].Samples > stackInfos[j].Samples
    })
    
    return stackInfos
}

func analyzeRecursion(traceOutput string) []RecursionInfo {
    recursions := make(map[string]*RecursionInfo)
    
    traces := strings.Split(traceOutput, "\n\n")
    for _, trace := range traces {
        lines := strings.Split(trace, "\n")
        functionCounts := make(map[string]int)
        
        for _, line := range lines {
            if strings.TrimSpace(line) != "" && !strings.Contains(line, "samples:") {
                // Extract function name (simplified)
                parts := strings.Fields(line)
                if len(parts) > 0 {
                    funcName := parts[len(parts)-1]
                    functionCounts[funcName]++
                }
            }
        }
        
        // Check for recursion
        for funcName, count := range functionCounts {
            if count > 1 { // Recursive call detected
                if existing, ok := recursions[funcName]; ok {
                    existing.Samples++
                    if count > existing.MaxDepth {
                        existing.MaxDepth = count
                    }
                } else {
                    recursions[funcName] = &RecursionInfo{
                        Function: funcName,
                        MaxDepth: count,
                        Samples:  1,
                    }
                }
            }
        }
    }
    
    var result []RecursionInfo
    for _, info := range recursions {
        result = append(result, *info)
    }
    
    sort.Slice(result, func(i, j int) bool {
        return result[i].Samples > result[j].Samples
    })
    
    return result
}

func analyzeCPUDistribution(output string) map[string]float64 {
    distribution := make(map[string]float64)
    
    lines := strings.Split(output, "\n")
    re := regexp.MustCompile(`^\s*\d+(?:\.\d+)?[a-zA-Z]*\s+(\d+(?:\.\d+)?)%.*?(\S+)$`)
    
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if len(matches) >= 3 {
            percent, _ := strconv.ParseFloat(matches[1], 64)
            name := matches[2]
            
            // Categorize by package/module
            category := categorizeFunction(name)
            distribution[category] += percent
        }
    }
    
    return distribution
}

func categorizeFunction(funcName string) string {
    switch {
    case strings.Contains(funcName, "runtime"):
        return "runtime"
    case strings.Contains(funcName, "main"):
        return "application"
    case strings.Contains(funcName, "net"):
        return "networking"
    case strings.Contains(funcName, "crypto"):
        return "crypto"
    case strings.Contains(funcName, "encoding"):
        return "encoding"
    case strings.Contains(funcName, "database"):
        return "database"
    default:
        return "other"
    }
}

func parseValue(valueStr string) float64 {
    // Parse values like "2.34s", "1.20ms", etc.
    re := regexp.MustCompile(`(\d+(?:\.\d+)?)([a-zA-Z]*)`)
    matches := re.FindStringSubmatch(valueStr)
    if len(matches) >= 2 {
        value, _ := strconv.ParseFloat(matches[1], 64)
        return value
    }
    return 0
}

// Generate analysis report
func (fa *FlameAnalysis) GenerateReport() string {
    var report strings.Builder
    
    report.WriteString("=== FLAMEGRAPH ANALYSIS REPORT ===\n\n")
    
    // Hot functions
    report.WriteString("TOP CPU CONSUMERS:\n")
    for i, fn := range fa.HotFunctions {
        if i >= 10 {
            break
        }
        report.WriteString(fmt.Sprintf("%2d. %s: %.2f%% (%d samples)\n", 
            i+1, fn.Name, fn.Percent, fn.Samples))
    }
    
    // Stack depth analysis
    report.WriteString("\nSTACK DEPTH ANALYSIS:\n")
    for i, stack := range fa.DeepStacks {
        if i >= 5 {
            break
        }
        report.WriteString(fmt.Sprintf("Depth %d: %d traces\n", 
            stack.Depth, stack.Samples))
    }
    
    // Recursion patterns
    if len(fa.RecursivePatterns) > 0 {
        report.WriteString("\nRECURSIVE PATTERNS DETECTED:\n")
        for i, rec := range fa.RecursivePatterns {
            if i >= 5 {
                break
            }
            report.WriteString(fmt.Sprintf("%s: max depth %d (%d occurrences)\n", 
                rec.Function, rec.MaxDepth, rec.Samples))
        }
    }
    
    // CPU distribution
    report.WriteString("\nCPU DISTRIBUTION BY CATEGORY:\n")
    for category, percent := range fa.CPUDistribution {
        if percent > 1.0 {
            report.WriteString(fmt.Sprintf("%s: %.2f%%\n", category, percent))
        }
    }
    
    return report.String()
}
```

## Best Practices for Flamegraph Analysis

### 1. Focus on Width, Not Height

```go
// Wide functions are CPU hotspots - optimize these first
// Tall stacks might be normal (deep call chains)

func identifyOptimizationTargets(analysis *FlameAnalysis) {
    fmt.Println("OPTIMIZATION PRIORITIES:")
    
    for i, fn := range analysis.HotFunctions {
        if i >= 5 { // Top 5 targets
            break
        }
        
        priority := "HIGH"
        if fn.Percent < 10 {
            priority = "MEDIUM"
        }
        if fn.Percent < 5 {
            priority = "LOW"
        }
        
        fmt.Printf("%s: %s (%.2f%% CPU)\n", priority, fn.Name, fn.Percent)
    }
}
```

### 2. Compare Before and After

```go
func validateOptimization(beforeProfile, afterProfile string) {
    beforeAnalysis, _ := analyzeFlameGraph(beforeProfile)
    afterAnalysis, _ := analyzeFlameGraph(afterProfile)
    
    fmt.Println("OPTIMIZATION RESULTS:")
    
    // Compare top functions
    beforeTop := beforeAnalysis.HotFunctions[0]
    afterTop := afterAnalysis.HotFunctions[0]
    
    improvement := beforeTop.Percent - afterTop.Percent
    fmt.Printf("Top hotspot improvement: %.2f%% → %.2f%% (%.2f%% reduction)\n",
        beforeTop.Percent, afterTop.Percent, improvement)
}
```

### 3. Automate Flamegraph Generation

```bash
#!/bin/bash
# flamegraph_pipeline.sh

PROFILE_DIR="./profiles"
OUTPUT_DIR="./flamegraphs"

mkdir -p $OUTPUT_DIR

# Generate flamegraphs for all profiles
for profile in $PROFILE_DIR/*.prof; do
    basename=$(basename "$profile" .prof)
    
    # Standard flamegraph
    go tool pprof -svg "$profile" > "$OUTPUT_DIR/${basename}_standard.svg"
    
    # Hot color scheme
    flamegraph.pl --colors hot < <(go tool pprof -raw "$profile") > "$OUTPUT_DIR/${basename}_hot.svg"
    
    # Memory-focused
    flamegraph.pl --colors mem < <(go tool pprof -raw "$profile") > "$OUTPUT_DIR/${basename}_mem.svg"
    
    echo "Generated flamegraphs for $basename"
done

echo "All flamegraphs generated in $OUTPUT_DIR"
```

## Next Steps

- Apply insights to [Memory Profiling](../memory-profiling/README.md)
- Learn [Goroutine Profiling](../goroutine-profiling/README.md) visualization
- Explore [Advanced Benchmarking](../../benchmarking/advanced/README.md) with flamegraphs

## Summary

Flamegraph analysis transforms raw profiling data into actionable insights:

1. **Visual pattern recognition** reveals performance bottlenecks
2. **Interactive exploration** enables deep-dive analysis
3. **Differential flamegraphs** validate optimizations
4. **Automated analysis** scales performance monitoring
5. **Custom visualizations** highlight specific concerns

Master flamegraph interpretation to accelerate performance optimization and make data-driven decisions about where to focus optimization efforts.
