// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package redact_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector/redact"
)

func TestLegacyMarkerVectors(t *testing.T) {
	key, err := redact.NewKey([]byte("proof-marker-key"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, raw, reason, source, digest, wantReason, wantSource string
	}{
		{"plain", "raw-secret", "sensitive", "azure.fixture", "2198721b3198e96733a687018fde5ed973e4ad5b65a2d428406c029186fd14ed", "sensitive", "azure.fixture"},
		{"unicode", "é雪", "sensitive", "azure.fixture", "0083406ec370d7fecd215fb5c0ee7167ee9c04674a52b60ae773bc92cfeb6422", "sensitive", "azure.fixture"},
		{"blank context", "value", " \t", "\n", "efcd0372c466332b135ceacd54ba2a079500ba5b98994146255333925685a3f1", "unknown", "unknown"},
		{"trim context", "value", " \twhy\n", " é ", "e5aee8e3d9f09c64b9295b658da1b67c1ed93be4f30ead54789667e5d2ad7539", "why", "é"},
		{"framing ab c", "raw", "ab", "c", "a051efc3106c76186d568107b198d35edb7b492bd629cd3b5110af92991714ee", "ab", "c"},
		{"framing a bc", "raw", "a", "bc", "89015896b39e84551dfa5ae21e7de59d266901b49322d166a3914286f1c66515", "a", "bc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := redact.Value{Marker: redact.MarkerPrefix + tc.digest, Reason: tc.wantReason, Source: tc.wantSource}
			if got := redact.String(tc.raw, tc.reason, tc.source, key); got != want {
				t.Fatalf("String() = %+v, want %+v", got, want)
			}
			if got := redact.Bytes([]byte(tc.raw), tc.reason, tc.source, key); got != want {
				t.Fatalf("Bytes() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestBinaryNilAndZeroKey(t *testing.T) {
	key, err := redact.NewKey([]byte("proof-marker-key"))
	if err != nil {
		t.Fatal(err)
	}
	wantBinary := redact.Value{Marker: redact.MarkerPrefix + "c66857e1b6bba57033fd9f09fecee0ca029b189c905f4b31e7daf5f7e54b9b7d", Reason: "sensitive", Source: "azure.fixture"}
	if got := redact.Bytes([]byte{0xff, 0xfe, 0}, "sensitive", "azure.fixture", key); got != wantBinary {
		t.Fatalf("binary = %+v, want %+v", got, wantBinary)
	}
	wantEmpty := redact.Value{Marker: redact.MarkerPrefix + "313878f0a67dabf15c7e427a2554197cb0cf088de3efe51eac7fa243e1657d31", Reason: "sensitive", Source: "azure.fixture"}
	for _, raw := range [][]byte{nil, {}} {
		if got := redact.Bytes(raw, "sensitive", "azure.fixture", key); got != wantEmpty {
			t.Fatalf("empty = %+v, want %+v", got, wantEmpty)
		}
	}
	var zero redact.Key
	if !zero.IsZero() {
		t.Fatal("zero Key must report zero")
	}
	wantZero := redact.Value{Marker: redact.MarkerPrefix + "7f449dbf6096322daa61b48cc97fba26722970a9e6d820f6a674150870467e3b", Reason: "sensitive", Source: "azure.fixture"}
	if got := redact.String("value", "sensitive", "azure.fixture", zero); got != wantZero {
		t.Fatalf("zero-key marker = %+v, want %+v", got, wantZero)
	}
	encoded, err := json.Marshal(wantZero)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"marker":"redacted:hmac-sha256:7f449dbf6096322daa61b48cc97fba26722970a9e6d820f6a674150870467e3b","reason":"sensitive","source":"azure.fixture"}` {
		t.Fatalf("JSON = %s", encoded)
	}
}

func TestKeyValidationAndCopy(t *testing.T) {
	for _, material := range [][]byte{nil, {}, []byte(" \t\n")} {
		if key, err := redact.NewKey(material); err == nil || !key.IsZero() {
			t.Fatalf("NewKey(%q) = (%v, %v), want zero and error", material, key.IsZero(), err)
		}
	}
	material := []byte(" key ")
	key, err := redact.NewKey(material)
	if err != nil || key.IsZero() {
		t.Fatalf("NewKey = (%v, %v)", key.IsZero(), err)
	}
	want := redact.Value{Marker: redact.MarkerPrefix + "55f5a00951ed473393689679c1de80975a908eaef2c0ca553884ec4a5ce03205", Reason: "sensitive", Source: "azure.fixture"}
	raw := []byte("value")
	before := bytes.Clone(raw)
	material[0] = 'x'
	if got := redact.Bytes(raw, "sensitive", "azure.fixture", key); got != want {
		t.Fatalf("copied key bytes marker = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(raw, before) {
		t.Fatalf("raw changed: %q", raw)
	}
}

func TestKeySharedAcrossCalls(t *testing.T) {
	key, err := redact.NewKey([]byte("proof-marker-key"))
	if err != nil {
		t.Fatal(err)
	}
	want := redact.String("raw-secret", "sensitive", "azure.fixture", key)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := redact.String("raw-secret", "sensitive", "azure.fixture", key); got != want {
				t.Errorf("shared key marker = %+v, want %+v", got, want)
			}
		}()
	}
	wg.Wait()
}
