# Collector transport helpers

`transport` provides reusable HTTP request, TLS trust, retry-status, and
bounded provider-failure helpers for collector authors. The package uses only
the Go standard library. It does not import Eshu core internals or the
collector wire-contract package.

`DoJSON` performs bounded request attempts and closes response bodies. Its
`JSONRequest` accepts a caller-owned client, headers, status handler, decoder,
and retry callback. `ParseBaseURL` rejects credential-bearing URLs;
`HTTPClientWithTLS` selects system trust, custom CA trust, or an explicit
local-test-only skip-verify mode. `HTTPError` and `ProviderFailure` keep error
strings free of provider response bodies and raw URLs. Callers must still
redact source-specific data and supply short, fixed `Operation` tokens rather
than request paths, IDs, or credentials.

The source package was Eshu `go/internal/collector/sdk` at the commit recorded
in [provenance](../../docs/provenance.md). Its `Version = "v0.1.0"` constant
is an inherited internal contract marker. It does not select a version for
this module or announce a release. Core collector call sites remain in Eshu.

Provider collectors own endpoint traversal, fact emission, workflow and retry
backoff decisions, and telemetry. This helper package emits no telemetry and
does not sleep between retries.

Run `GOWORK=off go test ./transport -count=1`,
`GOWORK=off go test -race ./transport -count=1`, and
`GOWORK=off go vet ./transport` from `collector/`.
