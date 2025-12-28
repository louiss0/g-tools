# Regex Extract Benchmark Report

Benchmarks were run with `go test -run ^$ -bench . -benchmem ./regex_extract`
using a local `GOCACHE`. Results capture compile-time regex usage and typed
conversion costs across the public API.

## Slowest Functions

- `ExtractTypedNamedGroups` is the slowest core API (around 16 us/op, 103
  allocs/op). It compiles the regex, parses the capture tree, builds nested
  maps, and performs reflective conversions per call, which adds both CPU time
  and allocations.
- `ExtractToStruct` and `ExtractTypedUnnamedGroups` follow (around 10-9 us/op,
  71-69 allocs/op). Both parse the capture tree and do reflection-heavy
  conversions into structs or slices, which is costly compared to simpler
  string extraction.
- `ExtractGroups` and `ExtractSlice` are faster but still pay for regex
  compilation on every call. The compile+extract path is significantly slower
  than a precompiled variant.

## Key Findings

- Regex compilation dominates the hot path when a pattern is reused. The
  precompiled `ExtractGroups` benchmark shows a large drop in time and
  allocations compared to compile+extract, indicating compilation and regex
  object setup are the primary overhead for simple extraction.
- Capture-tree parsing and reflective conversion are the next largest sources
  of time and allocations for typed APIs.
- `MapValues` is effectively free in comparison (sub-200 ns/op, zero allocs),
  highlighting that regex processing and reflection are the performance
  drivers in this package.

## Recommendations

- Cache or reuse compiled regexes in high-throughput callers.
- Keep patterns simple where possible; nested capture trees increase work.
- Prefer `ExtractGroups` or `ExtractSlice` when typed conversion is not needed.
