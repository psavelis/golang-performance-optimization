package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dmgo1014/interviewing-golang/pkg/metrics"
	"os"
	"sync"
	"time"

	"github.com/dmgo1014/interviewing-golang/pkg/model"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

const (
	batchSize       = 1000             // Process events in batches
	workerCount     = 4                // Number of concurrent workers
	bufferSize      = 100 * 1024       // 100KB buffer for reading
	maxOpenConns    = 20               // Maximum open connections
	maxIdleConns    = 10               // Maximum idle connections
	connMaxLifetime = time.Hour        // Connection maximum lifetime
	connMaxIdleTime = time.Minute * 30 // Connection maximum idle time
	retryAttempts   = 3                // Number of retry attempts for failed operations
	retryDelay      = time.Second      // Delay between retry attempts
)

// Optimized loader with streaming JSON, batch processing, and concurrent workers
func main() {
	// start optional metrics server (env-driven)
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

	dbUrl := os.Args[1]
	url, err := dburl.Parse(dbUrl)
	if err != nil {
		panic(fmt.Errorf("unable to parse database URL '%s' : %+v", url, err))
	}

	inputFile := os.Args[2]
	fmt.Printf("input file: %s\n", inputFile)

	// Open database connection pool
	db, err := sql.Open("postgres", url.DSN)
	if err != nil {
		panic(fmt.Errorf("unable to connect to database : %+v", err))
	}
	defer db.Close()

	// Configure connection pool with optimized settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("unable to ping database : %+v", err))
	}

	// Stream and process events
	totalProcessed, err := processEventsStreaming(db, inputFile)
	if err != nil {
		panic(fmt.Errorf("unable to process events : %+v", err))
	}

	fmt.Printf("successfully loaded %d events\n", totalProcessed)
	// optionally hold for scraping in short-lived runs
	metrics.HoldFromEnv()
}

// Process events using streaming JSON and batch processing
func processEventsStreaming(db *sql.DB, inputFile string) (int, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Create channels for event batches
	batchChannel := make(chan []*model.Event, workerCount*2)
	errorChannel := make(chan error, 1) // Buffered to prevent blocking

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			processWorker(db, batchChannel, errorChannel, workerID)
		}(i)
	}

	// Stream and batch events
	totalProcessed := 0
	decoder := json.NewDecoder(file)

	// Read opening bracket
	token, err := decoder.Token()
	if err != nil {
		return 0, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return 0, fmt.Errorf("expected array start")
	}

	batch := make([]*model.Event, 0, batchSize)

	// Progress tracking
	startTime := time.Now()
	var processedCount int64

	// Process events in batches
	processingDone := make(chan bool, 1)
	go func() {
		defer close(batchChannel)
		defer func() { processingDone <- true }()

		for decoder.More() {
			var event model.Event
			if err := decoder.Decode(&event); err != nil {
				errorChannel <- err
				return
			}

			batch = append(batch, &event)

			// Send batch when full
			if len(batch) >= batchSize {
				select {
				case batchChannel <- batch:
					processedCount += int64(len(batch))
					totalProcessed += len(batch)

					// Progress reporting every 10,000 events
					if totalProcessed%10000 == 0 {
						elapsed := time.Since(startTime)
						rate := float64(totalProcessed) / elapsed.Seconds()
						fmt.Printf("Processed %d events (%.0f events/sec)\n", totalProcessed, rate)
					}

					batch = make([]*model.Event, 0, batchSize)
				case err := <-errorChannel:
					errorChannel <- err
					return
				}
			}
		}

		// Send remaining batch
		if len(batch) > 0 {
			select {
			case batchChannel <- batch:
				processedCount += int64(len(batch))
				totalProcessed += len(batch)
			case err := <-errorChannel:
				errorChannel <- err
				return
			}
		}
	}()

	// Wait for processing to complete
	<-processingDone

	// Wait for all workers to complete
	wg.Wait()

	// Check for errors
	select {
	case err := <-errorChannel:
		return 0, err
	default:
		// No error
	}

	return totalProcessed, nil
}

// Worker function to process batches of events with retry logic
func processWorker(db *sql.DB, batchChannel <-chan []*model.Event, errorChannel chan<- error, workerID int) {
	for batch := range batchChannel {
		var err error
		var success bool

		// Retry logic for failed batches
		for attempt := 0; attempt < retryAttempts && !success; attempt++ {
			if attempt > 0 {
				fmt.Printf("Worker %d: Retrying batch (attempt %d/%d)\n", workerID, attempt+1, retryAttempts)
				time.Sleep(retryDelay * time.Duration(attempt)) // Exponential backoff
			}

			err = loadBatch(db, batch)
			if err == nil {
				success = true
				break
			}

			// Log transient errors but continue retrying
			fmt.Printf("Worker %d: Batch failed (attempt %d): %v\n", workerID, attempt+1, err)
		}

		// If all retries failed, report error
		if !success {
			select {
			case errorChannel <- fmt.Errorf("worker %d failed after %d attempts: %v", workerID, retryAttempts, err):
			default:
				// Channel full, ignore error to prevent blocking
			}
			return
		}
	}
}

// Load a batch of events using prepared statement with optimized transaction handling
func loadBatch(db *sql.DB, events []*model.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Start transaction with retry on connection issues
	var tx *sql.Tx
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		tx, err = db.Begin()
		if err == nil {
			break
		}
		if attempt < 2 {
			time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1))
		}
	}
	if err != nil {
		return fmt.Errorf("failed to begin transaction after retries: %v", err)
	}

	// Ensure transaction is properly handled
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Prepare statement once for the batch - using direct time parameter (no string formatting)
	stmt, err := tx.Prepare(`
		INSERT INTO event(event_source, event_ref, event_type, event_date, calling_number, called_number, location,
		                  duration_seconds, attr_1, attr_2, attr_3, attr_4, attr_5, attr_6, attr_7, attr_8)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Execute prepared statement for each event in batch
	for i, event := range events {
		_, err = stmt.Exec(
			event.EventSource,
			event.EventRef,
			event.EventType,
			event.EventDate, // Direct time.Time parameter - no string conversion needed
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
		if err != nil {
			return fmt.Errorf("failed to execute statement for event %d in batch: %v", i, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	return nil
}
