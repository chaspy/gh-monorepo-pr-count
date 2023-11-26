package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"os"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gh pr-count YYYY-MM-DD")
		return
	}
	// Get 1st argument as "since date"
	since := os.Args[1] // YYYY-MM-DD

	searchQuery := "merged:>=" + since

	// Get the current repository
	repo, err := repository.Current()
	if err != nil {
		log.Fatal(err)
	}

	targetRepo := repo.Owner + "/" + repo.Name
	// To overwrite the target repository, use the GH_REPO environment variable

	// Get default branch
	defaultBranch, _, err := gh.Exec("repo", "view", targetRepo, "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name", "-t", "{{.}}")
	if err != nil {
		log.Fatal(err)
	}
	// gh query doesn't work with \n
	baseBranch := strings.ReplaceAll(defaultBranch.String(), "\n", "")

	// Count a number of PR for each directory
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != "." && !filepath.HasPrefix(path, ".") {
			// Skip subdirectories
			if strings.Count(path, string(os.PathSeparator)) > 0 {
				return filepath.SkipDir
			}
			// TODO: handle with json format
			prList, _, err := gh.Exec("pr", "list", "--base", baseBranch, "--repo", targetRepo, "--label", path, "--search", searchQuery, "--label", "100")

			if err != nil {
				log.Fatal(err)
			}

			result := strings.Split(prList.String(), "\n")
			num := len(result) - 1
			fmt.Printf("%s,%d\n", path, num)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
