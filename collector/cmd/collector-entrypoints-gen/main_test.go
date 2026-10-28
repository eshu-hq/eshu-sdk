// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCheckStaleAndSymlink(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "manifest.yaml")
	content := "schema_version: 1\ncollectors:\n- schema_version: 1\n  command_dir: cmd/sample\n  handler_import_path: example.com/provider/handler\n  handler_func: Handle\n  supported_protocols: [collector-sdk/v1alpha1]\n  max_request_bytes: 1024\n  max_result_bytes: 2048\n"
	if err := os.WriteFile(manifest, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--manifest", manifest, "--repo-root", root}
	if err := run(append(args, "--check")); err == nil {
		t.Fatal("missing output accepted")
	}
	if err := run(append(args, "--write")); err != nil {
		t.Fatal(err)
	}
	if err := run(append(args, "--check")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "cmd/sample/main.go")
	if err := os.WriteFile(path, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(append(args, "--check")); err == nil {
		t.Fatal("stale output accepted")
	}
	if err := run(append(args, "--write")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}
	if err := run(append(args, "--write")); err == nil {
		t.Fatal("symlink overwrite accepted")
	}
	got, err := os.ReadFile(outside)
	if err != nil || string(got) != "safe" {
		t.Fatalf("outside changed: %q %v", got, err)
	}
}

func TestRejectInvalidOptionsAndSymlinkDirectory(t *testing.T) {
	for _, args := range [][]string{{}, {"--write"}, {"--check", "--write"}, {"--repo-root", ".", "--manifest", "x", "--write", "extra"}} {
		if err := run(args); err == nil {
			t.Fatalf("accepted options: %q", args)
		}
	}
	root := t.TempDir()
	outside := t.TempDir()
	manifest := filepath.Join(root, "manifest.yaml")
	content := "schema_version: 1\ncollectors:\n- schema_version: 1\n  command_dir: cmd/sample\n  handler_import_path: example.com/provider/handler\n  handler_func: Handle\n  supported_protocols: [collector-sdk/v1alpha1]\n  max_request_bytes: 1024\n  max_result_bytes: 2048\n"
	if err := os.WriteFile(manifest, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "cmd")); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"--manifest", manifest, "--repo-root", root, "--write"}); err == nil {
		t.Fatal("symlink directory accepted")
	}
}
