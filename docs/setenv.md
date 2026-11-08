# Command: setenv

```sh
easifem setenv
```

Currently prints the banner and help only. It does **not** write or change
any configuration. To set up directories persistently, edit the
[main configuration](configuration.md#main-configuration).

The command registers `--buildDir`/`-b`, `--sourceDir`/`-s`,
`--installDir`/`-i` and `--lintDir`/`-l`, but they have no lasting effect.
Older help examples that use `--build`, `--source` or `--install` do not match
these flag names.
