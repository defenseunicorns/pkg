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
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/modfile"
)

func getProjectPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", ".."), nil
}

func printNewVersion(module string) error {
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

	commits, err := getCommitMessagesFromLastTag(latestVersion, module)
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits affecting module %s since last tag", module)
	}

	category := getTypeOfChange(commits)
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

func getCommitMessagesFromLastTag(lastTagVersion *semver.Version, module string) ([]string, error) {
	repoPath, err := getProjectPath()
	if err != nil {
		return nil, err
	}
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	latestTag := fmt.Sprintf("%s/%s", module, lastTagVersion.Original())
	latestTagRef, err := r.Tag(latestTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	tagCommit, err := r.CommitObject(latestTagRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit from tag: %w", err)
	}

	headRef, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %w", err)
	}

	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	pathPrefix := fmt.Sprintf("%s/", module)
	commits, err := r.Log(&git.LogOptions{
		From: headCommit.Hash,
		PathFilter: func(path string) bool {
			return strings.HasPrefix(path, pathPrefix)
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	var commitMessages []string
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash == tagCommit.Hash {
			// Once we reach the tag's commit, stop iterating
			return nil
		}
		commitMessages = append(commitMessages, c.Message)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate over commits %w", err)
	}

	return commitMessages, nil
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

func getTypeOfChange(commits []string) string {
	// https://regex101.com/r/obSlh6/1
	// Regex for conventional commits
	commitRegex := regexp.MustCompile(`^(\w+)(\([\w\-.]+\))?(!)?:(\s+.*)`)
	category := patch
	for _, commit := range commits {
		matches := commitRegex.FindStringSubmatch(commit)
		if matches != nil {
			commitType := matches[1]
			isBreaking := matches[3] == "!"

			if isBreaking {
				category = major
				break
			} else if commitType == "feat" {
				category = minor
			}

		}
	}
	return category
}

func main() {
	module := os.Args[1]
	err := printNewVersion(module)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
