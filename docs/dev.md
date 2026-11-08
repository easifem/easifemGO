# Command: Dev and Install

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

## Details

Interactive terminals show one persistent dashboard per package operation.
Download, Configure, Build, and Install update the same panels in place.
Completed stages are listed with checkmarks; counters reset for the new stage,
recent compiler output is retained, and elapsed time covers the whole operation.

At 80 columns or wider, progress, elapsed time, and the current
task appear in a rounded box on the left; recent compiler output appears in
a matching box on the right.

Narrow terminals use a stacked bordered card.
Borders remain static; updates overwrite only changed text in a single buffered
write.

The cursor stays hidden across stage transitions and is restored only when the
package operation completes or fails. Compiler messages update within this area
without scrolling progress down; long lines are clipped to fit.

Ninja's default `[completed/total]` output updates a solid-block progress bar
and the current task description. Completed
blocks and the percentage use bold stage-colored text; the remaining track
is dimmed. Monochrome output keeps the same block shape. This percentage
counts tasks, not elapsed time; custom `NINJA_STATUS` formats may leave
the spinner active. Other build tools use the spinner without a
made-up percentage.

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

| Operation                 | Log file                             |
| ------------------------- | ------------------------------------ |
| `easifem dev pkgname`     | `<BuildDir>/easifem-build-log.txt`   |
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
