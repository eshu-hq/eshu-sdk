// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import "errors"

// Request is one core-owned claim and the contract/configuration given to a collector.
// Config is opaque to the SDK. The command.Run decoder preserves numbers as
// json.Number; its handlers should convert explicitly rather than assert float64.
type Request struct {
	ProtocolVersion string         `json:"protocol_version"`
	Claim           Claim          `json:"claim"`
	Contract        Contract       `json:"contract"`
	Config          map[string]any `json:"config,omitempty"`
}

// Validate checks request structure and protocol agreement. It does not check
// the clock, host admission, grants, fact declarations, or provider config.
func (r Request) Validate() error {
	if r.ProtocolVersion != ProtocolVersionV1Alpha1 && r.ProtocolVersion != ProtocolVersionV1Alpha2 {
		return errors.New("request protocol is unsupported")
	}
	if r.Contract.ProtocolVersion != r.ProtocolVersion {
		return errors.New("request and contract protocols differ")
	}
	return validateClaim(r.Claim)
}
