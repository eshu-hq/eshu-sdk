// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

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
	baselineScopeShape = regexp.MustCompile(`^vuln-intel://nvd/_baseline/([A-Za-z0-9_-]{1,64})/([0-9]{4})/(0|[1-9][0-9]{0,18})(?:/([A-Z0-9-]{1,64}))?$`)
	yearSHAShape       = regexp.MustCompile(`^[0-9a-f]{64}$`)
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
	Emitted       *bool                      `json:"emitted_by_collector"`
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
	Version          int              `json:"version"`
	SnapshotFactKind string           `json:"snapshot_fact_kind"`
	Comment          string           `json:"comment"`
	Window           verdictSection   `json:"window"`
	Baseline         verdictSection   `json:"baseline"`
	Position         *verdictPosition `json:"position"`
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
	if golden.Version != 1 || golden.SnapshotFactKind != "vulnerability.source_snapshot" {
		return fmt.Errorf("version or snapshot_fact_kind changed: %d %q", golden.Version, golden.SnapshotFactKind)
	}
	if strings.TrimSpace(golden.Comment) == "" {
		return fmt.Errorf("comment is blank, so the source revisions are unrecorded")
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
	if len(golden.Window.Reasons) != 0 || len(golden.Window.Classes) != 0 {
		return fmt.Errorf("window section declares reasons or classes, which only baseline has")
	}
	if !slices.Equal(golden.Baseline.Reasons, baselineReasons) {
		return fmt.Errorf("baseline reason set changed: %v", golden.Baseline.Reasons)
	}
	if !slices.Equal(golden.Baseline.Classes, baselineClasses) {
		return fmt.Errorf("baseline class set changed: %v", golden.Baseline.Classes)
	}
	return validateVerdictPosition(golden.Position)
}

func validateVerdictSection(name string, section verdictSection, scopePrefix, keyPrefix string, outcomes []string, seen map[string]bool, borrowed []verdictKey) error {
	if section.ScopePrefix != scopePrefix || section.KeyPrefix != keyPrefix || !slices.Equal(section.Outcomes, outcomes) {
		return fmt.Errorf("%s section scope_prefix, key_prefix or outcomes changed", name)
	}
	declared := map[string]verdictKey{}
	for _, key := range section.Keys {
		if err := validateVerdictKey(name, key, keyPrefix, seen); err != nil {
			return err
		}
		seen[key.Name] = true
		declared[key.Name] = key
	}
	if len(declared) == 0 {
		return fmt.Errorf("%s section declares no keys", name)
	}
	for _, key := range borrowed {
		declared[key.Name] = key
	}
	if len(section.Cases) == 0 {
		return fmt.Errorf("%s section declares no cases", name)
	}
	names := map[string]bool{}
	for _, c := range section.Cases {
		switch {
		case c.Name == "":
			return fmt.Errorf("%s section has a case with no name", name)
		case names[c.Name]:
			return fmt.Errorf("%s case name %q is used twice", name, c.Name)
		case c.ScopeID == "":
			return fmt.Errorf("%s case %q has no scope_id", name, c.Name)
		case len(c.Payload) == 0:
			return fmt.Errorf("%s case %q has an empty payload", name, c.Name)
		case len(c.Expect) == 0:
			return fmt.Errorf("%s case %q has an empty expect", name, c.Name)
		case c.Emitted == nil:
			return fmt.Errorf("%s case %q must declare emitted_by_collector", name, c.Name)
		}
		names[c.Name] = true
		if err := validateVerdictCase(name, section, c, declared); err != nil {
			return fmt.Errorf("%s case %q: %w", name, c.Name, err)
		}
	}
	return nil
}

func validateVerdictKey(section string, key verdictKey, keyPrefix string, seen map[string]bool) error {
	switch {
	case key.Field == "":
		return fmt.Errorf("%s key %q has no field name", section, key.Name)
	case !strings.HasPrefix(key.Name, keyPrefix):
		return fmt.Errorf("%s key %q lacks the section key prefix %q", section, key.Name, keyPrefix)
	case seen[key.Name]:
		return fmt.Errorf("%s key name %q is declared twice", section, key.Name)
	case !slices.Contains(verdictKeyTypes, key.Type):
		return fmt.Errorf("%s key %q has type %q outside the closed type set", section, key.Name, key.Type)
	case !slices.Contains(verdictKeyUnits, key.Unit):
		return fmt.Errorf("%s key %q has unit %q outside the closed unit set", section, key.Name, key.Unit)
	case !slices.Contains(verdictKeyReaders, key.ReadBy):
		return fmt.Errorf("%s key %q has reader %q outside the closed reader set", section, key.Name, key.ReadBy)
	case !slices.Contains(verdictKeyPresences, key.Presence):
		return fmt.Errorf("%s key %q has presence %q outside the closed presence set", section, key.Name, key.Presence)
	case key.Field == "should_advance" && (key.Type != "bool" || key.ReadBy != "core"):
		return fmt.Errorf("%s should_advance must be a core-read bool", section)
	case key.Field == "advance_to_cve_id" && (key.Type != "string" || key.ReadBy != "core" || key.Presence != "conditional"):
		return fmt.Errorf("%s advance_to_cve_id must be a conditional core-read string", section)
	case key.Field == "year_generated" && (key.Type != "int64" || key.Unit != "unix_us" || key.ReadBy != "core" || key.Presence != "always"):
		return fmt.Errorf("%s year_generated must be an always-present core-read int64 unix_us", section)
	}
	return nil
}

func validateVerdictCase(section string, s verdictSection, c verdictCase, declared map[string]verdictKey) error {
	if !strings.HasPrefix(c.ScopeID, s.ScopePrefix) {
		return fmt.Errorf("scope_id %q does not start with the section scope prefix", c.ScopeID)
	}
	if section == "window" {
		m := windowScopeShape.FindStringSubmatch(c.ScopeID)
		if m == nil {
			return fmt.Errorf("scope_id %q is not a window scope", c.ScopeID)
		}
		start, _ := strconv.ParseInt(m[1], 10, 64)
		end, _ := strconv.ParseInt(m[2], 10, 64)
		if start <= 0 || end <= start {
			return fmt.Errorf("scope_id %q has degenerate window bounds", c.ScopeID)
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
	if err := validatePayloadTypes(c, declared); err != nil {
		return err
	}
	if *c.Emitted {
		if err := validateEmittedPayload(section, s, c); err != nil {
			return err
		}
	}
	if err := validateVerdictExpect(section, s, c); err != nil {
		return err
	}
	if section == "baseline" {
		return validateCasePosition(c)
	}
	return nil
}

func validatePayloadTypes(c verdictCase, declared map[string]verdictKey) error {
	for key, value := range c.Payload {
		spec, ok := declared[key]
		if !ok {
			return fmt.Errorf("payload key %q is not declared", key)
		}
		matches := jsonMatchesType(value, spec.Type)
		marked := slices.Contains(c.MalformedKeys, key)
		if !matches && !marked {
			return fmt.Errorf("payload key %q is wrong-typed but not listed in malformed_keys", key)
		}
		if matches && marked {
			return fmt.Errorf("payload key %q is listed in malformed_keys but matches its declared type", key)
		}
		if matches && len(spec.Enum) > 0 {
			var text string
			_ = json.Unmarshal(value, &text)
			if !slices.Contains(spec.Enum, text) {
				return fmt.Errorf("payload key %q value %q is outside its enum", key, text)
			}
		}
	}
	return nil
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

func stringMember(members map[string]json.RawMessage, key string) string {
	var text string
	_ = json.Unmarshal(members[key], &text)
	return text
}

func boolMember(members map[string]json.RawMessage, key string) bool {
	return string(members[key]) == "true"
}
