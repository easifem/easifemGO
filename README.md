# EASIFEM CLI

`easifemGO` provides the `easifem` command-line program. It configures local
EASIFEM packages, builds and installs the Fortran libraries, and builds and
runs Fortran applications and examples against those installed libraries.

## Installation

Install the latest release with Go:

```sh
go install github.com/easifem/easifemGO/cmd/easifem@latest
```

Or install from a clone of this repository:

```sh
go install ./cmd/easifem
```

`go install` writes the executable to `GOBIN`, or to `$(go env GOPATH)/bin`
when `GOBIN` is unset. Make sure that directory is on your `PATH`.

Installing the CLI does not install EASIFEM's Fortran libraries. After
[configuring](docs/configuration.md) the CLI, install those with
`easifem install` (for example, `easifem install easifem`).

## Common workflows

Run these from the appropriate repository or examples directory. They require
a working EASIFEM [configuration](docs/configuration.md).

```sh
# Build local classes sources in Debug, without installing them
easifem dev classes

# Build and install local sources without downloading them again
easifem install classes --no-download

# Build and execute one program from its examples directory
easifem run Initiate_test_1.F90

# List the standalone tests in a directory without building them
easifem test /path/to/classes/src/modules/FEDOF/examples --list

# Run selected tests and keep a JSON report
easifem test /path/to/classes/src/modules/FEDOF/examples \
  --filter 'DG*' --timeout 5m --json /tmp/dg-results.json
```

To run the complete suite, omit `--filter`. To include subdirectories, add
`--recursive`.

`easifem run` and `easifem test` compile examples against the **installed**
libraries; they do not rebuild or install those libraries. After changing
library source, run `easifem install <pkg> --no-download` first if you want
the examples to use your changes.

## User guide

- [Command reference](docs/commands.md): global flags and an overview of every
  command.
- [Configuration](docs/configuration.md): main configuration, package plugins,
  `runner.toml`, generated files and troubleshooting.
- Commands: [`dev`](docs/dev.md), [`install`](docs/install.md),
  [`run`](docs/run.md), [`test`](docs/test.md), [`clean`](docs/clean.md),
  [`lint`](docs/lint.md), [`setenv`](docs/setenv.md).
- [Library progress and logs](docs/library-progress.md): the `dev`/`install`
  dashboard and build logs.

`easifem <command> --help` shows the flags of the installed executable. Some
older banner and long-help text describes historical behavior; these guides
describe the current implementation.

## Developer guide

- [Development](docs/development.md): building, testing and validating changes
  to the CLI.
- [Architecture](docs/architecture.md): the repository map and design rules.
- [AGENTS.md](AGENTS.md): instructions for AI coding agents.
