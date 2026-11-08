# Command: Run

```sh
easifem run filename [flags]
```

Build and, by default, execute a standalone source file. Run it from the
examples/application directory so its `runner.toml` and runtime data are found.

```sh
easifem run Initiate_test_1.F90
easifem run Initiate_test_1.F90 --no-run
```

| Flag            | Effect                                                 |
| --------------- | ------------------------------------------------------ |
| `--no-run`      | Configure and build the executable without running it. |
| `--cache-clean` | Add `--fresh` to CMake configuration.                  |
| `--rebuild`     | Add `--clean-first` to the example build.              |

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
