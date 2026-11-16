// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// nvdVerdictKeysGoldenSHA256 pins the file bytes. A deliberate vector change
// updates this digest in the same commit; an accidental one fails here first.
const nvdVerdictKeysGoldenSHA256 = "8c5ae1b9638e968174df3098db3a2d0ad1ad4b596c0f4b8676c55bde0c799248"

// Closed sets the validator holds the vector to. They are repeated here on
// purpose: a vector edit that widens a set must also edit this test.
var (
	verdictKeyTypes     = []string{"bool", "int", "int64", "string"}
	verdictKeyUnits     = []string{"count", "none", "unix_us", "year"}
	verdictKeyReaders   = []string{"core", "none"}
	verdictKeyPresences = []string{"always", "conditional"}

	windowOutcomes   = []string{"advanced", "held"}
	baselineOutcomes = []string{"advanced", "held", "not_covered", "refused"}
	baselineReasons  = []string{"advance_behind_start", "advance_missing", "advance_not_past_start", "sha_malformed"}
	baselineClasses  = []string{"nvd_baseline_window_writer_armed"}
	windowAnomalies  = []string{"", "head_unknown", "record_timestamp_unparsable", "record_timestamp_beyond_head", "head_behind_cursor"}
)

var (
	windowScopeShape   = regexp.MustCompile(`^vuln-intel://nvd/_window/([0-9]+)-([0-9]+)$`)
	baselineScopeShape = regexp.MustCompile(`^vuln-intel://nvd/_baseline/[A-Za-z0-9_-]{1,64}/([0-9]{4})/([0-9]+)$`)
)

type verdictKey struct {
	Field    string   `json:"field"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Unit     string   `json:"unit"`
	ReadBy   string   `json:"read_by"`
	Presence string   `json:"presence"`
	Enum     []string `json:"enum"`
	Note     string   `json:"note"`
}

type verdictCase struct {
	Name          string                     `json:"name"`
	ScopeID       string                     `json:"scope_id"`
	Payload       map[string]json.RawMessage `json:"payload"`
	Expect        map[string]json.RawMessage `json:"expect"`
	MalformedKeys []string                   `json:"malformed_keys"`
	Note          string                     `json:"note"`
}

type verdictSection struct {
	ScopePrefix string        `json:"scope_prefix"`
	KeyPrefix   string        `json:"key_prefix"`
	Outcomes    []string      `json:"outcomes"`
	Reasons     []string      `json:"reasons"`
	Classes     []string      `json:"classes"`
	Keys        []verdictKey  `json:"keys"`
	Cases       []verdictCase `json:"cases"`
}

type verdictGolden struct {
	Version          int            `json:"version"`
	SnapshotFactKind string         `json:"snapshot_fact_kind"`
	Comment          string         `json:"comment"`
	Window           verdictSection `json:"window"`
	Baseline         verdictSection `json:"baseline"`
}

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
	wantWindow := map[string]string{
		"advance_to":          "nvd_window_advance_to_unix_us",
		"head":                "nvd_window_head_unix_us",
		"should_advance":      "nvd_window_should_advance",
		"time_anomaly_reason": "nvd_window_time_anomaly_reason",
	}
	wantBaseline := map[string]string{
		"run":            "nvd_baseline_run",
		"year":           "nvd_baseline_year",
		"advance_to":     "nvd_baseline_advance_to_unix_us",
		"year_sha256":    "nvd_baseline_year_sha256",
		"year_complete":  "nvd_baseline_year_complete",
		"should_advance": "nvd_baseline_should_advance",
	}
	for section, want := range map[string]map[string]string{"window": wantWindow, "baseline": wantBaseline} {
		keys := golden.Window.Keys
		if section == "baseline" {
			keys = golden.Baseline.Keys
		}
		got := map[string]string{}
		for _, key := range keys {
			got[key.Field] = key.Name
		}
		for field, name := range want {
			if got[field] != name {
				t.Errorf("%s key %q = %q, want %q", section, field, got[field], name)
			}
		}
	}
	for section, keys := range map[string][]verdictKey{"window": golden.Window.Keys, "baseline": golden.Baseline.Keys} {
		for _, key := range keys {
			if key.Field == "should_advance" && (key.Type != "bool" || key.ReadBy != "core") {
				t.Errorf("%s should_advance must be a core-read bool, got %s read by %s", section, key.Type, key.ReadBy)
			}
		}
	}
	for _, key := range golden.Window.Keys {
		if key.Field == "time_anomaly_reason" && !slices.Equal(key.Enum, windowAnomalies) {
			t.Errorf("time_anomaly_reason enum = %v, want %v", key.Enum, windowAnomalies)
		}
	}
	if golden.SnapshotFactKind != "vulnerability.source_snapshot" || golden.Version != 1 {
		t.Fatalf("version %d / fact kind %q, want 1 / vulnerability.source_snapshot", golden.Version, golden.SnapshotFactKind)
	}
	if !slices.Equal(golden.Baseline.Outcomes, baselineOutcomes) || !slices.Equal(golden.Window.Outcomes, windowOutcomes) {
		t.Fatalf("outcome sets drifted: window %v baseline %v", golden.Window.Outcomes, golden.Baseline.Outcomes)
	}
}

// TestNVDVerdictKeysGoldenRejectsSeededCorruption plants one violation of each
// rule and requires the validator to reject it, so it cannot pass vacuously.
func TestNVDVerdictKeysGoldenRejectsSeededCorruption(t *testing.T) {
	original := string(NVDVerdictKeysGolden())
	cases := map[string][2]string{
		"renamed key in the table":      {`"name": "nvd_window_should_advance"`, `"name": "nvd_window_should_advanced"`},
		"unprefixed key":                {`"name": "nvd_baseline_run"`, `"name": "baseline_run"`},
		"duplicate key name":            {`"name": "nvd_baseline_year"`, `"name": "nvd_baseline_run"`},
		"key type outside the set":      {`"type": "int64"`, `"type": "float64"`},
		"should_advance declared text":  {`"nvd_window_should_advance", "type": "bool"`, `"nvd_window_should_advance", "type": "string"`},
		"non-bool should_advance value": {`"nvd_window_should_advance": true`, `"nvd_window_should_advance": "yes"`},
		"undeclared payload key":        {`"nvd_window_pages_refetched": 0`, `"nvd_window_pages_refetched_total": 0`},
		"window outcome outside set":    {`"outcome": "advanced"`, `"outcome": "moved"`},
		"baseline outcome outside set":  {`"outcome": "not_covered"`, `"outcome": "uncovered"`},
		"baseline reason outside set":   {`"reason": "advance_missing"`, `"reason": "no_advance"`},
		"sub-millisecond advance_to":    {`"advance_to": 1041379200000000`, `"advance_to": 1041379200000500`},
		"missing scope_id":              {`"scope_id": "vuln-intel://nvd/_window/1767225600000000-1767398400000000",`, ``},
		"missing expect block":          {`"expect": {`, `"expected": {`},
		"missing case name":             {`"name": "advances",`, ``},
		"stale malformed marker":        {`"malformed_keys": [`, `"malformed_keys": ["nvd_window_head_unix_us", `},
		"unsupported version":           {`"version": 1`, `"version": 2`},
	}
	for name, replacement := range cases {
		t.Run(name, func(t *testing.T) {
			mutated := strings.Replace(original, replacement[0], replacement[1], 1)
			if mutated == original {
				t.Fatal("seeded violation did not change the fixture")
			}
			err := validateNVDVerdictKeysGolden([]byte(mutated))
			if err == nil {
				t.Fatal("seeded violation passed golden validation")
			}
			t.Logf("rejected: %v", err)
		})
	}
}

func decodeVerdictGolden(t *testing.T, raw []byte) verdictGolden {
	t.Helper()
	var golden verdictGolden
	if err := strictDecode(raw, &golden); err != nil {
		t.Fatal(err)
	}
	return golden
}

// strictDecode rejects unknown members, so a misspelled case field cannot
// silently become an absent one.
func strictDecode(raw []byte, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("decode NVD verdict keys golden: %w", err)
	}
	return nil
}

func validateNVDVerdictKeysGolden(raw []byte) error {
	var golden verdictGolden
	if err := strictDecode(raw, &golden); err != nil {
		return err
	}
	if golden.Version != 1 || golden.SnapshotFactKind != "vulnerability.source_snapshot" || strings.TrimSpace(golden.Comment) == "" {
		return fmt.Errorf("NVD verdict keys golden header changed: version %d kind %q", golden.Version, golden.SnapshotFactKind)
	}
	seen := map[string]bool{}
	// A baseline payload may also carry window keys: that mixed shape is the
	// refused one, so a baseline case may name any declared window key.
	if err := validateVerdictSection("window", golden.Window, "vuln-intel://nvd/_window/", "nvd_window_", windowOutcomes, seen, nil); err != nil {
		return err
	}
	if err := validateVerdictSection("baseline", golden.Baseline, "vuln-intel://nvd/_baseline/", "nvd_baseline_", baselineOutcomes, seen, golden.Window.Keys); err != nil {
		return err
	}
	if !slices.Equal(golden.Baseline.Reasons, baselineReasons) || !slices.Equal(golden.Baseline.Classes, baselineClasses) {
		return fmt.Errorf("baseline reason or class set changed: %v %v", golden.Baseline.Reasons, golden.Baseline.Classes)
	}
	return nil
}

func validateVerdictSection(name string, section verdictSection, scopePrefix, keyPrefix string, outcomes []string, seen map[string]bool, borrowed []verdictKey) error {
	if section.ScopePrefix != scopePrefix || section.KeyPrefix != keyPrefix || !slices.Equal(section.Outcomes, outcomes) {
		return fmt.Errorf("%s section prefixes or outcomes changed", name)
	}
	declared := map[string]verdictKey{}
	for _, key := range section.Keys {
		switch {
		case !strings.HasPrefix(key.Name, keyPrefix) || key.Field == "":
			return fmt.Errorf("%s key %q lacks prefix %q or field name", name, key.Name, keyPrefix)
		case seen[key.Name]:
			return fmt.Errorf("%s key name %q is not unique", name, key.Name)
		case !slices.Contains(verdictKeyTypes, key.Type), !slices.Contains(verdictKeyUnits, key.Unit),
			!slices.Contains(verdictKeyReaders, key.ReadBy), !slices.Contains(verdictKeyPresences, key.Presence):
			return fmt.Errorf("%s key %q has a type, unit, reader or presence outside the closed sets", name, key.Name)
		case key.Field == "should_advance" && (key.Type != "bool" || key.ReadBy != "core"):
			return fmt.Errorf("%s should_advance must be a core-read bool", name)
		}
		seen[key.Name] = true
		declared[key.Name] = key
	}
	own := len(declared)
	for _, key := range borrowed {
		declared[key.Name] = key
	}
	if own == 0 || len(section.Cases) == 0 {
		return fmt.Errorf("%s section has no keys or no cases", name)
	}
	names := map[string]bool{}
	for _, c := range section.Cases {
		if c.Name == "" || c.ScopeID == "" || len(c.Payload) == 0 || len(c.Expect) == 0 {
			return fmt.Errorf("%s case %q is missing name, scope_id, payload or expect", name, c.Name)
		}
		if names[c.Name] {
			return fmt.Errorf("%s case name %q is not unique", name, c.Name)
		}
		names[c.Name] = true
		if err := validateVerdictCase(name, section, c, declared); err != nil {
			return fmt.Errorf("%s case %q: %w", name, c.Name, err)
		}
	}
	return nil
}

func validateVerdictCase(section string, s verdictSection, c verdictCase, declared map[string]verdictKey) error {
	if !strings.HasPrefix(c.ScopeID, s.ScopePrefix) {
		return fmt.Errorf("scope_id %q lacks prefix %q", c.ScopeID, s.ScopePrefix)
	}
	if section == "window" {
		m := windowScopeShape.FindStringSubmatch(c.ScopeID)
		if m == nil {
			return fmt.Errorf("scope_id %q is not a window scope", c.ScopeID)
		}
		start, _ := strconv.ParseInt(m[1], 10, 64)
		end, _ := strconv.ParseInt(m[2], 10, 64)
		if start <= 0 || end <= start {
			return fmt.Errorf("scope_id %q has degenerate bounds", c.ScopeID)
		}
	} else if baselineScopeShape.FindStringSubmatch(c.ScopeID) == nil {
		return fmt.Errorf("scope_id %q is not a baseline scope", c.ScopeID)
	}
	for _, malformed := range c.MalformedKeys {
		if _, ok := declared[malformed]; !ok {
			return fmt.Errorf("malformed_keys names undeclared key %q", malformed)
		}
		if _, ok := c.Payload[malformed]; !ok {
			return fmt.Errorf("malformed_keys names %q, which the payload does not carry", malformed)
		}
	}
	for key, value := range c.Payload {
		spec, ok := declared[key]
		if !ok {
			return fmt.Errorf("payload key %q is not declared", key)
		}
		matches := jsonMatchesType(value, spec.Type)
		if slices.Contains(c.MalformedKeys, key) == matches {
			return fmt.Errorf("payload key %q: type match %v disagrees with its malformed marker", key, matches)
		}
		if matches && len(spec.Enum) > 0 {
			var text string
			_ = json.Unmarshal(value, &text)
			if !slices.Contains(spec.Enum, text) {
				return fmt.Errorf("payload key %q value %q is outside its enum", key, text)
			}
		}
	}
	return validateVerdictExpect(section, s, c)
}

func jsonMatchesType(value json.RawMessage, typ string) bool {
	text := strings.TrimSpace(string(value))
	switch typ {
	case "bool":
		return text == "true" || text == "false"
	case "string":
		return strings.HasPrefix(text, `"`)
	default:
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil
	}
}

func intMember(members map[string]json.RawMessage, key string) (int64, error) {
	raw, ok := members[key]
	if !ok {
		return 0, fmt.Errorf("member %q is absent", key)
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("member %q is not a whole JSON number: %w", key, err)
	}
	return value, nil
}

func stringMember(members map[string]json.RawMessage, key string) (string, error) {
	var text string
	if err := json.Unmarshal(members[key], &text); err != nil {
		return "", fmt.Errorf("member %q is not a string", key)
	}
	return text, nil
}

func validateVerdictExpect(section string, s verdictSection, c verdictCase) error {
	outcome, err := stringMember(c.Expect, "outcome")
	if err != nil || !slices.Contains(s.Outcomes, outcome) {
		return fmt.Errorf("expect.outcome %q is outside the closed set", outcome)
	}
	advanceTo, err := intMember(c.Expect, "advance_to")
	if err != nil {
		return err
	}
	prefix := "nvd_" + section + "_"
	payloadAdvance, havePayloadAdvance := c.Payload[prefix+"advance_to_unix_us"]
	if section == "window" {
		head, err := intMember(c.Expect, "head")
		if err != nil {
			return err
		}
		if outcome == "held" && (advanceTo != 0 || head != 0) {
			return fmt.Errorf("a held window arms no writer, so expect advance_to and head must be 0")
		}
		if outcome == "advanced" {
			if got, _ := intMember(c.Payload, prefix+"advance_to_unix_us"); got != advanceTo {
				return fmt.Errorf("advanced window expects advance_to %d but stamps %d", advanceTo, got)
			}
			if got, _ := intMember(c.Payload, prefix+"head_unix_us"); got != head {
				return fmt.Errorf("advanced window expects head %d but stamps %d", head, got)
			}
		}
		return nil
	}
	// Baseline: expect.advance_to is the whole-millisecond value the armed
	// checkpoint writer would carry, 0 when nothing is armed.
	if advanceTo%1000 != 0 {
		return fmt.Errorf("expect.advance_to %d is not a whole millisecond", advanceTo)
	}
	completeRaw := string(c.Expect["year_complete"])
	if completeRaw != "true" && completeRaw != "false" {
		return fmt.Errorf("expect.year_complete must be a JSON bool")
	}
	complete := completeRaw == "true"
	reason, _ := stringMember(c.Expect, "reason")
	class, _ := stringMember(c.Expect, "class")
	if (outcome == "not_covered") != (reason != "") || (outcome == "not_covered" && !slices.Contains(s.Reasons, reason)) {
		return fmt.Errorf("expect.reason %q does not fit outcome %q", reason, outcome)
	}
	if (outcome == "refused") != (class != "") || (outcome == "refused" && !slices.Contains(s.Classes, class)) {
		return fmt.Errorf("expect.class %q does not fit outcome %q", class, outcome)
	}
	if outcome != "advanced" {
		if advanceTo != 0 || complete {
			return fmt.Errorf("outcome %q arms no writer, so expect advance_to and year_complete must be 0 and false", outcome)
		}
		return nil
	}
	if !havePayloadAdvance {
		return fmt.Errorf("an advanced baseline case must stamp advance_to")
	}
	stamped, err := strconv.ParseInt(strings.TrimSpace(string(payloadAdvance)), 10, 64)
	if err != nil || advanceTo != stamped-stamped%1000 {
		return fmt.Errorf("advanced baseline expects advance_to %d, want the stamp floored to a whole millisecond", advanceTo)
	}
	if string(c.Payload[prefix+"year_complete"]) != completeRaw {
		return fmt.Errorf("advanced baseline expects year_complete to equal the stamp")
	}
	return nil
}
