# Command: dev

```sh
easifem dev pkgname
```

Configure and build a configured package from its local sources, for
development. The package's source, build and install paths come from its
[plugin](configuration.md#package-plugins).

```sh
easifem dev classes
easifem --verbose dev classes
```

- Always uses CMake with Ninja, a `Debug` build and shared libraries,
  regardless of the plugin's `buildSystem`, `buildType` or library settings.
  The plugin's `buildOptions` are appended.
- Passes the installed prefixes of all registered packages as
  `CMAKE_PREFIX_PATH`, so dependencies such as `base` must already be
  installed.
- Does not download sources and does not install. Examples run with
  `easifem run`/`easifem test` keep using the previously installed library
  until you run [`easifem install`](install.md).
- Reads only the first package name; extra arguments are ignored.
- An inactive package (`isActive = false`) is skipped silently.

Progress is shown as a live dashboard, and the full output of every stage is
written to `<buildDir>/easifem-build-log.txt`. See
[library progress and logs](library-progress.md).
