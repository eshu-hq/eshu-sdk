# Contract test helper agent rules

Read the repository root and `factschema/AGENTS.md` before edits. Keep this
package test-only in purpose and dependent only on `factschema` plus the Go
standard library. Preserve subset fact-kind checks, required-key presence
semantics, and separate collector-kind validation. Assertions take `*testing.T`
first and call `t.Helper()`; prove failure behavior in an isolated test process
with real `*testing.T`. Do not add AWS scanner, boundary, service-kind, client,
or generated profile APIs here. Run focused test, race, and vet before handoff.
