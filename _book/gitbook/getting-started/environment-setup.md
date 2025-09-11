# Environment Setup

Setting up your profiling environment correctly is crucial for effective performance analysis. This guide covers all necessary tools and configurations.

## Core Requirements

### Go Installation

Ensure you have Go 1.21+ installed (Go 1.24+ recommended for latest profiling features):

```bash
# Check your Go version
go version

# Should output: go version go1.24.x platform/arch
```

If you need to upgrade Go:

```bash
# Download from https://golang.org/dl/
# Or using Homebrew (macOS)
brew install go

# Or using package manager (Linux)
sudo apt-get update && sudo apt-get install golang-go
```

### Essential Tools Installation

Install the complete profiling toolkit:

```bash
# Install graphviz for flamegraph generation
# macOS
brew install graphviz

# Ubuntu/Debian
sudo apt-get install graphviz

# CentOS/RHEL
sudo yum install graphviz

# Install Go profiling tools
go install github.com/google/pprof@latest
go install github.com/pkg/profile@latest

# Install flamegraph tools
git clone https://github.com/brendangregg/FlameGraph.git
sudo cp FlameGraph/*.pl /usr/local/bin/
```

### Verification Script

Create a verification script to ensure everything is installed correctly:

```bash
#!/bin/bash
# save as verify-setup.sh

echo "🔍 Verifying Go Performance Tools Setup..."

# Check Go version
echo -n "Go version: "
go version || echo "❌ Go not installed"

# Check graphviz
echo -n "Graphviz: "
dot -V 2>&1 | head -1 || echo "❌ Graphviz not installed"

# Check pprof
echo -n "pprof: "
go tool pprof -help > /dev/null 2>&1 && echo "✅ Available" || echo "❌ Not available"

# Check flamegraph tools
echo -n "FlameGraph tools: "
which flamegraph.pl > /dev/null 2>&1 && echo "✅ Available" || echo "❌ Not available"

echo "✅ Setup verification complete!"
```

Run the verification:

```bash
chmod +x verify-setup.sh
./verify-setup.sh
```

## Development Environment Configuration

### VS Code Setup (Recommended)

Install essential extensions for Go profiling:

```bash
# Install VS Code extensions
code --install-extension golang.go
code --install-extension ms-vscode.go
code --install-extension alefragnani.project-manager
```

Configure VS Code settings for profiling:

```json
// .vscode/settings.json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.buildOnSave": "workspace",
    "go.lintOnSave": "workspace",
    "go.testFlags": ["-v", "-race"],
    "go.buildFlags": ["-race"],
    "files.associations": {
        "*.pprof": "plaintext",
        "*.svg": "xml"
    }
}
```

### Terminal Configuration

Add these aliases to your shell configuration (`~/.bashrc`, `~/.zshrc`):

```bash
# Go profiling aliases
alias pprof-cpu='go tool pprof -http=:8080'
alias pprof-mem='go tool pprof -http=:8080 -alloc_space'
alias pprof-web='go tool pprof -http=:8080'
alias bench-run='go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof'

# Flamegraph generation
alias flame-cpu='go tool pprof -raw -output=cpu.raw cpu.prof && stackcollapse-go.pl cpu.raw | flamegraph.pl > cpu-flame.svg'
alias flame-mem='go tool pprof -raw -output=mem.raw mem.prof && stackcollapse-go.pl mem.raw | flamegraph.pl > mem-flame.svg'
```

## Project Structure Setup

Create a standard project structure for profiling:

```bash
mkdir -p go-perf-project/{cmd,pkg,internal,profiles,benchmarks,docs}
cd go-perf-project

# Initialize Go module
go mod init go-perf-project

# Create essential directories
mkdir -p {profiles/{cpu,memory,goroutine,mutex,block},benchmarks,docs/flamegraphs}
```

### Makefile Template

Create a Makefile for common profiling tasks:

```makefile
# Makefile for Go profiling project

.PHONY: help setup bench profile clean

# Variables
BINARY_NAME=app
PROFILE_DIR=profiles
BENCHMARK_DIR=benchmarks

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Setup profiling environment
	@echo "Setting up profiling environment..."
	@mkdir -p $(PROFILE_DIR)/{cpu,memory,goroutine,mutex,block}
	@mkdir -p $(BENCHMARK_DIR)
	@mkdir -p docs/flamegraphs
	@go mod download

build: ## Build the application
	@go build -o bin/$(BINARY_NAME) ./cmd/...

bench: ## Run benchmarks with profiling
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem \
		-cpuprofile=$(PROFILE_DIR)/cpu/cpu.prof \
		-memprofile=$(PROFILE_DIR)/memory/mem.prof \
		-blockprofile=$(PROFILE_DIR)/block/block.prof \
		-mutexprofile=$(PROFILE_DIR)/mutex/mutex.prof \
		./...

profile-cpu: ## Analyze CPU profile
	@go tool pprof -http=:8080 $(PROFILE_DIR)/cpu/cpu.prof

profile-mem: ## Analyze memory profile  
	@go tool pprof -http=:8080 $(PROFILE_DIR)/memory/mem.prof

flamegraph-cpu: ## Generate CPU flamegraph
	@go tool pprof -raw -output=$(PROFILE_DIR)/cpu/cpu.raw $(PROFILE_DIR)/cpu/cpu.prof
	@stackcollapse-go.pl $(PROFILE_DIR)/cpu/cpu.raw | flamegraph.pl > docs/flamegraphs/cpu.svg
	@echo "CPU flamegraph: docs/flamegraphs/cpu.svg"

flamegraph-mem: ## Generate memory flamegraph
	@go tool pprof -raw -output=$(PROFILE_DIR)/memory/mem.raw $(PROFILE_DIR)/memory/mem.prof
	@stackcollapse-go.pl $(PROFILE_DIR)/memory/mem.raw | flamegraph.pl > docs/flamegraphs/memory.svg
	@echo "Memory flamegraph: docs/flamegraphs/memory.svg"

clean: ## Clean profiles and artifacts
	@rm -rf $(PROFILE_DIR)/*/*.prof $(PROFILE_DIR)/*/*.raw
	@rm -rf docs/flamegraphs/*.svg
	@echo "Cleaned profiles and artifacts"

server: ## Start pprof web server
	@echo "Starting pprof server at http://localhost:8080"
	@go tool pprof -http=:8080

.DEFAULT_GOAL := help
```

## Environment Validation

Create a test application to validate your setup:

```go
// main.go - Test application for environment validation
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func main() {
	// Enable profiling endpoint
	go func() {
		log.Println("Profiling server starting at :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Print environment info
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())

	// Simulate some work for profiling
	fmt.Println("Performing CPU-intensive work...")
	for i := 0; i < 5; i++ {
		cpuIntensiveWork()
		fmt.Printf("Iteration %d complete\n", i+1)
	}

	fmt.Println("✅ Environment validation complete!")
	fmt.Println("🌐 Profiling available at: http://localhost:6060/debug/pprof/")
	
	// Keep server running
	select {}
}

func cpuIntensiveWork() {
	start := time.Now()
	sum := 0
	for i := 0; i < 10000000; i++ {
		sum += i * i
	}
	duration := time.Since(start)
	fmt.Printf("Work completed in %v (sum: %d)\n", duration, sum)
}
```

Run the validation:

```bash
# Build and run
go build -o validation main.go
./validation

# In another terminal, test profiling
curl http://localhost:6060/debug/pprof/
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
```

## Next Steps

With your environment properly configured, you're ready to:

1. **[Generate Your First Profile](first-profile.md)** - Create and analyze profiles
2. **Explore profiling tools** - Learn pprof, flamegraphs, and analysis techniques
3. **Practice with real applications** - Apply profiling to actual Go projects

## Troubleshooting

### Common Issues

**"go tool pprof not found"**
```bash
# Reinstall Go tools
go install -a std
```

**"graphviz not working"**
```bash
# Verify installation
dot -V
# Reinstall if needed
```

**"flamegraph.pl not found"**
```bash
# Add to PATH or use full path
export PATH=$PATH:/usr/local/bin
```

**Permission issues**
```bash
# Fix permissions
chmod +x /usr/local/bin/*.pl
```

Your environment is now ready for advanced Go performance engineering! 🚀
