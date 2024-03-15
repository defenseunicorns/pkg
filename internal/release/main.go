// package main gets the next semver version for the given package using git history
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/modfile"
)

func getProjectPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", ".."), nil
}

func cli(module string) error {
	// Ensure the path is consistent with the go mod
	baseProjectPath, err := getProjectPath()
	if err != nil {
		return fmt.Errorf("cannot get current file path")
	}
	modPath := filepath.Join(baseProjectPath, module, "go.mod")
	bytes, err := os.ReadFile(modPath)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(modPath, bytes, nil)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(modFile.Module.Mod.Path, module) {
		return fmt.Errorf("module is not named consistently with path %s", module)
	}

	filteredTags, err := getModuleTags(module)
	if err != nil {
		return err
	}

	vs := make([]*semver.Version, len(filteredTags))
	for i, r := range filteredTags {
		v, err := semver.NewVersion(r)
		if err != nil {
			return err
		}
		vs[i] = v
	}

	if len(vs) == 0 {
		// If there is not already a version, just make the version 0.0.1
		fmt.Printf("%s/v%s", module, "0.0.1")
		return nil
	}

	sort.Sort(semver.Collection(vs))
	latestVersion := vs[len(vs)-1]

	commit, err := gitLatestCommit()
	if err != nil {
		return err
	}
	category := getTypeOfChange(commit)
	var newVersion semver.Version
	switch category {
	case major:
		newVersion = latestVersion.IncMajor()
	case minor:
		newVersion = latestVersion.IncMinor()
	default:
		newVersion = latestVersion.IncPatch()
	}
	fmt.Printf("%s/v%s", module, newVersion.String())
	return nil
}

func gitLatestCommit() (string, error) {
	repoPath, err := getProjectPath()
	if err != nil {
		return "", err
	}
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get repo: %w", err)
	}

	ref, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	return commit.Message, nil
}

func getModuleTags(module string) ([]string, error) {
	repoPath, err := getProjectPath()
	if err != nil {
		return nil, err
	}
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// List all tag references
	tagRefs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	var filteredTags []string
	tagPrefix := fmt.Sprintf("refs/tags/%s/", module)
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(string(ref.Name()), tagPrefix) {
			filteredTags = append(filteredTags, strings.TrimPrefix(string(ref.Name()), tagPrefix))
		}
		return nil
	})

	return filteredTags, err
}

const (
	major = "major"
	minor = "minor"
	patch = "patch"
)

func getTypeOfChange(commit string) string {
	commitRegex := regexp.MustCompile(`^(\w+)(\([\w\-.]+\))?(!)?:(\s+.*)`)
	matches := commitRegex.FindStringSubmatch(commit)
	if matches != nil {
		commitType := matches[1]
		isBreaking := matches[3] == "!"

		if isBreaking {
			return major
		} else if commitType == "feat" {
			return minor
		}
	}
	return patch
}

func main() {
	module := os.Args[1]
	err := cli(module)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
