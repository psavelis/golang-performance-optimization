package main

import (
    "flag"
    "fmt"
    "os"
    "strconv"
    "time"

    fluent "github.com/fluent/fluent-logger-golang/fluent"
)

func main() {
    host := getenv("FLUENTD_HOST", "127.0.0.1")
    port := getenvInt("FLUENTD_PORT", 24224)
    tag := getenv("FLUENTD_TAG", "interviewing.golang")
    count := flag.Int("n", 100, "number of log entries to emit")
    flag.Parse()

    logger, err := fluent.New(fluent.Config{FluentHost: host, FluentPort: port, Async: true})
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to connect fluentd: %v\n", err)
        os.Exit(1)
    }
    defer logger.Close()

    for i := 1; i <= *count; i++ {
        data := map[string]interface{}{
            "message": fmt.Sprintf("log %d", i),
            "app":     "interviewing-golang",
            "env":     "dev",
            "seq":     i,
            "ts":      time.Now().Format(time.RFC3339Nano),
        }
        if err := logger.Post(tag, data); err != nil {
            fmt.Fprintf(os.Stderr, "post error: %v\n", err)
            os.Exit(2)
        }
        time.Sleep(10 * time.Millisecond)
    }
    fmt.Printf("emitted %d logs to fluentd %s:%d tag=%s\n", *count, host, port, tag)
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}

func getenvInt(k string, def int) int {
    if v := os.Getenv(k); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    return def
}
