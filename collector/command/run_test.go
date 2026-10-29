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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eshu-hq/eshu-sdk/collector"
)

func requestBytes(t *testing.T) []byte {
	t.Helper()
	r := collector.Request{
		ProtocolVersion: collector.ProtocolVersionV1Alpha1,
		Claim:           collector.Claim{ComponentID: "c", InstanceID: "i", CollectorKind: "k", SourceSystem: "s", Scope: collector.Scope{ID: "scope", Kind: "repo"}, SourceRunID: "run", GenerationID: "gen", WorkItemID: "work", FencingToken: "fence", Attempt: 1, Deadline: time.Now().Add(time.Hour), ConfigHandle: "config"},
		Contract:        collector.Contract{ProtocolVersion: collector.ProtocolVersionV1Alpha1},
		Config:          map[string]any{"large": json.Number("9007199254740993")},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func optionsFor(raw []byte) Options {
	return Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}
}

func TestRunPreservesRequestAndResult(t *testing.T) {
	raw := requestBytes(t)
	result := collector.Result{ProtocolVersion: collector.ProtocolVersionV1Alpha1, State: collector.ResultRetryable}
	for _, state := range []collector.ResultState{collector.ResultComplete, collector.ResultPartial, collector.ResultUnchanged, collector.ResultRetryable, collector.ResultTerminal} {
		t.Run(string(state), func(t *testing.T) {
			result.State = state
			var out bytes.Buffer
			err := Run(context.Background(), bytes.NewReader(raw), &out, func(_ context.Context, req collector.Request) (collector.Result, error) {
				if got := req.Config["large"].(json.Number).String(); got != "9007199254740993" {
					t.Errorf("config number = %s", got)
				}
				return result, nil
			}, optionsFor(raw))
			if err != nil {
				t.Fatal(err)
			}
			var got collector.Result
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.State != state {
				t.Fatalf("state = %q", got.State)
			}
			want, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			want = append(want, '\n')
			if !bytes.Equal(out.Bytes(), want) {
				t.Fatalf("result bytes changed: %q", out.Bytes())
			}
		})
	}
}

func TestRunRejectsInputAndOptionsBeforeHandler(t *testing.T) {
	raw := requestBytes(t)
	cases := map[string]struct {
		input []byte
		opts  Options
	}{
		"null":                         {[]byte("null"), optionsFor(raw)},
		"array":                        {[]byte("[]"), optionsFor(raw)},
		"unknown field":                {append(bytes.TrimSuffix(raw, []byte("}")), []byte(`,"unknown":1}`)...), optionsFor(raw)},
		"unknown claim field":          {bytes.Replace(raw, []byte(`"component_id":"c"`), []byte(`"component_id":"c","secret_field":"SECRET"`), 1), optionsFor(raw)},
		"wrong typed claim field":      {bytes.Replace(raw, []byte(`"attempt":1`), []byte(`"attempt":"1"`), 1), optionsFor(raw)},
		"trailing object":              {append(append([]byte{}, raw...), []byte(" {}")...), optionsFor(raw)},
		"overflow":                     {append(append([]byte{}, raw...), ' '), optionsFor(raw)},
		"max int budget":               {raw, Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1}, MaxRequestBytes: math.MaxInt64, MaxResultBytes: 4096}},
		"zero result budget":           {raw, Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1}, MaxRequestBytes: int64(len(raw))}},
		"unknown allowlist":            {raw, Options{SupportedProtocols: []string{"unknown"}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}},
		"unsupported request protocol": {bytes.ReplaceAll(raw, []byte(collector.ProtocolVersionV1Alpha1), []byte(collector.ProtocolVersionV1Alpha2)), optionsFor(raw)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			called := false
			var out bytes.Buffer
			err := Run(context.Background(), bytes.NewReader(tc.input), &out, func(context.Context, collector.Request) (collector.Result, error) {
				called = true
				return collector.Result{}, nil
			}, tc.opts)
			if err == nil || called || out.Len() != 0 {
				t.Fatalf("err=%v called=%v output=%q", err, called, out.String())
			}
			if strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("input leaked: %v", err)
			}
		})
	}
}

func TestRunExplicitSecondProtocolAndConcurrentCalls(t *testing.T) {
	raw := bytes.ReplaceAll(requestBytes(t), []byte(collector.ProtocolVersionV1Alpha1), []byte(collector.ProtocolVersionV1Alpha2))
	opts := Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1, collector.ProtocolVersionV1Alpha2}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out bytes.Buffer
			err := Run(context.Background(), bytes.NewReader(raw), &out, func(_ context.Context, req collector.Request) (collector.Result, error) {
				return collector.Result{ProtocolVersion: req.ProtocolVersion, State: collector.ResultUnchanged}, nil
			}, opts)
			if err != nil {
				t.Errorf("Run: %v", err)
				return
			}
			var result collector.Result
			if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.ProtocolVersion != collector.ProtocolVersionV1Alpha2 {
				t.Errorf("result=%+v err=%v", result, err)
			}
		}()
	}
	wg.Wait()
}

func TestRunDeadlineCancellationAndSafeErrors(t *testing.T) {
	raw := requestBytes(t)
	var expired collector.Request
	if err := json.Unmarshal(raw, &expired); err != nil {
		t.Fatal(err)
	}
	expired.Claim.Deadline = time.Now().Add(-time.Second)
	expiredRaw, err := json.Marshal(expired)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		ctx   context.Context
		input []byte
	}{
		{"parent canceled", canceledContext(), raw},
		{"expired claim", context.Background(), expiredRaw},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			opts := optionsFor(raw)
			opts.MaxRequestBytes = int64(len(tc.input))
			err := Run(tc.ctx, bytes.NewReader(tc.input), io.Discard, func(context.Context, collector.Request) (collector.Result, error) {
				called = true
				return collector.Result{}, nil
			}, opts)
			if err == nil || called {
				t.Fatalf("err=%v called=%v", err, called)
			}
		})
	}
	var out bytes.Buffer
	err = Run(context.Background(), bytes.NewReader(raw), &out, func(context.Context, collector.Request) (collector.Result, error) {
		return collector.Result{}, errors.New("SECRET-PROVIDER")
	}, optionsFor(raw))
	if err == nil || strings.Contains(err.Error(), "SECRET") || out.Len() != 0 {
		t.Fatalf("unsafe handler failure: %v %q", err, out.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = Run(ctx, bytes.NewReader(raw), &out, func(context.Context, collector.Request) (collector.Result, error) {
		cancel()
		return collector.Result{}, nil
	}, optionsFor(raw))
	if err == nil || out.Len() != 0 {
		t.Fatalf("late cancellation: %v %q", err, out.String())
	}
	parent, stop := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer stop()
	err = Run(parent, bytes.NewReader(raw), io.Discard, func(call context.Context, _ collector.Request) (collector.Result, error) {
		deadline, ok := call.Deadline()
		if !ok || !deadline.Equal(parentDeadline(t, parent)) {
			t.Errorf("effective deadline=%v", deadline)
		}
		return collector.Result{}, nil
	}, optionsFor(raw))
	if err != nil {
		t.Fatal(err)
	}
	var requested collector.Request
	if err := json.Unmarshal(raw, &requested); err != nil {
		t.Fatal(err)
	}
	claimDeadline := requested.Claim.Deadline
	err = Run(context.Background(), bytes.NewReader(raw), io.Discard, func(call context.Context, _ collector.Request) (collector.Result, error) {
		deadline, ok := call.Deadline()
		if !ok || !deadline.Equal(claimDeadline) {
			t.Errorf("claim deadline=%v", deadline)
		}
		return collector.Result{}, nil
	}, optionsFor(raw))
	if err != nil {
		t.Fatal(err)
	}
}

func parentDeadline(t *testing.T, ctx context.Context) time.Time {
	t.Helper()
	deadline, _ := ctx.Deadline()
	return deadline
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return 0, nil }

func TestRunStagesOutputAndReportsWriteFailure(t *testing.T) {
	raw := requestBytes(t)
	handler := func(context.Context, collector.Request) (collector.Result, error) {
		return collector.Result{State: collector.ResultTerminal}, nil
	}
	var out bytes.Buffer
	opts := optionsFor(raw)
	opts.MaxResultBytes = 1
	if err := Run(context.Background(), bytes.NewReader(raw), &out, handler, opts); err == nil || out.Len() != 0 {
		t.Fatalf("limit err=%v out=%q", err, out.String())
	}
	if err := Run(context.Background(), bytes.NewReader(raw), shortWriter{}, handler, optionsFor(raw)); err == nil {
		t.Fatal("short write accepted")
	}
	var expected bytes.Buffer
	if err := json.NewEncoder(&expected).Encode(collector.Result{State: collector.ResultTerminal}); err != nil {
		t.Fatal(err)
	}
	opts.MaxResultBytes = int64(expected.Len())
	if err := Run(context.Background(), bytes.NewReader(raw), &out, handler, opts); err != nil {
		t.Fatalf("exact result limit: %v", err)
	}
	if !bytes.Equal(out.Bytes(), expected.Bytes()) {
		t.Fatalf("output=%q expected=%q", out.Bytes(), expected.Bytes())
	}
	out.Reset()
	opts.MaxResultBytes--
	if err := Run(context.Background(), bytes.NewReader(raw), &out, handler, opts); err == nil || out.Len() != 0 {
		t.Fatalf("newline limit: %v %q", err, out.String())
	}
	out.Reset()
	bad := func(context.Context, collector.Request) (collector.Result, error) {
		return collector.Result{Facts: []collector.Fact{{Payload: map[string]any{"bad": make(chan int)}}}}, nil
	}
	if err := Run(context.Background(), bytes.NewReader(raw), &out, bad, optionsFor(raw)); err == nil || out.Len() != 0 {
		t.Fatalf("encode failure: %v %q", err, out.String())
	}
}

func coverageRequestBytes(t *testing.T) []byte {
	t.Helper()
	var request collector.Request
	if err := json.Unmarshal(requestBytes(t), &request); err != nil {
		t.Fatal(err)
	}
	request.ProtocolVersion = collector.ProtocolVersionV1Alpha3
	request.Contract.ProtocolVersion = collector.ProtocolVersionV1Alpha3
	request.Contract.Coverage = &collector.CoverageContract{
		Scope: collector.CoverageScope{TenantID: "tenant", ComponentID: request.Claim.ComponentID, AccountID: "123456789012", Partition: "aws", Service: "ec2", Region: "us-east-1"},
		Root:  collector.RootEC2Resources, MaxResources: 1,
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestRunRejectsInvalidUTF8CoverageWire(t *testing.T) {
	valid := coverageRequestBytes(t)
	invalidTenant := append([]byte(`"tenant_id":"`), 0xff)
	invalidTenant = append(invalidTenant, '"')
	raw := bytes.Replace(valid, []byte(`"tenant_id":"tenant"`), invalidTenant, 1)
	if bytes.Equal(raw, valid) {
		t.Fatal("did not seed invalid tenant bytes")
	}
	opts := Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha3}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}
	called := false
	var output bytes.Buffer
	err := Run(context.Background(), bytes.NewReader(raw), &output, func(context.Context, collector.Request) (collector.Result, error) {
		called = true
		return collector.Result{}, nil
	}, opts)
	if err == nil || called || output.Len() != 0 {
		t.Fatalf("invalid request reached handler: err=%v called=%t bytes=%d", err, called, output.Len())
	}
}

func TestRunRejectsInvalidUTF8CoverageResultBeforeEncode(t *testing.T) {
	raw := coverageRequestBytes(t)
	opts := Options{SupportedProtocols: []string{collector.ProtocolVersionV1Alpha3}, MaxRequestBytes: int64(len(raw)), MaxResultBytes: 4096}
	bad := string([]byte{'i', 0xff})
	for _, tc := range []struct {
		name   string
		change func(*collector.Result)
	}{
		{"resource identity", func(r *collector.Result) {
			r.Coverage.Resources = []collector.ResourceObservation{{Kind: collector.ResourceEC2Core, ID: bad, State: collector.ResourcePresent}}
		}},
		{"set identity", func(r *collector.Result) {
			r.Coverage.Sets = []collector.SetCoverage{{Set: collector.OwnershipSet{Collection: collector.CollectionEC2Core, ResourceKind: collector.ResourceEC2Core, ResourceID: bad}}}
		}},
		{"fact ownership", func(r *collector.Result) {
			r.Facts = []collector.Fact{{Ownership: &collector.OwnershipSet{Collection: collector.CollectionEC2Core, ResourceKind: collector.ResourceEC2Core, ResourceID: bad}}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			err := Run(context.Background(), bytes.NewReader(raw), &output, func(context.Context, collector.Request) (collector.Result, error) {
				result := collector.Result{ProtocolVersion: collector.ProtocolVersionV1Alpha3, Coverage: &collector.CoverageReport{}}
				tc.change(&result)
				return result, nil
			}, opts)
			if err == nil || output.Len() != 0 {
				t.Fatalf("invalid result was encoded: err=%v bytes=%d", err, output.Len())
			}
		})
	}
}
