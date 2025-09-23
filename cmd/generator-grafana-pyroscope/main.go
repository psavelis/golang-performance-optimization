package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/generator"
	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/google/uuid"
	pyroscope "github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Production-ready generator integrated with Grafana Pyroscope for continuous profiling.
// Args:
// 1) number of events to generate
// 2) output file path
func main() {
	start := time.Now()
	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	profStop := startPyroscope("interviewing-golang.generator")
	defer profStop()

	go func() {
		<-sigCh
		log.Println("received shutdown signal; exiting...")
		os.Exit(0)
	}()

	if len(os.Args) != 3 {
		panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
	}

	// number of events
	numEvents, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("unable to parse number of events : %+v", err))
	}

	outPutFile := os.Args[2]

	fmt.Printf("number event : %d\n", numEvents)
	fmt.Printf("dump output: %s\n", outPutFile)

	events := make([]*model.Event, 0, numEvents)
	// Optional OTel root span for correlation (set PYROSCOPE_TRACE_CORRELATION=true)
	if os.Getenv("PYROSCOPE_TRACE_CORRELATION") == "true" {
		tracer := otel.GetTracerProvider().Tracer("generator-pyroscope")
		ctx, span := tracer.Start(context.Background(), "generator.run", trace.WithAttributes(attribute.Int("events.requested", numEvents)))
		if span.SpanContext().HasTraceID() && span.SpanContext().HasSpanID() {
			appendDynamicTraceTags(span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
		}
		defer span.End()
		_ = ctx
	}
	for i := 0; i < numEvents; i++ {
		events = append(events, generateEvent())
	}

	content, err := json.Marshal(events)
	if err != nil {
		panic(fmt.Errorf("unable to marshall events : %+v", err))
	}

	if err := os.WriteFile(outPutFile, content, 0o666); err != nil {
		panic(fmt.Errorf("unable to write file : %+v", err))
	}

	fmt.Println("================")
	fmt.Printf("Execution Time : %v\n", time.Since(start))
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

// startPyroscope starts the Grafana Pyroscope Go profiler with sane production defaults.
// Config is driven by environment variables, with safe fallbacks:
//   - PYROSCOPE_SERVER_ADDRESS (default: http://localhost:4040)
//   - PYROSCOPE_APPLICATION_NAME (default: provided defaultApp)
//   - PYROSCOPE_TENANT_ID (optional)
//   - PYROSCOPE_TAGS (comma-separated key=value list)
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

// appendDynamicTraceTags mutates the static tag map once to include trace/span identifiers for correlation.
// It keeps tag cardinality low by only adding the root IDs of the run.
func appendDynamicTraceTags(traceID, spanID string) {
	// Exposed via env override: PYROSCOPE_TAGS="..."; we append if not already present.
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

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseTags(s string) map[string]string {
	if s == "" {
		return map[string]string{"service": "generator"}
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
		out["service"] = "generator"
	}
	return out
}
