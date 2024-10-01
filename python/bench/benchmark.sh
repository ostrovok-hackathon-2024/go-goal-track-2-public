#!/bin/bash

# Run pytest with coverage
# pytest --cov=src --cov-report=xml

# Run hyperfine
hyperfine --warmup 3 --runs 10 --export-json benchmark_dirty.json "poetry run cli --verbose predict "./notebooks/rates_dirty.csv"  -o o.txt -f json" "poetry run cli --verbose predict "./notebooks/rates_dirty.csv"  -o o.txt -f csv" "poetry run cli --verbose predict "./notebooks/rates_dirty.csv"  -o o.txt -f parquet"
