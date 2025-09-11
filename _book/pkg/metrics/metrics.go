package metrics

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/collectors"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    regOnce sync.Once
)

// StartFromEnv starts a Prometheus metrics HTTP server if enabled via env.
// Environment variables:
//   METRICS_ENABLED: "true"/"1" to force enable. Defaults to disabled unless METRICS_ADDR is set.
//   METRICS_ADDR: address to bind, e.g. ":2112". Required if not using METRICS_ENABLED.
//   METRICS_PATH: HTTP path, defaults to "/metrics".
// Returns a no-op stop func if not enabled.
func StartFromEnv() func() {
    enabled := isTrue(os.Getenv("METRICS_ENABLED"))
    addr := strings.TrimSpace(os.Getenv("METRICS_ADDR"))
    path := os.Getenv("METRICS_PATH")
    if path == "" { path = "/metrics" }

    if !enabled && addr == "" {
        return func() {}
    }
    if addr == "" {
        addr = ":2112"
    }

    // Register standard collectors once on the default registry.
    regOnce.Do(func() {
        prometheus.MustRegister(
            collectors.NewGoCollector(),
            collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
        )
    })

    mux := http.NewServeMux()
    mux.Handle(path, promhttp.Handler())
    srv := &http.Server{Addr: addr, Handler: mux}

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("metrics server error: %v\n", err)
        }
    }()

    return func() {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        _ = srv.Shutdown(ctx)
    }
}

// HoldFromEnv blocks for METRICS_HOLD_FOR duration if set (e.g., "60s").
// Useful for short-lived CLIs to keep the metrics endpoint available for scraping.
func HoldFromEnv() {
    v := strings.TrimSpace(os.Getenv("METRICS_HOLD_FOR"))
    if v == "" { return }
    if d, err := time.ParseDuration(v); err == nil && d > 0 {
        time.Sleep(d)
    }
}

func isTrue(s string) bool {
    if s == "" { return false }
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "1", "true", "yes", "y", "on":
        return true
    default:
        return false
    }
}
