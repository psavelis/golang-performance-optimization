package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/google/uuid"
)

// String pool for efficient string generation
var stringPool = []string{
	"ABC123", "XYZ456", "LOC789", "MNO012", "PQR345", "STU678", "VWX901", "DEF234",
	"GHI567", "JKL890", "ABC123456", "XYZ789012", "LOC345678", "MNO901234", "PQR567890",
	"STU123456", "VWX789012", "DEF345678", "GHI901234", "JKL567890", "ABC456789",
	"XYZ123456", "LOC789012", "MNO345678", "PQR901234", "STU456789", "VWX012345",
	"DEF678901", "GHI234567", "JKL890123", "MNO456789", "PQR012345", "STU678901",
	"VWX234567", "DEF890123", "GHI456789", "JKL012345", "ABC789012", "XYZ345678",
	"LOC901234", "MNO567890", "PQR123456", "STU789012", "VWX345678", "DEF901234",
	"GHI567890", "JKL123456", "ABC890123", "XYZ456789", "LOC012345", "MNO678901",
}

// Streaming JSON generator with constant memory usage
func main() {
	// Enable enhanced profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(5)

	// CPU profiling
	cpuProfile, err := os.Create("generator_streaming_cpu.prof")
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
			file, err := os.Create(fmt.Sprintf("generator_streaming_%s.prof", profileType))
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

	numEventsStr := os.Args[1]
	numEvents, err := strconv.Atoi(numEventsStr)
	if err != nil {
		panic(fmt.Errorf("unable to parse number of events : %+v", err))
	}

	outputFile := os.Args[2]

	fmt.Printf("number event : %d\n", numEvents)
	fmt.Printf("dump output: %s\n", outputFile)

	ctx := context.Background()

	// Stream events with profile labels
	pprof.Do(ctx, pprof.Labels("stage", "streaming_generation"), func(ctx context.Context) {
		streamEvents(ctx, numEvents, outputFile)
	})
}

func streamEvents(ctx context.Context, numEvents int, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("unable to create file : %+v", err))
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	pprof.Do(ctx, pprof.Labels("op", "json_streaming"), func(ctx context.Context) {
		// Write JSON array start
		writer.WriteString("[")

		for i := 0; i < numEvents; i++ {
			if i > 0 {
				writer.WriteString(",")
			}

			event := generateEventOptimized()

			// Stream each event directly to JSON
			eventJSON, err := json.Marshal(event)
			if err != nil {
				panic(fmt.Errorf("unable to marshal event : %+v", err))
			}

			writer.Write(eventJSON)

			// Flush periodically to maintain constant memory
			if i%1000 == 0 {
				writer.Flush()
			}
		}

		// Write JSON array end
		writer.WriteString("]")
	})
}

func generateEventOptimized() *model.Event {
	// Use single random call to generate multiple values
	r1 := rand.Uint64()
	r2 := rand.Uint64()
	r3 := rand.Uint64()

	// Extract different parts of the random values
	eventSource := int(r1 & 0xFFFFFFFF)
	callingNumber := int((r1 >> 32) & 0xFFFFFFFF)
	calledNumber := int(r2 & 0xFFFFFFFF)
	durationSeconds := int((r2 >> 32) & 0x7F) // Max 127

	// Generate random date more efficiently
	year := 2010 + int((r3&0xFF)%11)
	month := 1 + int(((r3>>8)&0xFF)%12)
	day := 1 + int(((r3>>16)&0xFF)%28)
	hour := int(((r3 >> 24) & 0xFF) % 24)
	minute := int(((r3 >> 32) & 0xFF) % 60)
	second := int(((r3 >> 40) & 0xFF) % 60)

	eventDate := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	return &model.Event{
		EventSource:     eventSource,
		EventRef:        uuid.New().String(),
		EventType:       generateEventTypeOptimized(),
		EventDate:       eventDate,
		CallingNumber:   callingNumber,
		CalledNumber:    calledNumber,
		Location:        getRandomStringFromPool(),
		DurationSeconds: durationSeconds,
		Attr1:           getRandomStringFromPool(),
		Attr2:           getRandomStringFromPool(),
		Attr3:           getRandomStringFromPool(),
		Attr4:           getRandomStringFromPool(),
		Attr5:           getRandomStringFromPool(),
		Attr6:           getRandomStringFromPool(),
		Attr7:           getRandomStringFromPool(),
		Attr8:           getRandomStringFromPool(),
	}
}

func generateEventTypeOptimized() int {
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

func getRandomStringFromPool() string {
	return stringPool[rand.Intn(len(stringPool))]
}
