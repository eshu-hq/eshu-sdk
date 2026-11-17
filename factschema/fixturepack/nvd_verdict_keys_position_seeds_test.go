// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

// payloadEdit and expectEdit adapt a one-member edit to caseEdit.
func payloadEdit(section, name string, edit func(payload map[string]any)) func(map[string]any) {
	return caseEdit(section, name, func(_, payload map[string]any) { edit(payload) })
}

func expectEdit(section, name string, edit func(expect map[string]any)) func(map[string]any) {
	return caseEdit(section, name, func(c, _ map[string]any) { edit(c["expect"].(map[string]any)) })
}

// positionSeeds plants one violation of each rule the position block and the
// position-aware case checks add. Like verdictSeeds, every seed names the
// error it must provoke.
func positionSeeds() []verdictSeed {
	const (
		idKey  = "nvd_baseline_advance_to_cve_id"
		genKey = "nvd_baseline_year_generated_unix_us"
	)
	dropOrder := func(name string) func(map[string]any) {
		return func(d map[string]any) {
			pos := d["position"].(map[string]any)
			orders := pos["orders"].([]any)
			for i, raw := range orders {
				if raw.(map[string]any)["name"] == name {
					pos["orders"] = append(orders[:i:i], orders[i+1:]...)
					return
				}
			}
			panic("no order case " + name)
		}
	}
	return []verdictSeed{
		{name: "position block absent", mutate: func(d map[string]any) { delete(d, "position") }, want: "position block is absent"},
		{name: "position comment blank", mutate: func(d map[string]any) { d["position"].(map[string]any)["comment"] = " " }, want: "position comment is blank"},
		{name: "position scope grammar changed", old: `[/<from_cve_id>]"`, new: `/<from_cve_id>"`, want: "position grammar, patterns, canonical rule or order rule changed"},
		{name: "position id pattern admits lowercase", old: `"cve_id_pattern": "^[A-Z0-9-]{1,64}$"`, new: `"cve_id_pattern": "^[A-Za-z0-9-]{1,64}$"`, want: "position grammar, patterns"},
		{name: "position order rule changed", old: `by unsigned byte order; the empty id sorts before every id`, new: `numerically; the empty id sorts before every id`, want: "position grammar, patterns"},
		{name: "position order case dropped", mutate: dropOrder("hyphen_before_digit"), want: "position order case names"},

		{name: "canonical id case flipped invalid", old: `"canonical": "CVE-2025-1500", "valid": true, "note": "A record id is upper-cased.`, new: `"canonical": "", "valid": false, "note": "A record id is upper-cased.`, want: "position canonical id case"},
		{name: "upper-casing admits a non-ASCII id", old: `"name": "non_ascii_upper_cases_to_ascii", "raw": "cve-2025-\u017f1", "canonical": "", "valid": false`, new: `"name": "non_ascii_upper_cases_to_ascii", "raw": "cve-2025-\u017f1", "canonical": "CVE-2025-S1", "valid": true`, want: "position canonical id case"},

		{name: "round trip with a trailing slash for the empty id", old: `"name": "three_segment_empty_id", "scope_id": "vuln-intel://nvd/_baseline/run-1/2025/1792357800120000"`, new: `"name": "three_segment_empty_id", "scope_id": "vuln-intel://nvd/_baseline/run-1/2025/1792357800120000/"`, want: "position round-trip case"},
		{name: "round trip names another id", old: `"scope_id": "vuln-intel://nvd/_baseline/run-1/2025/1792357800120000/CVE-2025-1500", "run": "run-1", "year": 2025, "start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-1500"`, new: `"scope_id": "vuln-intel://nvd/_baseline/run-1/2025/1792357800120000/CVE-2025-1500", "run": "run-1", "year": 2025, "start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-1501"`, want: "position round-trip case"},
		{name: "rejection the grammar accepts", old: `1792357800120000/cve-2025-1500"`, new: `1792357800120000/CVE-2025-1500"`, want: "is accepted by the grammar"},
		{name: "rejection set loses the long-s trap", mutate: func(d map[string]any) {
			pos := d["position"].(map[string]any)
			var kept []any
			for _, raw := range pos["rejections"].([]any) {
				if raw.(map[string]any)["name"] != "non_ascii_id_upper_cases_to_ascii" {
					kept = append(kept, raw)
				}
			}
			pos["rejections"] = kept
		}, want: "position rejection case names"},

		{name: "order case flipped", old: `"b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-9999"}, "cmp": -1`, new: `"b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-9999"}, "cmp": 1`, want: "position order case"},
		{name: "numeric pair no longer the ruled one", old: `"a": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-10000"}, "b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-9999"}`, new: `"a": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-1000"}, "b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-9999"}`, want: "must put CVE-2025-10000 before CVE-2025-9999"},
		{name: "empty-id pair no longer the ruled one", old: `"a": {"start_unix_us": 1792357800120000, "from_cve_id": ""}, "b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-0001"}`, new: `"a": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-0000"}, "b": {"start_unix_us": 1792357800120000, "from_cve_id": "CVE-2025-0001"}`, want: "must put (T, empty) before (T, id)"},
		{name: "order case with a lowercase id", old: `"from_cve_id": "CVE-2025-0001"}, "cmp": -1`, new: `"from_cve_id": "cve-2025-0001"}, "cmp": -1`, want: "non-canonical id"},
		{name: "order case with a non-positive stamp", old: `{"name": "equal_position", "a": {"start_unix_us": 1792357800120000`, new: `{"name": "equal_position", "a": {"start_unix_us": 0`, want: "non-positive stamp"},

		{name: "advance_to_cve_id declared an int", old: `"name": "nvd_baseline_advance_to_cve_id", "type": "string"`, new: `"name": "nvd_baseline_advance_to_cve_id", "type": "int"`, want: "must be a conditional core-read string"},
		{name: "advance_to_cve_id declared always present", old: `"read_by": "core", "presence": "conditional", "note": "The id half`, new: `"read_by": "core", "presence": "always", "note": "The id half`, want: "must be a conditional core-read string"},
		{name: "year_generated declared conditional", old: `"unit": "unix_us", "read_by": "core", "presence": "always", "note": "The seed manifest lastModified`, new: `"unit": "unix_us", "read_by": "core", "presence": "conditional", "note": "The seed manifest lastModified`, want: "must be an always-present core-read int64 unix_us"},
		{name: "year_generated not read by core", old: `"unit": "unix_us", "read_by": "core", "presence": "always", "note": "The seed manifest lastModified`, new: `"unit": "unix_us", "read_by": "none", "presence": "always", "note": "The seed manifest lastModified`, want: "must be an always-present core-read int64 unix_us"},

		{name: "four-segment scope with a lowercase id", old: `_baseline/run-1/2025/1792357800120000/CVE-2025-1500",`, new: `_baseline/run-1/2025/1792357800120000/cve-2025-1500",`, want: "is not a baseline scope"},
		{name: "emitted baseline case drops year_generated", mutate: payloadEdit("baseline", "advances_past_start", func(p map[string]any) { delete(p, genKey) }), want: "omits the always-present key"},
		{name: "emitted baseline case stamps a zero year_generated", mutate: payloadEdit("baseline", "advances_past_start", func(p map[string]any) { p[genKey] = num("0") }), want: "must stamp a positive year_generated_unix_us"},
		{name: "emitted baseline case drops the id it expects", mutate: payloadEdit("baseline", "advances_cut_inside_instant", func(p map[string]any) { delete(p, idKey) }), want: "an advanced case expects position id"},
		{name: "emitted resume inside an instant drops the id", mutate: payloadEdit("baseline", "advances_within_instant_by_cve_id", func(p map[string]any) { delete(p, idKey) }), want: "at or after its scope start position"},
		{name: "emitted byte-order advance compared numerically", mutate: payloadEdit("baseline", "advances_id_byte_order_past_numeric", func(p map[string]any) { p[idKey] = "CVE-2025-1000" }), want: "at or after its scope start position"},
		{name: "emitted tie stamps should_advance true", mutate: payloadEdit("baseline", "holds_no_progress_at_position", func(p map[string]any) { p["nvd_baseline_should_advance"] = true }), want: "stamps should_advance="},
		{name: "emitted advance stamps a lowercase id", mutate: payloadEdit("baseline", "advances_cut_inside_instant", func(p map[string]any) { p[idKey] = "cve-2025-1500" }), want: "not a non-empty canonical id"},
		{name: "emitted advance stamps an empty id instead of omitting it", mutate: payloadEdit("baseline", "advances_cut_inside_instant", func(p map[string]any) { p[idKey] = "" }), want: "not a non-empty canonical id"},

		{name: "byte-order pin expects numeric order", mutate: expectEdit("baseline", "not_covered_id_numeric_order_behind_position", func(e map[string]any) { e["reason"] = "advance_not_past_start" }), want: "advance_not_past_start is expected but"},
		{name: "behind position expected as an advance", mutate: expectEdit("baseline", "not_covered_dropped_id_behind_position", func(e map[string]any) {
			e["outcome"], e["reason"], e["advance_to"] = "advanced", "", num("1792357800120000")
		}), want: "that is behind or not past its scope position"},
		{name: "dropped id expected as a tie", mutate: expectEdit("baseline", "not_covered_dropped_id_behind_position", func(e map[string]any) { e["reason"] = "advance_not_past_start" }), want: "advance_not_past_start is expected but"},
		{name: "tie expected as behind", mutate: expectEdit("baseline", "not_covered_same_position_incomplete", func(e map[string]any) { e["reason"] = "advance_behind_start" }), want: "advance_behind_start is expected but"},
		{name: "held case arms an id", mutate: expectEdit("baseline", "holds_no_progress_at_position", func(e map[string]any) { e["advance_to_cve_id"] = "CVE-2025-2500" }), want: "arms no writer, so expect advance_to_cve_id"},
		{name: "advanced expect carries another id", mutate: expectEdit("baseline", "advances_cut_inside_instant", func(e map[string]any) { e["advance_to_cve_id"] = "CVE-2025-1501" }), want: "an advanced case expects position id"},
		{name: "advanced expect carries an empty id member", mutate: expectEdit("baseline", "advances_cut_inside_instant", func(e map[string]any) { e["advance_to_cve_id"] = "" }), want: "not a non-empty canonical id"},
		{name: "advanced expect carries another year_generated", mutate: expectEdit("baseline", "advances_cut_inside_instant", func(e map[string]any) { e["year_generated_unix_us"] = num("1") }), want: "an advanced case expects position id"},
		{name: "advanced expect drops year_generated", mutate: expectEdit("baseline", "advances_cut_inside_instant", func(e map[string]any) { delete(e, "year_generated_unix_us") }), want: "an advanced case expects position id"},
		{name: "misspelled expect member", mutate: expectEdit("baseline", "advances_cut_inside_instant", func(e map[string]any) {
			e["advance_to_cve_idd"] = e["advance_to_cve_id"]
			delete(e, "advance_to_cve_id")
		}), want: "outside the closed baseline expect set"},
	}
}
