# Wire up a collector

A collector builds its result with `collector` and fills fact payloads
from `factschema`. Core reads the same two modules on its side of the
wire. The picture is in [Diagrams](../diagrams/README.md#how-a-call-moves).

## Steps

1. **Add the modules.** `go get` exact versions of `collector` and
   `factschema` (see [Getting started](../getting-started/README.md)).
2. **Handle one request, return one result.** The host sends one
   `collector.Request` as JSON and reads one `collector.Result` back.
   The SDK validators run before the host commits anything, and they
   fail closed.
3. **Fill payloads from `factschema`.** Use the shared `Envelope` plus
   the typed struct for each fact kind under `<family>/v1`
   (for example `aws/v1`, `iam/v1`, `incident/v1`).
4. **Let paging handle big results.** If a result won't fit one bounded
   stdout write, it splits into pages the host resumes with a cursor
   (v1alpha3). The helpers are `PageBoundForStdout`, `FactWireBytes`,
   and `SplitFacts`.
5. **Run it over stdio or generate the command.** Use `collector/command`
   for one bounded request over stdio, or `collector/entrypoints` to
   generate a collector command from a manifest.

## Read next

- [Command behaviour](../command.md) — the exact request/result swap with a core host.
- [Entrypoint generation](../entrypoints.md) — rendering a stdio command from a provider handler.
- [Conformance grants](../conformance-grants.md) — trusted core fact context for reserved fact kinds.
- [collector README](../../collector/README.md) — the full contract surface.
