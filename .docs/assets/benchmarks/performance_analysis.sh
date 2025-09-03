#!/bin/bash

# Performance Analysis Script for Event Processing System
# This script generates comprehensive performance reports and flame graphs

set -e

echo "=== Event Processing System Performance Analysis ==="
echo "Date: $(date)"
echo

# Clean up previous runs
echo "Cleaning up previous results..."
make clean > /dev/null 2>&1

# Build all versions
echo "Building all versions..."
make build build_profiling build_optimized > /dev/null 2>&1

echo "=== Benchmark Results ==="
echo

# Test configurations
EVENTS_10K=10000
EVENTS_100K=100000
EVENTS_1M=1000000

echo "Testing with ${EVENTS_100K} events:"
echo

# Original performance
echo "Original Implementation:"
echo -n "  Generator: "
(time ./target/build/bin/generator ${EVENTS_100K} test_original.json) 2>&1 | grep real | awk '{print $2}'
echo -n "  Loader: "
(time ./target/build/bin/loader 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_original.json) 2>&1 | grep real | awk '{print $2}' | head -1

echo
echo "Optimized Implementation:"
echo -n "  Generator: "
(time ./target/build/bin/generator-optimized ${EVENTS_100K} test_optimized.json) 2>&1 | grep real | awk '{print $2}'
echo -n "  Loader: "
(time ./target/build/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_optimized.json) 2>&1 | grep real | awk '{print $2}' | head -1

echo
echo "=== File Size Comparison ==="
echo "Original JSON: $(ls -lah test_original.json | awk '{print $5}')"
echo "Optimized JSON: $(ls -lah test_optimized.json | awk '{print $5}')"

echo
echo "=== Memory Usage Analysis ==="

# Profile generator
echo "Profiling generator..."
./target/build/bin/generator-profiling ${EVENTS_100K} test_profiled.json > /dev/null

# Profile loader  
echo "Profiling loader..."
./target/build/bin/loader-profiling 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_profiled.json > /dev/null

echo "Generator Memory Profile:"
go tool pprof -top -cum -lines generator_mem.prof | head -10

echo
echo "Loader CPU Profile Top Functions:"
go tool pprof -top -cum -lines loader_cpu.prof | head -10

echo
echo "=== Flame Graph Servers Started ==="
echo "Access flame graphs at:"
echo "  Generator CPU: http://localhost:8080"
echo "  Generator Memory: http://localhost:8081" 
echo "  Loader CPU: http://localhost:8082"
echo "  Loader Memory: http://localhost:8083"

# Start flame graph servers in background
go tool pprof -http=:8080 generator_cpu.prof > /dev/null 2>&1 &
go tool pprof -http=:8081 generator_mem.prof > /dev/null 2>&1 &
go tool pprof -http=:8082 loader_cpu.prof > /dev/null 2>&1 &
go tool pprof -http=:8083 loader_mem.prof > /dev/null 2>&1 &

echo
echo "Flame graph servers started in background."
echo "Press any key to stop servers and exit..."
read -n 1 -s

# Kill background processes
pkill -f "go tool pprof -http"
echo "Servers stopped."
