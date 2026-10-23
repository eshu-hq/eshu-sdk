# Collector contract test helpers

This package gives collector tests a small neutral assertion surface over
`[]factschema.Envelope`. `CollectorContract` declares expected fact kinds and
required payload keys. `Lookup` finds one declaration. `AssertFactKinds`
rejects an emitted kind absent from the contract; `AssertRequiredPayloadKeys`
checks only declared kinds and only key presence. A present nil value counts
as present, and extra keys are allowed. `AssertFactShape` composes those two
checks. It does not check `CollectorKind`; call `ValidateCollectorKind`
separately when that identity must be asserted. Empty input passes. Use
`EnvelopeCounts` for diagnostic counts, including repeated kinds.

Each assertion takes `*testing.T` first and calls `t.Helper()`. These helpers
are opt-in test code, with no scanner, service-kind, client, or generated AWS
profile API. The candidate is copied from pinned Eshu core source at
`eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`, with envelope types adapted
to this module. See [migration](../../docs/contracttest-migration.md) for the
temporary overlap and later core adoption boundary. No module release is
selected by this extraction.

From `factschema/`, run `GOWORK=off go test ./contracttest -count=1`,
`GOWORK=off go test -race ./contracttest -count=1`, and
`GOWORK=off go vet ./contracttest`.
