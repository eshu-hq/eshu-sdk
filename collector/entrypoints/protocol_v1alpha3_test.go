// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package entrypoints

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector"
)

// TestGenerateAdmitsPagedProtocol proves a provider manifest may list
// collector-sdk/v1alpha3 so a generated stdio command can serve paged
// results (eshu-hq/eshu#9385), and that the generated main carries the
// protocol verbatim.
func TestGenerateAdmitsPagedProtocol(t *testing.T) {
	m := validManifest()
	m.SupportedProtocols = []string{collector.ProtocolVersionV1Alpha1, collector.ProtocolVersionV1Alpha3}
	files, err := Generate(m)
	if err != nil {
		t.Fatalf("Generate(v1alpha3) error = %v, want nil", err)
	}
	if !bytes.Contains(files[0].Contents, []byte(strconv.Quote(collector.ProtocolVersionV1Alpha3))) {
		t.Fatalf("generated main omits %q", collector.ProtocolVersionV1Alpha3)
	}
}
