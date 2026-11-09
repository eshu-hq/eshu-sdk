# Diagrams

Every diagram for eshu-sdk lives on this page. Other pages link here
instead of embedding their own copies. The repo [README](../../README.md)
keeps the first diagram only.

## Where the SDK sits

Collectors emit facts shaped by `collector` and `factschema`, the
reducer reads those same shapes, and API clients decode answers with
`wirecontract`. You only get types and validators here, no running code.

```mermaid
flowchart LR
    C["Collector programs<br/>(separate repos)"] -->|"facts, using<br/>collector + factschema"| R["Eshu core reducer"]
    R --> G["Graph and API"]
    G -->|"answers, decoded with<br/>wirecontract"| CLI["Thin command-line clients"]
    SDK["eshu-sdk (this repo):<br/>collector · factschema · wirecontract"] -.->|"shared types only"| C
    SDK -.-> R
    SDK -.-> CLI
```

Text version: Collector programs send facts (shaped by `collector` +
`factschema`) to the Eshu core reducer. The reducer feeds the graph
and API. Thin clients decode API answers with `wirecontract`. This
repo supplies shared types to all three sides and runs nothing itself.

## Who imports what

A collector builds its result with `collector` and fills fact payloads
from `factschema`. Core reads the same two modules on its side of the
wire. A client decodes API answers with `wirecontract`, which core
doesn't import at all.

```mermaid
flowchart TB
    subgraph coll["A collector repo"]
        CH["collector handler code"]
    end
    subgraph sdk["eshu-sdk"]
        C1["collector<br/>request/result, validators, paging"]
        C2["collector/command · transport · entrypoints · marker · conformance"]
        F1["factschema<br/>Envelope + typed payloads per family/v1"]
        F2["factschema/identity · fixturepack · contracttest"]
        W1["wirecontract<br/>API envelopes, profiles, tool shape, env registry"]
    end
    subgraph core["Eshu core"]
        HOST["extension host"]
        RED["reducer"]
        API["API / MCP surface"]
    end
    CLI["Thin client or CLI"]
    CH --> C1
    CH --> F1
    HOST --> C1
    RED --> F1
    RED --> F2
    API -.->|"shapes pinned by golden files"| W1
    CLI --> W1
    C2 -.-> CH
```

Text version: Collector handler code imports `collector` and
`factschema`. In core, the extension host imports `collector` and the
reducer imports `factschema` (including `identity`, `fixturepack`,
`contracttest`). The API/MCP surface does not import `wirecontract`;
its shapes are pinned to it by golden files. Thin clients import
`wirecontract`. The `collector` subpackages support the handler code.

## How a call moves

The host sends one request and reads one result per call. Validators
run before anything is committed, and big results come back in pages.

```mermaid
sequenceDiagram
    participant H as Core host
    participant Cmd as Collector command
    participant R as Reducer
    H->>Cmd: collector.Request as JSON on stdin
    Cmd->>Cmd: build facts with factschema payloads
    Cmd-->>H: collector.Result as JSON on stdout
    Note over Cmd,H: a big result comes back in pages,<br/>resumed with a cursor (v1alpha3)
    H->>H: validate before anything is committed
    H->>R: accepted fact envelopes
    R->>R: decode each kind into its typed struct
```

Text version: (1) Core host sends `collector.Request` JSON on stdin.
(2) The collector builds facts with `factschema` payloads. (3) It
returns `collector.Result` JSON on stdout, paged with a resume cursor
if large. (4) The host validates before committing. (5) Accepted
envelopes go to the reducer, which decodes each kind into its typed
struct.
