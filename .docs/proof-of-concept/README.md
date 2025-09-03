# Performance Verification Artifacts

This directory contains test scripts and results used to independently verify the performance claims in the documentation.

## Test Scripts

- `mem_test.sh`: Script to measure memory usage of the generator components
- `loader_mem_test.sh`: Script to measure memory usage of the loader components
- `perf_calc.py`: Python script to calculate and compare performance improvements

## Test Data

- `test_original.json`: Sample output from the original generator (100,000 events)
- `test_optimized.json`: Sample output from the optimized generator (100,000 events)

## Analysis Results

- [Performance Analysis Report](performance_analysis_report.md): Detailed comparison between claimed and measured performance metrics
