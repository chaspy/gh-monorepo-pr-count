/*
Copyright © 2023 Takeshi Kondo
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-pr-count",
	Short: "gh extension to count the number of PRs with the same label as the directory name",
	Long: `
	gh extension to count the number of PRs with the same label as the directory name.

	For example:

	gh pr-count 2023-10-01 2023-10-31 # Count the number of PRs in October 2023
	gh pr-count 2023-11-01            # Count the number of PRs since November 1st, 2023 until now
	`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gh-pr-count.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkArgs(args []string) {
	if len(args) > 3 || len(args) < 2 {
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

func get() error {
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

func run() {
	err := get()
	if err != nil {
		log.Fatal(err) //nolint:forbidigo
	}
}
