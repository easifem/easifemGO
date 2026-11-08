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
```

`IncludePath` entries are added before the installed include directories.
`Flags` in `linter.toml` is ignored: the built-in flag list always applies.
`CompilerPath` is read but not used.

A second positional argument (a project name) is accepted for compatibility
but does not change anything.
