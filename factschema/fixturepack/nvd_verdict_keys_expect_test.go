// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// validateEmittedPayload holds a case flagged emitted_by_collector to the
// shape a real collector builder can stamp: every always-present key of its
// section, a self-consistent verdict, and no deliberate wrong types. It
// states the builders' output shape; it does not re-implement core's reader.
func validateEmittedPayload(section string, s verdictSection, c verdictCase) error {
	if len(c.MalformedKeys) > 0 {
		return fmt.Errorf("an emitted case must not carry malformed_keys")
	}
	for _, key := range s.Keys {
		if _, ok := c.Payload[key.Name]; !ok && key.Presence == "always" {
			return fmt.Errorf("an emitted case omits the always-present key %q", key.Name)
		}
	}
	prefix := "nvd_" + section + "_"
	if section == "window" {
		advance, _ := intMember(c.Payload, prefix+"advance_to_unix_us")
		head, _ := intMember(c.Payload, prefix+"head_unix_us")
		switch should := boolMember(c.Payload, prefix+"should_advance"); {
		case should && (advance <= 0 || advance > head):
			return fmt.Errorf("an emitted window case stamps should_advance true without 0 < advance_to <= head")
		case !should && advance != 0:
			return fmt.Errorf("an emitted window case stamps a non-zero advance_to with should_advance false")
		}
		return nil
	}
	for name := range c.Payload {
		if strings.HasPrefix(name, "nvd_window_") {
			return fmt.Errorf("an emitted baseline case carries the window key %q", name)
		}
	}
	m := baselineScopeShape.FindStringSubmatch(c.ScopeID)
	start, _ := strconv.ParseInt(m[2], 10, 64)
	advance, _ := intMember(c.Payload, prefix+"advance_to_unix_us")
	if advance%1000 != 0 || advance < start {
		return fmt.Errorf("an emitted baseline case stamps advance_to %d, which is not a whole millisecond at or after its scope start %d", advance, start)
	}
	complete := boolMember(c.Payload, prefix+"year_complete")
	if should := boolMember(c.Payload, prefix+"should_advance"); should != (complete || advance > start) {
		return fmt.Errorf("an emitted baseline case stamps should_advance=%v, want complete or advance_to past start", should)
	}
	if !yearSHAShape.MatchString(stringMember(c.Payload, prefix+"year_sha256")) {
		return fmt.Errorf("an emitted baseline case stamps a year_sha256 that is not 64 lowercase hex")
	}
	return nil
}

func validateVerdictExpect(section string, s verdictSection, c verdictCase) error {
	outcome := stringMember(c.Expect, "outcome")
	if !slices.Contains(s.Outcomes, outcome) {
		return fmt.Errorf("expect.outcome %q is outside the closed outcome set", outcome)
	}
	advanceTo, err := intMember(c.Expect, "advance_to")
	if err != nil {
		return err
	}
	prefix := "nvd_" + section + "_"
	if section == "window" {
		return validateWindowExpect(c, outcome, advanceTo, prefix)
	}
	// Baseline: expect.advance_to is the whole-millisecond value the armed
	// checkpoint writer would carry, 0 when nothing is armed.
	if advanceTo%1000 != 0 {
		return fmt.Errorf("expect.advance_to %d must be a whole millisecond", advanceTo)
	}
	completeRaw := string(c.Expect["year_complete"])
	if completeRaw != "true" && completeRaw != "false" {
		return fmt.Errorf("expect.year_complete must be a JSON bool")
	}
	reason, class := stringMember(c.Expect, "reason"), stringMember(c.Expect, "class")
	switch {
	case outcome == "not_covered" && !slices.Contains(s.Reasons, reason):
		return fmt.Errorf("a not_covered case must expect a reason from the closed reason set, got %q", reason)
	case outcome != "not_covered" && reason != "":
		return fmt.Errorf("expect.reason %q is set on outcome %q, but only not_covered carries one", reason, outcome)
	case outcome == "refused" && !slices.Contains(s.Classes, class):
		return fmt.Errorf("a refused case must expect a class from the closed class set, got %q", class)
	case outcome != "refused" && class != "":
		return fmt.Errorf("expect.class %q is set on outcome %q, but only refused carries one", class, outcome)
	}
	if outcome != "advanced" {
		if advanceTo != 0 || completeRaw == "true" {
			return fmt.Errorf("outcome %q arms no writer, so expect advance_to and year_complete must be 0 and false", outcome)
		}
		return nil
	}
	payloadAdvance, ok := c.Payload[prefix+"advance_to_unix_us"]
	if !ok {
		return fmt.Errorf("an advanced baseline case must stamp advance_to")
	}
	stamped, err := strconv.ParseInt(strings.TrimSpace(string(payloadAdvance)), 10, 64)
	if err != nil || advanceTo != stamped-stamped%1000 {
		return fmt.Errorf("an advanced baseline case expects advance_to %d, want the stamp floored to a whole millisecond", advanceTo)
	}
	if string(c.Payload[prefix+"year_complete"]) != completeRaw {
		return fmt.Errorf("an advanced baseline case expects year_complete to equal the stamp")
	}
	return nil
}

func validateWindowExpect(c verdictCase, outcome string, advanceTo int64, prefix string) error {
	head, err := intMember(c.Expect, "head")
	if err != nil {
		return err
	}
	if outcome == "held" && (advanceTo != 0 || head != 0) {
		return fmt.Errorf("a held window arms no writer, so expect advance_to and head must be 0")
	}
	if outcome == "advanced" {
		if got, _ := intMember(c.Payload, prefix+"advance_to_unix_us"); got != advanceTo {
			return fmt.Errorf("an advanced window expects advance_to %d but stamps %d", advanceTo, got)
		}
		if got, _ := intMember(c.Payload, prefix+"head_unix_us"); got != head {
			return fmt.Errorf("an advanced window expects head %d but stamps %d", head, got)
		}
	}
	return nil
}
