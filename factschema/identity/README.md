# Stable fact identity

`StableID(factType, identity)` returns lowercase SHA-256 hex over JSON of a
record containing `fact_type` and normalized `identity`. The algorithm was
copied from pinned Eshu core commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`.
The package uses only the Go standard library.

Normalization recursively copies `map[string]any`, `[]any`, and `[]string`;
converts `time.Time` to UTC RFC3339Nano; then lets `encoding/json` sort map
keys. A nil identity map normalizes to an empty object. Unsupported JSON values
panic with the inherited marshal error. Input maps and slices are not mutated.
The result is durable identity: changing these bytes changes fact keys.

This is an extraction candidate, not a selected module version or a core
cutover. See [identity migration](../../docs/identity-migration.md) for the
temporary source overlap and the proof needed before core imports it.

From `factschema/`, run `GOWORK=off go test ./identity -count=1`,
`GOWORK=off go test -race ./identity -count=1`, and
`GOWORK=off go vet ./identity`.
