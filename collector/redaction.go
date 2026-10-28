// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

// IsSensitiveKeyName reports whether key matches the context-free sensitive
// name heuristic: a case-insensitive match against sensitiveQueryPattern
// (token|secret|password|credential|api[_-]?key|authorization) that is not
// carved out by the exact-name redactionSafePayloadKeys allowlist.
//
// This read-only predicate lets callers classify names before sharing a
// document. It uses the same regex and allowlist as validatePayloadKeys but
// has no fact kind, schema version, path, or value context. Fact validation
// can apply a narrower rule when all of that context is available.
func IsSensitiveKeyName(key string) bool {
	if _, safe := redactionSafePayloadKeys[key]; safe {
		return false
	}
	return sensitiveQueryPattern.MatchString(key)
}

// ValidateShareSafeKeys walks an arbitrary decoded JSON document (the result of
// json.Unmarshal into any) and rejects object keys that are sensitive-shaped
// per IsSensitiveKeyName at any depth.
//
// This context-free check does not require a map[string]any root or perform
// JSON-serializability checks. It calls validatePayloadKeys for an already
// decoded artifact. It cannot authorize the fact-specific Azure tag exception:
// a tagged Fact can admit an HMAC marker under an original sensitive-looking
// tag name, but an arbitrary shared document with that key is still rejected.
func ValidateShareSafeKeys(doc any) error {
	return validatePayloadKeys("", doc)
}

// redactionSafePayloadKeys allowlists exact payload field names that trip
// sensitiveQueryPattern's substring heuristic but are not, in fact, raw
// credentials: they are redaction-safe join-key fingerprints a collector
// emits by design to correlate facts (for example a hash of a Vault ACL
// policy name), never the secret value itself. The allowlist is keyed by
// EXACT field name, not a broader pattern, so it cannot silently widen the
// heuristic to cover an actually-sensitive field that happens to share a
// name; each entry is a deliberate, reviewed exception, not a generic escape
// hatch. Add an entry here only when the field's own type documents why it
// is a fingerprint/reference rather than a credential (see
// secretsiam/v1.VaultAuthRole.TokenPolicyJoinKeys for the motivating case).
var redactionSafePayloadKeys = map[string]struct{}{
	"token_policy_join_keys": {},
	// token_count is an integer body-size metric (lexer tokens), not a
	// credential; see reducerderived/v1.CodeDriftedMember.TokenCount.
	"token_count": {},
	// env_from_secret is a bool recording only WHETHER container env references
	// secrets (SecretKeyRef != nil; values never read), so "secret" is a
	// false positive (eshu-hq/eshu#9327). Value-bound to JSON bool/null;
	// see redactionSafeValue.
	"env_from_secret": {},
	// token_ttl_seconds is an integer Vault role TTL duration in seconds,
	// not a credential (eshu-hq/eshu#8869: vault_auth_role emission went
	// invalid_result on this key).
	"token_ttl_seconds": {},
	// token_policy_count is an integer count of attached Vault policies,
	// not a credential.
	"token_policy_count": {},
	// token_policy_name_fingerprints holds HMAC redaction markers over
	// Vault policy names, never the names themselves.
	"token_policy_name_fingerprints": {},
}

// redactionSafeValue reports whether value falls inside an allowlisted key's
// carve-out. Most allowlisted keys admit any value; env_from_secret admits
// only JSON booleans (or null) — any other type there could be a credential.
// Non-allowlisted keys are never safe.
func redactionSafeValue(key string, value any) bool {
	if _, safe := redactionSafePayloadKeys[key]; !safe {
		return false
	}
	if key != "env_from_secret" {
		return true
	}
	switch value.(type) {
	case bool, nil:
		return true
	default:
		return false
	}
}
