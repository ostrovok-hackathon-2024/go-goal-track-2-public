#!/bin/bash

# Run pytest with coverage
# pytest --cov=src --cov-report=xml

# Run hyperfine
hyperfine --warmup 3 --runs 10 --export-json benchmark_dirty.json "poetry run cli -i "../inputs/rates_dirty.csv" -o out.csv -f csv" "poetry run cli -i "../inputs/rates_dirty.csv" -o out.json -f json" "poetry run cli -i "../inputs/rates_dirty.csv" -o out.parquet -f parquet"
