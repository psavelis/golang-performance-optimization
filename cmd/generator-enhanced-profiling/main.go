package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/generator"
	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/google/uuid"
)

// Enhanced profiling with block/mutex contention tracking
func main() {
	// Enable block and mutex profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(5)

	// CPU profiling
	cpuProfile, err := os.Create("generator_cpu.prof")
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
		memProfile, err := os.Create("generator_mem.prof")
		if err != nil {
			panic(err)
		}
		defer memProfile.Close()

		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			panic(err)
		}

		// Block profiling
		blockProfile, err := os.Create("generator_block.prof")
		if err != nil {
			panic(err)
		}
		defer blockProfile.Close()

		p := pprof.Lookup("block")
		if err := p.WriteTo(blockProfile, 0); err != nil {
			panic(err)
		}

		// Mutex profiling
		mutexProfile, err := os.Create("generator_mutex.prof")
		if err != nil {
			panic(err)
		}
		defer mutexProfile.Close()

		p = pprof.Lookup("mutex")
		if err := p.WriteTo(mutexProfile, 0); err != nil {
			panic(err)
		}

		// Goroutine profiling
		goroutineProfile, err := os.Create("generator_goroutine.prof")
		if err != nil {
			panic(err)
		}
		defer goroutineProfile.Close()

		p = pprof.Lookup("goroutine")
		if err := p.WriteTo(goroutineProfile, 0); err != nil {
			panic(err)
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

	numEventsStr := os.Args[1]
	numEvents, err := strconv.Atoi(numEventsStr)
	if err != nil {
		panic(fmt.Errorf("unable to parse number of events : %+v", err))
	}

	outputFile := os.Args[2]

	fmt.Printf("number event : %d\n", numEvents)
	fmt.Printf("dump output: %s\n", outputFile)

	ctx := context.Background()

	// Generate events with profile labels
	pprof.Do(ctx, pprof.Labels("stage", "generation"), func(ctx context.Context) {
		generateEvents(ctx, numEvents, outputFile)
	})
}

func generateEvents(ctx context.Context, numEvents int, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("unable to create file : %+v", err))
	}
	defer file.Close()

	pprof.Do(ctx, pprof.Labels("op", "file_creation"), func(ctx context.Context) {
		// Generation logic with labeled sections
		events := make([]*model.Event, 0, numEvents)

		pprof.Do(ctx, pprof.Labels("op", "event_generation"), func(ctx context.Context) {
			for i := 0; i < numEvents; i++ {
				event := generateEvent()
				events = append(events, event)
			}
		})

		pprof.Do(ctx, pprof.Labels("op", "json_marshal"), func(ctx context.Context) {
			content, err := json.Marshal(events)
			if err != nil {
				panic(fmt.Errorf("unable to marshal events : %+v", err))
			}

			if _, err := file.Write(content); err != nil {
				panic(fmt.Errorf("unable to write file : %+v", err))
			}
		})
	})
}

func generateEvent() *model.Event {
	return &model.Event{
		EventSource:     rand.Intn(1000000),
		EventRef:        uuid.New().String(),
		EventType:       generateEventType(),
		EventDate:       generateRandomDate(),
		CallingNumber:   rand.Intn(1000000000),
		CalledNumber:    rand.Intn(1000000000),
		Location:        generator.RandomString(),
		DurationSeconds: rand.Intn(3600),
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
	if r < 15 {
		return 1
	} else if r < 35 {
		return 2
	} else if r < 55 {
		return 3
	} else {
		return 5
	}
}

func generateRandomDate() time.Time {
	year := 2010 + rand.Intn(11)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)
	hour := rand.Intn(24)
	minute := rand.Intn(60)
	second := rand.Intn(60)
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}
