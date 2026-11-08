# Command: test

```sh
easifem test [directory] [flags]
```

Discover the standalone `*_test_*.F90`/`*_test_*.f90` programs in a directory
(the current directory by default), then build and run each one, one at a
time, with a separate log for each.

```sh
easifem test ./examples --list
easifem test ./examples --filter 'Get*' --filter 'DG*'
easifem test ./src/modules --recursive --timeout 5m --json results.json
```

Test programs are compiled against the **installed** libraries. `easifem test`
does not build, install or clean the EASIFEM libraries themselves; after
changing library source, run [`easifem install`](install.md) first.

## Flags

| Flag                | Default                    | Meaning                                                           |
| ------------------- | -------------------------- | ----------------------------------------------------------------- |
| `--recursive`, `-r` | false                      | Also search subdirectories (see [Selection](#selection)).         |
| `--filter`          | all programs               | Glob on the filename or relative path. Repeat to OR several.      |
| `--list`            | false                      | Print the selection without building or running anything.         |
| `--logdir`          | system temporary directory | Parent of a unique per-invocation log directory.                  |
| `--json`            | disabled                   | Write a version-1 JSON report to this `.json` file.               |
| `--timeout`         | `0` (no limit)             | Deadline per program, covering configure, build and execution.    |
| `--fail-fast`       | false                      | Stop after the first unsuccessful program.                        |
| `--rebuild`         | false                      | `--clean-first` in each program's isolated build directory.       |
| `--allow-no-check`  | false                      | Treat `NOCHECK` as acceptable (it is still reported as `NOCHECK`).|

`--list` cannot be combined with `--json`. A negative `--timeout`, a `--json`
path without a `.json` extension, an invalid filter, an invalid directory and
an empty selection are all errors.

## Selection

```sh
# Preview the selection without building or executing examples
easifem test ./examples --list

# Search below a modules directory, selecting either prefix
easifem test ./src/modules --recursive --filter 'DG*' --filter 'H1*'
```

- Programs are sorted by relative path.
- Without `--recursive`, only the given directory is searched. With it,
  hidden directories, `build` and `CMakeFiles` are skipped, and symlinks below
  the root are not followed (a symlinked root directory itself is resolved).
- Filters are Go [`filepath.Match`](https://pkg.go.dev/path/filepath#Match)
  globs, matched against the basename or the relative path. They are not
  regular expressions, and `*` does not match `/`.
- The EASIFEM [configuration](configuration.md) is required even for `--list`,
  and start-up still rewrites `cache.toml` and `easifemvar.fish`.

## Builds and working directory

Each program reads `runner.toml` from its own directory (see
[configuration](configuration.md#runnertoml)). Without one, the runner uses
`Debug`, `gfortran`, Ninja and the installed `easifemBase` target. Malformed
`runner.toml` gives a `COMPILE` result.

- **Isolated builds.** Every program gets a fresh workspace,
  `<BuildDir>/easifem-test-XXXX/`, containing a generated
  `source/CMakeLists.txt` and `build/`. `BuildDir` defaults to
  `easifem-example-builds` in the system temporary directory; a relative value
  is resolved from the example directory. Workspaces are kept afterwards, also
  after failures, so you can inspect them. The examples' own `CMakeLists.txt`
  files are never touched.
- **Source settings.** The main source is always the discovered program
  (`FileName` is ignored). `ExtraFileNames`, `IncludePath`,
  `ExtraIncludePath`, `LibraryPath`, `ExtraCMakePrefixPath`, `ExtraLibs` and
  `Flags` are honored. Relative paths are resolved from `SourceDir`, which
  defaults to the example directory. The example directory and `SourceDir` are
  always on the include path. `CompilerPath` may be a compiler executable or
  the directory that contains `Compiler`.
- **Rebuilds.** `--rebuild` or `ReBuild = true` adds `--clean-first` to that
  program's isolated build. `CacheClean` and `IsExecute` have no effect here:
  workspaces are always fresh, and every selected program is executed. A
  failed build is never followed by running an older executable.
- **Runtime directory.** Each executable runs from its example directory, so
  relative mesh and fixture paths keep working. Standard input is empty
  (EOF). Isolation covers the build only: programs can still write files next
  to the example, so do not run two invocations that write the same output
  files in the same directory at the same time.

## Results

| Status      | Meaning                                                                        |
| ----------- | ------------------------------------------------------------------------------ |
| `PASS`      | Build and execution succeeded; at least one check passed and none failed.     |
| `FAIL`      | A `not ok` check, a TAP bail-out, or an inconsistent TAP plan or numbering.   |
| `COMPILE`   | Build setup, `runner.toml`, CMake configuration, compilation or linking failed.|
| `CRASH`     | Execution failed, or printed a recognized runtime-error diagnostic.           |
| `TIMEOUT`   | The `--timeout` deadline expired during build or execution.                   |
| `CANCELLED` | The run was interrupted (Ctrl-C or SIGTERM).                                   |
| `NOCHECK`   | Execution succeeded but printed no checks.                                     |
| `SKIP`      | Only explicit TAP skips/TODOs, or a zero-check plan marked as skipped.         |

Checks are read only from the program's output, never from build output:

- Flat `Test_Method`/TAP-style `ok` and `not ok` lines are recognized; ANSI
  color codes are ignored.
- An optional `1..N` plan is verified when present, and numbered checks must
  be sequential.
- `# SKIP` and `# TODO` checks are reported as skipped.
- Nested TAP subtests are not supported.
- A process failure overrides earlier passing checks (the result is `CRASH`).

## Exit status

The command exits with status 1 when any selected program is unsuccessful or
was not run, when the run is interrupted, or on an infrastructure or reporting
error. `PASS` and `SKIP` are successful. `NOCHECK` is unsuccessful unless
`--allow-no-check` is given; use that flag only when you deliberately accept
programs without checks. `--fail-fast` stops after the first unsuccessful
result, taking `--allow-no-check` into account.

## Logs, JSON and interruption

```sh
easifem test ./examples --logdir ./test-logs --json ./results.json
easifem test ./examples --timeout 5m --fail-fast
```

**Logs.** Each invocation creates a unique log directory under `--logdir` (or
the system temporary directory) and prints its path. Each program's log holds
the build and execution output, each command and its working directory. Log
filenames carry an index, so tests with the same name in different
directories do not collide.

**JSON report.** The report has `"version": 1`, selected/completed/not-run
counts, timestamps, and for each program its status, check counts, error
details, log path and executable path. Durations are integer nanoseconds. The
report is written via a temporary file, so an existing report is replaced
only once the new one is complete, and an interrupted run still produces a
report with partial results.

**Interruption.** Ctrl-C or SIGTERM stops the running program; the remaining
programs are counted as not run. On Unix the whole child process group is
terminated, including compiler subprocesses. On other platforms only the
immediate child process is terminated. Waiting for output streams to close is
bounded, so a stray background process cannot hang the runner.
