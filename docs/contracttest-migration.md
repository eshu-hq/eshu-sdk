# Collector contract helper extraction boundary

This candidate copies the generic helpers from Eshu core
`go/internal/collector/conformance/contract/assert.go` at commit
`eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5` into
`factschema/contracttest`. The helper inputs become `[]factschema.Envelope`.
The four generic source tests retain their positive cases, with AWS fact-kind
constants replaced by their identical literal values. Additional tests use
real isolated Go test processes to prove that undeclared kinds, missing keys,
nil payloads, and wrong collector kinds fail as intended. Positive controls
cover extra keys, present nil values, unknown kinds skipped by the standalone
key check, empty input, and repeated counts.

The core implementation stays in place for now. Its AWS `ScanFunc`, boundary,
service-kind and client assertions, and generated `AWSContract` profile stay
outside this neutral package. A later core adoption can use the published SDK
module with its existing fact-envelope adapter. S3 can then call the generic
API directly, while generated AWS data can refer to SDK contract types. Core
should remove duplicate generic helpers after that compatibility proof;
permanent twin implementations would allow their assertion semantics to drift.

No module version or tag is selected. This extraction alone does not prove a
runtime consumer, S3 adoption, or production cutover.
