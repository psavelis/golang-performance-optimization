# Profiling Instrumentation

This directory contains profiling-enabled versions of the event generator and loader components. These files add CPU and memory profiling instrumentation to the standard implementations.

## Usage

Build profiling versions:
```bash
make build_profiling
```

Run profiling on the generator:
```bash
make profile_generator
```

Run profiling on the loader:
```bash
make profile_loader
```

## Generated Artifacts

The profiling runs produce `.prof` files in the current directory:
- `generator_cpu.prof`: CPU profile of generator execution
- `generator_mem.prof`: Memory profile of generator execution
- `loader_cpu.prof`: CPU profile of loader execution
- `loader_mem.prof`: Memory profile of loader execution

These files can be analyzed with Go's pprof tool:
```bash
go tool pprof -http=:8080 generator_cpu.prof
```

Or converted to SVG flamegraphs:
```bash
go tool pprof -svg generator_cpu.prof > generator_cpu.svg
```

## Implementation Details

The profiling code uses Go's runtime/pprof package to enable CPU and memory profiling:
- CPU profiling begins at application start and ends at termination
- Memory profiling captures a heap snapshot at application exit

These instrumented versions maintain identical business logic to their standard counterparts, but with added profiling hooks.
