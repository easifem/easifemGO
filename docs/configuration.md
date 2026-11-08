# Configuration

The CLI uses three kinds of configuration:

1. **Main configuration** (`easifem.toml`) selects an environment and its root
   directories.
2. **Package plugins** (`plugins/<name>.toml`) describe each library: where its
   source, build and install directories are, and how to build it.
3. **`runner.toml`** describes how to build the programs in one examples or
   application directory.

Editing any of them does not rebuild a library by itself; run the relevant
command afterwards.

## Main configuration

Select a file explicitly with `--config`:

```sh
easifem --config /path/to/easifem.toml --env development test ./examples --list
```

Without `--config`, the CLI searches for `easifem.toml` in these directories
and uses the **first** one found:

1. `$HOME`
2. `$HOME/easifem`
3. `$HOME/.config`
4. `$HOME/.config/easifem`
5. The current directory

A file in the current directory is therefore used only when none of the home
locations has one. Use `--config` when the choice matters.

A minimal example:

```toml
envName = "development"

[development]
sourceDir = "$HOME/code/easifem"
buildDir = "$HOME/temp/easifem/build"
installDir = "$HOME/.local/easifem"
lintDir = "$HOME/temp/easifem/lint"
```

- The active environment is `--env` if given, else `envName`, else `env`.
- The directory keys are read from that environment's table. `$HOME` and other
  environment variables are expanded; a literal `~` is not.
- An environment table may also set `buildSystem`, `env` (a list of
  `KEY=value` strings), `ldLibraryPath` and `extpkgs`. These serve as defaults
  for packages that do not set them.
- A `plugins` directory must exist next to the main configuration file, and
  the directory must be writable (see [generated files](#generated-files)).

## Package plugins

The package `classes` is defined in `plugins/classes.toml` next to the main
configuration:

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

Adapt paths and options to your machine; the layout is only an example.
Define `base` and any external libraries the same way.

| Key                                   | Meaning and default                                                                                              |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `name`                                | Required for active packages. This is the name used on the command line.                                         |
| `isActive`                            | Default `true`. Inactive packages are skipped by every command and left out of the cache.                        |
| `isExtPkg`                            | Default `false`. `true` adds the package to the `extpkgs` group.                                                 |
| `sourceDir`, `buildDir`, `installDir` | If omitted: `<name>.<key>` in the main configuration, else `<environment dir>/<name>`.                           |
| `git` / `url`                         | Download source for `easifem install` (without `--no-download`).                                                 |
| `buildSystem`                         | `cmake` (default) or `make`. See [install](install.md#stages).                                                    |
| `buildType`                           | Default `Release`. Used by `install`; `easifem dev` always builds `Debug`.                                        |
| `buildSharedLibs`, `buildStaticLibs`  | Passed to CMake as `BUILD_SHARED_LIBS`/`BUILD_STATIC_LIBS`.                                                       |
| `buildOptions`                        | Extra arguments for CMake configure (or `./configure` for `make`).                                               |
| `targetName`                          | Default: `name`.                                                                                                 |
| `cmakePrefixPath`                     | Extra CMake package-search prefixes, added to the cache.                                                         |
| `envVars`                             | Environment variables set while building. If omitted, the `env` list from `[<name>]` or the environment is used. |
| `ldLibraryPath`                       | Extra library directories written to `easifemvar.fish`.                                                          |
| `dependencies`                        | Informational only: dependencies are **not** installed automatically.                                            |

Every entry in `plugins` is read as a package definition (the `.toml` suffix is
stripped from the name), so keep only plugin files in that directory. A plugin
that fails to parse stops every command.

Entries in an inherited `env` list must have the form `KEY=value`; the value
may itself contain `=`. A malformed entry is a configuration error.

## Generated files

On every start-up (every command except `--help` and `--version`), the CLI:

- loads the main configuration and all package plugins;
- creates the configured lint directory if it is missing;
- writes `cache.toml` next to the main configuration, listing the install
  prefixes (plus `cmakePrefixPath`) and library directories of all active
  packages;
- writes `easifemvar.fish` next to it, with one
  `set -gx LD_LIBRARY_PATH ...` line per library directory.

The cache lists configured prefixes; it does not prove that a library has been
installed there. `easifemvar.fish` is not loaded automatically: run
`source ~/.config/easifem/easifemvar.fish` (adjusting the path) in Fish, or set
`LD_LIBRARY_PATH` yourself in other shells. No POSIX shell file is written.

## `runner.toml`

Place a `runner.toml` in each examples or application directory. For the
classes examples:

```toml
BuildType = "Debug"
BuildDir = "/tmp/easifemClasses/FEDOF"
TargetLibs = ["easifemClasses"]
```

`TargetLibs` names installed CMake packages; each is found with
`find_package(<name>)` and linked as `<name>::<name>`. If the list is empty,
`easifemBase` is used. The install prefixes from the cache are added to the
CMake search path automatically.

Other accepted keys:

```toml
Compiler = "gfortran"                  # used by easifem test only
CompilerPath = "/opt/gcc/bin"          # used by easifem test only
TargetName = "main.out"
SourceDir = "."
ExtraFileNames = ["helper_module.F90"]
IncludePath = ["include"]              # used by easifem test only
ExtraIncludePath = ["include"]         # used by easifem test only
LibraryPath = ["lib"]                  # used by easifem test only
ExtraCMakePrefixPath = ["/path/to/another/install"]
ExtraLibs = ["m"]
Flags = ["-Wall"]                      # used by easifem test only
IsExecute = false                      # true means: build only, do not run
CacheClean = false
ReBuild = false
```

`IsExecute` is named the wrong way round: `IsExecute = true` *skips* running
the program, the same as `easifem run --no-run`.

The two commands treat `runner.toml` differently:

|                        | `easifem run`                                  | `easifem test`                                         |
| ---------------------- | ---------------------------------------------- | ------------------------------------------------------ |
| Default `BuildType`    | `Release`                                      | `Debug`                                                |
| Default `BuildDir`     | `./build`, used directly                       | system temp dir; parent of fresh per-test workspaces   |
| Generated CMake file   | **Overwrites** `./CMakeLists.txt`              | Written inside the workspace; sources untouched        |
| Main source            | The filename(s) on the command line            | Always the discovered test (`FileName` ignored)        |
| `IsExecute`/`CacheClean` | Honored; command-line flags override them    | Ignored                                                |
| `ReBuild`              | Honored; `--rebuild` overrides it              | Honored, or forced by `--rebuild`                       |
| Malformed TOML         | Error                                          | `COMPILE` result for that program                      |
| Missing file           | Defaults                                       | Defaults                                               |

See [run](run.md) and [test](test.md) for details.

## Troubleshooting

| Symptom                                              | What to check                                                                                     |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| Main configuration not found                         | The `--config` path, or the search directories above.                                             |
| Cannot read package metadata                         | The `plugins` directory next to the selected configuration; stray files in it; plugin TOML syntax.|
| `unknown package "x"`                                | A plugin with `name = "x"` exists in `plugins`. (An inactive package is skipped silently instead.) |
| CMake cannot find `easifemBase` or `easifemClasses`  | The package's `installDir` and whether it was actually installed there.                           |
| Missing `.mod` files or unresolved symbols           | The library was installed from matching sources; the example's `TargetLibs`.                      |
| Examples do not see your library changes             | Run `easifem install <pkg> --no-download`; `easifem dev` does not install.                        |
| Missing mesh or fixture file at runtime              | Relative paths are resolved from the example directory; check the data exists.                     |
| `NOCHECK`                                            | The program succeeded but printed no recognized checks; read its log.                              |
| CLI behaves differently from these docs              | `command -v easifem` points at an older build; reinstall the CLI.                                 |

Avoid `easifem clean` as a troubleshooting step: it also deletes the installed
library, which must then be rebuilt.
