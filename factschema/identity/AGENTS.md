# Identity package agent rules

Read the repository root and `factschema/AGENTS.md` before editing. StableID
produces durable fact keys; preserve the pinned algorithm, including JSON
marshaling, UTC time precision, nil-map normalization, and panic behavior.
Do not add a core or collector dependency. Keep normalization private. New
behavior requires fixed digest vectors derived independently of this package
and a differential comparison with the pinned core source. Run focused tests,
race, and vet from `factschema/` before handoff. Do not treat this extraction
as a module release or a completed core cutover.
