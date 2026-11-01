// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"errors"
	"strings"
)

// Request is one core-owned claim and the contract/configuration given to a collector.
// Config is opaque to the SDK. The command.Run decoder preserves numbers as
// json.Number; its handlers should convert explicitly rather than assert float64.
type Request struct {
	ProtocolVersion string         `json:"protocol_version"`
	Claim           Claim          `json:"claim"`
	Contract        Contract       `json:"contract"`
	Config          map[string]any `json:"config,omitempty"`
	// Continuation carries the resume cursor on a collector-sdk/v1alpha3
	// continuation call. It is absent on the first call and on every call
	// to a collector that never paginates.
	Continuation *PageContinuation `json:"continuation,omitempty"`
}

// Validate checks request structure and protocol agreement. A continuation
// is admitted only on ProtocolVersionV1Alpha3 and must name a non-blank
// cursor. It does not check the clock, host admission, grants, fact
// declarations, provider config, or whether the cursor is one the collector
// minted; the collector owns its cursor format.
func (r Request) Validate() error {
	switch r.ProtocolVersion {
	case ProtocolVersionV1Alpha1, ProtocolVersionV1Alpha2, ProtocolVersionV1Alpha3:
	default:
		return errors.New("request protocol is unsupported")
	}
	if r.Contract.ProtocolVersion != r.ProtocolVersion {
		return errors.New("request and contract protocols differ")
	}
	if r.Continuation != nil {
		if r.ProtocolVersion != ProtocolVersionV1Alpha3 {
			return errors.New("request continuation requires collector-sdk/v1alpha3")
		}
		if strings.TrimSpace(r.Continuation.Cursor) == "" {
			return errors.New("request continuation cursor is required")
		}
	}
	return validateClaim(r.Claim)
}
