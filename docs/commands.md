# Command reference

See [configuration](configuration.md) before first use. Commands normally
initialize the main configuration, package metadata, cache, and generated Fish
shell variables. Even a discovery-only command is not guaranteed to be a
filesystem read-only operation.

## Global flags

| Flag | Effect |
|---|---|
| `--config`, `-c` | Select a main configuration file explicitly. |
| `--env` | Select the environment section; the default flag value is `env`. |
| `--quiet`, `-q` | Suppress the CLI logger and library progress. Failed stages still show recent diagnostics and a log path. Test summaries still print. |
| `--verbose` | Show configuration details and library stage commands, working directories, full output, and log paths; disable animation. |
| `--no-download` | Skip package download in the installation path. Does not skip building or installation. |

Examples:

```sh
easifem --config /path/to/easifem.toml --env development dev classes
easifem --quiet test /path/to/examples --list
```

The `--no-download` flag's legacy help text is misleading; its implementation
skips downloading when the flag is true.

## `easifem test [directory]`

Discover, build and run standalone `*_test_*.F90`/`.f90` programs. The directory
defaults to the current directory. Tests run sequentially with isolated builds,
while runtime fixture paths are resolved from each example's directory.

```sh
easifem test ./examples --list
easifem test ./examples --filter 'Get*' --filter 'DG*'
easifem test ./src/modules --recursive --timeout 5m --json results.json
```

| Flag | Default | Meaning |
|---|---|---|
| `--recursive`, `-r` | false | Search subdirectories, excluding hidden/build directories. |
| `--filter` | all matching programs | Repeatable filename/relative-path glob; alternatives are ORed. |
| `--list` | false | Print selection without compiling or executing examples. |
| `--logdir` | system temporary directory | Parent of a unique per-invocation log directory. |
| `--json` | disabled | Write a version-1 report to a `.json` filename. |
| `--timeout` | `0` | Per-test build plus execution deadline; `0` disables it. |
| `--fail-fast` | false | Stop after the first unsuccessful result. |
| `--rebuild` | false | Clean only the test's isolated build before compiling. |
| `--allow-no-check` | false | Accept `NOCHECK` results without relabeling them as passes. |

Empty selections and invalid filters are errors. `--list` cannot be combined
with `--json`. The command returns nonzero for failures, unrun selected tests,
interruption, or infrastructure/reporting errors. Explicit skips are accepted.
See [the full runner guide](test-runner.md) for statuses, JSON and cancellation.

## `easifem run filename [flags]`

Build and, by default, execute a standalone source file. Run it from the
examples/application directory so its `runner.toml` and runtime data are found.

```sh
easifem run Initiate_test_1.F90
easifem run Initiate_test_1.F90 --no-run
```

| Flag | Effect |
|---|---|
| `--no-run` | Configure and build the executable without running it. |
| `--cache-clean` | Add `--fresh` to CMake configuration. |
| `--rebuild` | Add `--clean-first` to the example build. |

Without `runner.toml`, this older command defaults to Release, a relative
`build` directory, and `easifemBase`. It generates/overwrites `CMakeLists.txt`
in the current directory. In contrast, `easifem test` defaults to Debug and
creates isolated generated projects without replacing source CMake files.
Configuration and build failures stop `run`; an older executable is not then
executed. These flags do not rebuild the installed EASIFEM libraries.

Explicit `--no-run`, `--cache-clean`, and `--rebuild` flags override the
corresponding `runner.toml` values, including explicit false flags. Malformed
`runner.toml` is an error.

Use `--cache-clean`/`--rebuild` only when clearing the example artifacts is
intended; they are not library-clean commands.

## Library progress (`dev` and `install`)

Interactive terminals show one persistent dashboard per package operation.
Download, Configure, Build, and Install update the same panels in place.
Completed stages are listed with checkmarks; counters reset for the new stage,
recent compiler output is retained, and elapsed time covers the whole operation. At 80 columns or wider, progress, elapsed time, and the current
task appear in a rounded box on the left; recent compiler output appears in
a matching box on the right. Narrow terminals use a stacked bordered card.
Borders remain static; updates overwrite only changed text in a single buffered
write. The cursor stays hidden across stage transitions and is restored only when the
package operation completes or fails. Compiler messages update within this area
without scrolling progress down; long lines are clipped to fit. Ninja's default
`[completed/total]` output
updates a solid-block progress bar and the current task description. Completed
blocks and the percentage use bold stage-colored text; the remaining track is
dimmed. Monochrome output keeps the same block shape. This percentage
counts tasks, not elapsed time; custom `NINJA_STATUS` formats may leave the
spinner active. Other build tools use the spinner without a made-up percentage.

Downloads show bytes, average transfer speed, and estimated time remaining when
the protocol reports a content length. Unknown-length transfers and Git clones
use a spinner. No download stage is shown with `--no-download`.

The dashboard retains the last 20 compiler output lines and displays the most
recent lines that fit. At package completion, the boxes remain in scrollback with
a success or failure marker; diagnostics are not printed a second time.
Diagnostics have no logging prefixes. Compiler terminal controls are removed
from the dashboard; raw logs preserve the original output. Each command stage
records raw output, the command, working directory, timestamp, and stage name
in the package's configured build directory:

| Operation | Log file |
|---|---|
| `easifem dev pkgname` | `<BuildDir>/easifem-build-log.txt` |
| `easifem install pkgname` | `<BuildDir>/easifem-install-log.txt` |

Configure and build stages from an installation are included in the install
log. The first command stage replaces that operation's previous log; subsequent
stages append to it, preserving configuration diagnostics when a build fails.
The build and install logs are independent. Missing build directories are created
before opening a log; log creation/write errors fail the operation. These logs
follow the configured build directory, including when it is under `/tmp`, and
are removed with it by `easifem clean`.

Failures stop the operation and print the full log path. Animated operations
print their log path after the dashboard closes; `--verbose` prints it after
each successful stage. Successful operations print a final elapsed time and
installations show their destination.

Animation is disabled with `--quiet`, `--verbose`, a nonempty `CI` environment
variable, `TERM=dumb`, or when stderr is redirected. Redirected/CI output uses
ordinary lines. Interactive headers and borders use stage colors: cyan for
Download, blue for Configure, magenta for Build, and amber for Install. Header text
turns green on completion and red on failure. Backgrounds keep the terminal
default. Colors are disabled
when `NO_COLOR` is present (even if empty); monochrome animation still works.
Styling is applied only to the display and never added to raw command logs.
Quiet mode suppresses successful output; failed stages print up to the last 20
output lines plus the error and log path. Progress and diagnostics use stderr.

```sh
easifem dev classes
easifem install classes --no-download
easifem --verbose dev classes
easifem --quiet install classes --no-download
```

## `easifem dev pkgname`

Build a configured package from its local sources. The CMake development path
uses Ninja, Debug and shared libraries. It neither downloads nor installs the
package. Package names and source/build paths come from configuration.

```sh
easifem dev classes
```

Pass one package name. Although Cobra accepts additional arguments, this
command currently reads only its first package argument.

## `easifem install pkgname... --no-download`

Build and install configured packages, optionally skipping the download step.
The implementation supports CMake and make-based installation. `BuildType`,
paths and build options come from package metadata.

```sh
easifem install base --no-download
easifem install classes --no-download
```

Without `--no-download`, the install path can download package sources.
`extpkgs` selects packages marked external. The `easifem` group currently
selects the configured `base` and `classes` packages in that order; it does not install this
Go CLI. Use `go install ./cmd/easifem` from this repository to install the CLI.
Do not assume a dependency list causes automatic dependency installation;
install required packages explicitly in the appropriate order.

## `easifem clean pkgname...`

Remove the configured build **and install** directories for each package.
Source directories are not the intended cleanup target. The `extpkgs` cleanup
group uses the environment's `extpkgs` list when set. Its fallback selects
only active packages marked `isExtPkg = true`.

This command is destructive and has no documented dry-run mode. Read the
actual configured paths and obtain explicit permission for every cleanup.
Cleanup rejects empty paths and paths containing the current directory,
validates both paths before deletion, and returns deletion errors.
Do not use it to investigate an ordinary failed example.

## `easifem lint filename [projectname]`

Invoke the Fortran compiler with diagnostic flags and configured include paths.
This is compilation, not a read-only source linter: it creates object and module
artifacts under the configured lint directory and reads `linter.toml` from the
current directory when present.

```sh
easifem lint src/example.F90
```

The optional project-name argument is accepted but currently does not change
the compiler invocation. Check `internal/lint.go` before relying on additional
compiler settings; the diagnostic flag list is set by that implementation.

## `easifem setenv`

Currently prints the banner and command help. Directory flags registered on
this command are `--buildDir`/`-b`, `--sourceDir`/`-s`, `--installDir`/`-i`,
and `--lintDir`/`-l`. The command does not write a persistent main configuration.
Edit the TOML configuration for persistent setup. Historical help examples
using `--build`, `--source` or `--install` do not match the registered flag names.

## `easifem version`

Prints the easifem CLI release version, currently `2026.10.31`.
`easifem --version` prints the same version without configuration initialization.
The banner uses the same release value. To inspect a Go executable's embedded
build information, use `go version -m` on its path.
