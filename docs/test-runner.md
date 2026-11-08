# Running standalone Fortran examples

[Documentation index](README.md) · [Command reference](commands.md) ·
[Configuration](configuration.md) · [AI-agent guide](ai-guide.md)

The `easifem test` command discovers standalone `*_test_*.F90` and
`*_test_*.f90` programs in a directory. It builds and executes them one at a
time and produces a separate log for each program.

```sh
easifem test /path/to/classes/src/modules/FEDOF/examples
```

This compiles test programs against already-installed libraries; it does
not build, install, or clean the EASIFEM libraries themselves.

## Selection

```sh
# Preview the selected filenames without building or executing examples.
easifem test ./examples --list

# Search below a modules directory, selecting either prefix.
easifem test ./src/modules --recursive --filter 'DG*' --filter 'H1*'
```

Discovery is sorted by relative filename. Without `--recursive`, only the
supplied directory is searched. Recursive discovery skips hidden directories,
`build`, and `CMakeFiles`. It does not follow symlinks below the root; an
explicitly supplied root symlink is resolved. Filters use Go `filepath.Match`
globs against a basename or relative path and are ORed together. They are not
regular expressions; `*` does not cross path separators.

An invalid directory, invalid filter, or empty selection is an error.
The CLI's existing EASIFEM configuration is required, including for `--list`.
`--list` cannot be combined with `--json`.

## Builds and fixture directories

Each example reads `runner.toml` from its own directory. Malformed TOML is a
build error. Without this file, the test runner defaults to Debug, gfortran,
Ninja, and the installed `easifemBase` target. Existing `runner.toml` settings
such as `TargetLibs = ["easifemClasses"]` are preserved.

Every program gets a fresh temporary workspace beneath the configured
`BuildDir`. Relative `BuildDir` values are resolved from the example directory;
without a setting, the parent is the system temporary directory under
`easifem-example-builds`. The workspace contains `source/CMakeLists.txt` and
`build/`. The original examples' `CMakeLists.txt` files are never overwritten.
Build workspaces are retained for inspection, including after failures.

`SourceDir`, extra source files, include paths, library directories, and extra
CMake prefix paths are honored. Relative source-related paths are resolved
from `SourceDir`, whose default is the example's directory. Include paths
always include the example directory and `SourceDir`. `Flags` supplies
individual compile options. `CompilerPath` may name a compiler executable or
its containing directory. `FileName` is always the discovered program;
`ExtraFileNames` supplies its additional sources.

The generated executable is run from the original example directory, so
relative mesh and fixture paths retain their existing meaning. Programs are
noninteractive: standard input is EOF. Existing programs can still modify
files in their example directory; isolation applies to the generated build,
not to the program's runtime filesystem.

`CacheClean` and `IsExecute` in `runner.toml` do not control this command:
workspaces are fresh and selected programs are always executed. `--rebuild`
or `ReBuild = true` requests `cmake --build --clean-first` only in that test's
isolated build directory. No prior test executable is reused after a failed
build. Separate invocations have separate workspaces, though programs writing
fixed runtime filenames must still not be run concurrently in one directory.

## Results and exit status

| Status | Meaning |
|---|---|
| `PASS` | Build and execution succeeded; at least one assertion passed and none failed. |
| `FAIL` | `not ok`, a TAP bailout, or an inconsistent TAP plan/check numbering. |
| `COMPILE` | Configuration, compilation, linking, or build setup failed. |
| `CRASH` | Execution failed or emitted a recognized runtime-error diagnostic. |
| `TIMEOUT` | The per-program deadline expired during build or execution. |
| `CANCELLED` | Execution was interrupted. |
| `NOCHECK` | Execution succeeded but reported no checks. |
| `SKIP` | Only explicit TAP skips/TODO checks, or a skipped zero-check plan. |

Checks use flat `Test_Method`/TAP-style `ok` and `not ok` records. Optional
`1..N` plans are checked when present. Explicit numbered checks must be
sequential. `# SKIP` and `# TODO` records are reported as skipped. Nested TAP
subtests are not supported. ANSI color escapes are ignored when recognizing
records. Build output is logged but never interpreted as execution checks.
A process failure overrides any earlier passing assertions.

The exit status is nonzero for any unsuccessful program, unrun selected
program, interruption, or infrastructure/reporting error. `NOCHECK` is
unsuccessful by default; `--allow-no-check` explicitly accepts it without
relabeling it as `PASS`. Explicit `SKIP` results are acceptable. `--fail-fast`
stops after the first unsuccessful result, respecting `--allow-no-check`.

## Logs, JSON, and interruption

```sh
easifem test ./examples --logdir ./test-logs --json ./results.json
easifem test ./examples --timeout 5m --fail-fast
```

Each invocation creates a unique log directory under `--logdir`, or the
system temporary directory. Log filenames include an index to distinguish
identically named tests in different directories. Logs contain build and
execution output, the commands, and their working directories.

The optional JSON report requires a `.json` filename. It contains format
version `1`, selected/completed/unrun counts, timestamps, result statuses,
assertion counts, error details, log paths, and executable paths. Durations
are integer nanoseconds. Reports are written through a temporary file and
retain partial results on interruption. A previous report at that path is
replaced only when the new report has been written successfully.

`--timeout` covers configuration, build, and execution for each test; zero
(the default) disables the deadline. Ctrl-C/SIGTERM stops the active test and
leaves remaining programs unrun. On Unix, cancellation terminates the child
process group, including compiler descendants. On other platforms, Go's
default cancellation terminates the immediate process only. Stream-drain
waiting is bounded to avoid hanging indefinitely on inherited output pipes.

## Implementation and validation

`internal/testrunner` uses only the Go standard library. Its executor interface
separates discovery/reporting from the CLI's CMake and `runner.toml` adapter.
Cobra flag handling remains in `internal/test.go`.

The existing `easifem run` command also now propagates subprocess failures:
it stops after failed configuration or compilation instead of continuing to
an earlier executable. Its process output is forwarded directly.

Suggested validation commands for the maintainer:

```sh
go test ./internal/testrunner
go test ./internal -run '^TestRunRunCmdPropagatesExitFailure$'
go build -o /tmp/easifem-review ./cmd/easifem
/tmp/easifem-review test /path/to/classes/src/modules/FEDOF/examples --list
/tmp/easifem-review test /path/to/classes/src/modules/FEDOF/examples \
  --filter 'DG*' --json /tmp/dg-results.json
```

Unit tests use fake executors; the `run` regression starts only a Go helper
process with a failing exit code. They do not build or run Fortran programs.
Validation completed on macOS on 2026-10-31:

- CLI and runner package unit tests passed.
- Discovery found all 67 FEDOF examples; the DG and full FEDOF runs passed.
- Deliberate compilation, assertion, runtime and timeout failures returned
  COMPILE, FAIL, CRASH and TIMEOUT respectively, each with exit status one.
- A passing control returned zero. NOCHECK returned one by default and zero
  with `--allow-no-check`. Per-program logs and JSON reports were verified.
