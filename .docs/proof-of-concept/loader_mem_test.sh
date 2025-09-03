#!/bin/bash

echo "Testing original loader..."
/usr/bin/time -l target/build/bin/loader 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_original.json 2>&1 | grep "maximum resident set size"

echo -e "\nTesting optimized loader..."
/usr/bin/time -l target/build/bin/loader-optimized 'postgresql://test:test@localhost:5432/test?sslmode=disable' test_optimized.json 2>&1 | grep "maximum resident set size"
