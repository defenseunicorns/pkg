// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// package main generates a README.md file for the monorepo
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/mod/modfile"
)

type module struct {
	Name        string
	Description string
	Directory   string
}

func findModuleInfo(goModPath string) (module, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return module{}, fmt.Errorf("failed to read go.mod file: %w", err)
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return module{}, fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	if modFile.Module == nil {
		return module{}, fmt.Errorf("module declaration not found in go.mod")
	}

	moduleName := modFile.Module.Mod.Path

	moduleDir := filepath.Dir(goModPath)
	dirName := filepath.Base(moduleDir)

	description := ""
	err = filepath.WalkDir(moduleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != moduleDir {
			// Only process files in the root directory of the module
			if filepath.Dir(path) != moduleDir {
				return filepath.SkipDir
			}
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			lines := strings.Split(string(content), "\n")
			packageLineIdx := -1

			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "package ") {
					packageLineIdx = i
					break
				}
			}

			if packageLineIdx > 0 {
				for i := packageLineIdx - 1; i >= 0; i-- {
					line := strings.TrimSpace(lines[i])

					if line == "" || strings.HasPrefix(line, "// SPDX-") {
						continue
					}

					if strings.HasPrefix(line, "// Package ") {
						fields := strings.Fields(line)
						if len(fields) < 3 {
							break
						}
						packageName := fields[2]

						pos := strings.Index(line, "// Package ") + len("// Package ") + len(packageName) + 1

						if pos < len(line) {
							description = line[pos:]
							return filepath.SkipAll
						}
					}

					break
				}
			}
		}
		return nil
	})

	if err != nil && !errors.Is(err, filepath.SkipAll) {
		return module{}, fmt.Errorf("error scanning module directory: %w", err)
	}

	module := module{
		Description: description,
		Directory:   dirName,
		Name:        moduleName,
	}

	return module, nil
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("Error getting current directory: %v", err)
	}

	repoRoot := filepath.Join(cwd, "..", "..")
	var modules []module

	walkErr := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() && (entry.Name() == "internal" || entry.Name() == ".git") {
			return filepath.SkipDir
		}

		if !entry.IsDir() && entry.Name() == "go.mod" {
			fmt.Printf("Processing %s\n", path)
			module, err := findModuleInfo(path)
			if err != nil {
				return fmt.Errorf("failed to process %s: %w", path, err)
			}

			modules = append(modules, module)
		}
		return nil
	})

	if walkErr != nil {
		fatalf("Error walking directory: %v", walkErr)
	}

	templatePath := filepath.Join(cwd, "template.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		fatalf("Error reading template file: %v", err)
	}

	tmpl, err := template.New("readme").Parse(string(templateContent))
	if err != nil {
		fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Modules []module
	}{
		Modules: modules,
	}); err != nil {
		fatalf("Error executing template: %v", err)
	}

	readmePath := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		fatalf("Error writing README.md: %v", err)
	}

	fmt.Printf("Successfully generated README.md at %s\n", readmePath)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
