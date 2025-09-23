package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/dmgo1014/interviewing-golang/pkg/model"
    "github.com/xo/dburl"
    pyroscope "github.com/grafana/pyroscope-go"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"

    _ "github.com/lib/pq"
)

// Loader with Grafana Pyroscope continuous profiling.
// Args:
// 1) DB URL
// 2) input file path (JSON array of Event)
func main() {
    start := time.Now()

    // graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    stop := startPyroscope("interviewing-golang.loader")
    defer stop()

    go func() {
        <-sigCh
        log.Println("received shutdown signal; exiting...")
        os.Exit(0)
    }()

    if len(os.Args) != 3 {
        panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
    }

    dbUrl := os.Args[1]
    url, err := dburl.Parse(dbUrl)
    if err != nil {
        panic(fmt.Errorf("unable to parse database URL '%s' : %+v", url, err))
    }

    inputFile := os.Args[2]
    fmt.Printf("input file: %s\n", inputFile)

    eventRaw, err := os.ReadFile(inputFile)
    if err != nil {
        panic(fmt.Errorf("unable to read input file : %+v", err))
    }
    var events []*model.Event
    if err := json.Unmarshal(eventRaw, &events); err != nil {
        panic(fmt.Errorf("unable to unmarshall event file content : %+v", err))
    }
    fmt.Printf("Total events to load : %d\n", len(events))

    db, err := sql.Open("postgres", url.DSN)
    if err != nil {
        panic(fmt.Errorf("unable to connecto to database : %+v", err))
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        panic(fmt.Errorf("unable to start transaction : %+v", err))
    }
    defer tx.Rollback()

    // Optional trace correlation root span (env: PYROSCOPE_TRACE_CORRELATION=true)
    if os.Getenv("PYROSCOPE_TRACE_CORRELATION") == "true" {
        tracer := otel.GetTracerProvider().Tracer("loader-pyroscope")
        ctx, span := tracer.Start(context.Background(), "loader.run",
            trace.WithAttributes(attribute.Int("events.count", len(events))))
        if span.SpanContext().HasTraceID() && span.SpanContext().HasSpanID() {
            appendDynamicTraceTagsLoader(span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
        }
        defer span.End()
        _ = ctx
    }

    for _, e := range events {
        if err := load(tx, e); err != nil {
            panic(fmt.Errorf("unable to load event : %+v", err))
        }
    }
    if err := tx.Commit(); err != nil {
        panic(fmt.Errorf("unable to commit transaction : %+v", err))
    }

    fmt.Printf("sucessfully loaded %d events\n", len(events))
    fmt.Println("================")
    fmt.Printf("Execution Time : %v\n", time.Since(start))
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

// Shared Pyroscope bootstrap with env-driven config.
func startPyroscope(defaultApp string) func() {
    serverAddr := getenvDefault("PYROSCOPE_SERVER_ADDRESS", "http://localhost:4040")
    appName := getenvDefault("PYROSCOPE_APPLICATION_NAME", defaultApp)
    tenantID := os.Getenv("PYROSCOPE_TENANT_ID")
    tags := parseTags(os.Getenv("PYROSCOPE_TAGS"))

    cfg := pyroscope.Config{
        ApplicationName: appName,
        ServerAddress:   serverAddr,
        TenantID:        tenantID,
        Tags:            tags,
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileAllocSpace,
            pyroscope.ProfileInuseObjects,
            pyroscope.ProfileInuseSpace,
            pyroscope.ProfileGoroutines,
            pyroscope.ProfileMutexCount,
            pyroscope.ProfileMutexDuration,
            pyroscope.ProfileBlockCount,
            pyroscope.ProfileBlockDuration,
        },
    }

    p, err := pyroscope.Start(cfg)
    if err != nil {
        log.Printf("pyroscope: failed to start profiler: %v", err)
        return func() {}
    }
    return func() { _ = p.Stop() }
}

func getenvDefault(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}

func parseTags(s string) map[string]string {
    if s == "" {
        return map[string]string{"service": "loader"}
    }
    out := map[string]string{}
    parts := strings.Split(s, ",")
    for _, p := range parts {
        kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
        if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
            out[kv[0]] = kv[1]
        }
    }
    if len(out) == 0 {
        out["service"] = "loader"
    }
    return out
}

func appendDynamicTraceTagsLoader(traceID, spanID string) {
    existing := os.Getenv("PYROSCOPE_TAGS")
    if existing == "" {
        os.Setenv("PYROSCOPE_TAGS", fmt.Sprintf("trace_id=%s,root_span_id=%s", traceID, spanID))
        return
    }
    if !strings.Contains(existing, "trace_id=") {
        existing += fmt.Sprintf(",trace_id=%s", traceID)
    }
    if !strings.Contains(existing, "root_span_id=") {
        existing += fmt.Sprintf(",root_span_id=%s", spanID)
    }
    os.Setenv("PYROSCOPE_TAGS", existing)
}
