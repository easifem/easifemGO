# AGENTS.md — easifemGO

Instructions for AI coding agents. Human-facing documentation starts at
[README.md](README.md); this file adds only what agents need beyond it.

This repository is the Go implementation of the `easifem` CLI (module
`github.com/easifem/easifemGO`, Go baseline 1.21.6, Cobra commands, Viper/TOML
configuration). Before changing a command, read its page in `docs/` and its
source; the long-help strings in `internal/common.go` are often out of date.

## 1. Authorization

Many `easifem` and Go commands compile code, execute programs, delete
directories, or replace the user's installed tools. Run them only with
explicit permission from the user's current instructions.

| Needs explicit permission                                                     | Why                                                        |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------- |
| `go test`, `go run`, `go build`                                               | Compile and (for `test`/`run`) execute code.               |
| `go install ./cmd/easifem`                                                    | Replaces the user's installed CLI.                         |
| `easifem dev`, `easifem install`                                              | Build (and install) Fortran libraries.                     |
| `easifem run`, `easifem test` (without `--list`)                              | Build and execute Fortran; `run` overwrites `./CMakeLists.txt`. |
| `easifem lint`                                                                | Compiles and writes object/module files.                   |
| `easifem clean` — **fresh permission for every invocation**                   | Deletes build **and install** directories; no dry run.     |
| `gofmt -w`, commits, pushes, dependency changes                               | Edit files or shared state.                                |

Allowed without asking: reading files, `git status`/`git diff`/`git log`,
`gofmt -l`/`gofmt -d`, and `easifem test DIR --list`. Note that even
`--list` (like every command except `--help`/`--version`) rewrites
`cache.toml` and `easifemvar.fish` in the configuration directory.

Rules:

- Permission covers only what it names. Approval for one focused example run
  does not cover a broader suite, a library install, or a cleanup. Do not
  re-ask for an operation already approved in scope; newer restrictions win.
- Do not work around a restriction with `go run`, direct `cmake`/compiler
  calls, shell wrappers or another binary.
- A documentation-only or review-only task authorizes no builds, tests,
  installs or edits (not even to a bug-report file in review-only mode).
- When validation is not authorized, finish the change and the static review,
  then give the user focused commands to run and say which checks were not
  run. Never claim a suggested command succeeded.
- Do not commit or push unless asked. Keep unrelated user changes intact.

## 2. Changing this repository

The repository map and design rules are in
[docs/architecture.md](docs/architecture.md); build and validation steps are in
[docs/development.md](docs/development.md). In short:

- `internal/testrunner` uses only the Go standard library and must not import
  Cobra, Viper, TOML or `internal`. Inject execution via
  `testrunner.Executor`; CMake/`runner.toml` logic belongs in
  `internal/test_executor.go`.
- Launch executables with separate arguments, set `exec.Cmd.Dir` (no
  `os.Chdir`), propagate real start/wait/exit errors, and never run an older
  executable after a failed build. Never pass an `error` by value into a
  goroutine expecting it to be updated.
- Return wrapped errors (`%w`) from reusable code; no `log.Fatal`/`os.Exit`
  outside command handlers.
- Keep captured output bounded in memory with full output in logs. Preserve
  cancellation and Unix process-group cleanup.
- Do not raise the Go baseline or add dependencies without a concrete need
  and the user's agreement.
- Add focused tests (fake executors, temporary directories, failure paths).
  Writing tests does not authorize running them.
- Update the affected `docs/` page and `--help` text in the same change.
  Document only implemented behavior.

Test-runner invariants (do not break without discussing it first):

- Discover sorted `*_test_*.F90`/`.f90` programs; an empty selection is an
  error.
- Run sequentially, with the example directory as the runtime working
  directory.
- Build in isolated directories; never overwrite a source `CMakeLists.txt` or
  rebuild installed libraries from `easifem test`.
- Parse checks only from execution output. A process failure overrides prior
  passes. No checks is `NOCHECK`, accepted only with `--allow-no-check`.
- Keep `COMPILE`, `FAIL`, `CRASH`, `TIMEOUT` and `CANCELLED` distinct.
- Keep logs and partial JSON on interruption. Do not change the version-1
  report schema without documenting the compatibility impact.

Suggested commands to offer the user (see [development](docs/development.md)):

```sh
go test ./internal/...                       # not ./internal/ — that skips testrunner
go build -o /tmp/easifem-review ./cmd/easifem
/tmp/easifem-review test /path/to/examples --list
/tmp/easifem-review test /path/to/examples --filter 'DG*' --timeout 5m --json /tmp/dg.json
```

## 3. Using the CLI in other EASIFEM repositories

When working on easifemBase, easifemClasses or a kernel, follow that
repository's own `AGENTS.md` first; this section covers only the CLI.

Pitfalls that cause wrong conclusions:

- **Examples link the installed library.** `easifem run`/`easifem test`
  never rebuild libraries, and `easifem dev` builds without installing. After
  editing library source, examples exercise old code until
  `easifem install <pkg> --no-download` runs. Tell the user this rather than
  installing silently.
- **`--no-download` is not a dry run.** It still builds and installs.
- **`easifem install` installs Fortran packages, not this CLI.** The CLI is
  installed with `go install ./cmd/easifem`.
- **`easifem run --no-run` still builds.** `run` also overwrites
  `./CMakeLists.txt`; prefer `easifem test` for `*_test_*` programs.
- **`easifem clean` is never a troubleshooting step.** It removes the install
  directory too.
- **An existing executable proves nothing**; it may be from an older build.
  Use the status and log of the current run.
- Use the paths from the user's actual configuration; do not assume a build
  or install location.

Recommended test workflow, once execution is authorized:

1. `easifem test DIR --list` to confirm the selection.
2. Run the smallest relevant subset: `--filter 'Name*' --timeout 5m --json
   /tmp/results.json`.
3. Widen (no filter, or `--recursive`) only if the user wants broader
   verification; a passing focused run is not a reason to launch everything.

Prefer `easifem test` to writing new module-specific shell runners, but keep
existing Fish scripts unless the user asks to remove them.

Interpreting results (details in [docs/test.md](docs/test.md#results)):

- `PASS` is the only verified pass. `NOCHECK` means "ran, but nothing was
  checked"; do not add `--allow-no-check` just to make a command succeed.
- `COMPILE` is a build problem, `CRASH` a runtime failure (earlier `ok` lines
  do not count), `FAIL` a failed check. `TIMEOUT`/`CANCELLED` are not code
  failures by themselves.
- A nonzero exit can also mean unrun programs or a reporting error. Read the
  per-test log (path printed under `Logs:`) and the JSON report before
  diagnosing; JSON durations are nanoseconds.

Reporting:

- State what changed, which commands actually ran, and their results.
  Separate results the user reported from checks you ran.
- For a failure, give the example name, status, key error and log path. Keep
  failed build workspaces and logs; do not delete evidence.
- Never edit an example, fixture or mesh just to make it pass. Fix the real
  issue within the approved scope, keeping docs and examples in sync.
