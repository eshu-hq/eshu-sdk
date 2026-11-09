# Reference

## Which module does what

| Module | Reach for it when | Typical imports |
| --- | --- | --- |
| `github.com/eshu-hq/eshu-sdk/collector` | You're writing or hosting an out-of-tree collector and need the request/result records, validators, paging for large results, `transport` HTTP helpers, the `command.Run` stdio runner, or the `entrypoints` generator | `collector`, `collector/transport`, `collector/command` |
| `github.com/eshu-hq/eshu-sdk/factschema` | You produce or decode fact payloads and want a typed struct per fact kind, the shared `Envelope`, checked-in JSON Schemas, stable fact IDs from `identity`, shared fixtures from `fixturepack`, or assertions from `contracttest` | `factschema`, `factschema/aws/v1` and the other `<family>/v1` packages |
| `github.com/eshu-hq/eshu-sdk/wirecontract` | You're building a thin client that decodes Eshu API answers and needs the envelopes, query profiles, freshness values, the MCP tool-definition shape, or the client env registry | `wirecontract` |

Each module has its own README with the full surface:
[collector](../../collector/README.md),
[factschema](../../factschema/README.md), and
[wirecontract](../../wirecontract/README.md).

## Requirements and dependencies

| Module | Go | Third-party dependencies |
| --- | --- | --- |
| `collector` | 1.26.10+ | `gopkg.in/yaml.v3` |
| `factschema` | 1.26.10+ | `github.com/invopop/jsonschema` |
| `wirecontract` | 1.26.10+ | none (standard library only) |

Check each module's `go.mod` for the exact requirements. Nothing here
imports Eshu core at build time.

## Configuration

There is none. These libraries read no environment variables and take
no flags. The one runnable tool in the tree is the entrypoint
generator, `collector/cmd/collector-entrypoints-gen`, documented in
[its own README](../../collector/cmd/collector-entrypoints-gen/README.md).

## Versioning, as it actually works today

Tags don't move once they're published. A fix ships as the next
version, so if you pin a tag you keep getting exactly that code.

Watch out for the old paths. Code that still imports
`github.com/eshu-hq/eshu/sdk/go/...` is using the in-core modules, and
those types are not interchangeable with the ones here, even where the
names match. The new modules aren't drop-in replacements yet, and the
old paths stay put until consumer and deployed acceptance say otherwise.

Don't copy a "latest version" number out of any README, this one
included. Check the [compatibility policy](../compatibility.md) and the
exact tag instead. To cut a release yourself, follow the
[release checklist](../release.md).
