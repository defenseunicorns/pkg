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

	parts := strings.Split(moduleName, "/")
	name := parts[len(parts)-1]

	if strings.HasSuffix(name, "v0") || strings.HasSuffix(name, "v1") {
		name = parts[len(parts)-2]
	}

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
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error walking directory: %v", walkErr)
		os.Exit(1)
	}

	templatePath := filepath.Join(cwd, "template.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading template file: %v", err)
		os.Exit(1)
	}

	tmpl, err := template.New("readme").Parse(string(templateContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Modules []module
	}{
		Modules: modules,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v", err)
		os.Exit(1)
	}

	readmePath := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing README.md: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated README.md at %s\n", readmePath)
}
