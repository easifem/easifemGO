# EASIFEM CLI documentation

`easifemGO` provides the `easifem` command-line program for configuring local
EASIFEM packages, building/installing libraries, and building/running Fortran
applications and examples.

## Guides

- [Command reference](commands.md): commands, flags and important side effects.
- [Configuration](configuration.md): main configuration, package plugins,
  `runner.toml`, and generated files.
- [Standalone test runner](test-runner.md): discovery, results, logs, JSON,
  cancellation and validation.
- [AI-agent operating guide](ai-guide.md): authorization and reliable workflows
  for agents using the CLI in EASIFEM repositories.
- [Repository development instructions](../AGENTS.md): instructions for models
  modifying this Go codebase.

## Common workflows

Run these from the appropriate repository or examples directory. A working
EASIFEM configuration and installed library dependencies are required.

```sh
# Build local classes sources in Debug without installing
easifem dev classes

# Build and install local sources without downloading them again
easifem install classes --no-download

# Build and execute one standalone program from its examples directory
easifem run Initiate_test_1.F90

# Preview all standalone tests in a directory
easifem test /path/to/classes/src/modules/FEDOF/examples --list

# Run selected tests and retain a JSON report
easifem test /path/to/classes/src/modules/FEDOF/examples \
  --filter 'DG*' --timeout 5m --json /tmp/dg-results.json
```

For the complete suite, omit `--filter`. For subdirectories, add `--recursive`.
Test commands compile examples against installed libraries; they do not rebuild
or install those libraries. Changes to library source must first be built and
installed when you want examples to exercise them.

AI agents must follow the user's execution restrictions and repository
instructions before using these commands. Examples here are not approval to
build, install, test, or clean.

## Building the Go CLI

From the `easifemGO` repository:

```sh
go test ./internal/...
go build -o /tmp/easifem-review ./cmd/easifem
go install ./cmd/easifem
```

`go install` uses `GOBIN`, or the Go default under `GOPATH/bin` when `GOBIN` is
unset. Ensure that directory is on `PATH`. Installing the CLI does not install
EASIFEM's Fortran libraries.

Use `easifem test --help` for the installed runner's flags. Source files and
these guides describe the current implementation; some older banner/help text
still describes historical behavior.
