// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import "testing"

// TestValidatePayloadKeysRejectsSensitiveLookingKeys locks the baseline
// behavior of the sensitive-key heuristic validatePayload/validatePayloadKeys
// enforces: a payload key whose name matches sensitiveQueryPattern
// (token/secret/password/credential/api_key/authorization, case-insensitive)
// is rejected as a payload-contract violation, so a collector can never emit
// an unredacted credential-shaped field.
func TestValidatePayloadKeysRejectsSensitiveLookingKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{name: "top-level secret", payload: map[string]any{"secret": "shhh"}},
		{name: "top-level password", payload: map[string]any{"password": "shhh"}},
		{name: "top-level api_key", payload: map[string]any{"api_key": "shhh"}},
		{name: "top-level authorization", payload: map[string]any{"authorization": "Bearer x"}},
		{name: "top-level credential", payload: map[string]any{"credential": "shhh"}},
		{name: "raw access_token", payload: map[string]any{"access_token": "shhh"}},
		{name: "nested under object", payload: map[string]any{"nested": map[string]any{"secret": "shhh"}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := validatePayload(tt.payload); err == nil {
				t.Fatalf("validatePayload(%+v) error = nil, want a sensitive-key rejection", tt.payload)
			}
		})
	}
}

// TestValidatePayloadKeysAllowsRedactionSafeJoinKeyFields proves the
// allowlist carve-out this test locks in: a field name that merely CONTAINS a
// sensitive-looking substring (e.g. "token") but is a redaction-safe join-key
// fingerprint the collector emits by design — never a raw credential — passes
// validatePayload instead of being rejected as if it were an unredacted
// secret. token_policy_join_keys (secretsiam.VaultAuthRole) is the motivating
// case: it is a slice of Vault ACL policy join-key fingerprints, not a Vault
// token. Before this allowlist existed, this payload was rejected with
// `sensitive-looking key "token_policy_join_keys" must be redacted before
// emission`, a false positive on the substring "token" that blocked a
// perfectly safe fingerprint field from ever being emitted by an in-contract
// collector.
func TestValidatePayloadKeysAllowsRedactionSafeJoinKeyFields(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"role_join_key":                   "role-fingerprint",
		"token_policy_join_keys":          []any{"sha256:policy-fingerprint"},
		"bound_service_account_join_keys": []any{"sha256:sa-fingerprint"},
	}
	if err := validatePayload(payload); err != nil {
		t.Fatalf("validatePayload(%+v) error = %v, want nil (token_policy_join_keys is an allowlisted redaction-safe join-key field)", payload, err)
	}
}

// TestValidatePayloadKeysAllowsEnvFromSecretBool proves the allowlist
// carve-out for the kubernetes_live.pod_template containers.env_from_secret
// field: a JSON boolean recording only WHETHER container env references
// secrets — never a value — passes validatePayload instead of being rejected
// as if it were an unredacted secret. Before this allowlist existed, the
// field was rejected with `sensitive-looking key
// "containers.env_from_secret" must be redacted before emission`, a false
// positive on the substring "secret" that terminally failed every
// extracted-collector claim (eshu-hq/eshu#9327, blocking #8859). The bool
// cannot carry a credential: the in-tree baseline carries it as false on all
// 1224 pod_template facts, and the satellite adapter sets it from
// env.ValueFrom.SecretKeyRef != nil without ever reading env.Value.
func TestValidatePayloadKeysAllowsEnvFromSecretBool(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"containers": []any{
			map[string]any{
				"name":            "api",
				"env_from_secret": false,
			},
		},
	}
	if err := validatePayload(payload); err != nil {
		t.Fatalf("validatePayload(%+v) error = %v, want nil (env_from_secret is an allowlisted redaction-safe bool)", payload, err)
	}
}

// TestValidatePayloadKeysAllowlistDoesNotWeakenGenericDetection proves the
// allowlist is scoped to exact field names, not a broadening of the regex
// itself: a genuinely sensitive-looking field that happens to share the
// "token" substring but is NOT one of the allowlisted join-key field names
// (for example a plain "token" or "auth_token" field) still fails closed.
func TestValidatePayloadKeysAllowlistDoesNotWeakenGenericDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{name: "plain token field", payload: map[string]any{"token": "shhh"}},
		{name: "auth_token field", payload: map[string]any{"auth_token": "shhh"}},
		{name: "vault_token field", payload: map[string]any{"vault_token": "shhh"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := validatePayload(tt.payload); err == nil {
				t.Fatalf("validatePayload(%+v) error = nil, want a sensitive-key rejection (allowlist must not broaden past its exact field names)", tt.payload)
			}
		})
	}
}
