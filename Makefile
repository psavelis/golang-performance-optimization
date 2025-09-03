
TARGET_DIR = $(CURDIR)/target
BUILD_DIR = $(TARGET_DIR)/build

$(TARGET_DIR):
	mkdir -p $(TARGET_DIR)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)


# Clean all the generated output
.PHONY: clean
clean:
	rm -fr $(TARGET_DIR)

# Builds generator and loader
.PHONY: build
build:
	go build -o $(BUILD_DIR)/bin/generator github.com/dmgo1014/interviewing-golang.git/cmd/generator
	go build -o $(BUILD_DIR)/bin/loader github.com/dmgo1014/interviewing-golang.git/cmd/loader

# stops pg instance
.PHONY: down_env
down_env:
	docker compose -f env/docker-compose.yaml down


# start pg instance
.PHONY: start_env
start_env:
	docker compose -f env/docker-compose.yaml up -d


# updates compose
.PHONY: update_env
update_env:
	docker compose -f env/docker-compose.yaml pull

# Test generate and load 10_000 of events
test_10k: clean build
	$(BUILD_DIR)/bin/generator 10000 test.json
	#$(BUILD_DIR)/bin/loader postgresql://test:test@localhost:5432/test?sslmode=disable test.json

# Test generate and load 100_000 of events
test_100k: clean build
	$(BUILD_DIR)/bin/generator 100000 test.json
	#$(BUILD_DIR)/bin/loader postgresql://test:test@localhost:5432/test?sslmode=disable test.json

# Test generate and load of 1_000_000 of events
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
	go build -o $(BUILD_DIR)/bin/generator-profiling github.com/dmgo1014/interviewing-golang.git/cmd/generator-profiling
	go build -o $(BUILD_DIR)/bin/loader-profiling github.com/dmgo1014/interviewing-golang.git/cmd/loader-profiling
	go build -o $(BUILD_DIR)/bin/generator-optimized-profiling github.com/dmgo1014/interviewing-golang.git/cmd/generator-optimized-profiling
	go build -o $(BUILD_DIR)/bin/loader-optimized-profiling github.com/dmgo1014/interviewing-golang.git/cmd/loader-optimized-profiling

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
	go build -o $(BUILD_DIR)/bin/generator-optimized github.com/dmgo1014/interviewing-golang.git/cmd/generator-optimized
	go build -o $(BUILD_DIR)/bin/loader-optimized github.com/dmgo1014/interviewing-golang.git/cmd/loader-optimized

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
