// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/eshu-hq/eshu-sdk/collector"
)

// pagedRequestBytes encodes one collector-sdk/v1alpha3 request, resumed
// with cursor when cursor is non-empty.
func pagedRequestBytes(t *testing.T, cursor string) []byte {
	t.Helper()
	r := collector.Request{
		ProtocolVersion: collector.ProtocolVersionV1Alpha3,
		Claim:           collector.Claim{ComponentID: "c", InstanceID: "i", CollectorKind: "k", SourceSystem: "s", Scope: collector.Scope{ID: "scope", Kind: "repo"}, SourceRunID: "run", GenerationID: "gen", WorkItemID: "work", FencingToken: "fence", Attempt: 1, Deadline: time.Now().Add(time.Hour), ConfigHandle: "config"},
		Contract:        collector.Contract{ProtocolVersion: collector.ProtocolVersionV1Alpha3},
	}
	if cursor != "" {
		r.Continuation = &collector.PageContinuation{Cursor: cursor}
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestRunPagedProtocolRoundTrip proves an explicitly admitted
// collector-sdk/v1alpha3 command receives the host's resume cursor and
// emits its page envelope unchanged, across a first call and a resumed call.
func TestRunPagedProtocolRoundTrip(t *testing.T) {
	handler := func(_ context.Context, req collector.Request) (collector.Result, error) {
		page := &collector.ResultPage{Cursor: "page-cursor-1", More: true, TotalFacts: 2}
		if req.Continuation != nil {
			if req.Continuation.Cursor != "page-cursor-1" {
				t.Errorf("cursor = %q, want page-cursor-1", req.Continuation.Cursor)
			}
			page = &collector.ResultPage{PageIndex: 1, TotalFacts: 2}
		}
		return collector.Result{ProtocolVersion: req.ProtocolVersion, State: collector.ResultComplete, Page: page}, nil
	}
	for _, tc := range []struct {
		cursor string
		want   collector.ResultPage
	}{
		{"", collector.ResultPage{Cursor: "page-cursor-1", More: true, TotalFacts: 2}},
		{"page-cursor-1", collector.ResultPage{PageIndex: 1, TotalFacts: 2}},
	} {
		raw := pagedRequestBytes(t, tc.cursor)
		opts := Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1, collector.ProtocolVersionV1Alpha3}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}
		var out bytes.Buffer
		if err := Run(context.Background(), bytes.NewReader(raw), &out, handler, opts); err != nil {
			t.Fatalf("Run(cursor=%q): %v", tc.cursor, err)
		}
		var result collector.Result
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.ProtocolVersion != collector.ProtocolVersionV1Alpha3 || result.Page == nil || *result.Page != tc.want {
			t.Fatalf("cursor=%q result=%+v page=%+v, want %+v", tc.cursor, result, result.Page, tc.want)
		}
	}
}

// TestRunRejectsUnadmittedOrMalformedContinuation proves the command
// boundary fails closed before the handler on a v1alpha3 request the
// collector did not admit, a blank resume cursor, or a continuation on an
// older protocol.
func TestRunRejectsUnadmittedOrMalformedContinuation(t *testing.T) {
	resumed := pagedRequestBytes(t, "page-cursor-1")
	blank := bytes.Replace(resumed, []byte(`"cursor":"page-cursor-1"`), []byte(`"cursor":""`), 1)
	older := bytes.ReplaceAll(resumed, []byte(collector.ProtocolVersionV1Alpha3), []byte(collector.ProtocolVersionV1Alpha1))
	both := []string{collector.ProtocolVersionV1Alpha1, collector.ProtocolVersionV1Alpha3}
	cases := map[string]struct {
		input     []byte
		protocols []string
	}{
		"v1alpha3 not admitted":  {resumed, []string{collector.ProtocolVersionV1Alpha1}},
		"blank cursor":           {blank, both},
		"continuation on alpha1": {older, both},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			called := false
			var out bytes.Buffer
			opts := Options{SupportedProtocols: tc.protocols, MaxRequestBytes: int64(len(tc.input)), MaxResultBytes: 4096}
			err := Run(context.Background(), bytes.NewReader(tc.input), &out, func(context.Context, collector.Request) (collector.Result, error) {
				called = true
				return collector.Result{}, nil
			}, opts)
			if err == nil || called || out.Len() != 0 {
				t.Fatalf("err=%v called=%v output=%q", err, called, out.String())
			}
		})
	}
}
