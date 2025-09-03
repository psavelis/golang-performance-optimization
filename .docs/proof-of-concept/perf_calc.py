#!/usr/bin/env python3

# Original performance
original_generator = 846.867  # ms
original_loader = 40.219  # seconds
original_system = original_generator/1000 + original_loader  # seconds

# Optimized performance
optimized_generator = 149.784  # ms
optimized_loader = 7.548  # seconds
optimized_system = optimized_generator/1000 + optimized_loader  # seconds

# Calculate improvements
generator_improvement = original_generator / optimized_generator
loader_improvement = original_loader / optimized_loader
system_improvement = original_system / optimized_system

# Original throughput (events/second)
original_throughput = 100000 / original_system
optimized_throughput = 100000 / optimized_system
throughput_improvement = optimized_throughput / original_throughput

print(f"Performance Analysis Results:")
print(f"-----------------------------")
print(f"Generator: {generator_improvement:.2f}x improvement")
print(f"Loader: {loader_improvement:.2f}x improvement")
print(f"Total System: {system_improvement:.2f}x improvement")
print(f"Throughput: {throughput_improvement:.2f}x improvement")
print(f"Original throughput: {original_throughput:.0f} events/sec")
print(f"Optimized throughput: {optimized_throughput:.0f} events/sec")
print(f"\nREADME values to compare against:")
print(f"Generator: 6.01x")
print(f"Loader: 5.85x")
print(f"System Total: 5.86x")
print(f"Throughput: 6.6x (2,305 → 15,257 events/sec)")
