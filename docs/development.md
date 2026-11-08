# Development

This guide is for people changing the CLI itself. See
[architecture](architecture.md) for the repository layout and design rules.

## Requirements

- Go 1.21.6 or newer (the baseline in `go.mod`).
- For integration checks: CMake, Ninja, a Fortran compiler and installed
  EASIFEM libraries.

## Build and test

```sh
# Unit tests for both packages
go test ./internal/...

# Build a review executable without replacing the installed CLI
go build -o /tmp/easifem-review ./cmd/easifem

# Install the CLI to GOBIN / GOPATH/bin
go install ./cmd/easifem

# Check formatting without rewriting files
gofmt -l .
```

`go test ./internal/` covers only the `internal` package; `./internal/...` also
includes `internal/testrunner`.

The unit tests use fake executors and temporary directories. They do not build
or run Fortran. The `run` regression test (`TestRunRunCmdPropagatesExitFailure`)
starts only a Go helper process that exits with an error.

## Integration checks

These build and execute real Fortran examples, so they need a working EASIFEM
installation:

```sh
/tmp/easifem-review test /path/to/classes/src/modules/FEDOF/examples --list
/tmp/easifem-review test /path/to/classes/src/modules/FEDOF/examples \
  --filter 'DG*' --timeout 5m --json /tmp/dg-results.json
```

When changing result classification, also check deliberately broken programs:
a compile error should give `COMPILE`, a `not ok` check `FAIL`, a runtime
error `CRASH`, and a hang (with `--timeout`) `TIMEOUT`, each with exit status
1. A program without checks should give `NOCHECK` with exit status 1, or 0
with `--allow-no-check`.

The last full manual validation (macOS, 2026-10-31) covered the cases above,
plus all 67 FEDOF examples.

## Changing behavior

- Add focused tests for the changed behavior, including failure paths.
- Update the relevant page in `docs/` and the command's `--help` text in the
  same change. Legacy long-help strings in `internal/common.go` are not always
  accurate; fix them when you touch a command.
- Do not change the version-1 JSON report schema without documenting the
  compatibility impact.
- Bump `Version` in `internal/version.go` for a release.
- Do not raise the Go baseline or add dependencies without a concrete need.
