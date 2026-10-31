# Configuration

The CLI has three separate configuration concerns:

1. Main configuration selects an environment and root directories.
2. Package plugins describe library sources, build/install paths and options.
3. `runner.toml` describes an individual examples/application directory.

Changing one does not automatically rebuild a library or refresh an installed
Go executable.

## Main configuration

Use `--config` to select a file explicitly:

```sh
easifem --config /path/to/easifem.toml --env development test ./examples --list
```

Without it, the CLI registers these search directories, in this order, for
`easifem.toml`:

1. `$HOME`
2. `$HOME/easifem`
3. `$HOME/.config`
4. `$HOME/.config/easifem`
5. The current directory

A local configuration is not automatically preferred over an existing home
configuration. Use `--config` when the choice matters.

A minimal environment example:

```toml
envName = "development"

[development]
sourceDir = "$HOME/code/easifem"
buildDir = "$HOME/temp/easifem/build"
installDir = "$HOME/.local/easifem"
lintDir = "$HOME/temp/easifem/lint"
```

The environment name is selected through `envName`/`--env`; directory keys are
read from that environment's section. `$HOME` and other environment variables
in directory settings are expanded by the CLI. Do not assume that a literal
`~` is expanded the same way.

Keep a `plugins` directory next to the selected main configuration. Startup
expects to read package definitions from there. A working example runner also
needs the relevant installed library packages registered, or appropriate extra
CMake prefix paths.

## Package plugins

A package named `classes` is read from `plugins/classes.toml` next to the main
configuration. Example:

```toml
name = "classes"
isActive = true
isExtPkg = false
sourceDir = "$HOME/code/easifem/classes"
buildDir = "$HOME/temp/easifem/build/classes"
installDir = "$HOME/.local/easifem/classes"
git = "github.com/easifem/classes.git"
buildSystem = "cmake"
buildType = "Debug"
buildSharedLibs = true
buildStaticLibs = false
targetName = "easifemClasses"
buildOptions = ["-D USE_LIS:BOOL=ON"]
```

Use paths and options matching your installation; the example is not a required
layout. Register `base` and external libraries similarly when needed. The
`name` field is required for active packages. `isExtPkg = true` puts a package
in the external-package group.

An inactive package is skipped by package operations and cache construction.
If a plugin omits its build type, the current package fallback is Release.
`easifem dev` explicitly uses Debug regardless of that install setting.

`CmakePrefixPath` supplies additional package-search prefixes. `EnvVars` and
`LdLibraryPath` can provide package-specific environment/library path settings.
The implementation's `Dependencies` field does not establish automatic
installation order; install dependencies explicitly. The `easifem` install
group installs `base` before `classes`. Environment entries inherited from
main configuration must use `KEY=value`; values may contain additional `=`
characters and malformed entries produce a configuration error.

Keep only intended package-definition files in `plugins`: the current package
name discovery enumerates its directory entries rather than filtering a clean
list of `.toml` files.

## Generated configuration files

Normal root initialization:

- Loads the main configuration and package definitions.
- Creates the configured lint directory if needed.
- Builds the installed-prefix/library-path cache and writes `cache.toml`
  next to the main configuration.
- Writes `easifemvar.fish` there with Fish commands extending
  `LD_LIBRARY_PATH`.

The configuration directory must therefore be writable. Cached prefixes come
from package metadata; their presence does not prove the libraries have been
built and installed. Generating the Fish file does not source it into your
current terminal. A comment in the source mentions a shell `.sh` file as well,
but the current writer creates only the `.fish` file.

`easifem test --list` does not compile or execute Fortran, but it still uses
normal root initialization. Do not describe it as a completely read-only CLI
invocation.

## `runner.toml`

Place this file in each examples/application directory. For classes examples:

```toml
BuildType = "Debug"
BuildDir = "/tmp/easifemClasses/FEDOF"
TargetLibs = ["easifemClasses"]
```

`TargetLibs` names installed CMake packages/targets. The test adapter defaults
to `easifemBase` if the list is empty. The package cache supplies common CMake
search prefixes.

Additional settings accepted by the test adapter include:

```toml
Compiler = "gfortran"
TargetName = "main.out"
SourceDir = "."
ExtraFileNames = ["helper_module.F90"]
ExtraIncludePath = ["include"]
ExtraCMakePrefixPath = ["/path/to/another/install"]
ExtraLibs = ["m"]
Flags = ["-Wall"]
```

For `easifem test`:

- Main source is always the discovered test; `FileName` does not replace it.
- `BuildDir` is a parent for fresh per-test workspaces, not a shared executable
  directory. The default parent is under the system temporary directory.
- Relative source-related paths are resolved from `SourceDir`, which defaults
  to the example directory. Runtime cwd always remains the example directory.
- Missing `runner.toml` uses defaults; malformed TOML is a `COMPILE` result.
- `IsExecute`/`CacheClean` do not disable execution or change the fresh-workspace
  policy. See [the runner guide](test-runner.md) for all details.

For `easifem run`, the older defaults are Release and a relative `build`
directory. It uses the configured build location directly and overwrites a
`CMakeLists.txt` in the current directory. Avoid using it in a directory whose
handwritten CMake project must be preserved; use `easifem test` for suites.
The `run` decoder rejects malformed TOML. Explicit command-line flags override
`IsExecute`, `CacheClean`, and `ReBuild` settings from the file.

## Troubleshooting

| Symptom | Check |
|---|---|
| Main configuration not found | Selected `--config` path or configured search directories. |
| Cannot read package metadata | `plugins` beside the selected configuration; package name and TOML syntax. |
| Cannot find `easifemClasses` or `easifemBase` | The registered install prefix and actual installed CMake package files. |
| Missing `.mod` files or unresolved symbols | Build/install of matching library sources and the example's `TargetLibs`. |
| Missing mesh or fixture | Run from the examples directory; check relative paths and available data. |
| `NOCHECK` | Execution exited successfully but emitted no recognized assertions; inspect the log. |
| Unexpected CLI behavior after editing Go source | Check `command -v easifem`; rebuild/install the intended CLI when authorized. |

Do not use `easifem clean` as a first troubleshooting step. It removes install
artifacts as well as build artifacts.
