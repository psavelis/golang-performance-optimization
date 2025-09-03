package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/dmgo1014/interviewing-golang.git/pkg/generator"
	"github.com/dmgo1014/interviewing-golang.git/pkg/model"
	"github.com/google/uuid"
)

// Profiling-enabled version of the generator
func main() {
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
	}()

	// log time duration on application shutdown
	start := time.Now()
	defer func() {
		fmt.Println("================")
		fmt.Printf("Execution Time : %v\n", time.Since(start))
	}()

	// validate inputs firstly
	if len(os.Args) != 3 {
		panic(fmt.Errorf("invalid number of arguments, 2 expected, got %d", len(os.Args)-1))
	}

	// number of events is the first argument
	numEventsStr := os.Args[1]
	numEvents, err := strconv.Atoi(numEventsStr)
	if err != nil {
		panic(fmt.Errorf("unable to parse number of events : %+v", err))
	}

	// output file is the second
	outPutFile := os.Args[2]

	fmt.Printf("number event : %d\n", numEvents)
	fmt.Printf("dump output: %s\n", outPutFile)

	// act - this is the current inefficient implementation for profiling
	events := []*model.Event{}

	// generate requested number of events
	for i := 0; i < numEvents; i++ {
		events = append(events, generateEvent())
	}

	// marshall for saving
	content, err := json.Marshal(events)
	if err != nil {
		panic(fmt.Errorf("unable to marshall events : %+v", err))
	}

	// and write everything
	err = os.WriteFile(outPutFile, content, 0777)
	if err != nil {
		panic(fmt.Errorf("unable to write file : %+v", err))
	}
}

// generateEvent creates and returns a new instance of model.Event populated with random values for all its fields.
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

// generateEventType generates a random event type (integer) based on predefined probability distributions:
// - Type 1: 15%
// - Type 2: 20%
// - Type 3: 20%
// - Type 5: 45%
func generateEventType() int {
	r := rand.Intn(100)

	if r < 15 {
		return 1
	} else if r < 35 {
		return 2
	} else if r < 55 {
		return 3
	}
	return 5
}
