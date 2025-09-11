package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "math/rand"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/dmgo1014/interviewing-golang/pkg/generator"
    "github.com/dmgo1014/interviewing-golang/pkg/model"
    "github.com/google/uuid"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    listenAddr = flag.String("listen", ":2112", "address for metrics HTTP server")

    eventsGenerated = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "events_generated_total",
        Help: "Total number of events generated",
    })
    generateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "generation_duration_seconds",
        Help:    "Time to generate all requested events",
        Buckets: prometheus.DefBuckets,
    })
    inProgress = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "generation_in_progress",
        Help: "Indicator for generation in progress",
    })
)

func init() {
    prometheus.MustRegister(eventsGenerated, generateDuration, inProgress)
}

func main() {
    flag.Parse()

    // Optional env overrides
    if v := os.Getenv("METRICS_ADDR"); v != "" {
        *listenAddr = v
    }
    holdFor := time.Duration(0)
    if v := os.Getenv("METRICS_HOLD_FOR"); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            holdFor = d
        }
    }

    // Start metrics server
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    srv := &http.Server{Addr: *listenAddr, Handler: mux}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("metrics server error: %v\n", err)
        }
    }()

    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        _ = srv.Shutdown(ctx)
        os.Exit(0)
    }()

    if len(os.Args) != 3 {
        panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
    }

    numEvents, err := strconv.Atoi(os.Args[1])
    if err != nil {
        panic(fmt.Errorf("unable to parse number of events : %+v", err))
    }
    out := os.Args[2]

    start := time.Now()
    inProgress.Set(1)
    defer inProgress.Set(0)

    events := make([]*model.Event, 0, numEvents)
    for i := 0; i < numEvents; i++ {
        events = append(events, generateEvent())
    }
    eventsGenerated.Add(float64(numEvents))

    content, err := json.Marshal(events)
    if err != nil {
        panic(fmt.Errorf("unable to marshall events : %+v", err))
    }
    if err := os.WriteFile(out, content, 0o666); err != nil {
        panic(fmt.Errorf("unable to write file : %+v", err))
    }
    generateDuration.Observe(time.Since(start).Seconds())

    fmt.Printf("generated %d events -> %s\n", numEvents, out)

    if holdFor > 0 {
        fmt.Printf("holding metrics endpoint for %s on %s\n", holdFor, *listenAddr)
        time.Sleep(holdFor)
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        _ = srv.Shutdown(ctx)
        cancel()
    }
}

func generateEvent() *model.Event {
    return &model.Event{
        EventSource:     rand.Intn(88005553535),
        EventRef:        uuid.New().String(),
        EventType:       generateEventType(),
        EventDate:       *generator.RandomDate(),
        CallingNumber:   rand.Intn(88005553535),
        CalledNumber:    rand.Intn(88005553535),
        Location:        generator.RandomString(),
        DurationSeconds: rand.Intn(100),
        Attr1:           generator.RandomString(),
        Attr2:           generator.RandomString(),
        Attr3:           generator.RandomString(),
        Attr4:           generator.RandomString(),
        Attr5:           generator.RandomString(),
        Attr6:           generator.RandomString(),
        Attr7:           generator.RandomString(),
        Attr8:           generator.RandomString(),
    }
}

func generateEventType() int {
    r := rand.Intn(100)
    switch {
    case r < 15:
        return 1
    case r < 35:
        return 2
    case r < 55:
        return 3
    default:
        return 5
    }
}
