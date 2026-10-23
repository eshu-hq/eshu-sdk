// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package entrypoints

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/eshu-hq/eshu-sdk/collector"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGeneratedCommandContract(t *testing.T) {
	manifest := validManifest()
	manifest.HandlerImportPath = "github.com/eshu-hq/eshu-sdk/collector/entrypoints/testdata/provider"
	manifest.HandlerFunc = "Handle"
	manifest.MaxRequestBytes = 1 << 20
	manifest.MaxResultBytes = 4 << 20
	files, err := Generate(manifest)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainPath, files[0].Contents, 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "collector")
	buildCtx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, mainPath)
	build.Env = append(os.Environ(), "GOWORK=off")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated command: %v: %s", err, output)
	}
	raw, err := os.ReadFile("../testdata/fixtures/request.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var request collector.Request
	if err := decoder.Decode(&request); err != nil {
		t.Fatal(err)
	}
	request.Contract.Facts = []collector.FactDeclaration{{Kind: "dev.eshu.examples.scorecard.check", SchemaVersions: []string{"1.0.0"}, SourceConfidence: []collector.SourceConfidence{collector.SourceConfidenceObserved}}}
	run := func(input []byte) ([]byte, string, error) {
		t.Helper()
		ctx, stop := context.WithTimeout(t.Context(), 3*time.Second)
		defer stop()
		cmd := exec.CommandContext(ctx, binary, "--sdk-stdio")
		cmd.Stdin = bytes.NewReader(input)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		return out.Bytes(), stderr.String(), err
	}
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	out, stderr, err := run(input)
	if err != nil || stderr != "" {
		t.Fatalf("valid command: %v stderr=%q", err, stderr)
	}
	var result collector.Result
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if _, err := collector.NewValidator(request.Contract).ValidateResult(result); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(input)
	if got := result.Facts[0].Payload["request_digest"]; got != hex.EncodeToString(sum[:]) {
		t.Fatalf("opaque request changed: %v", got)
	}
	request.Config["proof_mode"] = "handler_error"
	input, err = json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	out, stderr, err = run(input)
	if err == nil || len(out) != 0 || !strings.Contains(stderr, "command handler failed") || strings.Contains(stderr, "SECRET_CANARY_8851") {
		t.Fatalf("handler stage or redaction failed: err=%v out=%q stderr=%q", err, out, stderr)
	}
	out, stderr, err = run([]byte("{invalid SECRET_CANARY_8851"))
	if err == nil || len(out) != 0 || !strings.Contains(stderr, "command decode failed") || strings.Contains(stderr, "SECRET_CANARY_8851") {
		t.Fatalf("decode stage or redaction failed: err=%v out=%q stderr=%q", err, out, stderr)
	}
}
