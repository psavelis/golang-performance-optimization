
TARGET_DIR = $(CURDIR)/target
BUILD_DIR = $(TARGET_DIR)/build
VERSION ?= 0.1.0
LDFLAGS = -X main.gitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown) \
		  -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
		  -X main.version=$(VERSION)

# Core Build Commands
.PHONY: clean build
clean:
	rm -fr $(TARGET_DIR)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator github.com/dmgo1014/interviewing-golang/cmd/generator
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader github.com/dmgo1014/interviewing-golang/cmd/loader
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-grafana-pyroscope github.com/dmgo1014/interviewing-golang/cmd/generator-grafana-pyroscope
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-grafana-pyroscope github.com/dmgo1014/interviewing-golang/cmd/loader-grafana-pyroscope
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-prometheus github.com/dmgo1014/interviewing-golang/cmd/generator-prometheus
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-prometheus github.com/dmgo1014/interviewing-golang/cmd/loader-prometheus
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-otel github.com/dmgo1014/interviewing-golang/cmd/generator-otel
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-otel github.com/dmgo1014/interviewing-golang/cmd/loader-otel
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/logger-fluentd github.com/dmgo1014/interviewing-golang/cmd/logger-fluentd
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/logger-elk github.com/dmgo1014/interviewing-golang/cmd/logger-elk
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/mcp-cli github.com/dmgo1014/interviewing-golang/cmd/mcp-cli

# Optimized Implementations
.PHONY: build_streaming
build_streaming:
	go build -o $(BUILD_DIR)/bin/generator-streaming github.com/dmgo1014/interviewing-golang/cmd/generator-streaming

# (Performance Analysis duplicates removed; see enhanced targets below)

# Environment Management
.PHONY: start_env down_env
start_env:
	docker compose -f env/docker-compose.yaml up -d

down_env:
	docker compose -f env/docker-compose.yaml down

# Start full profiling stack (Postgres + Pyroscope + Grafana)
.PHONY: start_profiling_stack stop_profiling_stack
start_profiling_stack:
	docker compose -f env/docker-compose.yaml up -d pyroscope grafana postgres
	@echo "Pyroscope UI:  http://localhost:4040"; \
	echo "Grafana UI:   http://localhost:3000 (admin/admin)";

stop_profiling_stack:
	docker compose -f env/docker-compose.yaml down

# Start observability stack including Prometheus
.PHONY: start_observability_stack
start_observability_stack:
	docker compose -f env/docker-compose.yaml up -d postgres pyroscope prometheus grafana otel-collector tempo loki fluentd elasticsearch kibana logstash
	@echo "Prometheus:   http://localhost:9090"; \
	echo "Pyroscope:    http://localhost:4040"; \
	echo "Grafana:      http://localhost:3000 (admin/admin)";
	echo "OTLP gRPC:    localhost:4317"; \
	echo "OTLP HTTP:    http://localhost:4318"; \
	echo "Tempo API:    http://localhost:3200";
	echo "Loki:         http://localhost:3100";
	echo "Fluentd:      http://localhost:9880";
	echo "Elasticsearch: http://localhost:9200";
	echo "Kibana:       http://localhost:5601";
	
.PHONY: run_logger_elk
run_logger_elk: build start_observability_stack
	LOGSTASH_HOST=127.0.0.1 LOGSTASH_PORT=8080 $(BUILD_DIR)/bin/logger-elk -n 200

.PHONY: run_logger_fluentd
run_logger_fluentd: build start_observability_stack
	FLUENTD_HOST=127.0.0.1 FLUENTD_PORT=24224 $(BUILD_DIR)/bin/logger-fluentd -n 200

# Start observability stack with Dynatrace exporter enabled (set DT_OTLP_ENDPOINT and DT_API_TOKEN)
.PHONY: start_observability_stack_dynatrace
start_observability_stack_dynatrace:
	@if [ -z "$$DT_OTLP_ENDPOINT" ] || [ -z "$$DT_API_TOKEN" ]; then \
		echo "DT_OTLP_ENDPOINT and DT_API_TOKEN must be set"; \
		exit 1; \
	fi
	docker compose -f env/docker-compose.yaml -f env/docker-compose.dynatrace.yaml up -d postgres pyroscope prometheus grafana otel-collector tempo
	@echo "Dynatrace OTLP: $$DT_OTLP_ENDPOINT";

# Run Pyroscope-enabled generator and loader locally (requires stack running)
.PHONY: run_pyroscope_generator run_pyroscope_loader
run_pyroscope_generator: build
	mkdir -p .docs/artifacts/test-data
	PYROSCOPE_SERVER_ADDRESS=http://localhost:4040 \
	PYROSCOPE_APPLICATION_NAME=interviewing-golang.generator \
	PYROSCOPE_TAGS=env=dev,service=generator \
	$(BUILD_DIR)/bin/generator-grafana-pyroscope 100000 .docs/artifacts/test-data/test_pyroscope.json

run_pyroscope_loader: build
	mkdir -p .docs/artifacts/test-data
	PYROSCOPE_SERVER_ADDRESS=http://localhost:4040 \
	PYROSCOPE_APPLICATION_NAME=interviewing-golang.loader \
	PYROSCOPE_TAGS=env=dev,service=loader \
	$(BUILD_DIR)/bin/loader-grafana-pyroscope 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_pyroscope.json

# Run Prometheus-instrumented generator and loader locally
.PHONY: run_prometheus_generator run_prometheus_loader
run_prometheus_generator: build
	mkdir -p .docs/artifacts/test-data
	METRICS_ADDR=:2112 METRICS_HOLD_FOR=60s \
	$(BUILD_DIR)/bin/generator-prometheus 100000 .docs/artifacts/test-data/test_prom.json

run_prometheus_loader: build start_env
	mkdir -p .docs/artifacts/test-data
	METRICS_ADDR=:2113 METRICS_HOLD_FOR=60s \
	$(BUILD_DIR)/bin/loader-prometheus 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_prom.json

# Run base apps with unobtrusive env-driven metrics
.PHONY: run_metrics_generator run_metrics_loader
run_metrics_generator: build
	mkdir -p .docs/artifacts/test-data
	METRICS_ENABLED=true METRICS_ADDR=:2112 METRICS_HOLD_FOR=60s \
	$(BUILD_DIR)/bin/generator 100000 .docs/artifacts/test-data/test_prom.json

run_metrics_loader: build start_env
	mkdir -p .docs/artifacts/test-data
	METRICS_ENABLED=true METRICS_ADDR=:2113 METRICS_HOLD_FOR=60s \
	$(BUILD_DIR)/bin/loader 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_prom.json

# Run OTel-instrumented generator and loader locally (export to local Collector)
.PHONY: run_otel_generator run_otel_loader

run_otel_generator: build start_observability_stack
	mkdir -p .docs/artifacts/test-data
	OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317 OTEL_EXPORTER_OTLP_INSECURE=true OTEL_SERVICE_NAME=interviewing-golang.generator-otel \
	$(BUILD_DIR)/bin/generator-otel 100000 .docs/artifacts/test-data/test_otel.json

run_otel_loader: build start_observability_stack
	mkdir -p .docs/artifacts/test-data
	OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317 OTEL_EXPORTER_OTLP_INSECURE=true OTEL_SERVICE_NAME=interviewing-golang.loader-otel \
	$(BUILD_DIR)/bin/loader-otel 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_otel.json

# Run OTel apps with Dynatrace exporter enabled (requires start_observability_stack_dynatrace)
.PHONY: run_otel_generator_dynatrace run_otel_loader_dynatrace
run_otel_generator_dynatrace: build start_observability_stack_dynatrace
	mkdir -p .docs/artifacts/test-data
	OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317 OTEL_EXPORTER_OTLP_INSECURE=true OTEL_SERVICE_NAME=interviewing-golang.generator-otel \
	$(BUILD_DIR)/bin/generator-otel 100000 .docs/artifacts/test-data/test_otel.json

run_otel_loader_dynatrace: build start_observability_stack_dynatrace
	mkdir -p .docs/artifacts/test-data
	OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317 OTEL_EXPORTER_OTLP_INSECURE=true OTEL_SERVICE_NAME=interviewing-golang.loader-otel \
	$(BUILD_DIR)/bin/loader-otel 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_otel.json

# Quick Tests
.PHONY: test_100k test_1M
test_100k: build_enhanced_profiling
	$(BUILD_DIR)/bin/generator-enhanced-profiling 100000 test_optimized.json

test_1M: clean build
	$(BUILD_DIR)/bin/generator 1000000 test.json
	#$(BUILD_DIR)/bin/loader postgresql://test:test@localhost:5432/test?sslmode=disable test.json
# Test generate and load of 1_000_000 of events
test_1B: clean build
	$(BUILD_DIR)/bin/generator 1000000000 test.json
	#$(BUILD_DIR)/bin/loader postgresql://test:test@localhost:5432/test?sslmode=disable test.json

# Build profiling versions
.PHONY: build_profiling
build_profiling:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-profiling
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-profiling github.com/dmgo1014/interviewing-golang/cmd/loader-profiling
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-optimized-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-optimized-profiling
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-optimized-profiling github.com/dmgo1014/interviewing-golang/cmd/loader-optimized-profiling

# Profile generator with 100k events
.PHONY: profile_generator
profile_generator: clean build_profiling
	$(BUILD_DIR)/bin/generator-profiling 100000 test.json

# Profile loader with existing test.json
.PHONY: profile_loader
profile_loader: build_profiling
	$(BUILD_DIR)/bin/loader-profiling postgresql://test:test@localhost:5432/test?sslmode=disable test.json

# Profile optimized generator with 100k events
.PHONY: profile_generator_optimized
profile_generator_optimized: clean build_profiling
	$(BUILD_DIR)/bin/generator-optimized-profiling 100000 test_optimized.json

# Profile optimized loader with existing test.json
.PHONY: profile_loader_optimized
profile_loader_optimized: build_profiling
	$(BUILD_DIR)/bin/loader-optimized-profiling postgresql://test:test@localhost:5432/test?sslmode=disable test_optimized.json

# Generate flame graphs from profiles
.PHONY: flamegraph
flamegraph:
	go tool pprof -http=:8080 generator_cpu.prof &
	go tool pprof -http=:8081 generator_mem.prof &
	go tool pprof -http=:8082 loader_cpu.prof &
	go tool pprof -http=:8083 loader_mem.prof &

# Generate flame graphs from optimized profiles
.PHONY: flamegraph_optimized
flamegraph_optimized:
	go tool pprof -http=:8084 generator_optimized_cpu.prof &
	go tool pprof -http=:8085 generator_optimized_mem.prof &
	go tool pprof -http=:8086 loader_optimized_cpu.prof &
	go tool pprof -http=:8087 loader_optimized_mem.prof &

# Generate all SVG flamegraphs and move them to artifacts
.PHONY: generate_all_flamegraphs
generate_all_flamegraphs:
	mkdir -p .docs/artifacts/flamegraphs
	# Generate original SVGs
	go tool pprof -svg generator_cpu.prof > .docs/artifacts/flamegraphs/generator_original_cpu.svg
	go tool pprof -svg generator_mem.prof > .docs/artifacts/flamegraphs/generator_original_mem.svg
	go tool pprof -svg loader_cpu.prof > .docs/artifacts/flamegraphs/loader_original_cpu.svg
	go tool pprof -svg loader_mem.prof > .docs/artifacts/flamegraphs/loader_original_mem.svg
	# Generate optimized SVGs
	go tool pprof -svg generator_optimized_cpu.prof > .docs/artifacts/flamegraphs/generator_optimized_cpu.svg
	go tool pprof -svg generator_optimized_mem.prof > .docs/artifacts/flamegraphs/generator_optimized_mem.svg
	go tool pprof -svg loader_optimized_cpu.prof > .docs/artifacts/flamegraphs/loader_optimized_cpu.svg
	go tool pprof -svg loader_optimized_mem.prof > .docs/artifacts/flamegraphs/loader_optimized_mem.svg
	# Copy profile files
	mkdir -p .docs/artifacts/profiles
	cp *.prof .docs/artifacts/profiles/
	
# Run all profiling and generate all artifacts
.PHONY: profile_all
profile_all: clean build_profiling
	# Run original profiling
	$(BUILD_DIR)/bin/generator-profiling 100000 test.json
	$(BUILD_DIR)/bin/loader-profiling postgresql://test:test@localhost:5432/test?sslmode=disable test.json
	# Run optimized profiling
	$(BUILD_DIR)/bin/generator-optimized-profiling 100000 test_optimized.json
	$(BUILD_DIR)/bin/loader-optimized-profiling postgresql://test:test@localhost:5432/test?sslmode=disable test_optimized.json
	# Generate all flamegraphs
	$(MAKE) generate_all_flamegraphs

# Build optimized versions
.PHONY: build_optimized
build_optimized:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-optimized github.com/dmgo1014/interviewing-golang/cmd/generator-optimized
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-optimized github.com/dmgo1014/interviewing-golang/cmd/loader-optimized

# Test optimized generator with 100k events
.PHONY: test_optimized_100k
test_optimized_100k: clean build_optimized
	$(BUILD_DIR)/bin/generator-optimized 100000 test_optimized.json
	$(BUILD_DIR)/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_optimized.json

# Test optimized generator with 1M events
.PHONY: test_optimized_1M
test_optimized_1M: clean build_optimized
	$(BUILD_DIR)/bin/generator-optimized 1000000 test_optimized.json
	$(BUILD_DIR)/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_optimized.json

# Performance comparison
.PHONY: benchmark_comparison
benchmark_comparison: clean build build_optimized
	@echo "=== Original Performance ==="
	time $(BUILD_DIR)/bin/generator 100000 test_original.json
	time $(BUILD_DIR)/bin/loader 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_original.json
	@echo "=== Optimized Performance ==="
	time $(BUILD_DIR)/bin/generator-optimized 100000 test_optimized.json
	time $(BUILD_DIR)/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_optimized.json

# Enhanced profiling targets
.PHONY: build_enhanced_profiling
build_enhanced_profiling:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-enhanced-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-enhanced-profiling
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/generator-streaming github.com/dmgo1014/interviewing-golang/cmd/generator-streaming
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bin/loader-pgx github.com/dmgo1014/interviewing-golang/cmd/loader-pgx

# Enhanced profiling with block/mutex/goroutine tracking
.PHONY: profile_enhanced
profile_enhanced: clean build_enhanced_profiling
	mkdir -p .docs/artifacts/profiles
	$(BUILD_DIR)/bin/generator-enhanced-profiling 100000 test_enhanced.json
	mv *.prof .docs/artifacts/profiles/

# Streaming JSON profiling
.PHONY: profile_streaming
profile_streaming: clean build_enhanced_profiling
	mkdir -p .docs/artifacts/profiles
	$(BUILD_DIR)/bin/generator-streaming 100000 test_streaming.json
	
# pgx COPY profiling  
.PHONY: profile_pgx
profile_pgx: build_enhanced_profiling start_env
	sleep 3  # Wait for DB to be ready
	mkdir -p .docs/artifacts/profiles
	$(BUILD_DIR)/bin/loader-pgx postgresql://test:test@localhost:5432/test?sslmode=disable test_enhanced.json

# Comprehensive benchmarking
.PHONY: benchmark
benchmark:
	@echo "=== Running Comprehensive Benchmarks ==="
	go test -v ./pkg/benchmarks/... -run=^$$ -bench=. -benchtime=2s -benchmem | tee .docs/artifacts/benchmark_results.txt

# Benchmark comparison with benchstat
.PHONY: benchmark_compare
benchmark_compare:
	@echo "=== String Generation Comparison ==="
	go test -v ./pkg/benchmarks/... -run=^$$ -bench=BenchmarkStringGeneration -count=5 -benchmem > .docs/artifacts/string_bench.txt
	@echo "=== Memory Allocation Comparison ==="
	go test -v ./pkg/benchmarks/... -run=^$$ -bench=BenchmarkMemoryAllocations -count=5 -benchmem > .docs/artifacts/memory_bench.txt
	@echo "=== Generator Performance Comparison ==="
	go test -v ./pkg/benchmarks/... -run=^$$ -bench=BenchmarkGenerator -count=5 -benchmem > .docs/artifacts/generator_bench.txt

# Trace collection for detailed analysis
.PHONY: trace_enhanced
trace_enhanced: build_enhanced_profiling
	mkdir -p .docs/artifacts/traces
	$(BUILD_DIR)/bin/generator-enhanced-profiling -trace 50000 test_trace.json

# Complete profiling study
.PHONY: profiling_study
profiling_study: clean build_enhanced_profiling start_env
	@echo "=== Running Complete Profiling Study ==="
	mkdir -p .docs/artifacts/{profiles,traces,flamegraphs,benchmarks}
	# Enhanced profiling with block/mutex/goroutine
	$(BUILD_DIR)/bin/generator-enhanced-profiling 100000 test_enhanced.json
	$(BUILD_DIR)/bin/generator-streaming 100000 test_streaming.json
	sleep 3  # Wait for DB
	$(BUILD_DIR)/bin/loader-pgx postgresql://test:test@localhost:5432/test?sslmode=disable test_enhanced.json
	# Move profiles
	mv *.prof .docs/artifacts/profiles/
	# Generate enhanced flamegraphs with all profile types
	$(MAKE) generate_enhanced_flamegraphs
	# Run comprehensive benchmarks
	$(MAKE) benchmark_compare
	@echo "=== Profiling Study Complete ==="
	@echo "Results available in .docs/artifacts/"

# Generate enhanced flamegraphs including block/mutex/goroutine profiles
.PHONY: generate_enhanced_flamegraphs
generate_enhanced_flamegraphs:
	mkdir -p .docs/artifacts/flamegraphs
	# CPU profiles
	go tool pprof -svg .docs/artifacts/profiles/generator_enhanced_cpu.prof > .docs/artifacts/flamegraphs/generator_enhanced_cpu.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/generator_streaming_cpu.prof > .docs/artifacts/flamegraphs/generator_streaming_cpu.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/loader_pgx_cpu.prof > .docs/artifacts/flamegraphs/loader_pgx_cpu.svg 2>/dev/null || true
	# Memory profiles
	go tool pprof -svg .docs/artifacts/profiles/generator_enhanced_mem.prof > .docs/artifacts/flamegraphs/generator_enhanced_mem.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/generator_streaming_mem.prof > .docs/artifacts/flamegraphs/generator_streaming_mem.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/loader_pgx_mem.prof > .docs/artifacts/flamegraphs/loader_pgx_mem.svg 2>/dev/null || true
	# Block profiles
	go tool pprof -svg .docs/artifacts/profiles/generator_enhanced_block.prof > .docs/artifacts/flamegraphs/generator_enhanced_block.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/generator_streaming_block.prof > .docs/artifacts/flamegraphs/generator_streaming_block.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/loader_pgx_block.prof > .docs/artifacts/flamegraphs/loader_pgx_block.svg 2>/dev/null || true
	# Mutex profiles
	go tool pprof -svg .docs/artifacts/profiles/generator_enhanced_mutex.prof > .docs/artifacts/flamegraphs/generator_enhanced_mutex.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/generator_streaming_mutex.prof > .docs/artifacts/flamegraphs/generator_streaming_mutex.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/loader_pgx_mutex.prof > .docs/artifacts/flamegraphs/loader_pgx_mutex.svg 2>/dev/null || true
	# Goroutine profiles
	go tool pprof -svg .docs/artifacts/profiles/generator_enhanced_goroutine.prof > .docs/artifacts/flamegraphs/generator_enhanced_goroutine.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/generator_streaming_goroutine.prof > .docs/artifacts/flamegraphs/generator_streaming_goroutine.svg 2>/dev/null || true
	go tool pprof -svg .docs/artifacts/profiles/loader_pgx_goroutine.prof > .docs/artifacts/flamegraphs/loader_pgx_goroutine.svg 2>/dev/null || true

# Interactive profiling servers
.PHONY: pprof_server
pprof_server:
	@echo "Starting pprof servers for enhanced profiles..."
	@echo "Enhanced CPU: http://localhost:8080"
	@echo "Enhanced Memory: http://localhost:8081" 
	@echo "Enhanced Block: http://localhost:8082"
	@echo "Enhanced Mutex: http://localhost:8083"
	@echo "Enhanced Goroutine: http://localhost:8084"
	go tool pprof -http=:8080 .docs/artifacts/profiles/generator_enhanced_cpu.prof &
	go tool pprof -http=:8081 .docs/artifacts/profiles/generator_enhanced_mem.prof &
	go tool pprof -http=:8082 .docs/artifacts/profiles/generator_enhanced_block.prof &
	go tool pprof -http=:8083 .docs/artifacts/profiles/generator_enhanced_mutex.prof &
	go tool pprof -http=:8084 .docs/artifacts/profiles/generator_enhanced_goroutine.prof &
	@echo "Press Ctrl+C to stop all servers"
	wait

# Quick benchmark run for development
.PHONY: bench_quick
bench_quick:
	go test -v ./pkg/benchmarks/... -run=^$$ -bench=BenchmarkStringGeneration -benchtime=500ms

# Note: scripts/parse_bench_regressions.sh can be used in CI to fail on >5% slowdown (configurable via THRESHOLD_PERCENT)

# CI-focused lightweight profiling (generator only)
.PHONY: ci_profiles
ci_profiles: clean
	mkdir -p .docs/artifacts/ci/flamegraphs
	mkdir -p .docs/artifacts/ci/raw
	go build -o $(BUILD_DIR)/bin/generator-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-profiling
	go build -o $(BUILD_DIR)/bin/generator-optimized-profiling github.com/dmgo1014/interviewing-golang/cmd/generator-optimized-profiling
	$(BUILD_DIR)/bin/generator-profiling 50000 /dev/null || true
	$(BUILD_DIR)/bin/generator-optimized-profiling 50000 /dev/null || true
	# Flamegraphs (CPU)
	go tool pprof -svg generator_cpu.prof > .docs/artifacts/ci/flamegraphs/generator_cpu.svg 2>/dev/null || true
	go tool pprof -svg generator_optimized_cpu.prof > .docs/artifacts/ci/flamegraphs/generator_optimized_cpu.svg 2>/dev/null || true
	# Hotspots (CPU)
	go tool pprof -top -nodecount=15 generator_cpu.prof > .docs/artifacts/ci/generator_cpu_top.txt 2>/dev/null || true
	go tool pprof -top -nodecount=15 generator_optimized_cpu.prof > .docs/artifacts/ci/generator_optimized_cpu_top.txt 2>/dev/null || true
	# Memory allocations (if heap profiles present)
	go tool pprof -top -alloc_space -nodecount=15 generator_mem.prof > .docs/artifacts/ci/generator_mem_top.txt 2>/dev/null || true
	go tool pprof -top -alloc_space -nodecount=15 generator_optimized_mem.prof > .docs/artifacts/ci/generator_optimized_mem_top.txt 2>/dev/null || true
	# Diff (CPU time) optimized vs original
	go tool pprof -top -nodecount=15 -diff_base=generator_cpu.prof generator_optimized_cpu.prof > .docs/artifacts/ci/generator_cpu_diff_top.txt 2>/dev/null || true
	# Diff (alloc_space)
	go tool pprof -top -alloc_space -nodecount=15 -diff_base=generator_mem.prof generator_optimized_mem.prof > .docs/artifacts/ci/generator_mem_diff_top.txt 2>/dev/null || true
	# Raw & proto export for external tooling (e.g., speedscope conversion offline)
	cp generator_cpu.prof .docs/artifacts/ci/raw/ || true
	cp generator_optimized_cpu.prof .docs/artifacts/ci/raw/ || true
	cp generator_mem.prof .docs/artifacts/ci/raw/ || true
	cp generator_optimized_mem.prof .docs/artifacts/ci/raw/ || true
	go tool pprof -proto generator_cpu.prof > .docs/artifacts/ci/raw/generator_cpu.pb 2>/dev/null || true
	go tool pprof -proto generator_optimized_cpu.prof > .docs/artifacts/ci/raw/generator_optimized_cpu.pb 2>/dev/null || true
	@echo "CI profiling artifacts in .docs/artifacts/ci"

# Convert CPU profiles to Speedscope JSON (requires speedscope CLI or npx)
.PHONY: speedscope_convert
speedscope_convert:
	mkdir -p .docs/artifacts/ci/speedscope
	@if ! command -v npx >/dev/null 2>&1; then echo "npx not found (install Node.js)"; exit 1; fi
	@if [ -f generator_cpu.prof ]; then npx --yes speedscope generator_cpu.prof -o .docs/artifacts/ci/speedscope/generator_cpu.speedscope.json || true; fi
	@if [ -f generator_optimized_cpu.prof ]; then npx --yes speedscope generator_optimized_cpu.prof -o .docs/artifacts/ci/speedscope/generator_optimized_cpu.speedscope.json || true; fi
	@echo "Speedscope artifacts in .docs/artifacts/ci/speedscope"

# Run AI heuristic analysis locally (after ci_profiles and optionally benchdiff artifacts)
.PHONY: ai_analysis_local
ai_analysis_local:
	go build -o $(BUILD_DIR)/bin/ai-profiler-analyzer github.com/dmgo1014/interviewing-golang/cmd/ai-profiler-analyzer
	PROFILES_DIR=.docs/artifacts/ci BENCH_DIR=.docs/artifacts/benchdiff $(BUILD_DIR)/bin/ai-profiler-analyzer || true
	@echo "AI analysis written to .docs/artifacts/ci/ai_analysis"

# --- GitBook Documentation ---
.PHONY: docs_build
docs_build:
	cd gitbook && npx --yes honkit build . _book

.PHONY: docs_serve
docs_serve:
	@if lsof -i :4000 >/dev/null 2>&1; then echo "Port 4000 busy"; exit 1; fi
	cd gitbook && npx --yes honkit serve --port 4000

.PHONY: docs_static_serve
docs_static_serve:
	python3 -m http.server --directory gitbook/_book 4001

# List MCP capabilities
.PHONY: mcp_list
mcp_list: build
	$(BUILD_DIR)/bin/mcp-cli -list

# Show MCP schemas
.PHONY: mcp_schemas
mcp_schemas: build
	$(BUILD_DIR)/bin/mcp-cli -schemas

# Validate an example JSON file against a capability schema: make mcp_validate CAP=prometheus.query_assistant FILE=payload.json
.PHONY: mcp_validate
mcp_validate: build
	@if [ -z "$(CAP)" ] || [ -z "$(FILE)" ]; then echo "Usage: make mcp_validate CAP=capability.id FILE=path.json"; exit 1; fi
	$(BUILD_DIR)/bin/mcp-cli -cap $(CAP) -in $(FILE) -validate

# Contract test: iterate all capabilities with a minimal empty {} (expect failures allowed=0) to ensure schemas reachable
.PHONY: mcp_contract_test
mcp_contract_test: build
	@echo "== MCP Contract Test =="; set -e; FAIL=0; while IFS= read -r line; do \
	  cap=$$(echo $$line | awk '{print $$1}'); \
	  printf "Testing %s ... " "$$cap"; \
	  echo '{}' | $(BUILD_DIR)/bin/mcp-cli -cap $$cap -validate >/dev/null 2>&1 && { echo "unexpected PASS (empty object)"; FAIL=1; } || echo "schema enforced"; \
	done < <($(BUILD_DIR)/bin/mcp-cli -list); \
	if [ $$FAIL -eq 1 ]; then echo "Contract test failed"; exit 1; fi; echo "All schemas reachable & enforcing";

.PHONY: mcp_examples_validate
mcp_examples_validate: build
	@echo "== Validating Example Payloads =="; set -e; for f in mcp/examples/*.json; do \
	  cap=$$(basename $$f .json); \
	  printf "Example %s ... " "$$cap"; \
	  $(BUILD_DIR)/bin/mcp-cli -cap $$cap -in $$f -validate >/dev/null && echo OK || { echo FAIL; exit 1; }; \
	done
