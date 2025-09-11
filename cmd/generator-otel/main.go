package main

import (
    "context"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/url"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"

    "github.com/dmgo1014/interviewing-golang/pkg/generator"
    "github.com/dmgo1014/interviewing-golang/pkg/model"
    "github.com/google/uuid"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
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

// Env configuration keys (aligned with industry conventions)
// - OTEL_SERVICE_NAME (overrides service.name resource attribute)
// - OTEL_EXPORTER_OTLP_ENDPOINT (eg: "localhost:4317")
// - OTEL_EXPORTER_OTLP_INSECURE ("true" to disable TLS)
// - OTEL_RESOURCE_ATTRIBUTES (comma-separated key=value)
// - OTEL_TRACES_SAMPLER (parentbased_always_on, parentbased_traceidratio, etc.)

func main() {
    if len(os.Args) != 3 {
        panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
    }

    // Initialize OpenTelemetry
    tp, shutdown, err := setupTracerProvider("interviewing-golang.generator-otel")
    if err != nil {
        panic(err)
    }
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = shutdown(ctx)
    }()

    tracer := tp.Tracer("generator")

    // Graceful exit
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() { <-sigCh; os.Exit(0) }()

    // Args
    numEvents, err := strconv.Atoi(os.Args[1])
    if err != nil { panic(fmt.Errorf("unable to parse number of events : %+v", err)) }
    outFile := os.Args[2]

    // Root span first (so we can correlate trace/span IDs into Pyroscope tags if enabled)
    ctx := context.Background()
    ctx, span := tracer.Start(ctx, "generate_events", trace.WithAttributes(
        attribute.Int("events.requested", numEvents),
        attribute.String("build.git_commit", nonEmpty(gitCommit, "unknown")),
        attribute.String("build.version", nonEmpty(version, getenv("OTEL_SERVICE_VERSION", "0.1.0"))),
        attribute.String("build.time", nonEmpty(buildTime, "unknown")),
    ))

    // Optional Pyroscope integration for OTel binary (env: PYROSCOPE_ENABLE=true)
    var stopProf func()
    if os.Getenv("PYROSCOPE_ENABLE") == "true" {
        stopProf = startPyroscopeWithTrace("interviewing-golang.generator-otel", span)
        defer stopProf()
    }
    start := time.Now()

    events := make([]*model.Event, 0, numEvents)

    // Batch spans to avoid span storms in demos
    const batch = 1000
    for i := 0; i < numEvents; i += batch {
        end := i + batch
        if end > numEvents {
            end = numEvents
        }
        bctx, bspan := tracer.Start(ctx, "generate_batch",
            trace.WithAttributes(
                attribute.Int("batch.start_index", i),
                attribute.Int("batch.end_index", end),
            ),
        )
        _ = bctx
        for j := i; j < end; j++ {
            events = append(events, generateEvent())
        }
        bspan.End()
    }

    content, err := json.Marshal(events)
    if err != nil {
        span.RecordError(err)
        span.End()
        panic(fmt.Errorf("unable to marshall events : %+v", err))
    }
    if err := os.WriteFile(outFile, content, 0o666); err != nil {
        span.RecordError(err)
        span.End()
        panic(fmt.Errorf("unable to write file : %+v", err))
    }

    span.SetAttributes(
        attribute.Int("events.generated", numEvents),
        attribute.String("output.file", outFile),
        attribute.Float64("duration.ms", float64(time.Since(start).Milliseconds())),
    )
    if os.Getenv("PYROSCOPE_TRACE_CORRELATION") == "true" && span.SpanContext().HasTraceID() {
        traceID := span.SpanContext().TraceID().String()
        svc := getenv("OTEL_SERVICE_NAME", "interviewing-golang.generator-otel")
        pyroscopeBase := getenv("PYROSCOPE_SERVER_ADDRESS", "http://localhost:4040")
        link := fmt.Sprintf("%s/?query=service%%3D%s%%20trace_id%%3D%s", strings.TrimRight(pyroscopeBase, "/"), url.QueryEscape(svc), traceID)
        span.AddEvent("pyroscope.link", trace.WithAttributes(
            attribute.String("pyroscope.query_url", link),
            attribute.String("pyroscope.service", svc),
            attribute.String("pyroscope.trace_id", traceID),
        ))
    }
    span.End()

    fmt.Printf("generated %d events -> %s\n", numEvents, outFile)
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

// startPyroscopeWithTrace starts Pyroscope profiler and injects trace/span IDs into tags when correlation env flag is set.
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

func setupTracerProvider(defaultService string) (*sdktrace.TracerProvider, func(context.Context) error, error) {
    endpoint := normalizeOTLPEndpoint(getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"))
    insecureFlag := true
    if v := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); v != "" {
        insecureFlag = v == "true" || v == "1"
    }

    var dialOpts []grpc.DialOption
    if insecureFlag {
        dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    }

    opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
    if insecureFlag {
        opts = append(opts, otlptracegrpc.WithInsecure())
    } else if len(dialOpts) > 0 {
        opts = append(opts, otlptracegrpc.WithDialOption(dialOpts...))
    }
    exp, err := otlptracegrpc.New(context.Background(), opts...)
    if err != nil {
        return nil, nil, fmt.Errorf("create otlp exporter: %w", err)
    }

    // Resource: service name and common attrs
    serviceName := getenv("OTEL_SERVICE_NAME", defaultService)
    res, rerr := resource.New(context.Background(),
        resource.WithFromEnv(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String(getenv("OTEL_SERVICE_VERSION", "0.1.0")),
            attribute.String("deployment.environment", getenv("OTEL_ENV", "dev")),
            attribute.String("build.git_commit", nonEmpty(gitCommit, "unknown")),
            attribute.String("build.time", nonEmpty(buildTime, "unknown")),
        ),
    )
    if rerr != nil {
        return nil, nil, fmt.Errorf("create resource: %w", rerr)
    }

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

func nonEmpty(v, fallback string) string { if v == "" { return fallback }; return v }
