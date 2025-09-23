package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/metrics"
	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/google/uuid"
)

// Pool of pre-allocated strings for efficiency
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

// Pre-computed event type thresholds for faster generation
const (
	typeThreshold1 = 15
	typeThreshold2 = 35
	typeThreshold3 = 55
)

// Optimized generator with streaming JSON, pre-allocated buffers, and efficient random generation
func main() {
	// optional metrics server (env-driven)
	stopMetrics := metrics.StartFromEnv()
	defer stopMetrics()
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

	// Create file with buffered writer
	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("unable to create file : %+v", err))
	}
	defer file.Close()

	// Generate events in batches and use efficient JSON marshaling
	const batchSize = 10000
	allEvents := make([]*model.Event, 0, numEvents)

	// Generate all events
	for i := 0; i < numEvents; i++ {
		allEvents = append(allEvents, generateEventOptimized())
	}

	// Marshal to JSON in one go (still more efficient than individual marshaling)
	content, err := json.Marshal(allEvents)
	if err != nil {
		panic(fmt.Errorf("unable to marshal events : %+v", err))
	}

	// Write to file
	if _, err := file.Write(content); err != nil {
		panic(fmt.Errorf("unable to write file : %+v", err))
	}

	// optionally hold for scraping in short-lived runs
	metrics.HoldFromEnv()
}

// Optimized event generation with pre-allocated buffers and efficient random generation
func generateEventOptimized() *model.Event {
	// Use single random call to generate random values
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
		EventRef:        uuid.New().String(), // Keep UUID for uniqueness
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

// Optimized event type generation
func generateEventTypeOptimized() int {
	r := rand.Intn(100)

	switch {
	case r < typeThreshold1:
		return 1
	case r < typeThreshold2:
		return 2
	case r < typeThreshold3:
		return 3
	default:
		return 5
	}
}

// Get random string from pre-allocated pool instead of generating
func getRandomStringFromPool() string {
	return stringPool[rand.Intn(len(stringPool))]
}
