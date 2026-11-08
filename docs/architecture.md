# Architecture: Repository map

| Path | Responsibility |
|---|---|
| `cmd/easifem/main.go` | CLI entry point |
| `internal/root.go` | Root command, global flags, configuration initialization |
| `internal/common.go` | Configuration types, shared state, legacy help text |
| `internal/pkg.go` | Package metadata, builds, downloads, cache and shell-file generation |
| `internal/dev.go`, `install.go`, `clean.go` | Library/package lifecycle commands |
| `internal/run.go` | Single-program build/run command and CMake generation |
| `internal/test.go` | Test command, flags, reporting and exit policy |
| `internal/test_executor.go` | CMake/runner.toml adapter for standalone examples |
| `internal/testrunner/` | Standard-library-only discovery, execution policy, result parsing and reports |
| `docs/` | User and AI-agent documentation |
