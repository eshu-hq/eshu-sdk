# Decisions

Architecture decision records (ADRs) for eshu-sdk. An ADR captures one
decision, why we made it, and what it costs us — so future readers
don't have to re-derive it from code.

## Format

One file per decision: `ADR-NNNN-short-title.md`, numbered in order.
Each ADR uses these sections:

- **Status** — proposed, accepted, superseded (link the replacement).
- **Context** — the forces at play, in plain language.
- **Decision** — what we chose.
- **Consequences** — what gets easier, what gets harder.

Keep it short. Link to code and docs instead of restating them.

## Index

| ADR | Decision | Status |
| --- | --- | --- |
| [ADR-0001](ADR-0001-three-independent-modules.md) | Three independent Go modules; no root module, no Eshu core import | accepted |
