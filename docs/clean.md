# Command: clean

Remove the configured build **and install** directories for each package.
Source directories are not the intended cleanup target. The `extpkgs` cleanup
group uses the environment's `extpkgs` list when set. Its fallback selects
only active packages marked `isExtPkg = true`.

This command is destructive and has no documented dry-run mode. Read the
actual configured paths and obtain explicit permission for every cleanup.
Cleanup rejects empty paths and paths containing the current directory,
validates both paths before deletion, and returns deletion errors.
Do not use it to investigate an ordinary failed example.
