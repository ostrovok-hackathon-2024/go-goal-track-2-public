# Benchmark

## Setup

### Install dependencies

```sh
poetry install
```

### Install hyperfine

Or use [your](https://github.com/sharkdp/hyperfine) package manager.

## Run

```sh
./bench/benchmark.sh
```

## Results

| Benchmark | Command                                                                            | Time (mean ± σ)                                     | Range (min … max) | Runs |
| --------- | ---------------------------------------------------------------------------------- | --------------------------------------------------- | ----------------- | ---- |
| 1         | `poetry run cli --verbose predict ./notebooks/rates_dirty.csv -o o.txt -f json`    | 7.381 s ± 0.181 s [User: 18.339 s, System: 2.872 s] | 7.193 s … 7.691 s | 10   |
| 2         | `poetry run cli --verbose predict ./notebooks/rates_dirty.csv -o o.txt -f csv`     | 6.890 s ± 0.112 s [User: 17.928 s, System: 2.769 s] | 6.795 s … 7.163 s | 10   |
| 3         | `poetry run cli --verbose predict ./notebooks/rates_dirty.csv -o o.txt -f parquet` | 6.672 s ± 0.196 s [User: 17.698 s, System: 2.729 s] | 6.482 s … 7.132 s | 10   |

## Notes

Benchmarks were run on a machine with the following configuration:

- CPU: MacBook Pro M3 (11 CPU, 18 GB RAM)
- Python: 3.12.6
- OS: macOS 15.0
