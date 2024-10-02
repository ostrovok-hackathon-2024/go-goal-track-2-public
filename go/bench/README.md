# Benchmark

## Setup

### Install hyperfine

Or use [your](https://github.com/sharkdp/hyperfine) package manager.

## Run

```sh
./bench/benchmark.sh
```

## Results

| Benchmark | Command                                                                     | Time (mean ± σ)   | User     | System  | Range (min … max) | Runs |
| --------- | --------------------------------------------------------------------------- | ----------------- | -------- | ------- | ----------------- | ---- |
| 1         | `go run ./cmd/cli/main.go -i ../inputs/rates_dirty.csv -o ../outputs/o.csv` | 5.983 s ± 0.244 s | 38.477 s | 3.056 s | 5.610 s … 6.312 s | 10   |

## Notes

Benchmarks were run on a machine with the following configuration:

- CPU: MacBook Pro M3 (11 CPU, 18 GB RAM)
- Python: 3.12.6
- OS: macOS 15.0
