# Getting started with eshu-sdk

There's no root Go module here. Each module builds, tests, and versions
on its own, with its own `go.mod`, changelog, and tags that look like
`<module>/vX.Y.Z`. All three need Go 1.26.10 or newer (check the `go`
line in each `go.mod`).

## 1. Pick an exact version

Start with the [compatibility policy](../compatibility.md). It lists
checked tags and prepared next candidates per module. Use an exact
version and never `@latest` as release evidence, because a downloadable
candidate isn't the same thing as an accepted one.

## 2. Add the module you need

```sh
go get github.com/eshu-hq/eshu-sdk/collector@<exact-version>
go get github.com/eshu-hq/eshu-sdk/factschema@<exact-version>
go get github.com/eshu-hq/eshu-sdk/wirecontract@<exact-version>
```

You rarely need all three. Collectors use `collector` + `factschema`;
thin API clients use only `wirecontract`. See
[Reference](../reference/README.md) for the full module table.

## 3. Make a first import

```go
import (
    "github.com/eshu-hq/eshu-sdk/collector"
    "github.com/eshu-hq/eshu-sdk/collector/command"
    "github.com/eshu-hq/eshu-sdk/factschema"
    awsv1 "github.com/eshu-hq/eshu-sdk/factschema/aws/v1"
    "github.com/eshu-hq/eshu-sdk/wirecontract"
)
```

Build it. If that compiles, you're wired up.

## 4. Check the version before you rely on it

Before you depend on a version, recheck its tag and where it came from
through the public Go proxy, then build a fresh consumer with
`GOWORK=off` and no `replace` directive. That proves the published
module works on its own instead of quietly leaning on a local checkout.

The [provenance note](../provenance.md) records the Eshu commit the code
came from (`eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`), and the
[evidence notes](../evidence) hold the recorded consumer proofs for
specific candidates.

## Where to go next

- Building a collector? [Wire up a collector](../guides/wire-up-a-collector.md).
- Decoding facts or API answers? [Decode facts and answers](../guides/decode-facts-and-answers.md).
- Want the mental model first? [Concepts](../concepts/README.md).
