// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// The position block pins the NVD baseline resume position: the pair
// (start_unix_us, from_cve_id) that a baseline scope id names and a baseline
// verdict advances. These literals are repeated here on purpose, so a vector
// edit that changes the grammar must also edit this test.
const (
	positionScopeGrammar = "vuln-intel://nvd/_baseline/<run>/<year>/<start_unix_us>[/<from_cve_id>]"
	positionRunPattern   = "^[A-Za-z0-9_-]{1,64}$"
	positionYearPattern  = "^[0-9]{4}$"
	positionStartPattern = "^(0|[1-9][0-9]{0,18})$"
	positionIDPattern    = "^[A-Z0-9-]{1,64}$"
	positionCanonical    = "trim Unicode white space (strings.TrimSpace semantics), reject any non-ASCII byte, upper-case ASCII letters, then require cve_id_pattern"
	positionOrder        = "start_unix_us ascending as an integer, then from_cve_id by unsigned byte order; the empty id sorts before every id"
)

var (
	positionRoundTripNames = []string{"four_segment", "four_segment_longest_id", "three_segment_empty_id"}
	positionRejectionNames = []string{
		"empty_trailing_segment", "extra_segment", "id_in_start_slot", "leading_zero_start", "lowercase_id", "non_ascii_id",
		"non_ascii_id_upper_cases_to_ascii", "non_numeric_start", "overflow_start", "overlong_id", "padded_id_leading",
		"padded_id_nbsp", "padded_id_newline", "padded_id_trailing",
	}
	positionCanonicalNames = []string{
		"blank", "canonical_unchanged", "carriage_return_trimmed", "inner_space", "longest_64", "lowercase_upper_cased",
		"nbsp_trimmed", "newline_trimmed", "non_ascii_accent", "non_ascii_upper_cases_to_ascii", "overlong_65",
		"tab_and_space_trimmed", "underscore",
	}
	positionOrderNames = []string{
		"empty_id_before_any_id", "equal_empty_position", "equal_position", "hyphen_before_digit", "id_nine_thousand_after_ten_thousand",
		"id_ten_thousand_before_nine_thousand", "later_stamp_beats_id", "letter_after_digit", "shorter_prefix_first",
	}

	// baselineExpectMembers is the closed set of expect members a baseline case
	// may carry; advance_to_cve_id and year_generated_unix_us are optional, and
	// absent reads as "" and 0.
	baselineExpectMembers = []string{"outcome", "reason", "class", "advance_to", "year_complete", "advance_to_cve_id", "year_generated_unix_us"}
)

type positionFields struct {
	StartUnixUS int64  `json:"start_unix_us"`
	FromCVEID   string `json:"from_cve_id"`
}

type positionCanonicalID struct {
	Name      string `json:"name"`
	Raw       string `json:"raw"`
	Canonical string `json:"canonical"`
	Valid     bool   `json:"valid"`
	Note      string `json:"note"`
}

type positionRoundTrip struct {
	Name        string `json:"name"`
	ScopeID     string `json:"scope_id"`
	Run         string `json:"run"`
	Year        int    `json:"year"`
	StartUnixUS int64  `json:"start_unix_us"`
	FromCVEID   string `json:"from_cve_id"`
	Note        string `json:"note"`
}

type positionRejection struct {
	Name    string `json:"name"`
	ScopeID string `json:"scope_id"`
	Note    string `json:"note"`
}

type positionOrderCase struct {
	Name string         `json:"name"`
	A    positionFields `json:"a"`
	B    positionFields `json:"b"`
	Cmp  int            `json:"cmp"`
	Note string         `json:"note"`
}

type verdictPosition struct {
	Comment        string                `json:"comment"`
	ScopeGrammar   string                `json:"scope_grammar"`
	RunPattern     string                `json:"run_pattern"`
	YearPattern    string                `json:"year_pattern"`
	StartPattern   string                `json:"start_unix_us_pattern"`
	CVEIDPattern   string                `json:"cve_id_pattern"`
	CVEIDCanonical string                `json:"cve_id_canonical"`
	Order          string                `json:"order"`
	CanonicalIDs   []positionCanonicalID `json:"canonical_ids"`
	RoundTrips     []positionRoundTrip   `json:"round_trips"`
	Rejections     []positionRejection   `json:"rejections"`
	Orders         []positionOrderCase   `json:"orders"`
}

// scopePosition is a parsed baseline scope id.
type scopePosition struct {
	Run         string
	Year        int
	StartUnixUS int64
	FromCVEID   string
}

var positionIDShape = regexp.MustCompile(positionIDPattern)

// canonicalCVEID is the reference canonicalization, the collector's
// strings.ToUpper(strings.TrimSpace(id)) convention narrowed to ASCII. Non-ASCII
// bytes (after the trim) are rejected before upper-casing because unicode upper-casing maps some of them
// into ASCII (for example U+017F to "S"), which would admit ids the grammar
// excludes.
func canonicalCVEID(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] >= 0x80 {
			return "", false
		}
	}
	upper := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		}
		return r
	}, trimmed)
	if !positionIDShape.MatchString(upper) {
		return "", false
	}
	return upper, true
}

// parseBaselineScope is the reference scope parser. A scope id carries the
// canonical id verbatim: an id that canonicalization would change is not a
// scope id, and an empty id has no trailing segment.
func parseBaselineScope(scope string) (scopePosition, error) {
	m := baselineScopeShape.FindStringSubmatch(scope)
	if m == nil {
		return scopePosition{}, fmt.Errorf("%q is not a baseline scope id", scope)
	}
	year, _ := strconv.Atoi(m[2])
	start, err := strconv.ParseInt(m[3], 10, 64)
	if err != nil {
		return scopePosition{}, fmt.Errorf("%q has a start that is not an int64: %w", scope, err)
	}
	if m[4] != "" {
		if canonical, ok := canonicalCVEID(m[4]); !ok || canonical != m[4] {
			return scopePosition{}, fmt.Errorf("%q carries a non-canonical id", scope)
		}
	}
	return scopePosition{Run: m[1], Year: year, StartUnixUS: start, FromCVEID: m[4]}, nil
}

// formatBaselineScope is the reference scope builder: one position, one scope
// id, so an empty id is never written as a trailing segment.
func formatBaselineScope(p scopePosition) string {
	scope := fmt.Sprintf("vuln-intel://nvd/_baseline/%s/%04d/%d", p.Run, p.Year, p.StartUnixUS)
	if p.FromCVEID != "" {
		scope += "/" + p.FromCVEID
	}
	return scope
}

// comparePosition orders (stamp, id) pairs: the stamp as an integer, then the
// id by unsigned byte order, so "" sorts first and CVE-2025-10000 sorts before
// CVE-2025-9999.
func comparePosition(aStamp int64, aID string, bStamp int64, bID string) int {
	switch {
	case aStamp < bStamp:
		return -1
	case aStamp > bStamp:
		return 1
	}
	return strings.Compare(aID, bID)
}

func sortedNames[T any](items []T, name func(T) string) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, name(item))
	}
	slices.Sort(names)
	return names
}

func validateVerdictPosition(p *verdictPosition) error {
	if p == nil {
		return errors.New("position block is absent")
	}
	if strings.TrimSpace(p.Comment) == "" {
		return errors.New("position comment is blank")
	}
	if p.ScopeGrammar != positionScopeGrammar || p.RunPattern != positionRunPattern || p.YearPattern != positionYearPattern ||
		p.StartPattern != positionStartPattern || p.CVEIDPattern != positionIDPattern || p.CVEIDCanonical != positionCanonical ||
		p.Order != positionOrder {
		return errors.New("position grammar, patterns, canonical rule or order rule changed")
	}
	for _, check := range []struct {
		what string
		got  []string
		want []string
	}{
		{"canonical id", sortedNames(p.CanonicalIDs, func(c positionCanonicalID) string { return c.Name }), positionCanonicalNames},
		{"round-trip", sortedNames(p.RoundTrips, func(c positionRoundTrip) string { return c.Name }), positionRoundTripNames},
		{"rejection", sortedNames(p.Rejections, func(c positionRejection) string { return c.Name }), positionRejectionNames},
		{"order", sortedNames(p.Orders, func(c positionOrderCase) string { return c.Name }), positionOrderNames},
	} {
		want := slices.Clone(check.want)
		slices.Sort(want)
		if !slices.Equal(check.got, want) {
			return fmt.Errorf("position %s case names = %v, want %v", check.what, check.got, want)
		}
	}
	for _, c := range p.CanonicalIDs {
		got, ok := canonicalCVEID(c.Raw)
		if ok != c.Valid || got != c.Canonical || strings.TrimSpace(c.Note) == "" {
			return fmt.Errorf("position canonical id case %q: reference gives (%q, %v), vector says (%q, %v)", c.Name, got, ok, c.Canonical, c.Valid)
		}
	}
	for _, c := range p.RoundTrips {
		want := scopePosition{Run: c.Run, Year: c.Year, StartUnixUS: c.StartUnixUS, FromCVEID: c.FromCVEID}
		got, err := parseBaselineScope(c.ScopeID)
		if err != nil || got != want {
			return fmt.Errorf("position round-trip case %q parses to %+v (%v), want %+v", c.Name, got, err, want)
		}
		if rebuilt := formatBaselineScope(want); rebuilt != c.ScopeID {
			return fmt.Errorf("position round-trip case %q rebuilds %q, want the one scope id %q", c.Name, rebuilt, c.ScopeID)
		}
	}
	for _, c := range p.Rejections {
		if _, err := parseBaselineScope(c.ScopeID); err == nil {
			return fmt.Errorf("position rejection case %q is accepted by the grammar", c.Name)
		}
		if strings.TrimSpace(c.Note) == "" {
			return fmt.Errorf("position rejection case %q has no note", c.Name)
		}
	}
	return validatePositionOrders(p.Orders)
}

func validatePositionOrders(orders []positionOrderCase) error {
	byName := map[string]positionOrderCase{}
	for _, c := range orders {
		byName[c.Name] = c
		for _, side := range []positionFields{c.A, c.B} {
			if side.StartUnixUS <= 0 {
				return fmt.Errorf("position order case %q has a non-positive stamp", c.Name)
			}
			if side.FromCVEID != "" {
				if canonical, ok := canonicalCVEID(side.FromCVEID); !ok || canonical != side.FromCVEID {
					return fmt.Errorf("position order case %q uses the non-canonical id %q", c.Name, side.FromCVEID)
				}
			}
		}
		if got := comparePosition(c.A.StartUnixUS, c.A.FromCVEID, c.B.StartUnixUS, c.B.FromCVEID); got != c.Cmp {
			return fmt.Errorf("position order case %q compares %d, vector says %d", c.Name, got, c.Cmp)
		}
		if strings.TrimSpace(c.Note) == "" {
			return fmt.Errorf("position order case %q has no note", c.Name)
		}
	}
	// The two orders the baseline ruling names are pinned by content, so
	// renaming or rewriting either one fails here and not silently.
	numeric, empty := byName["id_ten_thousand_before_nine_thousand"], byName["empty_id_before_any_id"]
	if numeric.A.FromCVEID != "CVE-2025-10000" || numeric.B.FromCVEID != "CVE-2025-9999" || numeric.A.StartUnixUS != numeric.B.StartUnixUS || numeric.Cmp != -1 {
		return errors.New("order case id_ten_thousand_before_nine_thousand must put CVE-2025-10000 before CVE-2025-9999 at one stamp")
	}
	if empty.A.FromCVEID != "" || empty.B.FromCVEID == "" || empty.A.StartUnixUS != empty.B.StartUnixUS || empty.Cmp != -1 {
		return errors.New("order case empty_id_before_any_id must put (T, empty) before (T, id)")
	}
	return nil
}

// validateCasePosition holds a baseline case to the position order: its
// stamped advance position against its scope's start position. It checks the
// vector's own consistency (a reason cannot name behind-start for a position
// that is ahead); it is not core's reader.
func validateCasePosition(c verdictCase) error {
	for member := range c.Expect {
		if !slices.Contains(baselineExpectMembers, member) {
			return fmt.Errorf("expect member %q is outside the closed baseline expect set", member)
		}
	}
	start, err := parseBaselineScope(c.ScopeID)
	if err != nil {
		return err
	}
	id, err := optionalIDMember(c.Payload, "nvd_baseline_advance_to_cve_id", "payload")
	if err != nil {
		return err
	}
	wantID, err := optionalIDMember(c.Expect, "advance_to_cve_id", "expect")
	if err != nil {
		return err
	}
	stamped, advanceErr := intMember(c.Payload, "nvd_baseline_advance_to_unix_us")
	outcome, reason := stringMember(c.Expect, "outcome"), stringMember(c.Expect, "reason")
	if outcome == "not_covered" && (advanceErr != nil || stamped <= 0) && reason != "advance_missing" {
		return fmt.Errorf("a missing or zero advance_to must expect reason advance_missing, got %q", reason)
	}
	if advanceErr == nil && stamped > 0 {
		complete := boolMember(c.Payload, "nvd_baseline_year_complete")
		cmp := comparePosition(stamped-stamped%1000, id, start.StartUnixUS, start.FromCVEID)
		switch {
		case outcome == "advanced" && (cmp < 0 || (cmp == 0 && !complete)):
			return fmt.Errorf("an advanced case stamps a position (%d, %q) that is behind or not past its scope position", stamped, id)
		case reason == "advance_behind_start" && cmp >= 0:
			return fmt.Errorf("advance_behind_start is expected but the stamped position (%d, %q) is not behind its scope position", stamped, id)
		case reason == "advance_not_past_start" && (cmp != 0 || complete):
			return fmt.Errorf("advance_not_past_start is expected but the stamped position (%d, %q) is not an incomplete tie with its scope position", stamped, id)
		}
	}
	gen, _ := intMember(c.Payload, "nvd_baseline_year_generated_unix_us")
	wantGen := int64(0)
	if _, ok := c.Expect["year_generated_unix_us"]; ok {
		if wantGen, err = intMember(c.Expect, "year_generated_unix_us"); err != nil {
			return err
		}
	}
	if outcome == "advanced" {
		if wantID != id || wantGen != gen {
			return fmt.Errorf("an advanced case expects position id %q and year_generated %d, want the stamped %q and %d", wantID, wantGen, id, gen)
		}
	} else if wantID != "" || wantGen != 0 {
		return fmt.Errorf("outcome %q arms no writer, so expect advance_to_cve_id and year_generated_unix_us must be absent", outcome)
	}
	return nil
}

// optionalIDMember reads an id member that is absent when empty. A present id
// must be a non-empty canonical id: one position has one encoding.
func optionalIDMember(members map[string]json.RawMessage, key, where string) (string, error) {
	raw, ok := members[key]
	if !ok {
		return "", nil
	}
	var id string
	if err := json.Unmarshal(raw, &id); err != nil {
		if where == "expect" {
			return "", fmt.Errorf("expect member %q must be a JSON string, got %s", key, raw)
		}
		return "", nil // a wrong-typed payload stamp is the malformed_keys rule's to judge
	}
	if canonical, valid := canonicalCVEID(id); !valid || canonical != id {
		return "", fmt.Errorf("%s member %q is %q, which is not a non-empty canonical id (an empty id is absent)", where, key, id)
	}
	return id, nil
}
