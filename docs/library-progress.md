# Library progress and logs

`easifem dev` and `easifem install` share the same progress display and log
format.

## Logs

Each command stage writes its raw output to a log in the package's configured
build directory, along with the command, working directory, timestamp and stage
name:

| Operation                 | Log file                             |
| ------------------------- | ------------------------------------ |
| `easifem dev pkgname`     | `<buildDir>/easifem-build-log.txt`   |
| `easifem install pkgname` | `<buildDir>/easifem-install-log.txt` |

- The first stage of an invocation replaces that operation's previous log, and
  later stages append to it. A failed build therefore still shows the
  configure output from the same run.
- The build and install logs are independent of each other.
- A missing build directory is created before the log is opened. If the log
  cannot be created or written, the operation fails.
- Logs live wherever `buildDir` points (including `/tmp`) and are deleted with
  it by [`easifem clean`](clean.md).
- On failure, the operation stops and prints the full log path. In animated
  mode the log path is printed after the dashboard closes; `--verbose` also
  prints it after each successful stage.

## Interactive dashboard

In an interactive terminal, each package operation shows one dashboard that
stays in place while the stages (Download, Configure, Build, Install) run:

- Completed stages are ticked off. Counters reset at each stage; elapsed time
  covers the whole operation.
- At 80 columns or wider, progress and the current task are in a box on the
  left and recent compiler output in a box on the right. Narrower terminals
  use a single stacked card.
- Ninja's default `[completed/total]` status drives a progress bar and the
  current task. The percentage counts build tasks, not time. A custom
  `NINJA_STATUS` format, or a non-Ninja tool, shows a spinner instead.
- Downloads show bytes, average speed and time remaining when the size is
  known; otherwise (including Git clones) a spinner.
- The last 20 compiler output lines are kept, and as many as fit are shown.
  Long lines are clipped. Terminal control sequences from the compiler are
  removed from the display but kept in the log.
- When the operation ends, the dashboard remains in scrollback with a success
  or failure marker; diagnostics are not printed twice. A successful operation
  prints its elapsed time, and `install` also prints the install directory.

Stage colors: cyan for Download, blue for Configure, magenta for Build, and
amber for Install. Headers turn green on success and red on failure. Set
`NO_COLOR` (any value, including empty) to disable color; the animation still
works in monochrome. Styling never reaches the log files.

## Non-interactive output

Animation is disabled by `--quiet`, `--verbose`, a nonempty `CI` environment
variable, `TERM=dumb`, or redirected stderr. Output is then printed as
ordinary lines.

- `--verbose` prints each stage's command, working directory and full output.
- `--quiet` prints nothing on success. A failed stage prints up to its last 20
  output lines, the error and the log path.

All progress and diagnostics go to stderr.

```sh
easifem --verbose dev classes
easifem --quiet install classes --no-download
CI=1 easifem install base --no-download 2>install.txt
```
