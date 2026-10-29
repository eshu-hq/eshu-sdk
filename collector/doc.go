// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package collector defines the public Eshu collector extension SDK contracts.
//
// The SDK is intentionally independent from Eshu internal Go packages. It
// describes versioned collector-sdk wire records an out-of-tree collector can
// return to a core-owned extension host, plus validators that fail closed before
// the host commits facts. The validator admits narrowly scoped, legacy redacted
// AWS Lambda and ECS environment and reference payloads while recursively
// checking all other fields for sensitive keys.
package collector
