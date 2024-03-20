// package main gets the next semver version for the given package using git history
package main

import (
	"errors"
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
	"github.com/go-git/go-git/v5/plumbing/storer"
	"golang.org/x/mod/modfile"
)

func getProjectPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(wd)), nil
}

func validateModFile(path string) error {
	// Ensure the path is consistent with the go mod
	baseProjectPath, err := getProjectPath()
	if err != nil {
		return err
	}
	modPath := filepath.Join(baseProjectPath, path, "go.mod")
	bytes, err := os.ReadFile(modPath)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(modPath, bytes, nil)
	if err != nil {
		return err
	}
	expectedModPath := fmt.Sprintf("github.com/defenseunicorns/pkg/%s", path)
	if expectedModPath != modFile.Module.Mod.Path {
		return fmt.Errorf("the module name is incorrect or a %s does not exist as a module", path)
	}
	return nil
}

func bumpVersion(module string) (*semver.Version, error) {
	filteredTags, err := getModuleTags(module)
	if err != nil {
		return nil, err
	}

	versions := make([]*semver.Version, len(filteredTags))
	for i, r := range filteredTags {
		v, err := semver.NewVersion(r)
		if err != nil {
			return nil, err
		}
		versions[i] = v
	}

	if len(versions) == 0 {
		// If there is not already a version, just make the version 0.0.1
		version, err := semver.NewVersion("0.0.1")
		if err != nil {
			return nil, err
		}
		return version, nil
	}

	sort.Sort(semver.Collection(versions))
	latestVersion := versions[len(versions)-1]

	commits, err := getCommitMessagesFromLastTag(latestVersion, module)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits affecting module %s since last tag", module)
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
	return &newVersion, nil
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
	// These commits are in the order of most recent first
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash == tagCommit.Hash {
			// Once we reach the tag's commit, stop iterating
			return storer.ErrStop
		}
		commitMessages = append(commitMessages, c.Message)
		return nil
	})

	if err != nil && !errors.Is(err, storer.ErrStop) {
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
	if len(os.Args) != 2 {
		panic("this program should be called with the module name. For example, \"go run main.go helpers\"")
	}
	module := os.Args[1]
	err := validateModFile(module)
	if err != nil {
		panic(err)
	}
	newVersion, err := bumpVersion(module)
	fmt.Printf("%s/v%s", module, newVersion.String())
	if err != nil {
		panic(err)
	}
}
