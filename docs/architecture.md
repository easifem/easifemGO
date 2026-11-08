# Architecture

## Repository map

| Path                                    | Responsibility                                                                          |
| --------------------------------------- | --------------------------------------------------------------------------------------- |
| `cmd/easifem/main.go`                   | Entry point; calls `internal.Execute`.                                                  |
| `internal/root.go`                      | Root command, global flags, start-up (configuration, plugins, cache, shell file).       |
| `internal/common.go`                    | Constants, shared state, `Pkg`/`Runner`/`Linter` types and legacy long-help text.       |
| `internal/pkg.go`                       | Plugin loading and defaults, download, CMake/make install stages, cache and Fish file.  |
| `internal/dev.go`, `internal/install.go` | `dev` and `install` commands, including the `easifem`/`extpkgs` groups.               |
| `internal/clean.go`                     | `clean` command and its path-safety checks.                                             |
| `internal/build_output.go`              | `dev`/`install` dashboard, stage logs, colors and non-interactive output.               |
| `internal/install_progress_tracking.go` | Download byte-progress adapter for go-getter.                                           |
| `internal/spinner.go`                   | Bubble Tea spinner component.                                                           |
| `internal/run.go`                       | `run` command, `CMakeLists.txt` generation and subprocess execution.                    |
| `internal/test.go`                      | `test` command: flags, signal handling, reporting and exit policy.                      |
| `internal/test_executor.go`             | Adapter implementing `testrunner.Executor` with isolated CMake builds and `runner.toml`.|
| `internal/testrunner/`                  | Standard-library-only discovery, execution policy, result parsing and reports.          |
| `internal/lint.go`                      | `lint` command.                                                                         |
| `internal/setenv.go`, `internal/version.go` | `setenv` and `version` commands; `Version` constant.                               |
| `internal/*_test.go`                    | Unit tests for the CLI package.                                                         |
| `test/`                                 | Hand-run Fortran fixtures for `run`/`lint` (not Go tests).                              |
| `docs/`                                 | User and developer documentation.                                                       |

## Design rules

- **Runner independence.** `internal/testrunner` must not import Cobra,
  Viper, TOML or `internal`, and uses only the Go standard library. Builds and
  executions go through the `testrunner.Executor` interface; CMake and
  `runner.toml` handling live in the adapter (`internal/test_executor.go`).
- **Processes.** Launch an executable with separate arguments, never a shell
  string. Set `exec.Cmd.Dir` instead of calling `os.Chdir`. On Unix,
  cancellation kills the process group so compiler descendants stop too.
- **Errors.** Reusable functions return wrapped errors (`%w`) rather than
  calling `log.Fatal` or `os.Exit`. Real start/wait/exit errors are
  propagated, and an older executable is never run after a failed build.
- **Output.** Captured stdout/stderr is bounded in memory; full output goes
  to log files.

Some older commands (`clean`, `lint`, `run`, `setenv`) still call `log.Fatal`
directly; new code should follow the rules above.
