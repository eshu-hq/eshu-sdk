// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"time"
	"unicode/utf8"

	"github.com/eshu-hq/eshu-sdk/collector"
)

// Handler processes one structurally valid request with its effective deadline.
type Handler func(context.Context, collector.Request) (collector.Result, error)

// Options sets admitted SDK protocols and explicit encoded byte budgets.
type Options struct {
	SupportedProtocols []string
	MaxRequestBytes    int64
	MaxResultBytes     int64
}

// Run reads one JSON request and writes one JSON result, including its newline.
// It stages output before writing; an io.Writer can still partially write on failure.
// A canceled context cannot interrupt an arbitrary blocking reader or handler.
func Run(ctx context.Context, reader io.Reader, writer io.Writer, handler Handler, opts Options) error {
	if ctx == nil || reader == nil || writer == nil || handler == nil ||
		opts.MaxRequestBytes <= 0 || opts.MaxRequestBytes == math.MaxInt64 ||
		opts.MaxResultBytes <= 0 || len(opts.SupportedProtocols) == 0 {
		return errors.New("command options invalid")
	}
	for _, protocol := range opts.SupportedProtocols {
		if protocol != collector.ProtocolVersionV1Alpha1 && protocol != collector.ProtocolVersionV1Alpha2 && protocol != collector.ProtocolVersionV1Alpha3 {
			return errors.New("command options protocol invalid")
		}
	}
	if ctx.Err() != nil {
		return errors.New("command canceled")
	}
	raw, err := io.ReadAll(io.LimitReader(reader, opts.MaxRequestBytes+1))
	if err != nil {
		return errors.New("command read failed")
	}
	if int64(len(raw)) > opts.MaxRequestBytes {
		return errors.New("command request limit exceeded")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var request *collector.Request
	if decoder.Decode(&request) != nil || request == nil {
		return errors.New("command decode failed")
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return errors.New("command trailing data")
	}
	if request.ProtocolVersion == collector.ProtocolVersionV1Alpha3 && !utf8.Valid(raw) {
		return errors.New("command decode failed")
	}
	if request.Validate() != nil {
		return errors.New("command request invalid")
	}
	admitted := false
	for _, protocol := range opts.SupportedProtocols {
		if request.ProtocolVersion == protocol {
			admitted = true
			break
		}
	}
	if !admitted {
		return errors.New("command protocol unsupported")
	}
	if !request.Claim.Deadline.After(time.Now()) {
		return errors.New("command deadline expired")
	}
	callCtx, cancel := context.WithDeadline(ctx, request.Claim.Deadline)
	defer cancel()
	if callCtx.Err() != nil {
		return errors.New("command canceled")
	}
	result, err := handler(callCtx, *request)
	if err != nil {
		return errors.New("command handler failed")
	}
	if callCtx.Err() != nil {
		return errors.New("command canceled")
	}
	if request.ProtocolVersion == collector.ProtocolVersionV1Alpha3 && collector.ValidateCoverageUTF8(result) != nil {
		return errors.New("command encode failed")
	}
	var staged bytes.Buffer
	if err := json.NewEncoder(&staged).Encode(result); err != nil {
		return errors.New("command encode failed")
	}
	if int64(staged.Len()) > opts.MaxResultBytes {
		return errors.New("command result limit exceeded")
	}
	if callCtx.Err() != nil {
		return errors.New("command canceled")
	}
	n, err := writer.Write(staged.Bytes())
	if err != nil || n != staged.Len() {
		return errors.New("command write failed")
	}
	return nil
}
