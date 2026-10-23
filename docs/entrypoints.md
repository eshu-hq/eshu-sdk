# Collector entrypoint generation

The collector SDK can generate a stdio command for a provider-owned handler.
The handler has the Go signature
`func(context.Context, collector.Request) (collector.Result, error)`.
The provider keeps its payload and configuration logic. The host still owns
admission, identity, grants, process supervision, and durable state.

Create a YAML manifest set:

```yaml
schema_version: 1
collectors:
  - schema_version: 1
    command_dir: cmd/my-collector
    handler_import_path: example.com/provider/handler
    handler_func: Handle
    supported_protocols: [collector-sdk/v1alpha1]
    max_request_bytes: 1048576
    max_result_bytes: 1048576
```

From the `collector` module, run:

```sh
GOWORK=off go run ./cmd/collector-entrypoints-gen --manifest ./entrypoints.yaml --repo-root . --write
GOWORK=off go run ./cmd/collector-entrypoints-gen --manifest ./entrypoints.yaml --repo-root . --check
```

Both flags naming files are required. Choose exactly one mode. The generator
accepts one strict YAML document and rejects unknown fields, duplicate output
directories, unsafe paths/imports, unsupported protocols, invalid handler
names, and nonpositive budgets. `max_request_bytes` cannot be the maximum
signed 64-bit integer because the runner uses a budget plus one to detect
oversize input. Output directories must be relative to `--repo-root`; existing
symlinks in the output path are refused.

The generated command accepts `--sdk-stdio` and reads one request from stdin.
It writes one result to stdout. Failure diagnostics on stderr identify only
the stage and do not include raw flags, requests, configuration, handler
errors, or paths. Synchronous panics on the command goroutine produce a fixed panic stage without
the panic value or stack. Handler and encoding failures occur before staged
output is written. Providers must keep their own writes and goroutines free of
sensitive diagnostics; the generated command cannot contain those. Writer
failures can leave partial stdout. The command passes explicit protocol and byte limits to
`command.Run`. Test configured limits against real provider inputs before
adoption; this example does not set a provider budget.

The generator adds no release or deployment claim. Existing Eshu host
templates and generated entrypoints stay in place until each collector's
consumer and deployment proof is accepted.
