// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package apicontract

// EnvelopeMIMEType selects the stable Eshu response envelope.
//
// The value is copied from Eshu core
// (go/internal/query/querycontract http.go); core is the provenance of
// record. This module depends on nothing core-internal: the string is
// duplicated here so out-of-tree gates and thin clients can negotiate the
// envelope without importing go/internal packages. The frozen contract test
// pins the exact bytes; do not change this string without moving core and
// every consumer together.
const EnvelopeMIMEType = "application/eshu.envelope+json"
