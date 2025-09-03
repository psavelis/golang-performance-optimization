#!/bin/bash

echo "Testing original generator..."
rm -f test_original.json
/usr/bin/time -l target/build/bin/generator 100000 test_original.json 2>&1 | grep "maximum resident set size"

echo -e "\nTesting optimized generator..."
rm -f test_optimized.json
/usr/bin/time -l target/build/bin/generator-optimized 100000 test_optimized.json 2>&1 | grep "maximum resident set size"
