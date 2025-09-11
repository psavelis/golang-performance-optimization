package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "net/url"
    "strings"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/lib/pq"
    "github.com/dmgo1014/interviewing-golang/pkg/model"
    "github.com/xo/dburl"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pyroscope "github.com/grafana/pyroscope-go"
)

var (
    gitCommit string
    buildTime string
    version   string
)

func main() {
    if len(os.Args) != 3 {
        panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
    }

    tp, shutdown, err := setupTracerProvider("interviewing-golang.loader-otel")
    if err != nil { panic(err) }
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = shutdown(ctx)
    }()

    tracer := tp.Tracer("loader")

    // Graceful exit
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() { <-sigCh; os.Exit(0) }()

    dbURL := os.Args[1]
    inputFile := os.Args[2]

    data, err := os.ReadFile(inputFile)
    if err != nil { panic(fmt.Errorf("read input: %w", err)) }
    var events []*model.Event
    if err := json.Unmarshal(data, &events); err != nil { panic(fmt.Errorf("unmarshal: %w", err)) }

    ctx := context.Background()
    ctx, span := tracer.Start(ctx, "load_events", trace.WithAttributes(
        attribute.Int("events.count", len(events)),
        attribute.String("input.file", inputFile),
        attribute.String("build.git_commit", nonEmpty(gitCommit, "unknown")),
        attribute.String("build.version", nonEmpty(version, getenv("OTEL_SERVICE_VERSION", "0.1.0"))),
        attribute.String("build.time", nonEmpty(buildTime, "unknown")),
    ))

    // Optional Pyroscope integration (env: PYROSCOPE_ENABLE=true)
    var stopProf func()
    if os.Getenv("PYROSCOPE_ENABLE") == "true" {
        stopProf = startPyroscopeWithTrace("interviewing-golang.loader-otel", span)
        defer stopProf()
    }
    defer span.End()

    u, err := dburl.Parse(dbURL)
    if err != nil { span.RecordError(err); panic(fmt.Errorf("parse db url: %w", err)) }
    db, err := sql.Open("postgres", u.DSN)
    if err != nil { span.RecordError(err); panic(fmt.Errorf("connect db: %w", err)) }
    defer db.Close()

    tctx, txSpan := tracer.Start(ctx, "db.transaction")
    tx, err := db.BeginTx(tctx, nil)
    if err != nil { txSpan.RecordError(err); txSpan.SetStatus(codes.Error, err.Error()); panic(err) }
    defer tx.Rollback()

    var skipped int
    for i, e := range events {
        rctx, rspan := tracer.Start(tctx, "db.insert", trace.WithAttributes(
            attribute.Int("row.index", i),
            attribute.String("event.ref", e.EventRef),
        ))
        _ = rctx
        if err := load(tx, e); err != nil {
            // If duplicate, skip to keep demo idempotent
            if isDuplicateKey(err) {
                skipped++
                rspan.AddEvent("duplicate_skipped")
            } else {
                rspan.RecordError(err); rspan.SetStatus(codes.Error, err.Error()); txSpan.RecordError(err); panic(err)
            }
        }
        rspan.End()
    }
    if err := tx.Commit(); err != nil { txSpan.RecordError(err); txSpan.SetStatus(codes.Error, err.Error()); panic(err) }
    txSpan.SetAttributes(attribute.Int("rows.skipped_duplicates", skipped))
    if os.Getenv("PYROSCOPE_TRACE_CORRELATION") == "true" && span.SpanContext().HasTraceID() {
        traceID := span.SpanContext().TraceID().String()
        svc := getenv("OTEL_SERVICE_NAME", "interviewing-golang.loader-otel")
        pyroscopeBase := getenv("PYROSCOPE_SERVER_ADDRESS", "http://localhost:4040")
        link := fmt.Sprintf("%s/?query=service%%3D%s%%20trace_id%%3D%s", strings.TrimRight(pyroscopeBase, "/"), url.QueryEscape(svc), traceID)
        span.AddEvent("pyroscope.link", trace.WithAttributes(
            attribute.String("pyroscope.query_url", link),
            attribute.String("pyroscope.service", svc),
            attribute.String("pyroscope.trace_id", traceID),
        ))
    }
    txSpan.End()
}

func load(tx *sql.Tx, event *model.Event) error {
    q := `
insert into event(event_source, event_ref, event_type, event_date, calling_number, called_number, location,
                  duration_seconds, attr_1, attr_2, attr_3, attr_4, attr_5, attr_6, attr_7, attr_8)
values ($1, $2, $3, %s, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
on conflict (event_source, event_ref) do nothing
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

func timeToTimestampNoTz(t *time.Time) string { return fmt.Sprintf("to_timestamp(cast(%d as bigint))::date", t.Unix()) }

func isDuplicateKey(err error) bool {
    // best-effort check for Postgres duplicate key error text
    if err == nil { return false }
    s := err.Error()
    return strings.Contains(s, "duplicate key value") || strings.Contains(s, "unique constraint")
}

func setupTracerProvider(defaultService string) (*sdktrace.TracerProvider, func(context.Context) error, error) {
    endpoint := normalizeOTLPEndpoint(getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"))
    // Default to insecure for local dev; allow explicit override via env.
    insecureFlag := true
    if v := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); v != "" {
        insecureFlag = v == "true" || v == "1"
    }
    var dialOpts []grpc.DialOption
    if insecureFlag { dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials())) }

    opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
    if insecureFlag {
        opts = append(opts, otlptracegrpc.WithInsecure())
    } else if len(dialOpts) > 0 {
        opts = append(opts, otlptracegrpc.WithDialOption(dialOpts...))
    }
    exp, err := otlptracegrpc.New(context.Background(), opts...)
    if err != nil { return nil, nil, fmt.Errorf("create otlp exporter: %w", err) }

    serviceName := getenv("OTEL_SERVICE_NAME", defaultService)
    res, rerr := resource.New(context.Background(),
        resource.WithFromEnv(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String(getenv("OTEL_SERVICE_VERSION", "0.1.0")),
            attribute.String("deployment.environment", getenv("OTEL_ENV", "dev")),
        ),
    )
    if rerr != nil { return nil, nil, fmt.Errorf("create resource: %w", rerr) }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)
    return tp, tp.Shutdown, nil
}

func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

// normalizeOTLPEndpoint accepts values like "localhost:4317", "127.0.0.1:4317",
// or full URLs like "http://localhost:4317", "grpc://127.0.0.1:4317" and returns host:port.
func normalizeOTLPEndpoint(ep string) string {
    ep = strings.TrimSpace(ep)
    if ep == "" { return "localhost:4317" }
    if strings.Contains(ep, "://") {
        if u, err := url.Parse(ep); err == nil {
            if u.Host != "" { return u.Host }
        }
    }
    ep = strings.TrimPrefix(ep, "http://")
    ep = strings.TrimPrefix(ep, "https://")
    ep = strings.TrimPrefix(ep, "grpc://")
    ep = strings.TrimPrefix(ep, "otel://")
    ep = strings.TrimPrefix(ep, "tcp://")
    ep = strings.Trim(ep, "/")
    return ep
}

// startPyroscopeWithTrace mirrors generator variant for loader.
func startPyroscopeWithTrace(app string, span trace.Span) func() {
    traceCorrelation := os.Getenv("PYROSCOPE_TRACE_CORRELATION") == "true"
    tags := parseTagString(os.Getenv("PYROSCOPE_TAGS"))
    if tags == nil { tags = map[string]string{} }
    if _, ok := tags["service"]; !ok { tags["service"] = app }
    if traceCorrelation && span.SpanContext().HasTraceID() {
        tags["trace_id"] = span.SpanContext().TraceID().String()
        if span.SpanContext().HasSpanID() { tags["root_span_id"] = span.SpanContext().SpanID().String() }
    }
    cfg := pyroscope.Config{
        ApplicationName: app,
        ServerAddress:   getenv("PYROSCOPE_SERVER_ADDRESS", "http://localhost:4040"),
        Tags:            tags,
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocSpace,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileInuseSpace,
            pyroscope.ProfileInuseObjects,
        },
    }
    p, err := pyroscope.Start(cfg)
    if err != nil { fmt.Printf("pyroscope start failed: %v\n", err); return func() {} }
    return func() { _ = p.Stop() }
}

func parseTagString(s string) map[string]string {
    if s == "" { return nil }
    out := map[string]string{}
    parts := strings.Split(s, ",")
    for _, p := range parts {
        kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
        if len(kv) == 2 && kv[0] != "" && kv[1] != "" { out[kv[0]] = kv[1] }
    }
    if len(out) == 0 { return nil }
    return out
}

func nonEmpty(v, fallback string) string { if v == "" { return fallback }; return v }
