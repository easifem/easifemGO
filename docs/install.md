# Command: install

```sh
easifem install pkgname... [--no-download]
```

Build and install one or more configured packages, in the order given.
Build type, paths and options come from each package's
[plugin](configuration.md#package-plugins).

```sh
easifem install base --no-download
easifem install classes --no-download
easifem install easifem            # base, then classes
easifem --quiet install extpkgs
```

## Stages

1. **Download** — fetches the plugin's `git`/URL source into its `sourceDir`
   (using go-getter, so Git and archive URLs work). Skipped with
   `--no-download`; use that flag when building sources you have edited
   locally.
2. **Configure, Build, Install** — depends on the plugin's `buildSystem`:
   - `cmake` (default): `cmake -G Ninja` with the plugin's build type,
     shared/static settings, installed prefixes of registered packages and
     `buildOptions`; then `cmake --build` and `cmake --install`.
   - `make`: `<sourceDir>/configure --prefix=<installDir> <buildOptions>`,
     `make` and `make install`, all run in `buildDir`.

   Other values (including `mason`) currently perform no build or install
   step.

The install directory is created if needed. Inactive packages are skipped
silently. The command stops at the first failing package.

## Package groups

| Name      | Installs                                                          |
| --------- | ----------------------------------------------------------------- |
| `easifem` | The configured `base` and `classes` packages, in that order.     |
| `extpkgs` | Every plugin with `isExtPkg = true`, in alphabetical order.       |

Plugins may list `Dependencies`, but they are not installed automatically.
Install dependencies yourself, in order (for example `base` before `classes`).

`easifem install` manages Fortran packages. It does not install this Go CLI;
see the [README](../README.md#installation) for that.

## Progress and logs

The full output of every stage is written to
`<buildDir>/easifem-install-log.txt`. See
[library progress and logs](library-progress.md).
