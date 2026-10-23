// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package identity derives stable fact identifiers from a fact type and an
// identity map. StableID preserves the pinned Eshu core algorithm: recursive
// map and slice normalization, UTC RFC3339Nano time formatting, JSON encoding,
// and SHA-256 hex output. A changed result can orphan existing fact identity.
//
// This package is a candidate public copy. Core callers still use the in-tree
// forwarder until a separately validated SDK release and consumer cutover.
package identity
