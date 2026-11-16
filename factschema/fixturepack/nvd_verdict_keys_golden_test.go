// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// nvdVerdictKeysGoldenSHA256 pins the file bytes. A deliberate vector change
// updates this digest in the same commit; an accidental one fails here first.
const nvdVerdictKeysGoldenSHA256 = "6a489baa68ec96eb01034536a78a8b3643c2ce31f8fe84cb73e2367830bdd073"

func TestNVDVerdictKeysGoldenFullContentDigest(t *testing.T) {
	got := fmt.Sprintf("%x", sha256.Sum256(NVDVerdictKeysGolden()))
	if got != nvdVerdictKeysGoldenSHA256 {
		t.Fatalf("NVD verdict keys golden digest = %s, want %s; review every key and case before updating the pin", got, nvdVerdictKeysGoldenSHA256)
	}
}

func TestNVDVerdictKeysGoldenContentAndCopy(t *testing.T) {
	raw := NVDVerdictKeysGolden()
	if err := validateNVDVerdictKeysGolden(raw); err != nil {
		t.Fatal(err)
	}
	raw[0] = 0
	if NVDVerdictKeysGolden()[0] == 0 {
		t.Fatal("accessor returned mutable embedded fixture bytes")
	}
}

// TestNVDVerdictKeysGoldenPinsNamesAndOutcomes holds the load-bearing names
// literally, so a rename needs an edit here as well as in the vector.
func TestNVDVerdictKeysGoldenPinsNamesAndOutcomes(t *testing.T) {
	golden := decodeVerdictGolden(t, NVDVerdictKeysGolden())
	want := map[string]map[string]string{
		"window": {
			"advance_to": "nvd_window_advance_to_unix_us", "head": "nvd_window_head_unix_us",
			"should_advance": "nvd_window_should_advance", "time_anomaly_reason": "nvd_window_time_anomaly_reason",
			"pages_refetched": "nvd_window_pages_refetched", "pages_downsized": "nvd_window_pages_downsized",
			"split": "nvd_window_split", "retry_reason": "nvd_window_retry_reason",
		},
		"baseline": {
			"run": "nvd_baseline_run", "year": "nvd_baseline_year", "advance_to": "nvd_baseline_advance_to_unix_us",
			"year_sha256": "nvd_baseline_year_sha256", "year_complete": "nvd_baseline_year_complete",
			"should_advance": "nvd_baseline_should_advance",
		},
	}
	for section, keys := range map[string][]verdictKey{"window": golden.Window.Keys, "baseline": golden.Baseline.Keys} {
		got := map[string]string{}
		for _, key := range keys {
			got[key.Field] = key.Name
			if key.Field == "time_anomaly_reason" && !slices.Equal(key.Enum, windowAnomalies) {
				t.Errorf("time_anomaly_reason enum = %v, want %v", key.Enum, windowAnomalies)
			}
		}
		if fmt.Sprint(got) != fmt.Sprint(want[section]) {
			t.Errorf("%s keys = %v, want %v", section, got, want[section])
		}
	}
	if !slices.Equal(golden.Baseline.Outcomes, baselineOutcomes) || !slices.Equal(golden.Window.Outcomes, windowOutcomes) {
		t.Fatalf("outcome sets drifted: window %v baseline %v", golden.Window.Outcomes, golden.Baseline.Outcomes)
	}
}

// TestNVDVerdictKeysGoldenReaderPinSet pins which cases the collector cannot
// emit, so flipping a flag to make a case look collector-backed fails here.
func TestNVDVerdictKeysGoldenReaderPinSet(t *testing.T) {
	golden := decodeVerdictGolden(t, NVDVerdictKeysGolden())
	wantPins := map[string][]string{
		"window": {"holds_should_advance_false", "holds_zero_advance_to", "holds_string_should_advance"},
		"baseline": {
			"advances_sub_millisecond_floors", "advances_sha_with_surrounding_whitespace", "not_covered_incomplete_at_start",
			"not_covered_sub_millisecond_floors_to_start", "not_covered_behind_start", "not_covered_complete_behind_start",
			"not_covered_missing_advance", "not_covered_zero_advance", "not_covered_string_advance", "not_covered_sha_blank",
			"not_covered_sha_malformed", "holds_string_should_advance", "holds_no_snapshot_keys",
			"refused_window_verdict_present", "refused_window_verdict_beside_baseline_hold",
		},
	}
	for section, cases := range map[string][]verdictCase{"window": golden.Window.Cases, "baseline": golden.Baseline.Cases} {
		var pins []string
		for _, c := range cases {
			if !*c.Emitted {
				pins = append(pins, c.Name)
			}
		}
		slices.Sort(pins)
		want := slices.Clone(wantPins[section])
		slices.Sort(want)
		if !slices.Equal(pins, want) {
			t.Errorf("%s reader pins = %v, want %v", section, pins, want)
		}
	}
}

// verdictSeed is one planted violation. A text seed replaces the first
// occurrence of old with new; a mutate seed edits the decoded document. Every
// seed names the validator error it must provoke, so a violation that passes,
// or that fails through a different rule, is caught.
type verdictSeed struct {
	name   string
	old    string
	new    string
	mutate func(doc map[string]any)
	want   string
}

// TestNVDVerdictKeysGoldenRejectsSeededCorruption plants one violation of each
// validator rule and requires that rule's error, so no rule can be weakened
// without a seed going red.
func TestNVDVerdictKeysGoldenRejectsSeededCorruption(t *testing.T) {
	original := string(NVDVerdictKeysGolden())
	for _, seed := range verdictSeeds() {
		t.Run(seed.name, func(t *testing.T) {
			t.Logf("seed %q want %q", seed.name, seed.want)
			var mutated string
			if seed.mutate != nil {
				mutated = mutateVerdictDoc(t, original, seed.mutate)
			} else {
				mutated = strings.Replace(original, seed.old, seed.new, 1)
			}
			if mutated == original {
				t.Fatal("seeded violation did not change the fixture")
			}
			err := validateNVDVerdictKeysGolden([]byte(mutated))
			if err == nil {
				t.Fatal("seeded violation passed golden validation")
			}
			if !strings.Contains(err.Error(), seed.want) {
				t.Fatalf("seeded violation failed through another rule: %v, want error containing %q", err, seed.want)
			}
		})
	}
}

func mutateVerdictDoc(t *testing.T, raw string, edit func(doc map[string]any)) string {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	edit(doc)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// caseEdit edits the named case of a section in the decoded document.
func caseEdit(section, name string, edit func(c, payload map[string]any)) func(map[string]any) {
	return func(doc map[string]any) {
		for _, raw := range doc[section].(map[string]any)["cases"].([]any) {
			c := raw.(map[string]any)
			if c["name"] == name {
				edit(c, c["payload"].(map[string]any))
				return
			}
		}
		panic("no case " + name)
	}
}

func num(v string) json.Number { return json.Number(v) }

func verdictSeeds() []verdictSeed {
	const windowScope = `vuln-intel://nvd/_window/1767225600000000-1767398400000000`
	const baselineScope = `_baseline/run-1/2002/1009843200000000`
	return []verdictSeed{
		{name: "unsupported version", old: `"version": 1`, new: `"version": 2`, want: "version or snapshot_fact_kind changed"},
		{name: "other snapshot fact kind", old: `"snapshot_fact_kind": "vulnerability.source_snapshot"`, new: `"snapshot_fact_kind": "vulnerability.cve"`, want: "version or snapshot_fact_kind changed"},
		{name: "blank comment", mutate: func(d map[string]any) { d["comment"] = " " }, want: "comment is blank"},
		{name: "window outcome set widened", old: `"outcomes": ["advanced", "held"]`, new: `"outcomes": ["advanced", "held", "moved"]`, want: "scope_prefix, key_prefix or outcomes changed"},
		{name: "window scope prefix changed", old: `"scope_prefix": "vuln-intel://nvd/_window/"`, new: `"scope_prefix": "vuln-intel://nvd/_windows/"`, want: "scope_prefix, key_prefix or outcomes changed"},
		{name: "window declares reasons", old: `"reasons": [],`, new: `"reasons": ["advance_missing"],`, want: "declares reasons or classes"},
		{name: "baseline reason set edited", old: `"reasons": ["advance_behind_start", "advance_missing", "advance_not_past_start", "sha_malformed"]`, new: `"reasons": ["advance_behind_start", "advance_missing", "advance_not_past_start", "sha_malformed", "extra"]`, want: "baseline reason set changed"},
		{name: "baseline class set edited", old: `"classes": ["nvd_baseline_window_writer_armed"]`, new: `"classes": ["nvd_baseline_window_writer_armed", "other"]`, want: "baseline class set changed"},
		{name: "window with no keys", mutate: func(d map[string]any) { d["window"].(map[string]any)["keys"] = []any{} }, want: "section declares no keys"},
		{name: "baseline with no cases", mutate: func(d map[string]any) { d["baseline"].(map[string]any)["cases"] = []any{} }, want: "section declares no cases"},

		{name: "renamed key in the table", old: `"name": "nvd_window_should_advance"`, new: `"name": "nvd_window_should_advanced"`, want: "is not declared"},
		{name: "key with no field name", old: `{"field": "advance_to", "name": "nvd_window_advance_to_unix_us"`, new: `{"field": "", "name": "nvd_window_advance_to_unix_us"`, want: "has no field name"},
		{name: "unprefixed key", old: `"name": "nvd_baseline_run"`, new: `"name": "baseline_run"`, want: "lacks the section key prefix"},
		{name: "duplicate key name that cases use", old: `"name": "nvd_baseline_year"`, new: `"name": "nvd_baseline_run"`, want: "is declared twice"},
		{name: "duplicate key name nothing uses", old: `"name": "nvd_window_retry_reason"`, new: `"name": "nvd_window_split"`, want: "is declared twice"},
		{name: "key type outside the set", old: `"type": "int64"`, new: `"type": "float64"`, want: "outside the closed type set"},
		{name: "key unit outside the set", old: `"unit": "unix_us"`, new: `"unit": "furlongs"`, want: "outside the closed unit set"},
		{name: "key reader outside the set", old: `"read_by": "core"`, new: `"read_by": "everyone"`, want: "outside the closed reader set"},
		{name: "key presence outside the set", old: `"presence": "always"`, new: `"presence": "sometimes"`, want: "outside the closed presence set"},
		{name: "should_advance declared a string", old: `"nvd_window_should_advance", "type": "bool"`, new: `"nvd_window_should_advance", "type": "string"`, want: "must be a core-read bool"},
		{name: "should_advance not read by core", old: `"type": "bool", "unit": "none", "read_by": "core", "presence": "always", "note": "Strict: only the JSON boolean true can move`, new: `"type": "bool", "unit": "none", "read_by": "none", "presence": "always", "note": "Strict: only the JSON boolean true can move`, want: "must be a core-read bool"},

		{name: "missing case name", old: `"name": "advances",`, new: ``, want: "a case with no name"},
		{name: "missing scope_id", old: `"scope_id": "` + windowScope + `",`, new: ``, want: "has no scope_id"},
		{name: "empty payload", mutate: caseEdit("window", "advances", func(c, _ map[string]any) { c["payload"] = map[string]any{} }), want: "has an empty payload"},
		{name: "missing expect", old: `"expect": {"outcome": "advanced", "advance_to": 1767398400000000, "head": 1767398400000000},`, new: ``, want: "has an empty expect"},
		{name: "misspelled expect member", old: `"expect": {`, new: `"expected": {`, want: "unknown field"},
		{name: "missing emitted_by_collector", old: `"emitted_by_collector": true,`, new: ``, want: "must declare emitted_by_collector"},
		{name: "duplicate case name", old: `"name": "advances_clamped_to_head"`, new: `"name": "advances"`, want: "is used twice"},
		{name: "scope outside the family prefix", old: `"scope_id": "vuln-intel://nvd/_window/`, new: `"scope_id": "vuln-intel://nvd/_other/`, want: "does not start with the section scope prefix"},
		{name: "window scope shape", old: `_window/1767225600000000-1767398400000000`, new: `_window/1767225600000000_1767398400000000`, want: "is not a window scope"},
		{name: "window scope degenerate bounds", old: `_window/1767225600000000-1767398400000000`, new: `_window/1767398400000000-1767225600000000`, want: "degenerate window bounds"},
		{name: "baseline scope shape", old: baselineScope, new: `_baseline/run-1/02/1009843200000000`, want: "is not a baseline scope"},

		{name: "malformed marker on an undeclared key", old: `"malformed_keys": ["nvd_window_should_advance"]`, new: `"malformed_keys": ["nvd_window_bogus"]`, want: "names undeclared key"},
		{name: "malformed marker on a key the payload lacks", old: `"malformed_keys": ["nvd_window_should_advance"]`, new: `"malformed_keys": ["nvd_window_retry_reason"]`, want: "which the payload does not carry"},
		{name: "stale malformed marker", old: `"malformed_keys": [`, new: `"malformed_keys": ["nvd_window_head_unix_us", `, want: "matches its declared type"},
		{name: "non-bool should_advance value", old: `"nvd_window_should_advance": true`, new: `"nvd_window_should_advance": "yes"`, want: "wrong-typed but not listed"},
		{name: "string where an int is expected", old: `"nvd_window_head_unix_us": 1767398400000000`, new: `"nvd_window_head_unix_us": "1767398400000000"`, want: "wrong-typed but not listed"},
		{name: "enum value outside the key's enum", old: `"nvd_window_time_anomaly_reason": "head_unknown"`, new: `"nvd_window_time_anomaly_reason": "head_unkown"`, want: "is outside its enum"},
		{name: "undeclared payload key", old: `"nvd_window_pages_refetched": 0`, new: `"nvd_window_pages_refetched_total": 0`, want: "is not declared"},

		{name: "emitted case with malformed keys", mutate: caseEdit("window", "holds_string_should_advance", func(c, _ map[string]any) { c["emitted_by_collector"] = true }), want: "must not carry malformed_keys"},
		{name: "emitted window case omits an always key", mutate: caseEdit("window", "advances", func(_, p map[string]any) { delete(p, "nvd_window_pages_downsized") }), want: "omits the always-present key"},
		{name: "emitted window advance beyond the head", mutate: caseEdit("window", "advances", func(_, p map[string]any) { p["nvd_window_head_unix_us"] = num("0") }), want: "without 0 < advance_to <= head"},
		{name: "emitted window hold with an advance", mutate: caseEdit("window", "holds_head_unknown", func(_, p map[string]any) { p["nvd_window_advance_to_unix_us"] = num("5") }), want: "non-zero advance_to with should_advance false"},
		{name: "emitted baseline case with a window key", mutate: caseEdit("baseline", "advances_past_start", func(_, p map[string]any) { p["nvd_window_should_advance"] = true }), want: "carries the window key"},
		{name: "emitted baseline advance sub-millisecond", mutate: caseEdit("baseline", "advances_past_start", func(_, p map[string]any) { p["nvd_baseline_advance_to_unix_us"] = num("1041379200000500") }), want: "not a whole millisecond at or after its scope start"},
		{name: "emitted baseline advance behind start", mutate: caseEdit("baseline", "advances_past_start", func(_, p map[string]any) { p["nvd_baseline_advance_to_unix_us"] = num("1000000000000000") }), want: "not a whole millisecond at or after its scope start"},
		{name: "emitted baseline should_advance inconsistent", mutate: caseEdit("baseline", "advances_past_start", func(_, p map[string]any) { p["nvd_baseline_should_advance"] = false }), want: "stamps should_advance="},
		{name: "emitted baseline sha malformed", mutate: caseEdit("baseline", "advances_past_start", func(_, p map[string]any) { p["nvd_baseline_year_sha256"] = "ABC" }), want: "not 64 lowercase hex"},

		{name: "window outcome outside the set", old: `"outcome": "advanced"`, new: `"outcome": "moved"`, want: "outside the closed outcome set"},
		{name: "baseline outcome outside the set", old: `"outcome": "not_covered"`, new: `"outcome": "uncovered"`, want: "outside the closed outcome set"},
		{name: "expect advance_to a string", old: `"expect": {"outcome": "advanced", "advance_to": 1767398400000000`, new: `"expect": {"outcome": "advanced", "advance_to": "1767398400000000"`, want: "is not a whole JSON number"},
		{name: "held window with a non-zero advance_to", old: `"expect": {"outcome": "held", "advance_to": 0, "head": 0}`, new: `"expect": {"outcome": "held", "advance_to": 5, "head": 0}`, want: "a held window arms no writer"},
		{name: "held window with a non-zero head", old: `"expect": {"outcome": "held", "advance_to": 0, "head": 0}`, new: `"expect": {"outcome": "held", "advance_to": 0, "head": 5}`, want: "a held window arms no writer"},
		{name: "advanced window expects another advance_to", old: `"expect": {"outcome": "advanced", "advance_to": 1767398400000000, "head": 1767398400000000}`, new: `"expect": {"outcome": "advanced", "advance_to": 1767398400000001, "head": 1767398400000000}`, want: "an advanced window expects advance_to"},
		{name: "advanced window expects another head", old: `"expect": {"outcome": "advanced", "advance_to": 1767398400000000, "head": 1767398400000000}`, new: `"expect": {"outcome": "advanced", "advance_to": 1767398400000000, "head": 1767398400000001}`, want: "an advanced window expects head"},
		{name: "baseline expect advance_to sub-millisecond", old: `"advance_to": 1041379200000000, "year_complete": false}`, new: `"advance_to": 1041379200000500, "year_complete": false}`, want: "must be a whole millisecond"},
		{name: "baseline expect advance_to not the floored stamp", old: `"advance_to": 1041379200000000, "year_complete": false}`, new: `"advance_to": 1041379201000000, "year_complete": false}`, want: "want the stamp floored"},
		{name: "baseline expect year_complete differs from the stamp", old: `"advance_to": 1041379200000000, "year_complete": false}`, new: `"advance_to": 1041379200000000, "year_complete": true}`, want: "expects year_complete to equal the stamp"},
		{name: "baseline expect year_complete not a bool", old: `"advance_to": 1041379200000000, "year_complete": false}`, new: `"advance_to": 1041379200000000, "year_complete": "false"}`, want: "must be a JSON bool"},
		{name: "not_covered with a blank reason", old: `"reason": "advance_behind_start"`, new: `"reason": ""`, want: "must expect a reason from the closed reason set"},
		{name: "not_covered with an unknown reason", old: `"reason": "advance_missing"`, new: `"reason": "no_advance"`, want: "must expect a reason from the closed reason set"},
		{name: "advanced with a reason", old: `"outcome": "advanced", "reason": ""`, new: `"outcome": "advanced", "reason": "advance_missing"`, want: "only not_covered carries one"},
		{name: "advanced with a class", old: `"outcome": "advanced", "reason": "", "class": ""`, new: `"outcome": "advanced", "reason": "", "class": "nvd_baseline_window_writer_armed"`, want: "only refused carries one"},
		{name: "refused with a blank class", old: `"class": "nvd_baseline_window_writer_armed", "advance_to"`, new: `"class": "", "advance_to"`, want: "must expect a class from the closed class set"},
		{name: "refused with an unknown class", old: `"class": "nvd_baseline_window_writer_armed", "advance_to"`, new: `"class": "other", "advance_to"`, want: "must expect a class from the closed class set"},
		{name: "held baseline with a non-zero advance_to", old: `"outcome": "held", "reason": "", "class": "", "advance_to": 0`, new: `"outcome": "held", "reason": "", "class": "", "advance_to": 1000`, want: "arms no writer, so expect advance_to and year_complete"},
		{name: "advanced baseline without an advance_to stamp", mutate: caseEdit("baseline", "advances_sub_millisecond_floors", func(_, p map[string]any) { delete(p, "nvd_baseline_advance_to_unix_us") }), want: "must stamp advance_to"},
	}
}
