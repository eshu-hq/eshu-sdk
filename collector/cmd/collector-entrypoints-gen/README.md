# Entrypoint generator command

Run with `--manifest PATH --repo-root PATH` and exactly one of `--write` or
`--check`. The command validates all manifest entries and keeps generated
outputs under the given root. `--check` fails for absent or stale files.
Output paths containing existing symlinks are rejected.
