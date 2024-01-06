package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
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

func printPRCount(baseBranch string, targetRepo string, path string, searchQuery string) error {
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

	return nil
}

func walk(baseBranch string, targetRepo string, searchQuery string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	var maxConcurrentcy int
	if os.Getenv("MAX_CONCURRENTCY") == "" {
		maxConcurrentcy = 50
	} else {
		maxConcurrentcy, _ = strconv.Atoi(os.Getenv("MAX_CONCURRENTCY"))
	}
	sem := make(chan struct{}, maxConcurrentcy)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if isPathValid(info, path) {
			// Skip subdirectories
			if strings.Count(path, string(os.PathSeparator)) > 0 {
				return filepath.SkipDir
			}

			sem <- struct{}{} // Acquire semaphore

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer func() {
					<-sem // Release semaphore
					wg.Done()
				}()

				err := printPRCount(baseBranch, targetRepo, path, searchQuery)
				if err != nil {
					errCh <- fmt.Errorf("could not print PR count: %w", err)
				}
			}(&wg)

		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not walk: %w", err)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return fmt.Errorf("could not print PR count: %w", err)
		}
	}

	return nil
}

func run() error {
	// Get current date as yyyy-mm-dd
	today := time.Now().Format("2006-01-02")

	uaFlag := flag.Bool("uniq-author", false, "Optional: Count a number of PR for each directory by uniq author")
	sinceFlag := flag.String("since", "", "Required: Search PRs merged since this date. Format: yyyy-mm-dd")
	untilFlag := flag.String("until", today, "Optional: Search PRs merged until this date. Format: yyyy-mm-dd")
	flag.Parse()

	// debug
	fmt.Println(*uaFlag)
	fmt.Println(*sinceFlag)
	fmt.Println(*untilFlag)

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
