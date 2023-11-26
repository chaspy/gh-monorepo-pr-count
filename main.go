package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"os"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// nolint:forbidigo
func usage() {
	fmt.Println("Usage: gh pr-count since (until)")
	fmt.Println("example: gh pr-count 2023-10-01 // Search PRs merged since 2023-10-01 until now")
	fmt.Println("example: gh pr-count 2023-10-01 2023-11-01 // Search PRs merged since 2023-10-01 until 2023-11-01")
}

func checkArgs(args []string) {
	if len(args) > 3 || len(args) < 2 {
		usage()
		os.Exit(1)
	}
}

func makeMergedQuery(args []string) string {
	var mergedQuery string
	if len(args) == 2 {
		// If 2nd argument is empty, set until date as today
		mergedQuery = "merged:>=" + args[1]
	} else {
		mergedQuery = "merged:" + args[1] + ".." + args[2]
	}
	return mergedQuery
}

func getTargetRepo() (string, error) {
	targetRepo := os.Getenv("GH_REPO")
	if targetRepo == "" {
		// Get the current repository
		repo, err := repository.Current()
		if err != nil {
			return "", fmt.Errorf("could not determine current repository: %w", err)
		}
		targetRepo = repo.Owner + "/" + repo.Name
	}
	return targetRepo, nil
}

func isPathValid(info fs.FileInfo, path string) bool {
	if !info.IsDir() {
		return false
	}
	if path == "." {
		return false
	}
	if strings.HasPrefix(path, ".") {
		return false
	}
	return true
}

func walk(baseBranch string, targetRepo string, searchQuery string) error {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if isPathValid(info, path) {
			// Skip subdirectories
			if strings.Count(path, string(os.PathSeparator)) > 0 {
				return filepath.SkipDir
			}
			// TODO: handle with json format
			// NOTE: If GH_REPO env is set, then it is used as targetRepo in preference to the current repository
			// ref: https://cli.github.com/manual/gh_help_environment
			prList, _, err := gh.Exec("pr", "list", "--base", baseBranch, "--repo", targetRepo, "--label", path, "--search", searchQuery, "--limit", "100")

			if err != nil {
				return fmt.Errorf("could not get PR list: %w", err)
			}

			result := strings.Split(prList.String(), "\n")
			num := len(result) - 1
			fmt.Printf("%s,%d\n", path, num)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not walk: %w", err)
	}
	return nil
}

func run() error {
	checkArgs(os.Args)

	mergedQuery := makeMergedQuery(os.Args)

	// Add $SEARCH_QUERY from environment variable
	additionalSearchQuery := os.Getenv("SEARCH_QUERY")
	searchQuery := mergedQuery + " " + additionalSearchQuery

	targetRepo, err := getTargetRepo()
	if err != nil {
		return fmt.Errorf("could not get target repository: %w", err)
	}

	// Get default branch
	defaultBranch, _, err := gh.Exec("repo", "view", targetRepo, "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name", "-t", "{{.}}")
	if err != nil {
		return fmt.Errorf("could not get default branch: %w", err)
	}
	// gh query doesn't work with \n
	baseBranch := strings.ReplaceAll(defaultBranch.String(), "\n", "")

	// Count a number of PR for each directory
	err = walk(baseBranch, targetRepo, searchQuery)
	if err != nil {
		return fmt.Errorf("could not walk: %w", err)
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err) //nolint:forbidigo
	}
}
