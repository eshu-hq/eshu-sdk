# Architecture

## Who imports what

A collector builds its result with `collector` and fills fact payloads
from `factschema`. Core reads the same two modules on its side of the
wire. A client never touches either one: it decodes API answers with
`wirecontract`, which core doesn't import at all.

The full picture, with a text version, is in
[Diagrams: Who imports what](../diagrams/README.md#who-imports-what).

## How a call moves

The host sends one request and reads one result per call, and the SDK
validators run before the host commits anything. If a result won't fit
one bounded stdout write, it splits into pages that the host resumes
until it has the lot. The reducer then decodes each envelope into the
typed struct for that fact kind.

See [Diagrams: How a call moves](../diagrams/README.md#how-a-call-moves).

## Boundaries

- **No core import.** Nothing here imports Eshu core at build time.
  That separation is the whole point of the repo, and
  [AGENTS.md](../../AGENTS.md) states it as a rule.
- **Three independent modules.** No root module, no workspace, no
  `replace` directives. Each module proves it stands alone with
  `GOWORK=off` (see [Contributing](../contributing/README.md)).
- **Core keeps the authority.** Admission, grants, graph writes,
  queues, and durable state stay in core. This repo holds shapes and
  validators only.
- **Consumers.** Collector repos, the core reducer, and thin API
  clients. Which exact version each consumer has accepted is recorded
  in the [compatibility policy](../compatibility.md), because it
  changes per candidate.

## Running things

There's nothing to deploy. The one runnable tool is the entrypoint
generator, `collector/cmd/collector-entrypoints-gen`, with
[its own README](../../collector/cmd/collector-entrypoints-gen/README.md).
