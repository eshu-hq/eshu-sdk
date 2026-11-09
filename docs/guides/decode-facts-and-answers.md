# Decode facts and answers

Two different jobs, two different modules. Don't mix them up.

## Decode fact payloads with `factschema`

The reducer decodes each envelope into the typed struct for that fact
kind instead of picking through untyped maps. A missing required field
comes back as a classified decode error that names the field — never as
a silent zero value.

- Typed structs live under `<family>/v1` (for example `aws/v1`).
- JSON Schemas under `factschema/schema/` are generated from those
  structs and checked for drift.
- `factschema/identity` derives stable fact IDs.
- `fixturepack` and `contracttest` give you shared fixtures and assertions.

See the [factschema README](../../factschema/README.md) for the full surface.

## Decode API answers with `wirecontract`

A thin client never touches `collector` or `factschema`. It decodes API
answers with `wirecontract`: the response, truth, and error envelopes,
query profiles, freshness values, the MCP tool-definition shape, and
the client environment registry with its validator.

It's standard library only. Golden files in `wirecontract/testdata/golden`
pin every JSON field to core bytes, so drift shows up as a red test
rather than a quiet client bug.

`wirecontract` deliberately leaves out envelope constructors, the
capability matrix, freshness cause plumbing, renderers, and anything
server-side. The package comment in `wirecontract/doc.go` gives the
full list. See the [wirecontract README](../../wirecontract/README.md).
