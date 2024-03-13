// package main gets the next semver version for the given package using git history
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v52/github"
	"golang.org/x/mod/modfile"
)

func cli(module string) error {
	// Ensure the path is consistent with the go mod
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, "..", "..", module, "go.mod")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(path, bytes, nil)
	if err != nil {
		return err
	}
	fmt.Printf("modfile is %v", modFile.Module.Mod.Path)

	// Get the latest version
	// Configuration
	owner := "defenseunicorns"
	repo := "pkg"

	client := github.NewClient(nil)

	// Fetch tags
	tags, _, err := client.Repositories.ListTags(context.TODO(), owner, repo, &github.ListOptions{})
	if err != nil {
		return err
	}

	// Filter tags based on prefix and sort them
	var filteredTags []string
	for _, tag := range tags {
		if strings.HasPrefix(*tag.Name, module) {
			filteredTags = append(filteredTags, strings.TrimPrefix(*tag.Name, module))
		}
	}
	fmt.Printf("these are the tags %v", filteredTags)

	// Increment the version based on conventional commits

	return nil
}

func main() {
	module := os.Args[1]
	err := cli(module)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
