# Redaction marker package

Read the root and `collector/AGENTS.md` before editing this package. Preserve
the legacy `redact.v1` HMAC framing: four uint64 big-endian length-prefixed
fields in order (domain, normalized reason, normalized source, raw bytes).
Keep the package standard-library-only and do not add provider policy or
logging. Tests should use literal legacy vectors; do not derive expected
digests from the implementation under test. Run focused tests and `-race`.
