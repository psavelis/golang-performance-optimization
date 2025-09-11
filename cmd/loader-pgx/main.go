package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgx-based loader with COPY FROM for high-performance bulk inserts
func main() {
	// Enable enhanced profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(5)

	// CPU profiling
	cpuProfile, err := os.Create("loader_pgx_cpu.prof")
	if err != nil {
		panic(err)
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	// Memory profiling
	defer func() {
		profiles := []string{"mem", "block", "mutex", "goroutine"}
		for _, profileType := range profiles {
			file, err := os.Create(fmt.Sprintf("loader_pgx_%s.prof", profileType))
			if err != nil {
				continue
			}
			defer file.Close()

			var p *pprof.Profile
			switch profileType {
			case "mem":
				err = pprof.WriteHeapProfile(file)
				continue
			default:
				p = pprof.Lookup(profileType)
			}

			if p != nil {
				p.WriteTo(file, 0)
			}
		}
	}()

	start := time.Now()
	defer func() {
		fmt.Println("================")
		fmt.Printf("Execution Time : %v\n", time.Since(start))
	}()

	if len(os.Args) != 3 {
		panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
	}

	dbUrl := os.Args[1]
	inputFile := os.Args[2]

	fmt.Printf("input file: %s\n", inputFile)

	ctx := context.Background()

	// Connect to database with pgxpool
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		panic(fmt.Errorf("unable to parse database config: %+v", err))
	}

	// Optimize pool settings for bulk operations
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(fmt.Errorf("unable to connect to database: %+v", err))
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Errorf("unable to ping database: %+v", err))
	}

	// Process events with profile labels
	pprof.Do(ctx, pprof.Labels("stage", "pgx_loading"), func(ctx context.Context) {
		totalProcessed, err := processEventsWithCopy(ctx, pool, inputFile)
		if err != nil {
			panic(fmt.Errorf("unable to process events: %+v", err))
		}
		fmt.Printf("successfully loaded %d events\n", totalProcessed)
	})
}

func processEventsWithCopy(ctx context.Context, pool *pgxpool.Pool, inputFile string) (int, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var events []*model.Event

	pprof.Do(ctx, pprof.Labels("op", "json_decode"), func(ctx context.Context) {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&events); err != nil {
			panic(fmt.Errorf("unable to decode JSON: %+v", err))
		}
	})

	totalProcessed := 0
	const batchSize = 10000

	pprof.Do(ctx, pprof.Labels("op", "bulk_copy"), func(ctx context.Context) {
		// Process in batches using COPY FROM
		for i := 0; i < len(events); i += batchSize {
			end := i + batchSize
			if end > len(events) {
				end = len(events)
			}

			batch := events[i:end]
			if err := copyBatch(ctx, pool, batch); err != nil {
				panic(fmt.Errorf("failed to copy batch: %+v", err))
			}

			totalProcessed += len(batch)

			// Progress reporting
			if totalProcessed%50000 == 0 {
				fmt.Printf("Processed %d events\n", totalProcessed)
			}
		}
	})

	return totalProcessed, nil
}

func copyBatch(ctx context.Context, pool *pgxpool.Pool, events []*model.Event) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %+v", err)
	}
	defer conn.Release()

	// Use COPY FROM for maximum throughput
	_, err = conn.Conn().CopyFrom(
		ctx,
		pgx.Identifier{"event"},
		[]string{
			"event_source", "event_ref", "event_type", "event_date",
			"calling_number", "called_number", "location", "duration_seconds",
			"attr_1", "attr_2", "attr_3", "attr_4", "attr_5", "attr_6", "attr_7", "attr_8",
		},
		&eventCopySource{events: events, index: 0},
	)

	if err != nil {
		return fmt.Errorf("copy from failed: %+v", err)
	}

	return nil
}

// eventCopySource implements pgx.CopyFromSource for streaming data to COPY FROM
type eventCopySource struct {
	events []*model.Event
	index  int
}

func (e *eventCopySource) Next() bool {
	return e.index < len(e.events)
}

func (e *eventCopySource) Values() ([]interface{}, error) {
	if e.index >= len(e.events) {
		return nil, nil
	}

	event := e.events[e.index]
	e.index++

	return []interface{}{
		event.EventSource,
		event.EventRef,
		event.EventType,
		event.EventDate,
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
	}, nil
}

func (e *eventCopySource) Err() error {
	return nil
}
