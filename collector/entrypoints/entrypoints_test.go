// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package entrypoints

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector"
)

func validManifest() Manifest {
	return Manifest{SchemaVersion: 1, CommandDir: "cmd/sample", HandlerImportPath: "example.com/provider/handler", HandlerFunc: "Handle", SupportedProtocols: []string{collector.ProtocolVersionV1Alpha1}, MaxRequestBytes: 1024, MaxResultBytes: 2048}
}

func TestGenerateDeterministicAndBounded(t *testing.T) {
	m := validManifest()
	a, err := Generate(m)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 1 || a[0].Name != "main.go" || !bytes.Equal(a[0].Contents, b[0].Contents) {
		t.Fatalf("nondeterministic output: %#v", a)
	}
	for _, want := range []string{"command.Run", "provider.Handle", "signal.NotifyContext", "sdk-stdio", "MaxRequestBytes:", "1024", "MaxResultBytes:", "2048", "collector-sdk/v1alpha1"} {
		if !bytes.Contains(a[0].Contents, []byte(want)) {
			t.Errorf("missing %q", want)
		}
	}
	for _, forbidden := range []string{"/go/internal/", "Disposable local boundary proof"} {
		if bytes.Contains(a[0].Contents, []byte(forbidden)) {
			t.Errorf("unsafe output contains %q", forbidden)
		}
	}
}

func TestGenerateRejectsUnsafeManifest(t *testing.T) {
	cases := map[string]func(*Manifest){
		"absolute target":    func(m *Manifest) { m.CommandDir = "/tmp/out" },
		"parent target":      func(m *Manifest) { m.CommandDir = "../out" },
		"backslash target":   func(m *Manifest) { m.CommandDir = `cmd\out` },
		"quoted target":      func(m *Manifest) { m.CommandDir = `cmd/"out` },
		"quoted import":      func(m *Manifest) { m.HandlerImportPath = `evil\";panic(1)` },
		"core import":        func(m *Manifest) { m.HandlerImportPath = "github.com/eshu-hq/eshu/go/internal/query" },
		"private handler":    func(m *Manifest) { m.HandlerFunc = "handle" },
		"expression handler": func(m *Manifest) { m.HandlerFunc = "Handle()" },
		"unknown protocol":   func(m *Manifest) { m.SupportedProtocols = []string{"v3"} },
		"duplicate protocol": func(m *Manifest) {
			m.SupportedProtocols = append(m.SupportedProtocols, collector.ProtocolVersionV1Alpha1)
		},
		"zero request":     func(m *Manifest) { m.MaxRequestBytes = 0 },
		"overflow request": func(m *Manifest) { m.MaxRequestBytes = math.MaxInt64 },
		"zero result":      func(m *Manifest) { m.MaxResultBytes = 0 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := validManifest()
			mutate(&m)
			if _, err := Generate(m); err == nil {
				t.Fatal("accepted unsafe manifest")
			}
		})
	}
}

func TestLoadManifestStrict(t *testing.T) {
	base := "schema_version: 1\ncollectors:\n- schema_version: 1\n  command_dir: cmd/sample\n  handler_import_path: example.com/provider/handler\n  handler_func: Handle\n  supported_protocols: [collector-sdk/v1alpha1]\n  max_request_bytes: 1024\n  max_result_bytes: 2048\n"
	for name, content := range map[string]string{
		"valid":            base,
		"unknown":          base + "unknown: true\n",
		"second document":  base + "---\nschema_version: 1\n",
		"duplicate target": base + strings.TrimPrefix(base, "schema_version: 1\ncollectors:\n"),
		"duplicate key":    strings.Replace(base, "schema_version: 1\n", "schema_version: 1\nschema_version: 1\n", 1),
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "manifest.yaml")
			if err := os.WriteFile(p, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := LoadManifestFile(p)
			if (err == nil) != (name == "valid") {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestGenerateAllowsPublicExampleHandlerModule(t *testing.T) {
	manifest := validManifest()
	manifest.HandlerImportPath = "github.com/eshu-hq/eshu/examples/collector-extensions/scorecard"
	if _, err := Generate(manifest); err != nil {
		t.Fatalf("public consumer module rejected as core implementation: %v", err)
	}
}
