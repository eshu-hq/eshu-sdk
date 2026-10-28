// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/eshu-hq/eshu-sdk/collector/entrypoints"
)

func run(args []string) error {
	flags := flag.NewFlagSet("collector-entrypoints-gen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	manifestPath := flags.String("manifest", "", "provider-owned manifest")
	rootPath := flags.String("repo-root", "", "repository root")
	write := flags.Bool("write", false, "write generated commands")
	check := flags.Bool("check", false, "check generated commands")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *manifestPath == "" || *rootPath == "" || *write == *check {
		return errors.New("invalid generator options")
	}
	root, err := filepath.Abs(*rootPath)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return errors.New("invalid repository root")
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer func() { _ = rootHandle.Close() }()
	set, err := entrypoints.LoadManifestFile(*manifestPath)
	if err != nil {
		return err
	}
	for _, manifest := range set.Collectors {
		files, err := entrypoints.Generate(manifest)
		if err != nil {
			return err
		}
		for _, file := range files {
			if !filepath.IsLocal(file.Name) {
				return errors.New("invalid generated filename")
			}
			path := filepath.Join(manifest.CommandDir, file.Name)
			if err := rejectSymlinkPath(rootHandle, path); err != nil {
				return err
			}
			if *check {
				current, err := rootHandle.ReadFile(path)
				if err != nil || !bytes.Equal(current, file.Contents) {
					return errors.New("missing or stale generated command")
				}
				continue
			}
			if err := rootHandle.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := rejectSymlinkPath(rootHandle, path); err != nil {
				return err
			}
			if err := rootHandle.WriteFile(path, file.Contents, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectSymlinkPath(root *os.Root, target string) error {
	if !filepath.IsLocal(target) {
		return errors.New("output outside repository root")
	}
	for path := target; path != "."; path = filepath.Dir(path) {
		info, err := root.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlink in output path")
		}
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = os.Stderr.WriteString("collector entrypoint generation failed\n")
		os.Exit(1)
	}
}
