# Command: lint

```sh
easifem lint filename
```

Compile one Fortran file with strict `gfortran` diagnostics (`-std=f2018`,
`-Wall`, `-pedantic-errors`, `-Wimplicit-interface`, `-Wuse-without-only` and
more; the full list is in `internal/lint.go`). The `include` directory of
every registered package's install prefix is added to the include path.

```sh
easifem lint src/example.F90
```

This is a real compilation, not a read-only check. It writes
`<lintDir>/lib/<file>.o` and module files under `<lintDir>/include`, where
`lintDir` comes from the main configuration.

An optional `linter.toml` in the current directory can set:

```toml
Compiler = "gfortran"
IncludePath = ["$HOME/code/easifem/classes/include"]
LintDir = "$HOME/temp/easifem/lint"
CompilerPath = "/opt/gcc-14/bin"
Flags = ["-Wextra", "-fcheck=all"]
```

- `IncludePath` entries come before the installed include directories.
- `Flags` are appended after the built-in flags, so they can add options or
  override earlier ones (for example `-std=f2008` or `-Wno-conversion`).
- `CompilerPath` may be the compiler executable itself or the directory that
  contains `Compiler`. A relative path is resolved from the current directory.
- `$VAR` references in paths and flags are expanded.
- A malformed `linter.toml` is an error.

A second positional argument (a project name) is accepted for compatibility
but does not change anything.
