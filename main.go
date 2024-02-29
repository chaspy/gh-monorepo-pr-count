package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"os"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// nolint:forbidigo
func usage() {
	fmt.Println("Usage: gh monorepo-pr-count (--uniq-author) since (until)")
	fmt.Println("example: gh monorepo-pr-count 2023-10-01 // Search PRs merged since 2023-10-01 until now")
	fmt.Println("example: gh monorepo-pr-count 2023-10-01 2023-11-01 // Search PRs merged since 2023-10-01 until 2023-11-01")
}

func makeStateQuery(state string, since string, until string) string {
	var stateQuery string
	var queryState string

	if state == "open" {
		queryState = "created"
	} else {
		queryState = state
	}

	if until != "" {
		stateQuery = "is:" + state + " " + queryState + ":" + since + ".." + until
	} else {
		stateQuery = "is:" + state + " " + queryState + ":>=" + since
	}
	return stateQuery
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

func printPRCount(baseBranch string, targetRepo string, path string, searchQuery string, uaFlag bool, debugFlag bool) error {
	// TODO: handle with json format
	// NOTE: If GH_REPO env is set, then it is used as targetRepo in preference to the current repository
	// ref: https://cli.github.com/manual/gh_help_environment

	// This gh query only show author name of each PR, so we need to count uniq author by --uniq-author flag.
	// However, even if the --uniq-author flag is not set, the number of PRs can be obtained by counting the number of lines.

	prList, _, err := gh.Exec("pr", "list", "--base", baseBranch, "--repo", targetRepo, "--label", path, "--search", searchQuery, "--limit", "100",
		"--json", "author", "--template", "'{{range .}}{{tablerow .author.login }}{{end}}'")

	if err != nil {
		return fmt.Errorf("could not get PR list: %w", err)
	}

	result := strings.Split(prList.String(), "\n")

	if uaFlag {
		// count uniq author
		slices.Sort(result)
		result = slices.Compact(result)
	}

	num := len(result) - 1
	fmt.Printf("%s,%d\n", path, num)

	if debugFlag {
		fmt.Printf("https://github.com/%s/pulls?q=%s base:%s label:%s\n", targetRepo, searchQuery, baseBranch, path)
	}

	return nil
}

func getMaxConcurrency() (int, error) {
	var ret int
	defaultMaxConcurrency := 50
	if os.Getenv("MAX_CONCURRENCY") == "" {
		ret = defaultMaxConcurrency
	} else {
		maxConcurrentcy, err := strconv.Atoi(os.Getenv("MAX_CONCURRENCY"))
		if err != nil {
			return 0, fmt.Errorf("could not convert MAX_CONCURRENCY to int: %w", err)
		}
		ret = maxConcurrentcy
	}
	return ret, nil
}

// nolint:gocyclo
func walk(maxConcurrentcy int, baseBranch string, targetRepo string, searchQuery string, uaFlag bool, debugFlag bool) error {

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	sem := make(chan struct{}, maxConcurrentcy)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for err := range errCh {
			if err != nil {
				cancel()
			}
		}
	}()

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("problem during the walk: %w", err)
		}

		if isPathValid(info, path) {
			// Skip subdirectories
			if strings.Count(path, string(os.PathSeparator)) > 0 {
				return filepath.SkipDir
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("context is done: %w", ctx.Err())
			case sem <- struct{}{}: // Acquire semaphore
			}

			wg.Add(1)
			go func(path string){
				defer wg.Done()
				defer func() { <-sem}()

				if err := printPRCount(baseBranch, targetRepo, path, searchQuery, uaFlag, debugFlag); err != nil {
					select {
					case errCh <- err:
					case <-ctx.Done():
					}
				}
			}(path)
		}
		return nil
	})

	wg.Wait()
	close(errCh)

	if err != nil {
		return fmt.Errorf("problem during the walk: %w", err)
	}

	select {
		case err, ok := <-errCh:
			if ok && err != nil {
				return err
			}
		default:
	}

	return nil
}

func run() error {
	// Get current date as yyyy-mm-dd
	today := time.Now().Format("2006-01-02")

	uaFlag := flag.Bool("uniq-author", false, "Optional: Count a number of PR for each directory by uniq author")
	debugFlag := flag.Bool("debug-url", false, "Optional: Print debug url")

	sinceFlag := flag.String("since", "", "Required: Search PRs merged since this date. Format: yyyy-mm-dd")
	untilFlag := flag.String("until", today, "Optional: Search PRs merged until this date. Format: yyyy-mm-dd")

	stateFlag := flag.String("state", "merged", "Optional: Search PRs with the specified state(open|closed|merged). Default: merged")

	flag.Parse()

	// sinceFlag is required
	if *sinceFlag == "" {
		usage()
		os.Exit(1)
	}

	stateQuery := makeStateQuery(*stateFlag, *sinceFlag, *untilFlag)

	// Add $SEARCH_QUERY from environment variable
	additionalSearchQuery := os.Getenv("SEARCH_QUERY")
	searchQuery := stateQuery + " " + additionalSearchQuery

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

	maxConcurrentcy, err := getMaxConcurrency()
	if err != nil {
		return fmt.Errorf("could not get MAX_CONCURRENCY: %w", err)
	}

	// Count a number of PR for each directory
	err = walk(maxConcurrentcy, baseBranch, targetRepo, searchQuery, *uaFlag, *debugFlag)
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
