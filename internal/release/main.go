// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// package main outputs the next semver version for the given package using git history and conventional commits
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

func validateModFile(path string, version *semver.Version) error {
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

	actualModPath := modFile.Module.Mod.Path
	var actualMajorVersion int64

	// Strip the /vX from the end of any mod paths
	actualModPathSections := strings.Split(actualModPath, "/")
	if len(actualModPathSections) == 5 {
		actualModPath = strings.Join(actualModPathSections[:4], "/")
		actualMajorVersion, err = strconv.ParseInt(strings.TrimPrefix(actualModPathSections[4], "v"), 10, 0)
		if err != nil {
			return err
		}
	}

	// Check that the mod path is what we expect
	expectedModPath := fmt.Sprintf("github.com/defenseunicorns/pkg/%s", path)
	if expectedModPath != actualModPath {
		return fmt.Errorf("the module name is incorrect or a %s does not exist as a module", path)
	}

	// Check that the mod version is what we expect
	expectedMajorVersion := version.Major()
	if actualMajorVersion == 0 && expectedMajorVersion > 1 {
		return fmt.Errorf("the module name does not end in /v%d for a major version > 1", expectedMajorVersion)
	} else if actualMajorVersion > 1 && actualMajorVersion != expectedMajorVersion {
		return fmt.Errorf("the expected module version /v%d does not match /v%d", expectedMajorVersion, actualMajorVersion)
	}

	return nil
}

func bumpVersion(module string, prTitle string) (*semver.Version, error) {
	repoPath, err := getProjectPath()
	if err != nil {
		return nil, err
	}

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	filteredTags, err := getModuleTags(r, module)
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

	commits, err := getCommitMessagesFromLastTag(r, latestVersion, module)
	if err != nil {
		return nil, err
	}

	if prTitle != "" {
		commits = append(commits, prTitle)
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

func getCommitMessagesFromLastTag(r *git.Repository, lastTagVersion *semver.Version, module string) ([]string, error) {
	latestTag := fmt.Sprintf("%s/v%s", module, lastTagVersion.String())

	latestTagRef, err := r.Tag(latestTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag %s: %w", latestTag, err)
	}

	tagCommit, err := r.CommitObject(latestTagRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit from tag: %w", err)
	}

	allCommits, err := r.Log(&git.LogOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	var commitsAfterTag []*object.Commit

	err = allCommits.ForEach(func(c *object.Commit) error {
		if c.Hash == tagCommit.Hash {
			// Once we reach the tag's commit, stop iterating
			return storer.ErrStop
		}
		commitsAfterTag = append(commitsAfterTag, c)
		return nil
	})

	if err != nil && !errors.Is(err, storer.ErrStop) {
		return nil, fmt.Errorf("could not iterate over commits %w", err)
	}

	pathPrefix := fmt.Sprintf("%s/", module)

	commitsOnModule, err := r.Log(&git.LogOptions{
		PathFilter: func(path string) bool {
			return strings.HasPrefix(path, pathPrefix)
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	var commitsAfterTagOnModule []string
	err = commitsOnModule.ForEach(func(c *object.Commit) error {
		for _, commitAfterTag := range commitsAfterTag {
			if commitAfterTag.Hash == c.Hash {
				commitsAfterTagOnModule = append(commitsAfterTagOnModule, c.Message)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate over commits %w", err)
	}

	return commitsAfterTagOnModule, nil
}

func getModuleTags(r *git.Repository, module string) ([]string, error) {
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
	if len(os.Args) < 2 {
		panic("this program should be called with the module name. For example, \"go run main.go helpers\"")
	}

	var prCommitMsg string
	if len(os.Args) == 3 {
		prCommitMsg = os.Args[2]
	}

	module := os.Args[1]

	newVersion, err := bumpVersion(module, prCommitMsg)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	err = validateModFile(module, newVersion)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	fmt.Printf("%s/v%s", module, newVersion.String())
	os.Exit(0)
}
