# Generator CLI guidance

Read root and collector `AGENTS.md`. Preserve exclusive `--write`/`--check`
behavior, output containment, and deterministic bytes. Tests must exercise
the real CLI `run` path, including stale and symlink outputs.
