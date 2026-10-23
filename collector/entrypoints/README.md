# Collector entrypoints

`Generate` renders one deterministic `main.go` for a provider handler. The
manifest names a local command directory, a provider package, an exported
handler function, explicit supported protocols, and positive request/result
byte budgets. The handler must have the signature
`func(context.Context, collector.Request) (collector.Result, error)`.

`LoadManifestFile` accepts one strict YAML document. Unknown fields, invalid
paths, duplicate targets, unsupported protocols, and unsafe imports fail
before generation. The generator does not import Eshu core internals or own
host admission, grant, graph, queue, or storage behavior.

See [entrypoint usage](../../docs/entrypoints.md) for the CLI contract.
