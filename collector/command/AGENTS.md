# Command package instructions

Read the repository and collector `AGENTS.md` before changing this package.
Keep the runner standard-library-only beyond its parent collector module.
Require explicit protocol and byte budgets. Never include request, config, or
provider error text in returned command errors. Keep host admission and result
validation in the host. Run focused tests and `go test -race ./command`.
