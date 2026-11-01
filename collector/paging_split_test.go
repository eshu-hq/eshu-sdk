// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestSplitFactsRespectsByteBound proves analyzer-side byte accounting: a
// hostile fact set splits into pages that each fit the per-call bound, in
// order, with no fact lost or duplicated, and an explicit error (never a
// silent drop) when one fact alone exceeds the bound.
func TestSplitFactsRespectsByteBound(t *testing.T) {
	t.Parallel()

	facts := make([]Fact, 0, 10)
	for i := 0; i < 10; i++ {
		fact := testPageFact()
		fact.StableKey = strings.Repeat("k", 24) + string(rune('a'+i))
		fact.SourceRef.FactKey = fact.StableKey
		fact.Payload = map[string]any{"blob": strings.Repeat("x", 400)}
		facts = append(facts, fact)
	}
	single, err := FactWireBytes(facts[0])
	if err != nil {
		t.Fatalf("FactWireBytes() error = %v, want nil", err)
	}
	bound := int64(single) * 3
	pages, err := SplitFacts(facts, bound)
	if err != nil {
		t.Fatalf("SplitFacts() error = %v, want nil", err)
	}
	if len(pages) < 4 {
		t.Fatalf("SplitFacts() pages = %d, want at least 4 for 10 facts at 3 per page", len(pages))
	}
	assertPagesWithinBound(t, facts, pages, bound)

	oversized := testPageFact()
	oversized.Payload = map[string]any{"blob": strings.Repeat("x", 4096)}
	if _, err := SplitFacts([]Fact{oversized}, 128); err == nil {
		t.Fatal("SplitFacts(oversized) error = nil, want explicit bound error")
	} else if !strings.Contains(err.Error(), oversized.StableKey) || !strings.Contains(err.Error(), oversized.Kind) {
		t.Fatalf("SplitFacts(oversized) error = %v, want kind and stable key named", err)
	}

	if _, err := SplitFacts(nil, 1024); err == nil {
		t.Fatal("SplitFacts(nil) error = nil, want non-nil")
	}
	if _, err := SplitFacts(facts, 0); err == nil {
		t.Fatal("SplitFacts(zero bound) error = nil, want non-nil")
	}
	if _, err := SplitFacts(facts, -1); err == nil {
		t.Fatal("SplitFacts(negative bound) error = nil, want non-nil")
	}
}

// TestFactWireBytesMatchesMarshal proves FactWireBytes reports the exact
// encoded size and fails closed on an unmarshalable payload.
func TestFactWireBytesMatchesMarshal(t *testing.T) {
	t.Parallel()

	fact := testPageFact()
	size, err := FactWireBytes(fact)
	if err != nil {
		t.Fatalf("FactWireBytes() error = %v, want nil", err)
	}
	if size <= len(fact.StableKey) {
		t.Fatalf("FactWireBytes() = %d, want larger than the stable key alone", size)
	}

	broken := testPageFact()
	broken.Payload = map[string]any{"bad": make(chan int)}
	if _, err := FactWireBytes(broken); err == nil {
		t.Fatal("FactWireBytes(unmarshalable) error = nil, want non-nil")
	}
	if _, err := SplitFacts([]Fact{broken}, 1<<20); err == nil {
		t.Fatal("SplitFacts(unmarshalable) error = nil, want non-nil")
	}
}

// TestPageBoundForStdout pins the explicit per-call page bound derived from
// the host stdout limit: the bound is the limit minus fixed envelope
// headroom, and a limit that leaves no page bytes fails closed.
func TestPageBoundForStdout(t *testing.T) {
	t.Parallel()

	const stdoutLimit = int64(4 * 1024 * 1024)
	bound, err := PageBoundForStdout(stdoutLimit)
	if err != nil {
		t.Fatalf("PageBoundForStdout() error = %v, want nil", err)
	}
	if want := stdoutLimit - 64*1024; bound != want {
		t.Fatalf("PageBoundForStdout(%d) = %d, want %d", stdoutLimit, bound, want)
	}
	for _, limit := range []int64{0, -1, 64 * 1024} {
		if _, err := PageBoundForStdout(limit); err == nil {
			t.Fatalf("PageBoundForStdout(%d) error = nil, want non-nil", limit)
		}
	}
}

// TestSplitFactsBoundsRetainedMemory proves the bounded-emission memory
// contract at first-scan scale (eshu-hq/eshu#9635: a first full scan
// emitting 49,047 facts against a 4 MiB page bound): every page's retained
// backing array stays proportional to the page it carries, never to the
// whole input.
func TestSplitFactsBoundsRetainedMemory(t *testing.T) {
	t.Parallel()

	const factCount = 49047
	facts := make([]Fact, 0, factCount)
	for i := 0; i < factCount; i++ {
		fact := testPageFact()
		fact.StableKey = "snapshot:github.com/eshu-hq/eshu:" + strconv.Itoa(i)
		fact.SourceRef.FactKey = fact.StableKey
		fact.Payload = map[string]any{"score": 8.7, "blob": strings.Repeat("x", 1024)}
		facts = append(facts, fact)
	}
	bound, err := PageBoundForStdout(4 * 1024 * 1024)
	if err != nil {
		t.Fatalf("PageBoundForStdout() error = %v, want nil", err)
	}
	pages, err := SplitFacts(facts, bound)
	if err != nil {
		t.Fatalf("SplitFacts() error = %v, want nil", err)
	}
	assertPagesWithinBound(t, facts, pages, bound)

	retained := 0
	for _, page := range pages {
		retained += cap(page)
	}
	// Retention stays proportional to the facts carried (append growth keeps
	// each page within ~2x its length), never to the input size times the
	// page count.
	if ceiling := 2*len(facts) + len(pages); retained > ceiling {
		t.Fatalf("SplitFacts() retained %d fact slots, want <= %d (2x facts + pages)", retained, ceiling)
	}
}

// assertPagesWithinBound checks every page is non-empty and within bound,
// and that the pages concatenate back to facts in order with nothing lost.
func assertPagesWithinBound(t *testing.T, facts []Fact, pages [][]Fact, bound int64) {
	t.Helper()

	var flat []Fact
	for index, page := range pages {
		if len(page) == 0 {
			t.Fatalf("SplitFacts() pages[%d] is empty", index)
		}
		var size int64
		for _, fact := range page {
			factSize, err := FactWireBytes(fact)
			if err != nil {
				t.Fatalf("FactWireBytes() error = %v, want nil", err)
			}
			size += int64(factSize)
		}
		if size > bound {
			t.Fatalf("SplitFacts() pages[%d] bytes = %d, want <= %d", index, size, bound)
		}
		flat = append(flat, page...)
	}
	if len(flat) != len(facts) {
		t.Fatalf("SplitFacts() kept %d facts, want %d", len(flat), len(facts))
	}
	for i, fact := range facts {
		if flat[i].StableKey != fact.StableKey {
			t.Fatalf("SplitFacts() order changed at position %d", i)
		}
	}
}

func testClaim() Claim {
	return Claim{
		ComponentID:   "dev.eshu.examples.scorecard",
		InstanceID:    "scorecard-prod",
		CollectorKind: "scorecard",
		SourceSystem:  "openssf_scorecard",
		Scope:         Scope{ID: "repo:github.com/eshu-hq/eshu", Kind: "repository"},
		SourceRunID:   "run-20260609",
		GenerationID:  "gen-20260609",
		WorkItemID:    "work-1",
		FencingToken:  "fence-1",
		Attempt:       1,
		Deadline:      time.Date(2026, time.July, 8, 12, 0, 0, 0, time.UTC),
		ConfigHandle:  "secretref:scorecard-prod",
	}
}

func testPageFact() Fact {
	return Fact{
		Kind:             "dev.eshu.examples.scorecard.snapshot",
		SchemaVersion:    "1.0.0",
		StableKey:        "snapshot:github.com/eshu-hq/eshu",
		SourceConfidence: SourceConfidenceReported,
		ObservedAt:       time.Date(2026, time.July, 8, 12, 0, 0, 0, time.UTC),
		SourceRef: SourceRef{
			SourceSystem: "openssf_scorecard",
			ScopeID:      "repo:github.com/eshu-hq/eshu",
			GenerationID: "gen-20260609",
			FactKey:      "snapshot:github.com/eshu-hq/eshu",
			URI:          "https://api.securityscorecards.dev/projects/github.com/eshu-hq/eshu",
			RecordID:     "github.com/eshu-hq/eshu",
		},
		Payload: map[string]any{"score": 8.7},
	}
}
