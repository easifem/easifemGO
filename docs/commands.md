# Command Reference

See [configuration](configuration.md) before first use. Commands normally
initialize the main configuration, package metadata, cache, and generated Fish
shell variables. Even a discovery-only command is not guaranteed to be a
filesystem read-only operation.

## Global Flags

| Flag             | Effect                                                                                                                                |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `--config`, `-c` | Select a main configuration file explicitly.                                                                                          |
| `--env`          | Select the environment section; the default flag value is `env`.                                                                      |
| `--quiet`, `-q`  | Suppress the CLI logger and library progress. Failed stages still show recent diagnostics and a log path. Test summaries still print. |
| `--verbose`      | Show configuration details and library stage commands, working directories, full output, and log paths; disable animation.            |
| `--no-download`  | Skip package download in the installation path. Does not skip building or installation.                                               |

Examples:

```sh
easifem --config /path/to/easifem.toml --env development dev classes
easifem --quiet test /path/to/examples --list
```

The `--no-download` flag's legacy help text is misleading; its implementation
skips downloading when the flag is true.

## `easifem test [directory]`

Documentation of test command is given here [test](test.md)

## `easifem run filename [flags]`

Documentation of run command is given here [run](run.md)

## `easifem dev filename [flags]`

Documentation of dev command is given here [dev](dev.md)

## `easifem install filename [flags]`

Documentation of install command is given here [install](install.md)

## `easifem clean pkgname...`

Documentation of clean command is given here [clean](clean.md)

## `easifem lint filename [projectname]`

Documentation of lint command is given here [lint](lint.md)

## `easifem setenv`

Documentation of setenv command is given here [setenv](setenv.md)

## `easifem version`

Prints the easifem CLI release version, currently `2026.10.31`.
`easifem --version` prints the same version without configuration initialization.
The banner uses the same release value. To inspect a Go executable's embedded
build information, use `go version -m` on its path.
