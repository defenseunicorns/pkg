// package main gets the next semver version for the given package using git history
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/plumbing/storer"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v52/github"
	"golang.org/x/mod/modfile"
)

const repoPath = "/home/austin/code/austin-pkg"

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
	fmt.Printf("modfile is %v \n", modFile.Module.Mod.Path)

	// Get the latest version
	// Configuration
	// Fetch tags
	// Filter tags based on prefix and sort them
	filteredTags, err := getTagsGit(module)
	if err != nil {
		return err
	}

	// Increment the version based on conventional commits
	vs := make([]*semver.Version, len(filteredTags))
	for i, r := range filteredTags {
		v, err := semver.NewVersion(r)
		if err != nil {
			return err
		}
		vs[i] = v
	}

	sort.Sort(semver.Collection(vs))

	fmt.Printf("this is the semver collection %v\n", vs)
	latestVersion := vs[len(vs)-1]

	commits, err := getCommitMessagesFromLastTag(latestVersion, module)
	if err != nil {
		return err
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
	fmt.Println(newVersion.String())
	return nil
}

func getCommitMessagesFromLastTag(version *semver.Version, module string) ([]string, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	compareTag := fmt.Sprintf("%s/%s", module, version.Original())
	fmt.Printf("this is the compare tag %s\n", compareTag)
	latestTagRef, err := r.Tag(compareTag)
	if err != nil {
		log.Fatalf("Failed to get tag: %s", err)
	}

	// Get the commit object for this tag
	tagCommit, err := r.CommitObject(latestTagRef.Hash())
	if err != nil {
		log.Fatalf("Failed to get commit from tag: %s", err)
	}

	headRef, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %w", err)
	}

	// Get the commit object for HEAD
	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Use Log to get all commits from HEAD to the tag's commit
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
	commits.ForEach(func(c *object.Commit) error {
		if c.Hash == tagCommit.Hash {
			// Once we reach the tag's commit, stop iterating
			return storer.ErrStop
		}
		commitMessages = append(commitMessages, c.Message)
		return nil
	})

	for _, v := range commitMessages {
		fmt.Printf("new message %s\n", v)
	}
	return commitMessages, nil
}

func getTagsGit(module string) ([]string, error) {

	// TODO change this
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
	fmt.Printf("these are the filtered tag %v\n", filteredTags)

	return filteredTags, err
}

const(
	major = "major"
	minor = "minor"
	patch = "patch"
)

func getTypeOfChange(commits []string) string {
	commitRegex := regexp.MustCompile(`^(\w+)(\([\w\-.]+\))?(!)?:(\s+.*)`)
	category := patch
	for _, commit := range commits {
		matches := commitRegex.FindStringSubmatch(commit)
		if matches != nil {
			commitType := matches[1]
			isBreaking := matches[3] == "!"

			if isBreaking {
				category = major
			} else if commitType == "feat" {
				category = minor
			}

		}
	}
	return category
}

func getTagsGithub(module string) ([]string, error) {
	owner := "austinabro321"
	repo := "pkg"

	client := github.NewClient(nil)

	tags, _, err := client.Repositories.ListTags(context.TODO(), owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredTags []string
	tagPrefix := fmt.Sprintf("%s/", module)
	for _, tag := range tags {
		if strings.HasPrefix(*tag.Name, tagPrefix) {
			filteredTags = append(filteredTags, strings.TrimPrefix(*tag.Name, tagPrefix))
		}
	}
	fmt.Printf("these are the tags %v", filteredTags)
	return filteredTags, nil
}

func main() {
	module := os.Args[1]
	err := cli(module)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
