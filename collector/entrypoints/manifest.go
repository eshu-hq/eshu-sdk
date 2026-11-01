// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package entrypoints

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/eshu-hq/eshu-sdk/collector"
	"gopkg.in/yaml.v3"
)

// Manifest describes one generated stdio command and its bounded runner.
type Manifest struct {
	SchemaVersion      int      `yaml:"schema_version"`
	CommandDir         string   `yaml:"command_dir"`
	HandlerImportPath  string   `yaml:"handler_import_path"`
	HandlerFunc        string   `yaml:"handler_func"`
	SupportedProtocols []string `yaml:"supported_protocols"`
	MaxRequestBytes    int64    `yaml:"max_request_bytes"`
	MaxResultBytes     int64    `yaml:"max_result_bytes"`
}

// ManifestSet is a versioned list of commands with unique output directories.
type ManifestSet struct {
	SchemaVersion int        `yaml:"schema_version"`
	Collectors    []Manifest `yaml:"collectors"`
}

// LoadManifestFile parses one YAML document and validates every command.
func LoadManifestFile(path string) (ManifestSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return ManifestSet{}, fmt.Errorf("open manifest: %w", err)
	}
	defer func() { _ = file.Close() }()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	var set ManifestSet
	if err := decoder.Decode(&set); err != nil {
		return ManifestSet{}, fmt.Errorf("decode manifest: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ManifestSet{}, errors.New("manifest must contain exactly one document")
	}
	if set.SchemaVersion != 1 || len(set.Collectors) == 0 {
		return ManifestSet{}, errors.New("invalid manifest set")
	}
	seen := make(map[string]bool, len(set.Collectors))
	for _, manifest := range set.Collectors {
		if err := validateManifest(manifest); err != nil {
			return ManifestSet{}, err
		}
		dir := filepath.Clean(manifest.CommandDir)
		if seen[dir] {
			return ManifestSet{}, errors.New("duplicate command directory")
		}
		seen[dir] = true
	}
	return set, nil
}

func validateManifest(m Manifest) error {
	if m.SchemaVersion != 1 {
		return errors.New("unsupported command schema version")
	}
	if !filepath.IsLocal(m.CommandDir) || m.CommandDir == "." || strings.ContainsAny(m.CommandDir, "\\\"'`:;") || strings.IndexFunc(m.CommandDir, unicode.IsControl) >= 0 {
		return errors.New("invalid command directory")
	}
	if !validImportPath(m.HandlerImportPath) {
		return errors.New("invalid handler import path")
	}
	if !token.IsIdentifier(m.HandlerFunc) || !ast.IsExported(m.HandlerFunc) {
		return errors.New("invalid handler function")
	}
	if m.MaxRequestBytes <= 0 || m.MaxRequestBytes == math.MaxInt64 || m.MaxResultBytes <= 0 {
		return errors.New("invalid byte budget")
	}
	if len(m.SupportedProtocols) == 0 {
		return errors.New("missing supported protocols")
	}
	seen := make(map[string]bool, len(m.SupportedProtocols))
	for _, protocol := range m.SupportedProtocols {
		if !admittedProtocol(protocol) {
			return errors.New("unknown protocol")
		}
		if seen[protocol] {
			return errors.New("duplicate protocol")
		}
		seen[protocol] = true
	}
	return nil
}

// isImportPathChar reports whether char may appear in one path segment of a
// handler import path: ASCII letters and digits plus the unreserved marks
// "._-~". Anything else (including whitespace, control characters, and the
// shell-meaningful set rejected above) keeps the manifest invalid.
func isImportPathChar(char rune) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char >= '0' && char <= '9' ||
		strings.ContainsRune("._-~", char)
}

func validImportPath(path string) bool {
	if path == "" || path == "github.com/eshu-hq/eshu" || path == "github.com/eshu-hq/eshu/go" || strings.HasPrefix(path, "github.com/eshu-hq/eshu/go/") || strings.ContainsAny(path, "\\\"'`@:;(){}[]") {
		return false
	}
	if strings.IndexFunc(path, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.HasPrefix(segment, ".") {
			return false
		}
		for _, char := range segment {
			if !isImportPathChar(char) {
				return false
			}
		}
	}
	return true
}

// admittedProtocol reports whether protocol is a collector SDK wire protocol
// a generated command may admit: v1alpha1, v1alpha2, or the paged v1alpha3.
func admittedProtocol(protocol string) bool {
	switch protocol {
	case collector.ProtocolVersionV1Alpha1, collector.ProtocolVersionV1Alpha2, collector.ProtocolVersionV1Alpha3:
		return true
	default:
		return false
	}
}
