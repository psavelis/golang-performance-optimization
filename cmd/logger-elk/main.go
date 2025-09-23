package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "time"
)

func main() {
    host := getenv("LOGSTASH_HOST", "127.0.0.1")
    port := getenvInt("LOGSTASH_PORT", 8080)
    n := flag.Int("n", 100, "number of logs")
    flag.Parse()

    url := fmt.Sprintf("http://%s:%d", host, port)
    client := &http.Client{Timeout: 5 * time.Second}

    for i := 1; i <= *n; i++ {
        rec := map[string]any{
            "message": fmt.Sprintf("elk log %d", i),
            "app":     "interviewing-golang",
            "env":     "dev",
            "seq":     i,
            "ts":      time.Now().Format(time.RFC3339Nano),
        }
        body, _ := json.Marshal(rec)
        resp, err := client.Post(url, "application/json", bytes.NewReader(body))
        if err != nil {
            fmt.Fprintf(os.Stderr, "post error: %v\n", err)
            os.Exit(2)
        }
        _ = resp.Body.Close()
        time.Sleep(10 * time.Millisecond)
    }
    fmt.Printf("emitted %d logs to logstash %s\n", *n, url)
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}
func getenvInt(k string, def int) int {
    if v := os.Getenv(k); v != "" {
        if n, err := strconv.Atoi(v); err == nil { return n }
    }
    return def
}
