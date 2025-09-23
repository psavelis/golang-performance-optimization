package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/lib/pq"
    "github.com/dmgo1014/interviewing-golang/pkg/model"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/xo/dburl"
)

var (
    loadEventsTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: "interviewing_golang",
        Subsystem: "loader",
        Name:      "events_loaded_total",
        Help:      "Total number of events loaded to DB.",
    })
    loadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Namespace: "interviewing_golang",
        Subsystem: "loader",
        Name:      "load_duration_seconds",
        Help:      "Duration of full load runs in seconds.",
        Buckets:   prometheus.DefBuckets,
    })
    txDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Namespace: "interviewing_golang",
        Subsystem: "loader",
        Name:      "transaction_duration_seconds",
        Help:      "Duration of the DB transaction in seconds.",
        Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
    })
)

func main() {
    prometheus.MustRegister(loadEventsTotal, loadDuration, txDuration)

    if len(os.Args) != 3 {
        panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
    }

    dbURL := os.Args[1]
    inputFile := os.Args[2]

    metricsAddr := getenv("METRICS_ADDR", ":2113")
    holdFor := getenvDuration("METRICS_HOLD_FOR", 60*time.Second)

    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    srv := &http.Server{Addr: metricsAddr, Handler: mux}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("metrics server error: %v", err)
        }
    }()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        _ = srv.Shutdown(ctx)
        os.Exit(0)
    }()

    start := time.Now()
    fmt.Printf("input file: %s\n", inputFile)

    data, err := os.ReadFile(inputFile)
    if err != nil {
        panic(fmt.Errorf("unable to read input file : %+v", err))
    }
    var events []*model.Event
    if err := json.Unmarshal(data, &events); err != nil {
        panic(fmt.Errorf("unable to unmarshall event file content : %+v", err))
    }
    fmt.Printf("Total events to load : %d\n", len(events))

    url, err := dburl.Parse(dbURL)
    if err != nil {
        panic(fmt.Errorf("unable to parse database URL '%s' : %+v", dbURL, err))
    }
    db, err := sql.Open("postgres", url.DSN)
    if err != nil {
        panic(fmt.Errorf("unable to connect to database : %+v", err))
    }
    defer db.Close()

    txStart := time.Now()
    tx, err := db.Begin()
    if err != nil {
        panic(fmt.Errorf("unable to start transaction : %+v", err))
    }
    defer tx.Rollback()

    for _, e := range events {
        if err := load(tx, e); err != nil {
            panic(fmt.Errorf("unable to load event : %+v", err))
        }
    }
    if err := tx.Commit(); err != nil {
        panic(fmt.Errorf("unable to commit transaction : %+v", err))
    }
    txDuration.Observe(time.Since(txStart).Seconds())

    loadEventsTotal.Add(float64(len(events)))
    loadDuration.Observe(time.Since(start).Seconds())
    fmt.Printf("successfully loaded %d events\n", len(events))

    if holdFor > 0 {
        fmt.Printf("holding metrics endpoint for %s on %s\n", holdFor, metricsAddr)
        time.Sleep(holdFor)
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        _ = srv.Shutdown(ctx)
        cancel()
    }
}

func load(tx *sql.Tx, event *model.Event) error {
    q := `
insert into event(event_source, event_ref, event_type, event_date, calling_number, called_number, location,
                  duration_seconds, attr_1, attr_2, attr_3, attr_4, attr_5, attr_6, attr_7, attr_8)
values ($1, $2, $3, %s, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`
    q = fmt.Sprintf(q, timeToTimestampNoTz(&event.EventDate))

    _, err := tx.Exec(q,
        event.EventSource,
        event.EventRef,
        event.EventType,
        event.CallingNumber,
        event.CalledNumber,
        event.Location,
        event.DurationSeconds,
        event.Attr1,
        event.Attr2,
        event.Attr3,
        event.Attr4,
        event.Attr5,
        event.Attr6,
        event.Attr7,
        event.Attr8,
    )
    return err
}

func timeToTimestampNoTz(t *time.Time) string {
    return fmt.Sprintf("to_timestamp(cast(%d as bigint))::date", t.Unix())
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}

func getenvDuration(k string, def time.Duration) time.Duration {
    if v := os.Getenv(k); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return def
}
