// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package marker constructs compatible keyed markers for collector source evidence.
package marker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Prefix identifies the persisted marker format.
const (
	Prefix  = "redacted:hmac-sha256:"
	unknown = "unknown"
)

// Key is deployment-scoped secret material used to produce redaction markers.
//
// The zero Key preserves legacy marker behavior but is not suitable for
// protecting sensitive inputs. Construct a key with NewKey and reject IsZero
// before source acquisition in production collectors.
type Key struct {
	material []byte
}

// NewKey constructs a redaction key from deployment-scoped secret material.
//
// The material is copied; callers must keep it stable until NewKey returns.
// Blank key material is rejected because unkeyed
// redaction markers are vulnerable to offline guessing for low-entropy values.
func NewKey(material []byte) (Key, error) {
	if len(strings.TrimSpace(string(material))) == 0 {
		return Key{}, fmt.Errorf("redaction key material must not be blank")
	}
	copied := make([]byte, len(material))
	copy(copied, material)
	return Key{material: copied}, nil
}

// IsZero reports whether the key has no material.
func (k Key) IsZero() bool {
	return len(k.material) == 0
}

// Value is a redaction result that can replace a sensitive scalar in facts
// without retaining raw secret material.
//
// Marker is deterministic for the same key, raw value, reason, and source.
// Reason and Source should be stable classification labels such as
// "sensitive_output" or "aws_db_instance.password"; callers must not place raw
// secret values in either field. This package emits no logs.
type Value struct {
	// Marker is the deterministic replacement string.
	Marker string `json:"marker"`
	// Reason is the normalized classification explaining why redaction happened.
	Reason string `json:"reason"`
	// Source is the normalized caller-provided field or source label.
	Source string `json:"source"`
}

// String turns sensitive string bytes into a deterministic marker.
//
// Empty strings still produce a marker. Blank reason or source values are
// normalized to "unknown". Callers must not use raw secrets as labels.
func String(raw string, reason string, source string, key Key) Value {
	return Bytes([]byte(raw), reason, source, key)
}

// Bytes turns sensitive bytes into a deterministic marker.
//
// The marker digest includes the raw bytes, normalized reason, and normalized
// source. Only the digest is returned; raw bytes are not retained. Callers must
// keep raw stable until Bytes returns.
func Bytes(raw []byte, reason string, source string, key Key) Value {
	normalizedReason := normalizeContext(reason)
	normalizedSource := normalizeContext(source)
	return Value{
		Marker: marker(raw, normalizedReason, normalizedSource, key),
		Reason: normalizedReason,
		Source: normalizedSource,
	}
}

func marker(raw []byte, reason string, source string, key Key) string {
	sum := hmac.New(sha256.New, key.material)
	writeField(sum, []byte("redact.v1"))
	writeField(sum, []byte(reason))
	writeField(sum, []byte(source))
	writeField(sum, raw)
	return Prefix + hex.EncodeToString(sum.Sum(nil))
}

func normalizeContext(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return unknown
	}
	return value
}

type fieldWriter interface {
	Write([]byte) (int, error)
}

func writeField(writer fieldWriter, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write(value)
}
