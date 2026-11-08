# Command: lint

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
