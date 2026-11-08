# EASIFEM CLI operating guide for AI agents

This guide applies when using `easifem` to work on EASIFEM libraries and
examples. Agents changing the CLI itself should also read
[the repository AGENTS.md](../AGENTS.md).

## Establish scope and authorization

1. Read the target repository's `AGENTS.md` and the user's latest restrictions.
2. Inspect the relevant source, configuration and working-tree changes.
3. Separate the intended operations: source editing, Go unit tests, Go CLI
   build/install, Fortran library build/install, example execution and cleanup.
4. Execute only the operations authorized by the user. Permission for a focused
   example run does not authorize a library install, broader suite or cleanup.
5. If execution is prohibited, prepare the change and give the user precise
   validation commands. Do not bypass the restriction with `go run`, direct
   CMake/compiler calls, a Fish wrapper, or another binary.

Explicit approval remains valid within its stated scope. Do not ask repeatedly
for the same approved operation, but respect newer restrictions and ask when
an additional operation requires authorization. Cleaning requires fresh explicit
permission for each invocation.

## Operation reference

| Intention | Command | Actual effect |
|---|---|---|
| Inspect command definitions | Read `internal/*.go` | Source inspection only. |
| Preview example selection | `easifem test DIR --list` | No Fortran build/run; normal CLI configuration/cache initialization still occurs. |
| Build a library for development | `easifem dev classes` | Local Debug build, no install/download. |
| Build/install local library sources | `easifem install classes --no-download` | Builds and installs; download is skipped. |
| Run one example | `easifem run Name_test_1.F90` | Generates source-dir CMakeLists, builds and executes. |
| Run an example suite | `easifem test DIR` | Isolated executable builds, sequential execution, logs and statuses. |
| Diagnose using `easifem lint` | `easifem lint file.F90` | Invokes a compiler and writes object/module artifacts. |
| Remove package artifacts | `easifem clean classes` | Deletes configured build and install directories. |
| Build the Go CLI | `go build -o /tmp/easifem-review ./cmd/easifem` | Compiles Go; creates a review binary and updates Go caches. |
| Install the Go CLI | `go install ./cmd/easifem` | Builds and replaces the executable in GOBIN/default GOPATH bin. |
| Test the Go implementation | `go test ./internal/...` | Builds and runs Go tests; unit tests do not run Fortran examples. |

The `install` subcommand manages configured packages. It is not the command for
installing the Go CLI itself. `--no-download` does not mean dry-run or no-build.
`--no-run` on `easifem run` still builds the example executable. Explicit
`--no-run` overrides `IsExecute` from `runner.toml`.

## Prefer the reusable test runner

Use `easifem test` for standalone Fortran suites rather than introducing new
module-specific shell runners. Keep existing Fish scripts during a transition
unless the user asks to remove them.

When execution is authorized, start with the smallest relevant selection:

```sh
easifem test /path/to/examples --list
easifem test /path/to/examples --filter 'DG*' \
  --timeout 5m --json /tmp/dg-results.json
```

If broader verification is appropriate and authorized, omit the filter or use
`--recursive` on a modules directory. Do not launch a whole repository suite
merely because the previous focused run passed.

Examples depend on installed libraries. If library source changed, running an
example without rebuilding/installing that library can exercise older code.
Explain the additional required operation instead of silently performing it.

`easifem test` builds only example executables and preserves their source-dir
CMakeLists files. It runs programs from their own directory so relative meshes
and fixtures work. Programs can still write output files in that directory.
Use actual configured paths rather than assuming a machine-specific build or
install location.

## Interpret results accurately

- `PASS` means successful build/execution with passing recognized assertions.
- `FAIL` means an assertion or flat TAP consistency failure.
- `COMPILE` covers build setup, configuration, compilation and linking errors.
- `CRASH` means execution failure; earlier passing assertions do not override it.
- `TIMEOUT` and `CANCELLED` remain separate from ordinary failures.
- `NOCHECK` is not a verified pass. Do not add `--allow-no-check` just to make a
  failing command succeed; use it only when accepting unchecked programs is
  explicitly intended.
- Explicit `SKIP` results are reported as such, not as passing assertions.
- A nonzero suite exit can also indicate unrun selected tests or an
  infrastructure/reporting error. Inspect the result and log rather than
  inferring a compiler error from the exit code alone.

Read the per-test logs and optional JSON report. The report has `version: 1`,
run metadata, per-test statuses/check counts, log paths and error details;
durations are nanoseconds. An interrupted or fail-fast run may have a nonzero
`not_run` count. See [the runner guide](test-runner.md).

Do not infer success from the existence of an executable: it may belong to an
older build. The new runner isolates builds and distinguishes build and runtime
errors. Do not assume every legacy command has the same error-propagation
implementation; inspect relevant logs and artifacts.

## Report work and preserve user changes

State what changed, what commands actually ran, their outcome and material
limitations. Distinguish results reported by the user from checks you ran.
For unexpected failures, give the example name, status, useful error and log
path. Preserve the evidence; do not delete failed builds/logs to tidy up.

Never modify a user's source examples or fixture data simply to obtain a pass.
Fix the actual issue within the approved scope, keep documentation/examples in
sync, and follow the target repository's bug-report instructions. In review-only
mode, report findings without editing even a bug-report file.

Do not commit or push without a user request. A documentation-only task does
not authorize builds, installs, tests or cleanup.
