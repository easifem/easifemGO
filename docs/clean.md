# Command: clean

```sh
easifem clean pkgname...
easifem clean extpkgs
```

> **Warning:** this permanently deletes each package's configured build
> directory **and its install directory**. There is no dry run and no
> confirmation prompt. Source directories are not touched.

Check `buildDir` and `installDir` in the package's
[plugin file](configuration.md#package-plugins) (or the environment fallback)
before cleaning.
After cleaning, the package must be rebuilt with
[`easifem install`](install.md) before examples can link against it.

Safety checks, applied to both paths before either is removed:

- An empty path is refused.
- A path equal to, or containing, the current directory is refused (symlinks
  are resolved first).
- A deletion error stops the command.

`extpkgs` uses the environment's `extpkgs` list when the configuration sets
one; otherwise it selects every active plugin with `isExtPkg = true`.

`clean` is rarely the right fix for a failing example: it removes the installed
library too. To get a fresh build of one example, use `easifem run --rebuild`
or `easifem test --rebuild` instead.
