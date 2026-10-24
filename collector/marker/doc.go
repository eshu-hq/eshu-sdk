// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package marker creates deterministic, keyed HMAC-SHA256 markers for raw
// values that a collector must replace in source evidence. String and Bytes
// return a marker plus normalized reason and source classification labels;
// neither raw input nor the key is retained in Value. Callers must keep raw
// secrets out of the reason and source labels.
//
// The zero Key preserves legacy marker behavior but is not suitable for
// protecting sensitive inputs. Construct a key with NewKey and reject IsZero
// before source acquisition in production collectors.
package marker
