# Command reference

Read [configuration](configuration.md) before first use.

Every command except `easifem --version` and `easifem --help` first runs
normal start-up. Start-up loads the main configuration and package plugins,
creates the lint directory, and rewrites `cache.toml` and `easifemvar.fish` next
to the main configuration. So even a command that only lists files, such as
`easifem test --list`, writes to the configuration directory.

## Commands

| Command                                   | Purpose                                                       | Details                    |
| ----------------------------------------- | ------------------------------------------------------------- | -------------------------- |
| `easifem dev pkgname`                     | Debug build of a package from local sources; no install.      | [dev](dev.md)              |
| `easifem install pkgname...`              | Download (optional), build and install packages.              | [install](install.md)      |
| `easifem run filename [flags]`            | Build and run one program in the current directory.           | [run](run.md)              |
| `easifem test [directory] [flags]`        | Discover, build and run standalone test programs.             | [test](test.md)            |
| `easifem clean pkgname...`                | **Delete** package build and install directories.             | [clean](clean.md)          |
| `easifem lint filename`                   | Compile one file with strict diagnostic flags.                | [lint](lint.md)            |
| `easifem setenv`                          | Print help; it does not write configuration.                  | [setenv](setenv.md)        |
| `easifem version`                         | Print the CLI release version.                                | [below](#easifem-version)  |

## Global flags

| Flag             | Effect                                                                                                                                  |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `--config`, `-c` | Use this main configuration file instead of searching for `easifem.toml`.                                                               |
| `--env`          | Select the environment section. Without the flag, `envName` from the configuration file is used, falling back to `env`.                |
| `--quiet`, `-q`  | Suppress the CLI logger and library progress. Failed stages still show recent diagnostics and a log path. Test summaries still print.  |
| `--verbose`      | Show configuration details, each library stage's command, working directory, full output and log path. Disables animation.             |
| `--no-download`  | Skip the download stage of `easifem install`. Building and installing still happen.                                                    |

```sh
easifem --config /path/to/easifem.toml --env development dev classes
easifem --quiet test /path/to/examples --list
```

## `easifem version`

Prints `easifem version <release>`. `easifem --version` prints the same
release without running start-up, so it works without a configuration. The
release string is the `Version` constant in `internal/version.go`. To see the
Go build information embedded in an executable, run
`go version -m "$(command -v easifem)"`.
