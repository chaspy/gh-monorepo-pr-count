package main

import (
	"fmt"
	"log"
	"strings"

	"os"
	"path/filepath"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

func main() {
	// Get 1st argument as "since date"
	since := os.Args[1] // YYYY-MM-DD

	searchQuery := "merged:>=" + since

	// Get directory name under the current directory
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != "." && !filepath.HasPrefix(path, ".") {
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Get the current repository
	repo, err := repository.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current repository: %s/%s\n", repo.Owner, repo.Name)

	targetRepo := repo.Owner + "/" + repo.Name
	fmt.Printf("Target repository: %s\n", targetRepo)
	// To overwrite the target repository, use the GH_REPO environment variable

	// Get default branch
	defaultBranch, _, err := gh.Exec("repo", "view", targetRepo, "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name", "-t", "{{.}}")
	if err != nil {
		log.Fatal(err)
	}
	// gh query doesn't work with \n
	baseBranch := strings.ReplaceAll(defaultBranch.String(), "\n", "")

	// Count a number of PRs
	fmt.Println("counting a number of PRs...")

	// Test
	repos := []string{"api", "docs", "bin"}
	for _, repo := range repos {

		fmt.Println("checking " + repo + "...")
		// TODO: handle with json format
		prList, _, err := gh.Exec("pr", "list", "--base", baseBranch, "--repo", targetRepo, "--label", repo, "--search", searchQuery)
		if err != nil {
			log.Fatal(err)
		}

		result := strings.Split(prList.String(), "\n")
		fmt.Println(len(result) - 1)
	}
}
