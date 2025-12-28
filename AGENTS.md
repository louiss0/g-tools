# Repository Guidelines

This repository is a collection of Go utility packages (enum, mode,
regex_extract). Changes should keep APIs small, idiomatic, and fully tested.

## Project Structure & Module Organization

- Root Go module in `go.mod`.
- Packages live in `enum/`, `mode/`, `regex_extract/`; each package keeps
  sources and tests together (e.g. `enum/pkg.go`, `enum/pkg_test.go`).
- Examples live in `*_test.go` via `Example...` functions.

## Build, Test, and Development Commands

- `ginkgo run ./...` runs all Ginkgo specs and examples across packages.
- `ginkgo watch ./...` reruns tests on file changes during development.
- `ginkgo build ./...` builds test binaries without running them.
- `ginkgo help` lists commands; `ginkgo help run` shows flags and pass-throughs.
- Consumers can set mode at build time with:
  `go build -ldflags "-X github.com/louiss0/g-tools/mode.buildMode=production"`.

## Coding Style & Naming Conventions

- Format Go code with `gofmt` (tabs, K&R braces).
- Use descriptive names; exported identifiers are `CamelCase`, unexported are
  `lowerCamel`.
- Prefer declarative APIs and small, cohesive types; keep state mutation close
  to where state is created.
- Do not mutate parameters; use comments only when intent is unclear,
  explaining "why".

## Testing Guidelines

- Tests must match the exact style used in `regex_extract/pkg_test.go`.
- Use same-package tests in `*_test.go`, with a `Test<Package>` function that
  initializes `tAssert = testifyassert.New(t)` and calls
  `RunSpecs(t, "<Package> Suite")`.
- Define specs as package-level vars like `var DescribeX = Describe("X", ...)`.
- Use the shared `tAssert` for all assertions (`NoError`, `ErrorIs`, `Equal`,
  `Len`, `Nil`, `Empty`) and do not reinitialize it in `BeforeEach`.
- Keep test data explicit and local (`pattern`, `input`, `expected`, `actual`,
  `err`), and define helper types inside the `Describe` where they are used.

## Commit & Pull Request Guidelines

- Commit messages follow `<type>(<scope>)<!>: <subject>` (imperative, <=64
  chars, no period). Example: `fix(regex-extract): handle empty capture sets`.
- Optional subtitle starts with `# `; body lines wrap ~72 chars; footers use
  `BREAKING CHANGE:` or `Fixes #123`.
- Each commit is a single logical change and must leave the repo buildable and
  tests passing.
- PRs should describe intent, list tests run, and link relevant issues; note
  any API changes.
