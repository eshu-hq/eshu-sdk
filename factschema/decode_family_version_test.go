// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package factschema

import (
	"errors"
	"testing"
)

// TestDecodeEachKind_UnsupportedMajorDeadLetters proves every public decoder
// rejects an unsupported schema major as a classified error.
func TestDecodeEachKind_UnsupportedMajorDeadLetters(t *testing.T) {
	t.Parallel()
	for _, factKind := range allDecodedKinds {
		t.Run(factKind, func(t *testing.T) {
			t.Parallel()
			env := Envelope{FactKind: factKind, SchemaVersion: "2.0.0", Payload: fullPayloadForKind(t, factKind)}
			err := decodeEnvelopeByKind(t, env)
			if !errors.Is(err, ErrUnsupportedSchemaMajor) {
				t.Fatalf("decode %s unsupported major: error = %v, want errors.Is ErrUnsupportedSchemaMajor", factKind, err)
			}
		})
	}
}
