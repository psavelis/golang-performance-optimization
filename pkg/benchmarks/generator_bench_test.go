package benchmarks

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/generator"
	"github.com/dmgo1014/interviewing-golang/pkg/model"
	"github.com/google/uuid"
)

// String pool for optimized string generation
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

// Event types
const (
	EventTypeCall = iota
	EventTypeSMS
	EventTypeData
	EventTypeRoaming
	EventTypeMMS
	EventTypeVoicemail
)

// Utility functions for benchmarks
func randomInt() int {
	return rand.Intn(999999999) + 100000000
}

func randomEventType() int {
	return rand.Intn(6) // 0-5 for different event types
}

func randomDate() time.Time {
	return time.Date(2023, time.Month(rand.Intn(12)+1), rand.Intn(28)+1,
		rand.Intn(24), rand.Intn(60), rand.Intn(60), 0, time.UTC)
}

func generateRandomStringOriginal() string {
	return generator.RandomString()
}

func getRandomStringFromPool() string {
	return stringPool[rand.Intn(len(stringPool))]
}

func createTestEvent() model.Event {
	return model.Event{
		EventSource:     randomInt(),
		EventRef:        uuid.New().String(),
		EventType:       randomEventType(),
		EventDate:       randomDate(),
		CallingNumber:   randomInt(),
		CalledNumber:    randomInt(),
		Location:        generateRandomStringOriginal(),
		DurationSeconds: rand.Intn(3600),
		Attr1:           generateRandomStringOriginal(),
		Attr2:           generateRandomStringOriginal(),
		Attr3:           generateRandomStringOriginal(),
		Attr4:           generateRandomStringOriginal(),
		Attr5:           generateRandomStringOriginal(),
		Attr6:           generateRandomStringOriginal(),
		Attr7:           generateRandomStringOriginal(),
		Attr8:           generateRandomStringOriginal(),
	}
}

func createOptimizedEvent() model.Event {
	return model.Event{
		EventSource:     randomInt(),
		EventRef:        uuid.New().String(),
		EventType:       randomEventType(),
		EventDate:       randomDate(),
		CallingNumber:   randomInt(),
		CalledNumber:    randomInt(),
		Location:        getRandomStringFromPool(),
		DurationSeconds: rand.Intn(3600),
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

func streamEventsToFile(file *os.File, count int) error {
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString("[")
	for i := 0; i < count; i++ {
		if i > 0 {
			writer.WriteString(",")
		}

		event := createTestEvent()
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		writer.Write(data)

		if i%100 == 0 {
			writer.Flush()
		}
	}
	writer.WriteString("]")
	return nil
}

// Benchmark original generator (accumulates in memory)
func BenchmarkGeneratorOriginal(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		file, err := os.CreateTemp("", "bench_original_*.json")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		// Simulate original approach - accumulate all events in memory
		var events []model.Event
		for j := 0; j < 1000; j++ {
			events = append(events, createTestEvent())
		}

		// Marshal all at once
		data, err := json.Marshal(events)
		if err != nil {
			b.Fatal(err)
		}

		file.Write(data)
	}
}

// Benchmark optimized generator (uses string pool)
func BenchmarkGeneratorOptimized(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		file, err := os.CreateTemp("", "bench_optimized_*.json")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		// Use optimized approach with string pool
		var events []model.Event
		for j := 0; j < 1000; j++ {
			events = append(events, createOptimizedEvent())
		}

		data, err := json.Marshal(events)
		if err != nil {
			b.Fatal(err)
		}

		file.Write(data)
	}
}

// Benchmark streaming generator (constant memory)
func BenchmarkGeneratorStreaming(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		file, err := os.CreateTemp("", "bench_streaming_*.json")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		streamEventsToFile(file, 1000)
	}
}

// String generation benchmarks
func BenchmarkStringGenerationOriginal(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = generateRandomStringOriginal()
	}
}

func BenchmarkStringGenerationPool(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getRandomStringFromPool()
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("OriginalApproach", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var events []model.Event
			for j := 0; j < 100; j++ {
				events = append(events, createTestEvent())
			}
			runtime.KeepAlive(events)
		}
	})

	b.Run("OptimizedApproach", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var events []model.Event
			for j := 0; j < 100; j++ {
				events = append(events, createOptimizedEvent())
			}
			runtime.KeepAlive(events)
		}
	})

	b.Run("StreamingApproach", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			file, err := os.CreateTemp("", "bench_streaming_mem_*.json")
			if err != nil {
				b.Fatal(err)
			}

			streamEventsToFile(file, 100)

			file.Close()
			os.Remove(file.Name())
		}
	})
}

// JSON marshaling benchmarks
func BenchmarkJSONMarshaling(b *testing.B) {
	event := createOptimizedEvent()

	b.Run("SingleEvent", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(event)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("EventSlice1000", func(b *testing.B) {
		events := make([]model.Event, 1000)
		for i := range events {
			events[i] = createOptimizedEvent()
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(events)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// CPU and memory profiling helper benchmark
func BenchmarkProfilingTarget(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Mix of different operations for profiling
		events := make([]model.Event, 500)

		// CPU-intensive string generation
		for j := 0; j < 500; j++ {
			if j%2 == 0 {
				events[j] = createTestEvent() // Original expensive string gen
			} else {
				events[j] = createOptimizedEvent() // Pool-based
			}
		}

		// JSON marshaling
		data, err := json.Marshal(events)
		if err != nil {
			b.Fatal(err)
		}

		// Simulated I/O
		file, err := os.CreateTemp("", "bench_profile_*.json")
		if err != nil {
			b.Fatal(err)
		}

		file.Write(data)
		file.Close()
		os.Remove(file.Name())

		runtime.KeepAlive(data)
	}
}
