# Benchmark

## Setup

### Install dependencies

```sh
go mod download
```

### Install hyperfine

Or use [your](https://github.com/sharkdp/hyperfine) package manager.

## Run

```sh
./bench/benchmark.sh
```

## Results

| Benchmark | Command                                                                             | Time (mean ± σ)                                     | Range (min … max) | Runs |
| --------- | ----------------------------------------------------------------------------------- | --------------------------------------------------- | ----------------- | ---- |
| 1         | `go run ./cmd/cli/main.go -i ../inputs/rates_dirty.csv -o ../outputs/o.csv -f csv`  | 6.177 s ± 0.528 s [User: 38.327 s, System: 3.099 s] | 5.474 s … 6.952 s | 10   |
| 2         | `go run ./cmd/cli/main.go -i ../inputs/rates_dirty.csv -o ../outputs/o.csv -f json` | 6.533 s ± 0.304 s [User: 38.965 s, System: 3.068 s] | 6.082 s … 7.186 s | 10   |

## Notes

Benchmarks were run on a machine with the following configuration:

- CPU: MacBook Pro M3 (11 CPU, 18 GB RAM)
- Python: 3.12.6
- OS: macOS 15.0
