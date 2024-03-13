// package main gets the next semver version for the given package using git history
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func cli(module string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join( cwd, "..", "..", module, "go.mod")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(path, bytes, nil)
	if err != nil {
		return err
	}
	fmt.Printf("modfile is %v", modFile.Module.Mod.Path)
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
