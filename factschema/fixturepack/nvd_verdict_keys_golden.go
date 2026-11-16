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
// payload, its scope id, the commit outcome the core reader reaches, and
// whether a real collector builder can emit that exact payload. The collector
// and Eshu core are expected to read it in their own changes: the collector
// to check its key constants and verdict builders against the keys, core to
// replay every case through its commit path. This package validates only the
// vector's own shape.
func NVDVerdictKeysGolden() []byte {
	return append([]byte(nil), nvdVerdictKeysGolden...)
}
