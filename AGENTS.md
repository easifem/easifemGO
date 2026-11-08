# AGENTS.md — easifemGO

## Purpose and scope

This repository contains the Go implementation of the `easifem` command-line
interface. Its module path is `github.com/easifem/easifemGO`. The baseline in
`go.mod` is Go 1.21.6. Commands use Cobra; existing configuration uses Viper
and TOML. Do not raise the Go baseline or add dependencies without a concrete
need and user agreement.

Read [the documentation index](docs/README.md) and the relevant command source
before changing behavior. Legacy command help in `internal/common.go` is not
always synchronized with the implementation.

## Authorization for actions

- Do not compile, install, or execute Go tests or Fortran examples unless the
  user's current instructions explicitly authorize that action. This includes
  `go run`, which compiles and executes code, and `easifem lint`, which compiles
  Fortran into object/module files.
- Do not treat a documentation example, existing tool approval, or an earlier
  successful test as authorization for a new task. Honor explicit permission
  within its stated scope; newer user restrictions take precedence.
- Cleaning or deleting build/install artifacts needs explicit approval for
  each invocation. `easifem clean` removes both build and install directories.
  Never clean a library to troubleshoot a test runner without permission.
- When validation is not authorized, finish the implementation and static
  review, then provide focused commands for the user to run. State clearly
  which checks were not run. Do not claim a suggested command succeeded.
- Read source, inspect Git diffs, and check formatting/whitespace as appropriate.
  Do not edit files during a review-only task. Formatting with `gofmt -w` edits
  files; `gofmt -l` and `gofmt -d` inspect them without rewriting.
- Do not commit, push, change dependencies, or install over the user's CLI
  unless requested. Keep unrelated user changes intact.

## Repository map

| Path | Responsibility |
|---|---|
| `cmd/easifem/main.go` | CLI entry point |
| `internal/root.go` | Root command, global flags, configuration initialization |
| `internal/common.go` | Configuration types, shared state, legacy help text |
| `internal/pkg.go` | Package metadata, builds, downloads, cache and shell-file generation |
| `internal/dev.go`, `install.go`, `clean.go` | Library/package lifecycle commands |
| `internal/run.go` | Single-program build/run command and CMake generation |
| `internal/test.go` | Test command, flags, reporting and exit policy |
| `internal/test_executor.go` | CMake/runner.toml adapter for standalone examples |
| `internal/testrunner/` | Standard-library-only discovery, execution policy, result parsing and reports |
| `docs/` | User and AI-agent documentation |

## Development conventions

- Keep reusable runner logic independent of Cobra, Viper, TOML, and the
  `internal` package. Inject execution through the `testrunner.Executor`
  interface. CLI configuration and CMake integration belong in the adapter.
- Keep new dependencies out of `internal/testrunner`; use the Go standard
  library. Preserve compatibility with the declared Go baseline.
- Use `gofmt`, explicit error handling, and wrapped errors with `%w` when the
  underlying cause is useful. Return errors from reusable functions rather
  than calling `log.Fatal` or `os.Exit` inside them.
- Launch commands using an executable plus separate arguments, not a shell
  command string. Set `exec.Cmd.Dir`; avoid process-wide `os.Chdir` in new code.
- Propagate real start/wait/exit errors. Never pass an `error` by value into a
  goroutine and expect it to update the caller. Do not run an older executable
  after a failed build.
- Keep stdout/stderr capture bounded in memory and retain full output in logs.
  Preserve cancellation and process cleanup; Unix cancellation uses process
  groups to include compiler descendants.
- Add focused tests for changed behavior. Use fake executors for runner tests
  and temporary directories for fixtures. Test failure paths as well as success.
  Writing tests does not authorize executing them.
- Update relevant CLI help and documentation in the same change. Keep flags,
  defaults, working directories, configuration precedence, side effects and
  exit behavior accurate. Do not document planned behavior as implemented.

## Test-runner invariants

- Discover sorted `*_test_*.F90`/`.f90` programs; an empty selection is an error.
- Run sequentially and use the example directory as the runtime working
  directory so mesh/fixture paths continue to work.
- Use isolated test build directories. Do not overwrite a source repository's
  `CMakeLists.txt` from `easifem test` or rebuild its installed libraries.
- Parse assertions only from execution output. A process failure overrides
  prior passing checks. Successful execution without checks is `NOCHECK`, not
  `PASS`; accepting it requires `--allow-no-check`.
- Keep build failures (`COMPILE`), assertion failures (`FAIL`), execution
  failures (`CRASH`), deadlines and cancellation distinct.
- Keep logs and partial JSON results on interruption. Do not silently change
  the version-1 report schema; document compatibility implications first.
- `--list` does not build or run Fortran, but normal root initialization can
  still update the user's cache and generated shell-variable file.

## Commands to offer the user

These are examples, not permission to execute them:

```sh
# Go unit tests (do not execute Fortran examples)
go test ./internal/...

# Build a review CLI without overwriting the installed executable
go build -o /tmp/easifem-review ./cmd/easifem

# Install the Go CLI, only when requested
go install ./cmd/easifem

# Discover examples without compiling or executing them
/tmp/easifem-review test /path/to/examples --list

# Focused integration check: builds and executes the selected examples
/tmp/easifem-review test /path/to/examples --filter 'DG*' \
  --timeout 5m --json /tmp/dg-results.json
```

`go test ./internal/` does not run tests in `internal/testrunner`; use
`go test ./internal/...` for both packages. `go install ./cmd/easifem`
installs the CLI; `easifem install classes` builds/installs a library package.
They are different operations and require their own authorization.

For using the CLI while working on other repositories, see
[the AI-agent operating guide](docs/ai-guide.md).
