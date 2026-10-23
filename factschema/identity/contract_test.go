// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package identity

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestStableIDCollectorVectors pins identity maps copied from three collector
// call sites; expected digests were calculated from canonical JSON in Python.
func TestStableIDCollectorVectors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, factType string
		identity       map[string]any
		want           string
	}{
		{"confluence", "ConfluenceDocumentationFact", map[string]any{"fact_kind": "documentation_section", "stable_key": "section:deployment", "scope_id": "confluence:space:ops", "generation_id": "gen-17"}, "9fab5cfe8bdae768c3609767d8726633e7103b24f03c8b15b454ab668c7c96b3"},
		{"jira", "work_item.record", map[string]any{"provider": "jira_cloud", "scope_id": "jira:project:OPS", "issue_id": "12345", "issue_key": "OPS-17"}, "1a7b2bfd5af8b891868ef163463b243e67f95f85cc078db22ed323e373e1d78c"},
		{"pagerduty", "incident.lifecycle_event", map[string]any{"provider": "pagerduty", "scope_id": "pagerduty:account:example", "incident_id": "PINC123", "event_id": "PLOG456"}, "ddf1d62f8124a3d98697e673be63e18428f2cc6d234ba47e0be4b03830ca5e3d"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StableID(tt.factType, tt.identity); got != tt.want {
				t.Fatalf("StableID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStableIDEdgeVectors(t *testing.T) {
	t.Parallel()
	instant := time.Date(2026, 9, 22, 15, 4, 5, 123456789, time.UTC)
	for _, tt := range []struct {
		name     string
		identity map[string]any
		want     string
	}{
		{"nested map and slices", map[string]any{"nested": map[string]any{"a": []any{1, map[string]any{"z": "q"}}, "t": "x"}}, "ef1b53d47416fb3e759b5ad784729afd33c27e80edc26bb91e8c1e7cced3eb46"},
		{"nil map normalizes to empty", nil, "4c5aaab4476e6c1c0a0ba915bb8865186a912bb445b18515c410dd2bc688b434"},
		{"empty map", map[string]any{}, "4c5aaab4476e6c1c0a0ba915bb8865186a912bb445b18515c410dd2bc688b434"},
		{"nanosecond time", map[string]any{"at": instant}, "eff156be080461ddc14cac6c3b50d01bb9afee09031cb5bbd3df34cb31eae4b7"},
		{"JSON HTML escaping", map[string]any{"text": "a<b&c"}, "cdd2b04d6b2abdddd77e3da376f864df49b95de89f25011be81a37367478e1b9"},
		{"integer above float exact range", map[string]any{"n": int64(9007199254740993)}, "baea824fa55919e19b0f949bc3b6185949e2b26aec73e8dd79a14c09e0b28501"},
		{"maximum signed integer", map[string]any{"n": int64(9223372036854775807)}, "d697cdd1feac4c6c5a084ccb3d3c9646ee4794668911e15d10975e1bb2f663a5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StableID("edge", tt.identity); got != tt.want {
				t.Fatalf("StableID() = %q, want %q", got, tt.want)
			}
		})
	}
	shifted := instant.In(time.FixedZone("plus2", 2*60*60))
	if got := StableID("edge", map[string]any{"at": shifted}); got != "eff156be080461ddc14cac6c3b50d01bb9afee09031cb5bbd3df34cb31eae4b7" {
		t.Fatalf("zoned instant = %q", got)
	}
}

func TestStableIDPanicsOnUnsupportedJSONValue(t *testing.T) {
	t.Parallel()
	defer func() {
		value := recover()
		if value == nil || !strings.Contains(value.(string), "marshal stable id payload:") {
			t.Fatalf("panic = %v, want marshal stable id payload", value)
		}
	}()
	StableID("edge", map[string]any{"unsupported": make(chan int)})
}

func TestStableIDDoesNotMutateInput(t *testing.T) {
	t.Parallel()
	input := map[string]any{"nested": map[string]any{"slice": []any{map[string]any{"when": time.Date(2026, 9, 22, 15, 4, 5, 123456789, time.UTC)}}}}
	before := map[string]any{"nested": map[string]any{"slice": []any{map[string]any{"when": time.Date(2026, 9, 22, 15, 4, 5, 123456789, time.UTC)}}}}
	StableID("edge", input)
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("input mutated: %#v", input)
	}
}

func TestStableIDConcurrentCalls(t *testing.T) {
	t.Parallel()
	input := map[string]any{"nested": map[string]any{"slice": []any{"x", 17}}}
	want := StableID("edge", input)
	var workers sync.WaitGroup
	failures := make(chan string, 16)
	for range 16 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				if got := StableID("edge", input); got != want {
					failures <- got
					return
				}
			}
		}()
	}
	workers.Wait()
	close(failures)
	for got := range failures {
		t.Fatalf("concurrent StableID() = %q, want %q", got, want)
	}
}
