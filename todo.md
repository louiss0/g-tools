# Tests that are failing

None known. Latest run: `go test -run ^$ -bench . -benchmem ./regex_extract`.

# What bugs are present

None reported.

# What to do next

- Run the full test suite if needed: `ginkgo run ./...`.
- Consider adding precompiled-regex benchmarks for other extractors if
  performance work continues.
