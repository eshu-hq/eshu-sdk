// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	sdk "github.com/eshu-hq/eshu-sdk/collector"
	"os"
	"path/filepath"
	"time"
)

// Handle returns deterministic source facts for generated-command contract tests.
func Handle(ctx context.Context, r sdk.Request) (sdk.Result, error) {
	mode, _ := r.Config["proof_mode"].(string)
	if mode == "handler_error" {
		return sdk.Result{}, errors.New("SECRET_CANARY_8851")
	}
	if mode == "cooperative" {
		<-ctx.Done()
		return sdk.Result{}, ctx.Err()
	}
	if mode == "noncooperative" {
		time.Sleep(time.Hour)
	}
	if mode == "barrier" {
		dir, _ := r.Config["barrier_dir"].(string)
		id, _ := r.Config["barrier_id"].(string)
		peer := "a"
		if id == "a" {
			peer = "b"
		}
		if err := os.WriteFile(filepath.Join(dir, id), []byte("started"), 0600); err != nil {
			return sdk.Result{}, err
		}
		for {
			if _, err := os.Stat(filepath.Join(dir, peer)); err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return sdk.Result{}, ctx.Err()
			case <-time.After(time.Millisecond):
			}
		}
	}
	state := sdk.ResultComplete
	switch mode {
	case "partial":
		state = sdk.ResultPartial
	case "unchanged":
		state = sdk.ResultUnchanged
	case "retryable":
		state = sdk.ResultRetryable
	case "terminal":
		state = sdk.ResultTerminal
	}
	now := time.Now().UTC()
	result := sdk.Result{ProtocolVersion: r.ProtocolVersion, Claim: r.Claim, State: state, Generation: sdk.Generation{ID: r.Claim.GenerationID, ObservedAt: now}}
	if state == sdk.ResultRetryable || state == sdk.ResultTerminal {
		result.Statuses = []sdk.Status{{Class: sdk.StatusFailure, FailureClass: "fixture_failure", RetryAfterSeconds: 1}}
		return result, nil
	}
	if state == sdk.ResultUnchanged {
		return result, nil
	}
	raw, _ := json.Marshal(r)
	hash := sha256.Sum256(raw)
	key := "scorecard:repo-eshu"
	f := sdk.Fact{Kind: r.Contract.Facts[0].Kind, SchemaVersion: "1.0.0", StableKey: key, ObservedAt: now, SourceConfidence: sdk.SourceConfidenceObserved, SourceRef: sdk.SourceRef{SourceSystem: r.Claim.SourceSystem, ScopeID: r.Claim.Scope.ID, GenerationID: r.Claim.GenerationID, FactKey: key, URI: "git://repo-eshu", RecordID: "fixture"}, Payload: map[string]any{"score": 98, "state": "pass", "request_digest": hex.EncodeToString(hash[:])}}
	result.Facts = []sdk.Fact{f}
	if state == sdk.ResultPartial {
		result.Statuses = []sdk.Status{{Class: sdk.StatusWarning, Partial: true, WarningCount: 1, FactCount: 1}}
	}
	if mode == "identity" {
		result.Claim.FencingToken = "wrong"
	}
	if mode == "conflict" {
		other := f
		other.Payload = map[string]any{"score": 99}
		result.Facts = append(result.Facts, other)
	}
	if mode == "encode_error" {
		result.Facts[0].Payload["bad"] = make(chan int)
	}
	if mode == "output_secret" {
		result.Facts[0].Payload["token"] = "SECRET_CANARY_8851"
	}
	return result, nil
}
