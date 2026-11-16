// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import _ "embed"

//go:embed payloads/nvd_verdict_keys_golden.json
var nvdVerdictKeysGolden []byte

// NVDVerdictKeysGolden returns the shared verdict-key vector for the NVD
// window and baseline checkpoint verdicts a satellite stamps on the
// vulnerability.source_snapshot payload and the Eshu core extension host reads
// (eshu-hq/eshu-sdk#93, eshu-hq/eshu#10732). Each section lists the payload
// keys with their JSON type, unit and reader, then named cases: a stamped
// payload, its scope id, and the commit outcome the core reader reaches. The
// extracted collector asserts that its constants and real verdict builders
// match the keys and payload shapes; core replays every case through its
// commit path, so a rename or a changed verdict rule fails both suites.
func NVDVerdictKeysGolden() []byte {
	return append([]byte(nil), nvdVerdictKeysGolden...)
}
