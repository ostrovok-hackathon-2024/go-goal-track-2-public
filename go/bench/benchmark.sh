#!/bin/bash

# Run hyperfine
hyperfine --warmup 3 --runs 10 --export-json benchmark_dirty.json "go run ./cmd/cli/main.go -i ../inputs/rates_dirty.csv -o ../outputs/o.csv -f csv" "go run ./cmd/cli/main.go -i ../inputs/rates_dirty.csv -o ../outputs/o.csv -f json" 
