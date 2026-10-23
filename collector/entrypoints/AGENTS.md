# Entrypoint package guidance

Read the repository and collector `AGENTS.md` before edits. Keep generation
deterministic and fail closed on unsafe manifests. Preserve provider ownership
of payloads and configuration. Use red-first tests for new validation guards.
Generated commands must emit no request, config, path, or handler detail to
stderr. Never add Eshu core implementation imports.
