# Command: run

```sh
easifem run filename [flags]
```

Build one Fortran program against the installed EASIFEM libraries and, by
default, execute it. Run it from the program's directory so that its
`runner.toml` and runtime data files are found.

```sh
easifem run Initiate_test_1.F90
easifem run Initiate_test_1.F90 --no-run
easifem run Initiate_test_1.F90 --rebuild
```

| Flag            | Effect                                                 |
| --------------- | ------------------------------------------------------ |
| `--no-run`      | Configure and build the executable without running it. |
| `--cache-clean` | Pass `--fresh` to CMake configuration.                 |
| `--rebuild`     | Pass `--clean-first` to the build.                     |

> **Note:** `run` writes a generated `CMakeLists.txt` into the current
> directory, replacing any existing one. Do not use it in a directory whose
> own CMake project must be kept. For suites of `*_test_*` programs, use
> [`easifem test`](test.md), which builds in isolated directories.

## Behavior

- Settings come from `runner.toml` in the current directory (see
  [configuration](configuration.md#runnertoml)). Without one, `run` uses a
  `Release` build in `./build`, Ninja and the `easifemBase` target, and lets
  CMake pick the Fortran compiler (for example from `FC`). Malformed
  `runner.toml` is an error.
- `Compiler`/`CompilerPath`, `IncludePath`, `ExtraIncludePath`,
  `LibraryPath` and `Flags` are applied as in `easifem test`; relative paths
  are resolved from `SourceDir`.
- Additional filenames after the first are compiled into the same
  executable, as are `ExtraFileNames` from `runner.toml`.
- Explicit flags override the matching `runner.toml` keys (`IsExecute`,
  `CacheClean`, `ReBuild`), including flags explicitly set to false such as
  `--rebuild=false`.
- If configuration or compilation fails, `run` stops; an older executable is
  never executed.
- The program's output is forwarded directly to the terminal.
- `--cache-clean` and `--rebuild` affect only this program's build directory.
  They do not rebuild the installed EASIFEM libraries.
