package main

import (
  "encoding/json"
  "errors"
  "flag"
  "fmt"
  "io"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/xeipuuv/gojsonschema"
  "github.com/dmgo1014/interviewing-golang/mcp/handlers"
)

// Simple CLI wrapper that discovers capabilities from manifest and dispatches to stub handlers.
// In a real implementation each MCP would call upstream systems. Here we focus on manifest loading
// and input contract validation placeholder (schema validation can be added via a library like
// github.com/xeipuuv/gojsonschema if desired later).

type Manifest struct {
  Version      int            `json:"version"`
  Generated    string         `json:"generated"`
  Capabilities []Capability   `json:"capabilities"`
}

type Capability struct {
  ID          string `json:"id"`
  Domain      string `json:"domain"`
  Description string `json:"description"`
  Maturity    string `json:"maturity"`
  InputSchema string `json:"inputSchema"`
  OutputSchema string `json:"outputSchema"`
}

var (
  manifestPath = flag.String("manifest", "mcp/manifest.yaml", "Path to MCP manifest YAML (YAML or JSON)")
  capabilityID = flag.String("cap", "", "Capability ID to invoke")
  inputPath    = flag.String("in", "", "JSON input file path (defaults to stdin if empty)")
  listOnly     = flag.Bool("list", false, "List capabilities and exit")
  timeoutSec   = flag.Int("timeout", 30, "Execution timeout seconds")
  validateOnly = flag.Bool("validate", false, "Validate input against schema and exit (no execution)")
  showSchemas  = flag.Bool("schemas", false, "List capabilities with schema paths")
)

func main() {
  flag.Parse()
  if *listOnly { listCapabilities(false); return }
  if *showSchemas { listCapabilities(true); return }
  if *capabilityID == "" {
    fatal("-cap required (use -list to view IDs)")
  }
  man, err := loadManifest(*manifestPath)
  if err != nil {
    fatal(err.Error())
  }
  cap, ok := findCap(man, *capabilityID)
  if !ok {
    fatal("capability not found in manifest: " + *capabilityID)
  }
  payload, err := readInput(*inputPath)
  if err != nil {
    fatal(err.Error())
  }
  if err := validateJSONAgainstSchema(payload, cap.InputSchema); err != nil {
    fatal("schema validation failed: " + err.Error())
  }
  if *validateOnly {
    fmt.Println("INPUT OK (schema validated)")
    return
  }
  h, ok := handlers.Registry[cap.ID]
  if !ok { fatal("no handler registered for capability: " + cap.ID) }
  result, err := h(payload, time.Now().UTC())
  if err != nil { fatal(err.Error()) }
  // Extract internal doc link if set & clean prior to validation
  var docLink string
  if m, ok := result.(map[string]any); ok {
    if v, ok2 := m["__doc_link"].(string); ok2 { docLink = v; delete(m, "__doc_link") }
  }
  if cap.OutputSchema != "" {
    ob, _ := json.Marshal(result)
    if err := validateJSONAgainstSchema(ob, cap.OutputSchema); err != nil {
      fmt.Fprintf(os.Stderr, "warning: produced output failed schema validation: %v\n", err)
    }
  }
  // Wrap with envelope (metadata outside schema validation scope)
  out := map[string]any{
    "capability": cap.ID,
    "received_at": time.Now().UTC().Format(time.RFC3339),
    "result": result,
  }
  if docLink != "" { out["docs"] = docLink }
  enc := json.NewEncoder(os.Stdout)
  enc.SetIndent("", "  ")
  _ = enc.Encode(out)
}

func listCapabilities(withSchemas bool) {
  man, err := loadManifest(*manifestPath)
  if err != nil { fatal(err.Error()) }
  for _, c := range man.Capabilities {
    if withSchemas {
      fmt.Printf("%s\t%s\t%s\t%s\t%s\n", c.ID, c.Domain, c.Maturity, c.InputSchema, c.OutputSchema)
    } else {
      fmt.Printf("%s\t%s\t%s\n", c.ID, c.Domain, c.Maturity)
    }
  }
}

func loadManifest(path string) (*Manifest, error) {
  b, err := os.ReadFile(path)
  if err != nil { return nil, err }
  var man Manifest
  if json.Unmarshal(b, &man) == nil && len(man.Capabilities) > 0 {
    return &man, nil
  }
  // YAML fallback (very limited) – parse top-level keys and capability list items.
  lines := strings.Split(string(b), "\n")
  var caps []Capability
  var current *Capability
  for _, ln := range lines {
    t := strings.TrimSpace(ln)
    if strings.HasPrefix(t, "-") { // new capability
      if current != nil { caps = append(caps, *current) }
      current = &Capability{}
      t = strings.TrimSpace(strings.TrimPrefix(t, "-"))
      if t == "" { continue }
    }
    if current == nil { continue }
    if i := strings.Index(t, ":"); i > 0 {
      k := strings.TrimSpace(t[:i])
      v := strings.Trim(strings.TrimSpace(t[i+1:]), "\"')")
      switch k {
      case "id": current.ID = v
      case "domain": current.Domain = v
      case "description": current.Description = v
      case "maturity": current.Maturity = v
      case "inputSchema": current.InputSchema = v
      case "outputSchema": current.OutputSchema = v
      }
    }
  }
  if current != nil { caps = append(caps, *current) }
  if len(caps) == 0 {
    return nil, errors.New("manifest parse failed: ensure JSON or simple YAML format")
  }
  man.Capabilities = caps
  return &man, nil
}

func findCap(m *Manifest, id string) (Capability, bool) {
  for _, c := range m.Capabilities { if c.ID == id { return c, true } }
  return Capability{}, false
}

func readInput(path string) ([]byte, error) {
  if path == "" { return io.ReadAll(os.Stdin) }
  return os.ReadFile(filepath.Clean(path))
}

func validateJSONAgainstSchema(doc []byte, schemaPath string) error {
  sl := gojsonschema.NewReferenceLoader("file://" + filepath.Clean(schemaPath))
  dl := gojsonschema.NewBytesLoader(doc)
  res, err := gojsonschema.Validate(sl, dl)
  if err != nil { return err }
  if !res.Valid() {
    var b strings.Builder
    for _, e := range res.Errors() { b.WriteString(e.String()); b.WriteString("; ") }
    return errors.New(b.String())
  }
  return nil
}

func fatal(msg string) {
  fmt.Fprintln(os.Stderr, msg)
  os.Exit(1)
}
