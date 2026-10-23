# Collector transport agent rules

Read the repository root and `collector/AGENTS.md` before editing this package.
Keep it independent of Eshu internals and of provider-specific fact payloads,
endpoint traversal, and redaction. Preserve bounded retries, response-body
closure, credential-free errors, and fixed low-cardinality operation tokens.
Tests for transport behavior belong in this package. Run the focused test,
race, and vet commands in this directory's README before handoff. Do not
interpret the inherited `Version` constant as a module release version.
